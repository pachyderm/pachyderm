package server

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"

	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
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
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("is %d characters longer than the %d max", len(request.Pipeline.Name)-maxPipelineNameLength, maxPipelineNameLength))

	request.Pipeline.Name = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY"
	k8sName := ppsutil.PipelineRcName(&pps.PipelineInfo{Pipeline: &pps.Pipeline{Project: &pfs.Project{Name: request.Pipeline.GetProject().GetName()}, Name: request.Pipeline.Name}, Version: 99})
	err = a.validatePipelineRequest(request)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("is %d characters longer than the %d max", len(k8sName)-dnsLabelLimit, dnsLabelLimit))
}
