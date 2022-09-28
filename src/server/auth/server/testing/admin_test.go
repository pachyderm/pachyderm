//go:build !k8s

// admin_test.go tests various features related to pachyderm's auth admins.
// Because the cluster has one global set of admins, these tests can't be run in
// parallel

package server_test

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

// TestActivate tests the Activate API (in particular, verifying
// that Activate() also authenticates). Even though GetClient also activates
// auth, this makes sure the code path is exercised (as auth may already be
// active when the test starts)
func TestActivate(t *testing.T) {
	env := envWithAuth(t)
	c := env.PachClient
	// Get anonymous client (this will activate auth, which is about to be
	// deactivated, but it also activates Pacyderm enterprise, which is needed for
	// this test to pass)
	rootClient := tu.AuthenticateClient(t, c, auth.RootUser)

	_, err := rootClient.Deactivate(rootClient.Ctx(), &auth.DeactivateRequest{})
	require.NoError(t, err)
	resp, err := rootClient.AuthAPIClient.Activate(context.Background(), &auth.ActivateRequest{})
	require.NoError(t, err)
	rootClient.SetAuthToken(resp.PachToken)
	defer func() {
		_, err := rootClient.Deactivate(rootClient.Ctx(), &auth.DeactivateRequest{})
		require.NoError(t, err)
	}()

	// Check that the token 'c' received from pachd authenticates them as "pach:root"
	who, err := rootClient.WhoAmI(rootClient.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, auth.RootUser, who.Username)

	bindings, err := rootClient.GetClusterRoleBinding()
	require.NoError(t, err)
	require.Equal(t, tu.BuildClusterBindings(), bindings)
}

// TestActivateKnownToken tests activating auth with a known token.
// This should always authenticate the user as `pach:root` and give them
// super admin status.
func TestActivateKnownToken(t *testing.T) {
	env := envWithAuth(t)
	c := env.PachClient
	// Get anonymous client (this will activate auth, which is about to be
	// deactivated, but it also activates Pacyderm enterprise, which is needed for
	// this test to pass)
	rootClient := tu.AuthenticateClient(t, c, auth.RootUser)

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
	require.Equal(t, tu.BuildClusterBindings(), bindings)
}

