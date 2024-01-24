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
	require.Equal(t, res2.Details.JobInfo.Input.Data, inputData)
}
