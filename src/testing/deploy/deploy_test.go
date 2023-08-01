//go:build k8s

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/net"
)

const (
	detLoginPath = "/api/v1/auth/login" // DNJ TODO don't hardcorde namespace/port? - need to rotate for parallel
	detUserPath  = "/api/v1/users"
)

var valueOverrides map[string]string = make(map[string]string)

type DeterminedUser struct {
	Id          int    `json:"id"`
	Username    string `json:"username"`
	Admin       bool   `json:"admin"`
	Active      bool   `json:"active"`
	DisplayName string `json:"displayName"`
	Password    string `json:"password"`
	IsHashed    bool   `json:"isHashed"`
}
type DeterminedUserBody struct {
	User *DeterminedUser `json:"user"`
}

func TestInstallAndUpgradeEnterpriseWithEnv(t *testing.T) {
	t.Parallel()
	ns, portOffset := minikubetestenv.ClaimCluster(t)
	k := testutil.GetKubeClient(t)
	opts := &minikubetestenv.DeployOpts{
		AuthUser:   auth.RootUser,
		Enterprise: true,
		PortOffset: portOffset,
		Determined: true,
	}
	valueOverrides["pachd.replicas"] = "1"
	opts.ValueOverrides = valueOverrides
	// Test Install
	minikubetestenv.PutNamespace(t, ns)
	c := minikubetestenv.InstallRelease(t, context.Background(), ns, k, opts)
	whoami, err := c.AuthAPIClient.WhoAmI(c.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, auth.RootUser, whoami.Username)
	c.SetAuthToken("")
	mockIDPLogin(t, c)
	// Test Upgrade
	opts.CleanupAfter = false
	// set new root token via env
	opts.AuthUser = ""
	token := "new-root-token"
	opts.ValueOverrides = valueOverrides
	opts.ValueOverrides["pachd.rootToken"] = token
	// add config file with trusted peers & new clients
	opts.ValuesFiles = []string{createAdditionalClientsFile(t), createTrustedPeersFile(t)}
	// apply upgrade
	c = minikubetestenv.UpgradeRelease(t, context.Background(), ns, k, opts)
	c.SetAuthToken(token)
	whoami, err = c.AuthAPIClient.WhoAmI(c.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, auth.RootUser, whoami.Username)
	// old token should no longer work
	c.SetAuthToken(testutil.RootToken)
	_, err = c.AuthAPIClient.WhoAmI(c.Ctx(), &auth.WhoAmIRequest{})
	require.YesError(t, err)
	c.SetAuthToken("")
	mockIDPLogin(t, c)
	// assert new trusted peer and client
	resp, err := c.IdentityAPIClient.GetOIDCClient(c.Ctx(), &identity.GetOIDCClientRequest{Id: "pachd"})
	require.NoError(t, err)
	require.EqualOneOf(t, resp.Client.TrustedPeers, "example-app")
	require.EqualOneOf(t, resp.Client.TrustedPeers, "determined-local")
	_, err = c.IdentityAPIClient.GetOIDCClient(c.Ctx(), &identity.GetOIDCClientRequest{Id: "example-app"})
	require.NoError(t, err)
	_, err = c.IdentityAPIClient.GetOIDCClient(c.Ctx(), &identity.GetOIDCClientRequest{Id: "determined-local"})
	require.NoError(t, err)
}

func TestEnterpriseServerMember(t *testing.T) {
	t.Parallel()
	ns, portOffset := minikubetestenv.ClaimCluster(t)
	k := testutil.GetKubeClient(t)
	minikubetestenv.PutNamespace(t, "enterprise")
	valueOverrides["pachd.replicas"] = "2"
	ec := minikubetestenv.InstallRelease(t, context.Background(), "enterprise", k, &minikubetestenv.DeployOpts{
		AuthUser:         auth.RootUser,
		EnterpriseServer: true,
		CleanupAfter:     true,
		ValueOverrides:   valueOverrides,
	})
	whoami, err := ec.AuthAPIClient.WhoAmI(ec.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, auth.RootUser, whoami.Username)
	mockIDPLogin(t, ec)
	minikubetestenv.PutNamespace(t, ns)
	c := minikubetestenv.InstallRelease(t, context.Background(), ns, k, &minikubetestenv.DeployOpts{
		AuthUser:         auth.RootUser,
		EnterpriseMember: true,
		Enterprise:       true,
		PortOffset:       portOffset,
		CleanupAfter:     true,
		ValueOverrides:   valueOverrides,
	})
	whoami, err = c.AuthAPIClient.WhoAmI(c.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, auth.RootUser, whoami.Username)
	c.SetAuthToken("")
	loginInfo, err := c.GetOIDCLogin(c.Ctx(), &auth.GetOIDCLoginRequest{})
	require.NoError(t, err)
	require.True(t, strings.Contains(loginInfo.LoginUrl, ":31658"))
	mockIDPLogin(t, c)
}

