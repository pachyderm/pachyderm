//go:build k8s

// admin_test.go tests various features related to pachyderm's auth admins.
// Because the cluster has one global set of admins, these tests can't be run in
// parallel

package server

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/license"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

func RepoInfoToName(repoInfo interface{}) interface{} {
	return repoInfo.(*pfs.RepoInfo).Repo.Name
}

func user(email string) string {
	return auth.UserPrefix + email
}

func group(group string) string {
	return auth.GroupPrefix + group
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
		auth.InternalPrefix+"auth-server", auth.ClusterAdminRole,
	)...)
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
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
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
	require.Equal(t, buildClusterBindings(), bindings)
}

// TestActivateKnownToken tests activating auth with a known token.
// This should always authenticate the user as `pach:root` and give them
// super admin status.
func TestActivateKnownToken(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
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
	require.Equal(t, buildClusterBindings(), bindings)
}

// TestSuperAdminRWO tests adding and removing cluster super admins, as well as super admins
// reading, writing, and moderating (owning) all repos in the cluster.
func TestSuperAdminRWO(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	alice, bob := robot(tu.UniqueString("alice")), robot(tu.UniqueString("bob"))
	aliceClient, bobClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, bob)
	rootClient := tu.AuthenticateClient(t, c, auth.RootUser)

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
	require.NoError(t, bobClient.GetFile(commit, "/file", buf))
	require.Matches(t, "test data", buf.String())

	// bob can write to the repo
	commit, err = bobClient.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, bobClient.FinishCommit(repo, commit.Branch.Name, commit.ID))
	require.Equal(t, 2, CommitCnt(t, aliceClient, repo)) // check that a new commit was created

	// bob can update the repo's ACL
	err = bobClient.ModifyRepoRoleBinding(repo, robot("carol"), []string{auth.RepoReaderRole})
	require.NoError(t, err)
	// check that ACL was updated
	require.Equal(t, buildBindings(alice, auth.RepoOwnerRole, robot("carol"), auth.RepoReaderRole), getRepoRoleBinding(t, aliceClient, repo))

	// 'admin' revokes bob's admin status
	require.NoError(t, rootClient.ModifyClusterRoleBinding(bob, []string{}))

	// wait until bob is not in admin list
	bindings, err = aliceClient.GetClusterRoleBinding()
	require.NoError(t, err)
	require.Equal(t, buildClusterBindings(), bindings)

	// bob can no longer read from the repo
	buf.Reset()
	err = bobClient.GetFile(commit, "/file", buf)
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
		buildBindings(alice, auth.RepoOwnerRole, robot("carol"), auth.RepoReaderRole), getRepoRoleBinding(t, aliceClient, repo))
}

// TestFSAdminRWO tests adding and removing cluster FS admins, as well as FS admins
// reading, writing, and moderating (owning) all repos in the cluster.
func TestFSAdminRWO(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	alice, bob := robot(tu.UniqueString("alice")), robot(tu.UniqueString("bob"))
	aliceClient, bobClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, bob)
	rootClient := tu.AuthenticateClient(t, c, auth.RootUser)

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
	require.NoError(t, bobClient.GetFile(commit, "/file", buf))
	require.Matches(t, "test data", buf.String())

	// bob can write to the repo
	commit, err = bobClient.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, bobClient.FinishCommit(repo, commit.Branch.Name, commit.ID))
	require.Equal(t, 2, CommitCnt(t, aliceClient, repo)) // check that a new commit was created

	// bob can update the repo's ACL
	err = bobClient.ModifyRepoRoleBinding(repo, robot("carol"), []string{auth.RepoReaderRole})
	require.NoError(t, err)
	// check that ACL was updated
	require.Equal(t, buildBindings(alice, auth.RepoOwnerRole, robot("carol"), auth.RepoReaderRole), getRepoRoleBinding(t, aliceClient, repo))

	// 'admin' revokes bob's admin status
	require.NoError(t, rootClient.ModifyClusterRoleBinding(bob, []string{}))

	// wait until bob is not in admin list
	bindings, err = aliceClient.GetClusterRoleBinding()
	require.NoError(t, err)
	require.Equal(t, buildClusterBindings(), bindings)

	// bob can no longer read from the repo
	buf.Reset()
	err = bobClient.GetFile(commit, "/file", buf)
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
	require.Equal(t, buildBindings(alice, auth.RepoOwnerRole, robot("carol"), auth.RepoReaderRole), getRepoRoleBinding(t, aliceClient, repo))
}

// TestFSAdminFixBrokenRepo tests that an FS admin can modify the ACL of a repo even
// when the repo's ACL is empty (indicating that no user has explicit access to
// to the repo)
func TestFSAdminFixBrokenRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	alice, bob := robot(tu.UniqueString("alice")), robot(tu.UniqueString("bob"))
	aliceClient, bobClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, bob)
	rootClient := tu.AuthenticateClient(t, c, auth.RootUser)

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
	err = aliceClient.PutFile(commit, "/file", strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.Branch.Name, commit.ID))
	require.Equal(t, 1, CommitCnt(t, rootClient, repo)) // check that a new commit was created
}

