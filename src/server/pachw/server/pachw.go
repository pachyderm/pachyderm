package server

import (
	"context"
	"path"
	"time"

	"go.uber.org/zap"
	autoscaling_v1 "k8s.io/api/autoscaling/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/dlock"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/middleware/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/miscutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/server/pfs/server"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/driver"
)

const (
	tickerSeconds      = 5
	scaleDownThreshold = 60 / tickerSeconds
	period             = time.Second * time.Duration(tickerSeconds)
)

type pachW struct {
	env Env
}

func NewController(ctx context.Context, env *Env) {
	controller := &pachW{
		env: *env,
	}
	go controller.run(ctx)
}

func (p *pachW) run(ctx context.Context) {
	ctx = auth.AsInternalUser(ctx, "pachw-controller")
	backoff.RetryUntilCancel(ctx, func() (retErr error) { //nolint:errcheck
		lock := dlock.NewDLock(p.env.EtcdClient, path.Join(p.env.EtcdPrefix, "pachw-controller-lock"))
		ctx, err := lock.Lock(ctx)
		if err != nil {
			return errors.Wrap(err, "locking pachw-controller lock")
		}
		defer errors.Invoke1(&retErr, lock.Unlock, ctx, "error unlocking")
		var replicas int
		var scaleDownCount int
		ticker := time.NewTicker(period)
		defer ticker.Stop()
		for {
			numTasks, err := p.countTasks(ctx, []string{server.URLTaskNamespace, server.StorageTaskNamespace, driver.PreprocessingTaskNamespace(nil)})
			if err != nil {
				return errors.Wrap(err, "error counting tasks")
			}
			scaleDownCount++
			if numTasks > replicas/2 {
				scaleDownCount = 0
			}
			if numTasks > replicas {
				replicas = miscutil.Min(numTasks, p.env.MaxReplicas)
				if err := p.scalePachw(ctx, int32(replicas)); err != nil {
					return errors.Wrap(err, "error scaling up pachw")
				}
			} else if scaleDownCount == scaleDownThreshold {
				replicas = miscutil.Max(replicas/2, p.env.MinReplicas)
				scaleDownCount = 0
				if err := p.scalePachw(ctx, int32(replicas)); err != nil {
					return errors.Wrap(err, "error scaling down pachw")
				}
			}
			select {
			case <-ticker.C:
			case <-ctx.Done():
				return errors.EnsureStack(context.Cause(ctx))
			}
		}
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		log.Error(ctx, "error in pachw run; will retry", zap.Error(err), zap.Duration("retryAfter", d))
		return nil
	})
}

func (p *pachW) countTasks(ctx context.Context, namespaces []string) (int, error) {
	totalTasks := 0
	for _, ns := range namespaces {
		tasks, _, err := task.Count(ctx, p.env.TaskService, ns, "")
		if err != nil {
			return 0, errors.EnsureStack(err)
		}
		totalTasks += int(tasks)
	}
	return totalTasks, nil
}

func (p *pachW) scalePachw(ctx context.Context, replicas int32) error {
	// We call scalePachw and tell k8s the desired replica count regardless of whether or not
	// there's a change in the number of desired replicas.  That way, we can't get out of sync
	// between a read and a write; k8s always knows what scale we want at the same time we know
	// we want it.  To cut down on log spam, this message is logged at DEBUG, but if the replica
	// count changes between the current actual replica count and our new desired replica count,
	// we log that at level INFO below.
	log.Debug(ctx, "scaling pachw", zap.Int32("replicas", replicas))

	scale, err := p.env.KubeClient.AppsV1().Deployments(p.env.Namespace).UpdateScale(
		ctx, "pachw", &autoscaling_v1.Scale{
			ObjectMeta: meta_v1.ObjectMeta{
				Name:      "pachw",
				Namespace: p.env.Namespace,
			},
			Spec: autoscaling_v1.ScaleSpec{
				Replicas: replicas,
			},
		}, meta_v1.UpdateOptions{
			FieldManager: "pachw-controller",
		},
	)
	if err != nil {
		return errors.Wrap(err, "could not scale pachw deployment")
	}

	// Report the change.
	var existingReplicas int32
	if scale != nil {
		existingReplicas = scale.Status.Replicas
	}
	if replicas != existingReplicas {
		log.Info(ctx, "scaled pachw ok", zap.Int32("replicas", replicas), zap.Int32("observedReplicas", existingReplicas))
	} else {
		log.Debug(ctx, "scaled pachw ok", zap.Int32("replicas", replicas), zap.Int32("observedReplicas", existingReplicas))
	}
	return nil
}
