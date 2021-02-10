// admin_test.go tests various features related to pachyderm's auth admins.
// Because the cluster has one global set of admins, these tests can't be run in
// parallel

package server

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/license"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/version"

	"github.com/gogo/protobuf/types"
	"golang.org/x/net/context"
)

const secsInYear = 365 * 24 * 60 * 60

func RepoInfoToName(repoInfo interface{}) interface{} {
	return repoInfo.(*pfs.RepoInfo).Repo.Name
}

func TSProtoOrDie(t *testing.T, ts time.Time) *types.Timestamp {
	proto, err := types.TimestampProto(ts)
	require.NoError(t, err)
	return proto
}

func user(email string) string {
	return auth.UserPrefix + email
}

func pl(pipeline string) string {
	return auth.PipelinePrefix + pipeline
}

func robot(robot string) string {
	return auth.RobotPrefix + robot
}

func buildClusterBindings(s ...string) *auth.RoleBinding {
	return buildBindings(append(s,
		auth.RootUser, auth.ClusterAdminRole,
		auth.PpsUser, auth.ClusterAdminRole)...)
}

func buildBindings(s ...string) *auth.RoleBinding {
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

// TestActivate tests the Activate API (in particular, verifying
// that Activate() also authenticates). Even though GetClient also activates
// auth, this makes sure the code path is exercised (as auth may already be
// active when the test starts)
func TestActivate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)
	// Get anonymous client (this will activate auth, which is about to be
	// deactivated, but it also activates Pacyderm enterprise, which is needed for
	// this test to pass)
	rootClient := tu.GetAuthenticatedPachClient(t, auth.RootUser)

	_, err := rootClient.Deactivate(rootClient.Ctx(), &auth.DeactivateRequest{})
	require.NoError(t, err)
	resp, err := rootClient.AuthAPIClient.Activate(context.Background(), &auth.ActivateRequest{})
	require.NoError(t, err)
	rootClient.SetAuthToken(resp.PachToken)
	defer rootClient.Deactivate(rootClient.Ctx(), &auth.DeactivateRequest{})

	// Check that the token 'c' received from pachd authenticates them as "pach:root"
	who, err := rootClient.WhoAmI(rootClient.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, auth.RootUser, who.Username)

	bindings, err := rootClient.GetClusterRoleBinding()
	require.NoError(t, err)
	require.Equal(t, buildClusterBindings(), bindings)
}

// TestActivateKnownToken tests activating auth with a known token.
// This should always authenticate the user as `pach:root` and give them
// super admin status.
func TestActivateKnownToken(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)
	// Get anonymous client (this will activate auth, which is about to be
	// deactivated, but it also activates Pacyderm enterprise, which is needed for
	// this test to pass)
	rootClient := tu.GetAuthenticatedPachClient(t, auth.RootUser)

	_, err := rootClient.Deactivate(rootClient.Ctx(), &auth.DeactivateRequest{})
	require.NoError(t, err)
	resp, err := rootClient.AuthAPIClient.Activate(context.Background(), &auth.ActivateRequest{RootToken: tu.RootToken})
	require.NoError(t, err)
	require.Equal(t, resp.PachToken, tu.RootToken)

	rootClient.SetAuthToken(tu.RootToken)

	// Check that the token 'c' received from pachd authenticates them as "pach:root"
	who, err := rootClient.WhoAmI(rootClient.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, auth.RootUser, who.Username)

	bindings, err := rootClient.GetClusterRoleBinding()
	require.NoError(t, err)
	require.Equal(t, buildClusterBindings(), bindings)
}

// TestSuperAdminRWO tests adding and removing cluster super admins, as well as super admins
// reading, writing, and moderating (owning) all repos in the cluster.
func TestSuperAdminRWO(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)
	alice, bob := robot(tu.UniqueString("alice")), robot(tu.UniqueString("bob"))
	aliceClient, bobClient := tu.GetAuthenticatedPachClient(t, alice), tu.GetAuthenticatedPachClient(t, bob)
	rootClient := tu.GetAuthenticatedPachClient(t, auth.RootUser)

	// The initial set of admins is just the user "admin"
	bindings, err := aliceClient.GetClusterRoleBinding()
	require.NoError(t, err)
	require.Equal(t, buildClusterBindings(), bindings)

	// alice creates a repo (that only she owns) and puts a file
	repo := tu.UniqueString("TestAdminRWO")
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.Equal(t, buildBindings(alice, auth.RepoOwnerRole), getRepoRoleBinding(t, aliceClient, repo))
	commit, err := aliceClient.StartCommit(repo, "master")
	require.NoError(t, err)
	err = aliceClient.PutFile(repo, commit.ID, "/file", strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.ID))

	// bob cannot read from the repo
	buf := &bytes.Buffer{}
	err = bobClient.GetFile(repo, "master", "/file", buf)
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())

	// bob cannot write to the repo
	_, err = bobClient.StartCommit(repo, "master")
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	// Note: we must pass aliceClient to CommitCnt, because it calls
	// ListCommit(repo), which requires the caller to have READER access to
	// 'repo', which bob does not have (but alice does)
	require.Equal(t, 1, CommitCnt(t, aliceClient, repo)) // check that no commits were created

	// bob can't update the ACL
	err = bobClient.ModifyRepoRoleBinding(repo, robot("carol"), []string{auth.RepoReaderRole})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.Equal(t, buildBindings(alice, auth.RepoOwnerRole), getRepoRoleBinding(t, aliceClient, repo)) // check that ACL wasn't updated

	// 'admin' makes bob a super admin
	require.NoError(t, rootClient.ModifyClusterRoleBinding(bob, []string{auth.ClusterAdminRole}))

	// wait until bob shows up in admin list
	bindings, err = aliceClient.GetClusterRoleBinding()
	require.NoError(t, err)
	require.Equal(t, buildClusterBindings(bob, auth.ClusterAdminRole), bindings)

	// now bob can read from the repo
	buf.Reset()
	require.NoError(t, bobClient.GetFile(repo, "master", "/file", buf))
	require.Matches(t, "test data", buf.String())

	// bob can write to the repo
	commit, err = bobClient.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, bobClient.FinishCommit(repo, commit.ID))
	require.Equal(t, 2, CommitCnt(t, aliceClient, repo)) // check that a new commit was created

	// bob can update the repo's ACL
	err = bobClient.ModifyRepoRoleBinding(repo, robot("carol"), []string{auth.RepoReaderRole})
	require.NoError(t, err)
	// check that ACL was updated
	require.Equal(t, buildBindings(alice, auth.RepoOwnerRole, robot("carol"), "repoReader"), getRepoRoleBinding(t, aliceClient, repo))

	// 'admin' revokes bob's admin status
	require.NoError(t, rootClient.ModifyClusterRoleBinding(bob, []string{}))

	// wait until bob is not in admin list
	bindings, err = aliceClient.GetClusterRoleBinding()
	require.NoError(t, err)
	require.Equal(t, buildClusterBindings(), bindings)

	// bob can no longer read from the repo
	buf.Reset()
	err = bobClient.GetFile(repo, "master", "/file", buf)
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())

	// bob cannot write to the repo
	_, err = bobClient.StartCommit(repo, "master")
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.Equal(t, 2, CommitCnt(t, aliceClient, repo)) // check that no commits were created

	// bob can't update the ACL
	err = bobClient.ModifyRepoRoleBinding(repo, robot("carol"), []string{auth.RepoWriterRole})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	// check that ACL wasn't updated
	require.Equal(t,
		buildBindings(alice, auth.RepoOwnerRole, robot("carol"), "repoReader"), getRepoRoleBinding(t, aliceClient, repo))
}

