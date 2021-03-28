package collection

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/lib/pq"
	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/watch"
)

const (
	minReconnectInterval = time.Second * 1
	maxReconnectInterval = time.Second * 30
)

type postgresWatcher struct {
	l            *listener
	c            chan *watch.Event
	closeMutex   sync.Mutex
	closed       bool
	template     proto.Message
	channelNames []string
}

func newPostgresWatch(l *listener, channelNames []string, template proto.Message) *postgresWatcher {
	// Copy the channel names since it could be mutated
	namesCopy := make([]string, 0, len(channelNames))
	for _, name := range channelNames {
		namesCopy = append(namesCopy, name)
	}

	return &postgresWatcher{
		l:            l,
		c:            make(chan *watch.Event),
		template:     template,
		channelNames: namesCopy,
	}
}

func (pw *postgresWatcher) Watch() <-chan *watch.Event {
	return pw.c
}

func (pw *postgresWatcher) Close() {
	pw.closeMutex.Lock()
	defer pw.closeMutex.Unlock()

	pw.l.unregister(pw)

	if !pw.closed {
		pw.closed = true
		close(pw.c)
	}
}

type PostgresListener interface {
	Listen(channelNames []string, template proto.Message) (watch.Watcher, error)
	Close() error
}

type watcherSet = map[*postgresWatcher]struct{}

type listener struct {
	dsn    string
	pql    *pq.Listener
	mu     sync.Mutex
	eg     *errgroup.Group
	closed bool

	channels map[string]watcherSet
}

func NewPostgresListener(dsn string) PostgresListener {
	eg, _ := errgroup.WithContext(context.Background())

	l := &listener{
		dsn:      dsn,
		eg:       eg,
		channels: make(map[string]watcherSet),
	}

	return l
}

func (l *listener) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.closed = true

	if l.pql != nil {
		if err := l.pql.Close(); err != nil {
			return err
		}
	}

	if err := l.eg.Wait(); err != nil {
		return err
	}

	if len(l.channels) != 0 {
		return errors.Errorf("PostgresListener closed while there are still active watches")
	}
	return nil
}

// Lazily generate the pq.Listener because we don't want to trigger a race
// condition where an unused pq.Listener can't be closed properly. The mutex
// must be locked before calling this.
func (l *listener) getPQL() *pq.Listener {
	if l.pql == nil {
		l.pql = pq.NewListener(l.dsn, minReconnectInterval, maxReconnectInterval, nil)
		l.eg.Go(func() error {
			return l.multiplex(l.pql.Notify)
		})
	}
	return l.pql
}

func (l *listener) multiplex(notifyChan chan *pq.Notification) error {
	for {
		notification, ok := <-notifyChan
		if !ok {
			return nil
		}
		if notification == nil {
			// A 'nil' notification means that the connection was lost - error out all
			// current watchers so they can rebuild state.
			l.reset(errors.Errorf("lost connection to database"))
		} else {
			l.routeNotification(notification)
		}
	}
}

func (l *listener) reset(err error) {
	for _, watchers := range l.channels {
		sendError(watchers, err)
	}

	l.channels = make(map[string]watcherSet)
	if err := l.getPQL().UnlistenAll(); err != nil {
		// `reset` is only ever called in the case of an error, so it should be fine to discard this error
	}
}

func (l *listener) Listen(channelNames []string, template proto.Message) (watch.Watcher, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return nil, errors.New("PostgresListener has been closed")
	}

	pw := newPostgresWatch(l, channelNames, template)

	// Subscribe the watch to the given channels
	for _, name := range pw.channelNames {
		watchers, ok := l.channels[name]
		if !ok {
			watchers = make(watcherSet)
			l.channels[name] = watchers
			if err := l.getPQL().Listen(name); err != nil {
				// If an error occurs, error out all watches and reset the state of the
				// listener to prevent desyncs
				err = errors.EnsureStack(err)
				l.reset(err)
				return nil, err
			}
		}
		watchers[pw] = struct{}{}
	}

	return pw, nil
}

func (l *listener) unregister(pw *postgresWatcher) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return errors.New("PostgresListener has been closed")
	}

	// Remove the watch from the given channels, unlisten if any are empty
	for _, name := range pw.channelNames {
		watchers := l.channels[name]
		delete(watchers, pw)
		if len(watchers) == 0 {
			delete(l.channels, name)
			if err := l.getPQL().Unlisten(name); err != nil {
				// If an error occurs, error out all watches and reset the state of the
				// listener to prevent desyncs
				err = errors.EnsureStack(err)
				l.reset(err)
				return err
			}
		}
	}
	return nil
}

// Send the given event to the watcher, but abort the watcher if the send would
// block. If this happens, the watcher is not keeping up with events. Spawn a
// goroutine to write an error, then close the watcher.
func sendToWatcher(watcher *postgresWatcher, ev *watch.Event) {
	select {
	case watcher.c <- ev:
	default:
		// Unregister the watcher first, so we stop attempting to send it events
		// (this will happen again in watcher.Close(), but be a no-op).
		watcher.l.unregister(watcher)

		go func() {
			watcher.c <- &watch.Event{Err: errors.New("watcher channel is full, aborting watch"), Type: watch.EventError}
			watcher.Close()
		}()
	}
}

func sendError(watchers watcherSet, err error) {
	ev := &watch.Event{Err: err, Type: watch.EventError}
	for watcher := range watchers {
		sendToWatcher(watcher, ev)
	}
}

func (l *listener) routeNotification(n *pq.Notification) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Ignore any messages from channels we have no watchers for
	watchers, ok := l.channels[n.Channel]
	if ok {
		fmt.Printf("n.Extra: %s\n", n.Extra)
		parts := strings.Split(n.Extra, " ") // key, type, date, data

		eventType := watch.EventError
		switch parts[1] {
		case "INSERT", "UPDATE":
			eventType = watch.EventPut
		case "DELETE":
			eventType = watch.EventDelete
		}

		if eventType == watch.EventError {
			sendError(watchers, errors.Errorf("failed to decode notification payload operation type: %s", parts[1]))
		} else if protoData, err := base64.StdEncoding.DecodeString(parts[3]); err != nil {
			sendError(watchers, errors.Wrap(err, "failed to decode notification payload base64"))
		} else {
			for watcher := range watchers {
				// TODO: verify this watcher was interested in this data (since there may have been a hash collision)
				ev := &watch.Event{Key: []byte(parts[0]), Value: protoData, Type: eventType, Template: watcher.template}
				sendToWatcher(watcher, ev)
			}
		}
	}
}
