package chain

import (
	"context"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/common"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/datum"
)

type testHasher struct{}

func (th *testHasher) Hash(inputs []*common.Input) string {
	return common.HashDatum("", "", inputs)
}

type testIterator struct {
	metas []*datum.Meta
}

func newTestIterator(metas []*datum.Meta) *testIterator {
	return &testIterator{metas: metas}
}

func (ti *testIterator) Iterate(cb func(*datum.Meta) error) error {
	for _, meta := range ti.metas {
		if err := cb(meta); err != nil {
			return err
		}
	}
	return nil
}

func newTestChain(pachClient *client.APIClient, metas ...*datum.Meta) *JobChain {
	hasher := &testHasher{}
	if len(metas) > 0 {
		return NewJobChain(pachClient, hasher, WithBase(newTestIterator(metas)))
	}
	return NewJobChain(pachClient, hasher)
}

func newMeta(job *pps.Job, name string) *datum.Meta {
	inputs := []*common.Input{
		&common.Input{
			FileInfo: &pfs.FileInfo{
				File: &pfs.File{
					Path: name,
				},
			},
		},
	}
	return &datum.Meta{
		Job:    job,
		Inputs: inputs,
		Hash:   common.HashDatum("", "", inputs),
	}
}

func newTestMetas(job *pps.Job) []*datum.Meta {
	return []*datum.Meta{
		newMeta(job, "a"),
		newMeta(job, "b"),
		newMeta(job, "c"),
	}
}

func requireIteratorContents(t *testing.T, jdi *JobDatumIterator, metas []*datum.Meta) {
	require.NoError(t, jdi.Iterate(func(meta *datum.Meta) error {
		require.True(t, len(metas) > 0)
		require.Equal(t, metas[0].Inputs[0].FileInfo.File.Path, meta.Inputs[0].FileInfo.File.Path)
		metas = metas[1:]
		return nil
	}))
	require.Equal(t, 0, len(metas))
}

func TestEmptyBase(t *testing.T) {
	env := testpachd.NewRealEnv(t, testutil.NewTestDBConfig(t))
	chain := newTestChain(env.PachClient)
	job := client.NewJob("pipeline", uuid.NewWithoutDashes())
	jobMetas := newTestMetas(job)
	ti := newTestIterator(jobMetas)
	jdi := chain.CreateJob(context.Background(), job, ti, ti)
	requireIteratorContents(t, jdi, jobMetas)
}

func TestAdditiveOnBase(t *testing.T) {
	env := testpachd.NewRealEnv(t, testutil.NewTestDBConfig(t))
	chain := newTestChain(env.PachClient, newTestMetas(uuid.NewWithoutDashes())[:2]...)
	job := client.NewJob("pipeline", uuid.NewWithoutDashes())
	jobMetas := newTestMetas(job)
	ti := newTestIterator(jobMetas)
	jdi := chain.CreateJob(context.Background(), job, ti, ti)
	requireIteratorContents(t, jdi, jobMetas[2:])
}

func TestSubtractiveOnBase(t *testing.T) {
	env := testpachd.NewRealEnv(t, testutil.NewTestDBConfig(t))
	chain := newTestChain(env.PachClient, newTestMetas(uuid.NewWithoutDashes())...)
	job := client.NewJob("pipeline", uuid.NewWithoutDashes())
	jobMetas := newTestMetas(job)[1:]
	ti := newTestIterator(jobMetas)
	jdi := chain.CreateJob(context.Background(), job, ti, ti)
	requireIteratorContents(t, jdi, nil)
}

func TestAdditiveSubtractiveOnBase(t *testing.T) {
	env := testpachd.NewRealEnv(t, testutil.NewTestDBConfig(t))
	chain := newTestChain(env.PachClient, newTestMetas(uuid.NewWithoutDashes())[1:]...)
	job := client.NewJob("pipeline", uuid.NewWithoutDashes())
	jobMetas := newTestMetas(job)[:2]
	ti := newTestIterator(jobMetas)
	jdi := chain.CreateJob(context.Background(), job, ti, ti)
	requireIteratorContents(t, jdi, jobMetas[:1])
}

