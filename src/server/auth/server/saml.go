package server

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/crewjam/saml"
	logrus "github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
)

/* >>> */
// These imports are used only for debugging
import (
	"log"

	"github.com/beevik/etree"
)

/* >>> */
var dbgLog = log.New(ioutil.Discard, "", 0)

/* >>> */

func validateConfig() error {
	// TODO - take the validation stuff below (that's called by watchConfig) and move it in here (called by setConfig
	return nil
}

var defaultRedirectURL = &url.URL{
	Scheme: "http",
	Host:   "localhost:30080",
	Path:   path.Join("auth", "autologin"),
}

// lookupIDPMetadata takes the URL of a SAML IdP's Metadata service, queries it,
// parses the result, and returns it as a struct the crewjam/saml library can
// use
// This code is heavily based on the crewjam/saml/samlsp.Middleware constructor
// >>> remove names from return values
func lookupIDPMetadata(name string, mdURL *url.URL) (retDesc *saml.EntityDescriptor, retErr error) {
	defer func() {
		dbgLog.Printf("(apiServer.lookupIDPMetadata) returned error: %v\n", retErr)
	}()
	c := http.DefaultClient
	req, err := http.NewRequest("GET", mdURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve IdP metadata for %q: %v", name, err)
	}
	req.Header.Set("User-Agent", "Golang; github.com/pachyderm/pachdyerm")

	var rawMetadata []byte
	b := backoff.NewInfiniteBackOff()
	b.MaxElapsedTime = 90 * time.Second
	b.MaxInterval = 2 * time.Second
	backoff.RetryNotify(func() error {
		dbgLog.Printf("(apiServer.lookupIDPMetadata) sending req to %v\n", mdURL)
		resp, err := c.Do(req)
		if err != nil {
			return err
		}
		dbgLog.Printf("(apiServer.lookupIDPMetadata) got response:\n%+v\n", resp)

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
	})

	// Successfully retrieved metadata--try parsing it
	dbgLog.Printf("(apiServer.lookupIDPMetadata) about to parse metadata:\n%s\n", string(rawMetadata))
	entity := &saml.EntityDescriptor{}
	err = xml.Unmarshal(rawMetadata, entity)
	if err != nil {
		// this comparison is ugly, but it is how the error is generated in
		// encoding/xml
		if err.Error() != "expected element type <EntityDescriptor> but have <EntitiesDescriptor>" {
			return nil, fmt.Errorf("could not unmarshal EntityDescriptor from IdP metadata: %v", err)
		}
		// Search through <EntitiesDescriptor> & find IdP entity
		entities := &saml.EntitiesDescriptor{}
		if err := xml.Unmarshal(rawMetadata, entities); err != nil {
			return nil, fmt.Errorf("could not unmarshal EntitiesDescriptor from IdP metadata: %v", err)
		}
		for i, e := range entities.EntityDescriptors {
			if len(e.IDPSSODescriptors) > 0 {
				entity = &entities.EntityDescriptors[i]
				break
			}
		}
		// Make sure we found an IdP entity descriptor
		if len(entity.IDPSSODescriptors) == 0 {
			return nil, fmt.Errorf("no entity found with IDPSSODescriptor")
		}
	}
	return entity, nil
}

