package server

import (
	"context"
	goerr "errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/random"
	"github.com/pachyderm/pachyderm/v2/src/internal/watch"

	oidc "github.com/coreos/go-oidc"
	logrus "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

const threeMinutes = 3 * 60 // Passed to col.PutTTL (so value is in seconds)

// various oidc invalid argument errors. Use 'goerror' instead of internal
// 'errors' library b/c stack trace isn't useful
var (
	errNotConfigured = goerr.New("OIDC ID provider configuration not found")
	errAuthFailed    = goerr.New("authorization failed")
	errWatchFailed   = goerr.New("error watching OIDC state token (has it expired?)")
	errTokenDeleted  = goerr.New("error during authorization: OIDC state token expired")
)

// IDTokenClaims represents the set of claims in an OIDC ID token that we're concerned with
type IDTokenClaims struct {
	Email         string   `json:"email"`
	EmailVerified bool     `json:"email_verified"`
	Groups        []string `json:"groups"`
}

// validateOIDC validates an OIDC configuration before it's stored in etcd.
func validateOIDCConfig(ctx context.Context, config *auth.OIDCConfig) error {
	if _, err := url.Parse(config.Issuer); err != nil {
		return errors.Wrapf(err, "OIDC issuer must be a valid URL")
	}

	// this does a request to <issuer>/.well-known/openid-configuration to see if it works
	_, err := newOIDCConfig(ctx, config)

	if err != nil {
		return errors.Wrapf(err, "provided OIDC issuer does not implement OIDC protocol")
	}

	if _, err := url.Parse(config.RedirectURI); err != nil {
		return errors.Wrapf(err, "OIDC redirect_uri must be a valid URL")
	}

	if config.ClientID == "" {
		return errors.Errorf("OIDC configuration must have a non-empty client_id")
	}

	return nil
}

// half is a helper function used to log the first half of OIDC state tokens in
// logs.
//
// Per the description of handleOIDCLogin, we currently don't give error details
// to callers of Authenticate/handleOIDCCallback, to avoid accidentally leaking
// sensitive information to untrusted users, and instead log error information
// from pachd (where only kubernetes administrators can see it) with the state
// token inline. This way, legitimate users having trouble authenticating can
// show their state token to a cluster administrator and get error information
// from them. However, to avoid giving too much user information to Kubernetes
// cluster administrators, we don't want to log users' private credentials. So
// this function is used to log part of an OIDC state token--enough to associate
// error logs with a failing authentication flow, but not enough for a cluster
// administrator to impersonate a user.
func half(state string) string {
	return fmt.Sprintf("%s.../%d", state[:len(state)/2], len(state))
}

// oidcConfig is a struct that wraps an auth.OIDCConfig and adds useful
// structs that we would normally need to instantiate multiple times
type oidcConfig struct {
	*auth.OIDCConfig
	oidcProvider      *oidc.Provider
	oauthConfig       oauth2.Config
	rewriteClient     *http.Client
	userAccessAddress string
}

func newOIDCConfig(ctx context.Context, config *auth.OIDCConfig) (*oidcConfig, error) {
	var rewriteClient *http.Client
	var err error
	if config.LocalhostIssuer {
		rewriteClient, err = LocalhostRewriteClient(config.Issuer)
		if err != nil {
			return nil, err
		}
		ctx = oidc.ClientContext(ctx, rewriteClient)
	}
	oidcProvider, err := oidc.NewProvider(ctx, config.Issuer)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}

	return &oidcConfig{
		OIDCConfig:        config,
		oidcProvider:      oidcProvider,
		rewriteClient:     rewriteClient,
		userAccessAddress: config.UserAccessibleIssuerHost,
		oauthConfig: oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			RedirectURL:  config.RedirectURI,
			Endpoint:     oidcProvider.Endpoint(),
			Scopes:       config.Scopes,
		},
	}, nil
}

