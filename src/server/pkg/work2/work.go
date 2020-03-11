package work2

import (
	"context"
	"fmt"
	"path"
	"sync"

	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/errutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/server/pkg/watch"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
	types "github.com/gogo/protobuf/types"
	"golang.org/x/sync/errgroup"
)

const (
	subtaskPrefix = "/subtask"
	claimPrefix   = "/claim"
)

// Task thing
type Task interface {
	Run(
		ctx context.Context,
		subtaskData chan *types.Any,
		collect func(ctx context.Context, data *types.Any) error,
	) error
}

// Master thing
type Master interface {
	Clear(ctx context.Context) error
	WithTask(cb func(task Task) error) error
}

// Worker thing
type Worker interface {
	Run(
		ctx context.Context,
		cb func(ctx context.Context, data *types.Any) error,
	) error
}

type taskEtcd struct {
	etcdClient *etcd.Client
	subtasks   col.Collection
	claims     col.Collection
}

type master struct {
	*taskEtcd
}

type worker struct {
	*taskEtcd
}

type task struct {
	master    *master
	id        string
	timestamp *types.Timestamp
}

func newCollection(etcdClient *etcd.Client, etcdPrefix string, template proto.Message) col.Collection {
	return col.NewCollection(
		etcdClient,
		etcdPrefix,
		nil,
		template,
		nil,
		nil,
	)
}

func newTaskEtcd(etcdClient *etcd.Client, etcdPrefix string, taskNamespace string) *taskEtcd {
	return &taskEtcd{
		etcdClient: etcdClient,
		subtasks:   newCollection(etcdClient, path.Join(etcdPrefix, subtaskPrefix, taskNamespace), &Subtask{}),
		claims:     newCollection(etcdClient, path.Join(etcdPrefix, claimPrefix, taskNamespace), &Claim{}),
	}
}

// NewMaster thing
func NewMaster(etcdClient *etcd.Client, etcdPrefix string, taskNamespace string) Master {
	return &master{
		taskEtcd: newTaskEtcd(etcdClient, etcdPrefix, taskNamespace),
	}
}

// NewWorker thing
func NewWorker(etcdClient *etcd.Client, etcdPrefix string, taskNamespace string) Worker {
	return &worker{
		taskEtcd: newTaskEtcd(etcdClient, etcdPrefix, taskNamespace),
	}
}

func (m *master) Clear(ctx context.Context) error {
	return m.clear(ctx, "")
}

func (m *master) clear(ctx context.Context, prefix string) error {
	_, err := col.NewSTM(context.Background(), m.etcdClient, func(stm col.STM) error {
		m.subtasks.ReadWrite(stm).DeleteAllPrefix(prefix)
		return nil
	})
	return err
}

func (m *master) WithTask(cb func(task Task) error) (retErr error) {
	id := fmt.Sprintf("task-%s", uuid.NewWithoutDashes())

	// Remove all subtasks once the task goes out of scope
	defer func() {
		if err := m.clear(context.Background(), id); retErr == nil {
			retErr = err
		}
	}()

	return cb(&task{
		master:    m,
		id:        id,
		timestamp: types.TimestampNow(),
	})
}

type subtaskEntry struct {
	id        string
	timestamp *types.Timestamp
	ctx       context.Context
	cancel    context.CancelFunc
}

