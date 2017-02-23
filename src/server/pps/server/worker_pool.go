package server

import (
	"context"
	"fmt"
	"path"
	"sync"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
	workerpkg "github.com/pachyderm/pachyderm/src/server/pkg/worker"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/mirror"
	"go.pedge.io/lion/proto"
	"google.golang.org/grpc"
)

const (
	WorkerEtcdPrefix = "workers"
)

type WorkerPool interface {
	// Process blocks until a worker accepts the given datum set.
	Submit(datumSet []*pfs.FileInfo) error
	// Wait blocks until all datum sets that have been submitted up
	// until this point have been processed.
	Wait() error
}

type worker struct {
	sync.Mutex
	addr    string
	conn    *grpc.ClientConn
	client  workerpkg.WorkerClient
	removed bool
}

type workerPool struct {
	ctx        context.Context
	workerDir  string
	workers    chan *worker
	workersMap map[string]*worker
	// workersLock sync.Mutex
	etcdClient *etcd.Client
}

func (w *workerPool) discoverWorkers(ctx context.Context) {
	b := backoff.NewInfiniteBackOff()
	if err := backoff.RetryNotify(func() error {
		syncer := mirror.NewSyncer(w.etcdClient, w.workerDir, 0)
		respCh, errCh := syncer.SyncBase(ctx)
	getBaseWorkers:
		for {
			select {
			case resp, ok := <-respCh:
				if !ok {
					break getBaseWorkers
				}
				for _, kv := range resp.Kvs {
					addr := path.Base(string(kv.Key))
					if err := w.addWorker(addr); err != nil {
						return err
					}
				}
			case err := <-errCh:
				if err != nil {
					return err
				}
			}
		}
		watchCh := syncer.SyncUpdates(ctx)
		protolion.Infof("watching `%s` for workers", w.workerDir)
		for {
			resp, ok := <-watchCh
			if !ok {
				return fmt.Errorf("watcher for prefix %s closed for unknown reasons", w.workerDir)
			}
			if err := resp.Err(); err != nil {
				return err
			}
			for _, event := range resp.Events {
				addr := path.Base(string(event.Kv.Key))
				switch event.Type {
				case etcd.EventTypePut:
					if err := w.addWorker(addr); err != nil {
						return err
					}
				case etcd.EventTypeDelete:
					if err := w.delWorker(addr); err != nil {
						return err
					}
				}
			}
		}
		panic("unreachable")
	}, b, func(err error, d time.Duration) error {
		if err == context.Canceled {
			return err
		}
		protolion.Errorf("error discovering workers: %v; retrying in %v", err, d)
		return nil
	}); err != context.Canceled {
		panic(fmt.Sprintf("the retry loop should not exit with a non-context-cancelled error: %v", err))
	}
}

func (w *workerPool) Submit(datumSet []*pfs.FileInfo) error {
	var wr *worker
	for {
		wr := <-w.workers
		wr.Lock()
		defer wr.Unlock()
		if !wr.removed {
			break
		}
	}
	if wr == nil {
		protolion.Debugf("workers channel returns a nil worker; this is likely a bug")
		return nil
	}
	go func() retErr {
		defer func() {
			if retErr != nil {
				protolion.Errorf("error processing datum %v: %v", datumSet, retErr)
			}
			w.workers <- wr
		}()
		_, err := wr.client.Process(w.ctx, workerpkg.ProcessRequest{
			Data: datumSet,
		})
		if err != nil {
			go w.Submit(datumSet)
			return err
		}
	}()
}

func (w *workerPool) Wait() error {
	return nil
}

func (a *apiServer) workerPool(ctx context.Context, pipeline *pps.Pipeline) WorkerPool {
	a.workerPoolsLock.Lock()
	defer a.workerPoolsLock.Unlock()
	workerPool, ok := a.workerPools[pipeline.Name]
	if !ok {
		workerPool = a.newWorkerPool(ctx, pipeline)
		a.workerPools[pipeline.Name] = workerPool
	}
	return workerPool
}

func (a *apiServer) newJob(ctx context.Context, pipeline *pps.Pipeline) (chan<- []*pfs.FileInfo, <-chan hashtree.HashTree) {

}

func (a *apiServer) newWorkerPool(ctx context.Context, pipeline *pps.Pipeline) WorkerPool {
	wp := &workerPool{
		ctx:        ctx,
		workerDir:  path.Join(a.etcdPrefix, WorkerEtcdPrefix, pipeline.Name),
		etcdClient: a.etcdClient,
		workers:    make(chan worker, 1000),
	}
	// We need to make sure that the prefix ends with the trailing slash,
	// because
	if wp.workerDir[len(wp.workerDir)-1] != '/' {
		wp.workerDir += "/"
	}

	go wp.discoverWorkers(ctx)
	return wp
}
