package collection

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/watch"

	etcd "github.com/coreos/etcd/clientv3"
)

var (
	pipelineIndex    Index = Index{"Pipeline", false}
	inputsMultiIndex Index = Index{"Inputs", true}
)

func TestIndex(t *testing.T) {
	etcdClient, err := getEtcdClient()
	require.NoError(t, err)
	uuidPrefix := uuid.NewWithoutDashes()

	jobInfos := NewCollection(etcdClient, uuidPrefix, []Index{pipelineIndex}, &pps.JobInfo{})

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

	jobInfos := NewCollection(etcdClient, uuidPrefix, []Index{pipelineIndex}, &pps.JobInfo{})

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

	fmt.Println("BP1")

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

	fmt.Println("BP7")
}

func TestMultiIndex(t *testing.T) {
	etcdClient, err := getEtcdClient()
	require.NoError(t, err)
	uuidPrefix := uuid.NewWithoutDashes()

	jobInfos := NewCollection(etcdClient, uuidPrefix, []Index{inputsMultiIndex}, &pps.JobInfo{})

	j1 := &pps.JobInfo{
		Job: &pps.Job{"j1"},
		Inputs: []*pps.JobInput{
			{Name: "input1"},
			{Name: "input2"},
			{Name: "input3"},
		},
	}
	j2 := &pps.JobInfo{
		Job: &pps.Job{"j2"},
		Inputs: []*pps.JobInput{
			{Name: "input1"},
			{Name: "input2"},
			{Name: "input3"},
		},
	}
	_, err = NewSTM(context.Background(), etcdClient, func(stm STM) error {
		jobInfos := jobInfos.ReadWrite(stm)
		jobInfos.Put(j1.Job.ID, j1)
		jobInfos.Put(j2.Job.ID, j2)
		return nil
	})
	require.NoError(t, err)

	jobInfosReadonly := jobInfos.ReadOnly(context.Background())

	iter, err := jobInfosReadonly.GetByIndex(inputsMultiIndex, &pps.JobInput{
		Name: "input1",
	})
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

	iter, err = jobInfosReadonly.GetByIndex(inputsMultiIndex, &pps.JobInput{
		Name: "input2",
	})
	require.NoError(t, err)
	ok, err = iter.Next(&ID, job)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, j1.Job.ID, ID)
	require.Equal(t, j1, job)
	ok, err = iter.Next(&ID, job)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, j2.Job.ID, ID)
	require.Equal(t, j2, job)

	j1.Inputs[2] = &pps.JobInput{Name: "input4"}
	_, err = NewSTM(context.Background(), etcdClient, func(stm STM) error {
		jobInfos := jobInfos.ReadWrite(stm)
		jobInfos.Put(j1.Job.ID, j1)
		return nil
	})
	require.NoError(t, err)

	iter, err = jobInfosReadonly.GetByIndex(inputsMultiIndex, &pps.JobInput{
		Name: "input3",
	})
	require.NoError(t, err)
	ok, err = iter.Next(&ID, job)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, j2.Job.ID, ID)
	require.Equal(t, j2, job)

	iter, err = jobInfosReadonly.GetByIndex(inputsMultiIndex, &pps.JobInput{
		Name: "input4",
	})
	require.NoError(t, err)
	ok, err = iter.Next(&ID, job)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, j1.Job.ID, ID)
	require.Equal(t, j1, job)

	_, err = NewSTM(context.Background(), etcdClient, func(stm STM) error {
		jobInfos := jobInfos.ReadWrite(stm)
		jobInfos.Delete(j1.Job.ID)
		return nil
	})

	iter, err = jobInfosReadonly.GetByIndex(inputsMultiIndex, &pps.JobInput{
		Name: "input1",
	})
	require.NoError(t, err)
	ok, err = iter.Next(&ID, job)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, j2.Job.ID, ID)
	require.Equal(t, j2, job)
}

func getEtcdClient() (*etcd.Client, error) {
	etcdClient, err := etcd.New(etcd.Config{
		Endpoints:   []string{"localhost:2379"},
		DialOptions: client.EtcdDialOptions(),
	})
	if err != nil {
		return nil, err
	}
	return etcdClient, nil
}
