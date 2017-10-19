package worker

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"testing"

	"golang.org/x/sync/errgroup"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/pachyderm/pachyderm/src/client"
	pclient "github.com/pachyderm/pachyderm/src/client"
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

func TestMergeRanges(t *testing.T) {
	require.Equal(t, []*DatumRange(nil), mergeRanges(nil))
	require.Equal(t, []*DatumRange{}, mergeRanges([]*DatumRange{}))
	require.Equal(t, []*DatumRange{&DatumRange{Upper: 10}}, mergeRanges([]*DatumRange{&DatumRange{Upper: 10}}))
	require.Equal(t, []*DatumRange{
		&DatumRange{Upper: 30},
	}, mergeRanges([]*DatumRange{
		&DatumRange{Lower: 20, Upper: 30},
		&DatumRange{Lower: 10, Upper: 20},
		&DatumRange{Upper: 10},
	}))
	require.Equal(t, []*DatumRange{
		&DatumRange{Upper: 10},
		&DatumRange{Lower: 20, Upper: 30},
	}, mergeRanges([]*DatumRange{
		&DatumRange{Lower: 20, Upper: 30},
		&DatumRange{Upper: 10},
	}))
}

func TestAcquireDatums(t *testing.T) {
	c := getPachClient(t)
	etcdClient := getEtcdClient(t)
	dataRepo := uniqueString("TestAcquireDatums")
	require.NoError(t, c.CreateRepo(dataRepo))
	commit, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	for i := 0; i < 100; i++ {
		_, err = c.PutFile(dataRepo, "master", fmt.Sprintf("file-%d", i), strings.NewReader(fmt.Sprintf("%d", i)))
		require.NoError(t, err)
	}
	require.NoError(t, c.FinishCommit(dataRepo, "master"))

	jobInfo := &pps.JobInfo{
		Job: client.NewJob(uuid.New()),
		Input: &pps.Input{
			Atom: &pps.AtomInput{
				Name:   dataRepo,
				Repo:   dataRepo,
				Glob:   "/*",
				Commit: commit.ID,
			}},
	}
	progresses := col.NewCollection(etcdClient, path.Join("", progressPrefix), []col.Index{}, &Progress{}, nil)
	_, err = col.NewSTM(context.Background(), etcdClient, func(stm col.STM) error {
		return progresses.ReadWrite(stm).Create(jobInfo.Job.ID, &Progress{})
	})
	require.NoError(t, err)
	var eg errgroup.Group
	for i := 0; i < 1; i++ {
		server := newTestAPIServer(c, etcdClient, "", t)
		logger := server.getMasterLogger()
		eg.Go(func() error {
			return server.acquireDatums(context.Background(), jobInfo, logger)
		})
	}
	require.NoError(t, eg.Wait())
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
		jobs:       ppsdb.Jobs(etcdClient, etcdPrefix),
		pipelines:  ppsdb.Pipelines(etcdClient, etcdPrefix),
		progresses: col.NewCollection(etcdClient, path.Join(etcdPrefix, progressPrefix), []col.Index{}, &Progress{}, nil),
	}
}

func uniqueString(prefix string) string {
	return prefix + "-" + uuid.NewWithoutDashes()[0:12]
}
