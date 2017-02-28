package collection

import (
	"context"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/watch"

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

	iter, err := personsReadonly.GetByIndex(PipelineIndex, j1.Pipeline)
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

	iter, err = personsReadonly.GetByIndex(PipelineIndex, j3.Pipeline)
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

	eventCh := personsReadonly.WatchByIndex(PipelineIndex, j1.Pipeline.String())
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
		persons := persons.ReadWrite(stm)
		persons.Put(j1.Job.ID, j1)
		return nil
	})

	select {
	case event := <-eventCh:
		t.Fatalf("should not have received an event %s", event)
	case <-time.After(2 * time.Second):
	}

	j2 := &pps.JobInfo{
		Job:      &pps.Job{"j2"},
		Pipeline: &pps.Pipeline{"p1"},
	}

	_, err = NewSTM(context.Background(), etcdClient, func(stm STM) error {
		persons := persons.ReadWrite(stm)
		persons.Put(j2.Job.ID, j2)
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
		persons := persons.ReadWrite(stm)
		persons.Put(j1.Job.ID, j1Prime)
		return nil
	})
	require.NoError(t, err)

	event = <-eventCh
	require.NoError(t, event.Err)
	require.Equal(t, event.Type, watch.EventDelete)
	require.NoError(t, event.Unmarshal(&ID, job))
	require.Equal(t, j1.Job.ID, ID)

	_, err = NewSTM(context.Background(), etcdClient, func(stm STM) error {
		persons := persons.ReadWrite(stm)
		persons.Delete(j2.Job.ID, &pps.JobInfo{})
		return nil
	})
	require.NoError(t, err)

	event = <-eventCh
	require.NoError(t, event.Err)
	require.Equal(t, event.Type, watch.EventDelete)
	require.NoError(t, event.Unmarshal(&ID, job))
	require.Equal(t, j2.Job.ID, ID)
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
