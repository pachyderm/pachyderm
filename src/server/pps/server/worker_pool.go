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
	"github.com/pachyderm/pachyderm/src/server/pkg/watch"
	workerpkg "github.com/pachyderm/pachyderm/src/server/pkg/worker"

	etcd "github.com/coreos/etcd/clientv3"
	"go.pedge.io/lion/proto"
	"google.golang.org/grpc"
)

const (
	workerEtcdPrefix = "workers"
)

// An input/output pair for a single datum. When a worker has finished
// processing 'data', it writes the resulting hashtree to 'resp' (each job has
// its own response channel)
type datumAndResp struct {
	ctx    context.Context
	jobID  string // This is passed to workers, so they can annotate their logs
	datum  []*pfs.FileInfo
	respCh chan hashtree.HashTree
	errCh  chan struct{}
	retCh  chan *datumAndResp
}

// WorkerPool represents a pool of workers that can be used to process datums.
type WorkerPool interface {
	DataCh() chan *datumAndResp
}

type worker struct {
	ctx          context.Context
	addr         string
	workerClient workerpkg.WorkerClient
	pachClient   *client.APIClient
}

func (w *worker) run(dataCh chan *datumAndResp) {
	defer func() {
		protolion.Infof("goro for worker %s is exiting", w.addr)
	}()
	returnDatum := func(dr *datumAndResp) {
		select {
		case dr.retCh <- dr:
		case <-dr.ctx.Done():
		}
	}
	for {
		dr, ok := <-dataCh
		if !ok {
			protolion.Errorf("worker %s exiting", w.addr)
			return
		}
		resp, err := w.workerClient.Process(dr.ctx, &workerpkg.ProcessRequest{
			JobID: dr.jobID,
			Data:  dr.datum,
		})
		if err != nil {
			protolion.Errorf("worker %s failed to process datum %v with error %s", w.addr, dr.datum, err)
			returnDatum(dr)
			continue
		}
		if resp.Tag != nil {
			var buffer bytes.Buffer
			if err := w.pachClient.GetTag(resp.Tag.Name, &buffer); err != nil {
				protolion.Errorf("failed to retrieve hashtree after worker %s has ostensibly processed the datum %v: %v", w.addr, dr.datum, err)
				returnDatum(dr)
				continue
			}
			tree, err := hashtree.Deserialize(buffer.Bytes())
			if err != nil {
				protolion.Errorf("failed to serialize hashtree after worker %s has ostensibly processed the datum %v; this is likely a bug: %v", w.addr, dr.datum, err)
				returnDatum(dr)
				continue
			}
			dr.respCh <- tree
		} else if resp.Failed {
			close(dr.errCh)
		} else {
			protolion.Errorf("unrecognized response from worker %s when processing datum %v; this is likely a bug", w.addr, dr.datum)
			returnDatum(dr)
			continue
		}
	}
}

type workerPool struct {
	// Worker pool recieves work via this channel
	dataCh chan *datumAndResp

	// The prefix in etcd where new workers can be discovered
	workerDir string

	// Map of worker address to cancel fn -- call fn to kill worker goroutine
	workersMap map[string]context.CancelFunc

	// Used to check for workers added/deleted in etcd
	etcdClient *etcd.Client
}

func (w *workerPool) discoverWorkers() {
	b := backoff.NewInfiniteBackOff()
	backoff.RetryNotify(func() error {
		protolion.Infof("watching `%s` for workers", w.workerDir)
		watcher, err := watch.NewWatcher(context.Background(), w.etcdClient, w.workerDir)
		if err != nil {
			return err
		}
		defer watcher.Close()
		for {
			resp, ok := <-watcher.Watch()
			if !ok {
				return fmt.Errorf("watcher closed for unknown reasons")
			}
			if err := resp.Err; err != nil {
				return err
			}
			addr := path.Base(string(resp.Key))
			switch resp.Type {
			case watch.EventPut:
				if err := w.addWorker(addr); err != nil {
					return err
				}
			case watch.EventDelete:
				if err := w.delWorker(addr); err != nil {
					return err
				}
			}
		}
		panic("unreachable")
	}, b, func(err error, d time.Duration) error {
		protolion.Errorf("error discovering workers for %v: %v; retrying in %v", w.workerDir, err, d)
		return nil
	})
}

func (w *workerPool) addWorker(addr string) error {
	if cancel, ok := w.workersMap[addr]; ok {
		cancel()
	}

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", addr, client.PPSWorkerPort), grpc.WithInsecure(), grpc.WithTimeout(5*time.Second))
	if err != nil {
		return err
	}
	childCtx, cancelFn := context.WithCancel(context.Background())
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
	protolion.Infof("launching new worker for %s at %v", w.workerDir, addr)
	go wr.run(w.dataCh)
	return nil
}

func (w *workerPool) delWorker(addr string) error {
	cancel, ok := w.workersMap[addr]
	if !ok {
		return fmt.Errorf("deleting worker %s which is not in worker pool", addr)
	}
	cancel()
	protolion.Infof("deleting worker for %s at %v", w.workerDir, addr)
	return nil
}

func (w *workerPool) DataCh() chan *datumAndResp {
	return w.dataCh
}

// workerPool fetches the worker pool associated with 'id', or creates one if
// none exists.
func (a *apiServer) workerPool(id string) WorkerPool {
	a.workerPoolsLock.Lock()
	defer a.workerPoolsLock.Unlock()
	workerPool, ok := a.workerPools[id]
	if !ok {
		workerPool = a.newWorkerPool(id)
		a.workerPools[id] = workerPool
	}
	return workerPool
}

// newWorkerPool generates a new worker pool for the job or pipeline identified
// with 'id'.  Each 'id' used to create a new worker pool must correspond to
// a unique binary (in other words, all workers in the worker pool for 'id'
// will be running the same user binary)
func (a *apiServer) newWorkerPool(id string) WorkerPool {
	wp := &workerPool{
		dataCh:     make(chan *datumAndResp),
		workerDir:  path.Join(a.etcdPrefix, workerEtcdPrefix, id),
		workersMap: make(map[string]context.CancelFunc),
		etcdClient: a.etcdClient,
	}
	// We need to make sure that the prefix ends with the trailing slash,
	// because
	if wp.workerDir[len(wp.workerDir)-1] != '/' {
		wp.workerDir += "/"
	}

	go wp.discoverWorkers()
	return wp
}

func (a *apiServer) delWorkerPool(id string) {
	a.workerPoolsLock.Lock()
	defer a.workerPoolsLock.Unlock()
	delete(a.workerPools, id)
}
