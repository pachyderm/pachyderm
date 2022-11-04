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
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/internal/watch"
)

type PostgresListener interface {
	// Register registers a notifier with the postgres listener.
	// A notifier will receive notifications for the channel it is associated
	// with while registered with the postgres listener.
	Register(Notifier) error
	// Unregister unregisters a notifier with the postgres listener.
	// A notifier will no longer receive notifications when this call completes.
	Unregister(Notifier) error
	// Close closes the postgres listener.
	Close() error
}

type Notification = pq.Notification

type Notifier interface {
	// ID is a unique identifier for the notifier.
	ID() string
	// Channel is the channel that this notifier should receive notifications for.
	Channel() string
	// Notify sends a notification to the notifier.
	Notify(*Notification)
	// Error sends an error to the notifier.
	Error(error)
}

const (
	minReconnectInterval = time.Second * 1
	maxReconnectInterval = time.Second * 30
	ChannelBufferSize    = 1000
)

type postgresWatcher struct {
	db       *pachsql.DB
	listener PostgresListener
	c        chan *watch.Event
	buf      chan *postgresEvent // buffer for messages before the initial table list is complete
	done     chan struct{}       // closed when the watcher is closed to interrupt selects
	template proto.Message
	id       string
	channel  string
	closer   sync.Once

	// Filtering variables:
	opts  watch.WatchOptions // may filter by the operation type (put or delete)
	index *string            // only set if the watch filters by an index (for dealing with hash collisions)
	value *string            // only set if 'index' is set
}

func newPostgresWatcher(
	db *pachsql.DB,
	listener PostgresListener,
	channel string,
	template proto.Message,
	index *string,
	value *string,
	opts watch.WatchOptions,
) (*postgresWatcher, error) {
	pw := &postgresWatcher{
		db:       db,
		listener: listener,
		c:        make(chan *watch.Event),
		buf:      make(chan *postgresEvent, ChannelBufferSize),
		done:     make(chan struct{}),
		template: template,
		id:       uuid.NewWithoutDashes(),
		channel:  channel,
		opts:     opts,
		index:    index,
		value:    value,
	}
	if err := listener.Register(pw); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return pw, nil
}

func (pw *postgresWatcher) Watch() <-chan *watch.Event {
	return pw.c
}

func (pw *postgresWatcher) Close() {
	pw.closer.Do(func() {
		// Close the 'done' channel to interrupt any waiting writes
		close(pw.done)
		pw.listener.Unregister(pw) //nolint:errcheck
	})
}

// `forwardNotifications` is a blocking call that will forward all messages on
// the 'buf' channel to the 'c' channel until the watcher is closed.
func (pw *postgresWatcher) forwardNotifications(ctx context.Context) {
	for {
		select {
		case eventData := <-pw.buf:
			select {
			case pw.c <- eventData.WatchEvent(ctx, pw.db, pw.template):
			case <-pw.done:
				// watcher has been closed, safe to abort
				return
			case <-ctx.Done():
				// watcher (or the read collection that created it) has been canceled -
				// unregister the watcher and stop forwarding notifications.
				select {
				case pw.c <- &watch.Event{Type: watch.EventError, Err: ctx.Err()}:
					pw.listener.Unregister(pw) //nolint:errcheck
				case <-pw.done:
				}
				return
			}
		case <-pw.done:
			// watcher has been closed, safe to abort
			return
		case <-ctx.Done():
			// watcher (or the read collection that created it) has been canceled -
			// unregister the watcher and stop forwarding notifications.
			select {
			case pw.c <- &watch.Event{Type: watch.EventError, Err: ctx.Err()}:
				pw.listener.Unregister(pw) //nolint:errcheck
			case <-pw.done:
			}
			return
		}
	}
}

func (pw *postgresWatcher) sendInitial(ctx context.Context, event *watch.Event) error {
	select {
	case pw.c <- event:
		return nil
	case <-ctx.Done():
		return errors.EnsureStack(ctx.Err())
	case <-pw.done:
		return errors.New("failed to send initial event, watcher has been closed")
	}
}

func (pw *postgresWatcher) ID() string {
	return pw.id
}

func (pw *postgresWatcher) Channel() string {
	return pw.channel
}

func (pw *postgresWatcher) Notify(notification *pq.Notification) {
	event := parsePostgresEvent(notification.Extra)
	indexMatch := pw.index == nil || (*pw.index == event.index && *pw.value == event.value)
	typeMatch := (pw.opts.IncludePut && event.eventType == watch.EventPut) ||
		(pw.opts.IncludeDelete && event.eventType == watch.EventDelete)
	interested := indexMatch && typeMatch
	if interested {
		pw.send(event)
	}
}

