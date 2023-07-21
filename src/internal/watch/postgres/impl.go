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
	// EventError happens when an error occurred
	EventError
)

const (
	DefaultBufferSize = 1000
)

var stopWaitForNotificationSignal = errors.New("listener cancelled")

type watchers []watcher
type Listener struct {
	mu          sync.Mutex
	conn        *pgx.Conn
	channels    map[string]watchers
	cancelCause context.CancelCauseFunc // allow watchers to unblock conn.WaitForNotification
}

func NewListener(conn *pgx.Conn) Listener {
	return Listener{
		conn:        conn,
		channels:    make(map[string]watchers),
		cancelCause: nil,
	}
}

func (l *Listener) Watch(ctx context.Context, channel string) (<-chan *Event, error) {
	if l.cancelCause != nil {
		l.cancelCause(stopWaitForNotificationSignal)
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, ok := l.channels[channel]; !ok {
		if _, err := l.conn.Exec(ctx, fmt.Sprintf("LISTEN %s", channel)); err != nil {
			return nil, err
		}
	}

	w := watcher{events: make(chan *Event, DefaultBufferSize), done: ctx.Done()}
	l.channels[channel] = append(l.channels[channel], w)
	return w.events, nil
}

func (l *Listener) Start(ctx context.Context) error {
	ctx, l.cancelCause = context.WithCancelCause(ctx)

	for {
		l.mu.Lock()
		msg, err := l.conn.WaitForNotification(ctx)
		if err != nil {
			l.mu.Unlock()
			if context.Cause(ctx) == stopWaitForNotificationSignal {
				ctx, l.cancelCause = context.WithCancelCause(ctx)
				continue
			}
			return err
		}

		event := parseNotification(msg.Payload)
		for _, w := range l.channels[msg.Channel] {
			select {
			case w.events <- &event:
			default:
				close(w.events)
			}
		}
		l.mu.Unlock()
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
	events chan *Event
	done   <-chan struct{}
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
