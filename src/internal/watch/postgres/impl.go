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

var newWatcherSignal = errors.New("listener cancelled")

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
	fmt.Println("qqq watcher sending itself to listener")
	l.watchers <- &w
	go func() {
		<-ctx.Done()
		w.errs <- context.Cause(ctx)
		w.close()
	}()
	return w.events, w.errs
}

// Start creates a direct connection to the database and starts the main loop.
// In each loop, it either waits for a new watcher or a notification from postgres.
// The listener exposes a context cancellation func for new watchers to cancel pgx.Conn.WaitForNotitication asynchronosly.
func (l *Listener) Start(ctx context.Context, config *pgx.ConnConfig) error {
	conn, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		return err
	}

	cancelFns := make(chan context.CancelCauseFunc)
	notifications := make(chan *pgconn.Notification)
	errFromWaitForNotification := make(chan error, 1)
	channelsAndWatchers := make(map[string][]*watcher)

	// Start listening on postgres.
	go func(ctx context.Context) {
		defer close(cancelFns)
		defer close(errFromWaitForNotification)
		defer close(notifications)

		_ctx, cancel := context.WithCancelCause(ctx)
		cancelFns <- cancel
		loop := 0
		for {
			fmt.Println("qqq listener wait loop", loop)
			msg, err := conn.WaitForNotification(_ctx)
			fmt.Println("qqq listener wait loop, WaitForNotification done")
			if err != nil {
				if context.Cause(_ctx) == newWatcherSignal {
					fmt.Println("qqq listener wait loop received newWatcherSignal, renewing context")
					_ctx, cancel = context.WithCancelCause(ctx)
					cancelFns <- cancel // for the next watcher
					fmt.Println("qqq listener wait loop sending new cancel fn")
					loop++
					continue
				}
				errFromWaitForNotification <- err
				return
			}
			fmt.Println("qqq listener wait loop sending msg to notifications")
			notifications <- msg
			loop++
		}
	}(ctx)

	// Propagate events to watchers.
	var cancel context.CancelCauseFunc
	loop := 0
	for {
		fmt.Println("qqq listener main loop", loop)
		select {
		case cancel = <-cancelFns:
			fmt.Println("qqq listener main loop received cancel fn")
		case w := <-l.watchers:
			fmt.Println("qqq listener main loop received watcher", w.channel)
			if _, ok := channelsAndWatchers[w.channel]; !ok {
				if cancel != nil {
					fmt.Println("qqq listener main loop cancelling wait loop")
					cancel(newWatcherSignal)
				}
				if _, err := conn.Exec(ctx, fmt.Sprintf("LISTEN %s", w.channel)); err != nil {
					return err
				}
			}
			channelsAndWatchers[w.channel] = append(channelsAndWatchers[w.channel], w)
		case msg := <-notifications:
			fmt.Println("qqq listener main loop received msg", msg)
			event, err := parseNotification(msg.Payload)
			// TODO should we exist early on error?
			for _, w := range channelsAndWatchers[msg.Channel] {
				if err != nil {
					w.errs <- err
					continue
				}
				// TODO should we remove the watcher from channelsAndWatchers?
				select {
				case w.events <- event:
				case <-w.ctx.Done():
					w.errs <- context.Cause(w.ctx)
					fmt.Println("qqq listener main loop, watcher done", w.channel)
					w.close()
				default:
					w.errs <- errors.Errorf("buffer full, dropping event: %v", event)
					w.close()
				}
			}
		case <-ctx.Done():
			return context.Cause(ctx)
		case err := <-errFromWaitForNotification:
			return err
		case <-time.After(time.Minute):
			go func() {
				if err := conn.Ping(ctx); err != nil {
					errFromWaitForNotification <- err
				}
			}()
		}
		loop++
	}
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
		fmt.Println("qqq watcher closing itself")
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
