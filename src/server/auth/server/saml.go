package server

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/crewjam/saml"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/errutil"
)

// DefaultDashRedirectURL is the default URL used for redirecting the dashboard
var DefaultDashRedirectURL = &url.URL{
	Scheme: "http",
	Host:   "localhost:30080",
	// leading slash is for equality checks in tests (HTTP lib inserts it)
	Path: path.Join("/", "auth", "autologin"),
}

// handleSAMLResponseInternal is a helper function called by handleSAMLResponse
func (a *apiServer) handleSAMLResponseInternal(cfg *canonicalConfig, sp *saml.ServiceProvider, req *http.Request) (string, string, *errutil.HTTPError) {
	if err := req.ParseForm(); err != nil {
		return "", "", errutil.NewHTTPError(http.StatusConflict, "Could not parse request form: %v", err)
	}
	// No possible request IDs b/c only IdP-initiated auth is implemented for now
	assertion, err := sp.ParseResponse(req, []string{""})
	if err != nil {
		errMsg := fmt.Sprintf("Error parsing SAML response: %v", err)
		var invalidRespErr *saml.InvalidResponseError
		if errors.As(err, invalidRespErr) {
			errMsg += "\n(" + invalidRespErr.PrivateErr.Error() + ")"
		}
		return "", "", errutil.NewHTTPError(http.StatusBadRequest, errMsg)
	}

	// Make sure all the fields we need are present (avoid segfault)
	switch {
	case assertion == nil:
		return "", "", errutil.NewHTTPError(http.StatusConflict, "Error parsing SAML response: assertion is nil")
	case assertion.Subject == nil:
		return "", "", errutil.NewHTTPError(http.StatusConflict, "Error parsing SAML response: assertion.Subject is nil")
	case assertion.Subject.NameID == nil:
		return "", "", errutil.NewHTTPError(http.StatusConflict, "Error parsing SAML response: assertion.Subject.NameID is nil")
	case assertion.Subject.NameID.Value == "":
		return "", "", errutil.NewHTTPError(http.StatusConflict, "Error parsing SAML response: assertion.Subject.NameID.Value is unset")
	}

	// User is successfully authenticated
	var samlIDP *canonicalIDPConfig
	for i := range cfg.IDPs {
		if cfg.IDPs[i].SAML != nil {
			samlIDP = &cfg.IDPs[i]
			break
		}
	}
	if samlIDP == nil {
		return "", "", errutil.NewHTTPError(http.StatusInternalServerError,
			"Error processing SAML assertion: assertion appears valid, but could "+
				"not find configuration for SAML ID provider")
	}
	subject := fmt.Sprintf("%s:%s", samlIDP.Name, assertion.Subject.NameID.Value)

	// Get new OTP for user (exp. from config if set, or default session duration)
	expiration := time.Now().Add(time.Duration(defaultSAMLTTLSecs) * time.Second)
	if cfg.SAMLSvc.SessionDuration != 0 {
		expiration = time.Now().Add(cfg.SAMLSvc.SessionDuration)
	}
	authCode, err := a.getOneTimePassword(req.Context(), subject,
		defaultOTPTTLSecs, expiration)
	if err != nil {
		return "", "", errutil.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Update group memberships
	if samlIDP.SAML.GroupAttribute != "" {
		for _, attribute := range assertion.AttributeStatements {
			for _, attr := range attribute.Attributes {
				if attr.Name != samlIDP.SAML.GroupAttribute {
					continue
				}

				// Collect groups specified in this attribute and record them
				var groups []string
				for _, v := range attr.Values {
					groups = append(groups, fmt.Sprintf("group/%s:%s", samlIDP.Name, v.Value))
				}
				if err := a.setGroupsForUserInternal(context.Background(), subject, groups); err != nil {
					return "", "", errutil.NewHTTPError(http.StatusInternalServerError, err.Error())
				}
			}
		}
	}

	return subject, authCode, nil
}

// handleSAMLResponse is the HTTP handler for Pachyderm's ACS, which receives
// signed SAML assertions from this cluster's SAML ID provider (if one is
// configured)
func (a *apiServer) handleSAMLResponse(w http.ResponseWriter, req *http.Request) {
	var subject, authCode string
	var err *errutil.HTTPError

	logRequest := "SAML login request"
	a.LogReq(logRequest)
	defer func(start time.Time) {
		if subject != "" {
			logRequest = fmt.Sprintf("SAML login request for %s", subject)
		}
		a.LogResp(logRequest, errutil.PrettyPrintCode(err), err, time.Since(start))
	}(time.Now())

	cfg, sp := a.getSAMLSP()
	if cfg == nil {
		http.Error(w, "auth has no active config (either never set or disabled)", http.StatusConflict)
		return
	}
	if sp == nil {
		http.Error(w, "SAML ACS has not been configured or was disabled", http.StatusConflict)
		return
	}
	subject, authCode, err = a.handleSAMLResponseInternal(cfg, sp, req)
	if err != nil {
		http.Error(w, err.Error(), err.Code())
		return
	}

	// Redirect caller back to dash with auth code
	u := *DefaultDashRedirectURL
	if cfg.SAMLSvc != nil && cfg.SAMLSvc.DashURL != nil {
		u = *cfg.SAMLSvc.DashURL
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
