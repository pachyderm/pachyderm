package server

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"

	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestAPIServer_validatePipelineRequest(t *testing.T) {
	var (
		a       = new(apiServer)
		request = &pps.CreatePipelineRequest{
			Pipeline: &pps.Pipeline{
				Project: &pfs.Project{
					Name: "0123456789ABCDEF",
				},
				Name: "0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF",
			},
		}
	)
	err := a.validatePipelineRequest(request)
	require.YesError(t, err)
	require.ErrorContains(t, err, fmt.Sprintf("is %d characters longer than the %d max", len(request.Pipeline.Name)-maxPipelineNameLength, maxPipelineNameLength))

	request.Pipeline.Name = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY"
	k8sName := ppsutil.PipelineRcName(&pps.PipelineInfo{Pipeline: &pps.Pipeline{Project: &pfs.Project{Name: request.Pipeline.GetProject().GetName()}, Name: request.Pipeline.Name}, Version: 99})
	err = a.validatePipelineRequest(request)
	require.YesError(t, err)
	require.ErrorContains(t, err, fmt.Sprintf("is %d characters longer than the %d max", len(k8sName)-dnsLabelLimit, dnsLabelLimit))
}

func TestNewMessageFilterFunc(t *testing.T) {
	ctx := pctx.TestContext(t)
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	filter, err := newMessageFilterFunc(`{ source: ., output: "" } | until(.source == "quux"; {"foo": "bar"}) | .output`, nil)
	require.NoError(t, err)
	go func() {
		_, err = filter(ctx, &pps.JobInfo{})
		require.YesError(t, err)
	}()
	<-ctx.Done()
	require.Equal(t, context.DeadlineExceeded, ctx.Err())
}
