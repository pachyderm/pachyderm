package pjs

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/pjs"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestCreateJob(t *testing.T) {
	ctx := pctx.TestContext(t)
	c := NewTestClient(t)
	inputData := []byte("input data")
	res, err := c.CreateJob(ctx, &pjs.CreateJobRequest{
		Spec: &anypb.Any{
			TypeUrl: "pachyderm.io/test",
			Value:   []byte("hello world"),
		},
		Input: &pjs.QueueElement{
			Data: inputData,
		},
	})
	require.NoError(t, err)
	t.Log(res)
	res2, err := c.InspectJob(ctx, &pjs.InspectJobRequest{
		Job: res.Id,
	})
	require.NoError(t, err)
	t.Log(res2)
	require.NotNil(t, res2.Details)
	require.NotNil(t, res2.Details.JobInfo)
	jobInfo := res2.Details.JobInfo
	require.Equal(t, jobInfo.Input.Data, inputData)
	require.Equal(t, jobInfo.State, pjs.JobState_QUEUED)
}

func TestDo(t *testing.T) {
	ctx := pctx.TestContext(t)
	c := NewTestClient(t)
	in := &pjs.CreateJobRequest{
		Input: &pjs.QueueElement{
			Data: []byte("hello"),
		},
	}
	out, err := Do(ctx, c, in, func(qe *pjs.QueueElement) (*pjs.QueueElement, error) {
		// f(x) = x + x as a service
		var data []byte
		data = append(data, qe.Data...)
		data = append(data, qe.Data...)
		return &pjs.QueueElement{Data: data}, nil
	})
	require.NoError(t, err)
	expected := &pjs.QueueElement{
		Data: []byte("hellohello"),
	}
	require.Equal(t, expected, out)
}
