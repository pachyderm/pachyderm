package postgres

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

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

var newWatcherSignal = errors.New("cancel WaitForNotification")

type Listener struct {
	once       sync.Once
	ctx        context.Context
	connConfig *pgx.ConnConfig
	watchers   chan *watcher
	errs       chan error
}

func NewListener(ctx context.Context, cc *pgx.ConnConfig) *Listener {
	return &Listener{
		ctx:        ctx,
		connConfig: cc,
		watchers:   make(chan *watcher),
		errs:       make(chan error, 1),
	}
}

func (l *Listener) Watch(ctx context.Context, channel string, bufferSize int) (<-chan *Event, <-chan error) {
	l.once.Do(func() {
		fmt.Println("qqq starting listener")
		go func() {
			l.errs <- l.start()
		}()
	})

	events := make(chan *Event, bufferSize)
	errs := make(chan error, 1)
	w := watcher{channel: channel, events: events, errs: errs, ctx: ctx}
	l.watchers <- &w
	return w.events, w.errs
}

func (l *Listener) Errs() <-chan error {
	return l.errs
}

func (l *Listener) listen(ctx context.Context, pgChannels <-chan string) (<-chan *pgconn.Notification, <-chan context.CancelCauseFunc, <-chan error) {
	notifications := make(chan *pgconn.Notification)
	cancelFns := make(chan context.CancelCauseFunc, 1)
	errs := make(chan error, 1)

	go func() {
		fmt.Println("qqq starting main listener loop")
		defer close(errs)
		defer close(notifications)
		defer close(cancelFns)

		conn, err := pgx.ConnectConfig(ctx, l.connConfig)
		if err != nil {
			errs <- err
			return
		}

		_ctx, cancel := context.WithCancelCause(ctx)
		cancelFns <- cancel
		fmt.Println("qqq sent cancel fn")
		for {
			select {
			case channel := <-pgChannels:
				fmt.Println("qqq listening to channel", channel)
				if _, err := conn.Exec(ctx, fmt.Sprintf("LISTEN %s", channel)); err != nil {
					errs <- err
					return
				}
			case <-ctx.Done():
				errs <- context.Cause(ctx)
				return
			default:
				fmt.Println("qqq waiting for notification")
				msg, err := conn.WaitForNotification(_ctx)
				if err != nil {
					if context.Cause(_ctx) == newWatcherSignal {
						fmt.Println("qqq WaitForNotification cancelled, renewing context")
						_ctx, cancel = context.WithCancelCause(ctx)
						cancelFns <- cancel
						fmt.Println("qqq sent cancel fn")
						continue
					}
					errs <- err
					return
				}
				notifications <- msg
				fmt.Println("qqq sent notification", msg)
			}
		}
	}()
	return notifications, cancelFns, errs
}

// start creates a direct connection to the database and starts .
// In each loop, it either waits for a new watcher or a notification from postgres.
// The listener exposes a context cancellation func for new watchers to cancel pgx.Conn.WaitForNotitication asynchronosly.
func (l *Listener) start() error {
	ctx := l.ctx
	// pgChannels is a channel of postgres channels to listen to.
	pgChannels := make(chan string)
	defer close(pgChannels)
	// channelsToWatchers maps postgres channels to watchers.
	channelsToWatchers := make(map[string][]*watcher)

	// Start the main listener loop.
	notifications, cancelFns, errs := l.listen(ctx, pgChannels)

	// cancel enables us to cancel pgx.Conn.WaitForNotification asynchronously.
	cancel := <-cancelFns

	fmt.Println("qqq starting broadcast loop")
	for {
		select {
		case err := <-errs:
			fmt.Println("qqq received err", err)
			return err
		case w := <-l.watchers:
			fmt.Println("qqq received watcher")
			if _, ok := channelsToWatchers[w.channel]; !ok {
				fmt.Println("qqq received cancel fn")
				cancel(newWatcherSignal)
				fmt.Println("qqq sending channel to listener")
				pgChannels <- w.channel
				fmt.Println("qqq sent channel to listener")
				cancel = <-cancelFns // Receive the next cancel func and unblock the listener.
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
		case <-ctx.Done():
			return context.Cause(ctx)
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
	ctx     context.Context
	channel string
	events  chan *Event
	errs    chan error
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
