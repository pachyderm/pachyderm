package server

import (
	"context"
	"path"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/dlock"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/middleware/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/miscutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/server/pfs/server"
)

const (
	tickerSeconds      = 5
	scaleDownThreshold = 60 / tickerSeconds
	period             = time.Second * time.Duration(tickerSeconds)
)

type pachW struct {
	log *logrus.Logger
	env Env
}

func NewController(ctx context.Context, env *Env) {
	controller := &pachW{
		env: *env,
		log: env.Logger,
	}
	go controller.run(ctx)
}

func (p *pachW) run(ctx context.Context) {
	ctx = auth.AsInternalUser(ctx, "pachw-controller")
	backoff.RetryUntilCancel(ctx, func() (retErr error) { //nolint:errcheck
		lock := dlock.NewDLock(p.env.EtcdClient, path.Join(p.env.EtcdPrefix, "pachw-controller-lock"))
		defer func() {
			if err := lock.Unlock(ctx); err != nil {
				retErr = multierror.Append(retErr, errors.Wrap(err, "error unlocking"))
			}
		}()
		var replicas int
		var scaleDownCount int
		ticker := time.NewTicker(period)
		defer ticker.Stop()
		for {
			numTasks, err := p.countTasks(ctx, []string{server.URLTaskNamespace, server.StorageTaskNamespace})
			if err != nil {
				return errors.Wrap(err, "error counting tasks")
			}
			scaleDownCount++
			if numTasks > replicas/2 {
				scaleDownCount = 0
			}
			if numTasks > replicas {
				replicas = miscutil.Min(numTasks, p.env.MaxReplicas)
				p.log.Infof("scaling up pachw workers to %d replicas", replicas)
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
				return errors.EnsureStack(ctx.Err())
			}
		}
	}, backoff.NewInfiniteBackOff(), func(err error, _ time.Duration) error {
		p.log.Errorf("error in pachw run: %v", err)
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
	pachwRs, err := p.env.KubeClient.AppsV1().Deployments(p.env.Namespace).Get(ctx,
		"pachw", v1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "could not get pachw deployment")
	}
	pachwRs.Spec.Replicas = &replicas
	_, err = p.env.KubeClient.AppsV1().Deployments(p.env.Namespace).Update(
		ctx,
		pachwRs,
		v1.UpdateOptions{},
	)
	if err != nil {
		return errors.Wrap(err, "could not scale pachw deployment")
	}
	return nil
}
