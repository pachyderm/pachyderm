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
	DataCh() chan datumSet
}

type worker struct {
	addr   string
	client workerpkg.WorkerClient
	ctx    context.Context
}

// An input/output pair for a single datum. When a worker has finished
// processing 'data', it writes the resulting hashtree to 'resp' (each job has
// its own response channel)
type datumSet struct {
	data []*pfs.FileInfo
	resp chan hashtree.HashTree
}

type workerPool struct {
	dataCh chan *datumSet

	// Parent of all worker contexts (see workersMap)
	ctx context.Context

	// The prefix in etcd where new workers can be discovered
	workerDir string

	// Map of worker address to cancel fn -- call fn to kill worker goroutine
	workersMap map[string]context.CancelFunc

	// Used to check for workers added/deleted in etcd
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

func (wp *workerPool) addWorker(addr string) error {
	if _, ok := wp.workersMap[addr]; ok {
		return fmt.Errorf("worker already exists at %s", addr)
	}

	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithTimeout(5*time.Second))
	if err != nil {
		return err
	}
	childCtx, cancelFn := context.WithCancel(wp.ctx)
	wp.workersMap[addr] = cancelFn

	wr := &worker{
		addr:   addr,
		client: workerpkg.NewWorkerClient(),
		ctx:    childCtx,
	}
	go w.run(wp.datumCh)
}

func (w *worker) run(datumCh chan *datumSet) (retErr error) {
	var data []*pfs.FileInfo
	for {
		datumSet = <-datumCh
		if datumSet != nil {
			return
		}
		resp, err := w.client.Process(w.ctx, &workerpkg.ProcessRequest{
			Data: datumSet.data,
		})
		if err != nil {
			datumCh <- data
			if err == context.Cancelled {
				return err
			}
			protolion.Errorf("worker request to %s failed with error %s", w.addr, retErr)
		}
		// tree := ?
		datumSet.resp <- hashtree.NewHashTree()
	}
}

func (wp *workerPool) DataCh() chan datumSet {
	return wp.dataCh
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

func (a *apiServer) newWorkerPool(ctx context.Context, pipeline *pps.Pipeline) WorkerPool {
	wp := &workerPool{
		ctx:        ctx,
		dataCh:     make(chan datumSet),
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
