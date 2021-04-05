package collection

import (
	"context"
	"encoding/base64"
	"math"
	"strconv"
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
	channelBufferSize    = 1000
)

type postgresWatcher struct {
	listener   *PostgresListener
	c          chan *watch.Event
	buf        chan *postgresEvent // buffer for messages before the initial table list is complete
	done       chan struct{}       // closed when the watcher is closed to interrupt selects
	template   proto.Message
	sqlChannel string

	// Filtering variables:
	opts      watch.WatchOptions // may filter by the operation type (put or delete)
	index     *string            // only set if the watch filters by an index (for dealing with hash collisions)
	value     *string            // only set if 'index' is set
	startTime *time.Time         // only set once the initial list has completed and we know the start time for the watch
}

func newPostgresWatch(
	listener *PostgresListener,
	sqlChannel string,
	template proto.Message,
	index *string,
	value *string,
	opts watch.WatchOptions,
) *postgresWatcher {
	return &postgresWatcher{
		listener:   listener,
		c:          make(chan *watch.Event),
		buf:        make(chan *postgresEvent, channelBufferSize),
		done:       make(chan struct{}),
		template:   template,
		sqlChannel: sqlChannel,
		opts:       opts,
		index:      index,
		value:      value,
	}
}

func (pw *postgresWatcher) Watch() <-chan *watch.Event {
	return pw.c
}

func (pw *postgresWatcher) Close() {
	// Close the 'done' channel to interrupt any waiting writes - because of this,
	// a postgres watcher should not be closed multiple times.
	close(pw.done)
	pw.listener.unregister(pw)
}

func (pw *postgresWatcher) isClosed() bool {
	select {
	case <-pw.done:
		return true
	default:
		return false
	}
}

// `forwardNotifications` is a blocking call that will forward all messages on
// the 'buf' channel to the 'c' channel until the watcher is closed. It will
// apply the 'startTime' timestamp as a lower bound on forwarded notifications.
func (pw *postgresWatcher) forwardNotifications(ctx context.Context, startTime time.Time) {
	for {
		// This allows for double notification on a matching timestamp which is
		// better than missing a notification. This will _always_ happen on the last
		// put event if a change is made while doing the list.
		// We could try to use the postgres xmin value instead of the modified
		// timestamp, but that has to detect wraparounds, which is non-trivial.
		select {
		case eventData := <-pw.buf:
			if !eventData.time.Before(startTime) || eventData.err != nil {
				// This may block, but that is fine, it will put back pressure on the 'buf' channel
				pw.c <- eventData.WatchEvent(pw.template)
			}
		case <-pw.done:
			// watcher has been closed, safe to abort
			return
		case <-ctx.Done():
			// watcher (or the read collection that created it) has been canceled -
			// unregister the watcher and stop forwarding notifications
			pw.c <- &watch.Event{Type: watch.EventError, Err: ctx.Err()}
			pw.listener.unregister(pw)
			return
		}
	}
}

func (pw *postgresWatcher) sendInitial(event *watch.Event) error {
	select {
	case pw.c <- event:
		return nil
	case <-pw.done:
		return errors.New("failed to send initial event, watcher has been closed")
	}
}

// Send the given event to the watcher, but abort the watcher if the send would
// block. If this happens, the watcher is not keeping up with events. Spawn a
// goroutine to write an error, then close the watcher.
func (pw *postgresWatcher) sendChange(eventData *postgresEvent) {
	indexMatch := pw.index == nil || (*pw.index == eventData.index && *pw.value == eventData.value)
	typeMatch := (pw.opts.IncludePut && eventData.eventType == watch.EventPut) ||
		(pw.opts.IncludeDelete && eventData.eventType == watch.EventDelete)
	interested := indexMatch && typeMatch

	if interested || eventData.err != nil {
		select {
		case pw.buf <- eventData:
		default:
			// Sending would block which means the user is not keeping up with events
			// and the buffer is full. We have no option but to abort the watch.
			// Do this all in a separate goroutine because we need to avoid
			// recursively locking the listener.
			go func() {
				// Unregister the watcher first, so we stop attempting to send it events
				// (this will happen again in pw.Close(), but it will be a no-op).
				pw.listener.unregister(pw)

				select {
				case pw.buf <- &postgresEvent{err: errors.New("watcher channel is full, aborting watch")}:
				case <-pw.done:
				}
			}()
		}
	}
}

type postgresEvent struct {
	index     string          // the index that was notified by this event
	value     string          // the value of the index for the notified row
	key       string          // the primary key of the notified row
	eventType watch.EventType // the type of operation that generated the event
	time      time.Time       // the time that this event was created in postgres
	protoData []byte          // the serialized protobuf of the row data
	err       error           // any error that occurred in parsing
}

func parsePostgresEpoch(s string) (time.Time, error) {
	parts := strings.Split(s, ".")
	sec, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return time.Time{}, errors.EnsureStack(err)
	}
	if len(parts) == 1 {
		return time.Unix(sec, 0).In(time.UTC), nil
	}

	nsec, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return time.Time{}, errors.EnsureStack(err)
	}
	nsec = int64(float64(nsec) * math.Pow10(9-len(parts[1])))

	return time.Unix(sec, nsec).In(time.UTC), nil
}

