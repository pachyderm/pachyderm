package collection_test

import (
	"context"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
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
	Write(ctx context.Context, item *col.TestItem) // writes the given row
	Delete(ctx context.Context, id string)         // deletes a row with the given ID
	DeleteAll(ctx context.Context)                 // deletes all rows
	ExpectEvent(expected TestEvent)                // expects an event matching the given event
	ExpectEventSet(expected ...TestEvent)          // expects a set of events matching the given events (in no given order)
	ExpectError(err error)                         // expects an error that matches the given error type
	ExpectNoEvents()                               // expects no events to be received for a short period
}

type ChannelWatchTester struct {
	t       *testing.T
	writer  WriteCallback
	watcher watch.Watcher
}

func (tester *ChannelWatchTester) nextEvent(timeout time.Duration) *watch.Event {
	return nextEvent(tester.watcher.Watch(), timeout)
}

func (tester *ChannelWatchTester) Write(ctx context.Context, item *col.TestItem) {
	tester.t.Helper()
	err := tester.writer(ctx, func(rw col.ReadWriteCollection) error {
		return putItem(item)(rw)
	})
	require.NoError(tester.t, err)
}

func (tester *ChannelWatchTester) Delete(ctx context.Context, id string) {
	tester.t.Helper()
	err := tester.writer(ctx, func(rw col.ReadWriteCollection) error {
		return errors.EnsureStack(rw.Delete(id))
	})
	require.NoError(tester.t, err)
}

func (tester *ChannelWatchTester) DeleteAll(ctx context.Context) {
	tester.t.Helper()
	err := tester.writer(ctx, func(rw col.ReadWriteCollection) error {
		return errors.EnsureStack(rw.DeleteAll())
	})
	require.NoError(tester.t, err)
}

func (tester *ChannelWatchTester) ExpectEvent(expected TestEvent) {
	tester.ExpectEventSet(expected)
}

func (tester *ChannelWatchTester) ExpectEventSet(expected ...TestEvent) {
	tester.t.Helper()
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
	tester.t.Helper()
	ev := tester.nextEvent(5 * time.Second)
	require.NotNil(tester.t, ev)
	require.Equal(tester.t, TestEvent{watch.EventError, "", nil}, newTestEvent(tester.t, ev))
	require.True(tester.t, errors.Is(ev.Err, err))
}

