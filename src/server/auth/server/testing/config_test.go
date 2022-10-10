//go:build unit_test

package server_test

import (
	"strconv"
	"testing"

	"github.com/gogo/protobuf/proto"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	authserver "github.com/pachyderm/pachyderm/v2/src/server/auth/server"
)

// TestSetGetConfigBasic sets an auth config and then retrieves it, to make
// sure it's stored propertly
func TestSetGetConfigBasic(t *testing.T) {
	t.Parallel()
	env := realenv.NewRealEnvWithIdentity(t, dockertestenv.NewTestDBConfig(t))
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	c := env.PachClient
	tu.ActivateAuthClient(t, c, peerPort)
	// Configure OIDC login
	require.NoError(t, tu.ConfigureOIDCProvider(t, c, true))
	adminClient := tu.AuthenticateClient(t, c, auth.RootUser)
	issuerHost := c.GetAddress().Host
	issuerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort + 8))
	redirectPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort + 7))
	// Set a configuration
	conf := &auth.OIDCConfig{
		Issuer:          "http://" + issuerHost + ":" + issuerPort + "/dex",
		ClientID:        "configtest",
		ClientSecret:    "newsecret",
		RedirectURI:     "http://" + issuerHost + ":" + redirectPort + "/authorization-code/test",
		LocalhostIssuer: false,
	}
	_, err := adminClient.SetConfiguration(adminClient.Ctx(),
		&auth.SetConfigurationRequest{Configuration: conf})
	require.NoError(t, err)

	// Read the configuration that was just written
	configResp, err := adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	require.Equal(t, true, proto.Equal(conf, configResp.Configuration))
}

// TestIssuerNotLocalhost sets an auth config with LocalhostIssuer = false
func TestIssuerNotLocalhost(t *testing.T) {
	t.Parallel()
	env := realenv.NewRealEnvWithIdentity(t, dockertestenv.NewTestDBConfig(t))
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	c := env.PachClient
	tu.ActivateAuthClient(t, c, peerPort)
	require.NoError(t, tu.ConfigureOIDCProvider(t, c, true))
	issuerHost := c.GetAddress().Host
	issuerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort + 8))
	redirectPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort + 7))

	adminClient := tu.AuthenticateClient(t, c, auth.RootUser)

	// set the issuer to locahost:1658 so we don't need to set LocalhostIssuer = true
	_, err := adminClient.SetIdentityServerConfig(adminClient.Ctx(), &identity.SetIdentityServerConfigRequest{
		Config: &identity.IdentityServerConfig{
			Issuer: "http://" + issuerHost + ":" + issuerPort + "/dex",
		},
	})
	require.NoError(t, err)

	// Set a configuration
	conf := &auth.OIDCConfig{
		Issuer:          "http://" + issuerHost + ":" + issuerPort + "/dex",
		ClientID:        "configtest",
		ClientSecret:    "newsecret",
		RedirectURI:     "http://" + issuerHost + ":" + redirectPort + "/authorization-code/test",
		LocalhostIssuer: false,
	}
	_, err = adminClient.SetConfiguration(adminClient.Ctx(),
		&auth.SetConfigurationRequest{Configuration: conf})
	require.NoError(t, err)

	// Read the configuration that was just written
	configResp, err := adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	require.Equal(t, true, proto.Equal(conf, configResp.Configuration))
}

// TestGetSetConfigAdminOnly confirms that only cluster admins can get/set the
// auth config
func TestGetSetConfigAdminOnly(t *testing.T) {
	t.Parallel()
	env := realenv.NewRealEnvWithIdentity(t, dockertestenv.NewTestDBConfig(t))
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	c := env.PachClient
	tu.ActivateAuthClient(t, c, peerPort)
	issuerHost := c.GetAddress().Host
	issuerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort + 8))
	redirectPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort + 7))

	adminClient := tu.AuthenticateClient(t, c, auth.RootUser)
	// Confirm that the auth config starts out default
	configResp, err := adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)

	require.Equal(t, true, proto.Equal(&authserver.DefaultOIDCConfig, configResp.GetConfiguration()))

	alice := tu.Robot(tu.UniqueString("alice"))
	aliceClient := tu.AuthenticateClient(t, c, alice)

	// Alice tries to set the current configuration and fails
	conf := &auth.OIDCConfig{
		Issuer:          "http://" + issuerHost + ":" + issuerPort + "/dex",
		ClientID:        "configtest",
		ClientSecret:    "newsecret",
		RedirectURI:     "http://" + issuerHost + ":" + redirectPort + "/authorization-code/test",
		LocalhostIssuer: false,
	}
	_, err = aliceClient.SetConfiguration(aliceClient.Ctx(),
		&auth.SetConfigurationRequest{
			Configuration: conf,
		})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.Matches(t, "needs permissions \\[CLUSTER_AUTH_SET_CONFIG\\] on CLUSTER", err.Error())

	// Confirm that alice didn't modify the configuration by retrieving the empty
	// config
	configResp, err = adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	require.Equal(t, true, proto.Equal(&authserver.DefaultOIDCConfig, configResp.Configuration))

	require.NoError(t, tu.ConfigureOIDCProvider(t, c, true))

	// Modify the configuration and make sure alice can't read it, but admin can
	_, err = adminClient.SetConfiguration(adminClient.Ctx(),
		&auth.SetConfigurationRequest{
			Configuration: conf,
		})
	require.NoError(t, err)

	// Confirm that alice can't read the config
	_, err = aliceClient.GetConfiguration(aliceClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.Matches(t, "needs permissions \\[CLUSTER_AUTH_GET_CONFIG\\] on CLUSTER", err.Error())

	// Confirm that admin can read the config
	configResp, err = adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	require.Equal(t, true, proto.Equal(conf, configResp.Configuration))
}