// TestSuperAdminRWO tests adding and removing cluster super admins, as well as super admins
// reading, writing, and moderating (owning) all repos in the cluster.
func TestSuperAdminRWO(t *testing.T) {
	env := envWithAuth(t)
	c := env.PachClient
	alice, bob := tu.Robot(tu.UniqueString("alice")), tu.Robot(tu.UniqueString("bob"))
	aliceClient, bobClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, bob)
	rootClient := tu.AuthenticateClient(t, c, auth.RootUser)

	// The initial set of admins is just the user "admin"
	bindings, err := aliceClient.GetClusterRoleBinding()
	require.NoError(t, err)
	require.Equal(t, tu.BuildClusterBindings(), bindings)

	// alice creates a repo (that only she owns) and puts a file
	repo := tu.UniqueString("TestAdminRWO")
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.Equal(t, tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, repo))
	commit, err := aliceClient.StartCommit(repo, "master")
	require.NoError(t, err)
	err = aliceClient.PutFile(commit, "/file", strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.Branch.Name, commit.ID))

	// bob cannot read from the repo
	buf := &bytes.Buffer{}
	err = bobClient.GetFile(commit, "/file", buf)
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())

	// bob cannot write to the repo
	_, err = bobClient.StartCommit(repo, "master")
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	// Note: we must pass aliceClient to tu.CommitCnt, because it calls
	// ListCommit(repo), which requires the caller to have READER access to
	// 'repo', which bob does not have (but alice does)
	require.Equal(t, 1, tu.CommitCnt(t, aliceClient, repo)) // check that no commits were created

	// bob can't update the ACL
	err = bobClient.ModifyRepoRoleBinding(repo, tu.Robot("carol"), []string{auth.RepoReaderRole})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.Equal(t, tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, repo)) // check that ACL wasn't updated

	// 'admin' makes bob a super admin
	require.NoError(t, rootClient.ModifyClusterRoleBinding(bob, []string{auth.ClusterAdminRole}))

	// wait until bob shows up in admin list
	bindings, err = aliceClient.GetClusterRoleBinding()
	require.NoError(t, err)
	require.Equal(t, tu.BuildClusterBindings(bob, auth.ClusterAdminRole), bindings)

	// now bob can read from the repo
	buf.Reset()
	require.NoError(t, bobClient.GetFile(commit, "/file", buf))
	require.Matches(t, "test data", buf.String())

	// bob can write to the repo
	commit, err = bobClient.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, bobClient.FinishCommit(repo, commit.Branch.Name, commit.ID))
	require.Equal(t, 2, tu.CommitCnt(t, aliceClient, repo)) // check that a new commit was created

	// bob can update the repo's ACL
	err = bobClient.ModifyRepoRoleBinding(repo, tu.Robot("carol"), []string{auth.RepoReaderRole})
	require.NoError(t, err)
	// check that ACL was updated
	require.Equal(t, tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Robot("carol"), auth.RepoReaderRole), tu.GetRepoRoleBinding(t, aliceClient, repo))

	// 'admin' revokes bob's admin status
	require.NoError(t, rootClient.ModifyClusterRoleBinding(bob, []string{}))

	// wait until bob is not in admin list
	bindings, err = aliceClient.GetClusterRoleBinding()
	require.NoError(t, err)
	require.Equal(t, tu.BuildClusterBindings(), bindings)

	// bob can no longer read from the repo
	buf.Reset()
	err = bobClient.GetFile(commit, "/file", buf)
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())

	// bob cannot write to the repo
	_, err = bobClient.StartCommit(repo, "master")
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.Equal(t, 2, tu.CommitCnt(t, aliceClient, repo)) // check that no commits were created

	// bob can't update the ACL
	err = bobClient.ModifyRepoRoleBinding(repo, tu.Robot("carol"), []string{auth.RepoWriterRole})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	// check that ACL wasn't updated
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Robot("carol"), auth.RepoReaderRole), tu.GetRepoRoleBinding(t, aliceClient, repo))
}

