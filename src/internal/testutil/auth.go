package testutil

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

const (
	// RootToken is the hard-coded admin token used on all activated test clusters
	RootToken = "iamroot"
)

func TSProtoOrDie(t testing.TB, ts time.Time) *types.Timestamp {
	proto, err := types.TimestampProto(ts)
	require.NoError(t, err)
	return proto
}

func activateAuthHelper(tb testing.TB, client *client.APIClient) {
	client.SetAuthToken(RootToken)

	ActivateEnterprise(tb, client)

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
func ActivateAuthClient(tb testing.TB, c *client.APIClient) {
	tb.Helper()
	activateAuthHelper(tb, c)
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

func AuthenticatedPachClient(tb testing.TB, c *client.APIClient, subject string) *client.APIClient {
	tb.Helper()
	activateAuthHelper(tb, c)
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

func Pl(pipeline string) string {
	return auth.PipelinePrefix + pipeline
}

func Robot(robot string) string {
	return auth.RobotPrefix + robot
}

func BuildClusterBindings(s ...string) *auth.RoleBinding {
	return BuildBindings(append(s,
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

func GetRepoRoleBinding(t *testing.T, c *client.APIClient, repo string) *auth.RoleBinding {
	t.Helper()
	resp, err := c.GetRepoRoleBinding(repo)
	require.NoError(t, err)
	return resp
}

// CommitCnt uses 'c' to get the number of commits made to the repo 'repo'
func CommitCnt(t *testing.T, c *client.APIClient, repo string) int {
	t.Helper()
	commitList, err := c.ListCommitByRepo(client.NewRepo(repo))
	require.NoError(t, err)
	return len(commitList)
}

// PipelineNames returns the names of all pipelines that 'c' gets from
// ListPipeline
func PipelineNames(t *testing.T, c *client.APIClient) []string {
	t.Helper()
	ps, err := c.ListPipeline(false)
	require.NoError(t, err)
	result := make([]string, len(ps))
	for i, p := range ps {
		result[i] = p.Pipeline.Name
	}
	return result
}

func Group(group string) string {
	return auth.GroupPrefix + group
}
