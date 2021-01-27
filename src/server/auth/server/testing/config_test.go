package server

import (
	"testing"

	"github.com/gogo/protobuf/proto"

	"github.com/pachyderm/pachyderm/src/auth"
	"github.com/pachyderm/pachyderm/src/identity"
	"github.com/pachyderm/pachyderm/src/internal/backoff"
	"github.com/pachyderm/pachyderm/src/internal/errors"
	"github.com/pachyderm/pachyderm/src/internal/require"
	tu "github.com/pachyderm/pachyderm/src/internal/testutil"
	authserver "github.com/pachyderm/pachyderm/src/server/auth/server"
)

// TestSetGetConfigBasic sets an auth config and then retrieves it, to make
// sure it's stored propertly
func TestSetGetConfigBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)

	tu.ConfigureOIDCProvider(t)

	adminClient := tu.GetAuthenticatedPachClient(t, auth.RootUser)

	// Set a configuration
	conf := &auth.OIDCConfig{
		Issuer:          "http://localhost:30658/",
		ClientID:        "configtest",
		ClientSecret:    "newsecret",
		RedirectURI:     "http://pachd:657/authorization-code/test",
		LocalhostIssuer: true,
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
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)

	tu.ConfigureOIDCProvider(t)

	adminClient := tu.GetAuthenticatedPachClient(t, auth.RootUser)

	// set the issuer to locahost:658 so we don't need to set LocalhostIssuer = true
	_, err := adminClient.SetIdentityServerConfig(adminClient.Ctx(), &identity.SetIdentityServerConfigRequest{
		Config: &identity.IdentityServerConfig{
			Issuer: "http://localhost:658/",
		},
	})
	require.NoError(t, err)

	// Set a configuration
	conf := &auth.OIDCConfig{
		Issuer:          "http://localhost:658/",
		ClientID:        "configtest",
		ClientSecret:    "newsecret",
		RedirectURI:     "http://pachd:657/authorization-code/test",
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
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)

	adminClient := tu.GetAuthenticatedPachClient(t, auth.RootUser)
	// Confirm that the auth config starts out default
	configResp, err := adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)

	require.Equal(t, true, proto.Equal(&authserver.DefaultOIDCConfig, configResp.GetConfiguration()))

	alice := robot(tu.UniqueString("alice"))
	anonClient := tu.GetUnauthenticatedPachClient(t)
	aliceClient := tu.GetAuthenticatedPachClient(t, alice)

	// Alice tries to set the current configuration and fails
	conf := &auth.OIDCConfig{
		Issuer:          "http://localhost:30658/",
		ClientID:        "configtest",
		ClientSecret:    "newsecret",
		RedirectURI:     "http://pachd:657/authorization-code/test",
		LocalhostIssuer: true,
	}
	_, err = aliceClient.SetConfiguration(aliceClient.Ctx(),
		&auth.SetConfigurationRequest{
			Configuration: conf,
		})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.Matches(t, "admin", err.Error())
	require.Matches(t, "SetConfiguration", err.Error())

	// Confirm that alice didn't modify the configuration by retrieving the empty
	// config
	configResp, err = adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	require.Equal(t, true, proto.Equal(&authserver.DefaultOIDCConfig, configResp.Configuration))

	tu.ConfigureOIDCProvider(t)

	// Modify the configuration and make sure anon can't read it, but alice and
	// admin can
	_, err = adminClient.SetConfiguration(adminClient.Ctx(),
		&auth.SetConfigurationRequest{
			Configuration: conf,
		})
	require.NoError(t, err)

	// Confirm that anon can't read the config
	_, err = anonClient.GetConfiguration(anonClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.YesError(t, err)
	require.Matches(t, "no authentication token", err.Error())

	// Confirm that alice and admin can read the config
	configResp, err = aliceClient.GetConfiguration(aliceClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	require.Equal(t, true, proto.Equal(conf, configResp.Configuration))

	configResp, err = adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	require.Equal(t, true, proto.Equal(conf, configResp.Configuration))
}

// TestConfigRestartAuth sets a config, then Deactivates+Reactivates auth, then
// calls GetConfig on an empty cluster to be sure the config was cleared
func TestConfigRestartAuth(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)

	tu.ConfigureOIDCProvider(t)

	adminClient := tu.GetAuthenticatedPachClient(t, auth.RootUser)

	// Set a configuration
	conf := &auth.OIDCConfig{
		Issuer:          "http://localhost:30658/",
		ClientID:        "configtest",
		ClientSecret:    "newsecret",
		RedirectURI:     "http://pachd:657/authorization-code/test",
		LocalhostIssuer: true,
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
	tu.DeleteAll(t)
}

// TestSetGetNilConfig tests that setting an empty config and setting a nil
// config are treated & persisted differently
func TestSetGetNilConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)

	tu.ConfigureOIDCProvider(t)

	adminClient := tu.GetAuthenticatedPachClient(t, auth.RootUser)

	// Set a configuration
	conf := &auth.OIDCConfig{
		Issuer:          "http://localhost:30658/",
		ClientID:        "configtest",
		ClientSecret:    "newsecret",
		RedirectURI:     "http://pachd:657/authorization-code/test",
		LocalhostIssuer: true,
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
	tu.DeleteAll(t)
}
