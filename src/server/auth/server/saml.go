package server

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/crewjam/saml"
	logrus "github.com/sirupsen/logrus"

	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
)

// canonicalConfig contains the values specified in an auth.AuthConfig proto
// message, but as structured Go types. This is populated and returned by
// validateConfig
type canonicalConfig struct {
	Version int64

	ACSURL      *url.URL
	MetadataURL *url.URL
	DashURL     *url.URL

	IDPName        string
	IDPDescription string
	IDPMetadataURL *url.URL
	IDPMetadata    *saml.EntityDescriptor
}

func (c *canonicalConfig) ToProto() *auth.AuthConfig {
	if c == nil {
		return &auth.AuthConfig{}
	}
	metadataBytes, _ := xml.MarshalIndent(c.IDPMetadata, "", "  ")
	samlIDP := &auth.IDProvider{
		Name:        c.IDPName,
		Description: c.IDPDescription,
		SAML: &auth.IDProvider_SAMLOptions{
			IDPMetadata: metadataBytes,
		},
	}
	if c.IDPMetadataURL != nil {
		samlIDP.SAML.MetadataURL = c.IDPMetadataURL.String()
	}

	result := &auth.AuthConfig{
		IDProviders: []*auth.IDProvider{samlIDP},
		SAMLServiceOptions: &auth.AuthConfig_SAMLServiceOptions{
			ACSURL:      c.ACSURL.String(),
			MetadataURL: c.MetadataURL.String(),
		},
	}
	if c.DashURL != nil {
		result.SAMLServiceOptions.DashURL = c.DashURL.String()
	}
	return result
}

// fetchRawIDPMetadata is a helper of validateConfig, below. It takes the URL
// of a SAML IdP's Metadata service, queries it, parses the result, and returns
// it as a struct the crewjam/saml library can use.
// This code is heavily based on the crewjam/saml/samlsp.Middleware constructor
func fetchRawIDPMetadata(name string, mdURL *url.URL) ([]byte, error) {
	c := http.DefaultClient
	req, err := http.NewRequest("GET", mdURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve IdP metadata for %q: %v", name, err)
	}
	req.Header.Set("User-Agent", "Golang; github.com/pachyderm/pachdyerm")

	var rawMetadata []byte
	b := backoff.NewInfiniteBackOff()
	b.MaxElapsedTime = 30 * time.Second
	b.MaxInterval = 2 * time.Second
	if err := backoff.RetryNotify(func() error {
		resp, err := c.Do(req)
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("%d %s", resp.StatusCode, resp.Status)
		}
		rawMetadata, err = ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("could not read IdP metadata response body: %v", err)
		}
		if len(rawMetadata) == 0 {
			return fmt.Errorf("empty metadata from IdP")
		}
		return nil
	}, b, func(err error, d time.Duration) error {
		logrus.Printf("error retrieving IdP metadata: %v; retrying in %v", err, d)
		return nil
	}); err != nil {
		return nil, err
	}

	// Successfully retrieved metadata
	return rawMetadata, nil
}

type configSource uint8

const (
	internal configSource = iota
	external
)