// TestCannotRemoveRootAdmin tests that trying to remove the root user as an admin returns an error.
func TestCannotRemoveRootAdmin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	alice := robot(tu.UniqueString("alice"))
	aliceClient, rootClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, auth.RootUser)

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
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	alice := robot(tu.UniqueString("alice"))
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

func TestPreActivationCronPipelinesKeepRunningAfterActivation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	alice := robot(tu.UniqueString("alice"))
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
	pipeline1 := tu.UniqueString("cron1-")
	require.NoError(t, aliceClient.CreatePipeline(
		pipeline1,
		"",
		[]string{"/bin/bash"},
		[]string{"cp /pfs/time/* /pfs/out/"},
		nil,
		client.NewCronInput("time", "@every 3s"),
		"",
		false,
	))
	pipeline2 := tu.UniqueString("cron2-")
	require.NoError(t, aliceClient.CreatePipeline(
		pipeline2,
		"",
		[]string{"/bin/bash"},
		[]string{"cp " + fmt.Sprintf("/pfs/%s/*", pipeline1) + " /pfs/out/"},
		nil,
		client.NewPFSInput(pipeline1, "/*"),
		"",
		false,
	))

	// subscribe to the pipeline2 cron repo and wait for inputs
	repo := client.NewRepo(pipeline2)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel() //cleanup resources

	checkCronCommits := func(n int) error {
		count := 0
		return aliceClient.WithCtx(ctx).SubscribeCommit(repo, "master", "", pfs.CommitState_FINISHED, func(ci *pfs.CommitInfo) error {
			files, err := aliceClient.ListFileAll(ci.Commit, "")
			require.NoError(t, err)

			require.Equal(t, count, len(files))
			count++
			if count > n {
				return errutil.ErrBreak
			}
			return nil
		})
	}
	// make sure the cron is working
	require.NoError(t, checkCronCommits(1))

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
	require.NoError(t, rootClient.ModifyClusterRoleBinding(alice, []string{auth.RepoWriterRole}))

	// make sure the cron is working
	require.NoError(t, checkCronCommits(5))
}

func TestPipelinesRunAfterExpiration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	alice := robot(tu.UniqueString("alice"))
	aliceClient, rootClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, auth.RootUser)

	// alice creates a repo
	repo := tu.UniqueString("TestPipelinesRunAfterExpiration")
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.Equal(t, buildBindings(alice, auth.RepoOwnerRole), getRepoRoleBinding(t, aliceClient, repo))

	// alice creates a pipeline
	pipeline := tu.UniqueString("alice-pipeline")
	require.NoError(t, aliceClient.CreatePipeline(
		pipeline,
		"", // default image: DefaultUserImage
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
	err = aliceClient.PutFile(commit, tu.UniqueString("/file1"),
		strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.Branch.Name, commit.ID))
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := aliceClient.WaitCommit(pipeline, "master", commit.ID)
		return err
	})

	// Make current enterprise token expire
	_, err = rootClient.License.Activate(rootClient.Ctx(),
		&license.ActivateRequest{
			ActivationCode: tu.GetTestEnterpriseCode(t),
			Expires:        tu.TSProtoOrDie(t, time.Now().Add(-30*time.Second)),
		})
	require.NoError(t, err)
	_, err = rootClient.Enterprise.Activate(rootClient.Ctx(),
		&enterprise.ActivateRequest{
			LicenseServer: "localhost:1650",
			Id:            "localhost",
			Secret:        "localhost",
		})
	require.NoError(t, err)

	// wait for Enterprise token to expire
	require.NoError(t, backoff.Retry(func() error {
		resp, err := rootClient.Enterprise.GetState(rootClient.Ctx(),
			&enterprise.GetStateRequest{})
		if err != nil {
			return errors.EnsureStack(err)
		}
		if resp.State == enterprise.State_ACTIVE {
			return errors.New("Pachyderm Enterprise is still active")
		}
		return nil
	}, backoff.NewTestingBackOff()))

	// Make sure alice's pipeline still runs successfully
	commit, err = rootClient.StartCommit(repo, "master")
	require.NoError(t, err)
	err = rootClient.PutFile(commit, tu.UniqueString("/file2"),
		strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, rootClient.FinishCommit(repo, commit.Branch.Name, commit.ID))
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := rootClient.WaitCommit(pipeline, "master", commit.ID)
		return err
	})
}

