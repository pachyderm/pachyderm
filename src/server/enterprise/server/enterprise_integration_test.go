//go:build k8s

package server

import (
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

// TODO(fahad): figure out a way to simulate pausing and unpausing. This may require using a fake kube apiserver.
/*
   N.b.: for these tests to run successfully on Linux I needed to upgrade to the
   latest kubectl and run port forwards in a _loop_.  I.e.:

     while :; do  kubectl port-forward svc/pachd 30650:1653; done
*/
func TestPauseUnpause(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	client, _ := minikubetestenv.AcquireCluster(t)

	// Activate Pachyderm Enterprise and make sure the state is ACTIVE
	testutil.ActivateEnterprise(t, client)
	testutil.ActivateAuthClient(t, client)

	_, err := client.Enterprise.Pause(client.Ctx(), &enterprise.PauseRequest{})
	require.NoError(t, err)
	bo := backoff.NewExponentialBackOff()
	backoff.Retry(func() error { //nolint:errcheck
		resp, err := client.Enterprise.PauseStatus(client.Ctx(), &enterprise.PauseStatusRequest{})
		if err != nil {
			return errors.Errorf("could not get pause status %w", err)
		}
		if resp.Status == enterprise.PauseStatusResponse_PAUSED {
			return nil
		}
		return errors.Errorf("status: %v", resp.Status)
	}, bo)

	// ListRepo should return an error since the cluster is paused now
	_, err = client.ListRepo()
	require.YesError(t, err)

	_, err = client.Enterprise.Unpause(client.Ctx(), &enterprise.UnpauseRequest{})
	require.NoError(t, err)
	bo.Reset()
	backoff.Retry(func() error { //nolint:errcheck
		resp, err := client.Enterprise.PauseStatus(client.Ctx(), &enterprise.PauseStatusRequest{})
		if err != nil {
			return errors.Errorf("could not get pause status %v", err)
		}
		if resp.Status == enterprise.PauseStatusResponse_UNPAUSED {
			return nil
		}
		return errors.Errorf("status: %v", resp.Status)
	}, bo)
}

func TestPauseUnpauseNoWait(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	client, _ := minikubetestenv.AcquireCluster(t)

	// Activate Pachyderm Enterprise and make sure the state is ACTIVE
	testutil.ActivateEnterprise(t, client)
	testutil.ActivateAuthClient(t, client)

	_, err := client.Enterprise.Pause(client.Ctx(), &enterprise.PauseRequest{})
	require.NoError(t, err)

	_, err = client.Enterprise.Unpause(client.Ctx(), &enterprise.UnpauseRequest{})
	require.NoError(t, err)
	bo := backoff.NewExponentialBackOff()
	backoff.Retry(func() error { //nolint:errcheck
		resp, err := client.Enterprise.PauseStatus(client.Ctx(), &enterprise.PauseStatusRequest{})
		if err != nil {
			return errors.Errorf("could not get pause status %v", err)
		}
		if resp.Status == enterprise.PauseStatusResponse_UNPAUSED {
			return nil
		}
		return errors.Errorf("status: %v", resp.Status)
	}, bo)
	// ListRepo should not return an error since the cluster is unpaused now
	_, err = client.ListRepo()
	require.Nil(t, err)
}

func TestDoublePauseUnpause(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	client, _ := minikubetestenv.AcquireCluster(t)

	// Activate Pachyderm Enterprise and make sure the state is ACTIVE
	testutil.ActivateEnterprise(t, client)
	testutil.ActivateAuthClient(t, client)

	_, err := client.Enterprise.Pause(client.Ctx(), &enterprise.PauseRequest{})
	require.NoError(t, err)
	time.Sleep(time.Second)
	_, err = client.Enterprise.Pause(client.Ctx(), &enterprise.PauseRequest{})
	require.NoError(t, err)
	bo := backoff.NewExponentialBackOff()
	backoff.Retry(func() error { //nolint:errcheck
		resp, err := client.Enterprise.PauseStatus(client.Ctx(), &enterprise.PauseStatusRequest{})
		if err != nil {
			return errors.Errorf("could not get pause status %v", err)
		}
		if resp.Status == enterprise.PauseStatusResponse_PAUSED {
			return nil
		}
		return errors.Errorf("status: %v", resp.Status)
	}, bo)
	_, err = client.Enterprise.Unpause(client.Ctx(), &enterprise.UnpauseRequest{})
	require.NoError(t, err)
	time.Sleep(time.Second)
	_, err = client.Enterprise.Unpause(client.Ctx(), &enterprise.UnpauseRequest{})
	require.NoError(t, err)
	bo = backoff.NewExponentialBackOff()
	backoff.Retry(func() error { //nolint:errcheck
		resp, err := client.Enterprise.PauseStatus(client.Ctx(), &enterprise.PauseStatusRequest{})
		if err != nil {
			return errors.Errorf("could not get pause status %v", err)
		}
		if resp.Status == enterprise.PauseStatusResponse_UNPAUSED {
			return nil
		}
		return errors.Errorf("status: %v", resp.Status)
	}, bo)
}