func validateConfig(config *auth.AuthConfig, src configSource) (*canonicalConfig, error) {
	c := &canonicalConfig{
		Version: config.LiveConfigVersion,
	}
	var err error

	// Validate all ID providers (and fetch IDP metadata for all SAML ID
	// providers)
	for _, idp := range config.IDProviders {
		// Validate the ID Provider's name (must exist and must not be reserved)
		if idp.Name == "" {
			return nil, errors.New("All ID providers must have a name specified " +
				"(for use during authorization)")
		}
		// TODO(msteffen): make sure we don't have to extend this every time we add
		// a new built-in backend.
		switch idp.Name + ":" {
		case auth.GitHubPrefix:
			return nil, errors.New("cannot configure auth backend with reserved prefix " +
				auth.GitHubPrefix)
		case auth.RobotPrefix:
			return nil, errors.New("cannot configure auth backend with reserved prefix " +
				auth.RobotPrefix)
		case auth.PipelinePrefix:
			return nil, errors.New("cannot configure auth backend with reserved prefix " +
				auth.PipelinePrefix)
		}

		// Check if the IDP is a known type (right now the only type of IDP is SAML)
		if idp.SAML == nil {
			// render ID provider as json for error message
			idpConfigAsJSON, err := json.MarshalIndent(idp, "", "  ")
			idpConfigMsg := string(idpConfigAsJSON)
			if err != nil {
				idpConfigMsg = fmt.Sprintf("(could not marshal config json: %v)", err)
			}
			return nil, fmt.Errorf("ID provider has unrecognized type: %v", idpConfigMsg)
		}

		// confirm that there is only one SAML IDP (requirement for now)
		if c.IDPName != "" {
			return nil, fmt.Errorf("two SAML providers found in config, %q and %q, "+
				"but only one is allowed", idp.Name, c.IDPName)
		}
		c.IDPName = idp.Name
		c.IDPDescription = idp.Description

		// construct this SAML ID provider's metadata. There are three valid cases:
		// 1. This is a user-provided config (i.e. it's coming from an RPC), and the
		//    IDP's metadata was set directly in the config
		// 2. This is a user-provided config, and the IDP's metadata was not set
		//    in the config, but the config contains a URL where the IDP metadata
		//    can be retrieved
		// 3. This is an internal config (it has already been validated by a pachd
		//    worker, and it's coming from etcd)
		// Any other case should be rejected with an error
		//
		// Either download raw IDP metadata from metadata URL or get it from cfg
		var rawIDPMetadata []byte
		if idp.SAML.MetadataURL == "" {
			if len(idp.SAML.IDPMetadata) == 0 {
				return nil, fmt.Errorf("must set either idp_metadata or metadata_url "+
					"for the SAML ID provider %q", idp.Name)
			}
			rawIDPMetadata = idp.SAML.IDPMetadata
		} else {
			// Parse URL even if this is an internal cfg and IDPMetadata is already
			// set, so that GetConfig works
			c.IDPMetadataURL, err = url.Parse(idp.SAML.MetadataURL)
			if err != nil {
				return nil, fmt.Errorf("Could not parse SAML IDP metadata URL (%q) "+
					"to query it: %v", idp.SAML.MetadataURL, err)
			} else if c.IDPMetadataURL.Scheme == "" {
				return nil, fmt.Errorf("SAML IDP metadata URL %q is invalid (no scheme)",
					idp.SAML.MetadataURL)
			}

			switch src {
			case external: // user-provided config
				if len(idp.SAML.IDPMetadata) > 0 {
					return nil, fmt.Errorf("cannot set both idp_metadata and "+
						"metadata_url for the SAML ID provider %q", idp.Name)
				}
				rawIDPMetadata, err = fetchRawIDPMetadata(idp.Name, c.IDPMetadataURL)
				if err != nil {
					return nil, err
				}

			case internal: // config from etcd
				if len(idp.SAML.IDPMetadata) == 0 {
					return nil, fmt.Errorf("internal error: the SAML ID provider %q was "+
						"persisted without IDP metadata", idp.Name)
				}
				rawIDPMetadata = idp.SAML.IDPMetadata
			}
		}

		// Parse IDP metadata. This code is heavily based on the
		// crewjam/saml/samlsp.Middleware constructor
		c.IDPMetadata = &saml.EntityDescriptor{}
		err = xml.Unmarshal(rawIDPMetadata, c.IDPMetadata)
		if err != nil {
			// this comparison is ugly, but it is how the error is generated in
			// encoding/xml
			if err.Error() != "expected element type <EntityDescriptor> but have <EntitiesDescriptor>" {
				return nil, fmt.Errorf("could not unmarshal EntityDescriptor from IDP metadata: %v", err)
			}
			// Search through <EntitiesDescriptor> & find IDP entity
			entities := &saml.EntitiesDescriptor{}
			if err := xml.Unmarshal(rawIDPMetadata, entities); err != nil {
				return nil, fmt.Errorf("could not unmarshal EntitiesDescriptor from IDP metadata: %v", err)
			}
			for i, e := range entities.EntityDescriptors {
				if len(e.IDPSSODescriptors) > 0 {
					c.IDPMetadata = &entities.EntityDescriptors[i]
					break
				}
			}
			// Make sure we found an IDP entity descriptor
			if len(c.IDPMetadata.IDPSSODescriptors) == 0 {
				return nil, fmt.Errorf("no entity found with IDPSSODescriptor")
			}
		}
	}

	// Make sure saml_svc_options are set if using SAML
	if c.IDPName != "" && config.SAMLServiceOptions == nil {
		return nil, errors.New("must set saml_svc_options if a SAML ID provider has been configured")
	}

	// Validate saml_svc_options
	if config.SAMLServiceOptions != nil {
		if c.IDPName == "" {
			return nil, errors.New("cannot set saml_svc_options without configuring a SAML ID provider")
		}
		sso := config.SAMLServiceOptions
		// parse ACS URL
		if sso.ACSURL == "" {
			return nil, errors.New("invalid SAML service options: must set ACS URL")
		}
		if c.ACSURL, err = url.Parse(sso.ACSURL); err != nil {
			return nil, fmt.Errorf("could not parse SAML config ACS URL (%q): %v", sso.ACSURL, err)
		} else if c.ACSURL.Scheme == "" {
			return nil, fmt.Errorf("ACS URL %q is invalid (no scheme)", sso.ACSURL)
		}

		// parse Metadata URL
		if sso.MetadataURL == "" {
			return nil, errors.New("invalid SAML service options: must set Metadata URL")
		}
		if c.MetadataURL, err = url.Parse(sso.MetadataURL); err != nil {
			return nil, fmt.Errorf("could not parse SAML config metadata URL (%q): %v", sso.MetadataURL, err)
		} else if c.MetadataURL.Scheme == "" {
			return nil, fmt.Errorf("Metadata URL %q is invalid (no scheme)", sso.MetadataURL)
		}

		// parse Dash URL
		if sso.DashURL != "" {
			if c.DashURL, err = url.Parse(sso.DashURL); err != nil {
				return nil, fmt.Errorf("could not parse Pachyderm dashboard URL (%q): %v", sso.DashURL, err)
			} else if c.DashURL.Scheme == "" {
				return nil, fmt.Errorf("Pachyderm dashboard URL %q is invalid (no scheme)", sso.DashURL)
			}
		}
	}

	return c, nil
}

