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

type tokenInfo struct {
	token string
	err   error
}

var tokenChan chan tokenInfo
var stateToken map[string]string

func init() {
	tokenChan = make(chan tokenInfo)
	stateToken = make(map[string]string)
}

func getOIDCLoginURL(state string) (string, error) {
	// Okta documentation: https://developer.okta.com/docs/reference/api/oidc/

	idp, err := oidc.NewProvider(context.Background(), "http://172.17.0.2:8080/auth/realms/adele-testing")
	if err != nil {
		return "", err
	}

	// get the redirect URI from etcd? (i think when we activate we will require this to be entered)
	redirectURL := "http://localhost:14687/authorization-code/callback"

	clientID := "pachyderm"
	// clientSecret := "YAxzH2RMFunedy1XpDz8oc8mzSFapFODaIu7QSkn"

	nonce := "testing"
	// prepare request by filling out parameters

	conf := oauth2.Config{
		ClientID: clientID,
		// ClientSecret: clientSecret,
		RedirectURL: redirectURL,

		// Discovery returns the OAuth2 endpoints.
		Endpoint: idp.Endpoint(),

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
func OIDCTokenToUsername(ctx context.Context, state string) (string, error) {
	idp, err := oidc.NewProvider(ctx, "http://172.17.0.2:8080/auth/realms/adele-testing")
	if err != nil {
		return "", err
	}

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

	userInfo, err := idp.UserInfo(ctx, ts)
	if err != nil {
		return "", err
	}

	return userInfo.Email, nil
}

func handleExchange(w http.ResponseWriter, req *http.Request) {
	ctx := context.Background()

	idp, err := oidc.NewProvider(context.Background(), "http://172.17.0.2:8080/auth/realms/adele-testing")
	if err != nil {
		log.Fatal(err)
	}

	conf := &oauth2.Config{
		ClientID:    "pachyderm",
		RedirectURL: "http://localhost:14687/authorization-code/callback",
		// Scopes:       []string{"openid"},
		Endpoint: idp.Endpoint(),
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

	http.HandleFunc("/authorization-code/callback", handleExchange)

	log.Fatal(http.ListenAndServe(":14687", nil))
}