// TestListRepoAdminIsOwnerOfAllRepos tests that when an admin calls ListRepo,
// the result indicates that they're an owner of every repo in the cluster
// (needed by the Pachyderm dashboard)
func TestListRepoAdminIsOwnerOfAllRepos(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	// t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	rootClient := tu.AuthenticateClient(t, c, auth.RootUser)
	alice, bob := robot(tu.UniqueString("alice")), robot(tu.UniqueString("bob"))
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
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
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
	require.Equal(t, robot(robotUser), who.Username)
	require.Nil(t, who.Expiration)
}

// TestGetTemporaryRobotToken tests that an admin can generate a robot token that expires
func TestGetTemporaryRobotToken(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
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
	require.Equal(t, robot(robotUser), who.Username)

	require.True(t, who.Expiration.After(time.Now()))
	require.True(t, who.Expiration.Before(time.Now().Add(time.Duration(600)*time.Second)))
}

// TestGetRobotTokenErrorNonAdminUser tests that non-admin users can't call
// GetRobotToken
func TestGetRobotTokenErrorNonAdminUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	alice := robot(tu.UniqueString("alice"))
	aliceClient := tu.AuthenticateClient(t, c, alice)
	resp, err := aliceClient.GetRobotToken(aliceClient.Ctx(), &auth.GetRobotTokenRequest{
		Robot: tu.UniqueString("t-1000"),
	})
	require.Nil(t, resp)
	require.YesError(t, err)
	require.Matches(t, "needs permissions \\[CLUSTER_AUTH_GET_ROBOT_TOKEN\\] on CLUSTER", err.Error())
}

// TestRobotUserWhoAmI tests that robot users can call WhoAmI and get a response
// with the right prefix
func TestRobotUserWhoAmI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
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
	require.Equal(t, robot(robotUser), who.Username)
	require.True(t, strings.HasPrefix(who.Username, auth.RobotPrefix))
}

// TestRobotUserACL tests that a robot user can create a repo, add users
// to their repo, and be added to user's repo.
func TestRobotUserACL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	alice := robot(tu.UniqueString("alice"))
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
	require.Equal(t, buildBindings(robot(robotUser), auth.RepoOwnerRole), getRepoRoleBinding(t, robotClient, repo))

	require.NoError(t, robotClient.ModifyRepoRoleBinding(repo, alice, []string{auth.RepoWriterRole}))
	require.Equal(t, buildBindings(alice, auth.RepoWriterRole, robot(robotUser), auth.RepoOwnerRole), getRepoRoleBinding(t, robotClient, repo))

	// test that alice can commit to the robot user's repo
	commit, err := aliceClient.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.Branch.Name, commit.ID))

	// Now alice creates a repo, and adds robotUser as a writer
	repo2 := tu.UniqueString("TestRobotUserACL")
	require.NoError(t, aliceClient.CreateRepo(repo2))
	require.Equal(t, buildBindings(alice, auth.RepoOwnerRole), getRepoRoleBinding(t, aliceClient, repo2))
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(repo2, robot(robotUser), []string{auth.RepoWriterRole}))
	require.Equal(t, buildBindings(alice, auth.RepoOwnerRole, robot(robotUser), auth.RepoWriterRole), getRepoRoleBinding(t, aliceClient, repo2))

	// test that the robot can commit to alice's repo
	commit, err = robotClient.StartCommit(repo2, "master")
	require.NoError(t, err)
	require.NoError(t, robotClient.FinishCommit(repo2, commit.Branch.Name, commit.ID))
}

// TestGroupRoleBinding tests that a group can be added to a role binding
// and confers access to members
func TestGroupRoleBinding(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	alice := robot(tu.UniqueString("alice"))
	group := group(tu.UniqueString("testGroup"))
	aliceClient, rootClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, auth.RootUser)

	// root creates a repo and adds a group writer access
	repo := tu.UniqueString("TestGroupRoleBinding")
	require.NoError(t, rootClient.CreateRepo(repo))
	require.NoError(t, rootClient.ModifyRepoRoleBinding(repo, group, []string{auth.RepoWriterRole}))
	require.Equal(t, buildBindings(group, auth.RepoWriterRole, auth.RootUser, auth.RepoOwnerRole), getRepoRoleBinding(t, rootClient, repo))

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
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	alice := robot(tu.UniqueString("alice"))
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
	require.NoError(t, rootClient.ModifyClusterRoleBinding(robot(robotUser), []string{auth.ClusterAdminRole}))
	// wait until robotUser shows up in admin list
	bindings, err := rootClient.GetClusterRoleBinding()
	require.NoError(t, err)
	require.Equal(t, buildClusterBindings(robot(robotUser), auth.ClusterAdminRole), bindings)

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
	require.Equal(t, buildBindings(robot(robotUser2), auth.RepoOwnerRole), getRepoRoleBinding(t, robotClient, repo))
	require.NoError(t, robotClient.ModifyRepoRoleBinding(repo, alice, []string{auth.RepoWriterRole}))
	require.Equal(t, buildBindings(robot(robotUser2), auth.RepoOwnerRole, alice, auth.RepoWriterRole), getRepoRoleBinding(t, robotClient, repo))
	commit, err = aliceClient.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.Branch.Name, commit.ID))

	_, err = robotClient.Deactivate(robotClient.Ctx(), &auth.DeactivateRequest{})
	require.NoError(t, err)
}