// TestFSAdminRWO tests adding and removing cluster FS admins, as well as FS admins
// reading, writing, and moderating (owning) all repos in the cluster.
func TestFSAdminRWO(t *testing.T) {
	env := envWithAuth(t)
	c := env.PachClient
	alice, bob := tu.Robot(tu.UniqueString("alice")), tu.Robot(tu.UniqueString("bob"))
	aliceClient, bobClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, bob)
	rootClient := tu.AuthenticateClient(t, c, auth.RootUser)

	// The initial set of admins is just the user "admin"
	bindings, err := aliceClient.GetClusterRoleBinding()
	require.NoError(t, err)
	require.Equal(t, tu.BuildClusterBindings(), bindings)

	// alice creates a repo (that only she owns) and puts a file
	repo := tu.UniqueString("TestAdminRWO")
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.Equal(t, tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, repo))
	commit, err := aliceClient.StartCommit(repo, "master")
	require.NoError(t, err)
	err = aliceClient.PutFile(commit, "/file", strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.Branch.Name, commit.ID))

	// bob cannot read from the repo
	buf := &bytes.Buffer{}
	err = bobClient.GetFile(commit, "/file", buf)
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())

	// bob cannot write to the repo
	_, err = bobClient.StartCommit(repo, "master")
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	// Note: we must pass aliceClient to tu.CommitCnt, because it calls
	// ListCommit(repo), which requires the caller to have READER access to
	// 'repo', which bob does not have (but alice does)
	require.Equal(t, 1, tu.CommitCnt(t, aliceClient, repo)) // check that no commits were created

	// bob can't update the ACL
	err = bobClient.ModifyRepoRoleBinding(repo, tu.Robot("carol"), []string{auth.RepoReaderRole})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.Equal(t, tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, repo)) // check that ACL wasn't updated

	// 'admin' makes bob an fs admin
	require.NoError(t, rootClient.ModifyClusterRoleBinding(bob, []string{auth.RepoOwnerRole}))

	// wait until bob shows up in admin list
	bindings, err = aliceClient.GetClusterRoleBinding()
	require.NoError(t, err)
	require.Equal(t, tu.BuildClusterBindings(bob, auth.RepoOwnerRole), bindings)

	// now bob can read from the repo
	buf.Reset()
	require.NoError(t, bobClient.GetFile(commit, "/file", buf))
	require.Matches(t, "test data", buf.String())

	// bob can write to the repo
	commit, err = bobClient.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, bobClient.FinishCommit(repo, commit.Branch.Name, commit.ID))
	require.Equal(t, 2, tu.CommitCnt(t, aliceClient, repo)) // check that a new commit was created

	// bob can update the repo's ACL
	err = bobClient.ModifyRepoRoleBinding(repo, tu.Robot("carol"), []string{auth.RepoReaderRole})
	require.NoError(t, err)
	// check that ACL was updated
	require.Equal(t, tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Robot("carol"), auth.RepoReaderRole), tu.GetRepoRoleBinding(t, aliceClient, repo))

	// 'admin' revokes bob's admin status
	require.NoError(t, rootClient.ModifyClusterRoleBinding(bob, []string{}))

	// wait until bob is not in admin list
	bindings, err = aliceClient.GetClusterRoleBinding()
	require.NoError(t, err)
	require.Equal(t, tu.BuildClusterBindings(), bindings)

	// bob can no longer read from the repo
	buf.Reset()
	err = bobClient.GetFile(commit, "/file", buf)
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())

	// bob cannot write to the repo
	_, err = bobClient.StartCommit(repo, "master")
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.Equal(t, 2, tu.CommitCnt(t, aliceClient, repo)) // check that no commits were created

	// bob can't update the ACL
	err = bobClient.ModifyRepoRoleBinding(repo, tu.Robot("carol"), []string{auth.RepoWriterRole})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	// check that ACL wasn't updated
	require.Equal(t, tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Robot("carol"), auth.RepoReaderRole), tu.GetRepoRoleBinding(t, aliceClient, repo))
}

// TestFSAdminFixBrokenRepo tests that an FS admin can modify the ACL of a repo even
// when the repo's ACL is empty (indicating that no user has explicit access to
// to the repo)
func TestFSAdminFixBrokenRepo(t *testing.T) {
	env := envWithAuth(t)
	c := env.PachClient
	alice, bob := tu.Robot(tu.UniqueString("alice")), tu.Robot(tu.UniqueString("bob"))
	aliceClient, bobClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, bob)
	rootClient := tu.AuthenticateClient(t, c, auth.RootUser)

	// alice creates a repo (that only she owns) and puts a file
	repo := tu.UniqueString("TestAdmin")
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.Equal(t, tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, repo))

	// 'admin' makes bob an FS admin
	require.NoError(t, rootClient.ModifyClusterRoleBinding(bob, []string{auth.RepoOwnerRole}))

	// wait until bob shows up in admin list
	bindings, err := aliceClient.GetClusterRoleBinding()
	require.NoError(t, err)
	require.Equal(t, tu.BuildClusterBindings(bob, auth.RepoOwnerRole), bindings)

	// admin deletes the repo's ACL
	require.NoError(t, rootClient.ModifyRepoRoleBinding(repo, alice, []string{}))

	// Check that the ACL is empty
	require.Nil(t, tu.GetRepoRoleBinding(t, rootClient, repo).Entries)

	// alice cannot write to the repo
	_, err = aliceClient.StartCommit(repo, "master")
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.Equal(t, 0, tu.CommitCnt(t, rootClient, repo)) // check that no commits were created

	// bob, an FS admin, can update the ACL to put Alice back, even though reading the ACL
	// will fail
	require.NoError(t, bobClient.ModifyRepoRoleBinding(repo, alice, []string{auth.RepoOwnerRole}))
	require.Equal(t, tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, repo))

	// now alice can write to the repo
	commit, err := aliceClient.StartCommit(repo, "master")
	require.NoError(t, err)
	err = aliceClient.PutFile(commit, "/file", strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.Branch.Name, commit.ID))
	require.Equal(t, 1, tu.CommitCnt(t, rootClient, repo)) // check that a new commit was created
}

