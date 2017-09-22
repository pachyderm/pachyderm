package server

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/enterprise"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/testutil"
)

var (
	pachClient *client.APIClient
	clientOnce sync.Once
)

const year = 365 * 24 * time.Hour

// getPachClient creates a seed client with a grpc connection to a pachyderm
// cluster
func getPachClient(t testing.TB) *client.APIClient {
	clientOnce.Do(func() {
		var err error
		if _, ok := os.LookupEnv("PACHD_PORT_650_TCP_ADDR"); ok {
			pachClient, err = client.NewInCluster()
		} else {
			pachClient, err = client.NewOnUserMachine(false, "user")
		}
		if err != nil {
			t.Fatalf("error getting Pachyderm client: %s", err.Error())
		}
	})
	return pachClient
}

func TestValidateActivationCode(t *testing.T) {
	_, err := validateActivationCode(testutil.GetTestEnterpriseCode())
	require.NoError(t, err)
}

func TestGetState(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	client := getPachClient(t)

	// Activate Pachyderm Enterprise and make sure the state is ACTIVE
	_, err := client.Enterprise.Activate(context.Background(),
		&enterprise.ActivateRequest{ActivationCode: testutil.GetTestEnterpriseCode()})
	resp, err := client.Enterprise.GetState(context.Background(),
		&enterprise.GetStateRequest{})
	require.NoError(t, err)
	require.Equal(t, resp.State, enterprise.State_ACTIVE)
	expires, err := types.TimestampFromProto(resp.Info.Expires)
	require.NoError(t, err)
	require.True(t, expires.Sub(time.Now()) > year)

	// Make current enterprise token expire
	expiresProto, err := types.TimestampProto(time.Now().Add(-30 * time.Second))
	require.NoError(t, err)
	_, err = client.Enterprise.Activate(context.Background(),
		&enterprise.ActivateRequest{
			ActivationCode: testutil.GetTestEnterpriseCode(),
			Expires:        expiresProto,
		})
	resp, err = client.Enterprise.GetState(context.Background(),
		&enterprise.GetStateRequest{})
	require.NoError(t, err)
	require.Equal(t, enterprise.State_EXPIRED, resp.State)
	require.Equal(t, expiresProto, resp.Info.Expires)
}
