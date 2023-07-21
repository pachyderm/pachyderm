package postgres

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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
	// EventError happens when an error occurred
	EventError
)

const (
	DefaultBufferSize = 1000
)

var newWatcherSignal = errors.New("listener cancelled")

type watchers []watcher
type Listener struct {
	watchers chan watcher
}

func NewListener() *Listener {
	return &Listener{
		watchers: make(chan watcher),
	}
}

func (l *Listener) Watch(ctx context.Context, channel string) (<-chan *Event, error) {
	w := watcher{channel: channel, events: make(chan *Event, DefaultBufferSize)}
	l.watchers <- w
	return w.events, nil
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
	channelsAndWatchers := make(map[string]watchers)

	// Start listening on postgres.
	go func(ctx context.Context) {
		defer close(cancelFns)
		_ctx, cancel := context.WithCancelCause(ctx)
		cancelFns <- cancel
		for {
			msg, err := conn.WaitForNotification(_ctx)
			if err != nil {
				if context.Cause(_ctx) == newWatcherSignal {
					_ctx, cancel = context.WithCancelCause(ctx)
					cancelFns <- cancel // for the next watcher
					continue
				}
				return
			}
			notifications <- msg
		}
	}(ctx)

	// Propagate events to watchers.
	var cancel context.CancelCauseFunc
	for {
		select {
		case cancel = <-cancelFns:
		case w := <-l.watchers:
			if _, ok := channelsAndWatchers[w.channel]; !ok {
				if cancel != nil {
					cancel(newWatcherSignal)
				}
				if _, err := conn.Exec(ctx, fmt.Sprintf("LISTEN %s", w.channel)); err != nil {
					return err
				}
			}
			channelsAndWatchers[w.channel] = append(channelsAndWatchers[w.channel], w)
		case msg := <-notifications:
			event := parseNotification(msg.Payload)
			for _, w := range channelsAndWatchers[msg.Channel] {
				select {
				case w.events <- &event:
				default:
					close(w.events)
				}
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

type Event struct {
	Id        uint64
	EventType EventType
	Error     error
}

// New version of postgresWatcher
// implements both Watcher and Notifier interfaces
type watcher struct {
	channel string
	events  chan *Event
	done    <-chan struct{}
}

func parseNotification(payload string) Event {
	parts := strings.Split(payload, " ")
	// The payload is a string that consists of: "<TG_OP> <id> <key>"
	if len(parts) != 2 {
		return Event{Error: errors.Errorf("failed to parse notification payload '%s', wrong number of parts: %d", payload, len(parts))}
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
		return Event{Error: errors.Errorf("failed to parse notification payload '%s', unknown TG_OP: %s", payload, parts[0])}
	}
	id, err := strconv.Atoi(parts[1])
	if err != nil {
		return Event{Error: errors.Wrap(err, "failed to parse notification payload's id")}
	}
	event.Id = uint64(id)
	return event
}
