package server

import (
	"context"
	"sync"
	"testing"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/enterprise"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"google.golang.org/grpc"
)

type MyEnterprise struct {
	enterprise.APIClient
}

// GetState is required to keep the workerPodSpec from attempting to **actually** contact pachd
func (*MyEnterprise) GetState(ctx context.Context, in *enterprise.GetStateRequest, opts ...grpc.CallOption) (*enterprise.GetStateResponse, error) {
	return &enterprise.GetStateResponse{
		State: enterprise.State_NONE,
	}, nil
}

func TestIssue3483(t *testing.T) {

	var api *apiServer
	{
		var myOnce sync.Once
		myOnce = sync.Once{}
		// burn the Once to keep pachClientOnce from running the real initializer function
		myOnce.Do(func() {})

		var pachClient *client.APIClient
		pachClient = &client.APIClient{
			Enterprise: &MyEnterprise{},
		}
		api = &apiServer{
			pachClient:     pachClient,
			pachClientOnce: myOnce,
			workerImage:    "example.com/nope",
		}
	}
	expectedVolumeName := "vol0"
	jsonPatch0 := `[
		{
			"op": "add",
			"path": "/volumes/0",
			"value": {
				"name": "` + expectedVolumeName + `",
				"hostPath": {
					"path": "/volumePath"
				}
			}
		}
	]`

	options := &workerOptions{
		cacheSize: "1G",
		podPatch:  jsonPatch0,
	}
	podSpec, err := api.workerPodSpec(options)
	require.NoError(t, err)
	vol0 := podSpec.Volumes[0]

	// this demonstrates the bad behavior that was happening before
	require.Nil(t, vol0.EmptyDir)

	require.Equal(t, "/volumePath", vol0.HostPath.Path)
	volumeName := vol0.Name
	require.Equal(t, expectedVolumeName, volumeName)
}
