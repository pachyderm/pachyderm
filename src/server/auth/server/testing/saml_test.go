package server

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
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

// TestValidateConfigMultipleSAMLIdPs tests that SetConfig rejects
// configs with a multiple SAML IDProviders
func TestValidateConfigMultipleSAMLIdPs(t *testing.T) {
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
	requireConfigsEqual(t, &defaultAuthConfig, configResp.Configuration)
	deleteAll(t)
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
	requireConfigsEqual(t, &defaultAuthConfig, configResp.Configuration)
	deleteAll(t)
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
	newClient := getPachClientConfigAgnostic(t, "")
	authResp, err := newClient.Authenticate(newClient.Ctx(), &auth.AuthenticateRequest{
		OneTimePassword: otp,
	})
	require.NoError(t, err)
	newClient.SetAuthToken(authResp.PachToken)
	whoAmIResp, err := newClient.WhoAmI(newClient.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, "idp_1:jane.doe@example.com", whoAmIResp.Username)
	deleteAll(t)
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
	aliceClient := getPachClientConfigAgnostic(t, "") // empty string b/c want anon client
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
	deleteAll(t)
}