func (w *worker) runSubtasks(ctx context.Context, subtaskQueue []*subtaskEntry, cb func(context.Context, *types.Any) error) error {
	// Loop through subtasks, attempt to claim one
	for _, entry := range subtaskQueue {
		fmt.Printf("worker claiming entry: %v\n", entry.id)
		if err := w.claims.Claim(ctx, entry.id, &Claim{}, func(ctx context.Context) (retErr error) {
			// Read out the claimed subtask
			subtask := &Subtask{}
			if err := w.subtasks.ReadOnly(ctx).Get(entry.id, subtask); err != nil {
				if col.IsErrNotFound(err) {
					return nil
				}
				return err
			}

			if subtask.State != State_RUNNING {
				return nil
			}

			// When we return, write out the updated subtask
			defer func() {
				if _, err := col.NewSTM(ctx, w.etcdClient, func(stm col.STM) error {
					updateSubtask := &Subtask{}
					return w.subtasks.ReadWrite(stm).Update(entry.id, updateSubtask, func() error {
						if subtask.State != State_RUNNING {
							fmt.Printf("unexpected state: subtask was finished while claimed by a worker")
							return nil
						}
						updateSubtask.State = State_SUCCESS
						updateSubtask.UserData = subtask.UserData
						if retErr != nil {
							updateSubtask.State = State_FAILURE
							updateSubtask.Reason = retErr.Error()
							retErr = nil
						}
						return nil
					})
				}); retErr == nil && !col.IsErrNotFound(err) {
					retErr = err
				}
			}()

			// We need a different ctx that will be canceled if the task gets deleted
			userCtx, cancel := context.WithCancel(ctx)
			go func() {
				<-entry.ctx.Done()
				cancel()
			}()
			defer entry.cancel()

			return cb(userCtx, subtask.UserData)
		}); err == nil {
			// We processed a subtask - abort this loop and resume from the start of subtasks again
			return nil
		} else if err != col.ErrNotClaimed {
			return err
		}
	}
	return nil
}

func (w *worker) Run(
	ctx context.Context,
	cb func(ctx context.Context, subtask *types.Any) error,
) (retErr error) {
	eg, ctx := errgroup.WithContext(ctx)

	mutex := sync.Mutex{}
	subtaskQueue := []*subtaskEntry{}
	changeChan := make(chan struct{}, 1)

	// Watch subtasks collection for subtasks in the RUNNING state, organize by task timestamp
	eg.Go(func() (retErr error) {
		defer close(changeChan)

		if err := w.subtasks.ReadOnly(ctx).WatchF(func(e *watch.Event) error {
			if e.Type == watch.EventError {
				return e.Err
			}

			var subtaskKey string
			subtask := Subtask{}
			if err := e.Unmarshal(&subtaskKey, &subtask); err != nil {
				return err
			}

			mutex.Lock()
			defer mutex.Unlock()

			index := -1
			for i, entry := range subtaskQueue {
				if entry.id == subtask.ID {
					index = i
					break
				}
			}
			fmt.Printf("worker got event: %v %v %v\n", e.Type, subtaskKey, subtask.State)

			if e.Type == watch.EventPut && subtask.State == State_RUNNING {
				if index == -1 {
					ctx, cancel := context.WithCancel(ctx)
					// TODO: insert based on task timestamp
					subtaskQueue = append(subtaskQueue, &subtaskEntry{
						id:        subtask.ID,
						timestamp: subtask.TaskTime,
						ctx:       ctx,
						cancel:    cancel,
					})
				}
			} else if index != -1 {
				fmt.Printf("removing subtask: %d of %d\n", index, len(subtaskQueue))
				// Subtask is not valid, cancel and remove it
				subtaskQueue[index].cancel()
				subtaskQueue = append(subtaskQueue[:index], subtaskQueue[index+1:]...)
			}

			select {
			case changeChan <- struct{}{}:
			default:
			}
			return nil
		}); err != nil && err != context.Canceled {
			return err
		}
		return nil
	})

	eg.Go(func() (retErr error) {
		// We looped over all known subtasks and didn't get any claims, wait for
		// something to change
		for {
			if _, ok := <-changeChan; !ok {
				fmt.Printf("worker second coroutine exiting, changeChan closed\n")
				break
			}
			fmt.Printf("worker coroutine looping\n")
			queueCopy := []*subtaskEntry{}
			func() {
				mutex.Lock()
				defer mutex.Unlock()
				queueCopy = append(queueCopy, subtaskQueue...)
			}()

			if err := w.runSubtasks(ctx, queueCopy, cb); err != nil {
				if err == context.Canceled {
					return nil
				}
				fmt.Printf("worker second coroutine returning: %v\n", err)
				return err
			}
		}

		fmt.Printf("worker second coroutine returning nil\n")
		return nil
	})

	err := eg.Wait()
	fmt.Printf("worker returning err: %v\n", err)
	return err
}

