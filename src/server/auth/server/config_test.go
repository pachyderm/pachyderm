package server

import (
	"encoding/xml"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"

	"github.com/crewjam/saml" // used to format saml IdP in config
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
)

// requireConfigsEqual compares 'expected' and 'actual' using 'proto.Equal', but
// also formats any XML they contain to remove e.g. whitespace discrepancies
// that we want to ignore
func requireConfigsEqual(t testing.TB, expected, actual *auth.AuthConfig) {
	e, a := proto.Clone(expected).(*auth.AuthConfig), proto.Clone(actual).(*auth.AuthConfig)
	// Format IdP metadata (which is serialized XML) to fix e.g. whitespace
	formatIDP := func(c *auth.IDProvider) {
		if c.SAML == nil || c.SAML.MetadataXML == nil {
			return
		}
		var d saml.EntityDescriptor
		require.NoError(t, xml.Unmarshal(c.SAML.MetadataXML, &d))
		c.SAML.MetadataXML = tu.MustMarshalXML(t, &d)
	}
	for _, idp := range e.IDProviders {
		formatIDP(idp)
	}
	for _, idp := range a.IDProviders {
		formatIDP(idp)
	}
	require.True(t, proto.Equal(e, a),
		"expected: %v\n           actual: %v", e, a)
}

// TestSetGetConfigBasic sets an auth config and then retrieves it, to make
// sure it's stored propertly
func TestSetGetConfigBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)

	adminClient := getPachClient(t, admin)

	// Set a configuration
	conf := &auth.AuthConfig{
		IDProviders: []*auth.IDProvider{{
			Name:        "idp",
			Description: "fake IdP for testing",
			SAML: &auth.IDProvider_SAMLOptions{
				MetadataXML: SimpleSAMLIDPMetadata(t),
			},
		}},
		SAMLServiceOptions: &auth.AuthConfig_SAMLServiceOptions{
			ACSURL: "http://acs", MetadataURL: "http://metadata",
		},
	}
	_, err := adminClient.SetConfiguration(adminClient.Ctx(),
		&auth.SetConfigurationRequest{Configuration: conf})
	require.NoError(t, err)

	// Read the configuration that was just written
	configResp, err := adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	conf.LiveConfigVersion = 2 // increment version ("default" config has v=1)
	requireConfigsEqual(t, conf, configResp.Configuration)
	deleteAll(t)
}

// TestGetSetConfigAdminOnly confirms that only cluster admins can get/set the
// auth config
func TestGetSetConfigAdminOnly(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)

	adminClient := getPachClient(t, admin)
	// Confirm that the auth config starts out default
	configResp, err := adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	requireConfigsEqual(t, &defaultAuthConfig, configResp.GetConfiguration())

	alice := tu.UniqueString("alice")
	anonClient := getPachClient(t, "")
	aliceClient := getPachClient(t, alice)

	// Alice tries to set the current configuration and fails
	conf := &auth.AuthConfig{
		IDProviders: []*auth.IDProvider{{
			Name:        "idp",
			Description: "fake IdP for testing",
			SAML: &auth.IDProvider_SAMLOptions{
				MetadataXML: SimpleSAMLIDPMetadata(t),
			},
		}},
		SAMLServiceOptions: &auth.AuthConfig_SAMLServiceOptions{
			ACSURL: "http://acs", MetadataURL: "http://metadata",
		},
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
	requireConfigsEqual(t, &defaultAuthConfig, configResp.Configuration)

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
	conf.LiveConfigVersion = 2 // increment version ("default" config has v=1)
	requireConfigsEqual(t, conf, configResp.Configuration)

	configResp, err = adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	requireConfigsEqual(t, conf, configResp.Configuration)
	deleteAll(t)
}

