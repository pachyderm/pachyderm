package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/coreos/go-oidc"
	"github.com/davecgh/go-spew/spew"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

// handler function

var tokenChan chan tokenInfo
var stateToken map[string]string
var issuerChan chan string

func init() {
	tokenChan = make(chan tokenInfo)
	issuerChan = make(chan string)

	stateToken = make(map[string]string)
	// create nonce
}

type internalOIDCProvider struct {
	Provider     *oidc.Provider
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	StateToToken map[string]chan tokenInfo
}

func NewOIDCIDP(ctx context.Context, issuer, clientID string, clientSecret string) (*internalOIDCProvider, error) {
	o := &internalOIDCProvider{}
	var err error
	o.Provider, err = oidc.NewProvider(ctx, issuer)
	if o.RedirectURL == "" {
		o.RedirectURL = "http://localhost:14687/authorization-code/callback"
	}
	o.Issuer = issuer
	// issuerChan <- issuer
	o.ClientID = clientID
	o.ClientSecret = clientSecret
	return o, err
}

func (o *internalOIDCProvider) GetOIDCLoginURL(state string) (string, error) {
	// clientID := "pachyderm"
	nonce := "testing"
	var err error
	// prepare request by filling out parameters
	if o.Provider == nil {
		o.Provider, err = oidc.NewProvider(context.Background(), o.Issuer)
		if err != nil {
			return "", fmt.Errorf("provider could not be found: %v", err)
		}
		// issuerChan <- o.Issuer
	}
	conf := oauth2.Config{
		ClientID:     o.ClientID,
		ClientSecret: o.ClientSecret,
		RedirectURL:  o.RedirectURL,
		// Discovery returns the OAuth2 endpoints.
		Endpoint: o.Provider.Endpoint(),
		// "openid" is a required scope for OpenID Connect flows.
		Scopes: []string{oidc.ScopeOpenID, "profile", "email"},
	}

	url := conf.AuthCodeURL(state,
		oauth2.SetAuthURLParam("response_type", "code"),
		oauth2.SetAuthURLParam("nonce", nonce))
	return url, nil
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

	fmt.Println("token is", oauthToken)

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
	spew.Dump(userInfo)
	// return "adele@pachyderm.io", nil
	return userInfo.Email, nil
}

func (a *apiServer) handleExchange(w http.ResponseWriter, req *http.Request) {
	ctx := context.Background()

	var err error
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
	if sp.Provider == nil {
		// issuer := <-issuerChan
		sp.Provider, err = oidc.NewProvider(context.Background(), sp.Issuer)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "OIDC Provider could not be found", http.StatusConflict)
			return
		}
	}
	fmt.Println("clinet id", sp.ClientID)
	conf := &oauth2.Config{
		ClientID:     sp.ClientID,
		ClientSecret: sp.ClientSecret,
		RedirectURL:  "http://localhost:14687/authorization-code/callback",
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     sp.Provider.Endpoint(),
	}

	code := req.URL.Query()["code"][0]
	state := req.URL.Query()["state"][0]
	// do we get the nonce back here?

	// check nonce
	// replace nonce

	// Use the authorization code that is pushed to the redirect
	// URL
	fmt.Println("code is:", code)
	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		log.Fatal("failed exchange", err)
	}
	fmt.Println("exchanged")

	// Use the token passed in as our token source
	ts := oauth2.StaticTokenSource(tok)
	newtok, err := ts.Token()
	if err != nil {
		panic("no token")
	}
	if newtok.AccessToken != tok.AccessToken {
		panic("not equal")
	}
	userInfo, err := sp.Provider.UserInfo(context.Background(), ts)
	if err != nil {
		fmt.Fprintf(w, "Oh no: %v", err)
	}

	spew.Dump(userInfo)
	// fmt.Println("user info:", userInfo.Email)

	stateToken[state] = tok.AccessToken
	fmt.Println("tok", tok.AccessToken)

	tokenChan <- tokenInfo{token: tok.AccessToken, err: err}

	fmt.Fprintf(w, "You are now logged in. Go back to the terminal to use Pachyderm!")

}

func (a *apiServer) serveOIDC() {
	// serve OIDC handler to exchange the auth code
	http.HandleFunc("/authorization-code/callback", a.handleExchange)
	log.Fatal(http.ListenAndServe(":14687", nil))
}
