package task

import (
	"context"
	"fmt"
	"path"
	"sync/atomic"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachhash"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/internal/watch"
	"golang.org/x/sync/errgroup"
)

const (
	groupPrefix = "/group"
	taskPrefix  = "/task"
	claimPrefix = "/claim"
)

type Service interface {
	// TODO: Should group ID parameter?
	Doer(ctx context.Context, namespace, groupID string, cb func(context.Context, Doer) error) error
	Maker(namespace string) Maker
	TaskCount(ctx context.Context, namespace string) (tasks int64, claims int64, err error)
}

type etcdService struct {
	etcdClient *etcd.Client
	etcdPrefix string
}

func NewEtcdService(etcdClient *etcd.Client, etcdPrefix string) Service {
	return &etcdService{
		etcdClient: etcdClient,
		etcdPrefix: etcdPrefix,
	}
}

func (es *etcdService) Doer(ctx context.Context, namespace, groupID string, cb func(context.Context, Doer) error) error {
	namespaceEtcd := newNamespaceEtcd(es.etcdClient, es.etcdPrefix, namespace)
	return namespaceEtcd.groupCol.WithRenewer(ctx, func(ctx context.Context, renewer *col.Renewer) error {
		if err := renewer.Put(ctx, groupID, &Group{}); err != nil {
			return err
		}
		return namespaceEtcd.taskCol.WithRenewer(ctx, func(ctx context.Context, renewer *col.Renewer) error {
			ed := newEtcdDoer(namespaceEtcd, groupID, renewer)
			return cb(ctx, ed)
		})
	})
}

type namespaceEtcd struct {
	etcdClient                  *etcd.Client
	groupCol, taskCol, claimCol col.EtcdCollection
}

func newNamespaceEtcd(etcdClient *etcd.Client, etcdPrefix, namespace string) *namespaceEtcd {
	return &namespaceEtcd{
		etcdClient: etcdClient,
		groupCol:   newCollection(etcdClient, path.Join(etcdPrefix, groupPrefix, namespace), &Group{}),
		taskCol:    newCollection(etcdClient, path.Join(etcdPrefix, taskPrefix, namespace), &Task{}),
		claimCol:   newCollection(etcdClient, path.Join(etcdPrefix, claimPrefix, namespace), &Claim{}),
	}
}

func newCollection(etcdClient *etcd.Client, etcdPrefix string, template proto.Message) col.EtcdCollection {
	return col.NewEtcdCollection(
		etcdClient,
		etcdPrefix,
		nil,
		template,
		nil,
		nil,
	)
}

type Doer interface {
	Do(ctx context.Context, any *types.Any) (*types.Any, error)
	DoBatch(ctx context.Context, anys []*types.Any, cb CollectFunc) error
	DoChan(ctx context.Context, anyChan chan *types.Any, cb CollectFunc) error
}

type etcdDoer struct {
	*namespaceEtcd
	groupID string
	renewer *col.Renewer
}

func newEtcdDoer(namespaceEtcd *namespaceEtcd, groupID string, renewer *col.Renewer) Doer {
	return &etcdDoer{
		namespaceEtcd: namespaceEtcd,
		groupID:       groupID,
		renewer:       renewer,
	}
}

func (ed *etcdDoer) Do(ctx context.Context, any *types.Any) (*types.Any, error) {
	// TODO: include any type in hash?
	sum := pachhash.Sum(any.Value)
	taskID := pachhash.EncodeHash(sum[:])
	prefix := uuid.NewWithoutDashes()
	taskKey := path.Join(ed.groupID, prefix, taskID)
	task := &Task{
		ID:    taskID,
		Input: any,
		State: State_RUNNING,
	}
	if err := ed.renewer.Put(ctx, taskKey, task); err != nil {
		return nil, err
	}
	defer func() {
		if _, err := col.NewSTM(ctx, ed.etcdClient, func(stm col.STM) error {
			if err := ed.taskCol.ReadWrite(stm).Delete(taskKey); err != nil {
				return err
			}
			return ed.claimCol.ReadWrite(stm).Delete(taskKey)
		}); err != nil {
			fmt.Printf("errored deleting task %v: %v\n", taskKey, err)
		}
	}()
	var result *types.Any
	if err := ed.taskCol.ReadOnly(ctx).WatchOneF(taskKey, func(e *watch.Event) error {
		if e.Type == watch.EventDelete {
			return errors.New("task was deleted while waiting for results")
		}
		var key string
		task := &Task{}
		if err := e.Unmarshal(&key, task); err != nil {
			return err
		}
		if task.State == State_RUNNING {
			return nil
		}
		if task.State == State_FAILURE {
			return errors.New(task.Reason)
		}
		result = task.Output
		return errutil.ErrBreak
	}); err != nil {
		return nil, err
	}
	return result, nil
}

