package worker

import (
	"fmt"
	"os"
	"sync"
	"testing"

	"golang.org/x/sync/errgroup"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/server/worker/common"
	"github.com/pachyderm/pachyderm/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/src/server/worker/logs"
)

func TestAcquireDatums(t *testing.T) {
	// TODO: this test is very out of date and doesn't indicate completion to
	// terminate acquireDatums by setting the job context to done
	t.Skip()
	pachClient := getPachClient(t)
	etcdClient := getEtcdClient(t)
	driver := driver.NewMockDriver(etcdClient, &driver.MockOptions{})
	pipeline := &pps.Pipeline{Name: "testPipeline"}

	for nChunks := 1; nChunks < 200; nChunks += 50 {
		for nWorkers := 1; nWorkers < 40; nWorkers += 10 {
			var plan *common.Plan
			// acquireDatums requires the job to exist to update counters, make a barebones placeholder
			job := client.NewJob(uuid.NewWithoutDashes())
			outputCommit := client.NewCommit("testRepo", uuid.NewWithoutDashes())
			jobInfo := &pps.EtcdJobInfo{Job: job, Pipeline: pipeline, OutputCommit: outputCommit}

			_, err := col.NewSTM(pachClient.Ctx(), etcdClient, func(stm col.STM) error {
				if err := driver.Jobs().ReadWrite(stm).Put(jobInfo.Job.ID, jobInfo); err != nil {
					return err
				}

				plan = &common.Plan{}
				for i := 1; i <= nChunks; i++ {
					plan.Chunks = append(plan.Chunks, int64(i))
				}
				return driver.Plans().ReadWrite(stm).Create(jobInfo.Job.ID, plan)
			})
			require.NoError(t, err)

			var seenChunks []int64
			var chunksMu sync.Mutex
			var eg errgroup.Group
			for i := 0; i < nWorkers; i++ {
				server := newTestAPIServer(pachClient, etcdClient, "", driver, t)
				logger := &logs.MockLogger{}
				eg.Go(func() error {
					return server.acquireDatums(pachClient.Ctx(), jobInfo.Job.ID, plan, logger, func(low, high int64) (*processResult, error) {
						chunksMu.Lock()
						defer chunksMu.Unlock()
						seenChunks = append(seenChunks, high)
						return &processResult{}, nil
					})
				})
			}
			require.NoError(t, eg.Wait())
			require.Equal(t, nChunks, len(seenChunks))
		}
	}
}

var etcdClient *etcd.Client
var etcdOnce sync.Once

func getEtcdClient(t *testing.T) *etcd.Client {
	// src/server/pfs/server/driver.go expects an etcd server at "localhost:32379"
	// Try to establish a connection before proceeding with the test (which will
	// fail if the connection can't be established)
	etcdAddress := "localhost:32379"
	etcdOnce.Do(func() {
		require.NoError(t, backoff.Retry(func() error {
			var err error
			etcdClient, err = etcd.New(etcd.Config{
				Endpoints:   []string{etcdAddress},
				DialOptions: client.DefaultDialOptions(),
			})
			if err != nil {
				return fmt.Errorf("could not connect to etcd: %s", err.Error())
			}
			return nil
		}, backoff.NewTestingBackOff()))
	})
	return etcdClient
}

var pachClient *client.APIClient
var getPachClientOnce sync.Once

func getPachClient(t testing.TB) *client.APIClient {
	getPachClientOnce.Do(func() {
		var err error
		if addr := os.Getenv("PACHD_PORT_650_TCP_ADDR"); addr != "" {
			pachClient, err = client.NewInCluster()
		} else {
			pachClient, err = client.NewForTest()
		}
		require.NoError(t, err)
	})
	return pachClient
}

func newTestAPIServer(pachClient *client.APIClient, etcdClient *etcd.Client, etcdPrefix string, d driver.Driver, t *testing.T) *APIServer {
	return &APIServer{
		pachClient: pachClient,
		etcdClient: etcdClient,
		etcdPrefix: etcdPrefix,
		driver:     d,
	}
}