func (tester *ChannelWatchTester) ExpectNoEvents() {
	tester.t.Helper()
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
	ctx, cancel := pctx.WithCancel(ctx)
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
	rctx context.Context,
	parent *testing.T,
	newCollection func(context.Context, *testing.T, ...bool) (ReadCallback, WriteCallback),
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
		require.NotNil(t, ev)
		require.Equal(t, TestEvent{watch.EventError, "", nil}, newTestEvent(t, ev))
		require.YesError(t, ev.Err)
		require.True(t, errors.Is(ev.Err, context.Canceled))
		requireEmptyChannel(t, watcher.Watch())
	}

	watchAllTests := func(suite *testing.T, makeWatcher func(context.Context, *testing.T, ReadCallback) watch.Watcher) {
		suite.Run("InterruptionAfterInitial", func(t *testing.T) {
			t.Parallel()
			ctx, cancel := pctx.WithCancel(rctx)
			defer cancel()
			reader, writer := newCollection(rctx, t)
			row := makeProto(makeID(4))
			require.NoError(t, writer(rctx, putItem(row)))

			tester := NewWatchTester(t, writer, makeWatcher(ctx, t, reader))
			tester.ExpectEvent(TestEvent{watch.EventPut, row.Id, row})
			tester.ExpectNoEvents()
			tester.Delete(rctx, row.Id)
			tester.ExpectEvent(TestEvent{watch.EventDelete, row.Id, nil})
			tester.ExpectNoEvents()
			cancel()
			tester.ExpectError(context.Canceled)
			tester.ExpectNoEvents()
		})

		suite.Run("InterruptionDuringBackfill", func(t *testing.T) {
			t.Parallel()
			ctx, cancel := pctx.WithCancel(rctx)
			defer cancel()
			reader, writer := newCollection(rctx, t)
			for i := 0; i < 10; i++ {
				require.NoError(t, writer(rctx, putItem(makeProto(makeID(i)))))
			}

			watcher := makeWatcher(ctx, t, reader)
			tester := NewWatchTester(t, writer, watcher)
			cancel()
			var canceled bool
			for i := 0; i < 11; i++ {
				// Consume events until we receive the cancellation event.  (Some
				// events may have arrived before cancellation.)
				ev := nextEvent(watcher.Watch(), time.Second)
				require.NotNil(t, ev, "event %v should not be nil, since we haven't been canceled yet", i)
				if ev.Err == nil {
					continue
				}
				require.ErrorIs(t, ev.Err, context.Canceled)
				canceled = true
				break
			}
			require.True(t, canceled, "we should have gotten the cancellation event in the loop")
			tester.ExpectNoEvents()
		})

		suite.Run("Delete", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(rctx, t)
			row := makeProto(makeID(1))
			require.NoError(t, writer(rctx, putItem(row)))

			tester := NewWatchTester(t, writer, makeWatcher(rctx, t, reader))
			tester.ExpectEvent(TestEvent{watch.EventPut, row.Id, row})
			tester.ExpectNoEvents()
			tester.Delete(rctx, row.Id)
			tester.ExpectEvent(TestEvent{watch.EventDelete, row.Id, nil})
			tester.ExpectNoEvents()
		})

		suite.Run("DeleteAll", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(rctx, t)
			rowA := makeProto(makeID(1))
			rowB := makeProto(makeID(2))
			require.NoError(t, writer(rctx, putItem(rowA)))
			require.NoError(t, writer(rctx, putItem(rowB)))

			tester := NewWatchTester(t, writer, makeWatcher(rctx, t, reader))
			tester.ExpectEvent(TestEvent{watch.EventPut, rowA.Id, rowA})
			tester.ExpectEvent(TestEvent{watch.EventPut, rowB.Id, rowB})
			tester.ExpectNoEvents()
			tester.DeleteAll(rctx)
			tester.ExpectEventSet(
				TestEvent{watch.EventDelete, rowA.Id, nil},
				TestEvent{watch.EventDelete, rowB.Id, nil},
			)
			tester.ExpectNoEvents()
		})

		suite.Run("Create", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(rctx, t)
			rowA := makeProto(makeID(1))
			rowB := makeProto(makeID(2))

			tester := NewWatchTester(t, writer, makeWatcher(rctx, t, reader))
			tester.ExpectNoEvents()
			tester.Write(rctx, rowA)
			tester.ExpectEvent(TestEvent{watch.EventPut, rowA.Id, rowA})
			tester.ExpectNoEvents()
			tester.Write(rctx, rowB)
			tester.ExpectEvent(TestEvent{watch.EventPut, rowB.Id, rowB})
			tester.ExpectNoEvents()
		})

		suite.Run("Overwrite", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(rctx, t)
			row := makeProto(makeID(1), originalValue)
			require.NoError(t, writer(rctx, putItem(row)))

			tester := NewWatchTester(t, writer, makeWatcher(rctx, t, reader))
			tester.ExpectEvent(TestEvent{watch.EventPut, row.Id, row})
			tester.ExpectNoEvents()

			row.Value = changedValue
			tester.Write(rctx, row)
			tester.ExpectEvent(TestEvent{watch.EventPut, row.Id, row})
			tester.ExpectNoEvents()
		})
	}

	parent.Run("Watch", func(suite *testing.T) {
		suite.Parallel()

		suite.Run("InterruptionPreemptive", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(rctx, t)
			watchRead := reader(canceledContext())
			require.NoError(t, writer(rctx, putItem(makeProto(makeID(3)))))
			testInterruptionPreemptive(t, func() (watch.Watcher, error) {
				res, err := watchRead.Watch()
				return res, errors.EnsureStack(err)
			})
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
			reader, writer := newCollection(rctx, t)
			watchRead := reader(canceledContext())
			require.NoError(t, writer(rctx, putItem(makeProto(makeID(1)))))

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
			reader, writer := newCollection(rctx, t)
			watchRead := reader(rctx)
			rowA := makeProto(makeID(1))
			rowB := makeProto(makeID(2))
			require.NoError(t, writer(rctx, putItem(rowA)))
			require.NoError(t, writer(rctx, putItem(rowB)))

			events := []TestEvent{}
			err := watchRead.WatchF(func(ev *watch.Event) error {
				events = append(events, newTestEvent(t, ev))
				return errutil.ErrBreak
			})
			require.NoError(t, err)
			require.Equal(t, []TestEvent{{watch.EventPut, rowA.Id, rowA}}, events)
		})

		watchAllTests(suite, func(ctx context.Context, t *testing.T, reader ReadCallback) watch.Watcher {
			return NewWatchShim(ctx, t, func(ctx context.Context, cb func(*watch.Event) error) error {
				return errors.EnsureStack(reader(ctx).WatchF(cb))
			})
		})
	})

	watchOneTests := func(suite *testing.T, makeWatcher func(context.Context, *testing.T, ReadCallback, string) watch.Watcher) {
		suite.Run("LargeRow", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(rctx, t)
			row := makeProto(makeID(7))

			tester := NewWatchTester(t, writer, makeWatcher(rctx, t, reader, row.Id))
			tester.ExpectNoEvents()
			// postgres payload limit is 8KB, but it is base64-encoded, so a data
			// length of 3/4 * 8000 should be sufficient.
			for i := 5800; i < 6000; i += 5 {
				row.Data = random.String(i)
				tester.Write(rctx, row)
				tester.ExpectEvent(TestEvent{watch.EventPut, row.Id, row})
				tester.ExpectNoEvents()
			}
			tester.Delete(rctx, row.Id)
			tester.ExpectEvent(TestEvent{watch.EventDelete, row.Id, nil})
			tester.ExpectNoEvents()
		})

		suite.Run("InterruptionAfterInitial", func(t *testing.T) {
			t.Parallel()
			ctx, cancel := pctx.WithCancel(rctx)
			defer cancel()
			reader, writer := newCollection(rctx, t)
			row := makeProto(makeID(4))
			require.NoError(t, writer(rctx, putItem(row)))

			tester := NewWatchTester(t, writer, makeWatcher(ctx, t, reader, row.Id))
			tester.ExpectEvent(TestEvent{watch.EventPut, row.Id, row})
			tester.ExpectNoEvents()
			tester.Delete(rctx, row.Id)
			tester.ExpectEvent(TestEvent{watch.EventDelete, row.Id, nil})
			tester.ExpectNoEvents()
			cancel()
			tester.ExpectError(context.Canceled)
			tester.ExpectNoEvents()
		})

		suite.Run("Delete", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(rctx, t)
			rowA := makeProto(makeID(1))
			rowB := makeProto(makeID(2))
			require.NoError(t, writer(rctx, putItem(rowA)))
			require.NoError(t, writer(rctx, putItem(rowB)))

			tester := NewWatchTester(t, writer, makeWatcher(rctx, t, reader, rowA.Id))
			tester.ExpectEvent(TestEvent{watch.EventPut, rowA.Id, rowA})
			tester.ExpectNoEvents()
			tester.Delete(rctx, rowA.Id)
			tester.ExpectEvent(TestEvent{watch.EventDelete, rowA.Id, nil})
			tester.ExpectNoEvents()
			tester.Delete(rctx, rowB.Id)
			tester.ExpectNoEvents()
		})

		suite.Run("DeleteAll", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(rctx, t)
			rowA := makeProto(makeID(1))
			rowB := makeProto(makeID(2))
			require.NoError(t, writer(rctx, putItem(rowA)))
			require.NoError(t, writer(rctx, putItem(rowB)))

			tester := NewWatchTester(t, writer, makeWatcher(rctx, t, reader, rowA.Id))
			tester.ExpectEvent(TestEvent{watch.EventPut, rowA.Id, rowA})
			tester.ExpectNoEvents()
			tester.DeleteAll(rctx)
			tester.ExpectEvent(TestEvent{watch.EventDelete, rowA.Id, nil})
			tester.ExpectNoEvents()
		})

		suite.Run("Create", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(rctx, t)
			rowA := makeProto(makeID(1))
			rowB := makeProto(makeID(2))

			tester := NewWatchTester(t, writer, makeWatcher(rctx, t, reader, rowA.Id))
			tester.ExpectNoEvents()
			tester.Write(rctx, rowA)
			tester.ExpectEvent(TestEvent{watch.EventPut, rowA.Id, rowA})
			tester.ExpectNoEvents()
			tester.Write(rctx, rowB)
			tester.ExpectNoEvents()
		})

		suite.Run("Overwrite", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(rctx, t)
			rowA := makeProto(makeID(1), originalValue)
			rowB := makeProto(makeID(2))
			require.NoError(t, writer(rctx, putItem(rowA)))
			require.NoError(t, writer(rctx, putItem(rowB)))

			tester := NewWatchTester(t, writer, makeWatcher(rctx, t, reader, rowA.Id))
			tester.ExpectEvent(TestEvent{watch.EventPut, rowA.Id, rowA})
			tester.ExpectNoEvents()

			rowA.Value = changedValue
			tester.Write(rctx, rowA)
			tester.ExpectEvent(TestEvent{watch.EventPut, rowA.Id, rowA})
			tester.ExpectNoEvents()

			rowB.Value = changedValue
			tester.Write(rctx, rowB)
			tester.ExpectNoEvents()
		})
	}

	parent.Run("WatchOne", func(suite *testing.T) {
		suite.Parallel()

		suite.Run("InterruptionPreemptive", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(rctx, t)
			watchRead := reader(canceledContext())
			row := makeProto(makeID(1))
			require.NoError(t, writer(rctx, putItem(row)))
			testInterruptionPreemptive(t, func() (watch.Watcher, error) {
				res, err := watchRead.WatchOne(row.Id)
				return res, errors.EnsureStack(err)
			})
		})

		watchOneTests(suite, func(ctx context.Context, t *testing.T, reader ReadCallback, key string) watch.Watcher {
			watcher, err := reader(ctx).WatchOne(key)
			require.NoError(t, err)
			t.Cleanup(watcher.Close)
			return watcher
		})
	})

	parent.Run("WatchOneF", func(suite *testing.T) {
		suite.Parallel()

		suite.Run("InterruptionPreemptive", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(rctx, t)
			watchRead := reader(canceledContext())
			row := makeProto(makeID(1))
			require.NoError(t, writer(rctx, putItem(row)))

			events := []*watch.Event{}
			err := watchRead.WatchOneF(row.Id, func(ev *watch.Event) error {
				return errors.New("should have been canceled before receiving event")
			})
			require.YesError(t, err)
			require.True(t, errors.Is(err, context.Canceled))
			require.Equal(t, 0, len(events))
		})

		suite.Run("ErrBreak", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(rctx, t)
			watchRead := reader(rctx)
			rowA := makeProto(makeID(1))
			rowB := makeProto(makeID(2))
			require.NoError(t, writer(rctx, putItem(rowA)))
			require.NoError(t, writer(rctx, putItem(rowB)))

			events := []TestEvent{}
			err := watchRead.WatchOneF(rowA.Id, func(ev *watch.Event) error {
				events = append(events, newTestEvent(t, ev))
				return errutil.ErrBreak
			})
			require.NoError(t, err)
			require.Equal(t, []TestEvent{{watch.EventPut, rowA.Id, rowA}}, events)
		})

		watchOneTests(suite, func(ctx context.Context, t *testing.T, reader ReadCallback, key string) watch.Watcher {
			return NewWatchShim(ctx, t, func(ctx context.Context, cb func(ev *watch.Event) error) error {
				return errors.EnsureStack(reader(ctx).WatchOneF(key, cb))
			})
		})
	})

	watchByIndexTests := func(suite *testing.T, makeWatcher func(context.Context, *testing.T, ReadCallback, string) watch.Watcher) {
		suite.Run("InterruptionAfterInitial", func(t *testing.T) {
			t.Parallel()
			ctx, cancel := pctx.WithCancel(rctx)
			defer cancel()
			reader, writer := newCollection(rctx, t)
			row := makeProto(makeID(4), originalValue)
			require.NoError(t, writer(rctx, putItem(row)))

			tester := NewWatchTester(t, writer, makeWatcher(ctx, t, reader, originalValue))
			tester.ExpectEvent(TestEvent{watch.EventPut, row.Id, row})
			tester.ExpectNoEvents()
			tester.Delete(rctx, row.Id)
			tester.ExpectEvent(TestEvent{watch.EventDelete, row.Id, nil})
			tester.ExpectNoEvents()
			cancel()
			tester.ExpectError(context.Canceled)
			tester.ExpectNoEvents()
		})

		suite.Run("Create", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(rctx, t)
			rowA := makeProto(makeID(1), originalValue)
			rowB := makeProto(makeID(2), changedValue)

			tester := NewWatchTester(t, writer, makeWatcher(rctx, t, reader, originalValue))
			tester.ExpectNoEvents()
			tester.Write(rctx, rowA)
			tester.ExpectEvent(TestEvent{watch.EventPut, rowA.Id, rowA})
			tester.ExpectNoEvents()
			tester.Write(rctx, rowB)
			tester.ExpectNoEvents()
		})

		suite.Run("Overwrite", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(rctx, t)
			rowA := makeProto(makeID(1), originalValue)
			rowB := makeProto(makeID(2), changedValue)
			require.NoError(t, writer(rctx, putItem(rowA)))
			require.NoError(t, writer(rctx, putItem(rowB)))

			tester := NewWatchTester(t, writer, makeWatcher(rctx, t, reader, originalValue))
			tester.ExpectEvent(TestEvent{watch.EventPut, rowA.Id, rowA})
			tester.ExpectNoEvents()
			tester.Write(rctx, rowA)
			tester.ExpectEvent(TestEvent{watch.EventPut, rowA.Id, rowA})
			tester.ExpectNoEvents()
			tester.Write(rctx, rowB)
			tester.ExpectNoEvents()
		})

		suite.Run("Delete", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(rctx, t)
			rowA := makeProto(makeID(1), originalValue)
			rowB := makeProto(makeID(2), changedValue)
			require.NoError(t, writer(rctx, putItem(rowA)))
			require.NoError(t, writer(rctx, putItem(rowB)))

			tester := NewWatchTester(t, writer, makeWatcher(rctx, t, reader, originalValue))
			tester.ExpectEvent(TestEvent{watch.EventPut, rowA.Id, rowA})
			tester.ExpectNoEvents()
			tester.Delete(rctx, rowA.Id)
			tester.ExpectEvent(TestEvent{watch.EventDelete, rowA.Id, nil})
			tester.ExpectNoEvents()
			tester.Delete(rctx, rowB.Id)
			tester.ExpectNoEvents()
		})

		suite.Run("DeleteAll", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(rctx, t)
			rowA := makeProto(makeID(1), originalValue)
			rowB := makeProto(makeID(2), changedValue)
			require.NoError(t, writer(rctx, putItem(rowA)))
			require.NoError(t, writer(rctx, putItem(rowB)))

			tester := NewWatchTester(t, writer, makeWatcher(rctx, t, reader, originalValue))
			tester.ExpectEvent(TestEvent{watch.EventPut, rowA.Id, rowA})
			tester.ExpectNoEvents()
			tester.DeleteAll(rctx)
			tester.ExpectEvent(TestEvent{watch.EventDelete, rowA.Id, nil})
			tester.ExpectNoEvents()
		})

		suite.Run("MoveIntoSet", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(rctx, t)
			rowA := makeProto(makeID(1), originalValue)
			rowB := makeProto(makeID(2), changedValue)
			require.NoError(t, writer(rctx, putItem(rowA)))
			require.NoError(t, writer(rctx, putItem(rowB)))

			tester := NewWatchTester(t, writer, makeWatcher(rctx, t, reader, originalValue))
			tester.ExpectEvent(TestEvent{watch.EventPut, rowA.Id, rowA})
			tester.ExpectNoEvents()
			rowB.Value = originalValue
			tester.Write(rctx, rowB)
			tester.ExpectEvent(TestEvent{watch.EventPut, rowB.Id, rowB})
			tester.ExpectNoEvents()
		})

		suite.Run("MoveOutOfSet", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(rctx, t)
			rowA := makeProto(makeID(1), originalValue)
			rowB := makeProto(makeID(2), changedValue)
			require.NoError(t, writer(rctx, putItem(rowA)))
			require.NoError(t, writer(rctx, putItem(rowB)))

			tester := NewWatchTester(t, writer, makeWatcher(rctx, t, reader, originalValue))
			tester.ExpectEvent(TestEvent{watch.EventPut, rowA.Id, rowA})
			tester.ExpectNoEvents()
			rowA.Value = changedValue
			tester.Write(rctx, rowA)
			tester.ExpectEvent(TestEvent{watch.EventDelete, rowA.Id, nil})
			tester.ExpectNoEvents()
		})
	}

	parent.Run("WatchByIndex", func(suite *testing.T) {
		suite.Parallel()

		suite.Run("InterruptionPreemptive", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(rctx, t)
			watchRead := reader(canceledContext())
			row := makeProto(makeID(1), originalValue)
			require.NoError(t, writer(rctx, putItem(row)))
			testInterruptionPreemptive(t, func() (watch.Watcher, error) {
				res, err := watchRead.WatchByIndex(TestSecondaryIndex, originalValue)
				return res, errors.EnsureStack(err)
			})
		})

		watchByIndexTests(suite, func(ctx context.Context, t *testing.T, reader ReadCallback, value string) watch.Watcher {
			watcher, err := reader(ctx).WatchByIndex(TestSecondaryIndex, value)
			require.NoError(t, err)
			t.Cleanup(watcher.Close)
			return watcher
		})
	})

	parent.Run("WatchByIndexF", func(suite *testing.T) {
		suite.Parallel()

		suite.Run("InterruptionPreemptive", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(rctx, t)
			watchRead := reader(canceledContext())
			row := makeProto(makeID(1), originalValue)
			require.NoError(t, writer(rctx, putItem(row)))

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
			reader, writer := newCollection(rctx, t)
			watchRead := reader(rctx)
			rowA := makeProto(makeID(1), originalValue)
			rowB := makeProto(makeID(2), originalValue)
			require.NoError(t, writer(rctx, putItem(rowA)))
			require.NoError(t, writer(rctx, putItem(rowB)))

			events := []TestEvent{}
			err := watchRead.WatchByIndexF(TestSecondaryIndex, originalValue, func(ev *watch.Event) error {
				events = append(events, newTestEvent(t, ev))
				return errutil.ErrBreak
			})
			require.NoError(t, err)
			require.Equal(t, []TestEvent{{watch.EventPut, rowA.Id, rowA}}, events)
		})

		watchByIndexTests(suite, func(ctx context.Context, t *testing.T, reader ReadCallback, value string) watch.Watcher {
			watcher, err := reader(ctx).WatchByIndex(TestSecondaryIndex, value)
			require.NoError(t, err)
			t.Cleanup(watcher.Close)
			return watcher
		})
	})
}
