package collection

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/server/pkg/watch"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/types"
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
	etcdClient := getEtcdClient()
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
	_, err := NewSTM(context.Background(), etcdClient, func(stm STM) error {
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
	etcdClient := getEtcdClient()
	uuidPrefix := uuid.NewWithoutDashes()

	jobInfos := NewCollection(etcdClient, uuidPrefix, []Index{pipelineIndex}, &pps.JobInfo{}, nil)

	j1 := &pps.JobInfo{
		Job:      &pps.Job{"j1"},
		Pipeline: &pps.Pipeline{"p1"},
	}
	_, err := NewSTM(context.Background(), etcdClient, func(stm STM) error {
		jobInfos := jobInfos.ReadWrite(stm)
		jobInfos.Put(j1.Job.ID, j1)
		return nil
	})
	require.NoError(t, err)

	jobInfosReadonly := jobInfos.ReadOnly(context.Background())

	watcher, err := jobInfosReadonly.WatchByIndex(pipelineIndex, j1.Pipeline)
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
	etcdClient := getEtcdClient()
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
	_, err := NewSTM(context.Background(), etcdClient, func(stm STM) error {
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

func TestBoolIndex(t *testing.T) {
	etcdClient := getEtcdClient()
	uuidPrefix := uuid.NewWithoutDashes()
	boolValues := NewCollection(etcdClient, uuidPrefix, []Index{{
		Field: "Value",
		Multi: false,
	}}, &types.BoolValue{}, nil)

	r1 := &types.BoolValue{
		Value: true,
	}
	r2 := &types.BoolValue{
		Value: false,
	}
	_, err := NewSTM(context.Background(), etcdClient, func(stm STM) error {
		boolValues := boolValues.ReadWrite(stm)
		boolValues.Put("true", r1)
		boolValues.Put("false", r2)
		return nil
	})
	require.NoError(t, err)

	// Test that we don't format the index string incorrectly
	resp, err := etcdClient.Get(context.Background(), uuidPrefix, etcd.WithPrefix())
	require.NoError(t, err)
	for _, kv := range resp.Kvs {
		if !bytes.Contains(kv.Key, []byte("__index_")) {
			continue // not an index record
		}
		require.True(t,
			bytes.Contains(kv.Key, []byte("__index_Value/true")) ||
				bytes.Contains(kv.Key, []byte("__index_Value/false")), string(kv.Key))
	}
}

var epsilon = &types.BoolValue{Value: true}

func TestTTL(t *testing.T) {
	etcdClient := getEtcdClient()
	uuidPrefix := uuid.NewWithoutDashes()

	clxn := NewCollection(etcdClient, uuidPrefix, nil, &types.BoolValue{}, nil)
	const TTL = 5
	_, err := NewSTM(context.Background(), etcdClient, func(stm STM) error {
		return clxn.ReadWrite(stm).PutTTL("key", epsilon, TTL)
	})
	require.NoError(t, err)

	var actualTTL int64
	_, err = NewSTM(context.Background(), etcdClient, func(stm STM) error {
		var err error
		actualTTL, err = clxn.ReadWrite(stm).TTL("key")
		return err
	})
	require.NoError(t, err)
	require.True(t, actualTTL > 0 && actualTTL < TTL, "actualTTL was %v", actualTTL)
}

func TestTTLExpire(t *testing.T) {
	etcdClient := getEtcdClient()
	uuidPrefix := uuid.NewWithoutDashes()

	clxn := NewCollection(etcdClient, uuidPrefix, nil, &types.BoolValue{}, nil)
	const TTL = 5
	_, err := NewSTM(context.Background(), etcdClient, func(stm STM) error {
		return clxn.ReadWrite(stm).PutTTL("key", epsilon, TTL)
	})
	require.NoError(t, err)

	time.Sleep((TTL + 1) * time.Second)
	value := &types.BoolValue{}
	err = clxn.ReadOnly(context.Background()).Get("key", value)
	require.NotNil(t, err)
	require.Matches(t, "not found", err.Error())
}

func TestTTLExtend(t *testing.T) {
	etcdClient := getEtcdClient()
	uuidPrefix := uuid.NewWithoutDashes()

	// Put value with short TLL & check that it was set
	clxn := NewCollection(etcdClient, uuidPrefix, nil, &types.BoolValue{}, nil)
	const TTL = 5
	_, err := NewSTM(context.Background(), etcdClient, func(stm STM) error {
		return clxn.ReadWrite(stm).PutTTL("key", epsilon, TTL)
	})
	require.NoError(t, err)

	var actualTTL int64
	_, err = NewSTM(context.Background(), etcdClient, func(stm STM) error {
		var err error
		actualTTL, err = clxn.ReadWrite(stm).TTL("key")
		return err
	})
	require.NoError(t, err)
	require.True(t, actualTTL > 0 && actualTTL < TTL, "actualTTL was %v", actualTTL)

	// Put value with new, longer TLL and check that it was set
	const LongerTTL = 15
	_, err = NewSTM(context.Background(), etcdClient, func(stm STM) error {
		return clxn.ReadWrite(stm).PutTTL("key", epsilon, LongerTTL)
	})
	require.NoError(t, err)

	_, err = NewSTM(context.Background(), etcdClient, func(stm STM) error {
		var err error
		actualTTL, err = clxn.ReadWrite(stm).TTL("key")
		return err
	})
	require.NoError(t, err)
	require.True(t, actualTTL > TTL && actualTTL < LongerTTL, "actualTTL was %v", actualTTL)
}

func TestPagination(t *testing.T) {
	etcdClient := getEtcdClient()
	t.Run("one-val-per-txn", func(t *testing.T) {
		uuidPrefix := uuid.NewWithoutDashes()
		col := NewCollection(etcdClient, uuidPrefix, nil, &types.Empty{}, nil)
		numVals := 1000
		for i := 0; i < numVals; i++ {
			_, err := NewSTM(context.Background(), etcdClient, func(stm STM) error {
				return col.ReadWrite(stm).Put(fmt.Sprintf("%d", i), &types.Empty{})
			})
			require.NoError(t, err)
		}
		ro := col.ReadOnly(context.Background())
		val := &types.Empty{}
		i := numVals - 1
		require.NoError(t, ro.ListF(val, func(key string) error {
			require.Equal(t, fmt.Sprintf("%d", i), key)
			i--
			return nil
		}))
	})
	t.Run("many-vals-per-txn", func(t *testing.T) {
		uuidPrefix := uuid.NewWithoutDashes()
		col := NewCollection(etcdClient, uuidPrefix, nil, &types.Empty{}, nil)
		numBatches := 10
		valsPerBatch := 7
		for i := 0; i < numBatches; i++ {
			_, err := NewSTM(context.Background(), etcdClient, func(stm STM) error {
				for j := 0; j < valsPerBatch; j++ {
					if err := col.ReadWrite(stm).Put(fmt.Sprintf("%d", i*valsPerBatch+j), &types.Empty{}); err != nil {
						return err
					}
				}
				return nil
			})
			require.NoError(t, err)
		}
		vals := make(map[string]bool)
		ro := col.ReadOnly(context.Background())
		val := &types.Empty{}
		require.NoError(t, ro.ListF(val, func(key string) error {
			require.False(t, vals[key], "saw value %s twice", key)
			vals[key] = true
			return nil
		}))
		require.Equal(t, numBatches*valsPerBatch, len(vals), "didn't receive every value")
	})
	t.Run("large-vals", func(t *testing.T) {
		uuidPrefix := uuid.NewWithoutDashes()
		col := NewCollection(etcdClient, uuidPrefix, nil, &pfs.Repo{}, nil)
		numVals := 100
		longString := strings.Repeat("foo\n", 1024*256) // 1 MB worth of foo
		for i := 0; i < numVals; i++ {
			_, err := NewSTM(context.Background(), etcdClient, func(stm STM) error {
				if err := col.ReadWrite(stm).Put(fmt.Sprintf("%d", i), &pfs.Repo{Name: longString}); err != nil {
					return err
				}
				return nil
			})
			require.NoError(t, err)
		}
		vals := make(map[string]bool)
		ro := col.ReadOnly(context.Background())
		val := &pfs.Repo{}
		require.NoError(t, ro.ListF(val, func(key string) error {
			require.False(t, vals[key], "saw value %s twice", key)
			vals[key] = true
			return nil
		}))
		require.Equal(t, numVals, len(vals), "didn't receive every value")
	})
}

var etcdClient *etcd.Client
var etcdClientOnce sync.Once

func getEtcdClient() *etcd.Client {
	etcdClientOnce.Do(func() {
		var err error
		etcdClient, err = etcd.New(etcd.Config{
			Endpoints:   []string{"localhost:32379"},
			DialOptions: client.EtcdDialOptions(),
		})
		if err != nil {
			panic(err)
		}
	})
	return etcdClient
}
