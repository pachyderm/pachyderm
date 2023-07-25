package postgres

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
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
	watchers chan *watcher
}

func NewListener() *Listener {
	return &Listener{
		watchers: make(chan *watcher),
	}
}

func (l *Listener) Watch(ctx context.Context, channel string) (<-chan *Event, <-chan error) {
	events := make(chan *Event, DefaultBufferSize)
	errs := make(chan error, 1)
	w := watcher{channel: channel, events: events, errs: errs, ctx: ctx}
	l.watchers <- &w
	go func() {
		<-ctx.Done()
		w.errs <- context.Cause(ctx)
		w.close()
	}()
	return w.events, w.errs
}

// Start creates a direct connection to the database and starts .
// In each loop, it either waits for a new watcher or a notification from postgres.
// The listener exposes a context cancellation func for new watchers to cancel pgx.Conn.WaitForNotitication asynchronosly.
func (l *Listener) Start(ctx context.Context, config *pgx.ConnConfig) error {
	conn, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		return err
	}

	// cancelFns is a channel of context cancellation funcs
	// to help cancel WaitForNotification when we want to listen to a net new postgres channel.
	cancelFns := make(chan context.CancelCauseFunc)
	// notifications is a channel of notifications from postgres.
	notifications := make(chan *pgconn.Notification)
	// channelsToWatchers maps channels to watchers.
	channelsToWatchers := make(map[string][]*watcher)

	eg, ctx := errgroup.WithContext(ctx)

	// Listens for notifications from postgres.
	eg.Go(func() error {
		defer close(cancelFns)
		defer close(notifications)

		// Create a new context for WaitForNotification, because it can be repeatedly cancelled by net new watchers.
		_ctx, cancel := context.WithCancelCause(ctx)
		cancelFns <- cancel // block until the broadcaster loop receives the cancel func
		for {
			msg, err := conn.WaitForNotification(_ctx)
			if err != nil {
				if context.Cause(_ctx) == newWatcherSignal {
					_ctx, cancel = context.WithCancelCause(ctx)
					cancelFns <- cancel // for the next watcher
					continue
				}
				return err
			}
			notifications <- msg
		}
	})

	// Broadcast notifications to watchers.
	eg.Go(func() error {
		var cancel context.CancelCauseFunc
		for {
			select {
			case cancel = <-cancelFns:
			case w := <-l.watchers:
				if _, ok := channelsToWatchers[w.channel]; !ok {
					if cancel != nil {
						cancel(newWatcherSignal)
					}
					if _, err := conn.Exec(ctx, fmt.Sprintf("LISTEN %s", w.channel)); err != nil {
						return err
					}
				}
				channelsToWatchers[w.channel] = append(channelsToWatchers[w.channel], w)
			case msg := <-notifications:
				if msg == nil {
					continue
				}
				event, err := parseNotification(msg.Payload)
				if err != nil {
					for _, w := range channelsToWatchers[msg.Channel] {
						w.errs <- err
					}
					continue
				}
				for _, w := range channelsToWatchers[msg.Channel] {
					// TODO should we remove the watcher from channelsAndWatchers?
					select {
					case w.events <- event:
					case <-w.ctx.Done():
						w.errs <- context.Cause(w.ctx)
						w.close()
					default:
						w.errs <- errors.Errorf("buffer full, dropping event: %v", event)
						w.close()
					}
				}
			case <-ctx.Done():
				return context.Cause(ctx)
			case <-time.After(time.Minute):
				if err := conn.Ping(ctx); err != nil {
					// should we return this error instead?
					log.Error(ctx, fmt.Sprintf("failed to ping postgres: %v", err))
				}
			}
		}
	})
	return eg.Wait()
}

type Event struct {
	Id        uint64
	EventType EventType
}

// New version of postgresWatcher
// implements both Watcher and Notifier interfaces
type watcher struct {
	channel string
	events  chan *Event
	errs    chan error
	ctx     context.Context
	once    sync.Once
}

func (w *watcher) close() {
	w.once.Do(func() {
		close(w.events)
		close(w.errs)
	})
}

func parseNotification(payload string) (*Event, error) {
	parts := strings.Split(payload, " ")
	// The payload is a string that consists of: "<TG_OP> <id> <key>"
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