// updateConfig validates 'config', and if it valides successfully, loads it
// into the apiServer's config cache. The caller should already hold a.configMu
// and a.samlSPMu (as this updates a.samlSP)
func (a *apiServer) updateConfig(config *auth.AuthConfig) error {
	if config != nil {
		newConfig, err := validateConfig(config, internal)
		if err != nil {
			return err
		}
		a.configCache = newConfig
	} else {
		a.configCache = nil
	}
	if a.configCache != nil && a.configCache.IDPName != "" {
		// construct SAML handler
		a.samlSP = &saml.ServiceProvider{
			Logger:      logrus.New(),
			IDPMetadata: a.configCache.IDPMetadata,
			AcsURL:      *a.configCache.ACSURL,
			MetadataURL: *a.configCache.MetadataURL,

			// Not set:
			// Key: Private key for Pachyderm ACS. Unclear if needed
			// Certificate: Public key for Pachyderm ACS. Unclear if needed
			// ForceAuthn: (whether users need to re-authenticate with the IdP, even
			//             if they already have a session--leaving this false)
			// AuthnNameIDFormat: (format the ACS expects the AuthnName to be in)
			// MetadataValidDuration: (how long the SP endpoints are valid? Returned
			//                        by the Metadata service)
		}
		a.redirectAddress = a.configCache.DashURL // Set redirect address from config as well
	} else {
		a.samlSP = nil
		a.redirectAddress = nil
	}
	return nil
}

var defaultDashRedirectURL = &url.URL{
	Scheme: "http",
	Host:   "localhost:30080",
	// leading slash is for equality checks in tests (HTTP lib inserts it)
	Path: path.Join("/", "auth", "autologin"),
}

