package server

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/crewjam/saml"
	"github.com/gogo/protobuf/proto"

	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
)

// SimpleSAMLIDPMetadata returns a hand-generated XML snippet with metadata for
// a very minimal SAML IDP.
func SimpleSAMLIDPMetadata(t *testing.T) []byte {
	// TODO(msteffen): it should be possible to generate a simple
	// saml.IdentityProvider and generate metadata with that, but a minor bug in
	// the crewjam/saml library prevents saml.IdentityProvider instances from
	// generating metadata if they don't have a cert, so this returns
	// manually-generated metadata
	return []byte(`
	  <EntityDescriptor
		  xmlns="urn:oasis:names:tc:SAML:2.0:metadata"
			validUntil="` + time.Now().Format(time.RFC3339) + `"
			cacheDuration="PT48H"
			entityID="http://idp.example.com/metadata">
		  <IDPSSODescriptor
			  xmlns="urn:oasis:names:tc:SAML:2.0:metadata"
			  validUntil="` + time.Now().Add(24*time.Hour).Format(time.RFC3339) + `"
			  protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
		    <NameIDFormat>urn:oasis:names:tc:SAML:2.0:nameid-format:transient</NameIDFormat>
		    <SingleSignOnService
				  Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect"
				  Location="http://idp.example.com/sso">
				</SingleSignOnService>
		    <SingleSignOnService
				  Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"
				  Location="http://idp.example.com/sso">
				</SingleSignOnService>
		  </IDPSSODescriptor>
		</EntityDescriptor>`)
}

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
		LiveConfigVersion: 0,
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
	conf.LiveConfigVersion = 1 // increment version
	requireConfigsEqual(t, conf, configResp.Configuration)
}

// TestGetSetConfigAdminOnly confirms that only cluster admins can get/set the
// auth config
func TestGetSetConfigAdminOnly(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)

	alice := tu.UniqueString("alice")
	anonClient, aliceClient, adminClient := getPachClient(t, ""), getPachClient(t, alice), getPachClient(t, admin)

	// Confirm that the auth config starts out empty
	configResp, err := adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	requireConfigsEqual(t, &auth.AuthConfig{}, configResp.Configuration)

	// Alice tries to set the current configuration and fails
	conf := &auth.AuthConfig{
		LiveConfigVersion: 0,
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
	requireConfigsEqual(t, &auth.AuthConfig{}, configResp.Configuration)

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
	conf.LiveConfigVersion = 1 // increment version
	requireConfigsEqual(t, conf, configResp.Configuration)

	configResp, err = adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	requireConfigsEqual(t, conf, configResp.Configuration)
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
		LiveConfigVersion: 0,
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
	conf.LiveConfigVersion = 1 // increment version
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
	mod2.LiveConfigVersion = 2 // increment version
	requireConfigsEqual(t, mod2, configResp.Configuration)
}

// TestSetNilConfig sets a config, then overwrites it with a nil config to
// disable the SAML service (this tests that an empty config validates and
// disables the SAML service)
func TestSetNilConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)

	adminClient := getPachClient(t, admin)

	// Set a configuration
	conf := &auth.AuthConfig{
		LiveConfigVersion: 0,
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
	conf.LiveConfigVersion = 1 // increment version
	requireConfigsEqual(t, conf, configResp.Configuration)

	// Set the nil config
	_, err = adminClient.SetConfiguration(adminClient.Ctx(),
		&auth.SetConfigurationRequest{})
	require.NoError(t, err)

	// Get the empty config
	configResp, err = adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	requireConfigsEqual(t, &auth.AuthConfig{}, configResp.Configuration)

	// TODO Make sure SAML has been deactivated
}

