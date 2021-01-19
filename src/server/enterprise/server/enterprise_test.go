package server

import (
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/license"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
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

	require.NoError(t, backoff.Retry(func() error {
		resp, err := client.Enterprise.GetState(context.Background(),
			&enterprise.GetStateRequest{})
		if err != nil {
			return err
		}
		if resp.State != enterprise.State_ACTIVE {
			return errors.Errorf("expected enterprise state to be ACTIVE but was %v", resp.State)
		}
		expires, err := types.TimestampFromProto(resp.Info.Expires)
		if err != nil {
			return err
		}
		if time.Until(expires) <= year {
			return errors.Errorf("expected test token to expire >1yr in the future, but expires at %v (congratulations on making it to 2026!)", expires)
		}
		activationCode, err := license.Unmarshal(resp.ActivationCode)
		if err != nil {
			return err
		}
		if activationCode.Signature != "" {
			return errors.Errorf("incorrect activation code signature, expected empty string")
		}
		return nil
	}, backoff.NewTestingBackOff()))

	// Make current enterprise token expire
	expires := time.Now().Add(-30 * time.Second)
	expiresProto, err := types.TimestampProto(expires)
	require.NoError(t, err)

	_, err = client.License.Activate(context.Background(),
		&lc.ActivateRequest{
			ActivationCode: testutil.GetTestEnterpriseCode(t),
			Expires:        expiresProto,
		})
	require.NoError(t, err)

	_, err = client.Enterprise.Activate(context.Background(),
		&enterprise.ActivateRequest{
			Id:            "localhost",
			Secret:        "localhost",
			LicenseServer: "localhost:650",
		})
	require.NoError(t, err)

	require.NoError(t, err)
	require.NoError(t, backoff.Retry(func() error {
		resp, err := client.Enterprise.GetState(context.Background(),
			&enterprise.GetStateRequest{})
		if err != nil {
			return err
		}
		if resp.State != enterprise.State_EXPIRED {
			return errors.Errorf("expected enterprise state to be EXPIRED but was %v", resp.State)
		}
		respExpires, err := types.TimestampFromProto(resp.Info.Expires)
		if err != nil {
			return err
		}
		if expires.Unix() != respExpires.Unix() {
			return errors.Errorf("expected enterprise expiration to be %v, but was %v", expires, respExpires)
		}
		activationCode, err := license.Unmarshal(resp.ActivationCode)
		if err != nil {
			return err
		}
		if activationCode.Signature != "" {
			return errors.Errorf("incorrect activation code signature, expected empty string")
		}
		return nil
	}, backoff.NewTestingBackOff()))
}

func TestGetActivationCode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	testutil.DeleteAll(t)
	defer testutil.DeleteAll(t)
	client := testutil.GetPachClient(t)

	testutil.ActivateEnterprise(t, client)

	require.NoError(t, backoff.Retry(func() error {
		resp, err := client.Enterprise.GetActivationCode(context.Background(),
			&enterprise.GetActivationCodeRequest{})
		if err != nil {
			return err
		}
		if resp.State != enterprise.State_ACTIVE {
			return errors.Errorf("expected enterprise state to be ACTIVE but was %v", resp.State)
		}
		expires, err := types.TimestampFromProto(resp.Info.Expires)
		if err != nil {
			return err
		}
		if time.Until(expires) <= year {
			return errors.Errorf("expected test token to expire >1yr in the future, but expires at %v (congratulations on making it to 2026!)", expires)
		}
		if resp.ActivationCode != testutil.GetTestEnterpriseCode(t) {
			return errors.Errorf("incorrect activation code, got: %s, expected: %s", resp.ActivationCode, testutil.GetTestEnterpriseCode(t))
		}
		return nil
	}, backoff.NewTestingBackOff()))

	// Make current enterprise token expire
	expires := time.Now().Add(-30 * time.Second)
	expiresProto, err := types.TimestampProto(expires)
	require.NoError(t, err)
	_, err = client.License.Activate(context.Background(),
		&lc.ActivateRequest{
			ActivationCode: testutil.GetTestEnterpriseCode(t),
			Expires:        expiresProto,
		})
	require.NoError(t, err)
	_, err = client.Enterprise.Activate(context.Background(),
		&enterprise.ActivateRequest{
			Id:            "localhost",
			Secret:        "localhost",
			LicenseServer: "localhost:650",
		})
	require.NoError(t, err)

	require.NoError(t, backoff.Retry(func() error {
		resp, err := client.Enterprise.GetActivationCode(context.Background(),
			&enterprise.GetActivationCodeRequest{})
		if err != nil {
			return err
		}
		if resp.State != enterprise.State_EXPIRED {
			return errors.Errorf("expected enterprise state to be EXPIRED but was %v", resp.State)
		}
		respExpires, err := types.TimestampFromProto(resp.Info.Expires)
		if err != nil {
			return err
		}
		if expires.Unix() != respExpires.Unix() {
			return errors.Errorf("expected enterprise expiration to be %v, but was %v", expires, respExpires)
		}
		if resp.ActivationCode != testutil.GetTestEnterpriseCode(t) {
			return errors.Errorf("incorrect activation code, got: %s, expected: %s", resp.ActivationCode, testutil.GetTestEnterpriseCode(t))
		}
		return nil
	}, backoff.NewTestingBackOff()))
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

	require.NoError(t, backoff.Retry(func() error {
		resp, err := client.Enterprise.GetState(context.Background(),
			&enterprise.GetStateRequest{})
		if err != nil {
			return err
		}
		if resp.State != enterprise.State_ACTIVE {
			return errors.Errorf("expected enterprise state to be ACTIVE but was %v", resp.State)
		}
		return nil
	}, backoff.NewTestingBackOff()))

	// Deactivate cluster and make sure its state is NONE
	_, err := client.Enterprise.Deactivate(context.Background(),
		&enterprise.DeactivateRequest{})
	require.NoError(t, err)
	require.NoError(t, backoff.Retry(func() error {
		resp, err := client.Enterprise.GetState(context.Background(),
			&enterprise.GetStateRequest{})
		if err != nil {
			return err
		}
		if resp.State != enterprise.State_NONE {
			return errors.Errorf("expected enterprise state to be NONE but was %v", resp.State)
		}
		return nil
	}, backoff.NewTestingBackOff()))
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
	_, err := client.Enterprise.Deactivate(context.Background(),
		&enterprise.DeactivateRequest{})
	require.NoError(t, err)
	require.NoError(t, backoff.Retry(func() error {
		resp, err := client.Enterprise.GetState(context.Background(),
			&enterprise.GetStateRequest{})
		if err != nil {
			return err
		}
		if resp.State != enterprise.State_NONE {
			return errors.Errorf("expected enterprise state to be NONE but was %v", resp.State)
		}
		return nil
	}, backoff.NewTestingBackOff()))

	// Deactivate the cluster again to make sure deactivation with no token works
	_, err = client.Enterprise.Deactivate(context.Background(),
		&enterprise.DeactivateRequest{})
	require.NoError(t, err)
	resp, err := client.Enterprise.GetState(context.Background(),
		&enterprise.GetStateRequest{})
	require.NoError(t, err)
	require.Equal(t, enterprise.State_NONE, resp.State)
}
