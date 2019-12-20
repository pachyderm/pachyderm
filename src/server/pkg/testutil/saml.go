package testutil

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"hash/fnv"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/beevik/etree"
	"github.com/crewjam/saml"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/cert"
)

// MockServiceProviderProvider gives tests the opportunity to inject their own
// SAML service provider metadata into an ID provider
type MockServiceProviderProvider func(*http.Request, string) (*saml.EntityDescriptor, error)

// GetServiceProvider implements the corresponding method in the
// ServiceProviderProvider interface for MockServiceProviderProvider
func (f MockServiceProviderProvider) GetServiceProvider(r *http.Request, serviceProviderID string) (*saml.EntityDescriptor, error) {
	return f(r, serviceProviderID)
}

// ConstServiceProviderProvider returns a ServiceProviderProvider that always
// returns 'md'
func ConstServiceProviderProvider(md *saml.EntityDescriptor) MockServiceProviderProvider {
	return MockServiceProviderProvider(func(*http.Request, string) (*saml.EntityDescriptor, error) {
		return md, nil
	})
}

type mockSessionProvider string

func (s mockSessionProvider) getSessionInternal() *saml.Session {
	// Generate ID as a hash of the subject
	id := make([]byte, 0, 4)
	h := fnv.New32()
	h.Write([]byte(s))
	h.Sum(id)

	return &saml.Session{
		ID:     string(id),
		NameID: string(s),
	}
}

// GetSession implements the corresponding method in the saml.SessionProvider
// interface for MockSessionProvider
func (s mockSessionProvider) GetSession(w http.ResponseWriter, r *http.Request, req *saml.IdpAuthnRequest) *saml.Session {
	return s.getSessionInternal()
}

// PrettyPrintEl returns 'el' indented and marshalled to a string
func PrettyPrintEl(t testing.TB, el *etree.Element) string {
	t.Helper()
	doc := etree.NewDocument()
	elCopy := el.Copy()
	doc.SetRoot(elCopy)
	doc.Indent(2)
	str, err := doc.WriteToString()
	require.NoError(t, err)
	return str
}

// MustMarshalEl marshals 'el' to []byte, or fails the current test
func MustMarshalEl(t testing.TB, el *etree.Element) []byte {
	t.Helper()
	doc := etree.NewDocument()
	doc.SetRoot(el)
	bts, err := doc.WriteToBytes()
	require.NoError(t, err)
	return bts
}

// MustMarshalXML marshals 'm' to []byte, or fails the current test
func MustMarshalXML(t testing.TB, m xml.Marshaler) []byte {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, m.MarshalXML(xml.NewEncoder(&buf), xml.StartElement{}))
	return buf.Bytes()
}

// MustParseURL marshals 'u' to a *url.URL, or fails the current test
func MustParseURL(t testing.TB, u string) *url.URL {
	t.Helper()
	result, err := url.Parse(u)
	require.NoError(t, err)
	return result
}

// GetACSAddress takes a pachd address, assumed to be of the form
//    (host):(650|30650)
// and converts it to
//    (host):(654|30654)
// and returns that
func GetACSAddress(t testing.TB, pachdAddress string) string {
	pachdHost, pachdPort, err := net.SplitHostPort(pachdAddress)
	require.NoError(t, err, "Could not extract pachd port--is the SAML ACS service exposed?")
	pachdPortInt, err := strconv.ParseInt(pachdPort, 10, 16) // bitwidth == 16 b/c this is a port
	require.NoError(t, err, "Could not parse saml service port: %v", err)
	return net.JoinHostPort(pachdHost, strconv.FormatInt(pachdPortInt+4, 10))
}

// TestIDP is a simple SAML ID provider for testing. It can generate SAML
// metadata (for configs) or SAMLResponses containing signed auth assertions
// and group memberships (for authentication)
type TestIDP struct {
	sync.Mutex
	t   testing.TB
	idp *saml.IdentityProvider
	sp  *saml.ServiceProvider
}

// Every TestIDP will be configured with these endpoints; they're used for e.g.
// certs, but never queried
const idpSSOURL = "http://idp.example.com/sso"
const idpMetadataURL = "http://idp.example.com/metadata"

