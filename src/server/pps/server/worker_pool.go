package server

import (
	"bytes"
	"context"
	"fmt"
	"path"
	"sync"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
	"github.com/pachyderm/pachyderm/src/server/pkg/watch"
	workerpkg "github.com/pachyderm/pachyderm/src/server/pkg/worker"

	etcd "github.com/coreos/etcd/clientv3"
	"go.pedge.io/lion/proto"
	"google.golang.org/grpc"
	"k8s.io/kubernetes/pkg/api"
	kube "k8s.io/kubernetes/pkg/client/unversioned"
)

const (
	workerEtcdPrefix = "workers"
	maxBackoff       = 5 * time.Second
)

type datum struct {
	files   []*pfs.FileInfo
	retries int
}

// WorkerPool represents a pool of workers that can be used to process datums.
type WorkerPool interface {
	// Send datums to this channel to be processed
	DataCh() chan<- *datum
	// Receive datums that failed to be processed from this channel
	FailCh() <-chan *datum
	// Receive hashtrees of the outputs of successfully processing datums
	SuccessCh() <-chan hashtree.HashTree
}

type workerPool struct {
	// When this context is canceled, the worker pool should clean up all
	// its resources.
	ctx context.Context
	// The prefix in etcd where new workers can be discovered
	workerDir string
	// workersMap is a map from a worker's address to the function that
	// can be used to release its resources.
	workersMap     map[string]worker
	workersMapLock sync.Mutex
	// objClient is the client for Pachyderm's object store
	objClient pfs.ObjectAPIClient
	// Used to check for workers added/deleted in etcd
	etcdClient *etcd.Client
	// Used to delete worker pods
	kubeClient *kube.Client
	namespace  string
	// The job that spawned the worker pool
	jobID string
	// workers get datums from this channel.
	dataCh chan *datum
	// workers send datums to this channel when they fail to process
	// the datums.
	failCh chan *datum
	// workers send the hashtrees of the outputs of processing datums to
	// this channel.
	successCh chan hashtree.HashTree
}