// TestCannotRemoveRootAdmin tests that trying to remove the root user as an admin returns an error.
func TestCannotRemoveRootAdmin(t *testing.T) {
	env := envWithAuth(t)
	c := env.PachClient
	alice := tu.Robot(tu.UniqueString("alice"))
	aliceClient, rootClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, auth.RootUser)

	// Check that the initial set of admins is just "admin"
	bindings, err := rootClient.GetClusterRoleBinding()
	require.NoError(t, err)
	require.Equal(t, tu.BuildClusterBindings(), bindings)

	// root cannot remove themselves from the list of super admins
	require.YesError(t, rootClient.ModifyClusterRoleBinding(auth.RootUser, []string{}))

	bindings, err = rootClient.GetClusterRoleBinding()
	require.NoError(t, err)
	require.Equal(t, tu.BuildClusterBindings(), bindings)

	// root can make alice a cluster administrator
	require.NoError(t, rootClient.ModifyClusterRoleBinding(alice, []string{auth.ClusterAdminRole}))
	bindings, err = rootClient.GetClusterRoleBinding()
	require.NoError(t, err)
	require.Equal(t, tu.BuildClusterBindings(alice, auth.ClusterAdminRole), bindings)

	// Root still cannot remove themselves as a cluster admin
	require.YesError(t, rootClient.ModifyClusterRoleBinding(auth.RootUser, []string{}))
	bindings, err = rootClient.GetClusterRoleBinding()
	require.NoError(t, err)
	require.Equal(t, tu.BuildClusterBindings(alice, auth.ClusterAdminRole), bindings)

	// alice is an admin, and she cannot remove root as an admin
	require.YesError(t, aliceClient.ModifyClusterRoleBinding(auth.RootUser, []string{}))
	bindings, err = rootClient.GetClusterRoleBinding()
	require.NoError(t, err)
	require.Equal(t, tu.BuildClusterBindings(alice, auth.ClusterAdminRole), bindings)
}

func TestPreActivationPipelinesKeepRunningAfterActivation(t *testing.T) {
	env := envWithAuth(t)
	c := env.PachClient
	alice := tu.Robot(tu.UniqueString("alice"))
	aliceClient, rootClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, auth.RootUser)

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
		"", // default image: DefaultUserImage
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
	err = aliceClient.PutFile(commit, "/file1", strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.Branch.Name, commit.ID))

	// make sure the pipeline runs
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := aliceClient.WaitCommit(pipeline, "master", commit.ID)
		return err
	})

	// activate auth
	resp, err := rootClient.Activate(rootClient.Ctx(), &auth.ActivateRequest{RootToken: tu.RootToken})
	require.NoError(t, err)
	rootClient.SetAuthToken(resp.PachToken)

	// activate auth in PFS
	_, err = rootClient.PfsAPIClient.ActivateAuth(rootClient.Ctx(), &pfs.ActivateAuthRequest{})
	require.NoError(t, err)

	// activate auth in PPS
	_, err = rootClient.PpsAPIClient.ActivateAuth(rootClient.Ctx(), &pps.ActivateAuthRequest{})
	require.NoError(t, err)

	// re-authenticate, as old tokens were deleted
	aliceClient = tu.AuthenticateClient(t, c, alice)

	// Make sure alice cannot read the input repo (i.e. if the pipeline runs as
	// alice, it will fail)
	buf := &bytes.Buffer{}
	err = aliceClient.GetFile(commit, "/file1", buf)
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())

	// Admin creates an input commit
	commit, err = rootClient.StartCommit(repo, "master")
	require.NoError(t, err)
	err = rootClient.PutFile(commit, "/file2", strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, rootClient.FinishCommit(repo, commit.Branch.Name, commit.ID))

	// make sure the pipeline still runs (i.e. it's not running as alice)
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := rootClient.WaitCommit(pipeline, "master", commit.ID)
		return err
	})
}

