package testutil

import (
	"bytes"
	"encoding/xml"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"github.com/beevik/etree"
	"github.com/crewjam/saml"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
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

// MockSessionProvider gives tests the opportunity to inject their own
// SAML service provider metadata into an ID provider
type MockSessionProvider func(http.ResponseWriter, *http.Request, *saml.IdpAuthnRequest) *saml.Session

// GetSession implements the corresponding method in the saml.SessionProvider
// interface for MockSessionProvider
func (f MockSessionProvider) GetSession(w http.ResponseWriter, r *http.Request, req *saml.IdpAuthnRequest) *saml.Session {
	return f(w, r, req)
}

// ConstSessionProvider gets a SessionProvider that always returns 'md'
func ConstSessionProvider(s *saml.Session) MockSessionProvider {
	return MockSessionProvider(func(w http.ResponseWriter, r *http.Request, req *saml.IdpAuthnRequest) *saml.Session {
		return s
	})
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
func GetACSAddress(t *testing.T, pachdAddress string) string {
	pachdHost, pachdPort, err := net.SplitHostPort(pachdAddress)
	require.NoError(t, err, "Could not extract pachd port--is the SAML ACS service exposed?")
	pachdPortInt, err := strconv.ParseInt(pachdPort, 10, 16) // bitwidth == 16 b/c this is a port
	require.NoError(t, err, "Could not parse saml service port: %v", err)
	return net.JoinHostPort(pachdHost, strconv.FormatInt(pachdPortInt+4, 10))
}