func (a *apiServer) handleSAMLResponseInternal(req *http.Request) (string, int, error) {
	a.configMu.Lock()
	defer a.configMu.Unlock()
	a.samlSPMu.Lock()
	defer a.samlSPMu.Unlock()
	if a.configCache == nil {
		return "", http.StatusConflict, errors.New("auth has no active config (either never set or disabled)")
	}
	if a.samlSP == nil {
		return "", http.StatusConflict, errors.New("SAML ACS has not been configured or was disabled")
	}
	sp := a.samlSP

	if err := req.ParseForm(); err != nil {
		return "", http.StatusBadRequest, fmt.Errorf("Could not parse request form: %v", err)
	}
	// No possible request IDs b/c only IdP-initiated auth is implemented for now
	assertion, err := sp.ParseResponse(req, []string{""})

	if err != nil {
		errMsg := fmt.Sprintf("Error parsing SAML response: %v", err)
		if invalidRespErr, ok := err.(*saml.InvalidResponseError); ok {
			errMsg += "\n(" + invalidRespErr.PrivateErr.Error() + ")"
		}
		return "", http.StatusBadRequest, errors.New(errMsg)
	}

	// Make sure all the fields we need are present (avoid segfault)
	switch {
	case assertion == nil:
		return "", http.StatusBadRequest, errors.New("Error parsing SAML response: assertion is nil")
	case assertion.Subject == nil:
		return "", http.StatusBadRequest, errors.New("Error parsing SAML response: assertion.Subject is nil")
	case assertion.Subject.NameID == nil:
		return "", http.StatusBadRequest, errors.New("Error parsing SAML response: assertion.Subject.NameID is nil")
	case assertion.Subject.NameID.Value == "":
		return "", http.StatusBadRequest, errors.New("Error parsing SAML response: assertion.Subject.NameID.Value is unset")
	}

	out := io.MultiWriter(w, os.Stdout)
	possibleRequestIDs := []string{""} // only IdP-initiated auth enabled for now
	fmt.Printf(">>> (apiServer.handleSAMLResponse) req.PostFormValue(\"SAMLResponse\"): %s\n", req.PostFormValue("SAMLResponse"))
	assertion, err := sp.ParseResponse(req, possibleRequestIDs)
	healthyResponse := err == nil && assertion != nil &&
		assertion.Subject != nil && assertion.Subject.NameID != nil
	fmt.Printf(">>> (apiServer.handleSAMLResponse) healthyResponse: %t\n", healthyResponse)
	if !healthyResponse {
		w.WriteHeader(http.StatusInternalServerError)
		out.Write([]byte("<html><head></head><body>"))
		switch {
		case err != nil:
			out.Write([]byte("Error parsing SAML response: "))
			out.Write([]byte(err.Error()))
			if invalidRespErr, ok := err.(*saml.InvalidResponseError); ok {
				out.Write([]byte("\n(" + invalidRespErr.PrivateErr.Error() + ")"))
			}
		case assertion == nil:
			out.Write([]byte("Error parsing SAML response: assertion is nil"))
		case assertion.Subject == nil:
			out.Write([]byte("Error parsing SAML response: assertion.Subject is nil"))
		case assertion.Subject.NameID == nil:
			out.Write([]byte("Error parsing SAML response: assertion.Subject.NameID is nil"))
		default:
			out.Write([]byte("Something went wrong"))
			out.Write([]byte(fmt.Sprintf("<p>healthyResponse: %t", healthyResponse)))
			out.Write([]byte(fmt.Sprintf("<p>err: %v", err)))
			out.Write([]byte(fmt.Sprintf("<p>assertion: %v", assertion)))
		}
		out.Write([]byte("\n</body></html>"))
		return
	}
	func() {
		d := etree.NewDocument()
		d.Element = *assertion.Element()
		xml, err := d.WriteToString()
		if err != nil {
			dbgLog.Printf("(apiServer.handleSAMLResponse) could not marshall assertion: %v\n", err)
		} else {
			dbgLog.Printf("(apiServer.handleSAMLResponse) assertion: %s\n", string(xml))
		}
	}()

	// Print debug info for when we're adding groups
	for _, attribute := range assertion.AttributeStatements {
		d := etree.NewDocument()
		if attribute.Element().Parent() != nil {
			d.Element = *attribute.Element().Parent()
			xml, err := d.WriteToString()
			if err != nil {
				dbgLog.Printf("(apiServer.handleSAMLResponse) could not marshall attribute statement parent: %v\n", err)
			} else {
				dbgLog.Printf("(apiServer.handleSAMLResponse) attribute statement parent: %s\n", string(xml))
			}
		} else {
			dbgLog.Printf("(apiServer.handleSAMLResponse) could not marshall attribute statement parent: nil\n")
		}
		d.Element = *attribute.Element()
		xml, err := d.WriteToString()
		if err != nil {
			dbgLog.Printf("(apiServer.handleSAMLResponse) could not marshall attribute statement: %v\n", err)
		} else {
			dbgLog.Printf("(apiServer.handleSAMLResponse) attribute statement: %s\n", string(xml))
		}
		for _, attr := range attribute.Attributes {
			d := etree.NewDocument()
			d.Element = *attr.Element()
			xml, err := d.WriteToString()
			if err != nil {
				dbgLog.Printf("(apiServer.handleSAMLResponse) could not marshall attribute: %v\n", err)
			} else {
				dbgLog.Printf("(apiServer.handleSAMLResponse) attribute: %s\n", string(xml))
			}
			if attr.Name != "memberOf" {
				continue
			}

			var groups []string
			for _, v := range attr.Values {
				groups = append(groups, path.Join("group", auth.SAMLPrefix)+v.Value)
			}
			// TODO make this internal and call it
			dbgLog.Printf("(apiServer.handleSAMLResponse) a.setGroupsForUser(ctx, %#v)", groups)
			a.setGroupsForUser(context.Background(), username, groups)
		}
		// attribute.Attributes[0].
	}

	// Success
	username := fmt.Sprintf("%s:%s", a.configCache.IDPName, assertion.Subject.NameID.Value)
	return username, 0, nil

}

