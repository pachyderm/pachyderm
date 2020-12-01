package testutil

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/identity"
	"github.com/pachyderm/pachyderm/src/client/pkg/config"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
)

const (
	// AdminUser is the sole cluster admin after GetAuthenticatedPachClient is called
	AdminUser = auth.RootUser

	adminTokenFile = "/tmp/pach-auth-test_admin-token"
)

var (
	tokenMapMut sync.Mutex // guards both tokenMap and seedClient
	tokenMap    = make(map[string]string)
	seedClient  *client.APIClient
)

// helper function that prepends auth.GitHubPrefix to 'user'--useful for validating
// responses
func gh(user string) string {
	return auth.GitHubPrefix + user
}

// helper method to generate a super admin role
func superClusterRole() *auth.ClusterRoles {
	return &auth.ClusterRoles{Roles: []auth.ClusterRole{auth.ClusterRole_SUPER}}
}

// UpdateAuthToken manually updates the auth token for a given user
func UpdateAuthToken(user string, token string) {
	tokenMapMut.Lock()
	defer tokenMapMut.Unlock()

	tokenMap[user] = token
}

// isAuthActive is a helper that checks if auth is currently active in the
// target cluster
//
// Caller must hold tokenMapMut. Currently only called by getPachClient(),
// activateAuth (which is only called by getPachClient()) and deleteAll()
func isAuthActive(tb testing.TB, checkConfig bool) bool {
	_, err := seedClient.GetClusterRoleBindings(context.Background(),
		&auth.GetClusterRoleBindingsRequest{})
	switch {
	case auth.IsErrNotSignedIn(err):
		adminClient := getPachClientInternal(tb, AdminUser)
		if checkConfig {
			if err := backoff.Retry(func() error {
				resp, err := adminClient.GetConfiguration(adminClient.Ctx(), &auth.GetConfigurationRequest{})
				if err != nil {
					return errors.Wrapf(err, "could not get config")
				}
				cfg := resp.GetConfiguration()
				if cfg.SAMLServiceOptions != nil {
					return errors.Errorf("SAML config in fresh cluster: %+v", cfg)
				}
				if len(cfg.IDProviders) != 1 || cfg.IDProviders[0].SAML != nil || cfg.IDProviders[0].GitHub == nil || cfg.IDProviders[0].Name != "GitHub" {
					return errors.Errorf("problem with ID providers in config in fresh cluster: %+v", cfg)
				}
				return nil
			}, backoff.NewTestingBackOff()); err != nil {
				panic(err)
			}
		}
		return true
	case auth.IsErrNotActivated(err), auth.IsErrPartiallyActivated(err):
		return false
	default:
		panic(fmt.Sprintf("could not determine if auth is activated: %v", err))
	}
	return false
}

// getPachClientInternal is a helper function called by getPachClient. It
// returns (or creates) a pach client authenticated as 'subject', but doesn't
// do any any checks to confirm that auth is activated and the cluster is
// configured correctly (those are done by getPachClient). If subject has no
// prefix, they are assumed to be a GitHub user.
//
// Caller must hold tokenMapMut. Currently only called by getPachClient()
func getPachClientInternal(tb testing.TB, subject string) *client.APIClient {
	// copy seed, so caller can safely modify result
	resultClient := seedClient.WithCtx(context.Background())
	if subject == "" {
		return resultClient // anonymous client
	}
	if token, ok := tokenMap[subject]; ok {
		resultClient.SetAuthToken(token)
		return resultClient
	}
	if subject == AdminUser {
		bytes, err := ioutil.ReadFile(adminTokenFile)
		if err == nil {
			tb.Logf("couldn't find admin token in cache, reading from %q", adminTokenFile)
			resultClient.SetAuthToken(string(bytes))
			return resultClient
		}
		tb.Fatalf("couldn't get admin client from cache or %q, no way to reset "+
			"cluster. Please deactivate auth or redeploy Pachyderm", adminTokenFile)
	}
	if !strings.Contains(subject, ":") {
		subject = gh(subject)
	}
	colonIdx := strings.Index(subject, ":")
	prefix := subject[:colonIdx+1]
	switch prefix {
	case auth.RobotPrefix:
		adminClient := getPachClientInternal(tb, AdminUser)
		resp, err := adminClient.GetAuthToken(adminClient.Ctx(), &auth.GetAuthTokenRequest{
			Subject: subject,
		})
		require.NoError(tb, err)
		tokenMap[subject] = resp.Token
	case auth.GitHubPrefix:
		resp, err := seedClient.Authenticate(context.Background(),
			&auth.AuthenticateRequest{
				// When Pachyderm is deployed locally, GitHubToken automatically
				// authenicates the caller as a GitHub user whose name is equal to the
				// provided token
				GitHubToken: strings.TrimPrefix(subject, auth.GitHubPrefix),
			})
		require.NoError(tb, err)
		tokenMap[subject] = resp.PachToken
	default:
		tb.Fatalf("can't give you a client of type %s", prefix)
	}
	resultClient.SetAuthToken(tokenMap[subject])
	return resultClient
}

