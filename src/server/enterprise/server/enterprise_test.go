package server

import (
	"testing"
	"time"

	"github.com/gogo/protobuf/types"

	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/license"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	lc "github.com/pachyderm/pachyderm/v2/src/license"
)

const year = 365 * 24 * time.Hour

func TestValidateActivationCode(t *testing.T) {
	_, err := license.Validate(testutil.GetTestEnterpriseCode(t))
	require.NoError(t, err)
}

func TestGetState(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	testutil.DeleteAll(t)
	defer testutil.DeleteAll(t)
	client := testutil.GetPachClient(t)

	testutil.ActivateEnterprise(t, client)

	resp, err := client.Enterprise.GetState(client.Ctx(), &enterprise.GetStateRequest{})
	require.NoError(t, err)
	require.Equal(t, resp.State, enterprise.State_ACTIVE)

	expires, err := types.TimestampFromProto(resp.Info.Expires)
	require.NoError(t, err)
	require.True(t, time.Until(expires) >= year)

	activationCode, err := license.Unmarshal(resp.ActivationCode)
	require.NoError(t, err)

	require.Equal(t, "", activationCode.Signature)

	// Make current enterprise token expire
	expires = time.Now().Add(-30 * time.Second)
	expiresProto, err := types.TimestampProto(expires)
	require.NoError(t, err)

	_, err = client.License.Activate(client.Ctx(),
		&lc.ActivateRequest{
			ActivationCode: testutil.GetTestEnterpriseCode(t),
			Expires:        expiresProto,
		})
	require.NoError(t, err)

	_, err = client.Enterprise.Activate(client.Ctx(),
		&enterprise.ActivateRequest{
			Id:            "localhost",
			Secret:        "localhost",
			LicenseServer: "localhost:650",
		})
	require.NoError(t, err)

	resp, err = client.Enterprise.GetState(client.Ctx(), &enterprise.GetStateRequest{})
	require.NoError(t, err)
	require.Equal(t, resp.State, enterprise.State_EXPIRED)

	activationCode, err = license.Unmarshal(resp.ActivationCode)
	require.NoError(t, err)

	require.Equal(t, "", activationCode.Signature)
}

func TestGetActivationCode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	testutil.DeleteAll(t)
	defer testutil.DeleteAll(t)
	client := testutil.GetPachClient(t)

	testutil.ActivateEnterprise(t, client)

	resp, err := client.Enterprise.GetActivationCode(client.Ctx(), &enterprise.GetActivationCodeRequest{})
	require.NoError(t, err)
	require.Equal(t, testutil.GetTestEnterpriseCode(t), resp.ActivationCode)
	require.Equal(t, resp.State, enterprise.State_ACTIVE)

	// Make current enterprise token expire
	expires := time.Now().Add(-30 * time.Second)
	expiresProto, err := types.TimestampProto(expires)
	require.NoError(t, err)
	_, err = client.License.Activate(client.Ctx(),
		&lc.ActivateRequest{
			ActivationCode: testutil.GetTestEnterpriseCode(t),
			Expires:        expiresProto,
		})
	require.NoError(t, err)
	_, err = client.Enterprise.Activate(client.Ctx(),
		&enterprise.ActivateRequest{
			Id:            "localhost",
			Secret:        "localhost",
			LicenseServer: "localhost:650",
		})
	require.NoError(t, err)

	resp, err = client.Enterprise.GetActivationCode(client.Ctx(), &enterprise.GetActivationCodeRequest{})
	require.NoError(t, err)
	require.Equal(t, resp.State, enterprise.State_EXPIRED)
	require.Equal(t, testutil.GetTestEnterpriseCode(t), resp.ActivationCode)
}

func TestGetActivationCodeNotAdmin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	testutil.DeleteAll(t)
	defer testutil.DeleteAll(t)
	aliceClient := testutil.GetAuthenticatedPachClient(t, "robot:alice")
	_, err := aliceClient.Enterprise.GetActivationCode(aliceClient.Ctx(), &enterprise.GetActivationCodeRequest{})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
}