func (c *oidcConfig) Ctx(ctx context.Context) context.Context {
	if c.rewriteClient != nil {
		return oidc.ClientContext(ctx, c.rewriteClient)
	}
	return ctx
}

func (a *apiServer) getOIDCConfig(ctx context.Context) (*oidcConfig, error) {
	config, ok := a.configCache.Load().(*auth.OIDCConfig)
	if !ok {
		return nil, errors.New("unable to load cached OIDC configuration")
	}
	if config.Issuer == "" {
		return nil, errors.WithStack(errNotConfigured)
	}
	return newOIDCConfig(ctx, config)
}

// GetOIDCLoginURL uses the given state to generate a login URL for the OIDC provider object
func (a *apiServer) GetOIDCLoginURL(ctx context.Context) (string, string, error) {
	config, err := a.getOIDCConfig(ctx)
	if err != nil {
		return "", "", err
	}

	// state ties this login to the auth code retrieved by the user
	// nonce ties this login to the access/identity token returned from the IDP
	state := random.String(30)
	nonce := random.String(30)

	if _, err := col.NewSTM(ctx, a.env.EtcdClient, func(stm col.STM) error {
		return errors.EnsureStack(a.oidcStates.ReadWrite(stm).PutTTL(state, &auth.SessionInfo{
			Nonce: nonce, // read & verified by /authorization-code/callback
		}, threeMinutes))
	}); err != nil {
		return "", "", errors.Wrap(err, "could not create OIDC login session")
	}

	authURL := config.oauthConfig.AuthCodeURL(state,
		oauth2.SetAuthURLParam("response_type", "code"),
		oauth2.SetAuthURLParam("nonce", nonce))

	if config.userAccessAddress != "" {
		rewriteURL, err := url.Parse(authURL)
		if err != nil {
			return "", "", errors.Wrap(err, "could not parse Auth URL for Localhost Issuer rewrite")
		}

		userURL, err := url.Parse(config.userAccessAddress)
		if err != nil {
			return "", "", errors.Wrap(err, "could not parse User Accessible Oauth Issuer URL")
		}
		rewriteURL.Host = userURL.Host
		rewriteURL.Scheme = userURL.Scheme
		authURL = rewriteURL.String()
	}
	return authURL, state, nil
}

// OIDCStateToEmail takes the state token created for the OIDC session and
// uses it discover the email of the user who obtained the code (or verify that
// the code belongs to them). This is how Pachyderm currently implements OIDC
// authorization in a production cluster
func (a *apiServer) OIDCStateToEmail(ctx context.Context, state string) (email string, retErr error) {
	defer func() {
		logrus.Infof("converted OIDC state %q to email %q (or err: %v)",
			half(state), email, retErr)
	}()
	// reestablish watch in a loop, in case there's a watch error
	if err := backoff.RetryNotify(func() error {
		watcher, err := a.oidcStates.ReadOnly(ctx).WatchOne(state)
		if err != nil {
			logrus.Errorf("error watching OIDC state token %q during authorization: %v",
				half(state), err)
			return errors.WithStack(errWatchFailed)
		}
		defer watcher.Close()

		// lookup the token from the given state
		for e := range watcher.Watch() {
			if e.Type == watch.EventError {
				// reestablish watch (error not returned to user)
				return e.Err
			} else if e.Type == watch.EventDelete {
				return errors.WithStack(errTokenDeleted)
			}

			// see if there's an ID token attached to the OIDC state now
			var key string
			var si auth.SessionInfo
			if err := e.Unmarshal(&key, &si); err != nil {
				// retry watch (maybe a valid SessionInfo will appear later?)
				return errors.Wrapf(err, "error unmarshalling OIDC SessionInfo")
			}
			if si.ConversionErr {
				return errors.WithStack(errAuthFailed)
			} else if si.Email != "" {
				// Success
				email = si.Email
				return nil
			}
		}
		return nil
	}, backoff.New60sBackOff(), func(err error, d time.Duration) error {
		logrus.Errorf("error watching OIDC state token %q during authorization (retrying in %s): %v",
			half(state), d, err)
		if errors.Is(err, errWatchFailed) || errors.Is(err, errTokenDeleted) || errors.Is(err, errAuthFailed) {
			return err // don't retry, just return the error
		}
		return nil
	}); err != nil {
		return "", err
	}
	return email, nil
}

