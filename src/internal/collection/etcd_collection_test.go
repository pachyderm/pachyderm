package collection_test

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	etcd "go.etcd.io/etcd/client/v3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
)

var (
	pipelineIndex *col.Index = &col.Index{
		Name: "Pipeline",
		Extract: func(val proto.Message) string {
			return val.(*pps.JobInfo).Job.Pipeline.Name
		},
	}
)

func TestEtcdCollections(suite *testing.T) {
	ctx := pctx.TestContext(suite)
	etcdEnv := testetcd.NewEnv(ctx, suite)
	newCollection := func(ctx context.Context, t *testing.T, noIndex ...bool) (ReadCallback, WriteCallback) {
		prefix := testutil.UniqueString("test-etcd-collections-")
		index := []*col.Index{TestSecondaryIndex}
		if len(noIndex) > 0 && noIndex[0] {
			index = nil
		}
		testCol := col.NewEtcdCollection(etcdEnv.EtcdClient, prefix, index, &col.TestItem{}, nil, nil)

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

	collectionTests(ctx, suite, newCollection)
	watchTests(ctx, suite, newCollection)
}

func TestDryrun(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	env := testetcd.NewEnv(ctx, t)
	uuidPrefix := uuid.NewWithoutDashes()

	jobInfos := col.NewEtcdCollection(env.EtcdClient, uuidPrefix, nil, &pps.JobInfo{}, nil, nil)

	job := &pps.JobInfo{
		Job: client.NewJob(pfs.DefaultProjectName, "p1", "j1"),
	}
	err := col.NewDryrunSTM(context.Background(), env.EtcdClient, func(stm col.STM) error {
		return errors.EnsureStack(jobInfos.ReadWrite(stm).Put(ppsdb.JobKey(job.Job), job))
	})
	require.NoError(t, err)

	err = jobInfos.ReadOnly(context.Background()).Get("j1", job)
	require.True(t, col.IsErrNotFound(err))
}

func TestDeletePrefix(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	env := testetcd.NewEnv(ctx, t)
	uuidPrefix := uuid.NewWithoutDashes()

	jobInfos := col.NewEtcdCollection(env.EtcdClient, uuidPrefix, nil, &pps.JobInfo{}, nil, nil)

	j1 := &pps.JobInfo{
		Job: client.NewJob(pfs.DefaultProjectName, "p", "prefix/suffix/Job"),
	}
	j2 := &pps.JobInfo{
		Job: client.NewJob(pfs.DefaultProjectName, "p", "prefix/suffix/Job2"),
	}
	j3 := &pps.JobInfo{
		Job: client.NewJob(pfs.DefaultProjectName, "p", "prefix/Job3"),
	}
	j4 := &pps.JobInfo{
		Job: client.NewJob(pfs.DefaultProjectName, "p", "Job4"),
	}

	_, err := col.NewSTM(context.Background(), env.EtcdClient, func(stm col.STM) error {
		rw := jobInfos.ReadWrite(stm)
		if err := rw.Put(ppsdb.JobKey(j1.Job), j1); err != nil {
			return errors.EnsureStack(err)
		}
		if err := rw.Put(ppsdb.JobKey(j2.Job), j2); err != nil {
			return errors.EnsureStack(err)
		}
		if err := rw.Put(ppsdb.JobKey(j3.Job), j3); err != nil {
			return errors.EnsureStack(err)
		}
		if err := rw.Put(ppsdb.JobKey(j4.Job), j4); err != nil {
			return errors.EnsureStack(err)
		}
		return nil
	})
	require.NoError(t, err)

	_, err = col.NewSTM(context.Background(), env.EtcdClient, func(stm col.STM) error {
		job := &pps.JobInfo{}
		rw := jobInfos.ReadWrite(stm)

		if err := rw.DeleteAllPrefix("default/p@prefix/suffix"); err != nil {
			return errors.EnsureStack(err)
		}
		if err := rw.Get(ppsdb.JobKey(j1.Job), job); !col.IsErrNotFound(err) {
			return errors.Wrapf(err, "Expected ErrNotFound for key '%s', but got", j1.Job.Id)
		}
		if err := rw.Get(ppsdb.JobKey(j2.Job), job); !col.IsErrNotFound(err) {
			return errors.Wrapf(err, "Expected ErrNotFound for key '%s', but got", j2.Job.Id)
		}
		if err := rw.Get(ppsdb.JobKey(j3.Job), job); err != nil {
			return errors.EnsureStack(err)
		}
		if err := rw.Get(ppsdb.JobKey(j4.Job), job); err != nil {
			return errors.EnsureStack(err)
		}

		if err := rw.DeleteAllPrefix("default/p@prefix"); err != nil {
			return errors.EnsureStack(err)
		}
		if err := rw.Get(ppsdb.JobKey(j1.Job), job); !col.IsErrNotFound(err) {
			return errors.Wrapf(err, "Expected ErrNotFound for key '%s', but got", j1.Job.Id)
		}
		if err := rw.Get(ppsdb.JobKey(j2.Job), job); !col.IsErrNotFound(err) {
			return errors.Wrapf(err, "Expected ErrNotFound for key '%s', but got", j2.Job.Id)
		}
		if err := rw.Get(ppsdb.JobKey(j3.Job), job); !col.IsErrNotFound(err) {
			return errors.Wrapf(err, "Expected ErrNotFound for key '%s', but got", j3.Job.Id)
		}
		if err := rw.Get(ppsdb.JobKey(j4.Job), job); err != nil {
			return errors.EnsureStack(err)
		}

		if err := rw.Put(ppsdb.JobKey(j1.Job), j1); err != nil {
			return errors.EnsureStack(err)
		}
		if err := rw.Get(ppsdb.JobKey(j1.Job), job); err != nil {
			return errors.EnsureStack(err)
		}

		if err := rw.DeleteAllPrefix("default/p@prefix/suffix"); err != nil {
			return errors.EnsureStack(err)
		}
		if err := rw.Get(ppsdb.JobKey(j1.Job), job); !col.IsErrNotFound(err) {
			return errors.Wrapf(err, "Expected ErrNotFound for key '%s', but got", j1.Job.Id)
		}

		if err := rw.Put(ppsdb.JobKey(j2.Job), j2); err != nil {
			return errors.EnsureStack(err)
		}
		if err := rw.Get(ppsdb.JobKey(j2.Job), job); err != nil {
			return errors.EnsureStack(err)
		}

		return nil
	})
	require.NoError(t, err)

	job := &pps.JobInfo{}
	ro := jobInfos.ReadOnly(context.Background())
	require.True(t, col.IsErrNotFound(ro.Get(ppsdb.JobKey(j1.Job), job)))
	require.NoError(t, ro.Get(ppsdb.JobKey(j2.Job), job))
	require.Equal(t, j2, job)
	require.True(t, col.IsErrNotFound(ro.Get(ppsdb.JobKey(j3.Job), job)))
	require.NoError(t, ro.Get(ppsdb.JobKey(j4.Job), job))
	require.Equal(t, j4, job)
}

func TestIndex(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	env := testetcd.NewEnv(ctx, t)
	uuidPrefix := uuid.NewWithoutDashes()

	jobInfos := col.NewEtcdCollection(env.EtcdClient, uuidPrefix, []*col.Index{pipelineIndex}, &pps.JobInfo{}, nil, nil)

	j1 := &pps.JobInfo{
		Job: client.NewJob(pfs.DefaultProjectName, "p1", "j1"),
	}
	j2 := &pps.JobInfo{
		Job: client.NewJob(pfs.DefaultProjectName, "p1", "j2"),
	}
	j3 := &pps.JobInfo{
		Job: client.NewJob(pfs.DefaultProjectName, "p2", "j3"),
	}
	_, err := col.NewSTM(context.Background(), env.EtcdClient, func(stm col.STM) error {
		rw := jobInfos.ReadWrite(stm)
		if err := rw.Put(ppsdb.JobKey(j1.Job), j1); err != nil {
			return errors.EnsureStack(err)
		}
		if err := rw.Put(ppsdb.JobKey(j2.Job), j2); err != nil {
			return errors.EnsureStack(err)
		}
		if err := rw.Put(ppsdb.JobKey(j3.Job), j3); err != nil {
			return errors.EnsureStack(err)
		}
		return nil
	})
	require.NoError(t, err)

	ro := jobInfos.ReadOnly(context.Background())

	job := &pps.JobInfo{}
	i := 1
	require.NoError(t, ro.GetByIndex(pipelineIndex, j1.Job.Pipeline.Name, job, col.DefaultOptions(), func(string) error {
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
	require.NoError(t, ro.GetByIndex(pipelineIndex, j3.Job.Pipeline.Name, job, col.DefaultOptions(), func(string) error {
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
	ctx := pctx.TestContext(t)
	env := testetcd.NewEnv(ctx, t)
	uuidPrefix := uuid.NewWithoutDashes()
	index := &col.Index{
		Name: "Value",
		Extract: func(val proto.Message) string {
			return fmt.Sprintf("%v", val.(*wrapperspb.BoolValue).Value)
		},
	}
	boolValues := col.NewEtcdCollection(env.EtcdClient, uuidPrefix, []*col.Index{index}, &wrapperspb.BoolValue{}, nil, nil)

	r1 := &wrapperspb.BoolValue{
		Value: true,
	}
	r2 := &wrapperspb.BoolValue{
		Value: false,
	}
	_, err := col.NewSTM(ctx, env.EtcdClient, func(stm col.STM) error {
		boolValues := boolValues.ReadWrite(stm)
		if err := boolValues.Put("true", r1); err != nil {
			return errors.EnsureStack(err)
		}
		if err := boolValues.Put("false", r2); err != nil {
			return errors.EnsureStack(err)
		}
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

var epsilon = &wrapperspb.BoolValue{Value: true}

func TestTTL(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	env := testetcd.NewEnv(ctx, t)
	uuidPrefix := uuid.NewWithoutDashes()

	clxn := col.NewEtcdCollection(env.EtcdClient, uuidPrefix, nil, &wrapperspb.BoolValue{}, nil, nil)
	const TTL = 5
	_, err := col.NewSTM(ctx, env.EtcdClient, func(stm col.STM) error {
		return errors.EnsureStack(clxn.ReadWrite(stm).PutTTL("key", epsilon, TTL))
	})
	require.NoError(t, err)

	var actualTTL int64
	_, err = col.NewSTM(ctx, env.EtcdClient, func(stm col.STM) error {
		var err error
		actualTTL, err = clxn.ReadWrite(stm).TTL("key")
		return errors.EnsureStack(err)
	})
	require.NoError(t, err)
	require.True(t, actualTTL > 0 && actualTTL < TTL, "actualTTL was %v", actualTTL)
}

func TestTTLExpire(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	env := testetcd.NewEnv(ctx, t)
	uuidPrefix := uuid.NewWithoutDashes()

	clxn := col.NewEtcdCollection(env.EtcdClient, uuidPrefix, nil, &wrapperspb.BoolValue{}, nil, nil)
	const TTL = 5
	_, err := col.NewSTM(ctx, env.EtcdClient, func(stm col.STM) error {
		return errors.EnsureStack(clxn.ReadWrite(stm).PutTTL("key", epsilon, TTL))
	})
	require.NoError(t, err)

	time.Sleep((TTL + 1) * time.Second)
	value := &wrapperspb.BoolValue{}
	err = clxn.ReadOnly(ctx).Get("key", value)
	require.NotNil(t, err)
	require.True(t, errutil.IsNotFoundError(err))
}

func TestTTLExtend(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	env := testetcd.NewEnv(ctx, t)
	uuidPrefix := uuid.NewWithoutDashes()

	// Put value with short TLL & check that it was set
	clxn := col.NewEtcdCollection(env.EtcdClient, uuidPrefix, nil, &wrapperspb.BoolValue{}, nil, nil)
	const TTL = 5
	_, err := col.NewSTM(ctx, env.EtcdClient, func(stm col.STM) error {
		return errors.EnsureStack(clxn.ReadWrite(stm).PutTTL("key", epsilon, TTL))
	})
	require.NoError(t, err)

	var actualTTL int64
	_, err = col.NewSTM(ctx, env.EtcdClient, func(stm col.STM) error {
		var err error
		actualTTL, err = clxn.ReadWrite(stm).TTL("key")
		return errors.EnsureStack(err)
	})
	require.NoError(t, err)
	require.True(t, actualTTL > 0 && actualTTL < TTL, "actualTTL was %v", actualTTL)

	// Put value with new, longer TLL and check that it was set
	const LongerTTL = 15
	_, err = col.NewSTM(ctx, env.EtcdClient, func(stm col.STM) error {
		return errors.EnsureStack(clxn.ReadWrite(stm).PutTTL("key", epsilon, LongerTTL))
	})
	require.NoError(t, err)

	_, err = col.NewSTM(ctx, env.EtcdClient, func(stm col.STM) error {
		var err error
		actualTTL, err = clxn.ReadWrite(stm).TTL("key")
		return errors.EnsureStack(err)
	})
	require.NoError(t, err)
	require.True(t, actualTTL > TTL && actualTTL < LongerTTL, "actualTTL was %v", actualTTL)
}

func TestIteration(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	env := testetcd.NewEnv(ctx, t)
	t.Run("one-val-per-txn", func(t *testing.T) {
		uuidPrefix := uuid.NewWithoutDashes()
		c := col.NewEtcdCollection(env.EtcdClient, uuidPrefix, nil, &col.TestItem{}, nil, nil)
		numVals := 1000
		for i := 0; i < numVals; i++ {
			_, err := col.NewSTM(ctx, env.EtcdClient, func(stm col.STM) error {
				testProto := makeProto(makeID(i))
				return errors.EnsureStack(c.ReadWrite(stm).Put(testProto.Id, testProto))
			})
			require.NoError(t, err)
		}
		ro := c.ReadOnly(context.Background())
		testProto := &col.TestItem{}
		i := numVals - 1
		require.NoError(t, ro.List(testProto, col.DefaultOptions(), func(string) error {
			require.Equal(t, fmt.Sprintf("%d", i), testProto.Id)
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
			_, err := col.NewSTM(ctx, env.EtcdClient, func(stm col.STM) error {
				for j := 0; j < valsPerBatch; j++ {
					testProto := makeProto(makeID(i*valsPerBatch + j))
					if err := c.ReadWrite(stm).Put(testProto.Id, testProto); err != nil {
						return errors.EnsureStack(err)
					}
				}
				return nil
			})
			require.NoError(t, err)
		}
		vals := make(map[string]bool)
		ro := c.ReadOnly(ctx)
		testProto := &col.TestItem{}
		require.NoError(t, ro.List(testProto, col.DefaultOptions(), func(string) error {
			require.False(t, vals[testProto.Id], "saw value %s twice", testProto.Id)
			vals[testProto.Id] = true
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
			_, err := col.NewSTM(ctx, env.EtcdClient, func(stm col.STM) error {
				id := fmt.Sprintf("%d", i)
				return errors.EnsureStack(c.ReadWrite(stm).Put(id, &col.TestItem{Id: id, Value: longString}))
			})
			require.NoError(t, err)
		}
		ro := c.ReadOnly(context.Background())
		val := &col.TestItem{}
		vals := make(map[string]bool)
		valsOrder := []string{}
		require.NoError(t, ro.List(val, col.DefaultOptions(), func(string) error {
			require.False(t, vals[val.Id], "saw value %s twice", val.Id)
			vals[val.Id] = true
			valsOrder = append(valsOrder, val.Id)
			return nil
		}))
		for i, key := range valsOrder {
			require.Equal(t, key, strconv.Itoa(numVals-i-1), "incorrect order returned")
		}
		require.Equal(t, numVals, len(vals), "didn't receive every value")
		vals = make(map[string]bool)
		valsOrder = []string{}
		require.NoError(t, ro.List(val, &col.Options{Target: col.SortByCreateRevision, Order: col.SortAscend}, func(string) error {
			require.False(t, vals[val.Id], "saw value %s twice", val.Id)
			vals[val.Id] = true
			valsOrder = append(valsOrder, val.Id)
			return nil
		}))
		for i, key := range valsOrder {
			require.Equal(t, key, strconv.Itoa(i), "incorrect order returned")
		}
		require.Equal(t, numVals, len(vals), "didn't receive every value")
	})
}