// TestTokenRevoke tests that an admin can revoke that token and it no longer works
func TestTokenRevoke(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
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

func TestRevokeTokensForUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)

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
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	rootClient := tu.AuthenticateClient(t, c, auth.RootUser)

	_, err := rootClient.RevokeAuthTokensForUser(rootClient.Ctx(), &auth.RevokeAuthTokensForUserRequest{Username: auth.RootUser})
	require.YesError(t, err)
	require.Matches(t, "cannot revoke tokens for pach: users", err.Error())
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
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	alice := robot(tu.UniqueString("alice"))
	aliceClient, rootClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, auth.RootUser)

	// alice creates a pipeline
	repo := tu.UniqueString("TestDeleteAllAfterDeactivate")
	pipeline := tu.UniqueString("pipeline")
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
// commit should eventually finish (though the pipeline may fail rather than
// processing anything in this state)
//
// Note: Like 'TestNoOutputRepoDoesntCrashPPSMaster', this test doesn't use the
// admin client at all, but it uses the kubernetes client, so out of prudence it
// shouldn't be run in parallel with any other test
func TestDeleteRCInStandby(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, ns := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	alice := robot(tu.UniqueString("alice"))
	c = tu.AuthenticateClient(t, c, alice)

	// Create input repo w/ initial commit
	repo := tu.UniqueString(t.Name())
	require.NoError(t, c.CreateRepo(repo))
	err := c.PutFile(client.NewCommit(repo, "master", ""), "/file.1", strings.NewReader("1"))
	require.NoError(t, err)

	// Create pipeline
	pipeline := tu.UniqueString("pipeline")
	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pipeline),
			Transform: &pps.Transform{
				Image: "", // default image: DefaultUserImage
				Cmd:   []string{"bash"},
				Stdin: []string{"cp /pfs/*/* /pfs/out"},
			},
			ParallelismSpec: &pps.ParallelismSpec{Constant: 1},
			Input:           client.NewPFSInput(repo, "/*"),
			Autoscaling:     true,
		})
	require.NoError(t, err)

	// Wait for pipeline to process input commit & go into standby
	require.NoErrorWithinT(t, 30*time.Second, func() error {
		_, err := c.WaitCommit(pipeline, "master", "")
		return err
	})
	require.NoErrorWithinTRetry(t, 30*time.Second, func() error {
		pi, err := c.InspectPipeline(pipeline, false)
		if err != nil {
			return err
		}
		if pi.State != pps.PipelineState_PIPELINE_STANDBY {
			return errors.Errorf("pipeline should be in standby, but is in %s", pi.State.String())
		}
		return nil
	})

	// delete pipeline RC
	tu.DeletePipelineRC(t, pipeline, ns)

	// Create new input commit (to force pipeline out of standby) & make sure
	// the pipeline either fails or restarts RC & finishes
	err = c.PutFile(client.NewCommit(repo, "master", ""), "/file.2", strings.NewReader("1"))
	require.NoError(t, err)
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := c.WaitCommit(pipeline, "master", "")
		return err
	})
}

// TestPipelineFailingWithOpenCommit creates a pipeline, then revokes its access
// to its output repo while it's running, causing it to fail. Then it makes sure
// that FlushJob still works and that the pipeline's output commit was
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
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	alice := robot(tu.UniqueString("alice"))
	aliceClient, rootClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, auth.RootUser)

	// Create input repo w/ initial commit
	repo := tu.UniqueString(t.Name())
	commit := client.NewCommit(repo, "master", "")
	require.NoError(t, aliceClient.CreateRepo(repo))
	err := aliceClient.PutFile(commit, "/file.1", strings.NewReader("1"))
	require.NoError(t, err)

	// Create pipeline
	pipeline := tu.UniqueString("pipeline")
	require.NoError(t, aliceClient.CreatePipeline(
		pipeline,
		"", // default image: DefaultUserImage
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

	// make sure the pipeline either fails or restarts RC & finishes
	require.NoErrorWithinT(t, 30*time.Second, func() error {
		_, err := aliceClient.WaitCommit(pipeline, "master", commit.ID)
		return err
	})

	// make sure the pipeline is failed
	pi, err := rootClient.InspectPipeline(pipeline, false)
	require.NoError(t, err)
	require.Equal(t, pps.PipelineState_PIPELINE_FAILURE, pi.State)
}

func TestRotateRootToken(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
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