// handleOIDCExchange implements the /authorization-code/callback endpoint. In
// the success case, it converts the passed authorization code to an email
// address and associates the email address with the passed OIDC state token in
// the 'oidc-authns' collection.
//
// The error handling from this function is slightly delicate, as callers may
// have network access to Pachyderm, but may not have an OIDC account or any
// legitimate access to this cluster, so we want to avoid accidentally leaking
// operational details. In general:
// - This should not return an HTTP error with more information than pachctl
//   prints. Currently, pachctl only prints the OIDC state token presented by
//   the user and "Authorization failed" if the token exchange doesn't work
//   (indicated by SessionInfo.ConversionErr == true).
// - More information may be included in logs (which should only be accessible
//   Pachyderm administrators with kubectl access), and logs include enough
//   characters of any relevant OIDC state token to identify a particular login
//   flow. Thus if a user is legitimate, they can present their OIDC state token
//   (displayed by pachctl or their browser) to a cluster administrator, and the
//   cluster administrator can locate a detailed error in pachctl's logs.
//   Together they can resolve any authorization issues.
// - This should also not log any user credentials that would allow a
//   kubernetes cluster administrator to impersonate an individual user
//   undetected in Pachyderm or elsewhere. Where this logs OIDC state tokens, to
//   correlate authentication flows to error logs, it only logs the first half,
//   which is not enough to authenticate.
//
// If needed, Pachyderm cluster administrators can impersonate users by calling
// GetAuthToken(), but that call is logged and auditable.
func (a *apiServer) handleOIDCExchange(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	code := req.URL.Query()["code"][0]
	state := req.URL.Query()["state"][0]
	if state == "" || code == "" {
		http.Error(w,
			"invalid OIDC callback request: missing OIDC state token or authorization code",
			http.StatusBadRequest)
		return
	}

	// Verify the ID token, and if it's valid, add it to this state's SessionInfo
	// in postgres, so that any concurrent Authorize() calls can discover it and give
	// the caller a Pachyderm token.
	nonce, email, conversionErr := a.handleOIDCExchangeInternal(
		context.Background(), code, state)
	_, txErr := col.NewSTM(ctx, a.env.EtcdClient, func(stm col.STM) error {
		var si auth.SessionInfo
		err := a.oidcStates.ReadWrite(stm).Update(state, &si, func() error {
			// nonce can only be checked inside postgres txn, but if nonces don't match
			// that's a non-retryable authentication error, so set conversionErr as
			// if handleOIDCExchangeInternal had errored and proceed
			if conversionErr == nil && nonce != si.Nonce {
				conversionErr = fmt.Errorf(
					"IDP nonce %v did not match Pachyderm's session nonce %v",
					nonce, si.Nonce)
			}
			if conversionErr == nil {
				si.Email = email
			} else {
				si.ConversionErr = true
			}
			return nil
		})
		return errors.EnsureStack(err)
	})
	// Make exactly one call, to http.Error or http.Write, with either
	// conversionErr (non-retryable) or txErr (retryable) if either is set
	switch {
	case conversionErr != nil:
		// Don't give the user specific error information
		http.Error(w,
			fmt.Sprintf("authorization failed (OIDC state token: %q; Pachyderm "+
				"logs may contain more information)", half(state)),
			http.StatusUnauthorized)
	case txErr != nil:
		http.Error(w,
			fmt.Sprintf("temporary error during authorization (OIDC state token: "+
				"%q; Pachyderm logs may contain more information)", half(state)),
			http.StatusInternalServerError)
	default:
		// Success
		fmt.Fprintf(w, "You are now logged in. Go back to the terminal to use Pachyderm!")
	}
	// Wite more detailed error information into pachd's logs, if appropriate
	// (use two ifs here vs switch in case both are set)
	if conversionErr != nil {
		logrus.Errorf("could not convert authorization code (OIDC state: %q) %v",
			half(state), conversionErr)
	}
	if txErr != nil {
		logrus.Errorf("error storing OIDC authorization code in postgres (OIDC state: %q): %v",
			half(state), txErr)
	}
}