// TestRMWConfigConflict does two conflicting R+M+W operation on a config
// and confirms that one operation fails
func TestRMWConfigConflict(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)

	adminClient := getPachClient(t, admin)

	// Set a configuration
	conf := &auth.AuthConfig{
		IDProviders: []*auth.IDProvider{{
			Name:        "idp",
			Description: "fake IdP for testing",
			SAML: &auth.IDProvider_SAMLOptions{
				MetadataXML: SimpleSAMLIDPMetadata(t),
			},
		}},
		SAMLServiceOptions: &auth.AuthConfig_SAMLServiceOptions{
			ACSURL: "http://acs", MetadataURL: "http://metadata",
		},
	}

	// Set an initial configuration
	_, err := adminClient.SetConfiguration(adminClient.Ctx(),
		&auth.SetConfigurationRequest{Configuration: conf})
	require.NoError(t, err)

	// Read the configuration that was just written
	configResp, err := adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	conf.LiveConfigVersion = 2 // increment version ("default" config has v=1)
	requireConfigsEqual(t, conf, configResp.Configuration)

	// modify the config twice
	mod2 := proto.Clone(configResp.Configuration).(*auth.AuthConfig)
	mod2.IDProviders = []*auth.IDProvider{{
		Name: "idp_2", Description: "fake IDP for testing, #2",
		SAML: &auth.IDProvider_SAMLOptions{
			MetadataXML: SimpleSAMLIDPMetadata(t),
		}}}
	mod3 := proto.Clone(configResp.Configuration).(*auth.AuthConfig)
	mod3.IDProviders = []*auth.IDProvider{{
		Name: "idp_3", Description: "fake IDP for testing, #3",
		SAML: &auth.IDProvider_SAMLOptions{
			MetadataXML: SimpleSAMLIDPMetadata(t),
		}}}

	// Apply both changes -- the second should fail
	_, err = adminClient.SetConfiguration(adminClient.Ctx(), &auth.SetConfigurationRequest{
		Configuration: mod2,
	})
	require.NoError(t, err)
	_, err = adminClient.SetConfiguration(adminClient.Ctx(), &auth.SetConfigurationRequest{
		Configuration: mod3,
	})
	require.YesError(t, err)
	require.Matches(t, "config version", err.Error())

	// Read the configuration and make sure that only #2 was applied
	configResp, err = adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	mod2.LiveConfigVersion = 3 // increment version
	requireConfigsEqual(t, mod2, configResp.Configuration)
	deleteAll(t)
}

// TestSetGetEmptyConfig is like TestSetNilConfig, but it passes a non-nil but
// empty config. Currently the default config is an empty config, but this test
// will ensure that setting an empty config has the right behavior even if the
// default config is changed later.
func TestSetGetEmptyConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)

	adminClient := getPachClient(t, admin)

	// Set a configuration & read it back
	conf := &auth.AuthConfig{
		IDProviders: []*auth.IDProvider{{
			Name:        "idp",
			Description: "fake IdP for testing",
			SAML: &auth.IDProvider_SAMLOptions{
				MetadataXML: SimpleSAMLIDPMetadata(t)},
		}},
		SAMLServiceOptions: &auth.AuthConfig_SAMLServiceOptions{
			ACSURL: "http://acs", MetadataURL: "http://metadata",
		},
	}
	_, err := adminClient.SetConfiguration(adminClient.Ctx(),
		&auth.SetConfigurationRequest{Configuration: conf})
	require.NoError(t, err)
	// Read the config back
	configResp, err := adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	conf.LiveConfigVersion = 2 // increment version ("default" config has v=1)
	requireConfigsEqual(t, conf, configResp.Configuration)

	// Set the empty config
	_, err = adminClient.SetConfiguration(adminClient.Ctx(),
		&auth.SetConfigurationRequest{
			Configuration: &auth.AuthConfig{},
		})
	require.NoError(t, err)

	// Get the empty config
	expected := auth.AuthConfig{}
	expected.LiveConfigVersion = 3 // increment version
	configResp, err = adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	requireConfigsEqual(t, &expected, configResp.Configuration)

	// TODO Make sure SAML has been deactivated
	deleteAll(t)
}

