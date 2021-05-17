package collection_test

import (
	"context"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"golang.org/x/sync/errgroup"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil/random"
	"github.com/pachyderm/pachyderm/v2/src/internal/watch"
)

func nextEvent(c <-chan *watch.Event, timeout time.Duration) *watch.Event {
	select {
	case ev := <-c:
		return ev
	case <-time.After(timeout):
		return nil
	}
}

func requireEmptyChannel(t *testing.T, c <-chan *watch.Event) {
	select {
	case ev, ok := <-c:
		if ok {
			require.False(t, ok, "watcher had additional unread events at the end of the test, type: %v, key: %s", ev.Type, string(ev.Key))
		}
	default:
	}
}

type TestEvent struct {
	Type  watch.EventType
	Key   string
	Value *col.TestItem
}

// newTestEvent copies out the fields we want to check from an Event and
// unmarshals the protobuf since the serialization isn't deterministic so we
// can't just compare the bytes.
func newTestEvent(t *testing.T, ev *watch.Event) TestEvent {
	result := TestEvent{Type: ev.Type, Key: string(ev.Key)}
	if ev.Value != nil {
		result.Value = &col.TestItem{}
		require.NoError(t, proto.Unmarshal(ev.Value, result.Value))
	}
	return result
}

type WatchTester interface {
	Write(item *col.TestItem)             // writes the given row
	Delete(id string)                     // deletes a row with the given ID
	DeleteAll()                           // deletes all rows
	ExpectEvent(expected TestEvent)       // expects an event matching the given event
	ExpectEventSet(expected ...TestEvent) // expects a set of events matching the given events (in no given order)
	ExpectError(err error)                // expects an error that matches the given error type
	ExpectNoEvents()                      // expects no events to be received for a short period
}

type ChannelWatchTester struct {
	t       *testing.T
	writer  WriteCallback
	watcher watch.Watcher
}

func (tester *ChannelWatchTester) nextEvent(timeout time.Duration) *watch.Event {
	return nextEvent(tester.watcher.Watch(), timeout)
}

func (tester *ChannelWatchTester) Write(item *col.TestItem) {
	err := tester.writer(context.Background(), func(rw col.ReadWriteCollection) error {
		return putItem(item)(rw)
	})
	require.NoError(tester.t, err)
}

func (tester *ChannelWatchTester) Delete(id string) {
	err := tester.writer(context.Background(), func(rw col.ReadWriteCollection) error {
		return rw.Delete(id)
	})
	require.NoError(tester.t, err)
}

func (tester *ChannelWatchTester) DeleteAll() {
	err := tester.writer(context.Background(), func(rw col.ReadWriteCollection) error {
		return rw.DeleteAll()
	})
	require.NoError(tester.t, err)
}

func (tester *ChannelWatchTester) ExpectEvent(expected TestEvent) {
	tester.ExpectEventSet(expected)
}

func (tester *ChannelWatchTester) ExpectEventSet(expected ...TestEvent) {
	actual := []TestEvent{}
	for range expected {
		ev := tester.nextEvent(5 * time.Second)
		require.NotNil(tester.t, ev)
		actual = append(actual, newTestEvent(tester.t, ev))
	}

	// require.ElementsEqual doesn't do a deep equal, so implement set equality ourselves
	seen := map[int]struct{}{}
	for _, ex := range expected {
		found := false
		for i, ac := range actual {
			if _, ok := seen[i]; !ok {
				if require.EqualOrErr(ex, ac) == nil {
					seen[i] = struct{}{}
					found = true
				}
			}
		}
		require.True(tester.t, found, "Expected element not found in result set: %v\nactual: %v", ex, actual)
	}
}

func (tester *ChannelWatchTester) ExpectError(err error) {
	ev := tester.nextEvent(5 * time.Second)
	require.NotNil(tester.t, ev)
	require.Equal(tester.t, TestEvent{watch.EventError, "", nil}, newTestEvent(tester.t, ev))
	require.True(tester.t, errors.Is(ev.Err, err))
}

func (tester *ChannelWatchTester) ExpectNoEvents() {
	require.Nil(tester.t, tester.nextEvent(100*time.Millisecond))
}

func NewWatchTester(t *testing.T, writer WriteCallback, watcher watch.Watcher) WatchTester {
	t.Cleanup(func() {
		requireEmptyChannel(t, watcher.Watch())
	})
	return &ChannelWatchTester{t, writer, watcher}
}