// TestSetEmptyConfig is like TestSetNilConfig, but it passes a non-nil but
// empty config to disable SAML
func TestSetEmptyConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)

	adminClient := getPachClient(t, admin)

	// Set a configuration
	conf := &auth.AuthConfig{
		LiveConfigVersion: 0,
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
	conf.LiveConfigVersion = 1 // increment version
	requireConfigsEqual(t, conf, configResp.Configuration)

	// Set the empty config
	_, err = adminClient.SetConfiguration(adminClient.Ctx(),
		&auth.SetConfigurationRequest{
			Configuration: &auth.AuthConfig{},
		})
	require.NoError(t, err)

	// Get the empty config
	configResp, err = adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	requireConfigsEqual(t, &auth.AuthConfig{}, configResp.Configuration)

	// TODO Make sure SAML has been deactivated
}

// TestGetEmptyConfig sets a config, then Deactivates+Reactivates auth, then
// calls GetConfig on an empty cluster to be sure the config was cleared
func TestGetEmptyConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)

	adminClient := getPachClient(t, admin)

	// Set a configuration
	conf := &auth.AuthConfig{
		LiveConfigVersion: 0,
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
	conf.LiveConfigVersion = 1 // increment version
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
		GitHubToken: "admin",
	})
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
	requireConfigsEqual(t, &auth.AuthConfig{}, configResp.Configuration)

	// Set the configuration (again)
	conf.LiveConfigVersion = 0 // reset to 0 (since config is empty)
	_, err = adminClient.SetConfiguration(adminClient.Ctx(),
		&auth.SetConfigurationRequest{Configuration: conf})
	require.NoError(t, err)

	// Get the configuration, and confirm that the config has been updated
	configResp, err = adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	conf.LiveConfigVersion = 1 // increment version
	requireConfigsEqual(t, conf, configResp.Configuration)
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
		LiveConfigVersion: 0,
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
	requireConfigsEqual(t, &auth.AuthConfig{}, configResp.Configuration)
}

