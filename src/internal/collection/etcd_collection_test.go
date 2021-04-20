package collection_test

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"

	"github.com/pachyderm/pachyderm/v2/src/client"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

var (
	pipelineIndex *col.Index = &col.Index{
		Name: "Pipeline",
		Extract: func(val proto.Message) string {
			return val.(*pps.JobInfo).Pipeline.Name
		},
	}
)

func TestEtcdCollections(suite *testing.T) {
	suite.Parallel()
	etcdEnv := testetcd.NewEnv(suite)
	newCollection := func(ctx context.Context, t *testing.T) (ReadCallback, WriteCallback) {
		prefix := testutil.UniqueString("test-etcd-collections-")
		testCol := col.NewEtcdCollection(etcdEnv.EtcdClient, prefix, []*col.Index{TestSecondaryIndex}, &col.TestItem{}, nil, nil)

		readCallback := func(ctx context.Context) col.ReadOnlyCollection {
			return testCol.ReadOnly(ctx)
		}

		writeCallback := func(ctx context.Context, f func(col.ReadWriteCollection) error) error {
			_, err := col.NewSTM(ctx, etcdEnv.EtcdClient, func(stm col.STM) (retErr error) {
				return f(testCol.ReadWrite(stm))
			})
			return errors.EnsureStack(err)
		}

		return readCallback, writeCallback
	}

	collectionTests(suite, newCollection)
	watchTests(suite, newCollection)
}

func TestDryrun(t *testing.T) {
	t.Parallel()
	env := testetcd.NewEnv(t)
	uuidPrefix := uuid.NewWithoutDashes()

	jobInfos := col.NewEtcdCollection(env.EtcdClient, uuidPrefix, nil, &pps.JobInfo{}, nil, nil)

	job := &pps.JobInfo{
		Job:      client.NewJob("j1"),
		Pipeline: client.NewPipeline("p1"),
	}
	err := col.NewDryrunSTM(context.Background(), env.EtcdClient, func(stm col.STM) error {
		jobInfos := jobInfos.ReadWrite(stm)
		jobInfos.Put(job.Job.ID, job)
		return nil
	})
	require.NoError(t, err)

	jobInfosReadonly := jobInfos.ReadOnly(context.Background())
	err = jobInfosReadonly.Get("j1", job)
	require.True(t, col.IsErrNotFound(err))
}