type WatchShim struct {
	watchChan chan *watch.Event
}

func (shim *WatchShim) Watch() <-chan *watch.Event {
	return shim.watchChan
}

func (shim *WatchShim) Close() {
	close(shim.watchChan)
}

// NewWatchShim converts a callback-based watch into a channel-based Watcher. It
// spawns a goroutine that will run the callback-based watch and pass the events
// into a shim that fulfills the Watcher interface so we can reuse
// ChannelWatchTester.
func NewWatchShim(ctx context.Context, t *testing.T, doWatch func(context.Context, func(*watch.Event) error) error) watch.Watcher {
	watchChan := make(chan *watch.Event)

	done := false
	ctx, cancel := context.WithCancel(ctx)
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		// There is no way to test ErrBreak behavior unless we know the set of
		// expected events ahead of time, so WatchF-style tests need to test this
		// manually.
		if err := doWatch(ctx, func(ev *watch.Event) error {
			watchChan <- ev
			return nil
		}); err != nil {
			// Don't send an error if we were canceled in cleanup - there is no one listening
			if !done {
				watchChan <- &watch.Event{Type: watch.EventError, Err: err}
			}
		}
		return nil
	})

	watcher := &WatchShim{watchChan}

	t.Cleanup(func() {
		done = true
		cancel()
		require.NoError(t, eg.Wait())
		watcher.Close()
	})

	return watcher
}

// TODO: options: filter, initial sort

