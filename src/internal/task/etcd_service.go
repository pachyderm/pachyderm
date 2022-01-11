package task

import (
	"context"
	"fmt"
	"path"
	"strings"
	"sync/atomic"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachhash"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/internal/watch"
	etcd "go.etcd.io/etcd/client/v3"
	"golang.org/x/sync/errgroup"
)

const (
	groupPrefix = "/group"
	taskPrefix  = "/task"
	claimPrefix = "/claim"
)

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

func (es *etcdService) NewDoer(namespace, group string) Doer {
	if group == "" {
		group = uuid.NewWithoutDashes()
	}
	namespaceEtcd := newNamespaceEtcd(es.etcdClient, es.etcdPrefix, namespace)
	return newEtcdDoer(namespaceEtcd, group)
}

func (es *etcdService) NewSource(namespace string) Source {
	namespaceEtcd := newNamespaceEtcd(es.etcdClient, es.etcdPrefix, namespace)
	return newEtcdSource(namespaceEtcd)
}

func (es *etcdService) TaskCount(ctx context.Context, namespace string) (tasks int64, claims int64, _ error) {
	namespaceEtcd := newNamespaceEtcd(es.etcdClient, es.etcdPrefix, namespace)
	nTasks, rev, err := namespaceEtcd.taskCol.ReadOnly(ctx).CountRev(0)
	if err != nil {
		return 0, 0, errors.EnsureStack(err)
	}
	nClaims, _, err := namespaceEtcd.claimCol.ReadOnly(ctx).CountRev(rev)
	if err != nil {
		return 0, 0, errors.EnsureStack(err)
	}
	return nTasks, nClaims, nil
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

type etcdDoer struct {
	*namespaceEtcd
	group string
}

func newEtcdDoer(namespaceEtcd *namespaceEtcd, group string) Doer {
	return &etcdDoer{
		namespaceEtcd: namespaceEtcd,
		group:         group,
	}
}

func (ed *etcdDoer) Do(ctx context.Context, inputChan chan *types.Any, cb CollectFunc) error {
	return ed.withGroup(ctx, func(ctx context.Context, renewer *col.Renewer) error {
		var eg errgroup.Group
		prefix := path.Join(ed.group, uuid.NewWithoutDashes())
		done := make(chan struct{})
		var count int64
		ctx, cancel := context.WithCancel(ctx)
		defer func() {
			cancel()
			eg.Wait()
		}()
		eg.Go(func() error {
			err := ed.taskCol.ReadOnly(ctx).WatchOneF(prefix, func(e *watch.Event) error {
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
			return errors.EnsureStack(err)
		})
		defer func() {
			if _, err := col.NewSTM(ctx, ed.etcdClient, func(stm col.STM) error {
				if err := ed.taskCol.ReadWrite(stm).DeleteAllPrefix(prefix); err != nil {
					return errors.EnsureStack(err)
				}
				return errors.EnsureStack(ed.claimCol.ReadWrite(stm).DeleteAllPrefix(prefix))
			}); err != nil {
				fmt.Printf("errored deleting tasks with the prefix %v: %v\n", prefix, err)
			}
		}()
		var index int64
		for {
			select {
			case input, more := <-inputChan:
				if !more {
					close(done)
					// If the tasks have already been collected (or there were none), then just return.
					if atomic.LoadInt64(&count) == 0 {
						return nil
					}
					return errors.EnsureStack(eg.Wait())
				}
				taskID, err := computeTaskID(input)
				if err != nil {
					return err
				}
				taskKey := path.Join(prefix, taskID)
				task := &Task{
					ID:    taskID,
					Input: input,
					State: State_RUNNING,
					Index: index,
				}
				index++
				if err := renewer.Put(ctx, taskKey, task); err != nil {
					return err
				}
				atomic.AddInt64(&count, 1)
			case <-ctx.Done():
				return errors.EnsureStack(ctx.Err())
			}
		}
	})
}

func (ed *etcdDoer) withGroup(ctx context.Context, cb func(ctx context.Context, renewer *col.Renewer) error) error {
	err := ed.groupCol.WithRenewer(ctx, func(ctx context.Context, renewer *col.Renewer) error {
		if err := renewer.Put(ctx, path.Join(ed.group, uuid.NewWithoutDashes()), &Group{}); err != nil {
			return err
		}
		err := ed.taskCol.WithRenewer(ctx, func(ctx context.Context, renewer *col.Renewer) error {
			return cb(ctx, renewer)
		})
		return errors.EnsureStack(err)
	})
	return errors.EnsureStack(err)
}

func computeTaskID(input *types.Any) (string, error) {
	val, err := proto.Marshal(input)
	if err != nil {
		return "", errors.EnsureStack(err)
	}
	sum := pachhash.Sum(val)
	return pachhash.EncodeHash(sum[:]), nil
}

type etcdSource struct {
	*namespaceEtcd
}

func newEtcdSource(namespaceEtcd *namespaceEtcd) Source {
	return &etcdSource{
		namespaceEtcd: namespaceEtcd,
	}
}

func (es *etcdSource) Iterate(ctx context.Context, cb ProcessFunc) error {
	groups := make(map[string]map[string]struct{})
	tq := newTaskQueue(ctx)
	err := es.groupCol.ReadOnly(ctx).WatchF(func(e *watch.Event) error {
		group, uuid := path.Split(string(e.Key))
		group = strings.TrimRight(group, "/")
		groupMap, ok := groups[group]
		if e.Type == watch.EventDelete {
			if !ok {
				return nil
			}
			delete(groupMap, uuid)
			if len(groupMap) == 0 {
				tq.deleteGroup(group)
				delete(groups, group)
			}
			return nil
		}
		if ok {
			groupMap[uuid] = struct{}{}
			return nil
		}
		groupMap = make(map[string]struct{})
		groups[group] = groupMap
		groupMap[uuid] = struct{}{}
		return tq.group(ctx, group, func(ctx context.Context, taskFuncChan chan taskFunc) {
			if err := es.forEachTask(ctx, group, func(taskKey string) error {
				select {
				case taskFuncChan <- es.createTaskFunc(ctx, taskKey, cb):
					return nil
				case <-ctx.Done():
					return errors.EnsureStack(ctx.Err())
				}
			}); err != nil && !errors.Is(ctx.Err(), context.Canceled) {
				fmt.Printf("errored in group callback: %v\n", err)
			}
		})
	})
	return errors.EnsureStack(err)
}

func (es *etcdSource) forEachTask(ctx context.Context, group string, cb func(string) error) error {
	claimWatch, err := es.claimCol.ReadOnly(ctx).WatchOne(group, watch.IgnorePut)
	if err != nil {
		return errors.EnsureStack(err)
	}
	defer claimWatch.Close()
	taskWatch, err := es.taskCol.ReadOnly(ctx).WatchOne(group, watch.IgnoreDelete)
	if err != nil {
		return errors.EnsureStack(err)
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
			return errors.EnsureStack(ctx.Err())
		}
	}
}

func (es *etcdSource) createTaskFunc(ctx context.Context, taskKey string, cb ProcessFunc) taskFunc {
	return func() {
		if err := func() error {
			task := &Task{}
			if _, err := col.NewSTM(ctx, es.etcdClient, func(stm col.STM) error {
				return errors.EnsureStack(es.taskCol.ReadWrite(stm).Get(taskKey, task))
			}); err != nil {
				return err
			}
			if task.State != State_RUNNING {
				return nil
			}
			err := es.claimCol.Claim(ctx, taskKey, &Claim{}, func(ctx context.Context) error {
				taskOutput, taskErr := cb(ctx, task.Input)
				// If the task context was canceled or the claim was lost, just return with no error.
				if errors.Is(ctx.Err(), context.Canceled) {
					return nil
				}
				task := &Task{}
				_, err := col.NewSTM(ctx, es.etcdClient, func(stm col.STM) error {
					err := es.taskCol.ReadWrite(stm).Update(taskKey, task, func() error {
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
					return errors.EnsureStack(err)
				})
				return err
			})
			return errors.EnsureStack(err)
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
