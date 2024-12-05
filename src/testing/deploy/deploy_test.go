//go:build k8s

package main

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

var globalValueOverrides map[string]string = make(map[string]string)

func mockIDPLogin(t testing.TB, c *client.APIClient) {
	require.NoErrorWithinTRetryConstant(t, 90*time.Second, func() (retErr error) {
		// login using mock IDP admin
		hc := &http.Client{Timeout: 15 * time.Second}
		c.SetAuthToken("")
		loginInfo, err := c.GetOIDCLogin(c.Ctx(), &auth.GetOIDCLoginRequest{})
		if err != nil {
			return errors.EnsureStack(err)
		}
		state := loginInfo.State
		// Get the initial URL from the grpc, which should point to the dex login page
		getResp, err := hc.Get(loginInfo.LoginUrl)
		if err != nil {
			return errors.EnsureStack(err)
		}
		defer errors.Close(&retErr, getResp.Body, "close body")
		if got, want := http.StatusOK, getResp.StatusCode; got != want {
			testutil.LogHttpResponse(t, getResp, "mock login get")
			return errors.Errorf("retrieve mock login page satus code: got %v want %v", got, want)
		}

		vals := make(url.Values)
		vals.Add("login", "admin")
		vals.Add("password", "password")
		postResp, err := hc.PostForm(getResp.Request.URL.String(), vals)
		if err != nil {
			return errors.EnsureStack(err)
		}
		defer postResp.Body.Close()
		if got, want := http.StatusOK, postResp.StatusCode; got != want {
			testutil.LogHttpResponse(t, postResp, "mock login post")
			return errors.Errorf("POST to perform mock login got: %v, want: %v", got, want)
		}
		postBody, err := io.ReadAll(postResp.Body)
		if err != nil {
			return errors.Errorf("Could not read login form POST response: %v", err.Error())
		}
		// There is a login failure case in which the response returned is a redirect back to the login page that returns 200, but does not log in
		if got, want := string(postBody), "You are now logged in"; !strings.HasPrefix(got, want) {
			return errors.Errorf("response body from mock IDP login form got: %v, want: %v", postBody, want)
		}

		authResp, err := c.AuthAPIClient.Authenticate(c.Ctx(), &auth.AuthenticateRequest{OidcState: state})
		if err != nil {
			return errors.EnsureStack(err)
		}
		c.SetAuthToken(authResp.PachToken)
		whoami, err := c.AuthAPIClient.WhoAmI(c.Ctx(), &auth.WhoAmIRequest{})
		if err != nil {
			return errors.EnsureStack(err)
		}
		expectedUsername := "user:" + testutil.DexMockConnectorEmail
		if expectedUsername != whoami.Username {
			return errors.Errorf("username after mock IDP login got: %v, want: %v ", whoami.Username, expectedUsername)
		}
		return nil
	}, 5*time.Second, "Attempting login through mock IDP")
}