// TestFSAdminRWO tests adding and removing cluster FS admins, as well as FS admins
// reading, writing, and moderating (owning) all repos in the cluster.
func TestFSAdminRWO(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)
	alice, bob := robot(tu.UniqueString("alice")), robot(tu.UniqueString("bob"))
	aliceClient, bobClient := tu.GetAuthenticatedPachClient(t, alice), tu.GetAuthenticatedPachClient(t, bob)
	rootClient := tu.GetAuthenticatedPachClient(t, auth.RootUser)

	// The initial set of admins is just the user "admin"
	bindings, err := aliceClient.GetClusterRoleBinding()
	require.NoError(t, err)
	require.Equal(t, buildClusterBindings(), bindings)

	// alice creates a repo (that only she owns) and puts a file
	repo := tu.UniqueString("TestAdminRWO")
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.Equal(t, buildBindings(alice, auth.RepoOwnerRole), getRepoRoleBinding(t, aliceClient, repo))
	commit, err := aliceClient.StartCommit(repo, "master")
	require.NoError(t, err)
	err = aliceClient.PutFile(repo, commit.ID, "/file", strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.ID))

	// bob cannot read from the repo
	buf := &bytes.Buffer{}
	err = bobClient.GetFile(repo, "master", "/file", buf)
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())

	// bob cannot write to the repo
	_, err = bobClient.StartCommit(repo, "master")
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	// Note: we must pass aliceClient to CommitCnt, because it calls
	// ListCommit(repo), which requires the caller to have READER access to
	// 'repo', which bob does not have (but alice does)
	require.Equal(t, 1, CommitCnt(t, aliceClient, repo)) // check that no commits were created

	// bob can't update the ACL
	err = bobClient.ModifyRepoRoleBinding(repo, robot("carol"), []string{auth.RepoReaderRole})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.Equal(t, buildBindings(alice, auth.RepoOwnerRole), getRepoRoleBinding(t, aliceClient, repo)) // check that ACL wasn't updated

	// 'admin' makes bob an fs admin
	require.NoError(t, rootClient.ModifyClusterRoleBinding(bob, []string{auth.RepoOwnerRole}))

	// wait until bob shows up in admin list
	bindings, err = aliceClient.GetClusterRoleBinding()
	require.NoError(t, err)
	require.Equal(t, buildClusterBindings(bob, auth.RepoOwnerRole), bindings)

	// now bob can read from the repo
	buf.Reset()
	require.NoError(t, bobClient.GetFile(repo, "master", "/file", buf))
	require.Matches(t, "test data", buf.String())

	// bob can write to the repo
	commit, err = bobClient.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, bobClient.FinishCommit(repo, commit.ID))
	require.Equal(t, 2, CommitCnt(t, aliceClient, repo)) // check that a new commit was created

	// bob can update the repo's ACL
	err = bobClient.ModifyRepoRoleBinding(repo, robot("carol"), []string{auth.RepoReaderRole})
	require.NoError(t, err)
	// check that ACL was updated
	require.Equal(t, buildBindings(alice, auth.RepoOwnerRole, robot("carol"), "repoReader"), getRepoRoleBinding(t, aliceClient, repo))

	// 'admin' revokes bob's admin status
	require.NoError(t, rootClient.ModifyClusterRoleBinding(bob, []string{}))

	// wait until bob is not in admin list
	bindings, err = aliceClient.GetClusterRoleBinding()
	require.NoError(t, err)
	require.Equal(t, buildClusterBindings(), bindings)

	// bob can no longer read from the repo
	buf.Reset()
	err = bobClient.GetFile(repo, "master", "/file", buf)
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())

	// bob cannot write to the repo
	_, err = bobClient.StartCommit(repo, "master")
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.Equal(t, 2, CommitCnt(t, aliceClient, repo)) // check that no commits were created

	// bob can't update the ACL
	err = bobClient.ModifyRepoRoleBinding(repo, robot("carol"), []string{auth.RepoWriterRole})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	// check that ACL wasn't updated
	require.Equal(t, buildBindings(alice, auth.RepoOwnerRole, robot("carol"), "repoReader"), getRepoRoleBinding(t, aliceClient, repo))
}

// TestFSAdminFixBrokenRepo tests that an FS admin can modify the ACL of a repo even
// when the repo's ACL is empty (indicating that no user has explicit access to
// to the repo)
func TestFSAdminFixBrokenRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)
	alice, bob := robot(tu.UniqueString("alice")), robot(tu.UniqueString("bob"))
	aliceClient, bobClient := tu.GetAuthenticatedPachClient(t, alice), tu.GetAuthenticatedPachClient(t, bob)
	rootClient := tu.GetAuthenticatedPachClient(t, auth.RootUser)

	// alice creates a repo (that only she owns) and puts a file
	repo := tu.UniqueString("TestAdmin")
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.Equal(t, buildBindings(alice, auth.RepoOwnerRole), getRepoRoleBinding(t, aliceClient, repo))

	// 'admin' makes bob an FS admin
	require.NoError(t, rootClient.ModifyClusterRoleBinding(bob, []string{auth.RepoOwnerRole}))

	// wait until bob shows up in admin list
	bindings, err := aliceClient.GetClusterRoleBinding()
	require.NoError(t, err)
	require.Equal(t, buildClusterBindings(bob, auth.RepoOwnerRole), bindings)

	// admin deletes the repo's ACL
	require.NoError(t, rootClient.ModifyRepoRoleBinding(repo, alice, []string{}))

	// Check that the ACL is empty
	require.Nil(t, getRepoRoleBinding(t, rootClient, repo).Entries)

	// alice cannot write to the repo
	_, err = aliceClient.StartCommit(repo, "master")
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.Equal(t, 0, CommitCnt(t, rootClient, repo)) // check that no commits were created

	// bob, an FS admin, can update the ACL to put Alice back, even though reading the ACL
	// will fail
	require.NoError(t, bobClient.ModifyRepoRoleBinding(repo, alice, []string{auth.RepoOwnerRole}))
	require.Equal(t, buildBindings(alice, auth.RepoOwnerRole), getRepoRoleBinding(t, aliceClient, repo))

	// now alice can write to the repo
	commit, err := aliceClient.StartCommit(repo, "master")
	require.NoError(t, err)
	err = aliceClient.PutFile(repo, commit.ID, "/file", strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.ID))
	require.Equal(t, 1, CommitCnt(t, rootClient, repo)) // check that a new commit was created
}

// TestCannotRemoveRootAdmin tests that trying to remove the root user as an admin returns an error.
func TestCannotRemoveRootAdmin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)
	alice := robot(tu.UniqueString("alice"))
	aliceClient, rootClient := tu.GetAuthenticatedPachClient(t, alice), tu.GetAuthenticatedPachClient(t, auth.RootUser)

	// Check that the initial set of admins is just "admin"
	bindings, err := rootClient.GetClusterRoleBinding()
	require.NoError(t, err)
	require.Equal(t, buildClusterBindings(), bindings)

	// root cannot remove themselves from the list of super admins
	require.YesError(t, rootClient.ModifyClusterRoleBinding(auth.RootUser, []string{}))

	bindings, err = rootClient.GetClusterRoleBinding()
	require.NoError(t, err)
	require.Equal(t, buildClusterBindings(), bindings)

	// root can make alice a cluster administrator
	require.NoError(t, rootClient.ModifyClusterRoleBinding(alice, []string{auth.ClusterAdminRole}))
	bindings, err = rootClient.GetClusterRoleBinding()
	require.NoError(t, err)
	require.Equal(t, buildClusterBindings(alice, auth.ClusterAdminRole), bindings)

	// Root still cannot remove themselves as a cluster admin
	require.YesError(t, rootClient.ModifyClusterRoleBinding(auth.RootUser, []string{}))
	bindings, err = rootClient.GetClusterRoleBinding()
	require.NoError(t, err)
	require.Equal(t, buildClusterBindings(alice, auth.ClusterAdminRole), bindings)

	// alice is an admin, and she cannot remove root as an admin
	require.YesError(t, aliceClient.ModifyClusterRoleBinding(auth.RootUser, []string{}))
	bindings, err = rootClient.GetClusterRoleBinding()
	require.NoError(t, err)
	require.Equal(t, buildClusterBindings(alice, auth.ClusterAdminRole), bindings)
}