func (pw *postgresWatcher) Error(err error) {
	pw.send(&postgresEvent{err: err})
}

// Send the given event to the watcher, but abort the watcher if the send would
// block. If this happens, the watcher is not keeping up with events. Spawn a
// goroutine to write an error, then close the watcher.
func (pw *postgresWatcher) send(event *postgresEvent) {
	select {
	case pw.buf <- event:
	default:
		// Sending would block which means the user is not keeping up with events
		// and the buffer is full. We have no option but to abort the watch.
		// Do this all in a separate goroutine because we need to avoid
		// recursively locking the listener.
		go func() {
			// Unregister the watcher first, so we stop attempting to send it events
			// (this will happen again in pw.Close(), but it will be a no-op).
			pw.listener.Unregister(pw) //nolint:errcheck

			select {
			case pw.buf <- &postgresEvent{err: errors.New("watcher buffer is full, aborting watch")}:
			case <-pw.done:
			}
		}()
	}
}

type postgresEvent struct {
	index     string          // the index that was notified by this event
	value     string          // the value of the index for the notified row
	key       string          // the primary key of the notified row
	eventType watch.EventType // the type of operation that generated the event
	time      time.Time       // the time that this event was created in postgres
	protoData []byte          // the serialized protobuf of the row data (exclusive with storedID)
	storedID  string          // the ID of a temporary row in the large_notifications table
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
	parts := strings.Split(payload, " ")
	if len(parts) != 7 {
		return &postgresEvent{err: errors.Errorf("failed to parse notification payload, wrong number of parts: %d", len(parts))}
	}

	value, err := base64.StdEncoding.DecodeString(parts[4])
	if err != nil {
		return &postgresEvent{err: errors.Wrap(err, "failed to decode notification payload index value base64")}
	}

	key, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return &postgresEvent{err: errors.Wrap(err, "failed to decode notification payload key base64")}
	}

	result := &postgresEvent{
		index:     parts[3],
		value:     string(value),
		key:       string(key),
		eventType: watch.EventError,
	}

	switch parts[2] {
	case "INSERT", "UPDATE":
		result.eventType = watch.EventPut
	case "DELETE":
		result.eventType = watch.EventDelete
	}
	if result.eventType == watch.EventError {
		return &postgresEvent{err: errors.Errorf("failed to decode notification payload operation type: %s", parts[3])}
	}

	result.time, err = parsePostgresEpoch(parts[1])
	if err != nil {
		return &postgresEvent{err: errors.Wrapf(err, "failed to decode notification payload timestamp: %s", parts[4])}
	}

	switch parts[5] {
	case "inline":
		result.protoData, err = base64.StdEncoding.DecodeString(parts[6])
		if err != nil {
			return &postgresEvent{err: errors.Wrap(err, "failed to decode notification payload proto base64")}
		}
	case "stored":
		result.storedID = parts[6]
	}

	return result
}

func (pe *postgresEvent) WatchEvent(ctx context.Context, db *pachsql.DB, template proto.Message) *watch.Event {
	if pe.err != nil {
		return &watch.Event{Err: pe.err, Type: watch.EventError}
	}
	if pe.eventType == watch.EventDelete {
		// Etcd doesn't return deleted row values - we could, but let's maintain parity
		return &watch.Event{Key: []byte(pe.key), Type: pe.eventType, Template: template}
	}
	if pe.protoData == nil && pe.storedID != "" {
		// The proto data was too large to fit in the payload, read it from a temporary location.
		if err := db.QueryRowContext(ctx, "select proto from collections.large_notifications where id = $1", pe.storedID).Scan(&pe.protoData); err != nil {
			// If the row is gone, this watcher is lagging too much, error it out
			return &watch.Event{Err: errors.Wrap(err, "failed to read notification data from large_notifications table, watcher latency may be too high"), Type: watch.EventError}
		}
	}
	return &watch.Event{
		Key:      []byte(pe.key),
		Value:    pe.protoData,
		Type:     pe.eventType,
		Template: template,
		Rev:      pe.time.Unix(),
	}
}

type notifierSet = map[string]Notifier

type postgresListener struct {
	dsn                   string
	pql                   *pq.Listener
	registerMu, channelMu sync.Mutex
	eg                    *errgroup.Group
	closed                bool

	channels map[string]notifierSet
}