// TestValidateConfigErrReservedName tests that SetConfig rejects configs that
// try to use a reserved name
func TestValidateConfigErrReservedName(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	adminClient := getPachClient(t, admin)

	for _, name := range []string{"github", "robot", "pipeline"} {
		conf := &auth.AuthConfig{
			LiveConfigVersion: 0,
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
	requireConfigsEqual(t, &auth.AuthConfig{}, configResp.Configuration)
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
		LiveConfigVersion: 0,
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
	requireConfigsEqual(t, &auth.AuthConfig{}, configResp.Configuration)
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
		LiveConfigVersion: 0,
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
	requireConfigsEqual(t, &auth.AuthConfig{}, configResp.Configuration)
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
		LiveConfigVersion: 0,
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
	requireConfigsEqual(t, &auth.AuthConfig{}, configResp.Configuration)
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

	conf := &auth.AuthConfig{
		LiveConfigVersion: 0,
		IDProviders: []*auth.IDProvider{&auth.IDProvider{
			Name:        "idp",
			Description: "fake IdP for testing",
			SAML: &auth.IDProvider_SAMLOptions{
				MetadataURL: "http://idp.example.com/metadata",
				MetadataXML: SimpleSAMLIDPMetadata(t),
			},
		}},
		SAMLServiceOptions: &auth.AuthConfig_SAMLServiceOptions{
			ACSURL: "http://acs", MetadataURL: "http://metadata",
		},
	}
	_, err := adminClient.SetConfiguration(adminClient.Ctx(),
		&auth.SetConfigurationRequest{Configuration: conf})
	require.YesError(t, err)
	require.Matches(t, "either|both", err.Error())

	// Make sure config change wasn't applied
	configResp, err := adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	requireConfigsEqual(t, &auth.AuthConfig{}, configResp.Configuration)
}

// TestValidateConfigErrMissingSAMLConfig tests that SetConfig rejects configs
// that have a SAML IdP but don't have saml_svc_options
func TestValidateConfigErrMissingSAMLConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	adminClient := getPachClient(t, admin)

	conf := &auth.AuthConfig{
		LiveConfigVersion: 0,
		IDProviders: []*auth.IDProvider{&auth.IDProvider{
			Name:        "idp",
			Description: "fake IdP for testing",
			SAML: &auth.IDProvider_SAMLOptions{
				MetadataXML: SimpleSAMLIDPMetadata(t),
			},
		}},
	}
	_, err := adminClient.SetConfiguration(adminClient.Ctx(),
		&auth.SetConfigurationRequest{Configuration: conf})
	require.YesError(t, err)
	require.Matches(t, "saml_svc_options", err.Error())

	// Make sure config change wasn't applied
	configResp, err := adminClient.GetConfiguration(adminClient.Ctx(),
		&auth.GetConfigurationRequest{})
	require.NoError(t, err)
	requireConfigsEqual(t, &auth.AuthConfig{}, configResp.Configuration)
}

func TestSAMLBasic(t *testing.T) {
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
						MetadataXML: testIDP.Metadata(),
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

	// Send SAMLResponse to Pachd and inspect the response redirect.
	// Note: this code is similar to testutil.AuthenticateWithSAMLResponse(), but
	// because we want to inspect the redirect directly, it's reproduced here.
	var resp *http.Response
	require.NoErrorWithinT(t, 300*time.Second, func() (retErr error) {
		noRedirectClient := http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		resp, err = noRedirectClient.PostForm(pachdACSURL, url.Values{
			"RelayState": []string{""},
			"SAMLResponse": []string{base64.StdEncoding.EncodeToString(
				testIDP.NewSAMLResponse("jane.doe@example.com"),
			)},
		})
		return err
	})
	buf := bytes.Buffer{}
	buf.ReadFrom(resp.Body)
	require.Equal(t, http.StatusFound, resp.StatusCode, "response body:\n"+buf.String())
	redirectLocation, err := resp.Location()
	require.NoError(t, err)
	// Compare host and path (i.e. redirect location) but not e.g. query string,
	// which contains a randomly-generated OTP
	require.Equal(t, defaultDashRedirectURL.Host, redirectLocation.Host)
	require.Equal(t, defaultDashRedirectURL.Path, redirectLocation.Path)

	otp := redirectLocation.Query()["auth_code"][0]
	require.NotEqual(t, "", otp)
	newClient := getPachClient(t, "")
	authResp, err := newClient.Authenticate(newClient.Ctx(), &auth.AuthenticateRequest{
		PachAuthenticationCode: otp,
	})
	require.NoError(t, err)
	newClient.SetAuthToken(authResp.PachToken)
	whoAmIResp, err := newClient.WhoAmI(newClient.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, "idp_1:jane.doe@example.com", whoAmIResp.Username)
}

func TestGroupsBasic(t *testing.T) {
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

	alice, group := tu.UniqueString("alice"), tu.UniqueString("group")
	aliceClient := getPachClient(t, "") // empty string b/c want anon client
	tu.AuthenticateWithSAMLResponse(t, aliceClient, testIDP.NewSAMLResponse(alice, group))

	// alice should be able to see her groups
	groupPrincipal := "group/idp_1:" + group
	groupsResp, err := aliceClient.GetGroups(aliceClient.Ctx(), &auth.GetGroupsRequest{})
	require.NoError(t, err)
	require.ElementsEqual(t, []string{groupPrincipal}, groupsResp.Groups)

	// Create a repo only accessible by 'group'--alice should be able to access it
	dataRepo := tu.UniqueString("TestGroupsBasic")
	require.NoError(t, adminClient.CreateRepo(dataRepo))
	_, err = adminClient.PutFile(dataRepo, "master", "/data", strings.NewReader("file contents"))
	require.NoError(t, err)
	_, err = adminClient.SetScope(adminClient.Ctx(), &auth.SetScopeRequest{
		Repo:     dataRepo,
		Username: groupPrincipal,
		Scope:    auth.Scope_READER,
	})
	require.NoError(t, err)
	var buf bytes.Buffer
	require.NoError(t, aliceClient.GetFile(dataRepo, "master", "/data", 0, 0, &buf))
	require.NoError(t, err)
	require.Equal(t, "file contents", buf.String())
}
