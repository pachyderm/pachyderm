package testutil

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
)

const (
	// AdminUser is the sole cluster admin after GetAuthenticatedPachClient is called
	AdminUser = auth.RobotPrefix + "admin"

	adminToken = "pachadmin"
)

// GetAuthenticatedPachClient gets a client with an auth token for `subject`.
func GetAuthenticatedPachClient(tb testing.TB, subject string) *client.APIClient {
	c := GetPachClient(tb).WithCtx(context.Background())
	c.SetAuthToken(adminToken)

	if subject == AdminUser {
		return c
	}

	resp, err := c.GetAuthToken(c.Ctx(), &auth.GetAuthTokenRequest{Subject: subject})
	require.NoError(tb, err)
	c.SetAuthToken(resp.Token)
	return c
}

// ResetClusterState resets the set of cluster admins and activates auth and enterprise if necessary.
func ResetClusterState(tb testing.TB) {
	tb.Helper()

	c := GetPachClient(tb).WithCtx(context.Background())

	// Activate Pachyderm Enterprise (if it's not already active)
	require.NoError(tb, ActivateEnterprise(tb, c))

	// Cluster may be in one of four states:
	// 1) Auth is off (=> Activate auth)
	// 2) Auth is on, but client tokens have been invalidated (=> Deactivate +
	//    Reactivate auth)
	// 3) Auth is on and client tokens are valid, but the admin isn't "admin"
	//    (=> reset cluster admins to "admin")
	// 4) Auth is on, client tokens are valid, and the only admin is "admin" (do
	//    nothing)
	if !isAuthActive(tb) {
		// Case 1: auth is off. Activate auth
		activateAuth(tb)
		return
	}

	adminClient := GetAuthenticatedPachClient(tb, AdminUser)
	getBindingsResp, err := adminClient.GetClusterRoleBindings(adminClient.Ctx(),
		&auth.GetClusterRoleBindingsRequest{})

	// Detect case 2: auth was deactivated and reactivated during previous test.
	// This only happens during one auth test.
	if err != nil && auth.IsErrBadToken(err) {
		panic("admin token is not the expected hard-coded test value")
	}

	// Detect case 3: previous change shuffled admins. Reset list to just "admin"
	if len(getBindingsResp.Bindings) == 0 {
		panic("it should not be possible to leave a cluster with no admins")
	}

	for a, _ := range getBindingsResp.Bindings {
		if a != AdminUser {
			_, err := adminClient.ModifyClusterRoleBinding(adminClient.Ctx(),
				&auth.ModifyClusterRoleBindingRequest{
					Principal: a,
				})
			require.NoError(tb, err)
		}
	}

	// Wait for admin change to take effect
	require.NoError(tb, backoff.Retry(func() error {
		getBindingsResp, err = adminClient.GetClusterRoleBindings(adminClient.Ctx(),
			&auth.GetClusterRoleBindingsRequest{})
		_, hasAdmin := getBindingsResp.Bindings[AdminUser]
		hasExpectedAdmin := len(getBindingsResp.Bindings) == 1 && hasAdmin
		if !hasExpectedAdmin {
			return errors.Errorf("cluster admins haven't yet updated")
		}
		return nil
	}, backoff.NewTestingBackOff()))
}

// DeleteAll deletes all data in the cluster. This includes deleting all auth
// tokens, so all pachyderm clients must be recreated after calling deleteAll()
// (it should generally be called at the beginning or end of tests, before any
// clients have been created or after they're done being used).
func DeleteAll(tb testing.TB) {
	tb.Helper()
	c := GetPachClient(tb)
	if isAuthActive(tb) {
		c.SetAuthToken(adminToken)
	}
	require.NoError(tb, c.DeleteAll(), "DeleteAll()")
}

func AdminTokenHash() string {
	sum := sha256.Sum256([]byte(adminToken))
	return fmt.Sprintf("%x", sum)
}

// isAuthActive is a helper that checks if auth is currently active in the
// target cluster
func isAuthActive(tb testing.TB) bool {
	_, err := GetPachClient(tb).GetClusterRoleBindings(context.Background(),
		&auth.GetClusterRoleBindingsRequest{})
	switch {
	case auth.IsErrNotSignedIn(err):
		return true
	case auth.IsErrNotActivated(err), auth.IsErrPartiallyActivated(err):
		return false
	default:
		panic(fmt.Sprintf("could not determine if auth is activated: %v", err))
	}
	return false
}

// activateAuth activates the auth service in the test cluster
//
// Caller must hold tokenMapMut. Currently only called by getPachClient()
func activateAuth(tb testing.TB) {
	_, err := GetPachClient(tb).AuthAPIClient.Activate(context.Background(),
		&auth.ActivateRequest{RootTokenHash: AdminTokenHash()},
	)
	if err != nil && !strings.HasSuffix(err.Error(), "already activated") {
		tb.Fatalf("could not activate auth service: %v", err.Error())
	}

	// Wait for the Pachyderm Auth system to activate
	require.NoError(tb, backoff.Retry(func() error {
		if isAuthActive(tb) {
			return nil
		}
		return errors.Errorf("auth not active yet")
	}, backoff.NewTestingBackOff()))
}
