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
)

const testActivationCode = `eyJ0b2tlbiI6IntcImV4cGlyeVwiOlwiMjAyNy0wNy0xMlQwM` +
	`zowOTowNi4xODBaXCIsXCJzY29wZXNcIjp7XCJiYXNpY1wiOnRydWV9LFwibmFtZVwiOlwicGF` +
	`jaHlkZXJtRW5naW5lZXJpbmdcIn0iLCJzaWduYXR1cmUiOiJWdjZQbEkrL3RJamlWYUNHMGw0T` +
	`Ud6WldDS2YrUFMyWS9WUzFkZ3plcVdjS3RETlJvdkRnSnd3TXFXbWdCOUs5a2lPemVQRlh4eTh` +
	`3U2dMbTJ4dnBlTHN2bGlsTlc5MEhKbGxxcjhKWEVTbVV4R2tKQldMTHZHak5mYUlHZ0IvZTFEM` +
	`zQzMi95eUVnSW1LZDlpZ3J3RXZsRCtGdW0wa1hqS3Rrb2pPRmhkMDR6RHFEMSt5ZWpsTmRtUzB` +
	`TaDJKWHRTMnFqWk0zTE5lWlpTRldLcEVJTmlXa2dhOTdTNUw2ZVlCdXFZcFJLMTkwd1pXNTVCO` +
	`VFJSHJNNWtDWGQrWUN5aTh0QU9kcFY2a3FMSDNoVGgxVDIwVjYveFNZNUVheHZObm8yRmFYbDU` +
	`yQzRFSWIvZ05RWW8xVExDd1hJN0FYL2lpL0VTckVBQmYzdDlYZmlwWGxleE9OMmhJaWY5dDROZ` +
	`FBaQ1pmYlErbW8vSlQ3Um5VTGpTb2J3alNWVk1qMUozLzZKbmhQRFpFSWNDdlVvUnMyL2M2WUZ` +
	`xOVo1TFRJNkUxV2Q0bE1RczRJYXVsTHVQOEFVa3R3ejBiQmY2dUhPd3VvTlk4UjJ3ZTA1MmUxW` +
	`VVGbmNyUE4wd2ZJVHo5Vm51M1dNcktpaDhhRzNmMzRLb2x0R3hpWXJHL2JZQjgweUFaTytCbzF` +
	`mTTJwaDB0emRXejFLR0lNQUlEbjBFWHU2V0duSUFFUWN1NHVFc1pSVXRzNFhuYk5PTC9vYU1NK` +
	`3RLV3UzdnFMdEhMWWlPaWZHNHpEcUxwYnNNN2NhZGNXWjJ3QzNoZVh6Y1loaUwzMHJlOGJ4MFc` +
	`3Vm1FOSt4elJHZisyNEdvRjFaS1BvaDNhY3hCS0dsZzRxN2JQd0c3QWJESmxkak1HbkVEdz0if` +
	`Q==`

var (
	pachClient *client.APIClient
	clientOnce sync.Once
)

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
			t.Fatal("error getting Pachyderm client: %s", err.Error())
		}
	})
	return pachClient
}

func TestValidateActivationCode(t *testing.T) {
	_, err := validateActivationCode(testActivationCode)
	require.NoError(t, err)
}

func TestGetState(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	client := getPachClient(t)

	// Activate Pachyderm Enterprise and make sure the state is ACTIVE
	_, err := client.Enterprise.Activate(context.Background(),
		&enterprise.ActivateRequest{ActivationCode: testActivationCode})
	resp, err := client.Enterprise.GetState(context.Background(), &enterprise.GetStateRequest{})
	require.NoError(t, err)
	require.Equal(t, resp.State, enterprise.State_ACTIVE)

	// Make current enterprise token expire
	expiresProto, err := types.TimestampProto(time.Now().Add(-30 * time.Second))
	require.NoError(t, err)
	_, err = client.Enterprise.Activate(context.Background(),
		&enterprise.ActivateRequest{
			ActivationCode: testActivationCode,
			Expires:        expiresProto,
		})
	resp, err = client.Enterprise.GetState(context.Background(), &enterprise.GetStateRequest{})
	require.NoError(t, err)
	require.Equal(t, enterprise.State_EXPIRED, resp.State)
}