// TestListRepoAdminIsOwnerOfAllRepos tests that when an admin calls ListRepo,
// the result indicates that they're an owner of every repo in the cluster
// (needed by the Pachyderm dashboard)
func TestListRepoAdminIsOwnerOfAllRepos(t *testing.T) {
	env := envWithAuth(t)
	c := env.PachClient
	rootClient := tu.AuthenticateClient(t, c, auth.RootUser)
	alice, bob := tu.Robot(tu.UniqueString("alice")), tu.Robot(tu.UniqueString("bob"))
	aliceClient, bobClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, bob)

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
		require.ElementsEqual(t, []string{"clusterAdmin"}, info.AuthInfo.Roles)
	}
}

// TestGetIndefiniteRobotToken tests that an admin can generate a robot token that never
// times out - this is the default behaviour
func TestGetIndefiniteRobotToken(t *testing.T) {
	env := envWithAuth(t)
	c := env.PachClient
	rootClient := tu.AuthenticateClient(t, c, auth.RootUser)

	// Generate auth credentials
	robotUser := tu.UniqueString("rock-em-sock-em")
	resp, err := rootClient.GetRobotToken(rootClient.Ctx(), &auth.GetRobotTokenRequest{Robot: robotUser})
	require.NoError(t, err)
	token1 := resp.Token
	robotClient1 := tu.UnauthenticatedPachClient(t, c)
	robotClient1.SetAuthToken(token1)

	// Confirm identity tied to 'token1'
	who, err := robotClient1.WhoAmI(robotClient1.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, tu.Robot(robotUser), who.Username)
	require.Nil(t, who.Expiration)
}

// TestGetTemporaryRobotToken tests that an admin can generate a robot token that expires
func TestGetTemporaryRobotToken(t *testing.T) {
	env := envWithAuth(t)
	c := env.PachClient
	rootClient := tu.AuthenticateClient(t, c, auth.RootUser)

	// Generate auth credentials
	robotUser := tu.UniqueString("rock-em-sock-em")
	resp, err := rootClient.GetRobotToken(rootClient.Ctx(), &auth.GetRobotTokenRequest{Robot: robotUser, TTL: 600})
	require.NoError(t, err)
	token1 := resp.Token
	robotClient1 := tu.UnauthenticatedPachClient(t, c)
	robotClient1.SetAuthToken(token1)

	// Confirm identity tied to 'token1'
	who, err := robotClient1.WhoAmI(robotClient1.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, tu.Robot(robotUser), who.Username)

	require.True(t, who.Expiration.After(time.Now()))
	require.True(t, who.Expiration.Before(time.Now().Add(time.Duration(600)*time.Second)))
}

// TestRobotUserWhoAmI tests that robot users can call WhoAmI and get a response
// with the right prefix
func TestRobotUserWhoAmI(t *testing.T) {
	env := envWithAuth(t)
	c := env.PachClient
	rootClient := tu.AuthenticateClient(t, c, auth.RootUser)

	// Generate a robot user auth credential, and create a client for that user
	robotUser := tu.UniqueString("r2d2")
	resp, err := rootClient.GetRobotToken(rootClient.Ctx(),
		&auth.GetRobotTokenRequest{Robot: robotUser})
	require.NoError(t, err)
	// copy client & use resp token
	robotClient := rootClient.WithCtx(context.Background())
	robotClient.SetAuthToken(resp.Token)

	who, err := robotClient.WhoAmI(robotClient.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, tu.Robot(robotUser), who.Username)
	require.True(t, strings.HasPrefix(who.Username, auth.RobotPrefix))
}

