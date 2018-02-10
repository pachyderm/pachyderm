package server

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/enterprise"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
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
	require.NoError(t, err)
	require.NoError(t, backoff.Retry(func() error {
		resp, err := client.Enterprise.GetState(context.Background(),
			&enterprise.GetStateRequest{})
		if err != nil {
			return err
		}
		if resp.State != enterprise.State_ACTIVE {
			return fmt.Errorf("expected enterprise state to be ACTIVE but was %v", resp.State)
		}
		expires, err := types.TimestampFromProto(resp.Info.Expires)
		if err != nil {
			return err
		}
		if expires.Sub(time.Now()) <= year {
			return fmt.Errorf("Expected test token to expire >1yr in the future, but expires at %v (congratulations on making it to 2026!)", expires)
		}
		return nil
	}, backoff.NewTestingBackOff()))

	// Make current enterprise token expire
	expires := time.Now().Add(-30 * time.Second)
	expiresProto, err := types.TimestampProto(expires)
	require.NoError(t, err)
	_, err = client.Enterprise.Activate(context.Background(),
		&enterprise.ActivateRequest{
			ActivationCode: testutil.GetTestEnterpriseCode(),
			Expires:        expiresProto,
		})
	require.NoError(t, backoff.Retry(func() error {
		resp, err := client.Enterprise.GetState(context.Background(),
			&enterprise.GetStateRequest{})
		if err != nil {
			return err
		}
		if resp.State != enterprise.State_EXPIRED {
			return fmt.Errorf("expected enterprise state to be EXPIRED but was %v", resp.State)
		}
		respExpires, err := types.TimestampFromProto(resp.Info.Expires)
		if err != nil {
			return err
		}
		if expires.Unix() != respExpires.Unix() {
			return fmt.Errorf("expected enterprise expiration to be %v, but was %v", expires, respExpires)
		}
		return nil
	}, backoff.NewTestingBackOff()))
}
