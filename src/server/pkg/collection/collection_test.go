package collection

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
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
	pipelineIndex *Index = &Index{
		Field: "Pipeline",
		Multi: false,
	}
	commitMultiIndex *Index = &Index{
		Field: "Provenance",
		Multi: true,
	}
)

func TestIndex(t *testing.T) {
	etcdClient := getEtcdClient()
	uuidPrefix := uuid.NewWithoutDashes()

	jobInfos := NewCollection(etcdClient, uuidPrefix, []*Index{pipelineIndex}, &pps.JobInfo{}, nil, nil)

	j1 := &pps.JobInfo{
		Job:      client.NewJob("j1"),
		Pipeline: client.NewPipeline("p1"),
	}
	j2 := &pps.JobInfo{
		Job:      client.NewJob("j2"),
		Pipeline: client.NewPipeline("p1"),
	}
	j3 := &pps.JobInfo{
		Job:      client.NewJob("j3"),
		Pipeline: client.NewPipeline("p2"),
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

	job := &pps.JobInfo{}
	i := 1
	require.NoError(t, jobInfosReadonly.GetByIndex(pipelineIndex, j1.Pipeline, job, DefaultOptions, func(ID string) error {
		switch i {
		case 1:
			require.Equal(t, j1.Job.ID, ID)
			require.Equal(t, j1, job)
		case 2:
			require.Equal(t, j2.Job.ID, ID)
			require.Equal(t, j2, job)
		case 3:
			t.Fatal("too many jobs")
		}
		i++
		return nil
	}))

	i = 1
	require.NoError(t, jobInfosReadonly.GetByIndex(pipelineIndex, j3.Pipeline, job, DefaultOptions, func(ID string) error {
		switch i {
		case 1:
			require.Equal(t, j3.Job.ID, ID)
			require.Equal(t, j3, job)
		case 2:
			t.Fatal("too many jobs")
		}
		i++
		return nil
	}))
}

func TestIndexWatch(t *testing.T) {
	etcdClient := getEtcdClient()
	uuidPrefix := uuid.NewWithoutDashes()

	jobInfos := NewCollection(etcdClient, uuidPrefix, []*Index{pipelineIndex}, &pps.JobInfo{}, nil, nil)

	j1 := &pps.JobInfo{
		Job:      client.NewJob("j1"),
		Pipeline: client.NewPipeline("p1"),
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
		Job:      client.NewJob("j2"),
		Pipeline: client.NewPipeline("p1"),
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
		Job:      client.NewJob("j1"),
		Pipeline: client.NewPipeline("p3"),
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

	cis := NewCollection(etcdClient, uuidPrefix, []*Index{commitMultiIndex}, &pfs.CommitInfo{}, nil, nil)

	c1 := &pfs.CommitInfo{
		Commit: client.NewCommit("repo", "c1"),
		Provenance: []*pfs.CommitProvenance{
			client.NewCommitProvenance("in", "master", "c1"),
			client.NewCommitProvenance("in", "master", "c2"),
			client.NewCommitProvenance("in", "master", "c3"),
		},
	}
	c2 := &pfs.CommitInfo{
		Commit: client.NewCommit("repo", "c2"),
		Provenance: []*pfs.CommitProvenance{
			client.NewCommitProvenance("in", "master", "c1"),
			client.NewCommitProvenance("in", "master", "c2"),
			client.NewCommitProvenance("in", "master", "c3"),
		},
	}
	_, err := NewSTM(context.Background(), etcdClient, func(stm STM) error {
		cis := cis.ReadWrite(stm)
		cis.Put(c1.Commit.ID, c1)
		cis.Put(c2.Commit.ID, c2)
		return nil
	})
	require.NoError(t, err)

	cisReadonly := cis.ReadOnly(context.Background())

	// Test that the first key retrieves both r1 and r2
	ci := &pfs.CommitInfo{}
	i := 1
	require.NoError(t, cisReadonly.GetByIndex(commitMultiIndex, client.NewCommit("in", "c1"), ci, DefaultOptions, func(ID string) error {
		if i == 1 {
			require.Equal(t, c1.Commit.ID, ID)
			require.Equal(t, c1, ci)
		} else {
			require.Equal(t, c2.Commit.ID, ID)
			require.Equal(t, c2, ci)
		}
		i++
		return nil
	}))

	// Test that the second key retrieves both r1 and r2
	i = 1
	require.NoError(t, cisReadonly.GetByIndex(commitMultiIndex, client.NewCommit("in", "c2"), ci, DefaultOptions, func(ID string) error {
		if i == 1 {
			require.Equal(t, c1.Commit.ID, ID)
			require.Equal(t, c1, ci)
		} else {
			require.Equal(t, c2.Commit.ID, ID)
			require.Equal(t, c2, ci)
		}
		i++
		return nil
	}))

	// replace "c3" in the provenance of c1 with "c4"
	c1.Provenance[2].Commit.ID = "c4"
	_, err = NewSTM(context.Background(), etcdClient, func(stm STM) error {
		cis := cis.ReadWrite(stm)
		cis.Put(c1.Commit.ID, c1)
		return nil
	})
	require.NoError(t, err)

	// Now "c3" only retrieves c2 (indexes are updated)
	require.NoError(t, cisReadonly.GetByIndex(commitMultiIndex, client.NewCommit("in", "c3"), ci, DefaultOptions, func(ID string) error {
		require.Equal(t, c2.Commit.ID, ID)
		require.Equal(t, c2, ci)
		return nil
	}))

	// And "C4" only retrieves r1 (indexes are updated)
	require.NoError(t, cisReadonly.GetByIndex(commitMultiIndex, client.NewCommit("in", "c4"), ci, DefaultOptions, func(ID string) error {
		require.Equal(t, c1.Commit.ID, ID)
		require.Equal(t, c1, ci)
		return nil
	}))

	// Delete c1 from etcd completely
	_, err = NewSTM(context.Background(), etcdClient, func(stm STM) error {
		cis := cis.ReadWrite(stm)
		cis.Delete(c1.Commit.ID)
		return nil
	})

	// Now "c1" only retrieves c2
	require.NoError(t, cisReadonly.GetByIndex(commitMultiIndex, client.NewCommit("in", "c1"), ci, DefaultOptions, func(ID string) error {
		require.Equal(t, c2.Commit.ID, ID)
		require.Equal(t, c2, ci)
		return nil
	}))
}

func TestBoolIndex(t *testing.T) {
	etcdClient := getEtcdClient()
	uuidPrefix := uuid.NewWithoutDashes()
	boolValues := NewCollection(etcdClient, uuidPrefix, []*Index{{
		Field: "Value",
		Multi: false,
	}}, &types.BoolValue{}, nil, nil)

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

	clxn := NewCollection(etcdClient, uuidPrefix, nil, &types.BoolValue{}, nil, nil)
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

	clxn := NewCollection(etcdClient, uuidPrefix, nil, &types.BoolValue{}, nil, nil)
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
	clxn := NewCollection(etcdClient, uuidPrefix, nil, &types.BoolValue{}, nil, nil)
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

func TestIteration(t *testing.T) {
	etcdClient := getEtcdClient()
	t.Run("one-val-per-txn", func(t *testing.T) {
		uuidPrefix := uuid.NewWithoutDashes()
		col := NewCollection(etcdClient, uuidPrefix, nil, &types.Empty{}, nil, nil)
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
		require.NoError(t, ro.List(val, DefaultOptions, func(key string) error {
			require.Equal(t, fmt.Sprintf("%d", i), key)
			i--
			return nil
		}))
	})
	t.Run("many-vals-per-txn", func(t *testing.T) {
		uuidPrefix := uuid.NewWithoutDashes()
		col := NewCollection(etcdClient, uuidPrefix, nil, &types.Empty{}, nil, nil)
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
		require.NoError(t, ro.List(val, DefaultOptions, func(key string) error {
			require.False(t, vals[key], "saw value %s twice", key)
			vals[key] = true
			return nil
		}))
		require.Equal(t, numBatches*valsPerBatch, len(vals), "didn't receive every value")
	})
	t.Run("large-vals", func(t *testing.T) {
		uuidPrefix := uuid.NewWithoutDashes()
		col := NewCollection(etcdClient, uuidPrefix, nil, &pfs.Repo{}, nil, nil)
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
		ro := col.ReadOnly(context.Background())
		val := &pfs.Repo{}
		vals := make(map[string]bool)
		valsOrder := []string{}
		require.NoError(t, ro.List(val, DefaultOptions, func(key string) error {
			require.False(t, vals[key], "saw value %s twice", key)
			vals[key] = true
			valsOrder = append(valsOrder, key)
			return nil
		}))
		for i, key := range valsOrder {
			require.Equal(t, key, strconv.Itoa(numVals-i-1), "incorrect order returned")
		}
		require.Equal(t, numVals, len(vals), "didn't receive every value")
		vals = make(map[string]bool)
		valsOrder = []string{}
		require.NoError(t, ro.List(val, &Options{etcd.SortByCreateRevision, etcd.SortAscend, true}, func(key string) error {
			require.False(t, vals[key], "saw value %s twice", key)
			vals[key] = true
			valsOrder = append(valsOrder, key)
			return nil
		}))
		for i, key := range valsOrder {
			require.Equal(t, key, strconv.Itoa(i), "incorrect order returned")
		}
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
			DialOptions: client.DefaultDialOptions(),
		})
		if err != nil {
			panic(err)
		}
	})
	return etcdClient
}