// TestConfigRestartAuth sets a config, then Deactivates+Reactivates auth, then
// calls GetConfig on an empty cluster to be sure the config was cleared
func TestConfigRestartAuth(t *testing.T) {
	t.Parallel()
	env := realenv.NewRealEnvWithIdentity(t, dockertestenv.NewTestDBConfig(t))
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	c := env.PachClient
	tu.ActivateAuthClient(t, c, peerPort)
	require.NoError(t, tu.ConfigureOIDCProvider(t, c, true))
	issuerHost := c.GetAddress().Host
	issuerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort + 8))
	redirectPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort + 7))

	adminClient := tu.AuthenticateClient(t, c, auth.RootUser)

	// Set a configuration
	conf := &auth.OIDCConfig{
		Issuer:          "http://" + issuerHost + ":" + issuerPort + "/dex",
		ClientID:        "configtest",
		ClientSecret:    "newsecret",
		RedirectURI:     "http://" + issuerHost + ":" + redirectPort + "/authorization-code/test",
		LocalhostIssuer: false,
	}
	_, err := adminClient.SetConfiguration(adminClient.Ctx(),
		&auth.SetConfigurationRequest{Configuration: conf})
	require.NoError(t, err)

	// Read the configuration that was just written
	configResp, err := adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	require.Equal(t, true, proto.Equal(conf, configResp.Configuration))

	// Deactivate auth
	_, err = adminClient.Deactivate(adminClient.Ctx(), &auth.DeactivateRequest{})
	require.NoError(t, err)

	// Wait for auth to be deactivated
	require.NoError(t, backoff.Retry(func() error {
		_, err := adminClient.WhoAmI(adminClient.Ctx(), &auth.WhoAmIRequest{})
		if err != nil && auth.IsErrNotActivated(err) {
			return nil // WhoAmI should fail when auth is deactivated
		}
		return errors.New("auth is not yet deactivated")
	}, backoff.NewTestingBackOff()))

	// Try to set and get the configuration, and confirm that the calls have been
	// deactivated
	_, err = adminClient.SetConfiguration(adminClient.Ctx(),
		&auth.SetConfigurationRequest{Configuration: conf})
	require.YesError(t, err)
	require.Matches(t, "activated", err.Error())

	_, err = adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.YesError(t, err)
	require.Matches(t, "activated", err.Error())

	// activate auth
	activateResp, err := adminClient.Activate(adminClient.Ctx(), &auth.ActivateRequest{RootToken: tu.RootToken})
	require.NoError(t, err)
	adminClient.SetAuthToken(activateResp.PachToken)

	// Wait for auth to be re-activated
	require.NoError(t, backoff.Retry(func() error {
		_, err := adminClient.WhoAmI(adminClient.Ctx(), &auth.WhoAmIRequest{})
		return err
	}, backoff.NewTestingBackOff()))

	// Try to get the configuration, and confirm that the config is now empty
	configResp, err = adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	require.Equal(t, true, proto.Equal(&authserver.DefaultOIDCConfig, configResp.Configuration))

	// Set the configuration (again)
	_, err = adminClient.SetConfiguration(adminClient.Ctx(),
		&auth.SetConfigurationRequest{Configuration: conf})
	require.NoError(t, err)

	// Get the configuration, and confirm that the config has been updated
	configResp, err = adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	require.Equal(t, true, proto.Equal(conf, configResp.Configuration))
}

// TestSetGetNilConfig tests that setting an empty config and setting a nil
// config are treated & persisted differently
func TestSetGetNilConfig(t *testing.T) {
	t.Parallel()
	env := realenv.NewRealEnvWithIdentity(t, dockertestenv.NewTestDBConfig(t))
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	c := env.PachClient
	tu.ActivateAuthClient(t, c, peerPort)
	require.NoError(t, tu.ConfigureOIDCProvider(t, c, true))
	issuerHost := c.GetAddress().Host
	issuerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort + 8))
	redirectPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort + 7))

	adminClient := tu.AuthenticateClient(t, c, auth.RootUser)

	// Set a configuration
	conf := &auth.OIDCConfig{
		Issuer:          "http://" + issuerHost + ":" + issuerPort + "/dex",
		ClientID:        "configtest",
		ClientSecret:    "newsecret",
		RedirectURI:     "http://" + issuerHost + ":" + redirectPort + "/authorization-code/test",
		LocalhostIssuer: false,
	}
	_, err := adminClient.SetConfiguration(adminClient.Ctx(),
		&auth.SetConfigurationRequest{Configuration: conf})
	require.NoError(t, err)
	// config cfg was written
	configResp, err := adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	require.Equal(t, true, proto.Equal(conf, configResp.Configuration))

	// Now, set a nil config & make sure that's retrieved correctly
	_, err = adminClient.SetConfiguration(adminClient.Ctx(),
		&auth.SetConfigurationRequest{Configuration: nil})
	require.NoError(t, err)

	// Read the configuration that was just written
	configResp, err = adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	conf = proto.Clone(&authserver.DefaultOIDCConfig).(*auth.OIDCConfig)
	require.Equal(t, true, proto.Equal(conf, configResp.Configuration))
}