func TestPreActivationPipelinesKeepRunningAfterActivation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)
	alice := robot(tu.UniqueString("alice"))
	aliceClient, rootClient := tu.GetAuthenticatedPachClient(t, alice), tu.GetAuthenticatedPachClient(t, auth.RootUser)

	// Deactivate auth
	_, err := rootClient.Deactivate(rootClient.Ctx(), &auth.DeactivateRequest{})
	require.NoError(t, err)

	// Wait for auth to be deactivated
	require.NoError(t, backoff.Retry(func() error {
		_, err := aliceClient.WhoAmI(aliceClient.Ctx(), &auth.WhoAmIRequest{})
		if err != nil && auth.IsErrNotActivated(err) {
			return nil // WhoAmI should fail when auth is deactivated
		}
		return errors.New("auth is not yet deactivated")
	}, backoff.NewTestingBackOff()))

	// alice creates a pipeline
	repo := tu.UniqueString("TestPreActivationPipelinesKeepRunningAfterActivation")
	pipeline := tu.UniqueString("alice-pipeline")
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.NoError(t, aliceClient.CreatePipeline(
		pipeline,
		"", // default image: ubuntu:14.04
		[]string{"bash"},
		[]string{fmt.Sprintf("cp /pfs/%s/* /pfs/out/", repo)},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(repo, "/*"),
		"", // default output branch: master
		false,
	))

	// alice makes an input commit
	commit, err := aliceClient.StartCommit(repo, "master")
	require.NoError(t, err)
	err = aliceClient.PutFile(repo, commit.ID, "/file1", strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.ID))

	// make sure the pipeline runs
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := aliceClient.FlushCommitAll(
			[]*pfs.Commit{commit},
			[]*pfs.Repo{{Name: pipeline}},
		)
		return err
	})

	// activate auth
	resp, err := rootClient.Activate(rootClient.Ctx(), &auth.ActivateRequest{RootToken: tu.RootToken})
	require.NoError(t, err)
	rootClient.SetAuthToken(resp.PachToken)

	// activate auth in PPS
	_, err = rootClient.ActivateAuth(rootClient.Ctx(), &pps.ActivateAuthRequest{})
	require.NoError(t, err)

	// re-authenticate, as old tokens were deleted
	aliceClient = tu.GetAuthenticatedPachClient(t, alice)

	// Make sure alice cannot read the input repo (i.e. if the pipeline runs as
	// alice, it will fail)
	buf := &bytes.Buffer{}
	err = aliceClient.GetFile(repo, "master", "/file1", buf)
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())

	// Admin creates an input commit
	commit, err = rootClient.StartCommit(repo, "master")
	require.NoError(t, err)
	err = rootClient.PutFile(repo, commit.ID, "/file2", strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, rootClient.FinishCommit(repo, commit.ID))

	// make sure the pipeline still runs (i.e. it's not running as alice)
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := rootClient.FlushCommitAll(
			[]*pfs.Commit{commit},
			[]*pfs.Repo{{Name: pipeline}},
		)
		return err
	})
}

func TestExpirationRepoOnlyAccessibleToAdmins(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)

	// Configure OIDC so alice is a real IDP user
	tu.ConfigureOIDCProvider(t)
	alice := user(tu.DexMockConnectorEmail)
	aliceClient, rootClient := tu.GetAuthenticatedPachClient(t, alice), tu.GetAuthenticatedPachClient(t, auth.RootUser)

	// alice creates a repo
	repo := tu.UniqueString("TestExpirationRepoOnlyAccessibleToAdmins")
	require.NoError(t, aliceClient.CreateRepo(repo))

	// alice creates a commit
	commit, err := aliceClient.StartCommit(repo, "master")
	require.NoError(t, err)
	err = aliceClient.PutFile(repo, commit.ID, "/file1", strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.ID))
	require.Equal(t, 1, CommitCnt(t, aliceClient, repo))

	// Make current enterprise token expire
	rootClient.License.Activate(rootClient.Ctx(),
		&license.ActivateRequest{
			ActivationCode: tu.GetTestEnterpriseCode(t),
			Expires:        TSProtoOrDie(t, time.Now().Add(-30*time.Second)),
		})
	rootClient.Enterprise.Activate(rootClient.Ctx(),
		&enterprise.ActivateRequest{
			LicenseServer: "localhost:650",
			Id:            "localhost",
			Secret:        "localhost",
		})

	// wait for Enterprise token to expire
	require.NoError(t, backoff.Retry(func() error {
		resp, err := rootClient.Enterprise.GetState(rootClient.Ctx(),
			&enterprise.GetStateRequest{})
		if err != nil {
			return err
		}
		if resp.State == enterprise.State_ACTIVE {
			return errors.New("Pachyderm Enterprise is still active")
		}
		return nil
	}, backoff.NewTestingBackOff()))

	// now alice can't read from the repo
	buf := &bytes.Buffer{}
	err = aliceClient.GetFile(repo, "master", "/file1", buf)
	require.YesError(t, err)
	require.Matches(t, "not active", err.Error())

	// alice can't write to the repo
	_, err = aliceClient.StartCommit(repo, "master")
	require.YesError(t, err)
	require.Matches(t, "not active", err.Error())
	require.Equal(t, 1, CommitCnt(t, rootClient, repo)) // check that no commits were created

	// alice can't update the ACL
	err = aliceClient.ModifyRepoRoleBinding(repo, robot("carol"), []string{auth.RepoReaderRole})
	require.YesError(t, err)
	require.Matches(t, "not active", err.Error())
	// We don't delete the ACL because the user might re-enable enterprise pachyderm
	require.Equal(t, buildBindings(alice, auth.RepoOwnerRole), getRepoRoleBinding(t, rootClient, repo)) // check that ACL wasn't updated

	// alice also can't re-authenticate
	loginInfo, err := aliceClient.GetOIDCLogin(aliceClient.Ctx(), &auth.GetOIDCLoginRequest{})
	require.NoError(t, err)
	tu.DoOAuthExchange(t, loginInfo.LoginURL)
	_, err = aliceClient.Authenticate(aliceClient.Ctx(), &auth.AuthenticateRequest{OIDCState: loginInfo.State})
	require.Matches(t, "not active", err.Error())

	// admin can read from the repo
	buf.Reset()
	require.NoError(t, rootClient.GetFile(repo, "master", "/file1", buf))
	require.Matches(t, "test data", buf.String())

	// admin can write to the repo
	commit, err = rootClient.StartCommit(repo, "master")
	require.NoError(t, err)
	err = rootClient.PutFile(repo, commit.ID, "/file2", strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, rootClient.FinishCommit(repo, commit.ID))
	require.Equal(t, 2, CommitCnt(t, rootClient, repo)) // check that a new commit was created

	// admin can update the repo's ACL
	err = rootClient.ModifyRepoRoleBinding(repo, robot("carol"), []string{auth.RepoReaderRole})
	require.NoError(t, err)
	// check that ACL was updated
	require.Equal(t, buildBindings(alice, auth.RepoOwnerRole, robot("carol"), "repoReader"), getRepoRoleBinding(t, rootClient, repo))

	// Re-enable enterprise
	year := 365 * 24 * time.Hour
	rootClient.License.Activate(rootClient.Ctx(),
		&license.ActivateRequest{
			ActivationCode: tu.GetTestEnterpriseCode(t),
			// This will stop working some time in 2026
			Expires: TSProtoOrDie(t, time.Now().Add(year)),
		})

	rootClient.Enterprise.Activate(rootClient.Ctx(),
		&enterprise.ActivateRequest{
			LicenseServer: "localhost:650",
			Id:            "localhost",
			Secret:        "localhost",
		})

	// wait for Enterprise token to re-enable
	require.NoError(t, backoff.Retry(func() error {
		resp, err := rootClient.Enterprise.GetState(rootClient.Ctx(),
			&enterprise.GetStateRequest{})
		if err != nil {
			return err
		}
		if resp.State != enterprise.State_ACTIVE {
			return errors.New("Pachyderm Enterprise is still expired")
		}
		return nil
	}, backoff.NewTestingBackOff()))

	// alice can now re-authenticate
	loginInfo, err = aliceClient.GetOIDCLogin(aliceClient.Ctx(), &auth.GetOIDCLoginRequest{})
	require.NoError(t, err)
	tu.DoOAuthExchange(t, loginInfo.LoginURL)
	authResp, err := aliceClient.Authenticate(aliceClient.Ctx(), &auth.AuthenticateRequest{OIDCState: loginInfo.State})
	require.NoError(t, err)
	aliceClient.SetAuthToken(authResp.PachToken)

	// alice can read from the repo again
	buf = &bytes.Buffer{}
	require.NoError(t, aliceClient.GetFile(repo, "master", "/file1", buf))
	require.Matches(t, "test data", buf.String())

	// alice can write to the repo again
	commit, err = aliceClient.StartCommit(repo, "master")
	require.NoError(t, err)
	err = aliceClient.PutFile(repo, commit.ID, "/file3", strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.ID))
	require.Equal(t, 3, CommitCnt(t, aliceClient, repo)) // check that a new commit was created

	// alice can update the ACL again
	err = aliceClient.ModifyRepoRoleBinding(repo, robot("carol"), []string{auth.RepoWriterRole})
	require.NoError(t, err)
	// check that ACL was updated
	require.Equal(t, buildBindings(alice, auth.RepoOwnerRole, robot("carol"), auth.RepoWriterRole), getRepoRoleBinding(t, rootClient, repo))
}