// TestConfigRestartAuth sets a config, then Deactivates+Reactivates auth, then
// calls GetConfig on an empty cluster to be sure the config was cleared
func TestConfigRestartAuth(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)

	adminClient := getPachClient(t, admin)

	// Set a configuration
	conf := &auth.AuthConfig{
		IDProviders: []*auth.IDProvider{{
			Name:        "idp",
			Description: "fake IdP for testing",
			SAML: &auth.IDProvider_SAMLOptions{
				MetadataXML: SimpleSAMLIDPMetadata(t)},
		}},
		SAMLServiceOptions: &auth.AuthConfig_SAMLServiceOptions{
			ACSURL: "http://acs", MetadataURL: "http://metadata",
		},
	}
	_, err := adminClient.SetConfiguration(adminClient.Ctx(),
		&auth.SetConfigurationRequest{Configuration: conf})
	require.NoError(t, err)

	// Read the configuration that was just written
	configResp, err := adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	conf.LiveConfigVersion = 2 // increment version ("default" config has v=1)
	requireConfigsEqual(t, conf, configResp.Configuration)

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
	activateResp, err := adminClient.Activate(adminClient.Ctx(), &auth.ActivateRequest{
		Subject: admin,
	})
	require.NoError(t, err)
	adminClient.SetAuthToken(activateResp.PachToken)
	tokenMap[admin] = activateResp.PachToken

	// Wait for auth to be re-activated
	require.NoError(t, backoff.Retry(func() error {
		_, err := adminClient.WhoAmI(adminClient.Ctx(), &auth.WhoAmIRequest{})
		return err
	}, backoff.NewTestingBackOff()))

	// Try to get the configuration, and confirm that the config is now empty
	configResp, err = adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	requireConfigsEqual(t, &defaultAuthConfig, configResp.Configuration)

	// Set the configuration (again)
	conf.LiveConfigVersion = 0 // clear version, so we issue a blind write
	_, err = adminClient.SetConfiguration(adminClient.Ctx(),
		&auth.SetConfigurationRequest{Configuration: conf})
	require.NoError(t, err)

	// Get the configuration, and confirm that the config has been updated
	configResp, err = adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	conf.LiveConfigVersion = 2 // increment version ("default" config has v=1)
	requireConfigsEqual(t, conf, configResp.Configuration)
	deleteAll(t)
}

// TestValidateConfigErrNoName tests that SetConfig rejects configs with unnamed
// ID providers
func TestValidateConfigErrNoName(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	adminClient := getPachClient(t, admin)

	conf := &auth.AuthConfig{
		IDProviders: []*auth.IDProvider{&auth.IDProvider{
			Description: "fake IdP for testing",
			SAML: &auth.IDProvider_SAMLOptions{
				MetadataXML: SimpleSAMLIDPMetadata(t)},
		}},
		SAMLServiceOptions: &auth.AuthConfig_SAMLServiceOptions{
			ACSURL: "http://acs", MetadataURL: "http://metadata",
		},
	}
	_, err := adminClient.SetConfiguration(adminClient.Ctx(),
		&auth.SetConfigurationRequest{Configuration: conf})
	require.YesError(t, err)
	require.Matches(t, "must have a name", err.Error())

	// Make sure config change wasn't applied
	configResp, err := adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	requireConfigsEqual(t, &defaultAuthConfig, configResp.Configuration)
	deleteAll(t)
}

// TestValidateConfigErrReservedName tests that SetConfig rejects configs that
// try to use a reserved name
func TestValidateConfigErrReservedName(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	adminClient := getPachClient(t, admin)

	for _, name := range []string{"robot", "pipeline"} {
		conf := &auth.AuthConfig{
			IDProviders: []*auth.IDProvider{&auth.IDProvider{
				Name:        name,
				Description: "fake IdP for testing",
				SAML: &auth.IDProvider_SAMLOptions{
					MetadataXML: SimpleSAMLIDPMetadata(t)},
			}},
			SAMLServiceOptions: &auth.AuthConfig_SAMLServiceOptions{
				ACSURL: "http://acs", MetadataURL: "http://metadata",
			},
		}
		_, err := adminClient.SetConfiguration(adminClient.Ctx(),
			&auth.SetConfigurationRequest{Configuration: conf})
		require.YesError(t, err)
		require.Matches(t, "reserved", err.Error())
	}

	// Make sure config change wasn't applied
	configResp, err := adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	requireConfigsEqual(t, &defaultAuthConfig, configResp.Configuration)
	deleteAll(t)
}