func NewPostgresListener(dsn string) PostgresListener {
	// Apparently this is very important for lib/pq to work
	dsn = strings.ReplaceAll(dsn, "statement_cache_mode=describe", "")
	eg, _ := errgroup.WithContext(context.Background())

	l := &postgresListener{
		dsn:      dsn,
		eg:       eg,
		channels: make(map[string]notifierSet),
	}

	return l
}

func (l *postgresListener) Close() error {
	l.registerMu.Lock()
	defer l.registerMu.Unlock()

	l.closed = true

	if l.pql != nil {
		if err := l.pql.Close(); err != nil {
			return errors.EnsureStack(err)
		}
	}

	if err := l.eg.Wait(); err != nil {
		return errors.EnsureStack(err)
	}

	l.reset(errors.New("postgres listener has been closed"))
	return nil
}

// Lazily generate the pq.Listener because we don't want to trigger a race
// condition where an unused pq.Listener can't be closed properly. The mutex
// must be locked before calling this.
func (l *postgresListener) getPQL() *pq.Listener {
	if l.pql == nil {
		l.pql = pq.NewListener(l.dsn, minReconnectInterval, maxReconnectInterval, nil)
		l.eg.Go(func() error {
			return l.multiplex(l.pql.Notify)
		})
	}
	return l.pql
}

// reset will remove all watchers and unlisten from all channels - you must have
// the lock on the listener's register mutex before calling this.
func (l *postgresListener) reset(err error) {
	l.channelMu.Lock()
	for _, notifiers := range l.channels {
		for _, notifier := range notifiers {
			notifier.Error(err)
		}
	}
	l.channels = make(map[string]notifierSet)
	l.channelMu.Unlock()
	if !l.closed {
		l.getPQL().UnlistenAll() //nolint:errcheck // `reset` is only ever called in the case of an error
	}
}

func (l *postgresListener) Register(notifier Notifier) error {
	l.registerMu.Lock()
	defer l.registerMu.Unlock()

	if l.closed {
		return errors.New("postgres listener has been closed")
	}

	// Subscribe the notifier to the given channel.
	id := notifier.ID()
	channel := notifier.Channel()
	l.channelMu.Lock()
	defer l.channelMu.Unlock()
	notifiers, ok := l.channels[channel]
	if !ok {
		notifiers = make(notifierSet)
		l.channels[channel] = notifiers
		l.channelMu.Unlock()
		defer l.channelMu.Lock()
		if err := l.getPQL().Listen(channel); err != nil {
			// If an error occurs, error out all notifiers and reset the state of the
			// listener to prevent desyncs.
			err = errors.EnsureStack(err)
			l.reset(err)
			return err
		}
	}
	notifiers[id] = notifier

	return nil
}

func (l *postgresListener) Unregister(notifier Notifier) error {
	l.registerMu.Lock()
	defer l.registerMu.Unlock()

	// Remove the notifier from the given channels, unlisten if any are empty.
	id := notifier.ID()
	channel := notifier.Channel()
	l.channelMu.Lock()
	defer l.channelMu.Unlock()
	notifiers := l.channels[channel]
	if len(notifiers) > 0 {
		delete(notifiers, id)
		if len(notifiers) == 0 {
			delete(l.channels, channel)
			l.channelMu.Unlock()
			defer l.channelMu.Lock()
			if err := l.getPQL().Unlisten(channel); err != nil {
				// If an error occurs, error out all watches and reset the state of the
				// listener to prevent desyncs.
				err = errors.EnsureStack(err)
				l.reset(err)
				return err
			}
		}
	}
	return nil
}

func (l *postgresListener) multiplex(notifyChan chan *pq.Notification) error {
	for {
		notification, ok := <-notifyChan
		if !ok {
			return nil
		}
		if notification == nil {
			// A 'nil' notification means that the connection was lost - error out all
			// current watchers so they can rebuild state.
			l.registerMu.Lock()
			l.reset(errors.Errorf("lost connection to database"))
			l.registerMu.Unlock()
		} else {
			l.routeNotification(notification)
		}
	}
}

func (l *postgresListener) routeNotification(notification *pq.Notification) {
	l.channelMu.Lock()
	defer l.channelMu.Unlock()

	// Ignore any messages from channels we have no notifiers for.
	notifiers, ok := l.channels[notification.Channel]
	if ok {
		for _, notifier := range notifiers {
			notifier.Notify(notification)
		}
	}
}
