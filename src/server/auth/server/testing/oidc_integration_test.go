//go:build k8s

package server

import (
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/license"
)

// TestOIDCAuthCodeFlow tests that we can configure an OIDC provider and do the
// auth code flow
func TestOIDCAuthCodeFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	require.NoError(t, tu.ConfigureOIDCProvider(t, c))

	testClient := tu.UnauthenticatedPachClient(t, c)
	loginInfo, err := testClient.GetOIDCLogin(testClient.Ctx(), &auth.GetOIDCLoginRequest{})
	require.NoError(t, err)

	tu.DoOAuthExchange(t, testClient, testClient, loginInfo.LoginURL)
	// Check that pachd recorded the response from the redirect
	authResp, err := testClient.Authenticate(testClient.Ctx(),
		&auth.AuthenticateRequest{OIDCState: loginInfo.State})
	require.NoError(t, err)
	testClient.SetAuthToken(authResp.PachToken)

	// Check that testClient authenticated as the right user
	whoAmIResp, err := testClient.WhoAmI(testClient.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, tu.User(tu.DexMockConnectorEmail), whoAmIResp.Username)
}

// TestOIDCTrustedApp tests using an ID token issued to another OIDC app to authenticate.
func TestOIDCTrustedApp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	require.NoError(t, tu.ConfigureOIDCProvider(t, c))
	testClient := tu.UnauthenticatedPachClient(t, c)

	token := tu.GetOIDCTokenForTrustedApp(t, c)

	// Use the id token from the previous OAuth flow with Pach
	authResp, err := testClient.Authenticate(testClient.Ctx(),
		&auth.AuthenticateRequest{IdToken: token})
	require.NoError(t, err)
	testClient.SetAuthToken(authResp.PachToken)

	// Check that testClient authenticated as the right user
	whoAmIResp, err := testClient.WhoAmI(testClient.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, tu.User(tu.DexMockConnectorEmail), whoAmIResp.Username)
}

// TestCannotAuthenticateWithExpiredLicense tests that we cannot login when the
// enterprise license is expired. We test this through the configured OIDC provider
// and do the auth code flow.
func TestCannotAuthenticateWithExpiredLicense(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	require.NoError(t, tu.ConfigureOIDCProvider(t, c))

	testClient := tu.UnauthenticatedPachClient(t, c)
	loginInfo, err := testClient.GetOIDCLogin(testClient.Ctx(), &auth.GetOIDCLoginRequest{})
	require.NoError(t, err)

	tu.DoOAuthExchange(t, testClient, testClient, loginInfo.LoginURL)

	// Expire Enterprise License
	adminClient := tu.AuthenticateClient(t, c, auth.RootUser)
	// set Enterprise Token value to have expired
	ts := &types.Timestamp{Seconds: time.Now().Unix() - 100}
	resp, err := adminClient.License.Activate(adminClient.Ctx(),
		&license.ActivateRequest{
			ActivationCode: tu.GetTestEnterpriseCode(t),
			Expires:        ts,
		})
	require.NoError(t, err)
	require.Equal(t, ts.Seconds, resp.GetInfo().Expires.Seconds)

	// Heartbeat forces Enterprise Service to refresh it's view of the LicenseRecord
	_, err = adminClient.Enterprise.Heartbeat(adminClient.Ctx(), &enterprise.HeartbeatRequest{})
	require.NoError(t, err)

	// Verify failure to authenticate with the OIDC Login Info
	_, err = testClient.Authenticate(testClient.Ctx(),
		&auth.AuthenticateRequest{OIDCState: loginInfo.State})
	require.YesError(t, err)
	require.Matches(t, "Pachyderm Enterprise is not active", err.Error())

	// give test user the Cluster Admin Role
	// admin grants alice cluster admin role
	_, err = adminClient.AuthAPIClient.ModifyRoleBinding(adminClient.Ctx(),
		&auth.ModifyRoleBindingRequest{
			Principal: tu.User(tu.DexMockConnectorEmail),
			Roles:     []string{auth.ClusterAdminRole},
			Resource:  &auth.Resource{Type: auth.ResourceType_CLUSTER},
		})
	require.NoError(t, err)

	// try to authenticate again
	authResp, err := testClient.Authenticate(testClient.Ctx(),
		&auth.AuthenticateRequest{OIDCState: loginInfo.State})
	require.NoError(t, err)
	testClient.SetAuthToken(authResp.PachToken)

	// Check that testClient authenticated as the right user
	whoAmIResp, err := testClient.WhoAmI(testClient.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, tu.User(tu.DexMockConnectorEmail), whoAmIResp.Username)
}