// TestValidateConfigErrNoType tests that SetConfig rejects configs that
// have an IDProvider (IDP) with no type
func TestValidateConfigErrNoType(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	adminClient := getPachClient(t, admin)

	conf := &auth.AuthConfig{
		IDProviders: []*auth.IDProvider{&auth.IDProvider{
			Name:        "idp",
			Description: "fake IdP for testing",
		}},
		SAMLServiceOptions: &auth.AuthConfig_SAMLServiceOptions{
			ACSURL: "http://acs", MetadataURL: "http://metadata",
		},
	}
	_, err := adminClient.SetConfiguration(adminClient.Ctx(),
		&auth.SetConfigurationRequest{Configuration: conf})
	require.YesError(t, err)
	require.Matches(t, "type", err.Error())

	// Make sure config change wasn't applied
	configResp, err := adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	requireConfigsEqual(t, &defaultAuthConfig, configResp.Configuration)
	deleteAll(t)
}

// TestValidateConfigErrInvalidIDPMetadata tests that SetConfig rejects configs
// that have a SAML IDProvider with invalid XML in their IDPMetadata
func TestValidateConfigErrInvalidIDPMetadata(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	adminClient := getPachClient(t, admin)

	conf := &auth.AuthConfig{
		IDProviders: []*auth.IDProvider{&auth.IDProvider{
			Name:        "idp",
			Description: "fake IdP for testing",
			SAML: &auth.IDProvider_SAMLOptions{
				MetadataXML: []byte("invalid XML"),
			},
		}},
		SAMLServiceOptions: &auth.AuthConfig_SAMLServiceOptions{
			ACSURL: "http://acs", MetadataURL: "http://metadata",
		},
	}
	_, err := adminClient.SetConfiguration(adminClient.Ctx(),
		&auth.SetConfigurationRequest{Configuration: conf})
	require.YesError(t, err)
	require.Matches(t, "could not unmarshal", err.Error())

	// Make sure config change wasn't applied
	configResp, err := adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	requireConfigsEqual(t, &defaultAuthConfig, configResp.Configuration)
	deleteAll(t)
}

// TestValidateConfigErrInvalidIDPMetadata tests that SetConfig rejects configs
// that have a SAML IDProvider with an invalid metadata URL
func TestValidateConfigErrInvalidMetadataURL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	adminClient := getPachClient(t, admin)

	conf := &auth.AuthConfig{
		IDProviders: []*auth.IDProvider{&auth.IDProvider{
			Name:        "idp",
			Description: "fake IdP for testing",
			SAML: &auth.IDProvider_SAMLOptions{
				MetadataURL: "invalid URL",
			},
		}},
		SAMLServiceOptions: &auth.AuthConfig_SAMLServiceOptions{
			ACSURL: "http://acs", MetadataURL: "http://metadata",
		},
	}
	_, err := adminClient.SetConfiguration(adminClient.Ctx(),
		&auth.SetConfigurationRequest{Configuration: conf})
	require.YesError(t, err)
	require.Matches(t, "URL", err.Error())

	// Make sure config change wasn't applied
	configResp, err := adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	requireConfigsEqual(t, &defaultAuthConfig, configResp.Configuration)
	deleteAll(t)
}

// TestValidateConfigErrRedundantIDPMetadata tests that SetConfig rejects
// configs with a SAML IDProvider that has both explicit metadata and a metadata
// URL
func TestValidateConfigErrRedundantIDPMetadata(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	adminClient := getPachClient(t, admin)
	_, err := adminClient.WhoAmI(adminClient.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)

	var (
		pachdSAMLAddress = tu.GetACSAddress(t, adminClient.GetAddress())
		pachdACSURL      = fmt.Sprintf("http://%s/saml/acs", pachdSAMLAddress)
		pachdMetadataURL = fmt.Sprintf("http://%s/saml/metadata", pachdSAMLAddress)
	)
	testIDP := tu.NewTestIDP(t, pachdACSURL, pachdMetadataURL)
	conf := &auth.AuthConfig{
		IDProviders: []*auth.IDProvider{
			{
				Name:        "idp_1",
				Description: "fake IdP for testing",
				SAML: &auth.IDProvider_SAMLOptions{
					MetadataXML: testIDP.Metadata(),
				},
			},
			{
				Name:        "idp_2",
				Description: "another fake IdP for testing",
				SAML: &auth.IDProvider_SAMLOptions{
					MetadataXML: testIDP.Metadata(),
				},
			},
		},
		SAMLServiceOptions: &auth.AuthConfig_SAMLServiceOptions{
			ACSURL: "http://acs", MetadataURL: "http://metadata",
		},
	}
	_, err = adminClient.SetConfiguration(adminClient.Ctx(),
		&auth.SetConfigurationRequest{Configuration: conf})
	require.YesError(t, err)
	require.Matches(t, "only one is allowed", err.Error())

	// Make sure config change wasn't applied
	configResp, err := adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	requireConfigsEqual(t, &defaultAuthConfig, configResp.Configuration)
	deleteAll(t)
}