func parsePostgresEvent(payload string) *postgresEvent {
	// TODO: do this in a streaming manner rather than copying the string
	parts := strings.Split(payload, " ") // index, type, date, data
	if len(parts) != 6 {
		return &postgresEvent{err: errors.Errorf("failed to parse notification payload, wrong number of parts: %d", len(parts))}
	}

	value, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return &postgresEvent{err: errors.Wrap(err, "failed to decode notification payload index value base64")}
	}

	key, err := base64.StdEncoding.DecodeString(parts[2])
	if err != nil {
		return &postgresEvent{err: errors.Wrap(err, "failed to decode notification payload key base64")}
	}

	eventType := watch.EventError
	switch parts[3] {
	case "INSERT", "UPDATE":
		eventType = watch.EventPut
	case "DELETE":
		eventType = watch.EventDelete
	}
	if eventType == watch.EventError {
		return &postgresEvent{err: errors.Errorf("failed to decode notification payload operation type: %s", parts[3])}
	}

	eventTime, err := parsePostgresEpoch(parts[4])
	if err != nil {
		return &postgresEvent{err: errors.Wrapf(err, "failed to decode notification payload timestamp: %s", parts[4])}
	}

	protoData, err := base64.StdEncoding.DecodeString(parts[5])
	if err != nil {
		return &postgresEvent{err: errors.Wrap(err, "failed to decode notification payload proto base64")}
	}

	return &postgresEvent{
		index:     parts[0],
		value:     string(value),
		key:       string(key),
		eventType: eventType,
		time:      eventTime,
		protoData: protoData,
	}
}

func (pe *postgresEvent) WatchEvent(template proto.Message) *watch.Event {
	if pe.err != nil {
		return &watch.Event{Err: pe.err, Type: watch.EventError}
	}
	return &watch.Event{
		Key:      []byte(pe.key),
		Value:    pe.protoData,
		Type:     pe.eventType,
		Template: template,
	}
}

type watcherSet = map[*postgresWatcher]struct{}

type PostgresListener struct {
	dsn    string
	pql    *pq.Listener
	mu     sync.Mutex
	eg     *errgroup.Group
	closed bool

	channels map[string]watcherSet
}

func NewPostgresListener(dsn string) *PostgresListener {
	eg, _ := errgroup.WithContext(context.Background())

	l := &PostgresListener{
		dsn:      dsn,
		eg:       eg,
		channels: make(map[string]watcherSet),
	}

	return l
}

func (l *PostgresListener) Close() error {
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
func (l *PostgresListener) getPQL() *pq.Listener {
	if l.pql == nil {
		l.pql = pq.NewListener(l.dsn, minReconnectInterval, maxReconnectInterval, nil)
		l.eg.Go(func() error {
			return l.multiplex(l.pql.Notify)
		})
	}
	return l.pql
}

func (l *PostgresListener) reset(err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, watchers := range l.channels {
		eventData := &postgresEvent{err: err}
		for watcher := range watchers {
			watcher.sendChange(eventData)
		}
	}

	l.channels = make(map[string]watcherSet)
	if err := l.getPQL().UnlistenAll(); err != nil {
		// `reset` is only ever called in the case of an error, so it should be fine to discard this error
	}
}

// TODO: I don't think we actually ever need to listen to multiple channels for a watch?
func (l *PostgresListener) listen(sqlChannel string, template proto.Message, index *string, value *string, opts watch.WatchOptions) (*postgresWatcher, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return nil, errors.New("PostgresListener has been closed")
	}

	pw := newPostgresWatch(l, sqlChannel, template, index, value, opts)

	// Subscribe the watch to the given channel
	watchers, ok := l.channels[sqlChannel]
	if !ok {
		watchers = make(watcherSet)
		l.channels[sqlChannel] = watchers
		if err := l.getPQL().Listen(sqlChannel); err != nil {
			// If an error occurs, error out all watches and reset the state of the
			// listener to prevent desyncs
			err = errors.EnsureStack(err)
			l.reset(err)
			return nil, err
		}
	}
	watchers[pw] = struct{}{}

	return pw, nil
}

func (l *PostgresListener) unregister(pw *postgresWatcher) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Remove the watch from the given channels, unlisten if any are empty
	watchers := l.channels[pw.sqlChannel]
	if len(watchers) > 0 {
		delete(watchers, pw)
		if len(watchers) == 0 {
			delete(l.channels, pw.sqlChannel)
			if err := l.getPQL().Unlisten(pw.sqlChannel); err != nil {
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

func (l *PostgresListener) multiplex(notifyChan chan *pq.Notification) error {
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

func (l *PostgresListener) routeNotification(notification *pq.Notification) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Ignore any messages from channels we have no watchers for
	watchers, ok := l.channels[notification.Channel]
	if ok {
		eventData := parsePostgresEvent(notification.Extra)
		for watcher := range watchers {
			watcher.sendChange(eventData)
		}
	}
}