func TestDeterminedInstallAndIntegration(t *testing.T) {
	t.Parallel()
	ns, portOffset := minikubetestenv.ClaimCluster(t)
	k := testutil.GetKubeClient(t)
	opts := &minikubetestenv.DeployOpts{
		AuthUser:   auth.RootUser,
		Enterprise: true,
		PortOffset: portOffset,
		Determined: true,
	}
	valueOverrides["pachd.replicas"] = "1"
	opts.ValueOverrides = valueOverrides
	minikubetestenv.PutNamespace(t, ns)
	c := minikubetestenv.InstallRelease(t, context.Background(), ns, k, opts)
	whoami, err := c.AuthAPIClient.WhoAmI(c.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, auth.RootUser, whoami.Username)
	c.SetAuthToken("")
	mockIDPLogin(t, c)
	time.Sleep(30 * time.Second) // DNJ TODO - wait for determined in deploy
	detUrl := determinedBaseUrl(t, ns)
	authToken := determinedLogin(t, *detUrl)
	detUser := determinedCreateUser(t, *detUrl, authToken)
	require.Equal(t, testutil.DexMockConnectorEmail, detUser.Username)
	// create repo and pipeline that should make the determined service user
	repoName := "images"
	pipelineName := "edges"
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repoName))
	_, err = c.PpsAPIClient.CreatePipeline(
		c.Ctx(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipelineName),
			Transform: &pps.Transform{
				Image: repoName,
				Cmd:   []string{"python3", "/edges.py"},
				Stdin: nil,
			},
			ParallelismSpec: nil,
			Input:           &pps.Input{Pfs: &pps.PFSInput{Glob: "/", Repo: repoName}},
			OutputBranch:    "master",
			Update:          false,
			Determined: &pps.Determined{
				Workspaces: []string{"pach-test-workspace"},
			},
		},
	)
	require.NoError(t, err)
}

func mockIDPLogin(t testing.TB, c *client.APIClient) {
	require.NoErrorWithinTRetryConstant(t, 60*time.Second, func() error {
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
		defer getResp.Body.Close()
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

func determinedLogin(t testing.TB, detUrl url.URL) string {
	detUrl.Path = detLoginPath
	req, err := http.NewRequest("POST", detUrl.String(), strings.NewReader(`{"username":"admin","password":""}`))
	require.NoError(t, err, "Creating Determined login request")

	hc := &http.Client{Timeout: 15 * time.Second}
	var resp *http.Response
	require.NoErrorWithinTRetryConstant(t, 60*time.Second, func() error {
		resp, err = hc.Do(req)
		return errors.EnsureStack(err)
	}, 5*time.Second, "Attempting to log into determined")
	require.Equal(t, 200, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Reading Determined login")
	authToken := struct {
		Token string
	}{}
	err = json.Unmarshal(body, &authToken)
	require.NoError(t, err, "Parsing Determined login")
	require.NotEqual(t, "", authToken.Token)
	return authToken.Token
}

func determinedCreateUser(t testing.TB, detUrl url.URL, authToken string) *DeterminedUser {
	userReq := DeterminedUserBody{
		User: &DeterminedUser{
			Username:    testutil.DexMockConnectorEmail,
			Admin:       false,
			Active:      true,
			DisplayName: testutil.DexMockConnectorEmail,
			Password:    "password",
			IsHashed:    false,
		},
	}
	userJson, err := json.Marshal(userReq)
	require.NoError(t, err, "Marshal determined user json")
	detUrl.Path = detUserPath
	req, err := http.NewRequest("POST", detUrl.String(), bytes.NewReader(userJson)) // DNJ TODO maybe refactor and combine with login call marshal/unmarshal instead of repeating code
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authToken))
	require.NoError(t, err, "Creating Determined user request")

	hc := &http.Client{Timeout: 15 * time.Second}
	var resp *http.Response
	require.NoErrorWithinTRetryConstant(t, 60*time.Second, func() error {
		resp, err = hc.Do(req)
		return errors.EnsureStack(err)
	}, 5*time.Second, "Attempting to create determined user")
	require.Equal(t, 200, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Reading Determined user creation response")
	userResponse := &DeterminedUserBody{}
	t.Logf("DNJ TODO %v", string(body))
	err = json.Unmarshal(body, userResponse)
	require.NoError(t, err, "Parsing Determined user create")
	return userResponse.User
}

func determinedBaseUrl(t testing.TB, namespace string) *url.URL {
	ctx := context.Background()
	kube := testutil.GetKubeClient(t)
	service, err := kube.CoreV1().Services(namespace).Get(ctx, fmt.Sprintf("determined-master-service-%s", namespace), v1.GetOptions{}) // DNJ TODO - should this be in minikubetestenv?
	detPort := service.Spec.Ports[0].NodePort                                                                                           // DNJ TODO - this port in values needs to be dynamic
	require.NoError(t, err, "Fininding Determined service")
	node, err := kube.CoreV1().Nodes().Get(ctx, "minikube", v1.GetOptions{})
	require.NoError(t, err, "Fininding node for Determined")
	var detHost string
	for _, addr := range node.Status.Addresses {
		if addr.Type == "InternalIP" {
			detHost = addr.Address
		}
	}
	detUrl := net.FormatURL("http", detHost, int(detPort), "")
	return detUrl
}

func createTrustedPeersFile(t testing.TB) string {
	data := []byte(`pachd:
  additionalTrustedPeers:
    - example-app
`)
	tf, err := os.CreateTemp("", "pachyderm-trusted-peers-*.yaml")
	require.NoError(t, err)
	_, err = tf.Write(data)
	require.NoError(t, err)
	return tf.Name()
}

func createAdditionalClientsFile(t testing.TB) string {
	data := []byte(`oidc:
  additionalClients:
    - id: example-app
      secret: example-app-secret
      name: 'Example App'
      redirectURIs:
      - 'http://127.0.0.1:5555/callback'
`)
	tf, err := os.CreateTemp("", "pachyderm-additional-clients-*.yaml")
	require.NoError(t, err)
	_, err = tf.Write(data)
	require.NoError(t, err)
	return tf.Name()
}
