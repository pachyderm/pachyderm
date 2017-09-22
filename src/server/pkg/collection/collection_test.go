package collection

import (
	"context"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/watch"

	etcd "github.com/coreos/etcd/clientv3"
)

var (
	pipelineIndex Index = Index{
		Field: "Pipeline",
		Multi: false,
	}
	repoMultiIndex Index = Index{
		Field: "Provenance",
		Multi: true,
	}
)

func TestIndex(t *testing.T) {
	etcdClient, err := getEtcdClient()
	require.NoError(t, err)
	uuidPrefix := uuid.NewWithoutDashes()

	jobInfos := NewCollection(etcdClient, uuidPrefix, []Index{pipelineIndex}, &pps.JobInfo{}, nil)

	j1 := &pps.JobInfo{
		Job:      &pps.Job{"j1"},
		Pipeline: &pps.Pipeline{"p1"},
	}
	j2 := &pps.JobInfo{
		Job:      &pps.Job{"j2"},
		Pipeline: &pps.Pipeline{"p1"},
	}
	j3 := &pps.JobInfo{
		Job:      &pps.Job{"j3"},
		Pipeline: &pps.Pipeline{"p2"},
	}
	_, err = NewSTM(context.Background(), etcdClient, func(stm STM) error {
		jobInfos := jobInfos.ReadWrite(stm)
		jobInfos.Put(j1.Job.ID, j1)
		jobInfos.Put(j2.Job.ID, j2)
		jobInfos.Put(j3.Job.ID, j3)
		return nil
	})
	require.NoError(t, err)

	jobInfosReadonly := jobInfos.ReadOnly(context.Background())

	iter, err := jobInfosReadonly.GetByIndex(pipelineIndex, j1.Pipeline)
	require.NoError(t, err)
	var ID string
	job := new(pps.JobInfo)
	ok, err := iter.Next(&ID, job)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, j1.Job.ID, ID)
	require.Equal(t, j1, job)
	ok, err = iter.Next(&ID, job)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, j2.Job.ID, ID)
	require.Equal(t, j2, job)
	ok, err = iter.Next(&ID, job)
	require.NoError(t, err)
	require.False(t, ok)

	iter, err = jobInfosReadonly.GetByIndex(pipelineIndex, j3.Pipeline)
	require.NoError(t, err)
	ok, err = iter.Next(&ID, job)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, j3.Job.ID, ID)
	require.Equal(t, j3, job)
	ok, err = iter.Next(&ID, job)
	require.NoError(t, err)
	require.False(t, ok)
}

func TestIndexWatch(t *testing.T) {
	etcdClient, err := getEtcdClient()
	require.NoError(t, err)
	uuidPrefix := uuid.NewWithoutDashes()

	jobInfos := NewCollection(etcdClient, uuidPrefix, []Index{pipelineIndex}, &pps.JobInfo{}, nil)

	j1 := &pps.JobInfo{
		Job:      &pps.Job{"j1"},
		Pipeline: &pps.Pipeline{"p1"},
	}
	_, err = NewSTM(context.Background(), etcdClient, func(stm STM) error {
		jobInfos := jobInfos.ReadWrite(stm)
		jobInfos.Put(j1.Job.ID, j1)
		return nil
	})
	require.NoError(t, err)

	jobInfosReadonly := jobInfos.ReadOnly(context.Background())

	watcher, err := jobInfosReadonly.WatchByIndex(pipelineIndex, j1.Pipeline.String())
	eventCh := watcher.Watch()
	require.NoError(t, err)
	var ID string
	job := new(pps.JobInfo)
	event := <-eventCh
	require.NoError(t, event.Err)
	require.Equal(t, event.Type, watch.EventPut)
	require.NoError(t, event.Unmarshal(&ID, job))
	require.Equal(t, j1.Job.ID, ID)
	require.Equal(t, j1, job)

	// Now we will put j1 again, unchanged.  We want to make sure
	// that we do not receive an event.
	_, err = NewSTM(context.Background(), etcdClient, func(stm STM) error {
		jobInfos := jobInfos.ReadWrite(stm)
		jobInfos.Put(j1.Job.ID, j1)
		return nil
	})

	select {
	case event := <-eventCh:
		t.Fatalf("should not have received an event %v", event)
	case <-time.After(2 * time.Second):
	}

	j2 := &pps.JobInfo{
		Job:      &pps.Job{"j2"},
		Pipeline: &pps.Pipeline{"p1"},
	}

	_, err = NewSTM(context.Background(), etcdClient, func(stm STM) error {
		jobInfos := jobInfos.ReadWrite(stm)
		jobInfos.Put(j2.Job.ID, j2)
		return nil
	})
	require.NoError(t, err)

	event = <-eventCh
	require.NoError(t, event.Err)
	require.Equal(t, event.Type, watch.EventPut)
	require.NoError(t, event.Unmarshal(&ID, job))
	require.Equal(t, j2.Job.ID, ID)
	require.Equal(t, j2, job)

	j1Prime := &pps.JobInfo{
		Job:      &pps.Job{"j1"},
		Pipeline: &pps.Pipeline{"p3"},
	}
	_, err = NewSTM(context.Background(), etcdClient, func(stm STM) error {
		jobInfos := jobInfos.ReadWrite(stm)
		jobInfos.Put(j1.Job.ID, j1Prime)
		return nil
	})
	require.NoError(t, err)

	event = <-eventCh
	require.NoError(t, event.Err)
	require.Equal(t, event.Type, watch.EventDelete)
	require.NoError(t, event.Unmarshal(&ID, job))
	require.Equal(t, j1.Job.ID, ID)

	_, err = NewSTM(context.Background(), etcdClient, func(stm STM) error {
		jobInfos := jobInfos.ReadWrite(stm)
		jobInfos.Delete(j2.Job.ID)
		return nil
	})
	require.NoError(t, err)

	event = <-eventCh
	require.NoError(t, event.Err)
	require.Equal(t, event.Type, watch.EventDelete)
	require.NoError(t, event.Unmarshal(&ID, job))
	require.Equal(t, j2.Job.ID, ID)
}

