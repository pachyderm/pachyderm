package collection

import (
	"context"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/client/pps"

	etcd "github.com/coreos/etcd/clientv3"
)

const (
	PipelineIndex Index = "Pipeline"
)

func TestIndex(t *testing.T) {
	etcdClient, err := getEtcdClient()
	require.NoError(t, err)
	uuidPrefix := uuid.NewWithoutDashes()

	persons := NewCollection(etcdClient, uuidPrefix, []Index{PipelineIndex})

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
		persons := persons.ReadWrite(stm)
		persons.Put(j1.Job.ID, j1)
		persons.Put(j2.Job.ID, j2)
		persons.Put(j3.Job.ID, j3)
		return nil
	})
	require.NoError(t, err)

	personsReadonly := persons.ReadOnly(context.Background())

	iter, err := personsReadonly.GetByIndex(PipelineIndex, j1.Pipeline.String())
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

	iter, err = personsReadonly.GetByIndex(PipelineIndex, j3.Pipeline.String())
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

	persons := NewCollection(etcdClient, uuidPrefix, []Index{PipelineIndex})

	j1 := &pps.JobInfo{
		Job:      &pps.Job{"j1"},
		Pipeline: &pps.Pipeline{"p1"},
	}
	_, err = NewSTM(context.Background(), etcdClient, func(stm STM) error {
		persons := persons.ReadWrite(stm)
		persons.Put(j1.Job.ID, j1)
		return nil
	})
	require.NoError(t, err)

	personsReadonly := persons.ReadOnly(context.Background())

	iter, err := personsReadonly.WatchByIndex(PipelineIndex, j1.Pipeline.String())
	require.NoError(t, err)
	var ID string
	job := new(pps.JobInfo)
	event, err := iter.Next()
	require.NoError(t, err)
	require.Equal(t, event.Type(), EventPut)
	require.NoError(t, event.Unmarshal(&ID, job))
	require.Equal(t, j1.Job.ID, ID)
	require.Equal(t, j1, job)
}

func getEtcdClient() (*etcd.Client, error) {
	etcdClient, err := etcd.New(etcd.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return etcdClient, nil
}