// TestRobotUserACL tests that a robot user can create a repo, add users
// to their repo, and be added to user's repo.
func TestRobotUserACL(t *testing.T) {
	env := envWithAuth(t)
	c := env.PachClient
	alice := tu.Robot(tu.UniqueString("alice"))
	aliceClient, rootClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, auth.RootUser)

	// Generate a robot user auth credential, and create a client for that user
	robotUser := tu.UniqueString("voltron")
	resp, err := rootClient.GetRobotToken(rootClient.Ctx(),
		&auth.GetRobotTokenRequest{Robot: robotUser})
	require.NoError(t, err)
	// copy client & use resp token
	robotClient := rootClient.WithCtx(context.Background())
	robotClient.SetAuthToken(resp.Token)

	// robotUser creates a repo and adds alice as a writer
	repo := tu.UniqueString("TestRobotUserACL")
	require.NoError(t, robotClient.CreateRepo(repo))
	require.Equal(t, tu.BuildBindings(tu.Robot(robotUser), auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, robotClient, repo))

	require.NoError(t, robotClient.ModifyRepoRoleBinding(repo, alice, []string{auth.RepoWriterRole}))
	require.Equal(t, tu.BuildBindings(alice, auth.RepoWriterRole, tu.Robot(robotUser), auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, robotClient, repo))

	// test that alice can commit to the robot user's repo
	commit, err := aliceClient.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.Branch.Name, commit.ID))

	// Now alice creates a repo, and adds robotUser as a writer
	repo2 := tu.UniqueString("TestRobotUserACL")
	require.NoError(t, aliceClient.CreateRepo(repo2))
	require.Equal(t, tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, repo2))
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(repo2, tu.Robot(robotUser), []string{auth.RepoWriterRole}))
	require.Equal(t, tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Robot(robotUser), auth.RepoWriterRole), tu.GetRepoRoleBinding(t, aliceClient, repo2))

	// test that the robot can commit to alice's repo
	commit, err = robotClient.StartCommit(repo2, "master")
	require.NoError(t, err)
	require.NoError(t, robotClient.FinishCommit(repo2, commit.Branch.Name, commit.ID))
}

// TestGroupRoleBinding tests that a group can be added to a role binding
// and confers access to members
func TestGroupRoleBinding(t *testing.T) {
	env := envWithAuth(t)
	c := env.PachClient
	alice := tu.Robot(tu.UniqueString("alice"))
	group := tu.Group(tu.UniqueString("testGroup"))
	aliceClient, rootClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, auth.RootUser)

	// root creates a repo and adds a group writer access
	repo := tu.UniqueString("TestGroupRoleBinding")
	require.NoError(t, rootClient.CreateRepo(repo))
	require.NoError(t, rootClient.ModifyRepoRoleBinding(repo, group, []string{auth.RepoWriterRole}))
	require.Equal(t, tu.BuildBindings(group, auth.RepoWriterRole, auth.RootUser, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, rootClient, repo))

	// add alice to the group
	_, err := rootClient.ModifyMembers(rootClient.Ctx(), &auth.ModifyMembersRequest{
		Group: group,
		Add:   []string{alice},
	})
	require.NoError(t, err)

	// test that alice can commit to the repo
	commit, err := aliceClient.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.Branch.Name, commit.ID))
}