func watchTests(
	parent *testing.T,
	newCollection func(context.Context, *testing.T) (ReadCallback, WriteCallback),
) {
	testInterruptionPreemptive := func(t *testing.T, makeWatcher func() (watch.Watcher, error)) {
		watcher, err := makeWatcher()

		// Difference between postgres and etcd - etcd will error at the `Watch` call, postgres will error in the channel
		if err != nil {
			require.True(t, errors.Is(err, context.Canceled))
			return
		}
		defer watcher.Close()

		ev := nextEvent(watcher.Watch(), 5*time.Second)
		require.Equal(t, TestEvent{watch.EventError, "", nil}, newTestEvent(t, ev))
		require.YesError(t, ev.Err)
		require.True(t, errors.Is(ev.Err, context.Canceled))
		requireEmptyChannel(t, watcher.Watch())
	}

	watchAllTests := func(suite *testing.T, makeWatcher func(context.Context, *testing.T, ReadCallback) watch.Watcher) {
		suite.Run("InterruptionAfterInitial", func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			reader, writer := newCollection(context.Background(), t)
			row := makeProto(makeID(4))
			writer(context.Background(), putItem(row))

			tester := NewWatchTester(t, writer, makeWatcher(ctx, t, reader))
			tester.ExpectEvent(TestEvent{watch.EventPut, row.ID, row})
			tester.ExpectNoEvents()
			tester.Delete(row.ID)
			tester.ExpectEvent(TestEvent{watch.EventDelete, row.ID, nil})
			tester.ExpectNoEvents()
			cancel()
			tester.ExpectError(context.Canceled)
			tester.ExpectNoEvents()
		})

		suite.Run("Delete", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(context.Background(), t)
			row := makeProto(makeID(1))
			writer(context.Background(), putItem(row))

			tester := NewWatchTester(t, writer, makeWatcher(context.Background(), t, reader))
			tester.ExpectEvent(TestEvent{watch.EventPut, row.ID, row})
			tester.ExpectNoEvents()
			tester.Delete(row.ID)
			tester.ExpectEvent(TestEvent{watch.EventDelete, row.ID, nil})
			tester.ExpectNoEvents()
		})

		suite.Run("DeleteAll", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(context.Background(), t)
			rowA := makeProto(makeID(1))
			rowB := makeProto(makeID(2))
			writer(context.Background(), putItem(rowA))
			writer(context.Background(), putItem(rowB))

			tester := NewWatchTester(t, writer, makeWatcher(context.Background(), t, reader))
			tester.ExpectEvent(TestEvent{watch.EventPut, rowA.ID, rowA})
			tester.ExpectEvent(TestEvent{watch.EventPut, rowB.ID, rowB})
			tester.ExpectNoEvents()
			tester.DeleteAll()
			tester.ExpectEventSet(
				TestEvent{watch.EventDelete, rowA.ID, nil},
				TestEvent{watch.EventDelete, rowB.ID, nil},
			)
			tester.ExpectNoEvents()
		})

		suite.Run("Create", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(context.Background(), t)
			rowA := makeProto(makeID(1))
			rowB := makeProto(makeID(2))

			tester := NewWatchTester(t, writer, makeWatcher(context.Background(), t, reader))
			tester.ExpectNoEvents()
			tester.Write(rowA)
			tester.ExpectEvent(TestEvent{watch.EventPut, rowA.ID, rowA})
			tester.ExpectNoEvents()
			tester.Write(rowB)
			tester.ExpectEvent(TestEvent{watch.EventPut, rowB.ID, rowB})
			tester.ExpectNoEvents()
		})

		suite.Run("Overwrite", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(context.Background(), t)
			row := makeProto(makeID(1), originalValue)
			writer(context.Background(), putItem(row))

			tester := NewWatchTester(t, writer, makeWatcher(context.Background(), t, reader))
			tester.ExpectEvent(TestEvent{watch.EventPut, row.ID, row})
			tester.ExpectNoEvents()

			row.Value = changedValue
			tester.Write(row)
			tester.ExpectEvent(TestEvent{watch.EventPut, row.ID, row})
			tester.ExpectNoEvents()
		})
	}

	parent.Run("Watch", func(suite *testing.T) {
		suite.Parallel()

		suite.Run("InterruptionPreemptive", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(context.Background(), t)
			watchRead := reader(canceledContext())
			writer(context.Background(), putItem(makeProto(makeID(3))))
			testInterruptionPreemptive(t, func() (watch.Watcher, error) { return watchRead.Watch() })
		})

		watchAllTests(suite, func(ctx context.Context, t *testing.T, reader ReadCallback) watch.Watcher {
			watcher, err := reader(ctx).Watch()
			require.NoError(t, err)
			t.Cleanup(watcher.Close)
			return watcher
		})
	})

	parent.Run("WatchF", func(suite *testing.T) {
		suite.Parallel()

		suite.Run("InterruptionPreemptive", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(context.Background(), t)
			watchRead := reader(canceledContext())
			writer(context.Background(), putItem(makeProto(makeID(1))))

			events := []*watch.Event{}
			err := watchRead.WatchF(func(ev *watch.Event) error {
				return errors.New("should have been canceled before receiving event")
			})
			require.YesError(t, err)
			require.True(t, errors.Is(err, context.Canceled))
			require.Equal(t, 0, len(events))
		})

		// TODO: consolidate ErrBreak tests
		suite.Run("ErrBreak", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(context.Background(), t)
			watchRead := reader(context.Background())
			rowA := makeProto(makeID(1))
			rowB := makeProto(makeID(2))
			writer(context.Background(), putItem(rowA))
			writer(context.Background(), putItem(rowB))

			events := []TestEvent{}
			err := watchRead.WatchF(func(ev *watch.Event) error {
				events = append(events, newTestEvent(t, ev))
				return errutil.ErrBreak
			})
			require.NoError(t, err)
			require.Equal(t, []TestEvent{{watch.EventPut, rowA.ID, rowA}}, events)
		})

		watchAllTests(suite, func(ctx context.Context, t *testing.T, reader ReadCallback) watch.Watcher {
			return NewWatchShim(ctx, t, func(ctx context.Context, cb func(*watch.Event) error) error {
				return reader(ctx).WatchF(cb)
			})
		})
	})

	watchOneTests := func(suite *testing.T, makeWatcher func(context.Context, *testing.T, ReadCallback, string) watch.Watcher) {
		suite.Run("LargeRow", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(context.Background(), t)
			row := makeProto(makeID(7))
			row.Data = random.String(10000) // postgres payload limit should be 8KB

			tester := NewWatchTester(t, writer, makeWatcher(context.Background(), t, reader, row.ID))
			tester.ExpectNoEvents()
			tester.Write(row)
			tester.ExpectEvent(TestEvent{watch.EventPut, row.ID, row})
			tester.ExpectNoEvents()
			tester.Delete(row.ID)
			tester.ExpectEvent(TestEvent{watch.EventDelete, row.ID, nil})
			tester.ExpectNoEvents()
		})

		suite.Run("InterruptionAfterInitial", func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			reader, writer := newCollection(context.Background(), t)
			row := makeProto(makeID(4))
			writer(context.Background(), putItem(row))

			tester := NewWatchTester(t, writer, makeWatcher(ctx, t, reader, row.ID))
			tester.ExpectEvent(TestEvent{watch.EventPut, row.ID, row})
			tester.ExpectNoEvents()
			tester.Delete(row.ID)
			tester.ExpectEvent(TestEvent{watch.EventDelete, row.ID, nil})
			tester.ExpectNoEvents()
			cancel()
			tester.ExpectError(context.Canceled)
			tester.ExpectNoEvents()
		})

		suite.Run("Delete", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(context.Background(), t)
			rowA := makeProto(makeID(1))
			rowB := makeProto(makeID(2))
			writer(context.Background(), putItem(rowA))
			writer(context.Background(), putItem(rowB))

			tester := NewWatchTester(t, writer, makeWatcher(context.Background(), t, reader, rowA.ID))
			tester.ExpectEvent(TestEvent{watch.EventPut, rowA.ID, rowA})
			tester.ExpectNoEvents()
			tester.Delete(rowA.ID)
			tester.ExpectEvent(TestEvent{watch.EventDelete, rowA.ID, nil})
			tester.ExpectNoEvents()
			tester.Delete(rowB.ID)
			tester.ExpectNoEvents()
		})

		suite.Run("DeleteAll", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(context.Background(), t)
			rowA := makeProto(makeID(1))
			rowB := makeProto(makeID(2))
			writer(context.Background(), putItem(rowA))
			writer(context.Background(), putItem(rowB))

			tester := NewWatchTester(t, writer, makeWatcher(context.Background(), t, reader, rowA.ID))
			tester.ExpectEvent(TestEvent{watch.EventPut, rowA.ID, rowA})
			tester.ExpectNoEvents()
			tester.DeleteAll()
			tester.ExpectEvent(TestEvent{watch.EventDelete, rowA.ID, nil})
			tester.ExpectNoEvents()
		})

		suite.Run("Create", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(context.Background(), t)
			rowA := makeProto(makeID(1))
			rowB := makeProto(makeID(2))

			tester := NewWatchTester(t, writer, makeWatcher(context.Background(), t, reader, rowA.ID))
			tester.ExpectNoEvents()
			tester.Write(rowA)
			tester.ExpectEvent(TestEvent{watch.EventPut, rowA.ID, rowA})
			tester.ExpectNoEvents()
			tester.Write(rowB)
			tester.ExpectNoEvents()
		})

		suite.Run("Overwrite", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(context.Background(), t)
			rowA := makeProto(makeID(1), originalValue)
			rowB := makeProto(makeID(2))
			writer(context.Background(), putItem(rowA))
			writer(context.Background(), putItem(rowB))

			tester := NewWatchTester(t, writer, makeWatcher(context.Background(), t, reader, rowA.ID))
			tester.ExpectEvent(TestEvent{watch.EventPut, rowA.ID, rowA})
			tester.ExpectNoEvents()

			rowA.Value = changedValue
			tester.Write(rowA)
			tester.ExpectEvent(TestEvent{watch.EventPut, rowA.ID, rowA})
			tester.ExpectNoEvents()

			rowB.Value = changedValue
			tester.Write(rowB)
			tester.ExpectNoEvents()
		})
	}

	parent.Run("WatchOne", func(suite *testing.T) {
		suite.Parallel()

		suite.Run("InterruptionPreemptive", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(context.Background(), t)
			watchRead := reader(canceledContext())
			row := makeProto(makeID(1))
			writer(context.Background(), putItem(row))
			testInterruptionPreemptive(t, func() (watch.Watcher, error) { return watchRead.WatchOne(row.ID) })
		})

		watchOneTests(suite, func(ctx context.Context, t *testing.T, reader ReadCallback, key string) watch.Watcher {
			watcher, err := reader(ctx).WatchOne(key)
			require.NoError(t, err)
			return watcher
		})
	})

	parent.Run("WatchOneF", func(suite *testing.T) {
		suite.Parallel()

		suite.Run("InterruptionPreemptive", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(context.Background(), t)
			watchRead := reader(canceledContext())
			row := makeProto(makeID(1))
			writer(context.Background(), putItem(row))

			events := []*watch.Event{}
			err := watchRead.WatchOneF(row.ID, func(ev *watch.Event) error {
				return errors.New("should have been canceled before receiving event")
			})
			require.YesError(t, err)
			require.True(t, errors.Is(err, context.Canceled))
			require.Equal(t, 0, len(events))
		})

		suite.Run("ErrBreak", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(context.Background(), t)
			watchRead := reader(context.Background())
			rowA := makeProto(makeID(1))
			rowB := makeProto(makeID(2))
			writer(context.Background(), putItem(rowA))
			writer(context.Background(), putItem(rowB))

			events := []TestEvent{}
			err := watchRead.WatchOneF(rowA.ID, func(ev *watch.Event) error {
				events = append(events, newTestEvent(t, ev))
				return errutil.ErrBreak
			})
			require.NoError(t, err)
			require.Equal(t, []TestEvent{{watch.EventPut, rowA.ID, rowA}}, events)
		})

		watchOneTests(suite, func(ctx context.Context, t *testing.T, reader ReadCallback, key string) watch.Watcher {
			return NewWatchShim(ctx, t, func(ctx context.Context, cb func(ev *watch.Event) error) error {
				return reader(ctx).WatchOneF(key, cb)
			})
		})
	})

	watchByIndexTests := func(suite *testing.T, makeWatcher func(context.Context, *testing.T, ReadCallback, string) watch.Watcher) {
		suite.Run("InterruptionAfterInitial", func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			reader, writer := newCollection(context.Background(), t)
			row := makeProto(makeID(4), originalValue)
			writer(context.Background(), putItem(row))

			tester := NewWatchTester(t, writer, makeWatcher(ctx, t, reader, originalValue))
			tester.ExpectEvent(TestEvent{watch.EventPut, row.ID, row})
			tester.ExpectNoEvents()
			tester.Delete(row.ID)
			tester.ExpectEvent(TestEvent{watch.EventDelete, row.ID, nil})
			tester.ExpectNoEvents()
			cancel()
			tester.ExpectError(context.Canceled)
			tester.ExpectNoEvents()
		})

		suite.Run("Create", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(context.Background(), t)
			rowA := makeProto(makeID(1), originalValue)
			rowB := makeProto(makeID(2), changedValue)

			tester := NewWatchTester(t, writer, makeWatcher(context.Background(), t, reader, originalValue))
			tester.ExpectNoEvents()
			tester.Write(rowA)
			tester.ExpectEvent(TestEvent{watch.EventPut, rowA.ID, rowA})
			tester.ExpectNoEvents()
			tester.Write(rowB)
			tester.ExpectNoEvents()
		})

		suite.Run("Overwrite", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(context.Background(), t)
			rowA := makeProto(makeID(1), originalValue)
			rowB := makeProto(makeID(2), changedValue)
			writer(context.Background(), putItem(rowA))
			writer(context.Background(), putItem(rowB))

			tester := NewWatchTester(t, writer, makeWatcher(context.Background(), t, reader, originalValue))
			tester.ExpectEvent(TestEvent{watch.EventPut, rowA.ID, rowA})
			tester.ExpectNoEvents()
			tester.Write(rowA)
			tester.ExpectEvent(TestEvent{watch.EventPut, rowA.ID, rowA})
			tester.ExpectNoEvents()
			tester.Write(rowB)
			tester.ExpectNoEvents()
		})

		suite.Run("Delete", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(context.Background(), t)
			rowA := makeProto(makeID(1), originalValue)
			rowB := makeProto(makeID(2), changedValue)
			writer(context.Background(), putItem(rowA))
			writer(context.Background(), putItem(rowB))

			tester := NewWatchTester(t, writer, makeWatcher(context.Background(), t, reader, originalValue))
			tester.ExpectEvent(TestEvent{watch.EventPut, rowA.ID, rowA})
			tester.ExpectNoEvents()
			tester.Delete(rowA.ID)
			tester.ExpectEvent(TestEvent{watch.EventDelete, rowA.ID, nil})
			tester.ExpectNoEvents()
			tester.Delete(rowB.ID)
			tester.ExpectNoEvents()
		})

		suite.Run("DeleteAll", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(context.Background(), t)
			rowA := makeProto(makeID(1), originalValue)
			rowB := makeProto(makeID(2), changedValue)
			writer(context.Background(), putItem(rowA))
			writer(context.Background(), putItem(rowB))

			tester := NewWatchTester(t, writer, makeWatcher(context.Background(), t, reader, originalValue))
			tester.ExpectEvent(TestEvent{watch.EventPut, rowA.ID, rowA})
			tester.ExpectNoEvents()
			tester.DeleteAll()
			tester.ExpectEvent(TestEvent{watch.EventDelete, rowA.ID, nil})
			tester.ExpectNoEvents()
		})

		suite.Run("MoveIntoSet", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(context.Background(), t)
			rowA := makeProto(makeID(1), originalValue)
			rowB := makeProto(makeID(2), changedValue)
			writer(context.Background(), putItem(rowA))
			writer(context.Background(), putItem(rowB))

			tester := NewWatchTester(t, writer, makeWatcher(context.Background(), t, reader, originalValue))
			tester.ExpectEvent(TestEvent{watch.EventPut, rowA.ID, rowA})
			tester.ExpectNoEvents()
			rowB.Value = originalValue
			tester.Write(rowB)
			tester.ExpectEvent(TestEvent{watch.EventPut, rowB.ID, rowB})
			tester.ExpectNoEvents()
		})

		suite.Run("MoveOutOfSet", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(context.Background(), t)
			rowA := makeProto(makeID(1), originalValue)
			rowB := makeProto(makeID(2), changedValue)
			writer(context.Background(), putItem(rowA))
			writer(context.Background(), putItem(rowB))

			tester := NewWatchTester(t, writer, makeWatcher(context.Background(), t, reader, originalValue))
			tester.ExpectEvent(TestEvent{watch.EventPut, rowA.ID, rowA})
			tester.ExpectNoEvents()
			rowA.Value = changedValue
			tester.Write(rowA)
			tester.ExpectEvent(TestEvent{watch.EventDelete, rowA.ID, nil})
			tester.ExpectNoEvents()
		})
	}

	parent.Run("WatchByIndex", func(suite *testing.T) {
		suite.Parallel()

		suite.Run("InterruptionPreemptive", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(context.Background(), t)
			watchRead := reader(canceledContext())
			row := makeProto(makeID(1), originalValue)
			writer(context.Background(), putItem(row))
			testInterruptionPreemptive(t, func() (watch.Watcher, error) { return watchRead.WatchByIndex(TestSecondaryIndex, originalValue) })
		})

		watchByIndexTests(suite, func(ctx context.Context, t *testing.T, reader ReadCallback, value string) watch.Watcher {
			watcher, err := reader(ctx).WatchByIndex(TestSecondaryIndex, value)
			require.NoError(t, err)
			return watcher
		})
	})

	parent.Run("WatchByIndexF", func(suite *testing.T) {
		suite.Parallel()

		suite.Run("InterruptionPreemptive", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(context.Background(), t)
			watchRead := reader(canceledContext())
			row := makeProto(makeID(1), originalValue)
			writer(context.Background(), putItem(row))

			events := []*watch.Event{}
			err := watchRead.WatchByIndexF(TestSecondaryIndex, originalValue, func(ev *watch.Event) error {
				return errors.New("should have been canceled before receiving event")
			})
			require.YesError(t, err)
			require.True(t, errors.Is(err, context.Canceled))
			require.Equal(t, 0, len(events))
		})

		suite.Run("ErrBreak", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(context.Background(), t)
			watchRead := reader(context.Background())
			rowA := makeProto(makeID(1), originalValue)
			rowB := makeProto(makeID(2), originalValue)
			writer(context.Background(), putItem(rowA))
			writer(context.Background(), putItem(rowB))

			events := []TestEvent{}
			err := watchRead.WatchByIndexF(TestSecondaryIndex, originalValue, func(ev *watch.Event) error {
				events = append(events, newTestEvent(t, ev))
				return errutil.ErrBreak
			})
			require.NoError(t, err)
			require.Equal(t, []TestEvent{{watch.EventPut, rowA.ID, rowA}}, events)
		})

		watchByIndexTests(suite, func(ctx context.Context, t *testing.T, reader ReadCallback, value string) watch.Watcher {
			watcher, err := reader(ctx).WatchByIndex(TestSecondaryIndex, value)
			require.NoError(t, err)
			return watcher
		})
	})
}