func (a *apiServer) validateIDToken(ctx context.Context, rawIDToken string) (*oidc.IDToken, *IDTokenClaims, error) {
	config, err := a.getOIDCConfig(ctx)
	if err != nil {
		return nil, nil, err
	}

	var verifier = config.oidcProvider.Verifier(&oidc.Config{ClientID: config.ClientID})
	idToken, err := verifier.Verify(config.Ctx(ctx), rawIDToken)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "could not verify token")
	}

	var claims IDTokenClaims
	if err := idToken.Claims(&claims); err != nil {
		return nil, nil, errors.Wrapf(err, "could not get claims")
	}

	if !claims.EmailVerified && config.RequireEmailVerified {
		return nil, nil, errors.New("email_verified claim was false, and require_email_verified was set")
	}
	return idToken, &claims, nil
}

func (a *apiServer) syncGroupMembership(ctx context.Context, claims *IDTokenClaims) error {
	groups := make([]string, len(claims.Groups))
	for i, g := range claims.Groups {
		groups[i] = fmt.Sprintf("%s%s", auth.GroupPrefix, g)
	}
	// Sync group membership based on the groups claim, if any
	return a.setGroupsForUserInternal(ctx, auth.UserPrefix+claims.Email, groups)
}

// handleOIDCExchangeInternal is a convenience function for converting an
// authorization code into an access token. The caller (handleOIDCExchange) is
// responsible for storing any responses from this in postgres and sending an HTTP
// response to the user's browser.
func (a *apiServer) handleOIDCExchangeInternal(ctx context.Context, authCode, state string) (nonce, email string, retErr error) {
	// log request, but do not log auth code (short-lived, but senstive user authenticator)
	logrus.Infof("auth.OIDC.handleOIDCExchange { \"state\": %q }", half(state))
	defer func() {
		logrus.Infof("auth.OIDC.handleOIDCExchange { \"state\": %q, \"nonce\": %q, \"email\": %q }",
			half(state), nonce, email)
	}()

	config, err := a.getOIDCConfig(ctx)
	if err != nil {
		return "", "", err
	}

	// Use the authorization code that is pushed to the redirect
	tok, err := config.oauthConfig.Exchange(config.Ctx(ctx), authCode)
	if err != nil {
		return "", "", errors.Wrapf(err, "failed to exchange code")
	}

	// Extract the ID Token from OAuth2 token.
	rawIDToken, ok := tok.Extra("id_token").(string)
	if !ok {
		return "", "", errors.New("missing id token")
	}

	// Parse and verify ID Token payload.
	idToken, claims, err := a.validateIDToken(ctx, rawIDToken)
	if err != nil {
		return "", "", errors.Wrapf(err, "could not verify token")
	}

	if err := a.syncGroupMembership(ctx, claims); err != nil {
		return "", "", errors.Wrapf(err, "could not sync group membership")
	}

	return idToken.Nonce, claims.Email, nil
}

func (a *apiServer) serveOIDC() error {
	mux := http.NewServeMux()
	// serve 200 on '/' for health checks
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "200 OK")
	})
	// serve OIDC handler to exchange the auth code
	mux.HandleFunc("/authorization-code/callback", a.handleOIDCExchange)
	return errors.EnsureStack(http.ListenAndServe(fmt.Sprintf(":%v", a.env.Config.OidcPort), mux))
}