func TestPipelinesRunAfterExpiration(t *testing.T) {
	if os.Getenv("RUN_BAD_TESTS") == "" {
		t.Skip("Skipping because RUN_BAD_TESTS was empty")
	}
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)
	alice := robot(tu.UniqueString("alice"))
	aliceClient, rootClient := tu.GetAuthenticatedPachClient(t, alice), tu.GetAuthenticatedPachClient(t, auth.RootUser)

	// alice creates a repo
	repo := tu.UniqueString("TestPipelinesRunAfterExpiration")
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.Equal(t, buildBindings(alice, auth.RepoOwnerRole), getRepoRoleBinding(t, aliceClient, repo))

	// alice creates a pipeline
	pipeline := tu.UniqueString("alice-pipeline")
	require.NoError(t, aliceClient.CreatePipeline(
		pipeline,
		"", // default image: ubuntu:14.04
		[]string{"bash"},
		[]string{fmt.Sprintf("cp /pfs/%s/* /pfs/out/", repo)},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(repo, "/*"),
		"",    // default output branch: master
		false, // no update
	))
	require.OneOfEquals(t, pipeline, PipelineNames(t, aliceClient))
	// check that alice owns the output repo too,
	require.Equal(t, buildBindings(alice, auth.RepoOwnerRole, pl(pipeline), auth.RepoWriterRole), getRepoRoleBinding(t, aliceClient, pipeline))

	// Make sure alice's pipeline runs successfully
	commit, err := aliceClient.StartCommit(repo, "master")
	require.NoError(t, err)
	err = aliceClient.PutFile(repo, commit.ID, tu.UniqueString("/file1"),
		strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.ID))
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := aliceClient.FlushCommitAll(
			[]*pfs.Commit{commit},
			[]*pfs.Repo{{Name: pipeline}},
		)
		return err
	})

	// Make current enterprise token expire
	rootClient.License.Activate(rootClient.Ctx(),
		&license.ActivateRequest{
			ActivationCode: tu.GetTestEnterpriseCode(t),
			Expires:        TSProtoOrDie(t, time.Now().Add(-30*time.Second)),
		})
	rootClient.Enterprise.Activate(rootClient.Ctx(),
		&enterprise.ActivateRequest{
			LicenseServer: "localhost:650",
			Id:            "localhost",
			Secret:        "localhost",
		})

	// wait for Enterprise token to expire
	require.NoError(t, backoff.Retry(func() error {
		resp, err := rootClient.Enterprise.GetState(rootClient.Ctx(),
			&enterprise.GetStateRequest{})
		if err != nil {
			return err
		}
		if resp.State == enterprise.State_ACTIVE {
			return errors.New("Pachyderm Enterprise is still active")
		}
		return nil
	}, backoff.NewTestingBackOff()))

	// Make sure alice's pipeline still runs successfully
	commit, err = rootClient.StartCommit(repo, "master")
	require.NoError(t, err)
	err = rootClient.PutFile(repo, commit.ID, tu.UniqueString("/file2"),
		strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, rootClient.FinishCommit(repo, commit.ID))
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := rootClient.FlushCommitAll(
			[]*pfs.Commit{commit},
			[]*pfs.Repo{{Name: pipeline}},
		)
		return err
	})
}

// Tests that GetRoleBinding and ModifyRoleBinding respect expired
// Enterprise tokens (i.e. reject non-admin requests once the token is expired,
// and allow admin requests)
func TestGetAndModifyRoleBindingWithEnterpriseExpired(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)
	alice := robot(tu.UniqueString("alice"))
	aliceClient, rootClient := tu.GetAuthenticatedPachClient(t, alice), tu.GetAuthenticatedPachClient(t, auth.RootUser)

	// alice creates a repo
	repo := tu.UniqueString("TestGetSetScopeAndAclWithExpiredToken")
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.Equal(t, buildBindings(alice, auth.RepoOwnerRole), getRepoRoleBinding(t, aliceClient, repo))

	// Make current enterprise token expire
	rootClient.License.Activate(rootClient.Ctx(),
		&license.ActivateRequest{
			ActivationCode: tu.GetTestEnterpriseCode(t),
			Expires:        TSProtoOrDie(t, time.Now().Add(-30*time.Second)),
		})
	rootClient.Enterprise.Activate(rootClient.Ctx(),
		&enterprise.ActivateRequest{
			LicenseServer: "localhost:650",
			Id:            "localhost",
			Secret:        "localhost",
		})

	// wait for Enterprise token to expire
	require.NoError(t, backoff.Retry(func() error {
		resp, err := rootClient.Enterprise.GetState(rootClient.Ctx(),
			&enterprise.GetStateRequest{})
		if err != nil {
			return err
		}
		if resp.State == enterprise.State_ACTIVE {
			return errors.New("Pachyderm Enterprise is still active")
		}
		return nil
	}, backoff.NewTestingBackOff()))

	// alice can't call GetRoleBinding on repo, even though she owns it
	_, err := aliceClient.GetRepoRoleBinding(repo)
	require.YesError(t, err)
	require.Matches(t, "not active", err.Error())

	// alice can't call ModifyRoleBinding on repo
	err = aliceClient.ModifyRepoRoleBinding(repo, robot("carol"), []string{auth.RepoReaderRole})
	require.YesError(t, err)
	require.Matches(t, "not active", err.Error())
	require.Equal(t, buildBindings(alice, auth.RepoOwnerRole), getRepoRoleBinding(t, rootClient, repo))

	// admin *can* call GetRoleBinding on repo
	bindings, err := rootClient.GetRepoRoleBinding(repo)
	require.NoError(t, err)
	require.Equal(t, buildBindings(alice, auth.RepoOwnerRole), bindings)

	// admin can call ModifyRoleBinding on repo
	err = rootClient.ModifyRepoRoleBinding(repo, robot("carol"), []string{auth.RepoReaderRole})
	require.NoError(t, err)

	bindings, err = rootClient.GetRepoRoleBinding(repo)
	require.NoError(t, err)
	require.Equal(t, buildBindings(alice, auth.RepoOwnerRole, robot("carol"), auth.RepoReaderRole), bindings)
}

