package server

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/sirupsen/logrus"
	etcd "go.etcd.io/etcd/client/v3"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/dlock"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/middleware/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	authserver "github.com/pachyderm/pachyderm/v2/src/server/auth"
	ppsserver "github.com/pachyderm/pachyderm/v2/src/server/pps"
)

// Env is the dependencies needed to run the PFS API server
type Env struct {
	ObjectClient obj.Client
	DB           *pachsql.DB
	EtcdPrefix   string
	EtcdClient   *etcd.Client
	TaskService  task.Service
	TxnEnv       *txnenv.TransactionEnv
	Listener     col.PostgresListener
	KubeClient   kubernetes.Interface
	Namespace    string

	AuthServer authserver.APIServer
	// TODO: a reasonable repo metadata solution would let us get rid of this circular dependency
	// permissions might also work.
	GetPPSServer func() ppsserver.APIServer
	// TODO: remove this, the load tests need a pachClient
	GetPachClient func(ctx context.Context) *client.APIClient

	BackgroundContext context.Context
	StorageConfig     serviceenv.StorageConfiguration
	Logger            *logrus.Logger
}

func EnvFromServiceEnv(env serviceenv.ServiceEnv, txnEnv *txnenv.TransactionEnv) (*Env, error) {
	// Setup etcd, object storage, and database clients.
	objClient, err := obj.NewClient(env.Config().StorageBackend, env.Config().StorageRoot)
	if err != nil {
		return nil, err
	}
	etcdPrefix := path.Join(env.Config().EtcdPrefix, env.Config().PFSEtcdPrefix)
	if env.AuthServer() == nil {
		panic("auth server cannot be nil")
	}
	return &Env{
		ObjectClient:  objClient,
		DB:            env.GetDBClient(),
		TxnEnv:        txnEnv,
		EtcdPrefix:    etcdPrefix,
		EtcdClient:    env.GetEtcdClient(),
		TaskService:   env.GetTaskService(etcdPrefix),
		KubeClient:    env.GetKubeClient(),
		Namespace:     env.Config().Namespace,
		GetPachClient: env.GetPachClient,

		BackgroundContext: env.Context(),
		StorageConfig:     env.Config().StorageConfiguration,
		Logger:            env.Logger(),
	}, nil
}

type Pachw struct {
	env        Env
	log        *logrus.Logger
	etcdClient *etcd.Client
	txnEnv     *txnenv.TransactionEnv
	prefix     string
}

func NewMaster(ctx context.Context, env *Env) *Pachw {
	master := &Pachw{
		env:        *env,
		etcdClient: env.EtcdClient,
		log:        env.Logger,
		txnEnv:     env.TxnEnv,
		prefix:     env.EtcdPrefix,
	}
	go master.master(ctx)
	return master
}

func (p *Pachw) master(ctx context.Context) {
	ctx = auth.AsInternalUser(ctx, "pfs-master")
	backoff.RetryUntilCancel(ctx, func() error { //nolint:errcheck
		eg, ctx := errgroup.WithContext(ctx)
		eg.Go(func() error {
			lock := dlock.NewDLock(p.etcdClient, path.Join(p.prefix, "pachw-master-lock"))
			defer func() {
				if err := lock.Unlock(ctx); err != nil {
					p.log.Errorf("error unlocking in pfs master: %v", err)
				}
			}()
			tickerSeconds := 5
			samplesPerMinute := 60 / tickerSeconds
			tasksPerMinute := float32(0)
			prevTotalInstances := 0
			var taskSamples []int
			ticker := time.NewTicker(time.Second * time.Duration(tickerSeconds))
			defer ticker.Stop()
			for {
				totalTasks := p.countTasks(ctx, []string{"storage", "url"})
				taskSamples, tasksPerMinute = p.taskAverage(taskSamples, totalTasks, samplesPerMinute)
				fmt.Printf("samples: %v avg: %v\n", taskSamples, tasksPerMinute)
				totalInstances := p.countPipelineInstances(ctx)
				if tasksPerMinute == 0 {
					err := p.scalePachw(ctx, 0)
					if err != nil {
						p.log.Errorf("failed to scale down pachw: %v", err)
					}
				} else if totalTasks > 0 || totalInstances != prevTotalInstances {
					err := p.scalePachw(ctx, int32(totalInstances))
					if err != nil {
						p.log.Errorf("failed to scale up pachw: %v", err)
					}
				}
				prevTotalInstances = totalInstances
				select {
				case <-ctx.Done():
					fmt.Printf("fahad: terminating master pachw")
					return errors.EnsureStack(ctx.Err())
				case <-ticker.C:
				}
			}
		})
		return errors.EnsureStack(eg.Wait())
	}, backoff.NewInfiniteBackOff(), func(err error, _ time.Duration) error {
		fmt.Printf("fahad: terminating master pachw")
		p.log.Errorf("error in pachw master: %v", err)
		return nil
	})
}

func (p *Pachw) taskAverage(taskSamples []int, tasks, numSamples int) ([]int, float32) {
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

// todo: Update logic to iterate through projects x pipelines, when project logic gets merged.
func (p *Pachw) countPipelineInstances(ctx context.Context) int {
	pachClient := p.env.GetPachClient(ctx)
	infos, err := pachClient.ListPipeline(true)
	if err != nil {
		p.log.Errorf("error listing pipelines: %v", err)
	}
	totalInstances := 0
	for _, info := range infos {
		totalInstances += int(info.Parallelism)
	}
	return totalInstances
}

func (p *Pachw) countTasks(ctx context.Context, namespaces []string) int {
	totalTasks := 0
	for _, ns := range namespaces {
		tasks, _, err := task.Count(ctx, p.env.TaskService, ns, "")
		if err != nil {
			p.log.Errorf("error counting tasks: %v", err)
		}
		totalTasks += int(tasks)
	}
	return totalTasks
}

func (p *Pachw) scalePachw(ctx context.Context, replicas int32) error {
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
