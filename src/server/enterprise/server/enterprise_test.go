package server

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/enterprise"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
)

const year = 365 * 24 * time.Hour

func TestValidateActivationCode(t *testing.T) {
	_, err := validateActivationCode(tu.GetTestEnterpriseCode(t))
	require.NoError(t, err)
}

func TestGetState(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	client := tu.GetPachClient(t)
	// Clear enterprise activation
	_, err := client.Enterprise.Deactivate(context.Background(),
		&enterprise.DeactivateRequest{})
	require.NoError(t, err)

	// Activate Pachyderm Enterprise and make sure the state is ACTIVE
	_, err = client.Enterprise.Activate(context.Background(),
		&enterprise.ActivateRequest{ActivationCode: tu.GetTestEnterpriseCode(t)})
	require.NoError(t, err)
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
		activationCode, err := unmarshalActivationCode(resp.ActivationCode)
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
	_, err = client.Enterprise.Activate(context.Background(),
		&enterprise.ActivateRequest{
			ActivationCode: tu.GetTestEnterpriseCode(t),
			Expires:        expiresProto,
		})
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
		activationCode, err := unmarshalActivationCode(resp.ActivationCode)
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
	tu.DeleteAll(t)
	client := tu.GetPachClient(t)
	// Clear enterprise activation
	_, err := client.Enterprise.Deactivate(context.Background(),
		&enterprise.DeactivateRequest{})
	require.NoError(t, err)

	// Activate Pachyderm Enterprise and make sure the state is ACTIVE
	_, err = client.Enterprise.Activate(context.Background(),
		&enterprise.ActivateRequest{ActivationCode: tu.GetTestEnterpriseCode(t)})
	require.NoError(t, err)
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
		if resp.ActivationCode != tu.GetTestEnterpriseCode(t) {
			return errors.Errorf("incorrect activation code, got: %s, expected: %s", resp.ActivationCode, tu.GetTestEnterpriseCode(t))
		}
		return nil
	}, backoff.NewTestingBackOff()))

	// Make current enterprise token expire
	expires := time.Now().Add(-30 * time.Second)
	expiresProto, err := types.TimestampProto(expires)
	require.NoError(t, err)
	_, err = client.Enterprise.Activate(context.Background(),
		&enterprise.ActivateRequest{
			ActivationCode: tu.GetTestEnterpriseCode(t),
			Expires:        expiresProto,
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
		if resp.ActivationCode != tu.GetTestEnterpriseCode(t) {
			return errors.Errorf("incorrect activation code, got: %s, expected: %s", resp.ActivationCode, tu.GetTestEnterpriseCode(t))
		}
		return nil
	}, backoff.NewTestingBackOff()))
}

func TestGetActivationCodeNotAdmin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	client := tu.GetPachClient(t)
	// Clear enterprise activation
	_, err := client.Enterprise.Deactivate(context.Background(),
		&enterprise.DeactivateRequest{})
	require.NoError(t, err)

	// Activate Pachyderm Enterprise and make sure the state is ACTIVE
	_, err = client.Enterprise.Activate(context.Background(),
		&enterprise.ActivateRequest{ActivationCode: tu.GetTestEnterpriseCode(t)})
	require.NoError(t, err)
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

	aliceClient := tu.GetAuthenticatedPachClient(t, "alice")
	_, err = aliceClient.Enterprise.GetActivationCode(aliceClient.Ctx(), &enterprise.GetActivationCodeRequest{})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
}

func TestDeactivate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	client := tu.GetPachClient(t)
	// Clear enterprise activation
	_, err := client.Enterprise.Deactivate(context.Background(),
		&enterprise.DeactivateRequest{})
	require.NoError(t, err)

	// Activate Pachyderm Enterprise and make sure the state is ACTIVE
	_, err = client.Enterprise.Activate(context.Background(),
		&enterprise.ActivateRequest{ActivationCode: tu.GetTestEnterpriseCode(t)})
	require.NoError(t, err)
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
	_, err = client.Enterprise.Deactivate(context.Background(),
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
// https://github.com/pachyderm/pachyderm/issues/3013
func TestDoubleDeactivate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	client := tu.GetPachClient(t)
	// Clear enterprise activation
	_, err := client.Enterprise.Deactivate(context.Background(),
		&enterprise.DeactivateRequest{})
	require.NoError(t, err)

	// Activate Pachyderm Enterprise and make sure the state is ACTIVE
	_, err = client.Enterprise.Activate(context.Background(),
		&enterprise.ActivateRequest{ActivationCode: tu.GetTestEnterpriseCode(t)})
	require.NoError(t, err)
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
	_, err = client.Enterprise.Deactivate(context.Background(),
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

func TestParallelism(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	c := tu.GetPachClient(t)
	// Clear enterprise activation
	_, err := c.Enterprise.Deactivate(context.Background(),
		&enterprise.DeactivateRequest{})
	require.NoError(t, err)

	dataRepo := tu.UniqueString(t.Name() + "_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	for i := 0; i < 10; i++ {
		_, err = c.PutFile(dataRepo, "master", strconv.Itoa(i), strings.NewReader("foo"))
		require.NoError(t, err)
	}
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))

	pipeline := tu.UniqueString(t.Name())
	err = c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
		},
		&pps.ParallelismSpec{
			Constant: 10, // disallowed by enterprise check
		},
		client.NewPFSInput(dataRepo, "/*"),
		"",
		false,
	)
	require.YesError(t, err)
	require.Matches(t, "parallelism", err.Error())

	// Activate Pachyderm Enterprise and make sure the state is ACTIVE
	_, err = c.Enterprise.Activate(context.Background(),
		&enterprise.ActivateRequest{ActivationCode: tu.GetTestEnterpriseCode(t)})
	require.NoError(t, err)
	require.NoError(t, backoff.Retry(func() error {
		resp, err := c.Enterprise.GetState(context.Background(),
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
		activationCode, err := unmarshalActivationCode(resp.ActivationCode)
		if err != nil {
			return err
		}
		if activationCode.Signature != "" {
			return errors.Errorf("incorrect activation code signature, expected empty string")
		}
		return nil
	}, backoff.NewTestingBackOff()))

	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
		},
		&pps.ParallelismSpec{
			Constant: 10, // disallowed by enterprise check
		},
		client.NewPFSInput(dataRepo, "/*"),
		"",
		false,
	))
	_, err = c.FlushCommitAll([]*pfs.Commit{client.NewCommit(dataRepo, "master")}, nil)
	require.NoError(t, err)
	var buf bytes.Buffer
	for i := 0; i < 10; i++ {
		buf.Reset()
		require.NoError(t, c.GetFile(pipeline, "master", strconv.Itoa(i), 0, 0, &buf))
		require.NoError(t, err)
		require.Equal(t, "foo", buf.String())
	}
}
