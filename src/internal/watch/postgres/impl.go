package postgres

import (
	"strconv"
	"strings"
	"sync"

	"github.com/lib/pq"
	"github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
)

type EventType uint8

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

type Event struct {
	Id   uint64
	Type EventType
	Err  error
}

// New version of postgresWatcher
// implements both Watcher and Notifier interfaces
type Watcher struct {
	db       *pachsql.DB
	listener collection.PostgresListener
	events   chan *Event
	id       string
	channel  string
	once     sync.Once
}

type WatcherOption func(*Watcher)

func NewWatcher(db *pachsql.DB, listener collection.PostgresListener, id string, channel string, opts ...WatcherOption) (*Watcher, error) {
	w := &Watcher{
		db:       db,
		listener: listener,
		id:       id,
		channel:  channel,
	}
	for _, opt := range opts {
		opt(w)
	}
	if w.events == nil {
		w.events = make(chan *Event, DefaultBufferSize)
	}

	if err := listener.Register(w); err != nil {
		return nil, err
	}
	return w, nil
}

func WithBufferSize(size int) WatcherOption {
	return func(w *Watcher) {
		w.events = make(chan *Event, size)
	}
}

// Implement Watcher Interface
func (w *Watcher) Watch() <-chan *Event {
	return w.events
}

func (w *Watcher) Close() {
	w.once.Do(func() {
		w.listener.Unregister(w) //nolint:errcheck
		close(w.events)
	})
}

// Implement Notifier Interface

func (w *Watcher) ID() string {
	return w.id
}

func (w *Watcher) Channel() string {
	return w.channel
}

func (w *Watcher) Notify(m *pq.Notification) {
	event := parseNotification(m.Extra)
	w.send(&event)
}

func (w *Watcher) Error(err error) {
	w.send(&Event{Err: err})
}

// Send the given event to the Watcher, but abort the Watcher if the send would
// block. If this happens, the Watcher is not keeping up with events. Spawn a
// goroutine to write an error, then close the Watcher.
func (w *Watcher) send(event *Event) {
	select {
	case w.events <- event:
	default:
		go func() {
			w.listener.Unregister(w) //nolint:errcheck
			w.events <- &Event{Err: errors.Errorf("failed to send event, watcher %s is blocked", w.id)}
		}()
	}
}

func parseNotification(payload string) Event {
	parts := strings.Split(payload, " ")
	// The payload is a string that consists of: "<TG_OP> <id>"
	if len(parts) != 2 {
		return Event{Err: errors.Errorf("failed to parse notification payload '%s', wrong number of parts: %d", payload, len(parts))}
	}
	event := Event{}
	switch parts[0] {
	case "INSERT":
		event.Type = EventInsert
	case "UPDATE":
		event.Type = EventUpdate
	case "DELETE":
		event.Type = EventDelete
	default:
		return Event{Err: errors.Errorf("failed to parse notification payload '%s', unknown TG_OP: %s", payload, parts[0])}
	}
	id, err := strconv.Atoi(parts[1])
	if err != nil {
		return Event{Err: errors.Wrap(err, "failed to parse notification payload's id")}
	}
	event.Id = uint64(id)
	return event
}