// activateAuth activates the auth service in the test cluster
//
// Caller must hold tokenMapMut. Currently only called by getPachClient()
func activateAuth(tb testing.TB) {
	resp, err := seedClient.AuthAPIClient.Activate(context.Background(),
		&auth.ActivateRequest{Subject: AdminUser},
	)
	if err != nil && !strings.HasSuffix(err.Error(), "already activated") {
		tb.Fatalf("could not activate auth service: %v", err.Error())
	}
	tokenMap[AdminUser] = resp.PachToken
	ioutil.WriteFile(adminTokenFile, []byte(resp.PachToken), 0644)
	config.WritePachTokenToConfig(resp.PachToken)

	// Wait for the Pachyderm Auth system to activate
	require.NoError(tb, backoff.Retry(func() error {
		if isAuthActive(tb, true) {
			return nil
		}
		return errors.Errorf("auth not active yet")
	}, backoff.NewTestingBackOff()))
}

// initSeedClient is a helper function called by getPachClient that initializes
// the 'seedClient' global variable.
//
// Caller must hold tokenMapMut. Currently only called by getPachClient() and
// deleteAll()
func initSeedClient(tb testing.TB) {
	tb.Helper()
	var err error
	if _, ok := os.LookupEnv("PACHD_PORT_650_TCP_ADDR"); ok {
		seedClient, err = client.NewInCluster()
	} else {
		seedClient, err = client.NewForTest()
	}
	require.NoError(tb, err)
	// discard any credentials from the user's machine (seedClient is
	// anonymous)
	seedClient = seedClient.WithCtx(context.Background())
	seedClient.SetAuthToken("")
}

// getPachClientConfigAgnostic does not check that the auth config is in the default state.
// i.e. it can be used to retrieve the pach client even after the auth config has been manipulated
func getPachClientConfigAgnostic(tb testing.TB, subject string) *client.APIClient {
	return getPachClientP(tb, subject, false)
}

// GetAuthenticatedPachClient explicitly checks that the auth config is set to the default,
// and will fail otherwise.
func GetAuthenticatedPachClient(tb testing.TB, subject string) *client.APIClient {
	return getPachClientP(tb, subject, true)
}

