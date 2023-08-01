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
	once                 sync.Once
	ctx                  context.Context
	connConfig           *pgx.ConnConfig
	watchers             chan *watcher
	cancelFns            chan context.CancelCauseFunc
	errs                 chan error
	existingWatchers     map[string][]*watcher // map of channel to watchers
	existingWatchersLock sync.Mutex
}

func NewListener(ctx context.Context, cc *pgx.ConnConfig) *Listener {
	return &Listener{
		ctx:              ctx,
		connConfig:       cc,
		watchers:         make(chan *watcher),
		errs:             make(chan error, 1),
		cancelFns:        make(chan context.CancelCauseFunc, 1),
		existingWatchers: make(map[string][]*watcher),
	}
}

func (l *Listener) Watch(ctx context.Context, channel string, events chan<- *Event) <-chan error {
	fmt.Println("qqq Watch called on", channel)
	l.once.Do(func() {
		go func() {
			l.errs <- l.listen(l.ctx)
		}()
	})

	errs := make(chan error, 1)
	w := &watcher{channel: channel, events: events, errs: errs, ctx: ctx, doneListen: make(chan struct{})}

	l.existingWatchersLock.Lock()
	if _, ok := l.existingWatchers[channel]; !ok {
		cancel := <-l.cancelFns
		cancel(newWatcherSignal)
		l.existingWatchers[channel] = []*watcher{w}
		fmt.Println("qqq watcher sending itself to listener")
		l.watchers <- w
		fmt.Println("qqq watcher waiting for listener to finishg registering watcher")
	} else {
		l.existingWatchers[channel] = append(l.existingWatchers[channel], w)
	}
	l.existingWatchersLock.Unlock()

	<-w.doneListen
	return w.errs
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
	go func(originalCtx context.Context) {
		ctx, cancel := context.WithCancelCause(originalCtx)
		l.cancelFns <- cancel
		fmt.Println("qqq sent cancel fn")
		for {
			fmt.Println("qqq listener waiting for notification from postgres")
			msg, err := conn.WaitForNotification(ctx)
			if err != nil {
				if context.Cause(ctx) == newWatcherSignal {
					fmt.Println("qqq WaitForNotification cancelled, renewing context")
					ctx, cancel = context.WithCancelCause(originalCtx)
					l.cancelFns <- cancel
					fmt.Println("qqq sent cancel fn")
					continue
				}
				fmt.Println("qqq got errror", err)
				errs <- err
			}
			event, err := parseNotification(msg.Payload)
			if err != nil {
				errs <- err
				continue
			}
			event.Channel = msg.Channel
			events <- event
		}
	}(ctx)

	for {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case w := <-l.watchers:
			fmt.Println("qqq listener got new watcher watching for channel", w.channel)
			// We need to execute LISTEN for all existing channels
			l.existingWatchersLock.Lock()
			for channelName := range l.existingWatchers {
				fmt.Println("qqq listening to channel", channelName)
				if _, err := conn.Exec(ctx, fmt.Sprintf("LISTEN %s", channelName)); err != nil {
					fmt.Println("qqq got error", err)
					// propagate error to all watchers
					go func() {
						w.errs <- err
						for _, watchers := range l.existingWatchers {
							for _, w := range watchers {
								w.errs <- err
							}
						}
					}()
					return err
				}
			}
			l.existingWatchersLock.Unlock()
			fmt.Println("qqq listener done registering watcher, sending done signal")
			w.doneListen <- struct{}{}
		case event := <-events:
			// Instead of removing dead watchers, we simply create a new slice and keep the alive watchers instead.
			// TODO is there a more efficient way to do this?
			var watchers []*watcher
			for _, w := range l.existingWatchers[event.Channel] {
				select {
				case w.events <- event:
					watchers = append(watchers, w)
				case <-w.ctx.Done():
					w.errs <- context.Cause(w.ctx)
				default:
					w.errs <- errors.Errorf("buffer full, dropping watcher")
				}
			}
			l.existingWatchers[event.Channel] = watchers
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