type CollectFunc func(int64, *types.Any, error) error

func (ed *etcdDoer) DoBatch(ctx context.Context, anys []*types.Any, cb CollectFunc) error {
	var eg errgroup.Group
	anyChan := make(chan *types.Any)
	eg.Go(func() error {
		return ed.DoChan(ctx, anyChan, cb)
	})
	for _, any := range anys {
		select {
		case anyChan <- any:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	close(anyChan)
	return eg.Wait()
}

func (ed *etcdDoer) DoChan(ctx context.Context, anyChan chan *types.Any, cb CollectFunc) error {
	prefix := path.Join(ed.groupID, uuid.NewWithoutDashes())
	defer func() {
		if _, err := col.NewSTM(ctx, ed.etcdClient, func(stm col.STM) error {
			if err := ed.taskCol.ReadWrite(stm).DeleteAllPrefix(prefix); err != nil {
				return err
			}
			return ed.claimCol.ReadWrite(stm).DeleteAllPrefix(prefix)
		}); err != nil {
			fmt.Printf("errored deleting tasks with the prefix %v: %v\n", prefix, err)
		}
	}()
	var eg errgroup.Group
	done := make(chan struct{})
	var count int64
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	eg.Go(func() error {
		return ed.taskCol.ReadOnly(ctx).WatchOneF(prefix, func(e *watch.Event) error {
			if e.Type == watch.EventDelete {
				return errors.New("task was deleted while waiting for results")
			}
			var key string
			task := &Task{}
			if err := e.Unmarshal(&key, task); err != nil {
				return err
			}
			if task.State == State_RUNNING {
				return nil
			}
			var err error
			if task.State == State_FAILURE {
				err = errors.New(task.Reason)
			}
			if err := cb(task.Index, task.Output, err); err != nil {
				return err
			}
			atomic.AddInt64(&count, -1)
			select {
			case <-done:
				if count == 0 {
					return errutil.ErrBreak
				}
			default:
			}
			return nil
		})
	})
	var index int64
	for {
		select {
		case any, more := <-anyChan:
			if !more {
				close(done)
				// If the tasks have already been collected (or there were none), then just return.
				if atomic.LoadInt64(&count) == 0 {
					return nil
				}
				return eg.Wait()
			}
			// TODO: include any type in hash?
			sum := pachhash.Sum(any.Value)
			taskID := pachhash.EncodeHash(sum[:])
			taskKey := path.Join(prefix, taskID)
			task := &Task{
				ID:    taskID,
				Input: any,
				State: State_RUNNING,
				Index: index,
			}
			index++
			if err := ed.renewer.Put(ctx, taskKey, task); err != nil {
				return err
			}
			atomic.AddInt64(&count, 1)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (es *etcdService) Maker(namespace string) Maker {
	namespaceEtcd := newNamespaceEtcd(es.etcdClient, es.etcdPrefix, namespace)
	return newEtcdMaker(namespaceEtcd)
}

type Maker interface {
	Make(ctx context.Context, cb ProcessFunc) error
}

type etcdMaker struct {
	*namespaceEtcd
}

func newEtcdMaker(namespaceEtcd *namespaceEtcd) Maker {
	return &etcdMaker{
		namespaceEtcd: namespaceEtcd,
	}
}

type ProcessFunc func(context.Context, *types.Any) (*types.Any, error)

func (em *etcdMaker) Make(ctx context.Context, cb ProcessFunc) error {
	tq := newTaskQueue(ctx)
	return em.groupCol.ReadOnly(ctx).WatchF(func(e *watch.Event) error {
		groupID := string(e.Key)
		if e.Type == watch.EventDelete {
			tq.deleteGroup(groupID)
			return nil
		}
		return tq.group(ctx, groupID, func(ctx context.Context, taskFuncChan chan taskFunc) {
			if err := em.forEachTask(ctx, groupID, func(taskKey string) error {
				select {
				case taskFuncChan <- em.createTaskFunc(ctx, taskKey, cb):
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			}); err != nil && !errors.Is(ctx.Err(), context.Canceled) {
				fmt.Printf("errored in group callback: %v\n", err)
			}
		})
	})
}

func (em *etcdMaker) forEachTask(ctx context.Context, groupID string, cb func(string) error) error {
	claimWatch, err := em.claimCol.ReadOnly(ctx).WatchOne(groupID, watch.IgnorePut)
	if err != nil {
		return err
	}
	defer claimWatch.Close()
	taskWatch, err := em.taskCol.ReadOnly(ctx).WatchOne(groupID, watch.IgnoreDelete)
	if err != nil {
		return err
	}
	defer taskWatch.Close()
	for {
		select {
		case e := <-claimWatch.Watch():
			if e.Type == watch.EventError {
				return e.Err
			}
			taskKey := string(e.Key)
			if err := cb(taskKey); err != nil {
				return err
			}
		case e := <-taskWatch.Watch():
			if e.Type == watch.EventError {
				return e.Err
			}
			taskKey := string(e.Key)
			if err := cb(taskKey); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (em *etcdMaker) createTaskFunc(ctx context.Context, taskKey string, cb ProcessFunc) taskFunc {
	return func() {
		if err := func() error {
			// (bryce) this should be refactored to have the check and claim in the same stm.
			// there is a rare race condition that does not affect correctness, but it is less
			// than ideal because a task could get run once more than necessary.
			task := &Task{}
			if _, err := col.NewSTM(ctx, em.etcdClient, func(stm col.STM) error {
				return em.taskCol.ReadWrite(stm).Get(taskKey, task)
			}); err != nil {
				return err
			}
			if task.State != State_RUNNING {
				return nil
			}
			return em.claimCol.Claim(ctx, taskKey, &Claim{}, func(ctx context.Context) error {
				taskOutput, taskErr := cb(ctx, task.Input)
				// If the task context was canceled or the claim was lost, just return with no error.
				if errors.Is(ctx.Err(), context.Canceled) {
					return nil
				}
				task := &Task{}
				_, err := col.NewSTM(ctx, em.etcdClient, func(stm col.STM) error {
					return em.taskCol.ReadWrite(stm).Update(taskKey, task, func() error {
						// (bryce) remove when check and claim are in the same stm.
						if task.State != State_RUNNING {
							return nil
						}
						task.State = State_SUCCESS
						task.Output = taskOutput
						if taskErr != nil {
							task.State = State_FAILURE
							task.Reason = taskErr.Error()
						}
						return nil
					})
				})
				return err
			})
		}(); err != nil {
			// If the group context was canceled or the task was deleted / not claimed, then no error should be logged.
			if errors.Is(ctx.Err(), context.Canceled) ||
				col.IsErrNotFound(err) || errors.Is(err, col.ErrNotClaimed) {
				return
			}
			fmt.Printf("errored in task callback: %v\n", err)
		}
	}
}

// TaskCount returns how many tasks are in a namespace and how many are claimed.
func (es *etcdService) TaskCount(ctx context.Context, namespace string) (tasks int64, claims int64, _ error) {
	namespaceEtcd := newNamespaceEtcd(es.etcdClient, es.etcdPrefix, namespace)
	nTasks, rev, err := namespaceEtcd.taskCol.ReadOnly(ctx).CountRev(0)
	if err != nil {
		return 0, 0, err
	}
	nClaims, _, err := namespaceEtcd.claimCol.ReadOnly(ctx).CountRev(rev)
	if err != nil {
		return 0, 0, err
	}
	return nTasks, nClaims, nil
}