// getPachClient creates a seed client with a grpc connection to a pachyderm
// cluster, and then enable the auth service in that cluster
func getPachClientP(tb testing.TB, subject string, checkConfig bool) *client.APIClient {
	tb.Helper()
	tokenMapMut.Lock()
	defer tokenMapMut.Unlock()

	// Check if seed client exists -- if not, create it
	if seedClient == nil {
		initSeedClient(tb)
	}

	// Activate Pachyderm Enterprise (if it's not already active)
	require.NoError(tb, ActivateEnterprise(tb, seedClient))

	// Cluster may be in one of four states:
	// 1) Auth is off (=> Activate auth)
	// 2) Auth is on, but client tokens have been invalidated (=> Deactivate +
	//    Reactivate auth)
	// 3) Auth is on and client tokens are valid, but the admin isn't "admin"
	//    (=> reset cluster admins to "admin")
	// 4) Auth is on, client tokens are valid, and the only admin is "admin" (do
	//    nothing)
	if !isAuthActive(tb, checkConfig) {
		// Case 1: auth is off. Activate auth & return a new client
		tokenMap = make(map[string]string)
		activateAuth(tb)
		return getPachClientInternal(tb, subject)
	}

	adminClient := getPachClientInternal(tb, AdminUser)
	getBindingsResp, err := adminClient.GetClusterRoleBindings(adminClient.Ctx(),
		&auth.GetClusterRoleBindingsRequest{})

	// Detect case 2: auth was deactivated during previous test. De/Reactivate
	// TODO it may may sense to do this between every test, though it would mean
	// that we can't run tests in parallel. Currently, only the first test that
	// needs to reset auth will run this block
	if err != nil && auth.IsErrBadToken(err) {
		// Don't know which tokens are valid, so clear tokenMap
		tokenMap = make(map[string]string)
		adminClient := getPachClientInternal(tb, AdminUser)

		// "admin" may no longer be an admin, so get a list of admins, authorize as
		// the first admin, and deactivate auth
		getBindingsResp, err := adminClient.GetClusterRoleBindings(adminClient.Ctx(),
			&auth.GetClusterRoleBindingsRequest{})
		require.NoError(tb, err)
		if len(getBindingsResp.Bindings) == 0 {
			panic("it should not be possible to leave a cluster with no admins")
		}
		var currentAdmin string
		for a := range getBindingsResp.Bindings {
			currentAdmin = a
			break
		}
		adminClient = getPachClientInternal(tb, currentAdmin)
		_, err = adminClient.Deactivate(adminClient.Ctx(), &auth.DeactivateRequest{})
		require.NoError(tb, err)

		// Just deactivated. *All* tokens are now invalid, so clear the map again
		tokenMap = make(map[string]string)
		activateAuth(tb)
		return getPachClientInternal(tb, subject)
	}

	// Detect case 3: previous change shuffled admins. Reset list to just "admin"
	if len(getBindingsResp.Bindings) == 0 {
		panic("it should not be possible to leave a cluster with no admins")
	}
	_, hasAdmin := getBindingsResp.Bindings[AdminUser]
	hasExpectedAdmin := len(getBindingsResp.Bindings) == 1 && hasAdmin
	if !hasExpectedAdmin {
		var curAdminClient *client.APIClient
		for a, roles := range getBindingsResp.Bindings {
			var isSuper bool
			for _, r := range roles.Roles {
				if r == auth.ClusterRole_SUPER {
					isSuper = true
					break
				}
			}
			if !isSuper {
				continue
			}

			if strings.HasPrefix(a, auth.GitHubPrefix) {
				curAdminClient = getPachClientInternal(tb, a) // use first GH admin
				break
			}
			if a == AdminUser {
				curAdminClient = getPachClientInternal(tb, a) // use admin robot
				break
			}
		}

		if curAdminClient == nil {
			tb.Fatal("cluster has no GitHub admins; no way for auth test to grant itself admin access")
			// staticcheck does not realize tb.Fatal will not return (because it is an
			// interface), see https://github.com/dominikh/go-tools/issues/635
			return nil
		}

		_, err = curAdminClient.ModifyClusterRoleBinding(curAdminClient.Ctx(), &auth.ModifyClusterRoleBindingRequest{
			Principal: AdminUser,
			Roles:     superClusterRole(),
		})
		require.NoError(tb, err)
		for a := range getBindingsResp.Bindings {
			if a != AdminUser {
				_, err = curAdminClient.ModifyClusterRoleBinding(curAdminClient.Ctx(), &auth.ModifyClusterRoleBindingRequest{Principal: a})
				require.NoError(tb, err)
			}
		}
		if curAdminClient == nil {
			tb.Fatal("cluster has no GitHub admins; no way for auth test to grant itself admin access")
			// staticcheck does not realize tb.Fatal will not return (because it is an
			// interface), see https://github.com/dominikh/go-tools/issues/635
			return nil
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

	return getPachClientInternal(tb, subject)
}

// DeleteAll deletes all data in the cluster. This includes deleting all auth
// tokens, so all pachyderm clients must be recreated after calling deleteAll()
// (it should generally be called at the beginning or end of tests, before any
// clients have been created or after they're done being used).
func DeleteAll(tb testing.TB) {
	tb.Helper()
	var anonClient *client.APIClient
	var useAdminClient bool
	func() {
		tokenMapMut.Lock() // May initialize the seed client
		defer tokenMapMut.Unlock()

		if seedClient == nil {
			initSeedClient(tb)
		}
		anonClient = seedClient.WithCtx(context.Background())

		// config might be nil if this is
		// called after a test that changed
		// the config
		useAdminClient = isAuthActive(tb, false)
	}() // release tokenMapMut before getPachClient
	if useAdminClient {
		adminClient := getPachClientConfigAgnostic(tb, AdminUser)
		_, err := adminClient.IdentityAPIClient.DeleteAll(adminClient.Ctx(), &identity.DeleteAllRequest{})
		require.NoError(tb, err, "initial DeleteAll()")
		require.NoError(tb, adminClient.DeleteAll(), "initial DeleteAll()")
	} else {
		require.NoError(tb, anonClient.DeleteAll())
	}
}
