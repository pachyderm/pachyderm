package task

import (
	"context"
	"path"
	"strings"
	"sync/atomic"

	etcd "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachhash"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/internal/watch"
	"github.com/pachyderm/pachyderm/v2/src/version"
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
		etcdPrefix: path.Join(etcdPrefix, version.PrettyVersion()),
	}
}

func (es *etcdService) NewDoer(namespace, group string, cache Cache) Doer {
	if group == "" {
		group = uuid.NewWithoutDashes()
	}
	namespaceEtcd := newNamespaceEtcd(es.etcdClient, es.etcdPrefix, namespace)
	return newEtcdDoer(namespaceEtcd, group, cache)
}

func (es *etcdService) NewSource(namespace string) Source {
	namespaceEtcd := newNamespaceEtcd(es.etcdClient, es.etcdPrefix, namespace)
	return newEtcdSource(namespaceEtcd)
}

func (es *etcdService) List(ctx context.Context, namespace, group string, cb func(string, string, *Task, bool) error) error {
	if namespace == "" && group != "" {
		return errors.New("must provide a task namespace to list a group")
	}
	prefix := path.Join(namespace, group)
	etcdCols := newNamespaceEtcd(es.etcdClient, es.etcdPrefix, prefix)
	var taskData Task
	var claim Claim
	return errors.EnsureStack(etcdCols.taskCol.ReadOnly(ctx).List(&taskData, col.DefaultOptions(), func(key string) error {
		var claimed bool
		if taskData.State == State_RUNNING && etcdCols.claimCol.ReadOnly(ctx).Get(key, &claim) == nil {
			claimed = true
		}
		// parse out namespace and group from key in case they weren't provided
		fullKey := strings.TrimPrefix(path.Join(prefix, key), "/")

		// namespace/group/doerID/taskID
		keyParts := strings.Split(fullKey, "/")
		if len(keyParts) != 4 {
			return errors.Errorf("malformed task key %s", fullKey)
		}
		return cb(keyParts[0], keyParts[1], &taskData, claimed)
	}))
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
	cache Cache
}

func newEtcdDoer(namespaceEtcd *namespaceEtcd, group string, cache Cache) Doer {
	return &etcdDoer{
		namespaceEtcd: namespaceEtcd,
		group:         group,
		cache:         cache,
	}
}

func (ed *etcdDoer) Do(ctx context.Context, inputChan chan *anypb.Any, cb CollectFunc) error {
	return ed.withGroup(ctx, func(ctx context.Context, renewer *col.Renewer) error {
		var eg errgroup.Group
		prefix := path.Join(ed.group, uuid.NewWithoutDashes())
		done := make(chan struct{})
		var count int64
		ctx, cancel := pctx.WithCancel(ctx)
		defer func() {
			cancel()
			eg.Wait() //nolint:errcheck
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
				if ed.cache != nil && err == nil {
					if err := ed.cache.Put(ctx, task.Id, task.Output); err != nil {
						log.Info(ctx, "errored putting task in cache",
							zap.String("taskType", task.GetInput().GetTypeUrl()),
							zap.String("taskID", task.GetId()),
							zap.Error(err))
					}
				}

				log.Debug(ctx, "task callback starting",
					zap.String("taskType", task.GetInput().GetTypeUrl()),
					zap.String("taskID", task.GetId()),
					zap.Error(err))
				if err := cb(task.Index, task.Output, err); err != nil {
					log.Debug(ctx, "task callback errored",
						zap.String("taskType", task.GetInput().GetTypeUrl()),
						zap.String("taskID", task.GetId()),
						zap.Error(err))
					return err
				}
				log.Debug(ctx, "task callback finished ok",
					zap.String("taskType", task.GetInput().GetTypeUrl()),
					zap.String("taskID", task.GetId()))

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
				log.Info(ctx, "errored deleting tasks with the prefix", zap.String("prefix", prefix), zap.Error(err))
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

				log.Debug(ctx, "task created",
					zap.String("taskType", input.GetTypeUrl()),
					zap.String("taskID", taskID))
				if ed.cache != nil {
					output, err := ed.cache.Get(ctx, taskID)
					if err == nil {
						log.Debug(ctx, "result cached",
							zap.String("taskType", input.GetTypeUrl()),
							zap.String("taskID", taskID))
						if err := cb(index, output, nil); err != nil {
							return err
						}
						index++
						continue
					}
				}
				taskKey := path.Join(prefix, taskID)
				task := &Task{
					Id:    taskID,
					Input: input,
					State: State_RUNNING,
					Index: index,
				}
				index++
				if err := renewer.Put(ctx, taskKey, task); err != nil {
					return err
				}
				log.Debug(ctx, "task submitted",
					zap.String("taskType", input.GetTypeUrl()),
					zap.String("taskID", taskID))
				atomic.AddInt64(&count, 1)
			case <-ctx.Done():
				return errors.EnsureStack(context.Cause(ctx))
			}
		}
	})
}

func (ed *etcdDoer) withGroup(ctx context.Context, cb func(ctx context.Context, renewer *col.Renewer) error) error {
	err := ed.groupCol.WithRenewer(ctx, func(ctx context.Context, renewer *col.Renewer) error {
		key := path.Join(ed.group, uuid.NewWithoutDashes())
		defer func() {
			if _, err := col.NewSTM(ctx, ed.etcdClient, func(stm col.STM) error {
				return errors.EnsureStack(ed.groupCol.ReadWrite(stm).Delete(key))
			}); err != nil {
				log.Info(ctx, "errored deleting group key", zap.String("key", key), zap.Error(err))
			}
		}()
		if err := renewer.Put(ctx, key, &Group{}); err != nil {
			return err
		}
		err := ed.taskCol.WithRenewer(ctx, func(ctx context.Context, renewer *col.Renewer) error {
			return cb(ctx, renewer)
		})
		return errors.EnsureStack(err)
	})
	return errors.EnsureStack(err)
}

func computeTaskID(input *anypb.Any) (string, error) {
	val, err := proto.Marshal(input)
	if err != nil {
		return "", errors.EnsureStack(err)
	}
	val = append(val, []byte(version.PrettyVersion())...)
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
					return errors.EnsureStack(context.Cause(ctx))
				}
			}); err != nil && !errors.Is(context.Cause(ctx), context.Canceled) {
				log.Info(ctx, "errored in group callback", zap.String("group", group), zap.Error(err))
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
			return errors.EnsureStack(context.Cause(ctx))
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
				log.Debug(ctx, "task received",
					zap.String("taskType", task.GetInput().GetTypeUrl()),
					zap.String("taskID", task.GetId()))
				taskOutput, taskErr := cb(ctx, task.Input)
				log.Debug(ctx, "task completed",
					zap.String("taskType", task.GetInput().GetTypeUrl()),
					zap.String("taskID", task.GetId()),
					zap.Error(taskErr))

				// If the task context was canceled or the claim was lost, just return with no error.
				if errors.Is(context.Cause(ctx), context.Canceled) {
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
			if errors.Is(context.Cause(ctx), context.Canceled) ||
				col.IsErrNotFound(err) || errors.Is(err, col.ErrNotClaimed) {
				return
			}
			log.Info(ctx, "errored in task callback", zap.Error(err), zap.String("taskKey", taskKey))
		}
	}
}