func TestDeactivate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	testutil.DeleteAll(t)
	defer testutil.DeleteAll(t)
	client := testutil.GetPachClient(t)

	// Activate Pachyderm Enterprise and make sure the state is ACTIVE
	testutil.ActivateEnterprise(t, client)
	resp, err := client.Enterprise.GetState(client.Ctx(), &enterprise.GetStateRequest{})
	require.NoError(t, err)
	require.Equal(t, resp.State, enterprise.State_ACTIVE)

	// Deactivate cluster and make sure its state is NONE
	_, err = client.Enterprise.Deactivate(client.Ctx(),
		&enterprise.DeactivateRequest{})
	require.NoError(t, err)

	resp, err = client.Enterprise.GetState(client.Ctx(), &enterprise.GetStateRequest{})
	require.NoError(t, err)
	require.Equal(t, resp.State, enterprise.State_NONE)
}

// TestDoubleDeactivate makes sure calling Deactivate() when there is no
// enterprise token works. Fixes
// https://github.com/pachyderm/pachyderm/v2/issues/3013
func TestDoubleDeactivate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	testutil.DeleteAll(t)
	defer testutil.DeleteAll(t)
	client := testutil.GetPachClient(t)

	// Deactivate cluster and make sure its state is NONE (enterprise might be
	// active at the start of this test?)
	_, err := client.Enterprise.Deactivate(client.Ctx(),
		&enterprise.DeactivateRequest{})
	require.NoError(t, err)

	resp, err := client.Enterprise.GetState(client.Ctx(), &enterprise.GetStateRequest{})
	require.NoError(t, err)
	require.Equal(t, resp.State, enterprise.State_NONE)

	// Deactivate the cluster again to make sure deactivation with no token works
	_, err = client.Enterprise.Deactivate(client.Ctx(),
		&enterprise.DeactivateRequest{})
	require.NoError(t, err)
	resp, err = client.Enterprise.GetState(client.Ctx(),
		&enterprise.GetStateRequest{})
	require.NoError(t, err)
	require.Equal(t, enterprise.State_NONE, resp.State)
}

// TestHeartbeatDeleted tests that heartbeating fails if the pachd has been
// deleted from the license server
func TestHeartbeatDeleted(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	testutil.DeleteAll(t)
	defer testutil.DeleteAll(t)
	client := testutil.GetPachClient(t)

	// Activate Pachyderm Enterprise and make sure the state is ACTIVE
	testutil.ActivateEnterprise(t, client)

	resp, err := client.Enterprise.GetState(client.Ctx(), &enterprise.GetStateRequest{})
	require.NoError(t, err)
	require.Equal(t, enterprise.State_ACTIVE, resp.State)

	// Delete this pachd from the license server
	_, err = client.License.DeleteCluster(client.Ctx(), &lc.DeleteClusterRequest{Id: "localhost"})
	require.NoError(t, err)

	// Trigger a heartbeat and confirm the cluster is no longer active
	_, err = client.Enterprise.Heartbeat(client.Ctx(), &enterprise.HeartbeatRequest{})
	require.NoError(t, err)

	require.NoError(t, backoff.Retry(func() error {
		resp, err = client.Enterprise.GetState(client.Ctx(), &enterprise.GetStateRequest{})
		if err != nil {
			return err
		}
		if resp.State != enterprise.State_HEARTBEAT_FAILED {
			return errors.Errorf("expected enterprise state to be HEARTBEAT_FAILED but was %v", resp.State)
		}
		return nil
	}, backoff.NewTestingBackOff()))

	// Re-add the pachd and heartbeat successfully
	_, err = client.License.AddCluster(client.Ctx(),
		&lc.AddClusterRequest{
			Id:      "localhost",
			Secret:  "localhost",
			Address: "localhost:650",
		})
	require.NoError(t, err)

	_, err = client.Enterprise.Heartbeat(client.Ctx(), &enterprise.HeartbeatRequest{})
	require.NoError(t, err)

	require.NoError(t, backoff.Retry(func() error {
		resp, err = client.Enterprise.GetState(client.Ctx(), &enterprise.GetStateRequest{})
		if err != nil {
			return err
		}
		if resp.State != enterprise.State_ACTIVE {
			return errors.Errorf("expected enterprise state to be ACTIVE but was %v", resp.State)
		}
		return nil
	}, backoff.NewTestingBackOff()))

}