// TestListRepoAdminIsOwnerOfAllRepos tests that when an admin calls ListRepo,
// the result indicates that they're an owner of every repo in the cluster
// (needed by the Pachyderm dashboard)
func TestListRepoAdminIsOwnerOfAllRepos(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)
	// t.Parallel()
	rootClient := tu.GetAuthenticatedPachClient(t, auth.RootUser)
	alice, bob := robot(tu.UniqueString("alice")), robot(tu.UniqueString("bob"))
	aliceClient, bobClient := tu.GetAuthenticatedPachClient(t, alice), tu.GetAuthenticatedPachClient(t, bob)

	// alice creates a repo
	repoWriter := tu.UniqueString("TestListRepoAdminIsOwnerOfAllRepos")
	require.NoError(t, aliceClient.CreateRepo(repoWriter))

	// bob calls ListRepo, but has NONE access to all repos
	infos, err := bobClient.ListRepo()
	require.NoError(t, err)
	for _, info := range infos {
		require.Nil(t, info.AuthInfo.Permissions)
	}

	// admin calls ListRepo, and has OWNER access to all repos
	infos, err = rootClient.ListRepo()
	require.NoError(t, err)
	for _, info := range infos {
		require.Equal(t, []auth.Permission{
			auth.Permission_REPO_READ,
			auth.Permission_REPO_WRITE,
			auth.Permission_REPO_MODIFY_BINDINGS,
			auth.Permission_REPO_DELETE,
			auth.Permission_REPO_INSPECT_COMMIT,
			auth.Permission_REPO_LIST_COMMIT,
			auth.Permission_REPO_DELETE_COMMIT,
			auth.Permission_REPO_CREATE_BRANCH,
			auth.Permission_REPO_LIST_BRANCH,
			auth.Permission_REPO_DELETE_BRANCH,
			auth.Permission_REPO_LIST_FILE,
			auth.Permission_REPO_INSPECT_FILE,
		}, info.AuthInfo.Permissions)
	}
}

// TestGetAuthToken tests that an admin can manufacture auth credentials for
// arbitrary other users
func TestGetAuthToken(t *testing.T) {
	if os.Getenv("RUN_BAD_TESTS") == "" {
		t.Skip("Skipping because RUN_BAD_TESTS was empty")
	}
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)
	rootClient := tu.GetAuthenticatedPachClient(t, auth.RootUser)

	// Generate first set of auth credentials
	robotUser := robot(tu.UniqueString("optimus_prime"))
	resp, err := rootClient.GetAuthToken(rootClient.Ctx(),
		&auth.GetAuthTokenRequest{Subject: robotUser})
	require.NoError(t, err)
	token1 := resp.Token
	robotClient1 := tu.GetUnauthenticatedPachClient(t)
	robotClient1.SetAuthToken(token1)

	// Confirm identity tied to 'token1'
	who, err := robotClient1.WhoAmI(robotClient1.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, robotUser, who.Username)
	if version.IsAtLeast(1, 10) {
		require.True(t, who.TTL >= 0 && who.TTL < secsInYear)
	} else {
		require.Equal(t, -1, who.TTL)
	}

	// Generate a second set of auth credentials--confirm that a unique token is
	// generated, but that the identity tied to it is the same
	resp, err = rootClient.GetAuthToken(rootClient.Ctx(),
		&auth.GetAuthTokenRequest{Subject: robotUser})
	require.NoError(t, err)
	token2 := resp.Token
	require.NotEqual(t, token1, token2)
	robotClient2 := tu.GetUnauthenticatedPachClient(t)
	robotClient2.SetAuthToken(token2)

	// Confirm identity tied to 'token1'
	who, err = robotClient2.WhoAmI(robotClient2.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, robotUser, who.Username)
	require.True(t, who.TTL >= 0 && who.TTL < secsInYear)

	// robotClient1 creates a repo
	repo := tu.UniqueString("TestPipelinesRunAfterExpiration")
	require.NoError(t, robotClient1.CreateRepo(repo))
	require.Equal(t, buildBindings(robotUser, auth.RepoOwnerRole), getRepoRoleBinding(t, robotClient1, repo))

	// robotClient1 creates a pipeline
	pipeline := tu.UniqueString("optimus-prime-line")
	require.NoError(t, robotClient1.CreatePipeline(
		pipeline,
		"", // default image: ubuntu:14.04
		[]string{"bash"},
		[]string{fmt.Sprintf("cp /pfs/%s/* /pfs/out/", repo)},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(repo, "/*"),
		"",    // default output branch: master
		false, // no update
	))
	require.OneOfEquals(t, pipeline, PipelineNames(t, robotClient1))
	// check that robotUser owns the output repo
	require.Equal(t, buildBindings(robotUser, auth.RepoOwnerRole, pl(pipeline), auth.RepoWriterRole), getRepoRoleBinding(t, robotClient1, pipeline))

	// Make sure that robotClient2 can commit to the input repo and flush their
	// input commit
	commit, err := robotClient2.StartCommit(repo, "master")
	require.NoError(t, err)
	err = robotClient2.PutFile(repo, commit.ID, tu.UniqueString("/file1"),
		strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, robotClient2.FinishCommit(repo, commit.ID))
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := robotClient2.FlushCommitAll(
			[]*pfs.Commit{commit},
			[]*pfs.Repo{{Name: pipeline}},
		)
		return err
	})

	// Make sure robotClient2 can update the pipeline, and it still runs
	// successfully
	require.NoError(t, robotClient2.CreatePipeline(
		pipeline,
		"", // default image: ubuntu:14.04
		[]string{"bash"},
		[]string{fmt.Sprintf("cp /pfs/%s/* /pfs/out/", repo)},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(repo, "/*"),
		"",   // default output branch: master
		true, // update
	))
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := robotClient2.FlushCommitAll(
			[]*pfs.Commit{commit},
			[]*pfs.Repo{{Name: pipeline}},
		)
		return err
	})
}

// TestGetTokenForRootUser tests that nobody can get a new token for the
// root user, except for the one issued at cluster creation.
func TestGetTokenForRootUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)
	rootClient := tu.GetAuthenticatedPachClient(t, auth.RootUser)

	// Try and get credentials as the root user, specifying the name
	_, err := rootClient.GetAuthToken(rootClient.Ctx(),
		&auth.GetAuthTokenRequest{Subject: auth.RootUser})
	require.YesError(t, err)
	require.Equal(t, "rpc error: code = Unknown desc = GetAuthTokenRequest.Subject is invalid", err.Error())
}

// TestGetIndefiniteAuthToken tests that an admin can generate an auth token that never
// times out if explicitly requested (e.g. for a daemon)
func TestGetIndefiniteAuthToken(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)
	rootClient := tu.GetAuthenticatedPachClient(t, auth.RootUser)

	// Generate auth credentials
	robotUser := robot(tu.UniqueString("rock-em-sock-em"))
	resp, err := rootClient.GetAuthToken(rootClient.Ctx(),
		&auth.GetAuthTokenRequest{
			Subject: robotUser,
			TTL:     -1,
		})
	require.NoError(t, err)
	token1 := resp.Token
	robotClient1 := tu.GetUnauthenticatedPachClient(t)
	robotClient1.SetAuthToken(token1)

	// Confirm identity tied to 'token1'
	who, err := robotClient1.WhoAmI(robotClient1.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, robotUser, who.Username)
	require.Equal(t, int64(-1), who.TTL)
}

// TestRobotUserWhoAmI tests that robot users can call WhoAmI and get a response
// with the right prefix
func TestRobotUserWhoAmI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)
	rootClient := tu.GetAuthenticatedPachClient(t, auth.RootUser)

	// Generate a robot user auth credential, and create a client for that user
	robotUser := robot(tu.UniqueString("r2d2"))
	resp, err := rootClient.GetAuthToken(rootClient.Ctx(),
		&auth.GetAuthTokenRequest{Subject: robotUser})
	require.NoError(t, err)
	// copy client & use resp token
	robotClient := rootClient.WithCtx(context.Background())
	robotClient.SetAuthToken(resp.Token)

	who, err := robotClient.WhoAmI(robotClient.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, robotUser, who.Username)
	require.True(t, strings.HasPrefix(who.Username, auth.RobotPrefix))
}