func TestDeletePrefix(t *testing.T) {
	t.Parallel()
	env := testetcd.NewEnv(t)
	uuidPrefix := uuid.NewWithoutDashes()

	jobInfos := col.NewEtcdCollection(env.EtcdClient, uuidPrefix, nil, &pps.JobInfo{}, nil, nil)

	j1 := &pps.JobInfo{
		Job:      client.NewJob("prefix/suffix/job"),
		Pipeline: client.NewPipeline("p"),
	}
	j2 := &pps.JobInfo{
		Job:      client.NewJob("prefix/suffix/job2"),
		Pipeline: client.NewPipeline("p"),
	}
	j3 := &pps.JobInfo{
		Job:      client.NewJob("prefix/job3"),
		Pipeline: client.NewPipeline("p"),
	}
	j4 := &pps.JobInfo{
		Job:      client.NewJob("job4"),
		Pipeline: client.NewPipeline("p"),
	}

	_, err := col.NewSTM(context.Background(), env.EtcdClient, func(stm col.STM) error {
		jobInfos := jobInfos.ReadWrite(stm)
		jobInfos.Put(j1.Job.ID, j1)
		jobInfos.Put(j2.Job.ID, j2)
		jobInfos.Put(j3.Job.ID, j3)
		jobInfos.Put(j4.Job.ID, j4)
		return nil
	})
	require.NoError(t, err)

	_, err = col.NewSTM(context.Background(), env.EtcdClient, func(stm col.STM) error {
		job := &pps.JobInfo{}
		jobInfos := jobInfos.ReadWrite(stm)

		jobInfos.DeleteAllPrefix("prefix/suffix")
		if err := jobInfos.Get(j1.Job.ID, job); !col.IsErrNotFound(err) {
			return errors.Wrapf(err, "Expected ErrNotFound for key '%s', but got", j1.Job.ID)
		}
		if err := jobInfos.Get(j2.Job.ID, job); !col.IsErrNotFound(err) {
			return errors.Wrapf(err, "Expected ErrNotFound for key '%s', but got", j2.Job.ID)
		}
		if err := jobInfos.Get(j3.Job.ID, job); err != nil {
			return err
		}
		if err := jobInfos.Get(j4.Job.ID, job); err != nil {
			return err
		}

		jobInfos.DeleteAllPrefix("prefix")
		if err := jobInfos.Get(j1.Job.ID, job); !col.IsErrNotFound(err) {
			return errors.Wrapf(err, "Expected ErrNotFound for key '%s', but got", j1.Job.ID)
		}
		if err := jobInfos.Get(j2.Job.ID, job); !col.IsErrNotFound(err) {
			return errors.Wrapf(err, "Expected ErrNotFound for key '%s', but got", j2.Job.ID)
		}
		if err := jobInfos.Get(j3.Job.ID, job); !col.IsErrNotFound(err) {
			return errors.Wrapf(err, "Expected ErrNotFound for key '%s', but got", j3.Job.ID)
		}
		if err := jobInfos.Get(j4.Job.ID, job); err != nil {
			return err
		}

		jobInfos.Put(j1.Job.ID, j1)
		if err := jobInfos.Get(j1.Job.ID, job); err != nil {
			return err
		}

		jobInfos.DeleteAllPrefix("prefix/suffix")
		if err := jobInfos.Get(j1.Job.ID, job); !col.IsErrNotFound(err) {
			return errors.Wrapf(err, "Expected ErrNotFound for key '%s', but got", j1.Job.ID)
		}

		jobInfos.Put(j2.Job.ID, j2)
		if err := jobInfos.Get(j2.Job.ID, job); err != nil {
			return err
		}

		return nil
	})
	require.NoError(t, err)

	job := &pps.JobInfo{}
	jobs := jobInfos.ReadOnly(context.Background())
	require.True(t, col.IsErrNotFound(jobs.Get(j1.Job.ID, job)))
	require.NoError(t, jobs.Get(j2.Job.ID, job))
	require.Equal(t, j2, job)
	require.True(t, col.IsErrNotFound(jobs.Get(j3.Job.ID, job)))
	require.NoError(t, jobs.Get(j4.Job.ID, job))
	require.Equal(t, j4, job)
}

