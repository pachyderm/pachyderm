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
	cancelFns  chan func(string)
	errs       chan error
}

func NewListener(ctx context.Context, cc *pgx.ConnConfig) *Listener {
	return &Listener{
		ctx:        ctx,
		connConfig: cc,
		watchers:   make(chan *watcher),
		errs:       make(chan error, 1),
		cancelFns:  make(chan func(string), 1),
	}
}

func (l *Listener) Watch(ctx context.Context, channel string, bufferSize int) (<-chan *Event, <-chan error) {
	fmt.Println("qqq Watch called on", channel)
	l.once.Do(func() {
		go func() {
			l.errs <- l.listen(l.ctx)
		}()
	})

	events := make(chan *Event, bufferSize)
	errs := make(chan error, 1)
	w := &watcher{channel: channel, events: events, errs: errs, ctx: ctx, doneListen: make(chan struct{})}
	maybeCancel := <-l.cancelFns
	maybeCancel(channel)
	fmt.Println("qqq watcher sending itself to listener")
	l.watchers <- w
	fmt.Println("qqq watcher waiting for listener to finishg registering watcher")
	<-w.doneListen
	return w.events, w.errs
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

	channelsToWatchers := make(map[string][]*watcher)

	_ctx, cancel := context.WithCancelCause(ctx)
	l.cancelFns <- generateCancel(cancel, channelsToWatchers)
	fmt.Println("qqq sent cancel fn")
	for {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case w := <-l.watchers:
			fmt.Println("qqq listener got new watcher watching for channel", w.channel)
			if _, ok := channelsToWatchers[w.channel]; !ok {
				fmt.Println("qqq listener got net new channel, executing sql LISTEN")
				// new watcher with net new postgres channel
				for channelName := range channelsToWatchers {
					fmt.Println("qqq listening to channel", channelName)
					if _, err := conn.Exec(ctx, fmt.Sprintf("LISTEN %s", channelName)); err != nil {
						fmt.Println("qqq got error", err)
						// propagate error to all watchers
						go func() {
							w.errs <- err
							for _, watchers := range channelsToWatchers {
								for _, w := range watchers {
									w.errs <- err
								}
							}
						}()
						return err
					}
				}
			}
			fmt.Println("qqq listener done registering watcher, sending done signal")
			w.doneListen <- struct{}{}
			channelsToWatchers[w.channel] = append(channelsToWatchers[w.channel], w)
		default:
			fmt.Println("qqq listener waiting for notification from postgres")
			msg, err := conn.WaitForNotification(_ctx)
			if err != nil {
				if context.Cause(_ctx) == newWatcherSignal {
					fmt.Println("qqq WaitForNotification cancelled, renewing context")
					_ctx, cancel = context.WithCancelCause(ctx)
					l.cancelFns <- generateCancel(cancel, channelsToWatchers)
					fmt.Println("qqq sent cancel fn")
					continue
				}
				fmt.Println("qqq got errror", err)
				return err
			}
			event, err := parseNotification(msg.Payload)
			if err != nil {
				for _, w := range channelsToWatchers[msg.Channel] {
					fmt.Println("qqq got error", err)
					w.errs <- err
				}
				continue
			}
			// Instead of removing dead watchers, we simply create a new slice and keep the alive watchers instead.
			// TODO is there a more efficient way to do this?
			var watchers []*watcher
			for _, w := range channelsToWatchers[msg.Channel] {
				select {
				case w.events <- event:
					watchers = append(watchers, w)
				case <-w.ctx.Done():
					w.errs <- context.Cause(w.ctx)
				default:
					w.errs <- errors.Errorf("buffer full, dropping watcher")
				}
			}
			channelsToWatchers[msg.Channel] = watchers
		}
	}
}

type Event struct {
	Id        uint64
	EventType EventType
}

// New version of postgresWatcher
// implements both Watcher and Notifier interfaces
type watcher struct {
	ctx        context.Context
	channel    string
	events     chan *Event
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
