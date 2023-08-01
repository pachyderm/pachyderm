package postgres

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

type EventType int

const (
	// EventInsert happens when an item is added
	EventInsert EventType = iota
	// EventUpdate happens when an item is updated
	EventUpdate
	// EventDelete happens when an item is removed
	EventDelete
)

const (
	DefaultBufferSize = 1000
)

var newWatcherSignal = errors.New("cancel WaitForNotification")

type Listener struct {
	once       sync.Once
	ctx        context.Context
	connConfig *pgx.ConnConfig
	watchers   chan *watcher
	cancelFns  chan context.CancelCauseFunc
	errs       chan error
}

func NewListener(ctx context.Context, cc *pgx.ConnConfig) *Listener {
	return &Listener{
		ctx:        ctx,
		connConfig: cc,
		watchers:   make(chan *watcher),
		errs:       make(chan error, 1),
		cancelFns:  make(chan context.CancelCauseFunc, 1),
	}
}

func (l *Listener) Watch(ctx context.Context, channel string, size int) (<-chan *Event, <-chan error) {
	fmt.Println("qqq Watch called on", channel)
	l.once.Do(func() {
		go func() {
			l.errs <- l.listen(l.ctx)
		}()
	})

	events := make(chan *Event, size)
	errs := make(chan error, 1)
	w := &watcher{channel: channel, events: events, errs: errs, ctx: ctx, doneListen: make(chan struct{})}

	l.watchers <- w

	<-w.doneListen
	fmt.Println("qqq watcher done registering")
	return events, errs
}

func (l *Listener) Errs() <-chan error {
	return l.errs
}

func generateCancel(cancel context.CancelCauseFunc, channelsToWatchers map[string][]*watcher) func(string) {
	// Need a static list of existing channels to avoid cancelling the context of a new watcher
	var existingChannels []string
	for channel := range channelsToWatchers {
		existingChannels = append(existingChannels, channel)
	}
	return func(channel string) {
		for _, existingChannel := range existingChannels {
			if existingChannel == channel {
				fmt.Println("qqq already listening to channel", channel, "no need to cancel")
				return
			}
		}
		cancel(newWatcherSignal)
	}
}

func (l *Listener) listen(ctx context.Context) error {
	fmt.Println("qqq starting main listener loop")

	conn, err := pgx.ConnectConfig(ctx, l.connConfig)
	if err != nil {
		return err
	}

	events := make(chan *Event)
	errs := make(chan error, 1)
	newWatchers := make(chan *watcher)
	go func(originalCtx context.Context, newWatchers <-chan *watcher) {
		ctx, cancel := context.WithCancelCause(originalCtx)
		l.cancelFns <- cancel
		fmt.Println("qqq sent cancel fn")
		for {
			select {
			case w := <-newWatchers:
				fmt.Println("qqq listen loop listening to channel", w.channel)
				if _, err := conn.Exec(originalCtx, fmt.Sprintf("LISTEN %s", w.channel)); err != nil {
					fmt.Println("qqq got error", err)
					go func() {
						errs <- err
					}()
					// TODO should we return here? This means all watchers will be cancelled.
					return
				}
				w.doneListen <- struct{}{}
			default:
				fmt.Println("qqq listen loop waiting for notification from postgres")
				msg, err := conn.WaitForNotification(ctx)
				if err != nil {
					if context.Cause(ctx) == newWatcherSignal {
						fmt.Println("qqq listen loop context cancelled, renewing context")
						ctx, cancel = context.WithCancelCause(originalCtx)
						l.cancelFns <- cancel
						fmt.Println("qqq listen loop sent cancel fn")
						continue
					}
					fmt.Println("qqq listen loop got error", err)
					errs <- err
					return
				}
				event, err := parseNotification(msg.Payload)
				if err != nil {
					errs <- err
					continue
				}
				event.Channel = msg.Channel
				events <- event
			}
		}
	}(ctx, newWatchers)

	existingWatchers := make(map[string][]*watcher)
	for {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case err := <-errs:
			return err
		case w := <-l.watchers:
			if _, ok := existingWatchers[w.channel]; !ok {
				// Net new channel, so must break the listening wait loop and call new LISTEN.
				fmt.Println("qqq broadcast loop got new watcher, cancelling listen loop", w.channel)
				cancel := <-l.cancelFns
				fmt.Println("qqq broadcast loop received cancel func")
				cancel(newWatcherSignal)
				fmt.Println("qqq broadcast sending new watcher to listen loop", w.channel)
				newWatchers <- w
				fmt.Println("qqq broadcast loop sent new watcher to listen loop", w.channel)
			} else {
				w.doneListen <- struct{}{}
			}
			existingWatchers[w.channel] = append(existingWatchers[w.channel], w)
		case event := <-events:
			fmt.Println("qqq broadcast loop broadcasting event to all watchers")
			// Instead of removing dead watchers, we simply create a new slice and keep the alive watchers instead.
			// TODO is there a more efficient way to do this?
			var watchers []*watcher
			for _, w := range existingWatchers[event.Channel] {
				select {
				case w.events <- event:
					watchers = append(watchers, w)
				case <-w.ctx.Done():
					w.errs <- context.Cause(w.ctx)
				default:
					w.errs <- errors.Errorf("buffer full, dropping watcher")
				}
			}
			// TODO handle the case where no more watchers are listening on this channel.
			// Need to call UNLISTEN on the channel.
			existingWatchers[event.Channel] = watchers
		}
	}
}

type Event struct {
	Id        uint64
	EventType EventType
	Channel   string
}

// New version of postgresWatcher
// implements both Watcher and Notifier interfaces
type watcher struct {
	ctx        context.Context
	channel    string
	events     chan<- *Event
	doneListen chan struct{}
	errs       chan error
}

func parseNotification(payload string) (*Event, error) {
	parts := strings.Split(payload, " ")
	// The payload is a string that consists of: "<TG_OP> <id>"
	if len(parts) != 2 {
		return nil, errors.Errorf("failed to parse notification payload '%s', wrong number of parts: %d", payload, len(parts))
	}
	event := Event{}
	switch parts[0] {
	case "INSERT":
		event.EventType = EventInsert
	case "UPDATE":
		event.EventType = EventUpdate
	case "DELETE":
		event.EventType = EventDelete
	default:
		return nil, errors.Errorf("failed to parse notification payload '%s', unknown TG_OP: %s", payload, parts[0])
	}
	id, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse notification payload's id")
	}
	event.Id = uint64(id)
	return &event, nil
}