// TestRobotUserAdmin tests that robot users can
// 1) become admins
// 2) mint tokens for robot and non-robot users
// 3) access other users' repos
// 4) update repo ACLs,
func TestRobotUserAdmin(t *testing.T) {
	env := envWithAuth(t)
	c := env.PachClient
	alice := tu.Robot(tu.UniqueString("alice"))
	rootClient := tu.AuthenticateClient(t, c, auth.RootUser)
	aliceClient := tu.AuthenticateClient(t, c, alice)

	// Generate a robot user auth credential, and create a client for that user
	robotUser := tu.UniqueString("bender")
	resp, err := rootClient.GetRobotToken(rootClient.Ctx(),
		&auth.GetRobotTokenRequest{Robot: robotUser})
	require.NoError(t, err)
	// copy client & use resp token
	robotClient := rootClient.WithCtx(context.Background())
	robotClient.SetAuthToken(resp.Token)

	// make robotUser an admin
	require.NoError(t, rootClient.ModifyClusterRoleBinding(tu.Robot(robotUser), []string{auth.ClusterAdminRole}))
	// wait until robotUser shows up in admin list
	bindings, err := rootClient.GetClusterRoleBinding()
	require.NoError(t, err)
	require.Equal(t, tu.BuildClusterBindings(tu.Robot(robotUser), auth.ClusterAdminRole), bindings)

	// robotUser mints a token for robotUser2
	robotUser2 := tu.UniqueString("robocop")
	resp, err = robotClient.GetRobotToken(robotClient.Ctx(), &auth.GetRobotTokenRequest{
		Robot: robotUser2,
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
	require.NoError(t, robotClient.FinishCommit(repo, commit.Branch.Name, commit.ID))

	// robotUser adds alice to the repo, and checks that the ACL is updated
	require.Equal(t, tu.BuildBindings(tu.Robot(robotUser2), auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, robotClient, repo))
	require.NoError(t, robotClient.ModifyRepoRoleBinding(repo, alice, []string{auth.RepoWriterRole}))
	require.Equal(t, tu.BuildBindings(tu.Robot(robotUser2), auth.RepoOwnerRole, alice, auth.RepoWriterRole), tu.GetRepoRoleBinding(t, robotClient, repo))
	commit, err = aliceClient.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.Branch.Name, commit.ID))

	_, err = robotClient.Deactivate(robotClient.Ctx(), &auth.DeactivateRequest{})
	require.NoError(t, err)
}

