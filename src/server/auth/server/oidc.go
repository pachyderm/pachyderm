package server

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"

	"github.com/coreos/go-oidc"
	logrus "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

var tokenChan chan tokenInfo

type sessionInfo struct {
	Nonce string
	Token string
}

var stateInfoMap map[string]sessionInfo

func init() {
	tokenChan = make(chan tokenInfo)
	stateInfoMap = make(map[string]sessionInfo)
}

// InternalOIDCProvider is our internal representation of an oidc provider
type InternalOIDCProvider struct {
	Provider     *oidc.Provider
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// CryptoString returns a cryptographically random, URL safe string with length at least n
func CryptoString(n int) string {
	var numBytes int
	for n >= base64.StdEncoding.EncodedLen(numBytes) {
		numBytes++
	}
	b := make([]byte, numBytes)
	_, err := rand.Read(b)
	if err != nil {
		panic("could not generate cryptographically secure random string!")
	}

	return base64.StdEncoding.EncodeToString(b)
}

// NewOIDCIDP creates a new internalOIDCProvider object from the given parameters
func NewOIDCIDP(ctx context.Context, issuer, clientID, clientSecret, redirectURI string) (*InternalOIDCProvider, error) {
	o := &InternalOIDCProvider{}
	var err error
	o.Provider, err = oidc.NewProvider(ctx, issuer)
	o.Issuer = issuer
	o.ClientID = clientID
	o.ClientSecret = clientSecret
	o.RedirectURI = redirectURI
	return o, err
}

// GetOIDCLoginURL uses the given state to generate a login URL for the OIDC provider object
func (o *InternalOIDCProvider) GetOIDCLoginURL() (string, string, error) {
	state := CryptoString(30)
	nonce := CryptoString(30)
	var err error
	// prepare request by filling out parameters
	if o.Provider == nil {
		o.Provider, err = oidc.NewProvider(context.Background(), o.Issuer)
		if err != nil {
			return "", "", fmt.Errorf("provider could not be found: %v", err)
		}
	}
	conf := oauth2.Config{
		ClientID:     o.ClientID,
		ClientSecret: o.ClientSecret,
		RedirectURL:  o.RedirectURI,
		Endpoint:     o.Provider.Endpoint(),
		// "openid" is a required scope for OpenID Connect flows.
		// "profile" and "email" are necessary for using the email as an identifier
		Scopes: []string{oidc.ScopeOpenID, "profile", "email"},
	}

	si := stateInfoMap[state]
	si.Nonce = nonce
	stateInfoMap[state] = si

	url := conf.AuthCodeURL(state,
		oauth2.SetAuthURLParam("response_type", "code"),
		oauth2.SetAuthURLParam("nonce", nonce))
	return url, state, nil
}

// OIDCStateToEmail takes the state session created for the OIDC session
// and uses it discover the email of the user who obtained the
// code (or verify that the code belongs to them). This is how
// Pachyderm currently implements authorization in a production cluster
func (o *InternalOIDCProvider) OIDCStateToEmail(ctx context.Context, state string) (string, error) {
	// lookup the token from the given state
	si, ok := stateInfoMap[state]
	if !ok {
		return "", fmt.Errorf("did not have a valid state")
	}
	oauthToken := si.Token

	// Use the token passed in as our token source
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: oauthToken,
		},
	)
	userInfo, err := o.Provider.UserInfo(ctx, ts)
	if err != nil {
		return "", err
	}
	logrus.Infof("recovered user info with email: '%v'", userInfo.Email)
	return userInfo.Email, nil
}

func (a *apiServer) handleExchange(w http.ResponseWriter, req *http.Request) {
	ctx := context.Background()

	var err error
	cfg, sp := a.getOIDCSP()
	if cfg == nil {
		http.Error(w, "auth has no active config (either never set or disabled)", http.StatusConflict)
		return
	}
	if sp == nil {
		http.Error(w, "OIDC has not been configured or was disabled", http.StatusConflict)
		return
	}
	if sp.Provider == nil {
		logrus.Warn("cached provider was nil, but issuer info was present, so recovering")
		sp.Provider, err = oidc.NewProvider(context.Background(), sp.Issuer)
		if err != nil {
			logrus.Errorf("could not find provider %v: %v", sp.Issuer, err)
			http.Error(w, "OIDC Provider could not be found:", http.StatusConflict)
			return
		}
	}

	conf := &oauth2.Config{
		ClientID:     sp.ClientID,
		ClientSecret: sp.ClientSecret,
		RedirectURL:  sp.RedirectURI,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     sp.Provider.Endpoint(),
	}

	code := req.URL.Query()["code"][0]
	state := req.URL.Query()["state"][0]

	logrus.Infof("session state and code are obtained")

	// Use the authorization code that is pushed to the redirect
	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		logrus.Errorf("failed to exchange code: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	logrus.Info("exchanged OIDC code for token")

	var verifier = sp.Provider.Verifier(&oidc.Config{ClientID: conf.ClientID})
	// Extract the ID Token from OAuth2 token.
	rawIDToken, ok := tok.Extra("id_token").(string)
	if !ok {
		logrus.Errorf("missing id token")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	logrus.Infof("raw ID Token: %v", rawIDToken)

	// Parse and verify ID Token payload.
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		logrus.Errorf("could not verify token: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	si := stateInfoMap[state]
	if idToken.Nonce != si.Nonce {
		logrus.Errorf("expected nonces to match, instead set nonce %v but got nonce %v", si.Nonce, idToken.Nonce)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	logrus.Infof("nonce is %v", idToken.Nonce)

	si.Token = tok.AccessToken
	logrus.Infof("saving state with access token")
	stateInfoMap[state] = si

	// let the CLI know that we've successfully exchanged the code, and verified the token
	tokenChan <- tokenInfo{token: tok.AccessToken, err: err}
	// make sure the channel is cleaned up for future logins
	close(tokenChan)
	tokenChan = make(chan tokenInfo)

	fmt.Fprintf(w, "You are now logged in. Go back to the terminal to use Pachyderm!")
}

func (a *apiServer) serveOIDC() {
	// serve OIDC handler to exchange the auth code
	http.HandleFunc("/authorization-code/callback", a.handleExchange)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", OidcPort), nil))
}