// TestConfigDeadlock tests that Pachyderm's SAML endpoint releases Pachyderm's
// config mutex (a bug fix)
func TestConfigDeadlock(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	adminClient := getPachClient(t, admin)

	var (
		pachdSAMLAddress = tu.GetACSAddress(t, adminClient.GetAddress())
		pachdACSURL      = fmt.Sprintf("http://%s/saml/acs", pachdSAMLAddress)
		pachdMetadataURL = fmt.Sprintf("http://%s/saml/metadata", pachdSAMLAddress)
	)
	testIDP := tu.NewTestIDP(t, pachdACSURL, pachdMetadataURL)

	// Get current configuration, and then set a configuration (this may need to
	// be retried if e.g. scraping metadata fails)
	configResp, err := adminClient.GetConfiguration(adminClient.Ctx(), &auth.GetConfigurationRequest{})
	require.NoError(t, err)
	_, err = adminClient.SetConfiguration(adminClient.Ctx(), &auth.SetConfigurationRequest{
		Configuration: &auth.AuthConfig{
			LiveConfigVersion: configResp.Configuration.LiveConfigVersion,
			IDProviders: []*auth.IDProvider{
				{
					Name:        "idp_1",
					Description: "fake IdP for testing",
					SAML: &auth.IDProvider_SAMLOptions{
						MetadataXML:    testIDP.Metadata(),
						GroupAttribute: "memberOf",
					},
				},
			},
			SAMLServiceOptions: &auth.AuthConfig_SAMLServiceOptions{
				ACSURL:      pachdACSURL,
				MetadataURL: pachdMetadataURL,
			},
		},
	})
	require.NoError(t, err)

	// Send a SAMLResponse to Pachyderm
	alice, group := tu.UniqueString("alice"), tu.UniqueString("group")
	aliceClient := getPachClientConfigAgnostic(t, "") // empty string b/c want anon client
	tu.AuthenticateWithSAMLResponse(t, aliceClient, testIDP.NewSAMLResponse(alice, group))
	who, err := aliceClient.WhoAmI(aliceClient.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, "idp_1:"+alice, who.Username)

	// get/set Pachyderm's configuration again, to make sure the config mutex has
	// been released
	require.NoErrorWithinT(t, 30*time.Second, func() error {
		configResp, err = adminClient.GetConfiguration(adminClient.Ctx(), &auth.GetConfigurationRequest{})
		if err != nil {
			return err
		}
		_, err = adminClient.SetConfiguration(adminClient.Ctx(), &auth.SetConfigurationRequest{
			Configuration: &auth.AuthConfig{
				LiveConfigVersion: configResp.Configuration.LiveConfigVersion,
				IDProviders: []*auth.IDProvider{
					{
						Name:        "idp_2",
						Description: "fake IdP for testing",
						SAML: &auth.IDProvider_SAMLOptions{
							MetadataXML:    testIDP.Metadata(),
							GroupAttribute: "memberOf",
						},
					},
				},
				SAMLServiceOptions: &auth.AuthConfig_SAMLServiceOptions{
					ACSURL:      pachdACSURL,
					MetadataURL: pachdMetadataURL,
				},
			},
		})
		return err
	})
	deleteAll(t)
}