// TestRobotUserACL tests that a robot user can create a repo, add users
// to their repo, and be added to user's repo.
func TestRobotUserACL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	alice := robot(tu.UniqueString("alice"))
	aliceClient, rootClient := tu.GetAuthenticatedPachClient(t, alice), tu.GetAuthenticatedPachClient(t, auth.RootUser)

	// Generate a robot user auth credential, and create a client for that user
	robotUser := robot(tu.UniqueString("voltron"))
	resp, err := rootClient.GetAuthToken(rootClient.Ctx(),
		&auth.GetAuthTokenRequest{Subject: robotUser})
	require.NoError(t, err)
	// copy client & use resp token
	robotClient := rootClient.WithCtx(context.Background())
	robotClient.SetAuthToken(resp.Token)

	// robotUser creates a repo and adds alice as a writer
	repo := tu.UniqueString("TestRobotUserACL")
	require.NoError(t, robotClient.CreateRepo(repo))
	require.Equal(t, buildBindings(robotUser, auth.RepoOwnerRole), getRepoRoleBinding(t, robotClient, repo))

	require.NoError(t, robotClient.ModifyRepoRoleBinding(repo, alice, []string{auth.RepoWriterRole}))
	require.Equal(t, buildBindings(alice, auth.RepoWriterRole, robotUser, auth.RepoOwnerRole), getRepoRoleBinding(t, robotClient, repo))

	// test that alice can commit to the robot user's repo
	commit, err := aliceClient.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.ID))

	// Now alice creates a repo, and adds robotUser as a writer
	repo2 := tu.UniqueString("TestRobotUserACL")
	require.NoError(t, aliceClient.CreateRepo(repo2))
	require.Equal(t, buildBindings(alice, auth.RepoOwnerRole), getRepoRoleBinding(t, aliceClient, repo2))
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(repo2, robotUser, []string{auth.RepoWriterRole}))
	require.Equal(t, buildBindings(alice, auth.RepoOwnerRole, robotUser, auth.RepoWriterRole), getRepoRoleBinding(t, aliceClient, repo2))

	// test that the robot can commit to alice's repo
	commit, err = robotClient.StartCommit(repo2, "master")
	require.NoError(t, err)
	require.NoError(t, robotClient.FinishCommit(repo2, commit.ID))
}

// TestRobotUserAdmin tests that robot users can
// 1) become admins
// 2) mint tokens for robot and non-robot users
// 3) access other users' repos
// 4) update repo ACLs,
func TestRobotUserAdmin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)
	alice := robot(tu.UniqueString("alice"))
	rootClient := tu.GetAuthenticatedPachClient(t, auth.RootUser)
	aliceClient := tu.GetAuthenticatedPachClient(t, alice)

	// Generate a robot user auth credential, and create a client for that user
	robotUser := robot(tu.UniqueString("bender"))
	resp, err := rootClient.GetAuthToken(rootClient.Ctx(),
		&auth.GetAuthTokenRequest{Subject: robotUser})
	require.NoError(t, err)
	// copy client & use resp token
	robotClient := rootClient.WithCtx(context.Background())
	robotClient.SetAuthToken(resp.Token)

	// make robotUser an admin
	require.NoError(t, rootClient.ModifyClusterRoleBinding(robotUser, []string{auth.ClusterAdminRole}))
	// wait until robotUser shows up in admin list
	bindings, err := rootClient.GetClusterRoleBinding()
	require.NoError(t, err)
	require.Equal(t, buildClusterBindings(robotUser, auth.ClusterAdminRole), bindings)

	// robotUser mints a token for robotUser2
	robotUser2 := robot(tu.UniqueString("robocop"))
	resp, err = robotClient.GetAuthToken(robotClient.Ctx(), &auth.GetAuthTokenRequest{
		Subject: robotUser2,
	})
	require.NoError(t, err)
	require.NotEqual(t, "", resp.Token)
	robotClient2 := rootClient.WithCtx(context.Background())
	robotClient2.SetAuthToken(resp.Token)

	// robotUser2 creates a repo, and robotUser commits to it
	repo := tu.UniqueString("TestRobotUserAdmin")
	require.NoError(t, robotClient2.CreateRepo(repo))
	commit, err := robotClient.StartCommit(repo, "master")
	require.NoError(t, err) // admin privs means robotUser can commit
	require.NoError(t, robotClient.FinishCommit(repo, commit.ID))

	// robotUser adds alice to the repo, and checks that the ACL is updated
	require.Equal(t, buildBindings(robotUser2, auth.RepoOwnerRole), getRepoRoleBinding(t, robotClient, repo))
	require.NoError(t, robotClient.ModifyRepoRoleBinding(repo, alice, []string{auth.RepoWriterRole}))
	require.Equal(t, buildBindings(robotUser2, auth.RepoOwnerRole, alice, auth.RepoWriterRole), getRepoRoleBinding(t, robotClient, repo))
	commit, err = aliceClient.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.ID))

	robotClient.Deactivate(robotClient.Ctx(), &auth.DeactivateRequest{})
	require.NoError(t, err)
}

// TestTokenTTL tests that an admin can create a token with a TTL for a user,
// and that that token will expire
func TestTokenTTL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)
	rootClient := tu.GetAuthenticatedPachClient(t, auth.RootUser)

	// Create repo (so alice has something to list)
	repo := tu.UniqueString("TestTokenTTL")
	require.NoError(t, rootClient.CreateRepo(repo))

	// Create auth token for alice
	alice := robot(tu.UniqueString("alice"))
	resp, err := rootClient.GetAuthToken(rootClient.Ctx(), &auth.GetAuthTokenRequest{
		Subject: alice,
		TTL:     5, // seconds
	})
	require.NoError(t, err)
	aliceClient := rootClient.WithCtx(context.Background())
	aliceClient.SetAuthToken(resp.Token)

	// alice's token is valid, but expires quickly
	repos, err := aliceClient.ListRepo()
	require.NoError(t, err)
	require.ElementsEqualUnderFn(t, []string{repo}, repos, RepoInfoToName)
	require.NoError(t, backoff.Retry(func() error {
		repos, err = aliceClient.ListRepo()
		if err == nil {
			return errors.New("alice still has access to ListRepo")
		}
		require.True(t, auth.IsErrBadToken(err), err.Error())
		require.Equal(t, 0, len(repos))
		return nil
	}, backoff.NewTestingBackOff()))
}

// TestTokenTTLExtend tests that after creating a token, the admin can extend
// it, and it expires at the new time
func TestTokenTTLExtend(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)
	rootClient := tu.GetAuthenticatedPachClient(t, auth.RootUser)

	// Create repo (so alice has something to list)
	repo := tu.UniqueString("TestTokenTTLExtend")
	require.NoError(t, rootClient.CreateRepo(repo))

	alice := robot(tu.UniqueString("alice"))
	resp, err := rootClient.GetAuthToken(rootClient.Ctx(), &auth.GetAuthTokenRequest{
		Subject: alice,
		TTL:     5, // seconds
	})
	require.NoError(t, err)
	aliceClient := rootClient.WithCtx(context.Background())
	aliceClient.SetAuthToken(resp.Token)

	// alice's token is valid, but expires quickly
	repos, err := aliceClient.ListRepo()
	require.NoError(t, err)
	require.ElementsEqualUnderFn(t, []string{repo}, repos, RepoInfoToName)
	time.Sleep(10 * time.Second)
	repos, err = aliceClient.ListRepo()
	require.True(t, auth.IsErrBadToken(err), err.Error())
	require.Equal(t, 0, len(repos))

	// admin gives alice another token but extends it. Now it doesn't expire after
	// 10 seconds, but 20
	resp, err = rootClient.GetAuthToken(rootClient.Ctx(), &auth.GetAuthTokenRequest{
		Subject: alice,
		TTL:     5, // seconds
	})
	require.NoError(t, err)
	aliceClient.SetAuthToken(resp.Token)
	// token is valid
	repos, err = aliceClient.ListRepo()
	require.NoError(t, err)
	require.ElementsEqualUnderFn(t, []string{repo}, repos, RepoInfoToName)
	// admin extends token
	_, err = rootClient.ExtendAuthToken(rootClient.Ctx(), &auth.ExtendAuthTokenRequest{
		Token: resp.Token,
		TTL:   15,
	})
	require.NoError(t, err)

	// token is still valid after 10 seconds
	time.Sleep(10 * time.Second)
	time.Sleep(time.Second)
	repos, err = aliceClient.ListRepo()
	require.NoError(t, err)
	require.ElementsEqualUnderFn(t, []string{repo}, repos, RepoInfoToName)
	// wait longer, now token is expired
	time.Sleep(10 * time.Second)
	repos, err = aliceClient.ListRepo()
	require.True(t, auth.IsErrBadToken(err), err.Error())
	require.Equal(t, 0, len(repos))
}