func (w *workerPool) discoverWorkers() {
	b := backoff.NewInfiniteBackOff()
	backoff.RetryNotify(func() error {
		protolion.Infof("watching `%s` for workers", w.workerDir)
		watcher, err := watch.NewWatcher(w.ctx, w.etcdClient, w.workerDir)
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
			podName := string(resp.Value)
			switch resp.Type {
			case watch.EventPut:
				if err := w.addWorker(addr, podName); err != nil {
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
		select {
		case <-w.ctx.Done():
			// Exit the retry loop if context got cancelled
			return err
		default:
		}
		protolion.Errorf("error discovering workers for %v: %v; retrying in %v", w.workerDir, err, d)
		return nil
	})
}

type worker struct {
	cancel  context.CancelFunc
	podName string
}

func (w *workerPool) addWorker(addr string, podName string) error {
	w.workersMapLock.Lock()
	defer w.workersMapLock.Unlock()

	if worker, ok := w.workersMap[addr]; ok {
		worker.cancel()
	}

	workerCtx, cancelFn := context.WithCancel(w.ctx)
	w.workersMap[addr] = worker{
		cancel:  cancelFn,
		podName: podName,
	}

	protolion.Infof("launching new worker for %s at %v", w.workerDir, addr)
	go w.runWorker(workerCtx, addr)
	return nil
}

func (w *workerPool) delWorker(addr string) error {
	w.workersMapLock.Lock()
	defer w.workersMapLock.Unlock()

	worker, ok := w.workersMap[addr]
	if !ok {
		return fmt.Errorf("deleting worker %s which is not in worker pool", addr)
	}

	worker.cancel()
	zeroVal := int64(0)
	if err := w.kubeClient.Pods(w.namespace).Delete(worker.podName, &api.DeleteOptions{
		GracePeriodSeconds: &zeroVal,
	}); err != nil {
		return err
	}
	protolion.Infof("deleting worker for %s at %v", w.workerDir, addr)

	return nil
}

func (w *workerPool) runWorker(ctx context.Context, addr string) {
	defer func() {
		protolion.Infof("goro for worker %s for job %s is exiting", addr, w.jobID)
	}()

	var workerClient workerpkg.WorkerClient
	b := backoff.NewInfiniteBackOff()
	backoff.RetryNotify(func() error {
		conn, err := grpc.DialContext(ctx, fmt.Sprintf("%s:%d", addr, client.PPSWorkerPort), grpc.WithInsecure())
		if err != nil {
			return err
		}
		workerClient = workerpkg.NewWorkerClient(conn)
		return nil
	}, b, func(err error, d time.Duration) error {
		select {
		case <-w.ctx.Done():
			// Exit the retry loop if context got cancelled
			return err
		default:
		}
		protolion.Infof("error establishing connection with worker %s; retrying in %v", addr, d)
		return nil
	})

	for true {
		var dt *datum
		var ok bool
		select {
		case <-ctx.Done():
			return
		case dt, ok = <-w.dataCh:
			if !ok {
				return
			}
		}
		var resp *workerpkg.ProcessResponse
		var err error
		b := backoff.NewExponentialBackOff()
		if err := backoff.RetryNotify(func() error {
			resp, err = workerClient.Process(ctx, &workerpkg.ProcessRequest{
				JobID: w.jobID,
				Data:  dt.files,
			})
			return err
		}, b, func(err error, d time.Duration) error {
			if d > maxBackoff {
				return err
			}
			protolion.Errorf("worker %s for job %s failed to process datum %v with error %s; retrying in %s", addr, w.jobID, dt.files, err, d)
			return nil
		}); err != nil {
			select {
			case w.failCh <- dt:
			case <-ctx.Done():
			}
			protolion.Errorf("deleting worker %s for job %s", addr, w.jobID)
			// If this worker keeps failing to process the datum (note that
			// failing to process a datum is different than if the user code
			// ran and returned a non-zero exit code), we eventually give up
			// and delete the worker pod.
			if err := w.delWorker(addr); err != nil {
				// If we can't delete the worker for some reason, we will
				// just have to carry on.
				protolion.Errorf("error deleting worker: %v", addr)
				continue
			}
			return
		}
		func() (retErr error) {
			defer func() {
				if retErr != nil {
					protolion.Errorf("datum error in job %s: %v", w.jobID, retErr)
					select {
					case w.failCh <- dt:
					case <-ctx.Done():
					}
				}
			}()
			if resp.Tag != nil {
				var buffer bytes.Buffer
				getTagClient, err := w.objClient.GetTag(ctx, &pfs.Tag{resp.Tag.Name})
				if err != nil {
					return fmt.Errorf("failed to retrieve hashtree after worker %s has ostensibly processed the datum %v: %v", addr, dt.files, err)
				}
				if err := grpcutil.WriteFromStreamingBytesClient(getTagClient, &buffer); err != nil {
					return fmt.Errorf("failed to retrieve hashtree after worker %s has ostensibly processed the datum %v: %v", addr, dt.files, err)
				}
				tree, err := hashtree.Deserialize(buffer.Bytes())
				if err != nil {
					return fmt.Errorf("failed to serialize hashtree after worker %s has ostensibly processed the datum %v; this is likely a bug: %v", addr, dt.files, err)
				}
				w.successCh <- tree
			} else if resp.Failed {
				return fmt.Errorf("user code failed to process datum %v", dt.files)
			} else {
				return fmt.Errorf("unrecognized response from worker %s when processing datum %v; this is likely a bug", addr, dt.files)
			}
			return nil
		}()
	}
}

func (w *workerPool) DataCh() chan<- *datum {
	return w.dataCh
}

func status(ctx context.Context, id string, etcdClient *etcd.Client, etcdPrefix string) ([]*pps.WorkerStatus, error) {
	workerClients, err := workerClients(ctx, id, etcdClient, etcdPrefix)
	if err != nil {
		return nil, err
	}
	var result []*pps.WorkerStatus
	for _, workerClient := range workerClients {
		status, err := workerClient.Status(ctx, &types.Empty{})
		if err != nil {
			return nil, err
		}
		result = append(result, status)
	}
	return result, nil
}

func cancel(ctx context.Context, id string, etcdClient *etcd.Client,
	etcdPrefix string, jobID string, dataFilter []string) error {
	workerClients, err := workerClients(ctx, id, etcdClient, etcdPrefix)
	if err != nil {
		return err
	}
	success := false
	for _, workerClient := range workerClients {
		resp, err := workerClient.Cancel(ctx, &workerpkg.CancelRequest{
			JobID:       jobID,
			DataFilters: dataFilter,
		})
		if err != nil {
			return err
		}
		if resp.Success {
			success = true
		}
	}
	if !success {
		return fmt.Errorf("datum matching filter %+v could not be found for jobID %s", dataFilter, jobID)
	}
	return nil
}

func (w *workerPool) FailCh() <-chan *datum {
	return w.failCh
}

func (w *workerPool) SuccessCh() <-chan hashtree.HashTree {
	return w.successCh
}

// workerPool generates a new worker pool that talks to the replication
// controller identified by rcName.
// Each workerPool is supposed to be owned by a single job, identified
// by jobID.
func (a *apiServer) newWorkerPool(ctx context.Context, rcName string, jobID string) (WorkerPool, error) {
	objClient, err := a.getObjectClient()
	if err != nil {
		return nil, err
	}
	wp := &workerPool{
		ctx:        ctx,
		workerDir:  path.Join(a.etcdPrefix, workerEtcdPrefix, rcName),
		workersMap: make(map[string]worker),
		objClient:  objClient,
		etcdClient: a.etcdClient,
		kubeClient: a.kubeClient,
		namespace:  a.namespace,
		jobID:      jobID,
		dataCh:     make(chan *datum),
		failCh:     make(chan *datum),
		successCh:  make(chan hashtree.HashTree),
	}
	// We need to make sure that the prefix ends with the trailing slash,
	// because
	if wp.workerDir[len(wp.workerDir)-1] != '/' {
		wp.workerDir += "/"
	}

	go wp.discoverWorkers()
	return wp, nil
}

func workerClients(ctx context.Context, id string, etcdClient *etcd.Client, etcdPrefix string) ([]workerpkg.WorkerClient, error) {
	resp, err := etcdClient.Get(ctx, path.Join(etcdPrefix, workerEtcdPrefix, id), etcd.WithPrefix())
	if err != nil {
		return nil, err
	}

	var result []workerpkg.WorkerClient
	for _, kv := range resp.Kvs {
		conn, err := grpc.Dial(fmt.Sprintf("%s:%d", path.Base(string(kv.Key)), client.PPSWorkerPort), grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(5*time.Second))
		if err != nil {
			return nil, err
		}
		result = append(result, workerpkg.NewWorkerClient(conn))
	}
	return result, nil
}