//func TestEmptyBase(t *testing.T) {
//	jobDatums := []string{"a", "b"}
//	chain := newTestChain(t, []string{})
//	job := newTestJob(jobDatums)
//	jdi, err := chain.Start(job)
//	require.NoError(t, err)
//	requireIteratorContents(t, jdi, jobDatums)
//
//	require.NoError(t, chain.Succeed(job))
//	requireChainEmpty(t, chain, jobDatums)
//}
//
//func TestAdditiveOnBase(t *testing.T) {
//	jobDatums := []string{"a", "b", "c"}
//	chain := newTestChain(t, []string{"a"})
//	job := newTestJob(jobDatums)
//	jdi, err := chain.Start(job)
//	require.NoError(t, err)
//	requireIteratorContents(t, jdi, []string{"b", "c"})
//
//	require.NoError(t, chain.Succeed(job))
//	requireChainEmpty(t, chain, jobDatums)
//}
//
//func TestSubtractiveOnBase(t *testing.T) {
//	jobDatums := []string{"a", "c"}
//	chain := newTestChain(t, []string{"a", "b", "c"})
//	job := newTestJob(jobDatums)
//	jdi, err := chain.Start(job)
//	require.NoError(t, err)
//	requireIteratorContents(t, jdi, jobDatums)
//
//	require.NoError(t, chain.Succeed(job))
//	requireChainEmpty(t, chain, jobDatums)
//}
//
//func TestAdditiveSubtractiveOnBase(t *testing.T) {
//	jobDatums := []string{"b", "c", "d", "e"}
//	chain := newTestChain(t, []string{"a", "b", "c"})
//	job := newTestJob(jobDatums)
//	jdi, err := chain.Start(job)
//	require.NoError(t, err)
//	requireIteratorContents(t, jdi, jobDatums)
//
//	require.NoError(t, chain.Succeed(job))
//	requireChainEmpty(t, chain, jobDatums)
//}
//
//// Read from a channel until we have the expected datums, then verify they
//// are correct, then make sure the channel doesn't have anything else.
//func requireDatums(t *testing.T, datumChan <-chan string, expected []string) {
//	// Recvs should be near-instant, but set a decently long timeout to avoid flakiness
//	actual := []string{}
//loop:
//	for range expected {
//		select {
//		case x, ok := <-datumChan:
//			if !ok {
//				require.ElementsEqual(t, expected, actual)
//			}
//			actual = append(actual, x)
//		case <-time.After(time.Second):
//			break loop
//		}
//	}
//	require.ElementsEqual(t, expected, actual)
//
//	select {
//	case x, ok := <-datumChan:
//		require.False(t, ok, "datum channel contains extra datum: %s", x)
//	default:
//	}
//}
//
//func requireChannelClosed(t *testing.T, c <-chan string) {
//	select {
//	case x, ok := <-c:
//		require.False(t, ok, "datum channel should be closed, but found extra datum: %s", x)
//	case <-time.After(time.Second):
//		require.True(t, false, "datum channel should be closed, but it is blocked")
//	}
//}
//
//func requireChannelBlocked(t *testing.T, c <-chan string) {
//	select {
//	case x, ok := <-c:
//		require.True(t, ok, "datum channel should be blocked, but it is closed")
//		require.True(t, false, "datum channel should be blocked, but it contains datum: %s", x)
//	default:
//	}
//}
//
//func superviseTestJobWithError(
//	ctx context.Context,
//	eg *errgroup.Group,
//	jdi JobDatumIterator,
//	expectedErr string,
//) <-chan string {
//	datumsChan := make(chan string)
//	eg.Go(func() (retErr error) {
//		defer func() {
//			if retErr != nil && expectedErr != "" && strings.Contains(retErr.Error(), expectedErr) {
//				retErr = nil
//			}
//		}()
//
//		defer close(datumsChan)
//		for {
//			count, err := jdi.NextBatch(ctx)
//			if err != nil {
//				return err
//			}
//			if count == 0 {
//				return nil
//			}
//
//			for i := int64(0); i < count; i++ {
//				inputs, _ := jdi.NextDatum()
//				datum, err := inputsToDatum(inputs)
//				if err != nil {
//					return err
//				}
//
//				datumsChan <- datum
//			}
//		}
//	})
//
//	return datumsChan
//}
//
//func superviseTestJob(ctx context.Context, eg *errgroup.Group, jdi JobDatumIterator) <-chan string {
//	return superviseTestJobWithError(ctx, eg, jdi, "")
//}
//
//// Job 1: ABCD   -> 1. Succeed
//// Job 2:   CDEF  -> 2. Succeed
//// Job 3: AB DE GH -> 3. Succeed
//func TestSuccess(t *testing.T) {
//	chain := newTestChain(t, []string{})
//	job1 := newTestJob([]string{"a", "b", "c", "d"})
//	job2 := newTestJob([]string{"c", "d", "e", "f"})
//	job3 := newTestJob([]string{"a", "b", "d", "e", "g", "h"})
//
//	eg, ctx := errgroup.WithContext(context.Background())
//
//	jdi1, err := chain.Start(job1)
//	require.NoError(t, err)
//	datums1 := superviseTestJob(ctx, eg, jdi1)
//
//	jdi2, err := chain.Start(job2)
//	require.NoError(t, err)
//	datums2 := superviseTestJob(ctx, eg, jdi2)
//
//	jdi3, err := chain.Start(job3)
//	require.NoError(t, err)
//	datums3 := superviseTestJob(ctx, eg, jdi3)
//
//	requireDatums(t, datums1, []string{"a", "b", "c", "d"})
//	requireDatums(t, datums2, []string{"e", "f"})
//	requireDatums(t, datums3, []string{"g", "h"})
//	requireChannelClosed(t, datums1)
//	requireChannelBlocked(t, datums2)
//	requireChannelBlocked(t, datums3)
//
//	require.NoError(t, chain.Succeed(job1))
//	requireDatums(t, datums2, []string{"c", "d"})
//	requireDatums(t, datums3, []string{"a", "b"})
//	requireChannelClosed(t, datums2)
//
//	require.NoError(t, chain.Succeed(job2))
//	requireDatums(t, datums3, []string{"d", "e"})
//	requireChannelClosed(t, datums3)
//
//	require.NoError(t, chain.Succeed(job3))
//	require.NoError(t, eg.Wait())
//
//	requireChainEmpty(t, chain, []string{"a", "b", "d", "e", "g", "h"})
//}
//
//// Job 1: ABCD   -> 1. Fail
//// Job 2:   CDEF  -> 2. Fail
//// Job 3: AB DE GH -> 3. Succeed
//func TestFail(t *testing.T) {
//	chain := newTestChain(t, []string{})
//	job1 := newTestJob([]string{"a", "b", "c", "d"})
//	job2 := newTestJob([]string{"c", "d", "e", "f"})
//	job3 := newTestJob([]string{"a", "b", "d", "e", "g", "h"})
//
//	eg, ctx := errgroup.WithContext(context.Background())
//
//	jdi1, err := chain.Start(job1)
//	require.NoError(t, err)
//	datums1 := superviseTestJob(ctx, eg, jdi1)
//
//	jdi2, err := chain.Start(job2)
//	require.NoError(t, err)
//	datums2 := superviseTestJob(ctx, eg, jdi2)
//
//	jdi3, err := chain.Start(job3)
//	require.NoError(t, err)
//	datums3 := superviseTestJob(ctx, eg, jdi3)
//
//	requireDatums(t, datums1, []string{"a", "b", "c", "d"})
//	requireDatums(t, datums2, []string{"e", "f"})
//	requireDatums(t, datums3, []string{"g", "h"})
//	requireChannelClosed(t, datums1)
//	requireChannelBlocked(t, datums2)
//	requireChannelBlocked(t, datums3)
//
//	require.NoError(t, chain.Fail(job1))
//	requireDatums(t, datums2, []string{"c", "d"})
//	requireDatums(t, datums3, []string{"a", "b"})
//	requireChannelClosed(t, datums2)
//
//	require.NoError(t, chain.Fail(job2))
//	requireDatums(t, datums3, []string{"d", "e"})
//	requireChannelClosed(t, datums3)
//
//	require.NoError(t, chain.Succeed(job3))
//	require.NoError(t, eg.Wait())
//
//	requireChainEmpty(t, chain, []string{"a", "b", "d", "e", "g", "h"})
//}
//
//// Job 1: AB   -> 1. Succeed
//// Job 2: ABC  -> 2. Succeed
//func TestAdditiveSuccess(t *testing.T) {
//	chain := newTestChain(t, []string{})
//	job1 := newTestJob([]string{"a", "b"})
//	job2 := newTestJob([]string{"a", "b", "c"})
//
//	eg, ctx := errgroup.WithContext(context.Background())
//
//	jdi1, err := chain.Start(job1)
//	require.NoError(t, err)
//	datums1 := superviseTestJob(ctx, eg, jdi1)
//
//	jdi2, err := chain.Start(job2)
//	require.NoError(t, err)
//	datums2 := superviseTestJob(ctx, eg, jdi2)
//
//	requireDatums(t, datums1, []string{"a", "b"})
//	requireDatums(t, datums2, []string{"c"})
//	requireChannelClosed(t, datums1)
//	requireChannelBlocked(t, datums2)
//
//	require.NoError(t, chain.Succeed(job1))
//	requireChannelClosed(t, datums2)
//
//	require.NoError(t, chain.Succeed(job2))
//	require.NoError(t, eg.Wait())
//
//	requireChainEmpty(t, chain, []string{"a", "b", "c"})
//}
//
//// Job 1: AB   -> 1. Fail
//// Job 2: ABC  -> 2. Succeed
//func TestAdditiveFail(t *testing.T) {
//	chain := newTestChain(t, []string{})
//	job1 := newTestJob([]string{"a", "b"})
//	job2 := newTestJob([]string{"a", "b", "c"})
//
//	eg, ctx := errgroup.WithContext(context.Background())
//
//	jdi1, err := chain.Start(job1)
//	require.NoError(t, err)
//	datums1 := superviseTestJob(ctx, eg, jdi1)
//
//	jdi2, err := chain.Start(job2)
//	require.NoError(t, err)
//	datums2 := superviseTestJob(ctx, eg, jdi2)
//
//	requireDatums(t, datums1, []string{"a", "b"})
//	requireDatums(t, datums2, []string{"c"})
//	requireChannelClosed(t, datums1)
//	requireChannelBlocked(t, datums2)
//
//	require.NoError(t, chain.Fail(job1))
//	requireDatums(t, datums2, []string{"a", "b"})
//	requireChannelClosed(t, datums2)
//
//	require.NoError(t, chain.Succeed(job2))
//	require.NoError(t, eg.Wait())
//
//	requireChainEmpty(t, chain, []string{"a", "b", "c"})
//}
//
//// Job 1: AB   -> 1. Succeed
//// Job 2:  BC  -> 2. Succeed
//// Job 3:  BCD -> 3. Succeed
//func TestCascadeSuccess(t *testing.T) {
//	chain := newTestChain(t, []string{})
//	job1 := newTestJob([]string{"a", "b"})
//	job2 := newTestJob([]string{"b", "c"})
//	job3 := newTestJob([]string{"b", "c", "d"})
//
//	eg, ctx := errgroup.WithContext(context.Background())
//
//	jdi1, err := chain.Start(job1)
//	require.NoError(t, err)
//	datums1 := superviseTestJob(ctx, eg, jdi1)
//
//	jdi2, err := chain.Start(job2)
//	require.NoError(t, err)
//	datums2 := superviseTestJob(ctx, eg, jdi2)
//
//	jdi3, err := chain.Start(job3)
//	require.NoError(t, err)
//	datums3 := superviseTestJob(ctx, eg, jdi3)
//
//	requireDatums(t, datums1, []string{"a", "b"})
//	requireDatums(t, datums2, []string{"c"})
//	requireDatums(t, datums3, []string{"d"})
//	requireChannelClosed(t, datums1)
//	requireChannelBlocked(t, datums2)
//	requireChannelBlocked(t, datums3)
//
//	require.NoError(t, chain.Succeed(job1))
//	requireDatums(t, datums2, []string{"b"})
//	requireChannelClosed(t, datums2)
//	requireChannelBlocked(t, datums3)
//
//	require.NoError(t, chain.Succeed(job2))
//	requireChannelClosed(t, datums3)
//
//	require.NoError(t, chain.Succeed(job3))
//	require.NoError(t, eg.Wait())
//
//	requireChainEmpty(t, chain, []string{"b", "c", "d"})
//}
//
//// Job 1: AB   -> 1. Succeed
//// Job 2: ABC  -> 2. Fail
//// Job 3: ABCD -> 3. Succeed
//func TestCascadeFail(t *testing.T) {
//	chain := newTestChain(t, []string{})
//	job1 := newTestJob([]string{"a", "b"})
//	job2 := newTestJob([]string{"a", "b", "c"})
//	job3 := newTestJob([]string{"a", "b", "c", "d"})
//
//	eg, ctx := errgroup.WithContext(context.Background())
//
//	jdi1, err := chain.Start(job1)
//	require.NoError(t, err)
//	datums1 := superviseTestJob(ctx, eg, jdi1)
//
//	jdi2, err := chain.Start(job2)
//	require.NoError(t, err)
//	datums2 := superviseTestJob(ctx, eg, jdi2)
//
//	jdi3, err := chain.Start(job3)
//	require.NoError(t, err)
//	datums3 := superviseTestJob(ctx, eg, jdi3)
//
//	requireDatums(t, datums1, []string{"a", "b"})
//	requireDatums(t, datums2, []string{"c"})
//	requireDatums(t, datums3, []string{"d"})
//	requireChannelClosed(t, datums1)
//	requireChannelBlocked(t, datums2)
//	requireChannelBlocked(t, datums3)
//
//	require.NoError(t, chain.Succeed(job1))
//	requireChannelClosed(t, datums2)
//	requireChannelBlocked(t, datums3)
//
//	require.NoError(t, chain.Fail(job2))
//	requireDatums(t, datums3, []string{"c"})
//	requireChannelClosed(t, datums3)
//
//	require.NoError(t, chain.Succeed(job3))
//	require.NoError(t, eg.Wait())
//
//	requireChainEmpty(t, chain, []string{"a", "b", "c", "d"})
//}
//
//// Job 1: AB   -> 2. Succeed
//// Job 2:  BC  -> 1. Fail
//// Job 3:  BCD -> 3. Succeed
//func TestSplitFail(t *testing.T) {
//	chain := newTestChain(t, []string{})
//	job1 := newTestJob([]string{"a", "b"})
//	job2 := newTestJob([]string{"b", "c"})
//	job3 := newTestJob([]string{"b", "c", "d"})
//
//	eg, ctx := errgroup.WithContext(context.Background())
//
//	jdi1, err := chain.Start(job1)
//	require.NoError(t, err)
//	datums1 := superviseTestJob(ctx, eg, jdi1)
//
//	jdi2, err := chain.Start(job2)
//	require.NoError(t, err)
//	datums2 := superviseTestJobWithError(ctx, eg, jdi2, "job failed")
//
//	jdi3, err := chain.Start(job3)
//	require.NoError(t, err)
//	datums3 := superviseTestJob(ctx, eg, jdi3)
//
//	requireDatums(t, datums1, []string{"a", "b"})
//	requireDatums(t, datums2, []string{"c"})
//	requireDatums(t, datums3, []string{"d"})
//	requireChannelClosed(t, datums1)
//	requireChannelBlocked(t, datums2)
//	requireChannelBlocked(t, datums3)
//
//	require.NoError(t, chain.Fail(job2))
//	requireDatums(t, datums3, []string{"c"})
//	//requireChannelClosed(t, datums2)
//	requireChannelBlocked(t, datums3)
//
//	require.NoError(t, chain.Succeed(job1))
//	requireDatums(t, datums3, []string{"b"})
//	requireChannelClosed(t, datums3)
//
//	require.NoError(t, chain.Succeed(job3))
//	require.NoError(t, eg.Wait())
//
//	requireChainEmpty(t, chain, []string{"b", "c", "d"})
//}
//
//// Job 1: AB   -> 1. Succeed (A and B recovered)
//// Job 2: ABC  -> 2. Succeed (A and C recovered)
//// Job 3: ABCD -> 3. Succeed
//func TestRecoveredDatums(t *testing.T) {
//	chain := newTestChain(t, []string{})
//	job1 := newTestJob([]string{"a", "b"})
//	job2 := newTestJob([]string{"a", "b", "c"})
//	job3 := newTestJob([]string{"a", "b", "c", "d"})
//
//	eg, ctx := errgroup.WithContext(context.Background())
//
//	jdi1, err := chain.Start(job1)
//	require.NoError(t, err)
//	datums1 := superviseTestJob(ctx, eg, jdi1)
//
//	jdi2, err := chain.Start(job2)
//	require.NoError(t, err)
//	datums2 := superviseTestJob(ctx, eg, jdi2)
//
//	jdi3, err := chain.Start(job3)
//	require.NoError(t, err)
//	datums3 := superviseTestJob(ctx, eg, jdi3)
//
//	requireDatums(t, datums1, []string{"a", "b"})
//	requireDatums(t, datums2, []string{"c"})
//	requireDatums(t, datums3, []string{"d"})
//	requireChannelClosed(t, datums1)
//	requireChannelBlocked(t, datums2)
//	requireChannelBlocked(t, datums3)
//
//	require.NoError(t, chain.RecoveredDatums(job1, datumsToSet([]string{"a", "b"})))
//	require.NoError(t, chain.Succeed(job1))
//	requireDatums(t, datums2, []string{"a", "b"})
//	requireChannelClosed(t, datums2)
//	requireChannelBlocked(t, datums3)
//
//	require.NoError(t, chain.RecoveredDatums(job2, datumsToSet([]string{"a", "c"})))
//	require.NoError(t, chain.Succeed(job2))
//	requireDatums(t, datums3, []string{"a", "c"})
//	requireChannelClosed(t, datums3)
//
//	require.NoError(t, chain.RecoveredDatums(job3, datumsToSet([]string{"c", "d"})))
//	require.NoError(t, chain.Succeed(job3))
//	require.NoError(t, eg.Wait())
//
//	requireChainEmpty(t, chain, []string{"a", "b"})
//}
//
//func TestEarlySuccess(t *testing.T) {
//	chain := newTestChain(t, []string{})
//	job1 := newTestJob([]string{"a", "b"})
//
//	_, err := chain.Start(job1)
//	require.NoError(t, err)
//
//	require.YesError(t, chain.Succeed(job1), "items remaining")
//}
//
//func TestEarlyFail(t *testing.T) {
//	chain := newTestChain(t, []string{"e", "f"})
//	job := newTestJob([]string{"a", "b"})
//
//	_, err := chain.Start(job)
//	require.NoError(t, err)
//
//	require.NoError(t, chain.Fail(job))
//	requireChainEmpty(t, chain, []string{"e", "f"})
//}
//
//func TestRepeatedDatumAdditiveSubtractiveOnBase(t *testing.T) {
//	jobDatums := []string{"c", "c", "b"}
//	chain := newTestChain(t, []string{"a", "b", "a", "b", "c"})
//	job := newTestJob(jobDatums)
//
//	jdi, err := chain.Start(job)
//	require.NoError(t, err)
//
//	requireIteratorContents(t, jdi, jobDatums)
//	require.NoError(t, chain.Succeed(job))
//
//	requireChainEmpty(t, chain, jobDatums)
//}
//
//func TestRepeatedDatumSubtractiveOnBase(t *testing.T) {
//	jobDatums := []string{"a", "a"}
//	chain := newTestChain(t, []string{"a", "b", "a", "b", "c"})
//	job := newTestJob(jobDatums)
//
//	jdi, err := chain.Start(job)
//	require.NoError(t, err)
//
//	requireIteratorContents(t, jdi, jobDatums)
//	require.NoError(t, chain.Succeed(job))
//
//	requireChainEmpty(t, chain, jobDatums)
//}
//
//func TestRepeatedDatumAdditiveOnBase(t *testing.T) {
//	baseDatums := []string{"a", "b", "a", "b", "c"}
//	newDatums := []string{"a", "c", "d"}
//	jobDatums := append([]string{}, baseDatums...)
//	jobDatums = append(jobDatums, newDatums...)
//
//	chain := newTestChain(t, baseDatums)
//	job := newTestJob(jobDatums)
//
//	jdi, err := chain.Start(job)
//	require.NoError(t, err)
//
//	requireIteratorContents(t, jdi, newDatums)
//	require.NoError(t, chain.Succeed(job))
//
//	requireChainEmpty(t, chain, jobDatums)
//}
//
//func TestRepeatedDatumWithoutBase(t *testing.T) {
//	jobDatums := []string{"a", "b", "c", "a", "b", "a"}
//	chain := newTestChain(t, []string{})
//	job := newTestJob(jobDatums)
//
//	jdi, err := chain.Start(job)
//	require.NoError(t, err)
//
//	requireIteratorContents(t, jdi, jobDatums)
//	require.NoError(t, chain.Succeed(job))
//
//	requireChainEmpty(t, chain, jobDatums)
//}
//
//func TestNoSkipSuccess(t *testing.T) {
//	chain := NewNoSkipJobChain(&testHasher{})
//	job1 := newTestJob([]string{"a", "b", "c", "d"})
//	job2 := newTestJob([]string{"b", "c", "d", "e"})
//	job3 := newTestJob([]string{"a", "f", "g"})
//	job4 := newTestJob([]string{"h", "i"})
//
//	eg, ctx := errgroup.WithContext(context.Background())
//
//	jdi1, err := chain.Start(job1)
//	require.NoError(t, err)
//	datums1 := superviseTestJob(ctx, eg, jdi1)
//
//	jdi2, err := chain.Start(job2)
//	require.NoError(t, err)
//	datums2 := superviseTestJob(ctx, eg, jdi2)
//
//	jdi3, err := chain.Start(job3)
//	require.NoError(t, err)
//	datums3 := superviseTestJob(ctx, eg, jdi3)
//
//	jdi4, err := chain.Start(job4)
//	require.NoError(t, err)
//	datums4 := superviseTestJob(ctx, eg, jdi4)
//
//	requireDatums(t, datums1, []string{"a", "b", "c", "d"})
//	requireDatums(t, datums2, []string{"b", "c", "d", "e"})
//	requireDatums(t, datums3, []string{"a", "f", "g"})
//	requireDatums(t, datums4, []string{"h", "i"})
//	requireChannelClosed(t, datums1)
//	requireChannelClosed(t, datums2)
//	requireChannelClosed(t, datums3)
//	requireChannelClosed(t, datums4)
//
//	require.NoError(t, chain.Succeed(job1))
//	require.NoError(t, chain.Succeed(job2))
//	require.NoError(t, chain.Succeed(job3))
//	require.NoError(t, chain.Succeed(job4))
//	require.NoError(t, eg.Wait())
//}
