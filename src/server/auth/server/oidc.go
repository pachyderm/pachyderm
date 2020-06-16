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

	// rs := reflect.ValueOf(o.Provider).Elem()
	// rf := rs.FieldByName("userInfoURL")
	// // rf can't be read or set.
	// rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()

	// userInfoURL := rf.String()
	// // userInfoURL = strings.Replace(userInfoURL, "172.17.0.2", "localhost", 1)

	// fmt.Println(userInfoURL)
	// request, err := http.NewRequest("GET", userInfoURL, nil)
	// if err != nil {
	// 	// fmt.Fprintf(w, "oidc: create GET request: %v", err)
	// 	return "", err
	// }
	// // tok.SetAuthHeader(request)
	// request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", oauthToken))
	// // request.Header.Write(w)

	// client := &http.Client{
	// 	CheckRedirect: func(req *http.Request, via []*http.Request) error {
	// 		return http.ErrUseLastResponse
	// 	},
	// }

	// // curl -v -X GET -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICJGMm1RR0VqdkFBSlctNnRlMF9FMm1BWUUxMWdyY21oNEdaeTNwTHVIVWpjIn0.eyJleHAiOjE1OTIwOTQzNDQsImlhdCI6MTU5MjA5NDA0NCwiYXV0aF90aW1lIjoxNTkyMDg3ODc1LCJqdGkiOiI1MzM5YjFjNi1lZmM0LTQxODYtYWM2Yi01MTk2OWU0ODQ4OGEiLCJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjgwODAvYXV0aC9yZWFsbXMvYWRlbGUtdGVzdGluZyIsImF1ZCI6ImFjY291bnQiLCJzdWIiOiIwZWE4ZjQzNi1lYTQ3LTQ3N2YtYjAxOC0zNzQxNzliZWY4ZjMiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJwYWNoeWRlcm0iLCJub25jZSI6InRlc3RpbmciLCJzZXNzaW9uX3N0YXRlIjoiZjc1YjZmNGYtMjA1Yy00YjQzLTgxZTktZjQwZTg3ZWY4YTAxIiwiYWNyIjoiMCIsImFsbG93ZWQtb3JpZ2lucyI6WyJodHRwOi8vbG9jYWxob3N0OjE0Njg3Il0sInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJvZmZsaW5lX2FjY2VzcyIsInVtYV9hdXRob3JpemF0aW9uIl19LCJyZXNvdXJjZV9hY2Nlc3MiOnsiYWNjb3VudCI6eyJyb2xlcyI6WyJtYW5hZ2UtYWNjb3VudCIsIm1hbmFnZS1hY2NvdW50LWxpbmtzIiwidmlldy1wcm9maWxlIl19fSwic2NvcGUiOiJvcGVuaWQgcHJvZmlsZSBlbWFpbCIsImVtYWlsX3ZlcmlmaWVkIjpmYWxzZSwibmFtZSI6IkFkZWxlIExvcGV6IiwicHJlZmVycmVkX3VzZXJuYW1lIjoiYWRlbGUiLCJnaXZlbl9uYW1lIjoiQWRlbGUiLCJmYW1pbHlfbmFtZSI6IkxvcGV6IiwiZW1haWwiOiJhZGVsZUBwYWNoeWRlcm0uaW8ifQ.KAXNfM_Po22LWVr12uL4j2bb95-AXjExOHWGgcHgyiiVJPUfbAlpUhAhEPmPYXf-ZJuRYWVOsJ2SEYYTN0qVsAZUiNcx6gAThG9M8xM3RgGmW_ZLyiyERc1a-2NY4xsOozlSVuFrKtR5KQ2hR4xoOG0XKHaEXrBhlGbCNHy8j0f49kVzCWG5NDeTTBFKzCzymDv9GDpT407SSUFlx2KHjdPNiu-qIG9ObTe6g2fguF5VCaTl4_zRoxZSatt1TgDrdO92izYcfC5QUKHQohhExXaC-ATnf-CBU8oT9TzdOy8aAi6MGDGGeuARBsp-opfuZYRgvcScK1pRKamzOSyDJg" http://172.17.0.2:8080/auth/realms/adele-testing/protocol/openid-connect/userinfo

	// resp, err := client.Do(request)

	// fmt.Println("did it worked?", resp.Status)

	// type claimEmail struct {
	// 	Email string
	// }

	// token, err := jwt.ParseWithClaims(oauthToken, func(token *jwt.Token) (interface{}, error) {
	// 	// since we only use the one private key to sign the tokens,
	// 	// we also only use its public counter part to verify
	// 	return "", nil
	// })
	// if err != nil {
	// 	return "", err
	// }

	// spew.Dump(token.Claims)
	// claims := token.Claims.(*claimEmail)
	// fmt.Println(claims.Email)

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

	var verifier = sp.Provider.Verifier(&oidc.Config{ClientID: conf.ClientID})

	// Extract the ID Token from OAuth2 token.
	rawIDToken, ok := tok.Extra("id_token").(string)
	if !ok {
		fmt.Println("missing token")
	}

	// Parse and verify ID Token payload.
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		log.Fatal("did not verify token", err)
	}

	spew.Dump(idToken)
	// Extract custom claims
	var claims struct {
		Email    string `json:"email"`
		Verified bool   `json:"email_verified"`
	}
	if err := idToken.Claims(&claims); err != nil {
		log.Fatal("could not parse claims", err)
	}

	fmt.Println(claims)

	tok.TokenType = "Bearer"
	// Use the token passed in as our token source
	ts := oauth2.StaticTokenSource(tok)
	newtok, err := ts.Token()
	if err != nil {
		panic("no token")
	}
	if newtok.AccessToken != tok.AccessToken {
		panic("not equal")
	}

	// rs := reflect.ValueOf(sp.Provider).Elem()
	// rf := rs.FieldByName("userInfoURL")
	// // rf can't be read or set.
	// rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()

	// userInfoURL := rf.String()
	// // userInfoURL = strings.Replace(userInfoURL, "172.17.0.2", "localhost", 1)

	// fmt.Println(userInfoURL)
	// request, err := http.NewRequest("GET", userInfoURL, nil)
	// if err != nil {
	// 	fmt.Fprintf(w, "oidc: create GET request: %v", err)
	// }
	// // tok.SetAuthHeader(request)
	// request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", tok.AccessToken))
	// request.Header.Write(w)

	// client := &http.Client{
	// 	CheckRedirect: func(req *http.Request, via []*http.Request) error {
	// 		return http.ErrUseLastResponse
	// 	},
	// }

	// // curl -v -X GET -H "Authorization: Bearer " http://localhost:8080/auth/realms/adele-testing/protocol/openid-connect/userinfo

	// resp, err := client.Do(request)

	// apt := exec.Command("Bash")

	// err = apt.Run()
	// if err != nil {
	// 	fmt.Fprintf(w, "Apt error: %v", err)
	// }
	// _, err = apt.Output()
	// if err != nil {
	// 	fmt.Fprintf(w, "Out error: %v", err)
	// }

	// curl := exec.Command("ls", "-v", "-H", fmt.Sprintf("\"Authorization: Bearer %v\"", tok.AccessToken), userInfoURL)

	// err = curl.Run()
	// if err != nil {
	// 	fmt.Fprintf(w, "Run error: %v", err)
	// }
	// out, err := curl.Output()
	// if err != nil {
	// 	fmt.Fprintf(w, "Out error: %v", err)
	// }

	// fmt.Fprintf(w, "\n%v\n", out)

	// if err != nil {
	// 	fmt.Fprintf(w, "Oops: %v", err)
	// }
	// fmt.Fprintf(w, "Yay: %v", resp.Status)

	// userInfo, err := sp.Provider.UserInfo(context.Background(), ts)
	// if err != nil {
	// 	fmt.Fprintf(w, "Oh no: %v", err)
	// }

	// spew.Dump(userInfo)
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