func (a *apiServer) handleSAMLResponse(w http.ResponseWriter, req *http.Request) {
	var subject string
	var err error
	var errCode int

	a.LogReq("SAML login request")
	defer func(start time.Time) {
		request := fmt.Sprintf("SAML login request for %s", subject)
		if errCode == 0 {
			errCode = http.StatusFound
		}
		a.LogResp(request, errCode, err, time.Since(start))
	}(time.Now())

	subject, errCode, err = a.handleSAMLResponseInternal(req)
	if err != nil {
		if errCode == 0 {
			errCode = http.StatusInternalServerError
		}
		http.Error(w, err.Error(), errCode)
		return
	}

	// Get new OTP for user
	authCode, err := a.getAuthenticationCode(req.Context(), subject)

	// Redirect caller back to dash with auth code
	u := *defaultDashRedirectURL
	if a.redirectAddress != nil {
		u = *a.redirectAddress
	}
	u.RawQuery = url.Values{"auth_code": []string{authCode}}.Encode()
	w.Header().Set("Location", u.String())
	w.WriteHeader(http.StatusFound) // Send redirect
}

func (a *apiServer) handleMetadata(w http.ResponseWriter, req *http.Request) {
	a.samlSPMu.Lock()
	defer a.samlSPMu.Unlock()
	buf, _ := xml.MarshalIndent(a.samlSP.Metadata(), "", "  ")
	w.Header().Set("Content-Type", "application/samlmetadata+xml")
	w.Write(buf)
	return
}

func (a *apiServer) serveSAML() {
	samlMux := http.NewServeMux()
	samlMux.HandleFunc("/saml/acs", a.handleSAMLResponse)
	samlMux.HandleFunc("/saml/metadata", a.handleMetadata)
	samlMux.HandleFunc("/*", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	http.ListenAndServe(fmt.Sprintf(":%d", SamlPort), samlMux)
}