func TestIndex(t *testing.T) {
	t.Parallel()
	env := testetcd.NewEnv(t)
	uuidPrefix := uuid.NewWithoutDashes()

	jobInfos := col.NewEtcdCollection(env.EtcdClient, uuidPrefix, []*col.Index{pipelineIndex}, &pps.JobInfo{}, nil, nil)

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
	_, err := col.NewSTM(context.Background(), env.EtcdClient, func(stm col.STM) error {
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
	require.NoError(t, jobInfosReadonly.GetByIndex(pipelineIndex, j1.Pipeline.Name, job, col.DefaultOptions(), func(string) error {
		switch i {
		case 1:
			require.Equal(t, j1, job)
		case 2:
			require.Equal(t, j2, job)
		case 3:
			t.Fatal("too many jobs")
		}
		i++
		return nil
	}))

	i = 1
	require.NoError(t, jobInfosReadonly.GetByIndex(pipelineIndex, j3.Pipeline.Name, job, col.DefaultOptions(), func(string) error {
		switch i {
		case 1:
			require.Equal(t, j3, job)
		case 2:
			t.Fatal("too many jobs")
		}
		i++
		return nil
	}))
}

func TestBoolIndex(t *testing.T) {
	t.Parallel()
	env := testetcd.NewEnv(t)
	uuidPrefix := uuid.NewWithoutDashes()
	index := &col.Index{
		Name: "Value",
		Extract: func(val proto.Message) string {
			return fmt.Sprintf("%v", val.(*types.BoolValue).Value)
		},
	}
	boolValues := col.NewEtcdCollection(env.EtcdClient, uuidPrefix, []*col.Index{index}, &types.BoolValue{}, nil, nil)

	r1 := &types.BoolValue{
		Value: true,
	}
	r2 := &types.BoolValue{
		Value: false,
	}
	_, err := col.NewSTM(context.Background(), env.EtcdClient, func(stm col.STM) error {
		boolValues := boolValues.ReadWrite(stm)
		boolValues.Put("true", r1)
		boolValues.Put("false", r2)
		return nil
	})
	require.NoError(t, err)

	// Test that we don't format the index string incorrectly
	resp, err := env.EtcdClient.Get(context.Background(), uuidPrefix, etcd.WithPrefix())
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
	t.Parallel()
	env := testetcd.NewEnv(t)
	uuidPrefix := uuid.NewWithoutDashes()

	clxn := col.NewEtcdCollection(env.EtcdClient, uuidPrefix, nil, &types.BoolValue{}, nil, nil)
	const TTL = 5
	_, err := col.NewSTM(context.Background(), env.EtcdClient, func(stm col.STM) error {
		return clxn.ReadWrite(stm).PutTTL("key", epsilon, TTL)
	})
	require.NoError(t, err)

	var actualTTL int64
	_, err = col.NewSTM(context.Background(), env.EtcdClient, func(stm col.STM) error {
		var err error
		actualTTL, err = clxn.ReadWrite(stm).TTL("key")
		return err
	})
	require.NoError(t, err)
	require.True(t, actualTTL > 0 && actualTTL < TTL, "actualTTL was %v", actualTTL)
}

func TestTTLExpire(t *testing.T) {
	t.Parallel()
	env := testetcd.NewEnv(t)
	uuidPrefix := uuid.NewWithoutDashes()

	clxn := col.NewEtcdCollection(env.EtcdClient, uuidPrefix, nil, &types.BoolValue{}, nil, nil)
	const TTL = 5
	_, err := col.NewSTM(context.Background(), env.EtcdClient, func(stm col.STM) error {
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
	t.Parallel()
	env := testetcd.NewEnv(t)
	uuidPrefix := uuid.NewWithoutDashes()

	// Put value with short TLL & check that it was set
	clxn := col.NewEtcdCollection(env.EtcdClient, uuidPrefix, nil, &types.BoolValue{}, nil, nil)
	const TTL = 5
	_, err := col.NewSTM(context.Background(), env.EtcdClient, func(stm col.STM) error {
		return clxn.ReadWrite(stm).PutTTL("key", epsilon, TTL)
	})
	require.NoError(t, err)

	var actualTTL int64
	_, err = col.NewSTM(context.Background(), env.EtcdClient, func(stm col.STM) error {
		var err error
		actualTTL, err = clxn.ReadWrite(stm).TTL("key")
		return err
	})
	require.NoError(t, err)
	require.True(t, actualTTL > 0 && actualTTL < TTL, "actualTTL was %v", actualTTL)

	// Put value with new, longer TLL and check that it was set
	const LongerTTL = 15
	_, err = col.NewSTM(context.Background(), env.EtcdClient, func(stm col.STM) error {
		return clxn.ReadWrite(stm).PutTTL("key", epsilon, LongerTTL)
	})
	require.NoError(t, err)

	_, err = col.NewSTM(context.Background(), env.EtcdClient, func(stm col.STM) error {
		var err error
		actualTTL, err = clxn.ReadWrite(stm).TTL("key")
		return err
	})
	require.NoError(t, err)
	require.True(t, actualTTL > TTL && actualTTL < LongerTTL, "actualTTL was %v", actualTTL)
}

func TestIteration(t *testing.T) {
	t.Parallel()
	env := testetcd.NewEnv(t)
	t.Run("one-val-per-txn", func(t *testing.T) {
		uuidPrefix := uuid.NewWithoutDashes()
		c := col.NewEtcdCollection(env.EtcdClient, uuidPrefix, nil, &col.TestItem{}, nil, nil)
		numVals := 1000
		for i := 0; i < numVals; i++ {
			_, err := col.NewSTM(context.Background(), env.EtcdClient, func(stm col.STM) error {
				testProto := makeProto(makeID(i))
				return c.ReadWrite(stm).Put(testProto.ID, testProto)
			})
			require.NoError(t, err)
		}
		ro := c.ReadOnly(context.Background())
		testProto := &col.TestItem{}
		i := numVals - 1
		require.NoError(t, ro.List(testProto, col.DefaultOptions(), func(string) error {
			require.Equal(t, fmt.Sprintf("%d", i), testProto.ID)
			i--
			return nil
		}))
	})
	t.Run("many-vals-per-txn", func(t *testing.T) {
		uuidPrefix := uuid.NewWithoutDashes()
		c := col.NewEtcdCollection(env.EtcdClient, uuidPrefix, nil, &col.TestItem{}, nil, nil)
		numBatches := 10
		valsPerBatch := 7
		for i := 0; i < numBatches; i++ {
			_, err := col.NewSTM(context.Background(), env.EtcdClient, func(stm col.STM) error {
				for j := 0; j < valsPerBatch; j++ {
					testProto := makeProto(makeID(i*valsPerBatch + j))
					if err := c.ReadWrite(stm).Put(testProto.ID, testProto); err != nil {
						return err
					}
				}
				return nil
			})
			require.NoError(t, err)
		}
		vals := make(map[string]bool)
		ro := c.ReadOnly(context.Background())
		testProto := &col.TestItem{}
		require.NoError(t, ro.List(testProto, col.DefaultOptions(), func(string) error {
			require.False(t, vals[testProto.ID], "saw value %s twice", testProto.ID)
			vals[testProto.ID] = true
			return nil
		}))
		require.Equal(t, numBatches*valsPerBatch, len(vals), "didn't receive every value")
	})
	t.Run("large-vals", func(t *testing.T) {
		uuidPrefix := uuid.NewWithoutDashes()
		c := col.NewEtcdCollection(env.EtcdClient, uuidPrefix, nil, &col.TestItem{}, nil, nil)
		numVals := 100
		longString := strings.Repeat("foo\n", 10) // 1 MB worth of foo
		for i := 0; i < numVals; i++ {
			_, err := col.NewSTM(context.Background(), env.EtcdClient, func(stm col.STM) error {
				id := fmt.Sprintf("%d", i)
				return c.ReadWrite(stm).Put(id, &col.TestItem{ID: id, Value: longString})
			})
			require.NoError(t, err)
		}
		ro := c.ReadOnly(context.Background())
		val := &col.TestItem{}
		vals := make(map[string]bool)
		valsOrder := []string{}
		require.NoError(t, ro.List(val, col.DefaultOptions(), func(string) error {
			require.False(t, vals[val.ID], "saw value %s twice", val.ID)
			vals[val.ID] = true
			valsOrder = append(valsOrder, val.ID)
			return nil
		}))
		for i, key := range valsOrder {
			require.Equal(t, key, strconv.Itoa(numVals-i-1), "incorrect order returned")
		}
		require.Equal(t, numVals, len(vals), "didn't receive every value")
		vals = make(map[string]bool)
		valsOrder = []string{}
		require.NoError(t, ro.List(val, &col.Options{col.SortByCreateRevision, col.SortAscend}, func(string) error {
			require.False(t, vals[val.ID], "saw value %s twice", val.ID)
			vals[val.ID] = true
			valsOrder = append(valsOrder, val.ID)
			return nil
		}))
		for i, key := range valsOrder {
			require.Equal(t, key, strconv.Itoa(i), "incorrect order returned")
		}
		require.Equal(t, numVals, len(vals), "didn't receive every value")
	})
}
