package collection_test

import (
	"context"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"golang.org/x/sync/errgroup"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
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

func nextEventOrErr(t *testing.T, c <-chan *watch.Event) *watch.Event {
	result := nextEvent(c, 5*time.Second)
	require.NotNil(t, result)
	return result
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
	Write(id string, value ...string)     // writes a row with the given ID and value
	Delete(id string)                     // deletes a row with the given ID
	DeleteAll()                           // deletes all rows
	ExpectEvent(expected TestEvent)       // expects an event matching the given event
	ExpectEventSet(expected ...TestEvent) // expects a set of events matching the given events (in no given order)
	ExpectError(err error)                // expects an error that matches the given error type
	ExpectNothing()                       // expects no events to be received for a short period
}

type ChannelWatchTester struct {
	t       *testing.T
	writer  WriteCallback
	watcher watch.Watcher
}

func (tester *ChannelWatchTester) nextEvent(timeout time.Duration) *watch.Event {
	return nextEvent(tester.watcher.Watch(), timeout)
}

func (tester *ChannelWatchTester) Write(id string, value ...string) {
	err := tester.writer(context.Background(), func(rw col.ReadWriteCollection) error {
		testProto := makeProto(id)
		if len(value) > 0 {
			testProto.Value = value[0]
		}
		return rw.Put(testProto.ID, testProto)
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
	for _ = range expected {
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

func (tester *ChannelWatchTester) ExpectNothing() {
	require.Nil(tester.t, tester.nextEvent(100*time.Millisecond))
}

func newChannelWatchTester(t *testing.T, writer WriteCallback, makeWatcher func() (watch.Watcher, error)) WatchTester {
	watcher, err := makeWatcher()
	require.NoError(t, err)

	t.Cleanup(func() {
		requireEmptyChannel(t, watcher.Watch())
		watcher.Close()
	})
	return &ChannelWatchTester{t, writer, watcher}
}

type CallbackWatcherShim struct {
	watchChan chan *watch.Event
}

type CallbackWatchTester struct {
	ChannelWatchTester
	eg *errgroup.Group
}

func (shim *CallbackWatcherShim) Watch() <-chan *watch.Event {
	return shim.watchChan
}

func (shim *CallbackWatcherShim) Close() {
	close(shim.watchChan)
}

// newCallbackWatchTester provides a WatchTester that works with a
// callback-based watch (i.e. WatchF() instead of Watch()). It spawns a
// goroutine that will run the callback-based watch and pass the events into a
// shim that fulfills the Watcher interface so we can reuse ChannelWatchTester.
func newCallbackWatchTester(t *testing.T, writer WriteCallback, doWatch func(context.Context, func(*watch.Event) error) error) WatchTester {
	watchChan := make(chan *watch.Event)

	ctx, cancel := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		// There is no way to test ErrBreak behavior unless we know the set of
		// expected events ahead of time, so WatchF-style tests need to test this
		// manually.
		if err := doWatch(ctx, func(ev *watch.Event) error {
			watchChan <- ev
			return nil
		}); err != nil {
			watchChan <- &watch.Event{Type: watch.EventError, Err: err}
		}
		return nil
	})

	t.Cleanup(func() {
		cancel()
		require.NoError(t, eg.Wait())
	})

	makeWatcher := func() (watch.Watcher, error) {
		return &CallbackWatcherShim{watchChan}, nil
	}

	return newChannelWatchTester(t, writer, makeWatcher)
}

// TODO: inject errors into watches

func watchTests(
	parent *testing.T,
	newCollection func(context.Context, *testing.T) (ReadCallback, WriteCallback),
) {
	testPreemptiveInterruption := func(t *testing.T, makeWatcher func() (watch.Watcher, error)) {
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

	watchAllTests := func(suite *testing.T, cb func(readOnly col.ReadOnlyCollection) WatchTester) {
	}

	parent.Run("Watch", func(suite *testing.T) {
		suite.Parallel()

		suite.Run("Interruption", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(context.Background(), t)
			watchRead := reader(canceledContext())
			writer(context.Background(), doWrite(makeID(3)))
			testPreemptiveInterruption(t, func() (watch.Watcher, error) { return watchRead.Watch() })
		})

		suite.Run("InterruptionAfterInitial", func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			reader, writer := newCollection(context.Background(), t)
			watchRead := reader(ctx)
			id := makeID(4)
			writer(context.Background(), doWrite(id))

			tester := newChannelWatchTester(t, writer, func() (watch.Watcher, error) { return watchRead.Watch() })

			tester.ExpectEvent(TestEvent{watch.EventPut, id, makeProto(id)})
			tester.ExpectNothing()
			tester.Delete(id)
			tester.ExpectEvent(TestEvent{watch.EventDelete, id, nil})
			tester.ExpectNothing()
			cancel()
			tester.ExpectError(context.Canceled)
			tester.ExpectNothing()
		})

		suite.Run("Delete", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(context.Background(), t)
			watchRead := reader(context.Background())
			id := makeID(1)
			writer(context.Background(), doWrite(id))

			tester := newChannelWatchTester(t, writer, func() (watch.Watcher, error) { return watchRead.Watch() })

			tester.ExpectEvent(TestEvent{watch.EventPut, id, makeProto(id)})
			tester.ExpectNothing()
			tester.Delete(id)
			tester.ExpectEvent(TestEvent{watch.EventDelete, id, nil})
			tester.ExpectNothing()
		})

		suite.Run("DeleteAll", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(context.Background(), t)
			watchRead := reader(context.Background())
			idA := makeID(1)
			idB := makeID(2)
			writer(context.Background(), doWrite(idA))
			writer(context.Background(), doWrite(idB))

			tester := newChannelWatchTester(t, writer, func() (watch.Watcher, error) { return watchRead.Watch() })

			tester.ExpectEvent(TestEvent{watch.EventPut, idA, makeProto(idA)})
			tester.ExpectEvent(TestEvent{watch.EventPut, idB, makeProto(idB)})
			tester.ExpectNothing()
			tester.DeleteAll()
			tester.ExpectEventSet(
				TestEvent{watch.EventDelete, idA, nil},
				TestEvent{watch.EventDelete, idB, nil},
			)
			tester.ExpectNothing()
		})

		suite.Run("Create", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(context.Background(), t)
			watchRead := reader(context.Background())
			idA := makeID(1)
			idB := makeID(2)

			tester := newChannelWatchTester(t, writer, func() (watch.Watcher, error) { return watchRead.Watch() })

			tester.ExpectNothing()
			tester.Write(idA)
			tester.ExpectEvent(TestEvent{watch.EventPut, idA, makeProto(idA)})
			tester.ExpectNothing()
			tester.Write(idB)
			tester.ExpectEvent(TestEvent{watch.EventPut, idB, makeProto(idB)})
			tester.ExpectNothing()
		})

		suite.Run("Overwrite", func(t *testing.T) {
			t.Parallel()
			reader, writer := newCollection(context.Background(), t)
			watchRead := reader(context.Background())
			id := makeID(1)
			writer(context.Background(), doWrite(id))

			tester := newChannelWatchTester(t, writer, func() (watch.Watcher, error) { return watchRead.Watch() })

			tester.ExpectEvent(TestEvent{watch.EventPut, id, makeProto(id)})
			tester.ExpectNothing()
			tester.Write(id, changedValue)
			tester.ExpectEvent(TestEvent{watch.EventPut, id, makeProto(id, changedValue)})
			tester.ExpectNothing()
		})
	})
	/*
		parent.Run("WatchF", func(suite *testing.T) {
			suite.Parallel()

			// TODO: options: filter, initial sort

			suite.Run("Interruption", func(t *testing.T) {
				t.Parallel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(canceledContext())
				writer(context.Background(), doWrite(makeID(1)))

				events := []*watch.Event{}
				err := watchRead.WatchF(collectEventsCallback(1, &events))
				require.YesError(t, err)
				require.True(t, errors.Is(err, context.Canceled))
				require.Equal(t, 0, len(events))
			})

			suite.Run("InterruptionAfterInitial", func(t *testing.T) {
				t.Parallel()
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(ctx)
				id := makeID(1)
				writer(context.Background(), doWrite(id))
				asyncWrite(t, writer, doDelete(id))

				events := []*watch.Event{}
				err := watchRead.WatchF(func(ev *watch.Event) error {
					events = append(events, ev)
					if len(events) == 2 {
						cancel() // Stop iteration via cancel once we're out of the initial value stage
					}
					return nil
				})
				require.YesError(t, err)
				require.True(t, errors.Is(err, context.Canceled))
				checkEvents(t, events, []TestEvent{
					{watch.EventPut, id, makeProto(id)},
					{watch.EventDelete, id, nil},
				})
			})

			suite.Run("Delete", func(t *testing.T) {
				t.Parallel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(context.Background())
				id := makeID(1)
				writer(context.Background(), doWrite(id))
				asyncWrite(t, writer, doDelete(id))

				events := []*watch.Event{}
				err := watchRead.WatchF(collectEventsCallback(2, &events))
				require.NoError(t, err)
				checkEvents(t, events, []TestEvent{
					{watch.EventPut, id, makeProto(id)},
					{watch.EventDelete, id, nil},
				})
			})

			suite.Run("DeleteAll", func(t *testing.T) {
				t.Parallel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(context.Background())
				idA := makeID(1)
				idB := makeID(2)
				writer(context.Background(), doWrite(idA))
				writer(context.Background(), doWrite(idB))
				asyncWrite(t, writer, doDeleteAll())

				events := []*watch.Event{}
				err := watchRead.WatchF(collectEventsCallback(4, &events))
				require.NoError(t, err)
				checkEvents(t, events, []TestEvent{
					{watch.EventPut, idA, makeProto(idA)},
					{watch.EventPut, idB, makeProto(idB)},
					{watch.EventDelete, idA, nil}, // TODO: deleteAll order isn't well-defined?
					{watch.EventDelete, idB, nil},
				})
			})

			suite.Run("Create", func(t *testing.T) {
				t.Parallel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(context.Background())
				idA := makeID(1)
				idB := makeID(2)
				asyncWrite(t, writer, doWrite(idA), doWrite(idB))

				events := []*watch.Event{}
				err := watchRead.WatchF(collectEventsCallback(2, &events))
				require.NoError(t, err)
				checkEvents(t, events, []TestEvent{
					{watch.EventPut, idA, makeProto(idA)},
					{watch.EventPut, idB, makeProto(idB)},
				})
			})

			suite.Run("Overwrite", func(t *testing.T) {
				t.Parallel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(context.Background())
				id := makeID(1)
				writer(context.Background(), doWrite(id))
				asyncWrite(t, writer, doWrite(id))

				events := []*watch.Event{}
				err := watchRead.WatchF(collectEventsCallback(2, &events))
				require.NoError(t, err)
				checkEvents(t, events, []TestEvent{
					{watch.EventPut, id, makeProto(id)},
					{watch.EventPut, id, makeProto(id)},
				})
			})

			suite.Run("ErrBreak", func(t *testing.T) {
				t.Parallel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(context.Background())
				idA := makeID(1)
				idB := makeID(2)
				writer(context.Background(), doWrite(idA))
				writer(context.Background(), doWrite(idB))

				events := []*watch.Event{}
				err := watchRead.WatchF(collectEventsCallback(1, &events))
				require.NoError(t, err)
				checkEvents(t, events, []TestEvent{
					{watch.EventPut, idA, makeProto(idA)},
				})
			})
		})

		parent.Run("WatchOne", func(suite *testing.T) {
			suite.Parallel()
			suite.Run("ErrNotFound", func(t *testing.T) {
				t.Parallel()
			})
			suite.Run("Success", func(t *testing.T) {
				t.Parallel()
			})
		})

		parent.Run("WatchOneF", func(suite *testing.T) {
			suite.Parallel()

			// TODO: options: filter, initial sort

			suite.Run("Interruption", func(t *testing.T) {
				t.Parallel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(canceledContext())
				id := makeID(1)
				writer(context.Background(), doWrite(makeID(1)))

				events := []*watch.Event{}
				err := watchRead.WatchOneF(id, collectEventsCallback(1, &events))
				require.YesError(t, err)
				require.True(t, errors.Is(err, context.Canceled))
				require.Equal(t, 0, len(events))
			})

			suite.Run("InterruptionAfterInitial", func(t *testing.T) {
				t.Parallel()
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(ctx)
				id := makeID(1)
				writer(context.Background(), doWrite(id))
				asyncWrite(t, writer, doDelete(id))

				expected := []TestEvent{
					{watch.EventPut, id, makeProto(id)},
					{watch.EventDelete, id, nil},
				}
				events := []*watch.Event{}
				err := watchRead.WatchOneF(id, func(ev *watch.Event) error {
					events = append(events, ev)
					if len(events) == len(expected) {
						cancel() // Stop iteration via cancel once we're out of the initial value stage
					}
					return nil
				})
				require.YesError(t, err)
				require.True(t, errors.Is(err, context.Canceled))
				checkEvents(t, events, expected)
			})

			suite.Run("Delete", func(t *testing.T) {
				t.Parallel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(context.Background())
				idA := makeID(1)
				idB := makeID(2)
				writer(context.Background(), doWrite(idA))
				writer(context.Background(), doWrite(idB))
				asyncWrite(t, writer, doDelete(idA), doDelete(idB))

				expected := []TestEvent{
					{watch.EventPut, idA, makeProto(idA)},
					{watch.EventDelete, idA, nil},
				}
				events := []*watch.Event{}
				err := watchRead.WatchOneF(idA, collectEventsCallback(len(expected), &events))
				require.NoError(t, err)
				checkEvents(t, events, expected)
			})

			suite.Run("DeleteAll", func(t *testing.T) {
				t.Parallel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(context.Background())
				idA := makeID(1)
				idB := makeID(2)
				writer(context.Background(), doWrite(idA))
				writer(context.Background(), doWrite(idB))
				asyncWrite(t, writer, doDeleteAll())

				expected := []TestEvent{
					{watch.EventPut, idA, makeProto(idA)},
					{watch.EventDelete, idA, nil},
				}
				events := []*watch.Event{}
				err := watchRead.WatchOneF(idA, collectEventsCallback(len(expected), &events))
				require.NoError(t, err)
				checkEvents(t, events, expected)
			})

			suite.Run("Create", func(t *testing.T) {
				t.Parallel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(context.Background())
				idA := makeID(1)
				idB := makeID(2)
				asyncWrite(t, writer, doWrite(idA), doWrite(idB))

				expected := []TestEvent{
					{watch.EventPut, idA, makeProto(idA)},
				}
				events := []*watch.Event{}
				err := watchRead.WatchOneF(idA, collectEventsCallback(len(expected), &events))
				require.NoError(t, err)
				checkEvents(t, events, expected)
			})

			suite.Run("Overwrite", func(t *testing.T) {
				t.Parallel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(context.Background())
				idA := makeID(1)
				idB := makeID(2)
				writer(context.Background(), doWrite(idA))
				writer(context.Background(), doWrite(idB))
				asyncWrite(t, writer, doWrite(idA), doWrite(idB))

				expected := []TestEvent{
					{watch.EventPut, idA, makeProto(idA)},
					{watch.EventPut, idA, makeProto(idA)},
				}
				events := []*watch.Event{}
				err := watchRead.WatchOneF(idA, collectEventsCallback(len(expected), &events))
				require.NoError(t, err)
				checkEvents(t, events, expected)
			})

			suite.Run("ErrBreak", func(t *testing.T) {
				t.Parallel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(context.Background())
				idA := makeID(1)
				idB := makeID(2)
				writer(context.Background(), doWrite(idA))
				writer(context.Background(), doWrite(idB))

				expected := []TestEvent{
					{watch.EventPut, idA, makeProto(idA)},
				}
				events := []*watch.Event{}
				err := watchRead.WatchOneF(idA, collectEventsCallback(len(expected), &events))
				require.NoError(t, err)
				checkEvents(t, events, expected)
			})
		})

		parent.Run("WatchByIndex", func(suite *testing.T) {
			suite.Parallel()

			suite.Run("Interruption", func(t *testing.T) {
				t.Parallel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(canceledContext())
				writer(context.Background(), doWrite(makeID(3)))

				watcher, err := watchRead.WatchByIndex(TestSecondaryIndex, changedValue)
				// Difference between postgres and etcd - etcd will error at the `Watch` call, postgres will error in the channel
				if err != nil {
					require.True(t, errors.Is(err, context.Canceled))
					return
				}
				defer watcher.Close()

				events := []*watch.Event{}
				collectEventsChannel(watcher, 1, &events)
				checkEvents(t, events, []TestEvent{{watch.EventError, "", nil}})
				require.YesError(t, events[0].Err)
				require.True(t, errors.Is(events[0].Err, context.Canceled))
			})

			suite.Run("InterruptionAfterInitial", func(t *testing.T) {
				t.Parallel()
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(ctx)
				id := makeID(4)
				writer(context.Background(), doWrite(id))
				asyncWrite(t, writer, doDelete(id))

				watcher, err := watchRead.WatchByIndex(TestSecondaryIndex, changedValue)
				require.NoError(t, err)
				defer watcher.Close()

				events := []*watch.Event{}
				for {
					events = append(events, <-watcher.Watch())
					if events[len(events)-1].Type == watch.EventError {
						break
					}
					if len(events) == 2 {
						cancel()
					}
				}
				checkEvents(t, events, []TestEvent{
					{watch.EventPut, id, makeProto(id)},
					{watch.EventDelete, id, nil},
					{watch.EventError, "", nil},
				})
				require.YesError(t, events[2].Err)
				require.True(t, errors.Is(events[2].Err, context.Canceled))
			})

			suite.Run("Delete", func(t *testing.T) {
				t.Parallel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(context.Background())
				id := makeID(1)
				writer(context.Background(), doWrite(id))
				asyncWrite(t, writer, doDelete(id))

				events := []*watch.Event{}
				watcher, err := watchRead.WatchByIndex(TestSecondaryIndex, changedValue)
				require.NoError(t, err)
				defer watcher.Close()
				collectEventsChannel(watcher, 2, &events)
				checkEvents(t, events, []TestEvent{
					{watch.EventPut, id, makeProto(id)},
					{watch.EventDelete, id, nil},
				})
			})

			suite.Run("DeleteAll", func(t *testing.T) {
				t.Parallel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(context.Background())
				idA := makeID(1)
				idB := makeID(2)
				writer(context.Background(), doWrite(idA))
				writer(context.Background(), doWrite(idB))
				asyncWrite(t, writer, doDeleteAll())

				events := []*watch.Event{}
				watcher, err := watchRead.WatchByIndex(TestSecondaryIndex, changedValue)
				require.NoError(t, err)
				defer watcher.Close()
				collectEventsChannel(watcher, 4, &events)
				checkEvents(t, events, []TestEvent{
					{watch.EventPut, idA, makeProto(idA)},
					{watch.EventPut, idB, makeProto(idB)},
					{watch.EventDelete, idA, nil}, // TODO: deleteAll order isn't well-defined?
					{watch.EventDelete, idB, nil},
				})
			})

			suite.Run("Create", func(t *testing.T) {
				t.Parallel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(context.Background())
				idA := makeID(1)
				idB := makeID(2)
				asyncWrite(t, writer, doWrite(idA), doWrite(idB))

				events := []*watch.Event{}
				watcher, err := watchRead.WatchByIndex(TestSecondaryIndex, changedValue)
				require.NoError(t, err)
				defer watcher.Close()
				collectEventsChannel(watcher, 2, &events)
				checkEvents(t, events, []TestEvent{
					{watch.EventPut, idA, makeProto(idA)},
					{watch.EventPut, idB, makeProto(idB)},
				})
			})

			suite.Run("Overwrite", func(t *testing.T) {
				t.Parallel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(context.Background())
				id := makeID(1)
				writer(context.Background(), doWrite(id))
				asyncWrite(t, writer, doWrite(id))

				events := []*watch.Event{}
				watcher, err := watchRead.WatchByIndex(TestSecondaryIndex, changedValue)
				require.NoError(t, err)
				defer watcher.Close()
				collectEventsChannel(watcher, 2, &events)
				checkEvents(t, events, []TestEvent{
					{watch.EventPut, id, makeProto(id)},
					{watch.EventPut, id, makeProto(id)},
				})
			})
		})
	*/
}
