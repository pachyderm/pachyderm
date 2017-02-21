package server

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
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
	addr   string
	conn   *grpc.ClientConn
	client workerpkg.WorkerClient
	busy   bool
}

type workerPool struct {
	workerDir  string
	workers    []worker
	etcdClient *etcd.Client
}

func (w *workerPool) addWorker(addr string) error {
	for _, worker := range w.workers {
		if worker.addr == addr {
			return nil
		}
	}

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return err
	}

	w.workers = append(w.workers, worker{
		addr:   addr,
		conn:   conn,
		client: workerpkg.NewWorkerClient(conn),
	})
	return nil
}

func (w *workerPool) delWorker(addr string) error {
	for i, worker := range w.workers {
		if worker.addr == addr {
			worker.conn.Close()
			w.workers = append(w.workers[:i], w.workers[i+1:]...)
		}
	}
	return nil
}

func (w *workerPool) discoverWorkers(ctx context.Context) {
	b := backoff.NewInfiniteBackOff()
	if err := backoff.RetryNotify(func() error {
		syncer := mirror.NewSyncer(w.etcdClient, w.workerDir, 0)
		respCh, errCh := syncer.SyncBase(ctx)
		for {
			select {
			case resp := <-respCh:
				for _, kv := range resp.Kvs {
					addr := string(kv.Key)
					if err := w.addWorker(addr); err != nil {
						return err
					}
				}
			case err := <-errCh:
				return err
			}
		}
		watchCh := syncer.SyncUpdates(ctx)
		for {
			resp, ok := <-watchCh
			if !ok {
				return fmt.Errorf("watcher for prefix %s closed for unknown reasons", w.workerDir)
			}
			if err := resp.Err(); err != nil {
				return err
			}
			for _, event := range resp.Events {
				addr := string(event.Kv.Key)
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
		return nil
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
	return nil
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
	}
	return workerPool
}

func (a *apiServer) newWorkerPool(ctx context.Context, pipeline *pps.Pipeline) WorkerPool {
	wp := &workerPool{
		workerDir:  path.Join(a.etcdPrefix, WorkerEtcdPrefix, pipeline.Name),
		etcdClient: a.etcdClient,
	}
	// We need to make sure that the prefix ends with the trailing slash,
	// because
	if wp.workerDir[len(wp.workerDir)-1] != '/' {
		wp.workerDir += "/"
	}

	go wp.discoverWorkers(ctx)
	return wp
}
