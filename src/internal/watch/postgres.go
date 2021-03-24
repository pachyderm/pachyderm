package watch

import (
	"context"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

const (
	minReconnectInterval = time.Second * 1
	maxReconnectInterval = time.Second * 30
)

type postgresWatcher struct {
	l            *listener
	c            chan *Event
	channelNames []string
	template     proto.Message
}

func newPostgresWatch(l *listener, channelNames []string, template proto.Message) *postgresWatcher {
	// Copy the channel names since it could be mutated
	namesCopy := make([]string, 0, len(channelNames))
	for _, name := range channelNames {
		namesCopy = append(namesCopy, name)
	}

	return &postgresWatcher{
		l:            l,
		c:            make(chan *Event),
		channelNames: namesCopy,
		template:     template,
	}
}

func (pw *postgresWatcher) Watch() <-chan *Event {
	return pw.c
}

func (pw *postgresWatcher) Close() {
	pw.l.unregister(pw)
}

type PostgresListener interface {
	Listen(channelNames []string, template proto.Message) (Watcher, error)
	Close() error
}

type watcherSet = map[*postgresWatcher]struct{}

type listener struct {
	pql    *pq.Listener
	mu     sync.Mutex
	eg     *errgroup.Group
	cancel func()

	channels map[string]watcherSet
}

func NewPostgresListener(connStr string) PostgresListener {
	ctx, cancel := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(ctx)

	l := &listener{
		pql:      pq.NewListener(connStr, minReconnectInterval, maxReconnectInterval, nil),
		cancel:   cancel,
		channels: make(map[string]watcherSet),
	}
	eg.Go(l.multiplex)

	return l
}

func (l *listener) Close() error {
	if l.eg != nil {
		l.cancel()
		if err := l.eg.Wait(); err != nil {
			return err
		}
	}

	if len(l.channels) != 0 {
		return errors.Errorf("PostgresListener closed while there are still active watches")
	}
	return nil
}

func (l *listener) Listen(channelNames []string, template proto.Message) (Watcher, error) {
	pw := newPostgresWatch(l, channelNames, template)

	l.mu.Lock()
	defer l.mu.Unlock()

	// Subscribe the watch to the given channels
	for _, name := range pw.channelNames {
		watchers, ok := l.channels[name]
		if !ok {
			watchers = make(watcherSet)
			l.channels[name] = watchers
			if err := l.pql.Listen(name); err != nil {
				return nil, errors.EnsureStack(err)
			}
		}
		watchers[pw] = struct{}{}
	}

	return pw, nil
}

func (l *listener) unregister(pw *postgresWatcher) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Remove the watch from the given channels, unlisten if any are empty
	for _, name := range pw.channelNames {
		watchers := l.channels[name]
		delete(watchers, pw)
		if len(watchers) == 0 {
			if err := l.pql.Unlisten(name); err != nil {
				return errors.EnsureStack(err)
			}
		}
	}
	return nil
}

func (l *listener) multiplex() error {
	for {
		notification, ok := <-l.pql.Notify
		if !ok {
			return nil
		}
		if notification != nil {
			// A 'nil' notification means that the connection was lost - error out all
			// current watchers so they can rebuild state.
			l.broadcastError(errors.Errorf("lost connection to database"))
		} else {
			l.routeNotification(notification)
		}
	}
}

func (l *listener) broadcastError(err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, watchers := range l.channels {
		sendError(watchers, err)
	}
}

func sendError(watchers watcherSet, err error) {
	for watcher := range watchers {
		// TODO: close the channel if the send would block
		watcher.c <- &Event{
			Err:  err,
			Type: EventError,
		}
	}
}

func (l *listener) routeNotification(n *pq.Notification) {
	l.mu.Lock()
	defer l.mu.Unlock()

	watchers, ok := l.channels[n.Channel]
	if ok {
		payload := &NotifyPayload{}
		if payloadData, err := base64.StdEncoding.DecodeString(n.Extra); err != nil {
			sendError(watchers, errors.Wrap(err, "failed to decode notification payload base64"))
		} else if err := payload.Unmarshal(payloadData); err != nil {
			sendError(watchers, errors.Wrap(err, "failed to parse notification payload protobuf"))
		} else {
			for watcher := range watchers {
				// TODO: verify this watcher was interested in this data (since there may have been a hash collision)
				// TODO: load the value from the payload (if we decide to store it inline) or from the database
				ev := &Event{Key: []byte(payload.Key), Value: nil, Type: EventType(payload.Type), Template: watcher.template}
				// TODO: close the channel if the send would block
				watcher.c <- ev
			}
		}
	}
}

func PublishNotification(tx *sqlx.Tx, channelName string, payload *NotifyPayload) error {
	payloadData, err := payload.Marshal()
	if err != nil {
		return errors.EnsureStack(err)
	}
	payloadString := base64.StdEncoding.EncodeToString(payloadData)
	_, err = tx.Exec(fmt.Sprintf("notify %s, '%s'", channelName, payloadString))
	return err
}