// NewTestIDP generates a new x509 cert and new test SAML ID provider
func NewTestIDP(t testing.TB, pachdACSURL, pachdMetadataURL string) *TestIDP {
	idpCert, err := cert.GenerateSelfSignedCert(idpSSOURL, nil)
	require.NoError(t, err)

	result := &TestIDP{
		t: t,
	}

	// Generate service provider SAML metadata, which TestIDP.idp needs.
	// TestIDP.sp.Metadata() should be able to generate this, but there is a
	// minor bug in the crewjam/saml library that prevents ServiceProviders
	// without a certificate from generating metadata (see
	// https://github.com/crewjam/saml/pull/171). For now, we use
	// manually-generated metadata
	var spMetadata saml.EntityDescriptor
	err = xml.Unmarshal([]byte(`
		<EntityDescriptor
		  xmlns="urn:oasis:names:tc:SAML:2.0:metadata"
		  validUntil="`+time.Now().Format(time.RFC3339)+`"
		  entityID="`+pachdMetadataURL+`">
      <SPSSODescriptor
		    xmlns="urn:oasis:names:tc:SAML:2.0:metadata"
		    validUntil="`+time.Now().Add(24*time.Hour).Format(time.RFC3339)+`"
		    protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol"
		    AuthnRequestsSigned="false"
		    WantAssertionsSigned="true">
        <AssertionConsumerService
		      Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"
		      Location="`+pachdACSURL+`"
		      index="1">
		    </AssertionConsumerService>
      </SPSSODescriptor>
    </EntityDescriptor>`), &spMetadata)
	require.NoError(t, err)
	result.idp = &saml.IdentityProvider{
		Key:                     idpCert.PrivateKey,
		Certificate:             idpCert.Leaf,
		MetadataURL:             *MustParseURL(t, idpMetadataURL),
		SSOURL:                  *MustParseURL(t, idpSSOURL),
		ServiceProviderProvider: ConstServiceProviderProvider(&spMetadata),
	}
	result.sp = &saml.ServiceProvider{
		Logger:      log.New(os.Stdout, "[samlsp validation]", log.LstdFlags|log.Lshortfile),
		AcsURL:      *MustParseURL(t, pachdACSURL),
		MetadataURL: *MustParseURL(t, pachdMetadataURL),
		IDPMetadata: result.idp.Metadata(),
	}
	return result
}

// Metadata returns serialized IDP metadata for 't'
func (t *TestIDP) Metadata() []byte {
	return MustMarshalXML(t.t, t.idp.Metadata())
}

// NewSAMLResponse generates a new signed SAMLResponse asserting that the
// bearer is 'subject' and is a member of the groups in 'groups'
func (t *TestIDP) NewSAMLResponse(subject string, groups ...string) []byte {
	t.Lock()
	defer t.Unlock()

	// the SAML library doesn't have a way to generate SAMLResponses other
	// than for SP-created requests. But Pachyderm only supports IdP-initiated
	// auth, so we generate a SAMLRequest & response pair as required by the lib,
	// but unset the request ID as we'd see in a standalone response.
	clientReq, err := t.sp.MakeAuthenticationRequest(idpSSOURL)
	require.NoError(t.t, err)
	clientReq.ID = ""
	samlRequest := &saml.IdpAuthnRequest{
		Now:           time.Now(), // uh
		IDP:           t.idp,
		RequestBuffer: MustMarshalXML(t.t, clientReq),
	}
	samlRequest.HTTPRequest, err = http.NewRequest("POST", idpSSOURL, nil)
	require.NoError(t.t, err)
	// Populate SP metadata fields in samlRequest (a side effect of Validate())
	require.NoError(t.t, samlRequest.Validate())

	// generate SAML Assertion based on 'samlRequest'
	sessionProvider := mockSessionProvider(subject)
	t.idp.SessionProvider = sessionProvider
	require.NoError(t.t, saml.DefaultAssertionMaker{}.MakeAssertion(samlRequest,
		sessionProvider.getSessionInternal()))

	// Add group attribute to assertion
	groupAttribute := saml.Attribute{
		Name:       "memberOf",
		NameFormat: "urn:oasis:names:tc:SAML:2.0:attrname-format:unspecified",
		Values:     nil,
	}
	for _, group := range groups {
		groupAttribute.Values = append(groupAttribute.Values, saml.AttributeValue{
			Type:  "xs:string",
			Value: group,
		})
	}
	samlRequest.Assertion.AttributeStatements = append(
		samlRequest.Assertion.AttributeStatements, saml.AttributeStatement{
			Attributes: []saml.Attribute{groupAttribute},
		})

	// Convert Assertion struct to etree element
	require.NoError(t.t, samlRequest.MakeAssertionEl())
	// Sign assertion (a side effect of MakeResponse()
	require.NoError(t.t, samlRequest.MakeResponse())
	return MustMarshalEl(t.t, samlRequest.ResponseEl)
}

// AuthenticateWithSAMLResponse sends 'response' to Pachd's ACS, extracts the
// OTP from the response, converts it to a Pachyderm token, and automatically
// assigns it to the client 'c'
func AuthenticateWithSAMLResponse(t testing.TB, c *client.APIClient, samlResponse []byte) {
	pachdSAMLAddress := GetACSAddress(t, c.GetAddress())
	pachdACSURL := fmt.Sprintf("http://%s/saml/acs", pachdSAMLAddress)

	var resp *http.Response
	require.NoErrorWithinT(t, 300*time.Second, func() (retErr error) {
		// By default, http.Client.PostForm follows redirect, but we want to trap it
		noRedirectClient := http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		var err error
		resp, err = noRedirectClient.PostForm(pachdACSURL, url.Values{
			"RelayState": []string{""},
			// crewjam/saml library parses the SAML assertion using base64.StdEncoding
			// TODO(msteffen) is that is the SAML spec?
			"SAMLResponse": []string{base64.StdEncoding.EncodeToString(samlResponse)},
		})
		return err
	})

	// Extract OTP from redirect response & convert it to a Pachyderm token
	redirectLocation, err := resp.Location()
	require.NoError(t, err)
	otp := redirectLocation.Query()["auth_code"][0]
	require.NotEqual(t, "", otp)
	authResp, err := c.Authenticate(c.Ctx(),
		&auth.AuthenticateRequest{
			OneTimePassword: otp,
		})
	require.NoError(t, err)
	c.SetAuthToken(authResp.PachToken)
}
