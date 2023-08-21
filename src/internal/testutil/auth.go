package testutil

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

const (
	// RootToken is the hard-coded admin token used on all activated test clusters
	RootToken = "iamroot"
)

func TSProtoOrDie(t testing.TB, ts time.Time) *timestamppb.Timestamp {
	return timestamppb.New(ts)
}

func activateAuthHelper(tb testing.TB, client *client.APIClient, port ...string) {
	client.SetAuthToken(RootToken)

	ActivateEnterprise(tb, client, port...)

	_, err := client.Activate(client.Ctx(),
		&auth.ActivateRequest{RootToken: RootToken},
	)
	if err != nil && !strings.HasSuffix(err.Error(), "already activated") {
		tb.Fatalf("could not activate auth service: %v", err.Error())
	}
	require.NoError(tb, config.WritePachTokenToConfig(RootToken, false))

	// Activate auth for PPS
	client = client.WithCtx(context.Background())
	client.SetAuthToken(RootToken)
	_, err = client.PfsAPIClient.ActivateAuth(client.Ctx(), &pfs.ActivateAuthRequest{})
	require.NoError(tb, err)
	_, err = client.PpsAPIClient.ActivateAuth(client.Ctx(), &pps.ActivateAuthRequest{})
	require.NoError(tb, err)
}

// ActivateAuthClient activates the auth service in the test cluster, if it isn't already enabled
func ActivateAuthClient(tb testing.TB, c *client.APIClient, port ...string) {
	tb.Helper()
	activateAuthHelper(tb, c, port...)
}

// creates a new authenticated pach client, without re-activating
func AuthenticateClient(tb testing.TB, c *client.APIClient, subject string) *client.APIClient {
	tb.Helper()
	rootClient := UnauthenticatedPachClient(tb, c)
	rootClient.SetAuthToken(RootToken)
	if subject == auth.RootUser {
		return rootClient
	}
	token, err := rootClient.GetRobotToken(rootClient.Ctx(), &auth.GetRobotTokenRequest{Robot: subject})
	require.NoError(tb, err)
	client := UnauthenticatedPachClient(tb, c)
	client.SetAuthToken(token.Token)
	return client
}

func AuthenticatedPachClient(tb testing.TB, c *client.APIClient, subject string, port ...string) *client.APIClient {
	tb.Helper()
	activateAuthHelper(tb, c, port...)
	return AuthenticateClient(tb, c, subject)
}

// GetUnauthenticatedPachClient returns a copy of the testing pach client with no auth token
func UnauthenticatedPachClient(tb testing.TB, c *client.APIClient) *client.APIClient {
	tb.Helper()
	client := c.WithCtx(context.Background())
	client.SetAuthToken("")
	return client
}

func RepoInfoToName(repoInfo interface{}) interface{} {
	return repoInfo.(*pfs.RepoInfo).Repo.Name
}

func User(email string) string {
	return auth.UserPrefix + email
}

func Pl(projectName, pipelineName string) string {
	return fmt.Sprintf("%s%s/%s", auth.PipelinePrefix, projectName, pipelineName)
}

func Robot(robot string) string {
	return auth.RobotPrefix + robot
}

func BuildClusterBindings(s ...string) *auth.RoleBinding {
	return BuildBindings(append(s,
		auth.AllClusterUsersSubject, auth.ProjectCreatorRole,
		auth.RootUser, auth.ClusterAdminRole,
		auth.InternalPrefix+"auth-server", auth.ClusterAdminRole,
	)...)
}

func BuildBindings(s ...string) *auth.RoleBinding {
	var b auth.RoleBinding
	b.Entries = make(map[string]*auth.Roles)
	for i := 0; i < len(s); i += 2 {
		if _, ok := b.Entries[s[i]]; !ok {
			b.Entries[s[i]] = &auth.Roles{Roles: make(map[string]bool)}
		}
		b.Entries[s[i]].Roles[s[i+1]] = true
	}
	return &b
}

func GetRepoRoleBinding(t *testing.T, c *client.APIClient, projectName, repoName string) *auth.RoleBinding {
	t.Helper()
	resp, err := c.GetRepoRoleBinding(projectName, repoName)
	require.NoError(t, err)
	return resp
}

func GetProjectRoleBinding(t *testing.T, c *client.APIClient, project string) *auth.RoleBinding {
	t.Helper()
	resp, err := c.GetProjectRoleBinding(project)
	require.NoError(t, err)
	return resp
}

// CommitCnt uses 'c' to get the number of commits made to the repo 'repo'
func CommitCnt(t *testing.T, c *client.APIClient, repo *pfs.Repo) int {
	t.Helper()
	commitList, err := c.ListCommitByRepo(repo)
	require.NoError(t, err)
	return len(commitList)
}

// PipelineNames returns the names of all pipelines that 'c' gets from
// ListPipeline in the specified project.
func PipelineNames(t *testing.T, c *client.APIClient, project string) []string {
	t.Helper()
	ps, err := c.ListPipeline(false)
	require.NoError(t, err)
	var result []string
	if project == "" {
		project = pfs.DefaultProjectName
	}
	for _, p := range ps {
		if project == p.GetPipeline().GetProject().GetName() {
			result = append(result, p.GetPipeline().GetName())
		}
	}
	return result
}

func Group(group string) string {
	return auth.GroupPrefix + group
}

func RandomRobot(t *testing.T, c *client.APIClient, name string) (string, *client.APIClient) {
	name = Robot(UniqueString(name))
	return name, AuthenticateClient(t, c, name)
}
