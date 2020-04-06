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

	// Use the authorization code that is pushed to the redirect
	// URL
	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintf(w, "Your access token is:\n\n%v", tok.AccessToken)

}

func (a *apiServer) serveOIDC() {
	// serve OIDC handler to exchange the auth code

	http.HandleFunc("/authorization-code/callback", handleExchange)

	log.Fatal(http.ListenAndServe(":14687", nil))
}