func TestMultiIndex(t *testing.T) {
	etcdClient, err := getEtcdClient()
	require.NoError(t, err)
	uuidPrefix := uuid.NewWithoutDashes()

	repoInfos := NewCollection(etcdClient, uuidPrefix, []Index{repoMultiIndex}, &pfs.RepoInfo{}, nil)

	r1 := &pfs.RepoInfo{
		Repo: &pfs.Repo{"r1"},
		Provenance: []*pfs.Repo{
			{"input1"},
			{"input2"},
			{"input3"},
		},
	}
	r2 := &pfs.RepoInfo{
		Repo: &pfs.Repo{"r2"},
		Provenance: []*pfs.Repo{
			{"input1"},
			{"input2"},
			{"input3"},
		},
	}
	_, err = NewSTM(context.Background(), etcdClient, func(stm STM) error {
		repoInfos := repoInfos.ReadWrite(stm)
		repoInfos.Put(r1.Repo.Name, r1)
		repoInfos.Put(r2.Repo.Name, r2)
		return nil
	})
	require.NoError(t, err)

	repoInfosReadonly := repoInfos.ReadOnly(context.Background())

	// Test that the first key retrieves both r1 and r2
	iter, err := repoInfosReadonly.GetByIndex(repoMultiIndex, &pfs.Repo{"input1"})
	require.NoError(t, err)
	var Name string
	repo := new(pfs.RepoInfo)
	ok, err := iter.Next(&Name, repo)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, r1.Repo.Name, Name)
	require.Equal(t, r1, repo)
	repo = new(pfs.RepoInfo)
	ok, err = iter.Next(&Name, repo)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, r2.Repo.Name, Name)
	require.Equal(t, r2, repo)

	// Test that the *second* key retrieves both r1 and r2
	iter, err = repoInfosReadonly.GetByIndex(repoMultiIndex, &pfs.Repo{"input2"})
	require.NoError(t, err)
	repo = new(pfs.RepoInfo)
	ok, err = iter.Next(&Name, repo)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, r1.Repo.Name, Name)
	require.Equal(t, r1, repo)
	repo = new(pfs.RepoInfo)
	ok, err = iter.Next(&Name, repo)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, r2.Repo.Name, Name)
	require.Equal(t, r2, repo)

	// replace "input3" in the provenance of r1 with "input4"
	r1.Provenance[2] = &pfs.Repo{"input4"}
	_, err = NewSTM(context.Background(), etcdClient, func(stm STM) error {
		repoInfos := repoInfos.ReadWrite(stm)
		repoInfos.Put(r1.Repo.Name, r1)
		return nil
	})
	require.NoError(t, err)

	// Now "input3" only retrieves r2 (indexes are updated)
	iter, err = repoInfosReadonly.GetByIndex(repoMultiIndex, &pfs.Repo{"input3"})
	require.NoError(t, err)
	repo = new(pfs.RepoInfo)
	ok, err = iter.Next(&Name, repo)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, r2.Repo.Name, Name)
	require.Equal(t, r2, repo)

	// As well, "input4" only retrieves r1 (indexes are updated)
	iter, err = repoInfosReadonly.GetByIndex(repoMultiIndex, &pfs.Repo{"input4"})
	require.NoError(t, err)
	repo = new(pfs.RepoInfo)
	ok, err = iter.Next(&Name, repo)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, r1.Repo.Name, Name)
	require.Equal(t, r1, repo)

	// Delete r1 from etcd completely
	_, err = NewSTM(context.Background(), etcdClient, func(stm STM) error {
		repoInfos := repoInfos.ReadWrite(stm)
		repoInfos.Delete(r1.Repo.Name)
		return nil
	})

	// Now "input1" only retrieves r2
	iter, err = repoInfosReadonly.GetByIndex(repoMultiIndex, &pfs.Repo{"input1"})
	require.NoError(t, err)
	repo = new(pfs.RepoInfo)
	ok, err = iter.Next(&Name, repo)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, r2.Repo.Name, Name)
	require.Equal(t, r2, repo)
}

func getEtcdClient() (*etcd.Client, error) {
	etcdClient, err := etcd.New(etcd.Config{
		Endpoints:   []string{"localhost:32379"},
		DialOptions: client.EtcdDialOptions(),
	})
	if err != nil {
		return nil, err
	}
	return etcdClient, nil
}