// TestTokenRevoke tests that an admin can revoke that token and it no longer works
func TestTokenRevoke(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)
	rootClient := tu.GetAuthenticatedPachClient(t, auth.RootUser)

	// Create repo (so alice has something to list)
	repo := tu.UniqueString("TestTokenRevoke")
	require.NoError(t, rootClient.CreateRepo(repo))

	alice := robot(tu.UniqueString("alice"))
	resp, err := rootClient.GetAuthToken(rootClient.Ctx(), &auth.GetAuthTokenRequest{
		Subject: alice,
	})
	require.NoError(t, err)
	aliceClient := rootClient.WithCtx(context.Background())
	aliceClient.SetAuthToken(resp.Token)

	// alice's token is valid
	repos, err := aliceClient.ListRepo()
	require.NoError(t, err)
	require.ElementsEqualUnderFn(t, []string{repo}, repos, RepoInfoToName)

	// admin revokes token
	_, err = rootClient.RevokeAuthToken(rootClient.Ctx(), &auth.RevokeAuthTokenRequest{
		Token: resp.Token,
	})
	require.NoError(t, err)

	// alice's token is no longer valid
	repos, err = aliceClient.ListRepo()
	require.True(t, auth.IsErrBadToken(err), err.Error())
	require.Equal(t, 0, len(repos))
}

// TestGetAuthTokenErrorNonAdminUser tests that non-admin users can't call
// GetAuthToken on behalf of another user
func TestGetAuthTokenErrorNonAdminUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)

	alice := robot(tu.UniqueString("alice"))
	aliceClient := tu.GetAuthenticatedPachClient(t, alice)
	resp, err := aliceClient.GetAuthToken(aliceClient.Ctx(), &auth.GetAuthTokenRequest{
		Subject: robot(tu.UniqueString("t-1000")),
	})
	require.Nil(t, resp)
	require.YesError(t, err)
	require.Matches(t, "needs permissions \\[CLUSTER_ADMIN\\] on CLUSTER", err.Error())
}

// TestGetAuthTokenErrorFSAdminUser tests that FS admin users can't call
// GetAuthToken on behalf of another user
func TestGetAuthTokenErrorFSAdminUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)

	alice := robot(tu.UniqueString("alice"))
	aliceClient := tu.GetAuthenticatedPachClient(t, alice)
	rootClient := tu.GetAuthenticatedPachClient(t, auth.RootUser)

	// 'admin' makes alice an fs admin
	require.NoError(t, rootClient.ModifyClusterRoleBinding(alice, []string{auth.RepoOwnerRole}))

	// wait until alice shows up in admin list
	bindings, err := aliceClient.GetClusterRoleBinding()
	require.NoError(t, err)
	require.Equal(t, buildClusterBindings(alice, auth.RepoOwnerRole), bindings)

	// Try to get a token for a robot as alice
	resp, err := aliceClient.GetAuthToken(aliceClient.Ctx(), &auth.GetAuthTokenRequest{
		Subject: robot(tu.UniqueString("t-1000")),
	})
	require.Nil(t, resp)
	require.YesError(t, err)
	require.Matches(t, "needs permissions \\[CLUSTER_ADMIN\\] on CLUSTER", err.Error())
}

// TestDeleteAllAfterDeactivate tests that deleting repos and (particularly)
// pipelines works if auth was deactivated after they were created. Pipelines
// store a unique auth token after auth is activated, and if that auth token
// is used in the deletion process, DeletePipeline (and therefore DeleteAll)
// fails.
func TestDeleteAllAfterDeactivate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)

	alice := robot(tu.UniqueString("alice"))
	aliceClient, rootClient := tu.GetAuthenticatedPachClient(t, alice), tu.GetAuthenticatedPachClient(t, auth.RootUser)

	// alice creates a pipeline
	repo := tu.UniqueString("TestDeleteAllAfterDeactivate")
	pipeline := tu.UniqueString("pipeline")
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.NoError(t, aliceClient.CreatePipeline(
		pipeline,
		"", // default image: ubuntu:14.04
		[]string{"bash"},
		[]string{fmt.Sprintf("cp /pfs/%s/* /pfs/out/", repo)},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(repo, "/*"),
		"", // default output branch: master
		false,
	))

	// alice makes an input commit
	commit, err := aliceClient.StartCommit(repo, "master")
	require.NoError(t, err)
	err = aliceClient.PutFile(repo, commit.ID, "/file1", strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.ID))

	// make sure the pipeline runs
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := aliceClient.FlushCommitAll(
			[]*pfs.Commit{commit},
			[]*pfs.Repo{{Name: pipeline}},
		)
		return err
	})

	// Deactivate auth
	_, err = rootClient.Deactivate(rootClient.Ctx(), &auth.DeactivateRequest{})
	require.NoError(t, err)

	// Wait for auth to be deactivated
	require.NoError(t, backoff.Retry(func() error {
		_, err := aliceClient.WhoAmI(aliceClient.Ctx(), &auth.WhoAmIRequest{})
		if err != nil && auth.IsErrNotActivated(err) {
			return nil // WhoAmI should fail when auth is deactivated
		}
		return errors.New("auth is not yet deactivated")
	}, backoff.NewTestingBackOff()))

	// Make sure DeleteAll() succeeds
	require.NoError(t, aliceClient.DeleteAll())
}

// TestDeleteRCInStandby creates a pipeline, waits for it to enter standby, and
// then deletes its RC. This should not crash the PPS master, and the
// flush-commit run on an input commit should eventually return (though the
// pipeline may fail rather than processing anything in this state)
//
// Note: Like 'TestNoOutputRepoDoesntCrashPPSMaster', this test doesn't use the
// admin client at all, but it uses the kubernetes client, so out of prudence it
// shouldn't be run in parallel with any other test
func TestDeleteRCInStandby(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)
	alice := robot(tu.UniqueString("alice"))
	c := tu.GetAuthenticatedPachClient(t, alice)

	// Create input repo w/ initial commit
	repo := tu.UniqueString(t.Name())
	require.NoError(t, c.CreateRepo(repo))
	err := c.PutFile(repo, "master", "/file.1", strings.NewReader("1"))
	require.NoError(t, err)

	// Create pipeline
	pipeline := tu.UniqueString("pipeline")
	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pipeline),
			Transform: &pps.Transform{
				Image: "ubuntu:16.04",
				Cmd:   []string{"bash"},
				Stdin: []string{"cp /pfs/*/* /pfs/out"},
			},
			ParallelismSpec: &pps.ParallelismSpec{Constant: 1},
			Input:           client.NewPFSInput(repo, "/*"),
			Standby:         true,
		})
	require.NoError(t, err)

	// Wait for pipeline to process input commit & go into standby
	require.NoErrorWithinT(t, 30*time.Second, func() error {
		_, err := c.FlushCommitAll(
			[]*pfs.Commit{client.NewCommit(repo, "master")},
			[]*pfs.Repo{client.NewRepo(pipeline)})
		return err
	})
	require.NoErrorWithinTRetry(t, 30*time.Second, func() error {
		pi, err := c.InspectPipeline(pipeline)
		if err != nil {
			return err
		}
		if pi.State != pps.PipelineState_PIPELINE_STANDBY {
			return errors.Errorf("pipeline should be in standby, but is in %s", pi.State.String())
		}
		return nil
	})

	// delete pipeline RC
	tu.DeletePipelineRC(t, pipeline)

	// Create new input commit (to force pipeline out of standby) & make sure
	// flush-commit returns (pipeline either fails or restarts RC & finishes)
	err = c.PutFile(repo, "master", "/file.2", strings.NewReader("1"))
	require.NoError(t, err)
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := c.FlushCommitAll(
			[]*pfs.Commit{client.NewCommit(repo, "master")},
			[]*pfs.Repo{client.NewRepo(pipeline)})
		return err
	})
}

