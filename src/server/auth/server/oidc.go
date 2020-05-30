package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/coreos/go-oidc"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

// handler function

var tokenChan chan tokenInfo
var stateToken map[string]string

func init() {
	tokenChan = make(chan tokenInfo)
	stateToken = make(map[string]string)
}

type internalOIDCProvider struct {
	Provider     *oidc.Provider
	ClientID     string
	RedirectURL  string
	StateToToken map[string]chan tokenInfo
}

func NewOIDCIDP(ctx context.Context, issuer, clientID string) (*internalOIDCProvider, error) {
	o := &internalOIDCProvider{}
	var err error
	o.Provider, err = oidc.NewProvider(ctx, issuer)
	if o.RedirectURL == "" {
		o.RedirectURL = "http://localhost:14687/authorization-code/callback"
	}
	o.ClientID = clientID
	return o, err
}

func (o *internalOIDCProvider) GetOIDCLoginURL(state string) (string, error) {
	clientID := "pachyderm"
	nonce := "testing"
	// prepare request by filling out parameters
	conf := oauth2.Config{
		ClientID:    clientID,
		RedirectURL: o.RedirectURL,
		// Discovery returns the OAuth2 endpoints.
		Endpoint: o.Provider.Endpoint(),
		// "openid" is a required scope for OpenID Connect flows.
		Scopes: []string{oidc.ScopeOpenID},
	}

	return conf.AuthCodeURL(state,
		oauth2.SetAuthURLParam("response_type", "code"),
		oauth2.SetAuthURLParam("nonce", nonce)), nil
}

// OIDCTokenToUsername takes a OAuth access token issued by OIDC and uses
// it discover the username of the user who obtained the code (or verify that
// the code belongs to OIDCUsername). This is how Pachyderm currently
// implements authorization in a production cluster
func (o *internalOIDCProvider) OIDCTokenToUsername(ctx context.Context, state string) (string, error) {
	oauthToken, ok := stateToken[state]
	if !ok {
		return "", fmt.Errorf("did not have a valid state")
	}

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

	return userInfo.Profile, nil
}

func (a *apiServer) handleExchange(w http.ResponseWriter, req *http.Request) {
	ctx := context.Background()
	//	"http://172.17.0.2:8080/auth/realms/adele-testing"
	cfg, sp := a.getOIDCSP()
	if cfg == nil {
		http.Error(w, "auth has no active config (either never set or disabled)", http.StatusConflict)
		return
	}
	if sp == nil {
		http.Error(w, "OIDC has not been configured or was disabled", http.StatusConflict)
		return
	}

	conf := &oauth2.Config{
		ClientID:    a.oidcSP.ClientID,
		RedirectURL: "http://localhost:14687/authorization-code/callback",
		// Scopes:       []string{"openid"},
		Endpoint: sp.Provider.Endpoint(),
	}

	code := req.URL.Query()["code"][0]
	state := req.URL.Query()["state"][0]

	// Use the authorization code that is pushed to the redirect
	// URL
	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		log.Fatal(err)
	}

	stateToken[state] = tok.AccessToken

	tokenChan <- tokenInfo{token: tok.AccessToken, err: err}

	fmt.Fprintf(w, "You are now logged in. Go back to the terminal to use Pachyderm!")

}

func (a *apiServer) serveOIDC() {
	// serve OIDC handler to exchange the auth code
	http.HandleFunc("/authorization-code/callback", a.handleExchange)
	log.Fatal(http.ListenAndServe(":14687", nil))
}
