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
	"strings"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/auth"
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
	detLoginPath       = "/api/v1/auth/login" // DNJ TODO don't hardcorde namespace/port? - need to rotate port for parallel?
	detUserPath        = "/api/v1/users"
	detWorkspacePath   = "/api/v1/workspaces"
	detNewUserPassword = "test-password"
)

type DeterminedUser struct {
	Id          int    `json:"id"`
	Username    string `json:"username"`
	Admin       bool   `json:"admin"`
	Active      bool   `json:"active"`
	DisplayName string `json:"displayName"`
}
type DeterminedUserBody struct {
	User     *DeterminedUser `json:"user"`
	Password string          `json:"password"`
	IsHashed bool            `json:"isHashed"`
}
type DeterminedUserList struct {
	Users *[]DeterminedUser `json:"users"`
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
	// log in and create a non-admin user with the kilgore email from pachyderm
	detUrl := determinedBaseUrl(t, ns)
	authToken := determinedLogin(t, *detUrl, "admin", "")
	detUser := determinedCreateUser(t, *detUrl, authToken)
	require.Equal(t, testutil.DexMockConnectorEmail, detUser.Username, "The new user has the same name as dex user")

	repoName := "images"
	pipelineName := "edges"
	workspaceName := "pach-test-workspace"
	// log in as non-admin test user to make and use the new workspace
	userToken := determinedLogin(t, *detUrl, detUser.Username, detNewUserPassword)
	determinedCreateWorkspace(t, *detUrl, userToken, workspaceName)
	previous := determinedGetUsers(t, *detUrl, userToken)
	// create repo and pipeline that should make the determined service user
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
				Workspaces: []string{workspaceName},
			},
		},
	)
	require.NoError(t, err)
	current := determinedGetUsers(t, *detUrl, userToken)
	require.Equal(t, len(*previous.Users)+1, len(*current.Users), "the new pipeline has created an additional service user in Determined")
}

func determinedLogin(t testing.TB, detUrl url.URL, username string, password string) string {
	detUrl.Path = detLoginPath
	req, err := http.NewRequest(
		"POST",
		detUrl.String(),
		strings.NewReader(fmt.Sprintf(`{"username":"%s","password":"%s"}`, username, password)),
	)
	require.NoError(t, err, "Creating Determined login request")

	body := doDeterminedRequest(t, req)

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
		},
		Password: detNewUserPassword,
		IsHashed: false,
	}
	userJson, err := json.Marshal(userReq)
	require.NoError(t, err, "Marshal determined user json")
	detUrl.Path = detUserPath
	req, err := http.NewRequest("POST", detUrl.String(), bytes.NewReader(userJson))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authToken))
	require.NoError(t, err, "Creating Determined create user request")

	body := doDeterminedRequest(t, req)

	userResponse := &DeterminedUserBody{}
	err = json.Unmarshal(body, userResponse)
	require.NoError(t, err, "Parsing Determined user create", string(body))
	return userResponse.User
}

func determinedGetUsers(t testing.TB, detUrl url.URL, authToken string) *DeterminedUserList {
	detUrl.Path = detUserPath
	req, err := http.NewRequest("GET", detUrl.String(), strings.NewReader("{}"))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authToken))
	require.NoError(t, err, "Creating Determined get users request")

	body := doDeterminedRequest(t, req)

	userResponse := &DeterminedUserList{}
	err = json.Unmarshal(body, userResponse)
	require.NoError(t, err, "Parsing Determined user list", string(body))
	return userResponse
}

func doDeterminedRequest(t testing.TB, req *http.Request) []byte {
	hc := testutil.NewLoggingHTTPClient(t)
	hc.Timeout = 15 * time.Second
	var resp *http.Response
	var err error
	require.NoErrorWithinTRetryConstant(t, 120*time.Second, func() error {
		resp, err = hc.Do(req)
		return errors.EnsureStack(err)
	}, 5*time.Second, "Attempting to make determined request")
	require.Equal(t, 200, resp.StatusCode, "Checking response code for Determined request")
	responseBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Reading Determined API response")
	return responseBody
}

func determinedCreateWorkspace(t testing.TB, detUrl url.URL, authToken string, workspace string) {
	workspaceReq := struct {
		Name string `json:"name"`
	}{
		Name: workspace,
	}
	workspaceJson, err := json.Marshal(workspaceReq)
	require.NoError(t, err, "Marshal determined workspace json")
	detUrl.Path = detWorkspacePath
	req, err := http.NewRequest("POST", detUrl.String(), bytes.NewReader(workspaceJson))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authToken))
	require.NoError(t, err, "Creating Determined create workspace request")
	_ = doDeterminedRequest(t, req)
}

func determinedBaseUrl(t testing.TB, namespace string) *url.URL {
	ctx := context.Background()
	kube := testutil.GetKubeClient(t)
	service, err := kube.CoreV1().Services(namespace).Get(ctx, fmt.Sprintf("determined-master-service-%s", namespace), v1.GetOptions{}) // DNJ TODO - should this be in minikubetestenv?
	detPort := service.Spec.Ports[0].NodePort
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
