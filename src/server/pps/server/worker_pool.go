package server

import (
	"bytes"
	"context"
	"fmt"
	"path"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
	workerpkg "github.com/pachyderm/pachyderm/src/server/pkg/worker"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/mirror"
	"go.pedge.io/lion/proto"
	"google.golang.org/grpc"
)

const (
	workerEtcdPrefix = "workers"
)

// WorkerPool represents a pool of workers that can be used to process datums.
type WorkerPool interface {
	DataCh() chan datumAndResp
}

type worker struct {
	ctx          context.Context
	addr         string
	workerClient workerpkg.WorkerClient
	pachClient   *client.APIClient
}

func (w *worker) run(dataCh chan datumAndResp) {
	for {
		dr, ok := <-dataCh
		if !ok {
			return
		}
		resp, err := w.workerClient.Process(w.ctx, &workerpkg.ProcessRequest{
			Data: dr.datum,
		})
		if err != nil {
			dataCh <- dr
			if err == context.Canceled {
				return
			} else if err != nil {
				protolion.Errorf("worker request to %s failed with error %s", w.addr, err)
			}
			continue
		}
		if resp.Tag != nil {
			var buffer bytes.Buffer
			if err := w.pachClient.GetTag(resp.Tag.Name, &buffer); err != nil {
				protolion.Errorf("failed to retrieve hashtree after worker %s has ostensibly processed the datum %v", w.addr, dr.datum)
				dataCh <- dr
				continue
			}
			tree, err := hashtree.Deserialize(buffer.Bytes())
			if err != nil {
				panic(err)
			}
			dr.respCh <- tree
		} else {
			dr.errCh <- resp.Log
		}
	}
}

// An input/output pair for a single datum. When a worker has finished
// processing 'data', it writes the resulting hashtree to 'resp' (each job has
// its own response channel)
type datumAndResp struct {
	datum  []*pfs.FileInfo
	respCh chan hashtree.HashTree
	errCh  chan string
}

type workerPool struct {
	dataCh chan datumAndResp

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
					if err := <-errCh; err != nil {
						return err
					}
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

func (w *workerPool) addWorker(addr string) error {
	if cancel, ok := w.workersMap[addr]; ok {
		cancel()
	}

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", addr, client.PPSWorkerPort), grpc.WithInsecure(), grpc.WithTimeout(5*time.Second))
	if err != nil {
		return err
	}
	childCtx, cancelFn := context.WithCancel(w.ctx)
	w.workersMap[addr] = cancelFn

	pachClient, err := client.NewInCluster()
	if err != nil {
		return err
	}

	wr := &worker{
		ctx:          childCtx,
		addr:         addr,
		workerClient: workerpkg.NewWorkerClient(conn),
		pachClient:   pachClient,
	}
	protolion.Infof("launching new worker at %v", addr)
	go wr.run(w.dataCh)
	return nil
}

func (w *workerPool) delWorker(addr string) error {
	cancel, ok := w.workersMap[addr]
	if !ok {
		return fmt.Errorf("deleting worker %s which is not in worker pool", addr)
	}
	cancel()
	return nil
}

func (w *workerPool) DataCh() chan datumAndResp {
	return w.dataCh
}

// workerPool fetches the worker pool associated with 'id', or creates one if
// none exists.
func (a *apiServer) workerPool(ctx context.Context, id string) WorkerPool {
	a.workerPoolsLock.Lock()
	defer a.workerPoolsLock.Unlock()
	workerPool, ok := a.workerPools[id]
	if !ok {
		workerPool = a.newWorkerPool(ctx, id)
		a.workerPools[id] = workerPool
	}
	return workerPool
}

// newWorkerPool generates a new worker pool for the job or pipeline identified
// with 'id'.  Each 'id' used to create a new worker pool must correspond to
// a unique binary (in other words, all workers in the worker pool for 'id'
// will be running the same user binary)
func (a *apiServer) newWorkerPool(ctx context.Context, id string) WorkerPool {
	wp := &workerPool{
		ctx:        ctx,
		dataCh:     make(chan datumAndResp),
		workerDir:  path.Join(a.etcdPrefix, workerEtcdPrefix, id),
		workersMap: make(map[string]context.CancelFunc),
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

func (a *apiServer) delWorkerPool(id string) {
	a.workerPoolsLock.Lock()
	defer a.workerPoolsLock.Unlock()
	delete(a.workerPools, id)
}
