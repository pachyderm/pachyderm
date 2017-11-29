package worker

import (
	// "context"
	"fmt"
	"os"
	"path"
	"sync"
	"testing"

	// "golang.org/x/sync/errgroup"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/pachyderm/pachyderm/src/client"
	pclient "github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsdb"
)

var (
	port int32 = 30653
)

// func TestAcquireDatums(t *testing.T) {
// 	c := getPachClient(t)
// 	etcdClient := getEtcdClient(t)
//
// 	chunks := col.NewCollection(etcdClient, path.Join("", chunksPrefix), []col.Index{}, &Chunks{}, nil)
// 	for nChunks := 1; nChunks < 200; nChunks += 50 {
// 		for nWorkers := 1; nWorkers < 40; nWorkers += 10 {
// 			jobInfo := &pps.JobInfo{
// 				Job: client.NewJob(uuid.New()),
// 			}
// 			_, err := col.NewSTM(context.Background(), etcdClient, func(stm col.STM) error {
// 				c := &Chunks{}
// 				for i := 1; i <= nChunks; i++ {
// 					c.Chunks = append(c.Chunks, int64(i))
// 				}
// 				return chunks.ReadWrite(stm).Create(jobInfo.Job.ID, c)
// 			})
// 			require.NoError(t, err)
// 			var chunks []int64
// 			var chunksMu sync.Mutex
// 			var eg errgroup.Group
// 			for i := 0; i < nWorkers; i++ {
// 				server := newTestAPIServer(c, etcdClient, "", t)
// 				logger := server.getMasterLogger()
// 				eg.Go(func() error {
// 					return server.acquireDatums(context.Background(), jobInfo, logger, func(low, high int64) error {
// 						chunksMu.Lock()
// 						defer chunksMu.Unlock()
// 						chunks = append(chunks, high)
// 						return nil
// 					})
// 				})
// 			}
// 			require.NoError(t, eg.Wait())
// 			require.Equal(t, nChunks, len(chunks))
// 		}
// 	}
// }

// TestJobInput tests the conversion logic in worker/master.go from pipeline
// inputs and commit provenance to job input
func TestJobInput(t *testing.T) {
	pipelineInputs := []*pps.Input{
		client.NewAtomInput("does_exist", "/*"),
		client.NewCrossInput(
			client.NewAtomInput("does_not_exist", "/*"),
			client.NewAtomInput("does_exist", "/*"),
			client.NewAtomInput("does_not_exist", "/*"),
		),
		client.NewCrossInput(
			&pps.Input{
				Cron: &pps.CronInput{Repo: "does_not_exist"},
			}, &pps.Input{
				Cron: &pps.CronInput{Repo: "does_exist"},
			}, &pps.Input{
				Cron: &pps.CronInput{Repo: "does_not_exist"},
			},
		),
		client.NewUnionInput(
			client.NewAtomInput("does_exist", "/*"),
			client.NewAtomInput("does_not_exist", "/*"),
		),
		client.NewCrossInput(
			client.NewAtomInput("does_exist", "/*"),
			client.NewUnionInput(
				client.NewAtomInput("does_not_exist_1", "/*"),
				client.NewAtomInput("does_not_exist_2", "/*"),
			),
		),
		client.NewCrossInput(
			client.NewAtomInput("does_exist", "/*"),
			client.NewUnionInput(
				client.NewCrossInput(
					client.NewAtomInput("does_not_exist_1", "/*"),
					client.NewAtomInput("does_not_exist_2", "/*"),
				),
			),
		),
	}
	expectedJobOutputs := []*pps.Input{
		client.NewAtomInput("does_exist", "/*"),
		client.NewCrossInput(
			client.NewAtomInput("does_exist", "/*"),
		),
		client.NewCrossInput(
			&pps.Input{
				Cron: &pps.CronInput{Repo: "does_exist"},
			},
		),
		client.NewUnionInput(
			client.NewAtomInput("does_exist", "/*"),
		),
		client.NewCrossInput(
			client.NewAtomInput("does_exist", "/*"),
		),
		client.NewCrossInput(
			client.NewAtomInput("does_exist", "/*"),
		),
	}
	cases := []string{
		"atom", "cross", "cross+cron", "union", "nested", "nested union",
	}
	require.Equal(t, len(cases), len(pipelineInputs))
	require.Equal(t, len(cases), len(expectedJobOutputs))
	for i, name := range cases {
		require.True(t, t.Run(name, func(t *testing.T) {
			actual := JobInput(&pps.PipelineInfo{
				Input: pipelineInputs[i],
			}, &pfs.CommitInfo{
				Provenance: []*pfs.Commit{{Repo: &pfs.Repo{Name: "does_exist"}}},
			})
			require.InputEquals(t, expectedJobOutputs[i], actual)
		}))
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
				DialOptions: pclient.EtcdDialOptions(),
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
			pachClient, err = client.NewOnUserMachine(false, "user")
		}
		require.NoError(t, err)
	})
	return pachClient
}

func newTestAPIServer(pachClient *client.APIClient, etcdClient *etcd.Client, etcdPrefix string, t *testing.T) *APIServer {
	return &APIServer{
		pachClient: pachClient,
		etcdClient: etcdClient,
		etcdPrefix: etcdPrefix,
		logMsgTemplate: pps.LogMessage{
			PipelineName: "test",
			WorkerID:     "local",
		},
		jobs:      ppsdb.Jobs(etcdClient, etcdPrefix),
		pipelines: ppsdb.Pipelines(etcdClient, etcdPrefix),
		chunks:    col.NewCollection(etcdClient, path.Join(etcdPrefix, chunksPrefix), []col.Index{}, &Chunks{}, nil),
	}
}

func uniqueString(prefix string) string {
	return prefix + "-" + uuid.NewWithoutDashes()[0:12]
}