// TestSetGetNilConfig tests that setting an empty config and setting a nil
// config are treated & persisted differently
func TestSetGetNilConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)

	adminClient := getPachClient(t, admin)

	// Set a configuration
	conf := &auth.AuthConfig{
		IDProviders: []*auth.IDProvider{{
			Name:        "idp",
			Description: "fake IdP for testing",
			SAML: &auth.IDProvider_SAMLOptions{
				MetadataXML: SimpleSAMLIDPMetadata(t),
			},
		}},
		SAMLServiceOptions: &auth.AuthConfig_SAMLServiceOptions{
			ACSURL: "http://acs", MetadataURL: "http://metadata",
		},
	}
	_, err := adminClient.SetConfiguration(adminClient.Ctx(),
		&auth.SetConfigurationRequest{Configuration: conf})
	require.NoError(t, err)
	// config cfg was written
	configResp, err := adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	conf.LiveConfigVersion = 2 // increment version ("default" config has v=1)
	requireConfigsEqual(t, conf, configResp.Configuration)

	// Now, set a nil config & make sure that's retrieved correctly
	_, err = adminClient.SetConfiguration(adminClient.Ctx(),
		&auth.SetConfigurationRequest{Configuration: nil})
	require.NoError(t, err)

	// Read the configuration that was just written
	configResp, err = adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	conf = proto.Clone(&defaultAuthConfig).(*auth.AuthConfig)
	conf.LiveConfigVersion = 3 // increment version
	requireConfigsEqual(t, conf, configResp.Configuration)
	deleteAll(t)
}

// TestConfigBlindWrite tests blind-writing a config by not setting
// live_config_version.
func TestConfigBlindWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)

	adminClient := getPachClient(t, admin)

	// Set a configuration
	conf := &auth.AuthConfig{
		IDProviders: []*auth.IDProvider{{
			Name:        "idp",
			Description: "fake IdP for testing",
			SAML: &auth.IDProvider_SAMLOptions{
				MetadataXML: SimpleSAMLIDPMetadata(t),
			},
		}},
		SAMLServiceOptions: &auth.AuthConfig_SAMLServiceOptions{
			ACSURL: "http://acs", MetadataURL: "http://metadata",
		},
	}
	_, err := adminClient.SetConfiguration(adminClient.Ctx(),
		&auth.SetConfigurationRequest{Configuration: conf})
	require.NoError(t, err)
	// Read the config back
	configResp, err := adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	conf.LiveConfigVersion = 2 // increment version ("default" config has v=1)
	requireConfigsEqual(t, conf, configResp.Configuration)

	// blind-write a new config & read it back
	conf.LiveConfigVersion = 0         // unset
	conf.IDProviders[0].Name = "idp_2" // change config
	_, err = adminClient.SetConfiguration(adminClient.Ctx(),
		&auth.SetConfigurationRequest{Configuration: conf})
	require.NoError(t, err)
	configResp, err = adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	conf.LiveConfigVersion = 3 // increment version (*past* previously-read cfg)
	requireConfigsEqual(t, conf, configResp.Configuration)
	deleteAll(t)
}

// TestInitialConfigConflict tests that a R+M+W even of pachyderm's initial
// config won't work if the change is interrupted (in other words, that the
// default config has version > 0, and the R+M+W isn't treated as a blind write)
func TestInitialConfigConflict(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)

	adminClient := getPachClient(t, admin)
	resp, err := adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	initialConfig := resp.Configuration

	// Set a configuration
	interruptingConf := &auth.AuthConfig{
		IDProviders: []*auth.IDProvider{{
			Name:        "idp",
			Description: "fake IdP for testing",
			SAML: &auth.IDProvider_SAMLOptions{
				MetadataXML: SimpleSAMLIDPMetadata(t),
			},
		}},
		SAMLServiceOptions: &auth.AuthConfig_SAMLServiceOptions{
			ACSURL: "http://acs", MetadataURL: "http://metadata",
		},
	}
	_, err = adminClient.SetConfiguration(adminClient.Ctx(),
		&auth.SetConfigurationRequest{Configuration: interruptingConf})
	require.NoError(t, err)

	// try to update the config using a modified initialConfig (should fail b/c of
	// interruptingConfig above
	initialConfig.IDProviders = append(initialConfig.IDProviders,
		&auth.IDProvider{
			Name:        "idp",
			Description: "fake IdP for testing",
			SAML: &auth.IDProvider_SAMLOptions{
				MetadataXML: SimpleSAMLIDPMetadata(t),
			},
		})
	initialConfig.SAMLServiceOptions = &auth.AuthConfig_SAMLServiceOptions{
		ACSURL: "http://acs", MetadataURL: "http://metadata",
	}
	require.Equal(t, int64(1), initialConfig.LiveConfigVersion)
	_, err = adminClient.SetConfiguration(adminClient.Ctx(),
		&auth.SetConfigurationRequest{Configuration: initialConfig})
	require.YesError(t, err)
	require.Matches(t, "config version", err.Error())
	deleteAll(t)
}
