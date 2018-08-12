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
	"time"

	"github.com/crewjam/saml"
	logrus "github.com/sirupsen/logrus"

	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
)

func validateConfig() error {
	// TODO - take the validation stuff below (that's called by watchConfig) and move it in here (called by setConfig
	return nil
}

// lookupIDPMetadata takes the URL of a SAML IdP's Metadata service, queries it,
// parses the result, and returns it as a struct the crewjam/saml library can
// use
// This code is heavily based on the crewjam/saml/samlsp.Middleware constructor
func lookupIDPMetadata(name string, mdURL *url.URL) (*saml.EntityDescriptor, error) {
	c := http.DefaultClient
	req, err := http.NewRequest("GET", mdURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve IdP metadata for %q: %v", name, err)
	}
	req.Header.Set("User-Agent", "Golang; github.com/pachyderm/pachdyerm")

	var rawMetadata []byte
	b := backoff.NewInfiniteBackOff()
	b.MaxElapsedTime = 15 * time.Second
	backoff.RetryNotify(func() error {
		resp, err := c.Do(req)
		if err != nil {
			return nil
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("%d %s", resp.StatusCode, resp.Status)
		}
		rawMetadata, err = ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return err
		}
		return nil
	}, b, func(err error, d time.Duration) error {
		logrus.Printf("error retrieving IdP metadata: %v; retrying in %v", err, d)
		return nil
	})

	// Successfully retrieved metadata--try parsing it
	entity := &saml.EntityDescriptor{}
	err = xml.Unmarshal(rawMetadata, entity)
	if err != nil {
		// this comparison is ugly, but it is how the error is generated in
		// encoding/xml
		if err.Error() != "expected element type <EntityDescriptor> but have <EntitiesDescriptor>" {
			return nil, err
		}
		// Search through <EntitiesDescriptor> & find IdP entity
		entities := &saml.EntitiesDescriptor{}
		if err := xml.Unmarshal(rawMetadata, entities); err != nil {
			return nil, err
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
	fmt.Printf(">>> (apiServer.updateSAMLSP) updating server SAML options\n")
	a.samlSPMu.Lock()
	defer a.samlSPMu.Unlock()

	if a.configCache.SAMLServiceOptions == nil {
		return nil // no config options to copy
	}
	sso := a.configCache.SAMLServiceOptions

	// parse ACS URL
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

	var samlProvider string
	var idpMetadataURL *url.URL
	for _, idp := range a.configCache.IDProviders {
		// Check if the IDP is a known type (right now the only type of IdP is SAML)
		if idp.SAML == nil {
			idpConfigAsJSON, err := json.MarshalIndent(idp, "", "  ")
			idpConfigMsg := string(idpConfigAsJSON)
			if err != nil {
				idpConfigMsg = fmt.Sprint("(could not marshal config json: %v)", err)
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
		if err != nil {
			return fmt.Errorf("could not parse SAML IdP Metadata URL: %v", err)
		}
		if idpMetadataURL.Scheme == "" {
			return fmt.Errorf("invalid SAML IdP Metadata URL (no scheme): %v", err)
		}
	}
	// Create a.samlSP
	if a.samlSP == nil {
		// Lookup full IdP metadata from URL
		fmt.Printf(">>> looking up IdPMetdata")
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
	fmt.Printf(">>> (apiServer.updateSAMLSP) a.samlSP: %+v\n", a.samlSP)

	// Set ACS URL and metadata URL from config
	a.samlSP.AcsURL = *acsURL
	a.samlSP.MetadataURL = *metadataURL
	// a.samlSP.IDPMetadata

	return nil
}

func (a *apiServer) handleSAMLResponse(w http.ResponseWriter, req *http.Request) {
	a.samlSPMu.Lock()
	defer a.samlSPMu.Unlock()
	if a.samlSP == nil {
		fmt.Printf(">>> (apiServer.handleSAMLResponse) samlSP is nil\n")
		http.Error(w, "SAML ACS has not been configured", http.StatusConflict)
		return
	}
	sp := a.samlSP

	// stat := statReadCloser{
	// 	ReadCloser: req.Body,
	// }
	// req.Body = &stat
	out := io.MultiWriter(w, os.Stdout)
	possibleRequestIDs := []string{""} // only IdP-initiated auth enabled for now
	fmt.Printf(">>> (apiServer.handleSAMLResponse) req.PostFormValue(\"SAMLResponse\"): %s\n", req.PostFormValue("SAMLResponse"))
	assertion, err := sp.ParseResponse(req, possibleRequestIDs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		out.Write([]byte("<html><head></head><body>"))
		out.Write([]byte("Error parsing SAML response: "))
		out.Write([]byte(err.Error()))
		ie := err.(*saml.InvalidResponseError)
		out.Write([]byte("\nPrivate error: " + ie.PrivateErr.Error()))
		out.Write([]byte("\n"))
	}
	if assertion == nil {
		w.WriteHeader(http.StatusInternalServerError)
		out.Write([]byte("<html><head></head><body>"))
		out.Write([]byte("Nil assertion\n"))
	}
	out.Write([]byte("Would've authenticated as:\n"))
	fmt.Printf(">>> (apiServer.handleSAMLResponse) This is a test\n")
	if assertion != nil {
		out.Write([]byte(fmt.Sprintf("assertion: %v\n", assertion)))
		if assertion.Subject != nil {
			out.Write([]byte(fmt.Sprintf("assertion.Subject: %v\n", assertion.Subject)))
			if assertion.Subject.NameID != nil {
				out.Write([]byte(fmt.Sprintf("assertion.Subject.NameID.Value: %v\n", assertion.Subject.NameID.Value)))
			}
		}
		if assertion.Element() != nil {
			xmlBytes, err := xml.MarshalIndent(assertion.Element(), "", "  ")
			if err != nil {
				out.Write([]byte(fmt.Sprintf("could not marshall assertion: %v\n", err)))
			} else {
				out.Write([]byte(fmt.Sprintf("<pre>\n%s\n</pre>", xmlBytes)))
			}
		} else {
			out.Write([]byte("assertion.Element() was nil"))
		}
	} else {
		out.Write([]byte("assertion was nil"))
	}
	out.Write([]byte("</body></html>"))
}

func (a *apiServer) handleMetadata(w http.ResponseWriter, req *http.Request) {
	buf, _ := xml.MarshalIndent(a.samlSP.Metadata(), "", "  ")
	w.Header().Set("Content-Type", "application/samlmetadata+xml")
	w.Write(buf)
	return
}

func (a *apiServer) serveSAML() {
	fmt.Printf(">>> (apiServer.serveSAML) Entering serveSAML\n")
	samlMux := http.NewServeMux()
	samlMux.HandleFunc("/saml/acs", a.handleSAMLResponse)
	samlMux.HandleFunc("/saml/metadata", a.handleMetadata)
	samlMux.HandleFunc("/*", func(w http.ResponseWriter, req *http.Request) {
		fmt.Printf(">>> (apiServer.serveSAML) received request to %s\n", req.URL.Path)
		w.WriteHeader(http.StatusTeapot)
	})
	http.ListenAndServe(fmt.Sprintf(":%d", SamlPort), samlMux)
}