// TestNoOutputRepoDoesntCrashPPSMaster creates a pipeline, then deletes its
// output repo while it's running (failing the pipeline and preventing the PPS
// master from finishing the pipeline's output commit) and makes sure new
// pipelines can be created (i.e. that the PPS master doesn't crashloop due to
// the missing output repo). This test also exists in pachyderm_test.go, but
// duplicating it here ensures that Pachyderm handles this case correctly even
// when the missing output repo yields "access denied" instead of "not found"
// errors.
//
// Note: arguably deleting the output repo of a pipeline, even one that's
// stopped, should be prevented. However, the way that PPS currently uses PFS
// (stopping a pipeline = removing output branch subvenance) we have no way to
// prevent that. Thus, this test at least makes sure that doing such a thing
// doesn't break the PPS master.
//
// Note: This test actually doesn't use the admin client or admin privileges
// anywhere. However, it restarts pachd, so it shouldn't be run in parallel with
// any other test (which is expected of tests in auth_test.go)
func TestNoOutputRepoDoesntCrashPPSMaster(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)
	alice := robot(tu.UniqueString("alice"))
	aliceClient := tu.GetAuthenticatedPachClient(t, alice)

	// Create input repo w/ initial commit
	repo := tu.UniqueString(t.Name())
	require.NoError(t, aliceClient.CreateRepo(repo))
	err := aliceClient.PutFile(repo, "master", "/file.1", strings.NewReader("1"))
	require.NoError(t, err)

	// Create pipeline
	pipeline := tu.UniqueString("pipeline")
	require.NoError(t, aliceClient.CreatePipeline(
		pipeline,
		"", // default image: ubuntu:16.04
		[]string{"bash"},
		[]string{
			"sleep 10",
			"cp /pfs/*/* /pfs/out/",
		},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(repo, "/*"),
		"", // default output branch: master
		false,
	))

	// force-delete output repo while 'sleep 10' is running, failing the pipeline
	require.NoError(t, aliceClient.DeleteRepo(pipeline, true))

	// make sure the pipeline is failed
	require.NoErrorWithinTRetry(t, 30*time.Second, func() error {
		pi, err := aliceClient.InspectPipeline(pipeline)
		if err != nil {
			return err
		}
		if pi.State == pps.PipelineState_PIPELINE_FAILURE {
			return errors.Errorf("%q should be in state FAILURE but is in %q", pipeline, pi.State.String())
		}
		return nil
	})

	// Delete the pachd pod, so that it restarts and the PPS master has to process
	// the failed pipeline
	tu.DeletePachdPod(t) // delete the pachd pod
	require.NoErrorWithinTRetry(t, 30*time.Second, func() error {
		_, err := aliceClient.Version() // wait for pachd to come back
		return err
	})

	// Create a new input commit, and flush its output to 'pipeline', to make sure
	// the pipeline either restarts the RC and recreates the output repo, or fails
	err = aliceClient.PutFile(repo, "master", "/file.2", strings.NewReader("2"))
	require.NoError(t, err)
	require.NoErrorWithinT(t, 30*time.Second, func() error {
		// TODO(msteffen): While not currently possible, PFS could return
		// CommitDeleted here. This should detect that error, but first:
		// - src/server/pfs/pfs.go should be moved to src/client/pfs (w/ other err
		//   handling code)
		// - packages depending on that code should be migrated
		// Then this could add "|| pfs.IsCommitDeletedErr(err)" and satisfy the todo
		if _, err := aliceClient.FlushCommitAll(
			[]*pfs.Commit{client.NewCommit(repo, "master")},
			[]*pfs.Repo{client.NewRepo(pipeline)}); err != nil {
			return errors.Wrapf(err, "unexpected error value")
		}
		return nil
	})

	// Create a new pipeline, make sure FlushCommit eventually returns, and check
	// pipeline output (i.e. the PPS master does not crashloop--pipeline2
	// eventually starts successfully)
	pipeline2 := tu.UniqueString("pipeline")
	require.NoError(t, aliceClient.CreatePipeline(
		pipeline2,
		"", // default image: ubuntu:16.04
		[]string{"bash"},
		[]string{"cp /pfs/*/* /pfs/out/"},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(repo, "/*"),
		"", // default output branch: master
		false,
	))
	require.NoErrorWithinT(t, 30*time.Second, func() error {
		_, err := aliceClient.FlushCommitAll(
			[]*pfs.Commit{client.NewCommit(repo, "master")},
			[]*pfs.Repo{client.NewRepo(pipeline2)})
		return err
	})
	buf := &bytes.Buffer{}
	require.NoError(t, aliceClient.GetFile(pipeline2, "master", "/file.1", buf))
	require.Equal(t, "1", buf.String())
	buf.Reset()
	require.NoError(t, aliceClient.GetFile(pipeline2, "master", "/file.2", buf))
	require.Equal(t, "2", buf.String())
}

// TestPipelineFailingWithOpenCommit creates a pipeline, then revokes its access
// to its output repo while it's running, causing it to fail. Then it makes sure
// that FlushCommit still works and that the pipeline's output commit was
// successfully finished (though as an empty commit)
//
// Note: This test actually doesn't use the admin client or admin privileges
// anywhere. However, it restarts pachd, so it shouldn't be run in parallel with
// any other test
func TestPipelineFailingWithOpenCommit(t *testing.T) {
	// TODO: Reenable when finishing job state is transactional.
	t.Skip("Job state does not get finished in a transaction, so stats commit is left open")
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)
	alice := robot(tu.UniqueString("alice"))
	aliceClient, rootClient := tu.GetAuthenticatedPachClient(t, alice), tu.GetAuthenticatedPachClient(t, auth.RootUser)

	// Create input repo w/ initial commit
	repo := tu.UniqueString(t.Name())
	require.NoError(t, aliceClient.CreateRepo(repo))
	err := aliceClient.PutFile(repo, "master", "/file.1", strings.NewReader("1"))
	require.NoError(t, err)

	// Create pipeline
	pipeline := tu.UniqueString("pipeline")
	require.NoError(t, aliceClient.CreatePipeline(
		pipeline,
		"", // default image: ubuntu:16.04
		[]string{"bash"},
		[]string{
			"sleep 10",
			"cp /pfs/*/* /pfs/out/",
		},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(repo, "/*"),
		"", // default output branch: master
		false,
	))

	// Revoke pipeline's access to output repo while 'sleep 10' is running (so
	// that it fails)
	require.NoError(t, rootClient.ModifyRepoRoleBinding(repo, fmt.Sprintf("pipeline:%s", pipeline), []string{}))

	// make sure flush-commit returns (pipeline either
	// fails or restarts RC & finishes)
	require.NoErrorWithinT(t, 30*time.Second, func() error {
		_, err := aliceClient.FlushCommitAll(
			[]*pfs.Commit{client.NewCommit(repo, "master")},
			[]*pfs.Repo{client.NewRepo(pipeline)})
		return err
	})

	// make sure the pipeline is failed
	pi, err := rootClient.InspectPipeline(pipeline)
	require.NoError(t, err)
	require.Equal(t, pps.PipelineState_PIPELINE_FAILURE, pi.State)
}