// TestTokenRevoke tests that an admin can revoke that token and it no longer works
func TestTokenRevoke(t *testing.T) {
	env := envWithAuth(t)
	c := env.PachClient
	rootClient := tu.AuthenticateClient(t, c, auth.RootUser)

	// Create repo (so alice has something to list)
	repo := tu.UniqueString("TestTokenRevoke")
	require.NoError(t, rootClient.CreateRepo(repo))

	alice := tu.UniqueString("alice")
	resp, err := rootClient.GetRobotToken(rootClient.Ctx(), &auth.GetRobotTokenRequest{
		Robot: alice,
	})
	require.NoError(t, err)
	aliceClient := rootClient.WithCtx(context.Background())
	aliceClient.SetAuthToken(resp.Token)

	// alice's token is valid
	repos, err := aliceClient.ListRepo()
	require.NoError(t, err)
	require.ElementsEqualUnderFn(t, []string{repo}, repos, tu.RepoInfoToName)

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

func TestRevokeTokensForUser(t *testing.T) {
	env := envWithAuth(t)
	c := env.PachClient

	contains := func(tokens []*auth.TokenInfo, hashedToken string) bool {
		for _, v := range tokens {
			if v.HashedToken == hashedToken {
				return true
			}
		}
		return false
	}

	rootClient := tu.AuthenticateClient(t, c, auth.RootUser)

	// Create repo (so alice has something to list)
	repo := tu.UniqueString("TestTokenRevoke")
	require.NoError(t, rootClient.CreateRepo(repo))

	alice := tu.UniqueString("robot:alice")
	bob := tu.UniqueString("robot:bob")

	// mint two tokens for Alice
	aliceTokenA, err := rootClient.GetRobotToken(rootClient.Ctx(), &auth.GetRobotTokenRequest{
		Robot: alice,
	})
	require.NoError(t, err)

	aliceTokenB, err := rootClient.GetRobotToken(rootClient.Ctx(), &auth.GetRobotTokenRequest{
		Robot: alice,
	})
	require.NoError(t, err)

	// mint one token for Bob
	bobToken, err := rootClient.GetRobotToken(rootClient.Ctx(), &auth.GetRobotTokenRequest{
		Robot: bob,
	})
	require.NoError(t, err)

	// verify all three tokens are extractable
	extractTokensResp, extractErr := rootClient.ExtractAuthTokens(rootClient.Ctx(), &auth.ExtractAuthTokensRequest{})
	require.NoError(t, extractErr)

	preRevokeTokens := extractTokensResp.Tokens
	require.Equal(t, 3, len(preRevokeTokens), "all three tokens should be returned")
	require.True(t, contains(preRevokeTokens, auth.HashToken(aliceTokenA.Token)), "Alice's Token A should be extracted")
	require.True(t, contains(preRevokeTokens, auth.HashToken(aliceTokenB.Token)), "Alice's Token B should be extracted")
	require.True(t, contains(preRevokeTokens, auth.HashToken(bobToken.Token)), "Bob's Token should be extracted")

	aliceClient := tu.UnauthenticatedPachClient(t, c)
	aliceClient.SetAuthToken(aliceTokenA.Token)

	bobClient := tu.UnauthenticatedPachClient(t, c)
	bobClient.SetAuthToken(bobToken.Token)

	// delete all tokens for user Alice
	_, revokeErr := rootClient.RevokeAuthTokensForUser(rootClient.Ctx(), &auth.RevokeAuthTokensForUserRequest{Username: alice})
	require.NoError(t, revokeErr)

	// verify Alice can no longer authenticate with either of her tokens
	_, whoAmIErr := aliceClient.WhoAmI(aliceClient.Ctx(), &auth.WhoAmIRequest{})
	require.YesError(t, whoAmIErr)

	aliceClient.SetAuthToken(aliceTokenB.Token)
	_, whoAmIErr = aliceClient.WhoAmI(aliceClient.Ctx(), &auth.WhoAmIRequest{})
	require.YesError(t, whoAmIErr)

	// verify Bob can still authenticate with his token
	_, whoAmIErr = bobClient.WhoAmI(bobClient.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, whoAmIErr)

	// verify only Bob's tokens are extractable
	extractTokensResp, extractErr = rootClient.ExtractAuthTokens(rootClient.Ctx(), &auth.ExtractAuthTokensRequest{})
	require.NoError(t, extractErr)
	postRevokeTokens := extractTokensResp.Tokens
	require.Equal(t, 1, len(postRevokeTokens), "There should now be two fewer tokens extracted")
	require.True(t, contains(postRevokeTokens, auth.HashToken(bobToken.Token)), "Bob's Token should be extracted")
}

// TestRevokePachUserToken tests that the pps superuser and root tokens can't
// be revoked.
func TestRevokePachUserToken(t *testing.T) {
	env := envWithAuth(t)
	c := env.PachClient
	rootClient := tu.AuthenticateClient(t, c, auth.RootUser)

	_, err := rootClient.RevokeAuthTokensForUser(rootClient.Ctx(), &auth.RevokeAuthTokensForUserRequest{Username: auth.RootUser})
	require.YesError(t, err)
	require.Matches(t, "cannot revoke tokens for pach: users", err.Error())
}

func TestRotateRootToken(t *testing.T) {
	env := envWithAuth(t)
	c := env.PachClient
	rootClient := tu.AuthenticateClient(t, c, auth.RootUser)

	// create a repo for the purpose of testing access
	repo := tu.UniqueString("TestRotateRootToken")
	require.NoError(t, rootClient.CreateRepo(repo))

	// rotate token after creating the repo
	rotateReq := &auth.RotateRootTokenRequest{}
	rotateResp, err := rootClient.RotateRootToken(rootClient.Ctx(), rotateReq)
	require.NoError(t, err)

	_, err = rootClient.ListRepo()
	require.YesError(t, err, "the list operation is expected to fail since the token configured into the client is no longer valid")

	rootClient.SetAuthToken(rotateResp.RootToken)
	listResp, err := rootClient.ListRepo()
	require.NoError(t, err)
	require.Equal(t, 1, len(listResp), "now that the rotated token is configured in the client, the operation should work")

	// now try setting the token
	rotateReq = &auth.RotateRootTokenRequest{
		RootToken: tu.RootToken,
	}
	rotateResp, err = rootClient.RotateRootToken(rootClient.Ctx(), rotateReq)
	require.NoError(t, err)
	require.Equal(t, rotateResp.RootToken, tu.RootToken)

	_, err = rootClient.ListRepo()
	require.YesError(t, err)

	rootClient.SetAuthToken(rotateResp.RootToken)
	listResp, err = rootClient.ListRepo()
	require.NoError(t, err)
	require.Equal(t, 1, len(listResp))
}
