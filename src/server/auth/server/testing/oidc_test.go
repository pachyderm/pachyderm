package server

import (
	"strings"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/license"
)

func newClient(t *testing.T) *client.APIClient {
	return minikubetestenv.NewPachClient(t)
}

// TestOIDCAuthCodeFlow tests that we can configure an OIDC provider and do the
// auth code flow
func TestOIDCAuthCodeFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c := newClient(t)
	tu.DeleteAll(t, c)
	tu.ConfigureOIDCProvider(t, c)
	defer tu.DeleteAll(t, c)

	testClient := tu.GetUnauthenticatedPachClient(t, c)
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
	require.Equal(t, user(tu.DexMockConnectorEmail), whoAmIResp.Username)

	tu.DeleteAll(t, c)
}

// TestOIDCTrustedApp tests using an ID token issued to another OIDC app to authenticate.
func TestOIDCTrustedApp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c := newClient(t)
	tu.DeleteAll(t, c)
	tu.ConfigureOIDCProvider(t, c)
	defer tu.DeleteAll(t, c)
	testClient := tu.GetUnauthenticatedPachClient(t, c)

	token := tu.GetOIDCTokenForTrustedApp(t, c)

	// Use the id token from the previous OAuth flow with Pach
	authResp, err := testClient.Authenticate(testClient.Ctx(),
		&auth.AuthenticateRequest{IdToken: token})
	require.NoError(t, err)
	testClient.SetAuthToken(authResp.PachToken)

	// Check that testClient authenticated as the right user
	whoAmIResp, err := testClient.WhoAmI(testClient.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, user(tu.DexMockConnectorEmail), whoAmIResp.Username)

	tu.DeleteAll(t, c)
}

// TestCannotAuthenticateWithExpiredLicense tests that we cannot login when the
// enterprise license is expired. We test this through the configured OIDC provider
// and do the auth code flow.
func TestCannotAuthenticateWithExpiredLicense(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c := newClient(t)
	tu.DeleteAll(t, c)
	tu.ConfigureOIDCProvider(t, c)
	defer tu.DeleteAll(t, c)

	testClient := tu.GetUnauthenticatedPachClient(t, c)
	loginInfo, err := testClient.GetOIDCLogin(testClient.Ctx(), &auth.GetOIDCLoginRequest{})
	require.NoError(t, err)

	tu.DoOAuthExchange(t, testClient, testClient, loginInfo.LoginURL)

	// Expire Enterprise License
	adminClient := tu.GetAuthenticatedPachClient(t, c, auth.RootUser)
	// set Enterprise Token value to have expired
	ts := &types.Timestamp{Seconds: time.Now().Unix() - 100}
	resp, err := adminClient.License.Activate(adminClient.Ctx(),
		&license.ActivateRequest{
			ActivationCode: tu.GetTestEnterpriseCode(t),
			Expires:        ts,
		})
	require.NoError(t, err)
	require.True(t, resp.GetInfo().Expires.Seconds == ts.Seconds)

	// Heartbeat forces Enterprise Service to refresh it's view of the LicenseRecord
	_, err = adminClient.Enterprise.Heartbeat(adminClient.Ctx(), &enterprise.HeartbeatRequest{})
	require.NoError(t, err)

	// Verify failure to authenticate with the OIDC Login Info
	_, err = testClient.Authenticate(testClient.Ctx(),
		&auth.AuthenticateRequest{OIDCState: loginInfo.State})
	require.YesError(t, err)
	require.True(t, strings.Contains(err.Error(), "Pachyderm Enterprise is not active"))

	// give test user the Cluster Admin Role
	// admin grants alice cluster admin role
	_, err = adminClient.AuthAPIClient.ModifyRoleBinding(adminClient.Ctx(),
		&auth.ModifyRoleBindingRequest{
			Principal: user(tu.DexMockConnectorEmail),
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
	require.Equal(t, user(tu.DexMockConnectorEmail), whoAmIResp.Username)

	tu.DeleteAll(t, c)
}
