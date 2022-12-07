package server

import (
	"context"
	"path"
	"time"

	"golang.org/x/sync/errgroup"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/dlock"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/middleware/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/server/pfs/server"
)

const (
	backOffPeriod    = 30 // backOffPeriod is how many seconds the master waits between scale down operations.
	tickerSeconds    = 5
	samplesPerMinute = 60 / tickerSeconds
)

type pachW struct {
	env Env
}

func NewMaster(ctx context.Context, env *Env) {
	master := &pachW{
		env: *env,
	}
	go master.master(ctx)
}

func (p *pachW) master(ctx context.Context) {
	ctx = auth.AsInternalUser(ctx, "pfs-master")
	backoff.RetryUntilCancel(ctx, func() error { //nolint:errcheck
		eg, ctx := errgroup.WithContext(ctx)
		eg.Go(func() error {
			lock := dlock.NewDLock(p.env.EtcdClient, path.Join(p.env.EtcdPrefix, "pachw-master-lock"))
			defer func() {
				if err := lock.Unlock(ctx); err != nil {
					p.env.Logger.Errorf("error unlocking in pfs master: %v", err)
				}
			}()

			scaleBackoff := 0 // scaleBackoff tracks the time until the next scale down operation can occur.
			tasksPerMinute := float32(0)
			var taskSamples []int
			ticker := time.NewTicker(time.Second * time.Duration(tickerSeconds))
			defer ticker.Stop()
			for {
				if scaleBackoff > 0 {
					scaleBackoff -= tickerSeconds
				}
				totalTasks := p.countTasks(ctx, []string{server.URLTaskNamespace, server.StorageTaskNamespace})
				taskSamples, tasksPerMinute = p.taskAverage(taskSamples, totalTasks, samplesPerMinute)
				if tasksPerMinute == 0 && scaleBackoff <= 0 {
					err := p.exponentialPachwScaleDown(ctx)
					if err != nil {
						p.env.Logger.Errorf("failed to scale down pachw: %v", err)
					}
					scaleBackoff = backOffPeriod
				} else if totalTasks > 0 {
					totalInstances := totalTasks
					if p.env.MaxReplicas != 0 && totalInstances > p.env.MaxReplicas {
						totalInstances = p.env.MaxReplicas
					}
					err := p.scalePachw(ctx, int32(totalInstances))
					if err != nil {
						p.env.Logger.Errorf("failed to scale up pachw: %v", err)
					}
				}
				select {
				case <-ctx.Done():
					return errors.EnsureStack(ctx.Err())
				case <-ticker.C:
				}
			}
		})
		return errors.EnsureStack(eg.Wait())
	}, backoff.NewInfiniteBackOff(), func(err error, _ time.Duration) error {
		p.env.Logger.Errorf("error in pachw master: %v", err)
		return nil
	})
}

func (p *pachW) taskAverage(taskSamples []int, tasks, numSamples int) ([]int, float32) {
	taskSamples = append(taskSamples, tasks)
	if len(taskSamples) > numSamples {
		taskSamples = taskSamples[1:]
	}
	totalTasksLastMinute := 0
	for _, taskCount := range taskSamples {
		totalTasksLastMinute += taskCount
	}
	return taskSamples, float32(totalTasksLastMinute) / float32(numSamples)
}

func (p *pachW) countTasks(ctx context.Context, namespaces []string) int {
	totalTasks := 0
	for _, ns := range namespaces {
		tasks, _, err := task.Count(ctx, p.env.TaskService, ns, "")
		if err != nil {
			p.env.Logger.Errorf("error counting tasks: %v", err)
		}
		totalTasks += int(tasks)
	}
	return totalTasks
}

func (p *pachW) exponentialPachwScaleDown(ctx context.Context) error {
	pachwRs, err := p.env.KubeClient.AppsV1().ReplicaSets(p.env.Namespace).Get(ctx,
		"pachw", v1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "could not get pachw replicaSet")
	}
	replicas := *pachwRs.Spec.Replicas
	replicas /= 2
	if replicas < int32(p.env.MinReplicas) {
		replicas = int32(p.env.MinReplicas)
	}
	*pachwRs.Spec.Replicas = replicas
	_, err = p.env.KubeClient.AppsV1().ReplicaSets(p.env.Namespace).Update(
		ctx,
		pachwRs,
		v1.UpdateOptions{},
	)
	if err != nil {
		return errors.Wrap(err, "could not scale pachw replicaSet")
	}
	return nil
}

func (p *pachW) scalePachw(ctx context.Context, replicas int32) error {
	pachwRs, err := p.env.KubeClient.AppsV1().ReplicaSets(p.env.Namespace).Get(ctx,
		"pachw", v1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "could not get pachw replicaSet")
	}
	pachwRs.Spec.Replicas = &replicas
	_, err = p.env.KubeClient.AppsV1().ReplicaSets(p.env.Namespace).Update(
		ctx,
		pachwRs,
		v1.UpdateOptions{},
	)
	if err != nil {
		return errors.Wrap(err, "could not scale pachw replicaSet")
	}
	return nil
}