// updateSAMLSP Updates the saml SP library object in 'apiServer' after a new
// config update has been observed
// CAUTION: the caller should already hold a lock on a.configCache
func (a *apiServer) updateSAMLSP() error {
	a.samlSPMu.Lock()
	defer a.samlSPMu.Unlock()

	if a.configCache.SAMLServiceOptions == nil {
		return nil // no config options to copy
	}
	sso := a.configCache.SAMLServiceOptions

	if sso.DebugLogging {
		dbgLog = log.New(os.Stdout, "[saml] ", log.Lshortfile|log.LstdFlags|log.Lmicroseconds)
	}
	dbgLog.Printf("(apiServer.updateSAMLSP) updating server SAML options\n")

	// parse ACS URL
	dbgLog.Printf("(apiServer.updateSAMLSP) parsing ACS URL\n")
	if sso.ACSURL == "" {
		return errors.New("invalid SAML service options: must set ACS URL")
	}
	acsURL, err := url.Parse(sso.ACSURL)
	if err != nil {
		return fmt.Errorf("could not parse config ACS: %v", err)
	}
	if acsURL.Scheme == "" {
		return fmt.Errorf("ACS URL %q is invalid (no scheme)", acsURL)
	}

	// parse Metadata URL
	dbgLog.Printf("(apiServer.updateSAMLSP) parsing Metadata URL\n")
	if sso.MetadataURL == "" {
		return errors.New("invalid SAML service options: must set Metadata URL")
	}
	metadataURL, err := url.Parse(sso.MetadataURL)
	if err != nil {
		return fmt.Errorf("could not parse config ACS: %v", err)
	}
	if metadataURL.Scheme == "" {
		return fmt.Errorf("Metadata URL %q is invalid (no scheme)", metadataURL)
	}

	// parse Dash URL
	var dashURL *url.URL
	if sso.DashURL != "" {
		dashURL, err = url.Parse(sso.DashURL)
		if err != nil {
			return fmt.Errorf("could not parse dash URL \"%s\": %v", sso.DashURL, err)
		}
	}

	var samlProvider string
	var idpMetadataURL *url.URL
	for _, idp := range a.configCache.IDProviders {
		dbgLog.Printf("(apiServer.updateSAMLSP) analyzing IdP %v\n", idp.Name)
		// Check if the IDP is a known type (right now the only type of IdP is SAML)
		if idp.SAML == nil {
			idpConfigAsJSON, err := json.MarshalIndent(idp, "", "  ")
			idpConfigMsg := string(idpConfigAsJSON)
			if err != nil {
				idpConfigMsg = fmt.Sprintf("(could not marshal config json: %v)", err)
			}
			return fmt.Errorf("unrecognized ID provider: %v", idpConfigMsg)
		}

		// confirm that there is only one SAML IdP (requirement for now)
		if samlProvider != "" {
			return fmt.Errorf("two SAML providers found in config, %q and %q, but "+
				"only one is allowed", idp.Name, samlProvider)
		}
		samlProvider = idp.Name

		var err error
		idpMetadataURL, err = url.Parse(idp.SAML.MetadataURL)
		dbgLog.Printf("(apiServer.updateSAMLSP) Parsing IdP metadata URL\n")
		if err != nil {
			return fmt.Errorf("could not parse SAML IdP Metadata URL: %v", err)
		}
		if idpMetadataURL.Scheme == "" {
			return fmt.Errorf("invalid SAML IdP Metadata URL (no scheme): %v", err)
		}
	}
	dbgLog.Printf("(apiServer.updateSAMLSP) Creating samlSP\n")
	// Create a.samlSP
	if a.samlSP == nil {
		// Lookup full IdP metadata from URL
		dbgLog.Printf("looking up IdPMetdata\n")
		idpMeta, err := lookupIDPMetadata(samlProvider, idpMetadataURL)
		if err != nil {
			return err
		}

		// construct SAML handler
		a.samlSP = &saml.ServiceProvider{
			Logger:      logrus.New(),
			IDPMetadata: idpMeta,

			// Not set:
			// AcsURL: set below (derived from config)
			// MetadataURL: set below (derived from config)
			//
			// Key: Private key for Pachyderm ACS. Unclear if needed
			// Certificate: Public key for Pachyderm ACS. Unclear if needed
			// ForceAuthn: (whether users need to re-authenticate with the IdP, even
			//             if they already have a session--leaving this false)
			// AuthnNameIDFormat: (format the ACS expects the AuthnName to be in)
			// MetadataValidDuration: (how long the SP endpoints are valid? Returned
			//                        by the Metadata service)
		}
	}
	dbgLog.Printf("(apiServer.updateSAMLSP) a.samlSP: %+v\n", a.samlSP)

	// Set ACS URL and metadata URL from config
	a.samlSP.AcsURL = *acsURL
	a.samlSP.MetadataURL = *metadataURL
	// a.samlSP.IDPMetadata
	a.redirectAddress = dashURL

	return nil
}

func (a *apiServer) handleSAMLResponse(w http.ResponseWriter, req *http.Request) {
	a.configMu.Lock()
	a.samlSPMu.Lock()
	defer a.samlSPMu.Unlock()
	if a.samlSP == nil {
		dbgLog.Printf("(apiServer.handleSAMLResponse) samlSP is nil\n")
		http.Error(w, "SAML ACS has not been configured", http.StatusConflict)
		return
	}
	sp := a.samlSP

	out := io.MultiWriter(w, os.Stdout)
	possibleRequestIDs := []string{""} // only IdP-initiated auth enabled for now
	dbgLog.Printf("(apiServer.handleSAMLResponse) req.PostFormValue(\"SAMLResponse\"): %s\n", req.PostFormValue("SAMLResponse"))
	assertion, err := sp.ParseResponse(req, possibleRequestIDs)
	healthyResponse := err == nil && assertion != nil &&
		assertion.Subject != nil && assertion.Subject.NameID != nil
	dbgLog.Printf("(apiServer.handleSAMLResponse) healthyResponse: %t\n", healthyResponse)
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

	// Success
	dbgLog.Printf("(apiServer.handleSAMLResponse) Success\n")
	username := auth.SAMLPrefix + assertion.Subject.NameID.Value
	authCode, err := a.getAuthenticationCode(req.Context(), username)

	// Update group memberships
	var samlIDPConfig *auth.IDProvider_SAMLOptions
	for _, i := range a.configCache.IDProviders {
		if i.SAML != nil {
			samlIDPConfig = i.SAML
			break
		}
	}
	if samlIDPConfig != nil && samlIDPConfig.GroupAttribute != "" {
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
				if attr.Name != samlIDPConfig.GroupAttribute {
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
		}
	}

	// Redirect caller back to dash with auth code
	u := *defaultRedirectURL
	if a.redirectAddress != nil {
		u = *a.redirectAddress
	}
	u.RawQuery = url.Values{"auth_code": []string{authCode}}.Encode()
	dbgLog.Printf("(apiServer.handleSAMLResponse) Location: %s\n", u.String())
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
	dbgLog.Printf("(apiServer.serveSAML) Entering serveSAML\n")
	samlMux := http.NewServeMux()
	samlMux.HandleFunc("/saml/acs", a.handleSAMLResponse)
	samlMux.HandleFunc("/saml/metadata", a.handleMetadata)
	samlMux.HandleFunc("/*", func(w http.ResponseWriter, req *http.Request) {
		dbgLog.Printf("(apiServer.serveSAML) received request to %s\n", req.URL.Path)
		w.WriteHeader(http.StatusTeapot)
	})
	http.ListenAndServe(fmt.Sprintf(":%d", SamlPort), samlMux)
}