func (t *task) Run(
	ctx context.Context,
	subtaskData chan *types.Any,
	collect func(ctx context.Context, subtask *types.Any) error,
) (retErr error) {
	runID := path.Join(t.id, fmt.Sprintf("run-%s", uuid.NewWithoutDashes()))

	// Remove all subtasks from this run when we return
	defer func() {
		subtask := &Subtask{}
		t.master.subtasks.ReadOnly(ctx).ListPrefix(runID, subtask, &col.Options{}, func(key string) error {
			fmt.Printf("subtask remaining in etcd: %v\n", key)
			return nil
		})

		if err := t.master.clear(context.Background(), runID); retErr == nil {
			retErr = err
		}

		t.master.subtasks.ReadOnly(ctx).ListPrefix(runID, subtask, &col.Options{}, func(key string) error {
			fmt.Printf("subtask remaining in etcd: %v\n", key)
			return nil
		})
	}()

	egCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	eg, egCtx := errgroup.WithContext(egCtx)
	numSubtasks := 0
	subtasksReady := make(chan struct{})

	eg.Go(func() (retErr error) {
		count := 0
		defer close(subtasksReady)

		for userData := range subtaskData {
			fmt.Printf("master received subtask: %v\n", subtaskData)
			subtask := &Subtask{
				ID:       path.Join(runID, fmt.Sprintf("subtask-%s", uuid.NewWithoutDashes())),
				UserData: userData,
				TaskTime: t.timestamp,
				State:    State_RUNNING,
			}

			if _, err := col.NewSTM(egCtx, t.master.etcdClient, func(stm col.STM) error {
				fmt.Printf("master writing subtask: %v\n", subtask.ID)
				return t.master.subtasks.ReadWrite(stm).Put(subtask.ID, subtask)
			}); err != nil {
				return err
			}
			count++
		}

		// If there are no etcd events to watch for, we need to cancel the other goroutine
		if count == 0 {
			cancel()
		}

		numSubtasks = count
		return nil
	})

	eg.Go(func() (retErr error) {
		finishedSubtasks := 0
		return t.master.subtasks.ReadOnly(egCtx).WatchPrefixF(runID, func(e *watch.Event) error {
			switch e.Type {
			case watch.EventError:
				return e.Err
			case watch.EventPut:
				var subtaskKey string
				subtask := Subtask{}
				if err := e.Unmarshal(&subtaskKey, &subtask); err != nil {
					return err
				}

				if subtask.State == State_FAILURE {
					return fmt.Errorf("%s", subtask.Reason)
				}

				if subtask.State == State_SUCCESS {
					if err := collect(egCtx, subtask.UserData); err != nil {
						return err
					}

					finishedSubtasks++

					// If we've finished with all the subtasks, we can break
					if isDone(subtasksReady) && finishedSubtasks == numSubtasks {
						return errutil.ErrBreak
					}
				}

				return nil
			case watch.EventDelete:
				return fmt.Errorf("subtask was unexpectedly deleted")
			}
			return fmt.Errorf("unrecognized watch event: %v", e.Type)
		})
	})

	err := eg.Wait()

	// Handle the shitty corner case cancelling for when we didn't get any subtasks
	if err == context.Canceled && !isDone(ctx.Done()) {
		fmt.Printf("master returning err: nil\n")
		return nil
	}

	fmt.Printf("master returning err: %v\n", err)
	return err
}

func isDone(channel <-chan struct{}) bool {
	select {
	case <-channel:
		return true
	default:
		return false
	}
}
