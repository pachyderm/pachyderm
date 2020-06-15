// admin_test.go tests various features related to pachyderm's auth admins.
// Because the cluster has one global set of admins, these tests can't be run in
// parallel

package server

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/enterprise"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/client/version"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"

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

// helper method to generate a super admin role
func superAdminRole() *auth.AdminRoles {
	return &auth.AdminRoles{Roles: []auth.AdminRoles_Role{auth.AdminRoles_SUPER}}
}

// helper method to generate an fs admin role
func fsAdminRole() *auth.AdminRoles {
	return &auth.AdminRoles{Roles: []auth.AdminRoles_Role{auth.AdminRoles_FS}}
}

// helper function to generate map of admins
func admins(super ...string) func(fs ...string) map[string]*auth.AdminRoles {
	a := make(map[string]*auth.AdminRoles)
	for _, u := range super {
		a[u] = superAdminRole()
	}

	return func(fs ...string) map[string]*auth.AdminRoles {
		for _, u := range fs {
			if _, ok := a[u]; ok {
				a[u].Roles = append(a[u].Roles, auth.AdminRoles_FS)
			} else {
				a[u] = fsAdminRole()
			}
		}
		return a
	}
}

// helper function that prepends auth.GitHubPrefix to 'user'--useful for validating
// responses
func gh(user string) string {
	return auth.GitHubPrefix + user
}

func pl(pipeline string) string {
	return auth.PipelinePrefix + pipeline
}

func robot(robot string) string {
	return auth.RobotPrefix + robot
}

// TestActivate tests the Activate API (in particular, verifying
// that Activate() also authenticates). Even though GetClient also activates
// auth, this makes sure the code path is exercised (as auth may already be
// active when the test starts)
func TestActivate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	defer deleteAll(t)
	// Get anonymous client (this will activate auth, which is about to be
	// deactivated, but it also activates Pacyderm enterprise, which is needed for
	// this test to pass)
	adminClient := getPachClient(t, admin)
	_, err := adminClient.Deactivate(adminClient.Ctx(), &auth.DeactivateRequest{})
	require.NoError(t, err)
	resp, err := adminClient.AuthAPIClient.Activate(context.Background(),
		&auth.ActivateRequest{Subject: admin})
	require.NoError(t, err)
	tokenMap[admin] = resp.PachToken
	adminClient.SetAuthToken(resp.PachToken)

	// Check that the token 'c' received from pachd authenticates them as "admin"
	who, err := adminClient.WhoAmI(adminClient.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, auth.AdminRoles_SUPER, who.AdminRoles.Roles[0])
	require.Equal(t, admin, who.Username)
}

// TestSuperAdminRWO tests adding and removing cluster super admins, as well as super admins
// reading, writing, and moderating (owning) all repos in the cluster.
func TestSuperAdminRWO(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	defer deleteAll(t)
	alice, bob := tu.UniqueString("alice"), tu.UniqueString("bob")
	aliceClient, bobClient := getPachClient(t, alice), getPachClient(t, bob)
	adminClient := getPachClient(t, admin)

	// The initial set of admins is just the user "admin"
	resp, err := aliceClient.GetAdmins(aliceClient.Ctx(), &auth.GetAdminsRequest{})
	require.NoError(t, err)
	require.Equal(t, admins(admin)(), resp.Admins)

	// alice creates a repo (that only she owns) and puts a file
	repo := tu.UniqueString("TestAdminRWO")
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.Equal(t, entries(alice, "owner"), getACL(t, aliceClient, repo))
	commit, err := aliceClient.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = aliceClient.PutFile(repo, commit.ID, "/file", strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.ID))

	// bob cannot read from the repo
	buf := &bytes.Buffer{}
	err = bobClient.GetFile(repo, "master", "/file", 0, 0, buf)
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
	_, err = bobClient.SetScope(bobClient.Ctx(), &auth.SetScopeRequest{
		Repo:     repo,
		Username: "carol",
		Scope:    auth.Scope_READER,
	})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.Equal(t, entries(alice, "owner"), getACL(t, aliceClient, repo)) // check that ACL wasn't updated

	// 'admin' makes bob a super admin
	_, err = adminClient.ModifyAdmins(adminClient.Ctx(),
		&auth.ModifyAdminsRequest{Principal: gh(bob), Roles: superAdminRole()})
	require.NoError(t, err)

	// wait until bob shows up in admin list
	require.NoError(t, backoff.Retry(func() error {
		resp, err := aliceClient.GetAdmins(aliceClient.Ctx(), &auth.GetAdminsRequest{})
		require.NoError(t, err)
		return require.EqualOrErr(
			admins(admin, gh(bob))(), resp.Admins,
		)
	}, backoff.NewTestingBackOff()))

	// now bob can read from the repo
	buf.Reset()
	require.NoError(t, bobClient.GetFile(repo, "master", "/file", 0, 0, buf))
	require.Matches(t, "test data", buf.String())

	// bob can write to the repo
	commit, err = bobClient.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, bobClient.FinishCommit(repo, commit.ID))
	require.Equal(t, 2, CommitCnt(t, aliceClient, repo)) // check that a new commit was created

	// bob can update the repo's ACL
	_, err = bobClient.SetScope(bobClient.Ctx(), &auth.SetScopeRequest{
		Repo:     repo,
		Username: "carol",
		Scope:    auth.Scope_READER,
	})
	require.NoError(t, err)
	// check that ACL was updated
	require.ElementsEqual(t,
		entries(alice, "owner", "carol", "reader"), getACL(t, aliceClient, repo))

	// 'admin' revokes bob's admin status
	_, err = adminClient.ModifyAdmins(adminClient.Ctx(),
		&auth.ModifyAdminsRequest{Principal: gh(bob), Roles: &auth.AdminRoles{Roles: []auth.AdminRoles_Role{}}})
	require.NoError(t, err)

	// wait until bob is not in admin list
	backoff.Retry(func() error {
		resp, err := aliceClient.GetAdmins(aliceClient.Ctx(), &auth.GetAdminsRequest{})
		require.NoError(t, err)
		return require.EqualOrErr(admins(admin)(), resp.Admins)
	}, backoff.NewTestingBackOff())

	// bob can no longer read from the repo
	buf.Reset()
	err = bobClient.GetFile(repo, "master", "/file", 0, 0, buf)
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())

	// bob cannot write to the repo
	_, err = bobClient.StartCommit(repo, "master")
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.Equal(t, 2, CommitCnt(t, aliceClient, repo)) // check that no commits were created

	// bob can't update the ACL
	_, err = bobClient.SetScope(bobClient.Ctx(), &auth.SetScopeRequest{
		Repo:     repo,
		Username: "carol",
		Scope:    auth.Scope_WRITER,
	})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	// check that ACL wasn't updated
	require.ElementsEqual(t,
		entries(alice, "owner", "carol", "reader"), getACL(t, aliceClient, repo))
}

// TestFSAdminRWO tests adding and removing cluster FS admins, as well as FS admins
// reading, writing, and moderating (owning) all repos in the cluster.
func TestFSAdminRWO(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	defer deleteAll(t)
	alice, bob := tu.UniqueString("alice"), tu.UniqueString("bob")
	aliceClient, bobClient := getPachClient(t, alice), getPachClient(t, bob)
	adminClient := getPachClient(t, admin)

	// The initial set of admins is just the user "admin"
	resp, err := aliceClient.GetAdmins(aliceClient.Ctx(), &auth.GetAdminsRequest{})
	require.NoError(t, err)
	require.Equal(t, admins(admin)(), resp.Admins)

	// alice creates a repo (that only she owns) and puts a file
	repo := tu.UniqueString("TestAdminRWO")
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.Equal(t, entries(alice, "owner"), getACL(t, aliceClient, repo))
	commit, err := aliceClient.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = aliceClient.PutFile(repo, commit.ID, "/file", strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.ID))

	// bob cannot read from the repo
	buf := &bytes.Buffer{}
	err = bobClient.GetFile(repo, "master", "/file", 0, 0, buf)
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
	_, err = bobClient.SetScope(bobClient.Ctx(), &auth.SetScopeRequest{
		Repo:     repo,
		Username: "carol",
		Scope:    auth.Scope_READER,
	})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.Equal(t, entries(alice, "owner"), getACL(t, aliceClient, repo)) // check that ACL wasn't updated

	// 'admin' makes bob an fs admin
	_, err = adminClient.ModifyAdmins(adminClient.Ctx(),
		&auth.ModifyAdminsRequest{Principal: gh(bob), Roles: fsAdminRole()})
	require.NoError(t, err)

	// wait until bob shows up in admin list
	require.NoError(t, backoff.Retry(func() error {
		resp, err := aliceClient.GetAdmins(aliceClient.Ctx(), &auth.GetAdminsRequest{})
		require.NoError(t, err)
		return require.EqualOrErr(
			admins(admin)(gh(bob)), resp.Admins,
		)
	}, backoff.NewTestingBackOff()))

	// now bob can read from the repo
	buf.Reset()
	require.NoError(t, bobClient.GetFile(repo, "master", "/file", 0, 0, buf))
	require.Matches(t, "test data", buf.String())

	// bob can write to the repo
	commit, err = bobClient.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, bobClient.FinishCommit(repo, commit.ID))
	require.Equal(t, 2, CommitCnt(t, aliceClient, repo)) // check that a new commit was created

	// bob can update the repo's ACL
	_, err = bobClient.SetScope(bobClient.Ctx(), &auth.SetScopeRequest{
		Repo:     repo,
		Username: "carol",
		Scope:    auth.Scope_READER,
	})
	require.NoError(t, err)
	// check that ACL was updated
	require.ElementsEqual(t,
		entries(alice, "owner", "carol", "reader"), getACL(t, aliceClient, repo))

	// 'admin' revokes bob's admin status
	_, err = adminClient.ModifyAdmins(adminClient.Ctx(),
		&auth.ModifyAdminsRequest{Principal: gh(bob), Roles: &auth.AdminRoles{Roles: []auth.AdminRoles_Role{}}})
	require.NoError(t, err)

	// wait until bob is not in admin list
	backoff.Retry(func() error {
		resp, err := aliceClient.GetAdmins(aliceClient.Ctx(), &auth.GetAdminsRequest{})
		require.NoError(t, err)
		return require.EqualOrErr(admins(admin)(), resp.Admins)
	}, backoff.NewTestingBackOff())

	// bob can no longer read from the repo
	buf.Reset()
	err = bobClient.GetFile(repo, "master", "/file", 0, 0, buf)
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())

	// bob cannot write to the repo
	_, err = bobClient.StartCommit(repo, "master")
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.Equal(t, 2, CommitCnt(t, aliceClient, repo)) // check that no commits were created

	// bob can't update the ACL
	_, err = bobClient.SetScope(bobClient.Ctx(), &auth.SetScopeRequest{
		Repo:     repo,
		Username: "carol",
		Scope:    auth.Scope_WRITER,
	})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	// check that ACL wasn't updated
	require.ElementsEqual(t,
		entries(alice, "owner", "carol", "reader"), getACL(t, aliceClient, repo))
}

// TestFSAdminFixBrokenRepo tests that an FS admin can modify the ACL of a repo even
// when the repo's ACL is empty (indicating that no user has explicit access to
// to the repo)
func TestFSAdminFixBrokenRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	defer deleteAll(t)
	alice, bob := tu.UniqueString("alice"), tu.UniqueString("bob")
	aliceClient, bobClient := getPachClient(t, alice), getPachClient(t, bob)
	adminClient := getPachClient(t, admin)

	// alice creates a repo (that only she owns) and puts a file
	repo := tu.UniqueString("TestAdmin")
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.Equal(t, entries(alice, "owner"), getACL(t, aliceClient, repo))

	// 'admin' makes bob an FS admin
	_, err := adminClient.ModifyAdmins(adminClient.Ctx(),
		&auth.ModifyAdminsRequest{Principal: gh(bob), Roles: fsAdminRole()})
	require.NoError(t, err)

	// wait until bob shows up in admin list
	require.NoError(t, backoff.Retry(func() error {
		resp, err := aliceClient.GetAdmins(aliceClient.Ctx(), &auth.GetAdminsRequest{})
		require.NoError(t, err)
		return require.EqualOrErr(
			admins(admin)(gh(bob)), resp.Admins,
		)
	}, backoff.NewTestingBackOff()))

	// admin deletes the repo's ACL
	_, err = adminClient.AuthAPIClient.SetACL(adminClient.Ctx(),
		&auth.SetACLRequest{
			Repo:    repo,
			Entries: nil,
		})
	require.NoError(t, err)

	// Check that the ACL is empty
	require.Equal(t, entries(), getACL(t, adminClient, repo))

	// alice cannot write to the repo
	_, err = aliceClient.StartCommit(repo, "master")
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.Equal(t, 0, CommitCnt(t, adminClient, repo)) // check that no commits were created

	// bob, an FS admin, can update the ACL to put Alice back, even though reading the ACL
	// will fail
	_, err = bobClient.SetACL(bobClient.Ctx(),
		&auth.SetACLRequest{
			Repo: repo,
			Entries: []*auth.ACLEntry{
				{Username: alice, Scope: auth.Scope_OWNER},
			},
		})
	require.NoError(t, err)
	require.Equal(t, entries(alice, "owner"), getACL(t, aliceClient, repo))

	// now alice can write to the repo
	commit, err := aliceClient.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = aliceClient.PutFile(repo, commit.ID, "/file", strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.ID))
	require.Equal(t, 1, CommitCnt(t, adminClient, repo)) // check that a new commit was created
}

// TestCannotRemoveAllClusterAdmins tests that trying to remove all of a
// clusters admins yields an error
func TestCannotRemoveAllClusterAdmins(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	defer deleteAll(t)
	alice := tu.UniqueString("alice")
	aliceClient, adminClient := getPachClient(t, alice), getPachClient(t, admin)

	// Check that the initial set of admins is just "admin"
	resp, err := adminClient.GetAdmins(adminClient.Ctx(), &auth.GetAdminsRequest{})
	require.NoError(t, err)
	require.Equal(t, admins(admin)(), resp.Admins)

	// admin cannot remove themselves from the list of super admins (otherwise
	// there would be no admins)
	_, err = adminClient.ModifyAdmins(adminClient.Ctx(),
		&auth.ModifyAdminsRequest{Principal: admin})
	require.YesError(t, err)
	require.NoError(t, backoff.Retry(func() error {
		resp, err = adminClient.GetAdmins(adminClient.Ctx(), &auth.GetAdminsRequest{})
		require.NoError(t, err)
		return require.EqualOrErr(admins(admin)(), resp.Admins)
	}, backoff.NewTestingBackOff()))

	// admin can make alice an FS administrator but admin is still the only super admin
	_, err = adminClient.ModifyAdmins(adminClient.Ctx(),
		&auth.ModifyAdminsRequest{
			Principal: gh(alice),
			Roles:     fsAdminRole(),
		})
	require.NoError(t, err)
	require.NoError(t, backoff.Retry(func() error {
		resp, err = adminClient.GetAdmins(adminClient.Ctx(), &auth.GetAdminsRequest{})
		require.NoError(t, err)
		return require.EqualOrErr(admins(admin)(gh(alice)), resp.Admins)
	}, backoff.NewTestingBackOff()))

	// admin cannot remove themselves from the list of super admins (otherwise
	// there would be no admins), alice only has fs admin
	_, err = adminClient.ModifyAdmins(adminClient.Ctx(),
		&auth.ModifyAdminsRequest{Principal: admin})
	require.YesError(t, err)
	require.NoError(t, backoff.Retry(func() error {
		resp, err = adminClient.GetAdmins(adminClient.Ctx(), &auth.GetAdminsRequest{})
		require.NoError(t, err)
		return require.EqualOrErr(admins(admin)(gh(alice)), resp.Admins)
	}, backoff.NewTestingBackOff()))

	// admin can make alice a cluster administrator
	_, err = adminClient.ModifyAdmins(adminClient.Ctx(),
		&auth.ModifyAdminsRequest{
			Principal: gh(alice),
			Roles:     superAdminRole(),
		})
	require.NoError(t, err)
	require.NoError(t, backoff.Retry(func() error {
		resp, err = adminClient.GetAdmins(adminClient.Ctx(), &auth.GetAdminsRequest{})
		require.NoError(t, err)
		return require.EqualOrErr(admins(admin, gh(alice))(), resp.Admins)
	}, backoff.NewTestingBackOff()))

	// Now admin can remove themselves as a cluster admin
	_, err = adminClient.ModifyAdmins(adminClient.Ctx(),
		&auth.ModifyAdminsRequest{Principal: admin})
	require.NoError(t, err)
	require.NoError(t, backoff.Retry(func() error {
		resp, err = adminClient.GetAdmins(adminClient.Ctx(), &auth.GetAdminsRequest{})
		require.NoError(t, err)
		return require.EqualOrErr(admins(gh(alice))(), resp.Admins)
	}, backoff.NewTestingBackOff()))

	// now alice is the only admin, and she cannot remove herself as a cluster
	// administrator
	_, err = aliceClient.ModifyAdmins(aliceClient.Ctx(),
		&auth.ModifyAdminsRequest{Principal: alice})
	require.YesError(t, err)
	require.NoError(t, backoff.Retry(func() error {
		resp, err = aliceClient.GetAdmins(aliceClient.Ctx(), &auth.GetAdminsRequest{})
		require.NoError(t, err)
		return require.EqualOrErr(admins(gh(alice))(), resp.Admins)
	}, backoff.NewTestingBackOff()))
}

// TestModifyClusterAdminsAllowRobotOnlyAdmin tests the fix to
// https://github.com/pachyderm/pachyderm/issues/3010
// Basically, ModifyAdmins should not return an error if the only cluster admin
// is a robot user
func TestModifyClusterAdminsAllowRobotOnlyAdmin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	defer deleteAll(t)
	adminClient := getPachClient(t, admin)

	// Check that the initial set of admins is just "admin"
	resp, err := adminClient.GetAdmins(adminClient.Ctx(), &auth.GetAdminsRequest{})
	require.NoError(t, err)
	require.Equal(t, admins(admin)(), resp.Admins)

	// 'admin' gets credentials for a robot user, and swaps themself and the robot
	// so that the only cluster administrator is the robot user
	tokenResp, err := adminClient.GetAuthToken(adminClient.Ctx(),
		&auth.GetAuthTokenRequest{
			Subject: robot("rob"),
			TTL:     3600,
		})
	require.NoError(t, err)
	// copy client & use resp token
	robotClient := adminClient.WithCtx(context.Background())
	robotClient.SetAuthToken(tokenResp.Token)

	// Make the robot an admin
	_, err = adminClient.ModifyAdmins(adminClient.Ctx(),
		&auth.ModifyAdminsRequest{
			Principal: robot("rob"),
			Roles:     superAdminRole(),
		})
	require.NoError(t, err)
	require.NoError(t, backoff.Retry(func() error {
		resp, err = adminClient.GetAdmins(adminClient.Ctx(), &auth.GetAdminsRequest{})
		require.NoError(t, err)
		return require.EqualOrErr(admins(robot("rob"), admin)(), resp.Admins)
	}, backoff.NewTestingBackOff()))

	// Remove admin
	_, err = adminClient.ModifyAdmins(adminClient.Ctx(),
		&auth.ModifyAdminsRequest{
			Principal: admin,
		})
	require.NoError(t, err)
	require.NoError(t, backoff.Retry(func() error {
		resp, err = adminClient.GetAdmins(adminClient.Ctx(), &auth.GetAdminsRequest{})
		require.NoError(t, err)
		return require.EqualOrErr(admins(robot("rob"))(), resp.Admins)
	}, backoff.NewTestingBackOff()))

	// The robot user adds admin back as a cluster admin
	_, err = robotClient.ModifyAdmins(robotClient.Ctx(),
		&auth.ModifyAdminsRequest{
			Principal: admin,
			Roles:     superAdminRole(),
		})
	require.NoError(t, err)
	require.NoError(t, backoff.Retry(func() error {
		resp, err = adminClient.GetAdmins(adminClient.Ctx(), &auth.GetAdminsRequest{})
		require.NoError(t, err)
		return require.EqualOrErr(admins(robot("rob"), admin)(), resp.Admins)
	}, backoff.NewTestingBackOff()))
}

func TestPreActivationPipelinesKeepRunningAfterActivation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	defer deleteAll(t)
	alice := tu.UniqueString("alice")
	aliceClient, adminClient := getPachClient(t, alice), getPachClient(t, admin)

	// Deactivate auth
	_, err := adminClient.Deactivate(adminClient.Ctx(), &auth.DeactivateRequest{})
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
	_, err = aliceClient.PutFile(repo, commit.ID, "/file1", strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.ID))

	// make sure the pipeline runs
	iter, err := aliceClient.FlushCommit(
		[]*pfs.Commit{commit},
		[]*pfs.Repo{{Name: pipeline}},
	)
	require.NoError(t, err)
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := iter.Next()
		return err
	})

	// activate auth
	resp, err := adminClient.Activate(adminClient.Ctx(), &auth.ActivateRequest{
		Subject: admin,
	})
	require.NoError(t, err)
	tokenMap[admin] = resp.PachToken
	adminClient.SetAuthToken(resp.PachToken)

	// re-authenticate, as old tokens were deleted
	aliceResp, err := aliceClient.Authenticate(context.Background(), &auth.AuthenticateRequest{
		GitHubToken: alice,
	})
	require.NoError(t, err)
	tokenMap[alice] = aliceResp.PachToken
	aliceClient.SetAuthToken(aliceResp.PachToken)

	// Make sure alice cannot read the input repo (i.e. if the pipeline runs as
	// alice, it will fail)
	buf := &bytes.Buffer{}
	err = aliceClient.GetFile(repo, "master", "/file1", 0, 0, buf)
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())

	// Admin creates an input commit
	commit, err = adminClient.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = adminClient.PutFile(repo, commit.ID, "/file2", strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, adminClient.FinishCommit(repo, commit.ID))

	// make sure the pipeline still runs (i.e. it's not running as alice)
	iter, err = adminClient.FlushCommit(
		[]*pfs.Commit{commit},
		[]*pfs.Repo{{Name: pipeline}},
	)
	require.NoError(t, err)
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := iter.Next()
		return err
	})
}

func TestExpirationRepoOnlyAccessibleToAdmins(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	defer deleteAll(t)
	alice := tu.UniqueString("alice")
	aliceClient, adminClient := getPachClient(t, alice), getPachClient(t, admin)

	// alice creates a repo
	repo := tu.UniqueString("TestExpirationRepoOnlyAccessibleToAdmins")
	require.NoError(t, aliceClient.CreateRepo(repo))

	// alice creates a commit
	commit, err := aliceClient.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = aliceClient.PutFile(repo, commit.ID, "/file1", strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.ID))
	require.Equal(t, 1, CommitCnt(t, aliceClient, repo))

	// Make current enterprise token expire
	adminClient.Enterprise.Activate(adminClient.Ctx(),
		&enterprise.ActivateRequest{
			ActivationCode: tu.GetTestEnterpriseCode(t),
			Expires:        TSProtoOrDie(t, time.Now().Add(-30*time.Second)),
		})
	// wait for Enterprise token to expire
	require.NoError(t, backoff.Retry(func() error {
		resp, err := adminClient.Enterprise.GetState(adminClient.Ctx(),
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
	err = aliceClient.GetFile(repo, "master", "/file1", 0, 0, buf)
	require.YesError(t, err)
	require.Matches(t, "not active", err.Error())

	// alice can't write to the repo
	_, err = aliceClient.StartCommit(repo, "master")
	require.YesError(t, err)
	require.Matches(t, "not active", err.Error())
	require.Equal(t, 1, CommitCnt(t, adminClient, repo)) // check that no commits were created

	// alice can't update the ACL
	_, err = aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Repo:     repo,
		Username: "carol",
		Scope:    auth.Scope_READER,
	})
	require.YesError(t, err)
	require.Matches(t, "not active", err.Error())
	// We don't delete the ACL because the user might re-enable enterprise pachyderm
	require.Equal(t, entries(alice, "owner"), getACL(t, adminClient, repo)) // check that ACL wasn't updated

	// alice also can't re-authenticate
	_, err = aliceClient.Authenticate(context.Background(),
		&auth.AuthenticateRequest{GitHubToken: alice})
	require.YesError(t, err)
	require.Matches(t, "not active", err.Error())

	// admin can read from the repo
	buf.Reset()
	require.NoError(t, adminClient.GetFile(repo, "master", "/file1", 0, 0, buf))
	require.Matches(t, "test data", buf.String())

	// admin can write to the repo
	commit, err = adminClient.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = adminClient.PutFile(repo, commit.ID, "/file2", strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, adminClient.FinishCommit(repo, commit.ID))
	require.Equal(t, 2, CommitCnt(t, adminClient, repo)) // check that a new commit was created

	// admin can update the repo's ACL
	_, err = adminClient.SetScope(adminClient.Ctx(), &auth.SetScopeRequest{
		Repo:     repo,
		Username: "carol",
		Scope:    auth.Scope_READER,
	})
	require.NoError(t, err)
	// check that ACL was updated
	require.ElementsEqual(t,
		entries(alice, "owner", "carol", "reader"), getACL(t, adminClient, repo))

	// Re-enable enterprise
	year := 365 * 24 * time.Hour
	adminClient.Enterprise.Activate(adminClient.Ctx(),
		&enterprise.ActivateRequest{
			ActivationCode: tu.GetTestEnterpriseCode(t),
			// This will stop working some time in 2026
			Expires: TSProtoOrDie(t, time.Now().Add(year)),
		})
	// wait for Enterprise token to re-enable
	require.NoError(t, backoff.Retry(func() error {
		resp, err := adminClient.Enterprise.GetState(adminClient.Ctx(),
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
	resp, err := aliceClient.Authenticate(context.Background(),
		&auth.AuthenticateRequest{GitHubToken: alice})
	require.NoError(t, err)
	aliceClient.SetAuthToken(resp.PachToken)

	// alice can read from the repo again
	buf = &bytes.Buffer{}
	require.NoError(t, aliceClient.GetFile(repo, "master", "/file1", 0, 0, buf))
	require.Matches(t, "test data", buf.String())

	// alice can write to the repo again
	commit, err = aliceClient.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = aliceClient.PutFile(repo, commit.ID, "/file3", strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.ID))
	require.Equal(t, 3, CommitCnt(t, aliceClient, repo)) // check that a new commit was created

	// alice can update the ACL again
	_, err = aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Repo:     repo,
		Username: "carol",
		Scope:    auth.Scope_WRITER,
	})
	require.NoError(t, err)
	// check that ACL was updated
	require.ElementsEqual(t,
		entries(alice, "owner", "carol", "writer"), getACL(t, adminClient, repo))
}

func TestPipelinesRunAfterExpiration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	defer deleteAll(t)
	alice := tu.UniqueString("alice")
	aliceClient, adminClient := getPachClient(t, alice), getPachClient(t, admin)

	// alice creates a repo
	repo := tu.UniqueString("TestPipelinesRunAfterExpiration")
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.Equal(t, entries(alice, "owner"), getACL(t, aliceClient, repo))

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
	require.ElementsEqual(t,
		entries(alice, "owner", pl(pipeline), "writer"), getACL(t, aliceClient, pipeline))

	// Make sure alice's pipeline runs successfully
	commit, err := aliceClient.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = aliceClient.PutFile(repo, commit.ID, tu.UniqueString("/file1"),
		strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.ID))
	iter, err := aliceClient.FlushCommit(
		[]*pfs.Commit{commit},
		[]*pfs.Repo{{Name: pipeline}},
	)
	require.NoError(t, err)
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := iter.Next()
		return err
	})

	// Make current enterprise token expire
	adminClient.Enterprise.Activate(adminClient.Ctx(),
		&enterprise.ActivateRequest{
			ActivationCode: tu.GetTestEnterpriseCode(t),
			Expires:        TSProtoOrDie(t, time.Now().Add(-30*time.Second)),
		})
	// wait for Enterprise token to expire
	require.NoError(t, backoff.Retry(func() error {
		resp, err := adminClient.Enterprise.GetState(adminClient.Ctx(),
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
	commit, err = adminClient.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = adminClient.PutFile(repo, commit.ID, tu.UniqueString("/file2"),
		strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, adminClient.FinishCommit(repo, commit.ID))
	iter, err = adminClient.FlushCommit(
		[]*pfs.Commit{commit},
		[]*pfs.Repo{{Name: pipeline}},
	)
	require.NoError(t, err)
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := iter.Next()
		return err
	})
}

// Tests that GetAcl, SetAcl, GetScope, and SetScope all respect expired
// Enterprise tokens (i.e. reject non-admin requests once the token is expired,
// and allow admin requests)
func TestGetSetScopeAndAclWithExpiredToken(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	defer deleteAll(t)
	alice := tu.UniqueString("alice")
	aliceClient, adminClient := getPachClient(t, alice), getPachClient(t, admin)

	// alice creates a repo
	repo := tu.UniqueString("TestGetSetScopeAndAclWithExpiredToken")
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.Equal(t, entries(alice, "owner"), getACL(t, aliceClient, repo))

	// Make current enterprise token expire
	adminClient.Enterprise.Activate(adminClient.Ctx(),
		&enterprise.ActivateRequest{
			ActivationCode: tu.GetTestEnterpriseCode(t),
			Expires:        TSProtoOrDie(t, time.Now().Add(-30*time.Second)),
		})
	// wait for Enterprise token to expire
	require.NoError(t, backoff.Retry(func() error {
		resp, err := adminClient.Enterprise.GetState(adminClient.Ctx(),
			&enterprise.GetStateRequest{})
		if err != nil {
			return err
		}
		if resp.State == enterprise.State_ACTIVE {
			return errors.New("Pachyderm Enterprise is still active")
		}
		return nil
	}, backoff.NewTestingBackOff()))

	// alice can't call GetScope on repo, even though she owns it
	_, err := aliceClient.GetScope(aliceClient.Ctx(), &auth.GetScopeRequest{
		Repos: []string{repo},
	})
	require.YesError(t, err)
	require.Matches(t, "not active", err.Error())

	// alice can't call SetScope on repo
	_, err = aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Repo:     repo,
		Username: "carol",
		Scope:    auth.Scope_READER,
	})
	require.YesError(t, err)
	require.Matches(t, "not active", err.Error())
	require.Equal(t, entries(alice, "owner"), getACL(t, adminClient, repo))

	// alice can't call GetAcl on repo
	_, err = aliceClient.GetACL(aliceClient.Ctx(), &auth.GetACLRequest{
		Repo: repo,
	})
	require.YesError(t, err)
	require.Matches(t, "not active", err.Error())

	// alice can't call GetAcl on repo
	_, err = aliceClient.SetACL(aliceClient.Ctx(), &auth.SetACLRequest{
		Repo: repo,
		Entries: []*auth.ACLEntry{
			{Username: alice, Scope: auth.Scope_OWNER},
			{Username: "carol", Scope: auth.Scope_READER},
		},
	})
	require.YesError(t, err)
	require.Matches(t, "not active", err.Error())
	require.Equal(t, entries(alice, "owner"), getACL(t, adminClient, repo))

	// admin *can* call GetScope on repo
	resp, err := adminClient.GetScope(adminClient.Ctx(), &auth.GetScopeRequest{
		Repos: []string{repo},
	})
	require.NoError(t, err)
	require.Equal(t, []auth.Scope{auth.Scope_NONE}, resp.Scopes)

	// admin can call SetScope on repo
	_, err = adminClient.SetScope(adminClient.Ctx(), &auth.SetScopeRequest{
		Repo:     repo,
		Username: "carol",
		Scope:    auth.Scope_READER,
	})
	require.NoError(t, err)
	require.ElementsEqual(t,
		entries(alice, "owner", "carol", "reader"), getACL(t, adminClient, repo))

	// admin can call GetAcl on repo
	aclResp, err := adminClient.GetACL(adminClient.Ctx(), &auth.GetACLRequest{
		Repo: repo,
	})
	require.NoError(t, err)
	aclEntries := make([]aclEntry, 0, len(aclResp.Entries))
	for _, e := range aclResp.Entries {
		aclEntries = append(aclEntries, aclEntry{
			Username: e.Username,
			Scope:    e.Scope,
		})
	}
	require.ElementsEqual(t,
		entries(alice, "owner", "carol", "reader"), aclEntries)

	// admin can call SetAcl on repo
	_, err = adminClient.SetACL(adminClient.Ctx(), &auth.SetACLRequest{
		Repo: repo,
		Entries: []*auth.ACLEntry{
			{Username: alice, Scope: auth.Scope_OWNER},
			{Username: "carol", Scope: auth.Scope_WRITER},
		},
	})
	require.NoError(t, err)
	require.ElementsEqual(t,
		entries(alice, "owner", "carol", "writer"), getACL(t, adminClient, repo))
}

// TestAdminWhoAmI tests that when an admin calls WhoAmI(), the AdminRoles reflects their admin roles
func TestAdminWhoAmI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	defer deleteAll(t)
	alice, bob := tu.UniqueString("alice"), tu.UniqueString("bob")
	aliceClient, bobClient := getPachClient(t, alice), getPachClient(t, bob)
	adminClient := getPachClient(t, admin)

	// 'admin' makes bob an FS admin
	_, err := adminClient.ModifyAdmins(adminClient.Ctx(),
		&auth.ModifyAdminsRequest{Principal: gh(bob), Roles: fsAdminRole()})
	require.NoError(t, err)

	// wait until bob shows up in admin list
	require.NoError(t, backoff.Retry(func() error {
		resp, err := aliceClient.GetAdmins(aliceClient.Ctx(), &auth.GetAdminsRequest{})
		require.NoError(t, err)
		return require.EqualOrErr(
			admins(admin)(gh(bob)), resp.Admins,
		)
	}, backoff.NewTestingBackOff()))

	// alice has no admin roles
	resp, err := aliceClient.WhoAmI(aliceClient.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, gh(alice), resp.Username)
	require.Equal(t, 0, len(resp.AdminRoles.Roles))

	// admin has super admin
	resp, err = adminClient.WhoAmI(adminClient.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, admin, resp.Username)
	require.Equal(t, superAdminRole(), resp.AdminRoles)

	// bob has FS admin
	resp, err = bobClient.WhoAmI(bobClient.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, gh(bob), resp.Username)
	require.Equal(t, fsAdminRole(), resp.AdminRoles)
}

// TestListRepoAdminIsOwnerOfAllRepos tests that when an admin calls ListRepo,
// the result indicates that they're an owner of every repo in the cluster
// (needed by the Pachyderm dashboard)
func TestListRepoAdminIsOwnerOfAllRepos(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	defer deleteAll(t)
	// t.Parallel()
	adminClient := getPachClient(t, admin)
	alice, bob := tu.UniqueString("alice"), tu.UniqueString("bob")
	aliceClient, bobClient := getPachClient(t, alice), getPachClient(t, bob)

	// alice creates a repo
	repoWriter := tu.UniqueString("TestListRepoAdminIsOwnerOfAllRepos")
	require.NoError(t, aliceClient.CreateRepo(repoWriter))

	// bob calls ListRepo, but has NONE access to all repos
	infos, err := bobClient.ListRepo()
	require.NoError(t, err)
	for _, info := range infos {
		require.Equal(t, auth.Scope_NONE, info.AuthInfo.AccessLevel)
	}

	// admin calls ListRepo, and has OWNER access to all repos
	infos, err = adminClient.ListRepo()
	require.NoError(t, err)
	for _, info := range infos {
		require.Equal(t, auth.Scope_OWNER, info.AuthInfo.AccessLevel)
	}
}

// TestGetAuthToken tests that an admin can manufacture auth credentials for
// arbitrary other users
func TestGetAuthToken(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	defer deleteAll(t)
	adminClient := getPachClient(t, admin)

	// Generate first set of auth credentials
	robotUser := robot(tu.UniqueString("optimus_prime"))
	resp, err := adminClient.GetAuthToken(adminClient.Ctx(),
		&auth.GetAuthTokenRequest{Subject: robotUser})
	require.NoError(t, err)
	token1 := resp.Token
	robotClient1 := getPachClient(t, "")
	robotClient1.SetAuthToken(token1)

	// Confirm identity tied to 'token1'
	who, err := robotClient1.WhoAmI(robotClient1.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, robotUser, who.Username)
	require.Equal(t, 0, len(who.AdminRoles.Roles))
	if version.IsAtLeast(1, 10) {
		require.True(t, who.TTL >= 0 && who.TTL < secsInYear)
	} else {
		require.Equal(t, -1, who.TTL)
	}

	// Generate a second set of auth credentials--confirm that a unique token is
	// generated, but that the identity tied to it is the same
	resp, err = adminClient.GetAuthToken(adminClient.Ctx(),
		&auth.GetAuthTokenRequest{Subject: robotUser})
	require.NoError(t, err)
	token2 := resp.Token
	require.NotEqual(t, token1, token2)
	robotClient2 := getPachClient(t, "")
	robotClient2.SetAuthToken(token2)

	// Confirm identity tied to 'token1'
	who, err = robotClient2.WhoAmI(robotClient2.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, robotUser, who.Username)
	require.Equal(t, 0, len(who.AdminRoles.Roles))
	if version.IsAtLeast(1, 10) {
		require.True(t, who.TTL >= 0 && who.TTL < secsInYear)
	} else {
		require.Equal(t, -1, who.TTL)
	}

	// robotClient1 creates a repo
	repo := tu.UniqueString("TestPipelinesRunAfterExpiration")
	require.NoError(t, robotClient1.CreateRepo(repo))
	require.Equal(t, entries(robotUser, "owner"), getACL(t, robotClient1, repo))

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
	require.ElementsEqual(t,
		entries(robotUser, "owner", pl(pipeline), "writer"), getACL(t, robotClient1, pipeline))

	// Make sure that robotClient2 can commit to the input repo and flush their
	// input commit
	commit, err := robotClient2.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = robotClient2.PutFile(repo, commit.ID, tu.UniqueString("/file1"),
		strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, robotClient2.FinishCommit(repo, commit.ID))
	iter, err := robotClient2.FlushCommit(
		[]*pfs.Commit{commit},
		[]*pfs.Repo{{Name: pipeline}},
	)
	require.NoError(t, err)
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := iter.Next()
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
	iter, err = robotClient2.FlushCommit(
		[]*pfs.Commit{commit},
		[]*pfs.Repo{{Name: pipeline}},
	)
	require.NoError(t, err)
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := iter.Next()
		return err
	})
}

// TestGetIndefiniteAuthToken tests that an admin can generate an auth token that never
// times out if explicitly requested (e.g. for a daemon)
func TestGetIndefiniteAuthToken(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	defer deleteAll(t)
	adminClient := getPachClient(t, admin)

	// Generate auth credentials
	robotUser := robot(tu.UniqueString("rock-em-sock-em"))
	resp, err := adminClient.GetAuthToken(adminClient.Ctx(),
		&auth.GetAuthTokenRequest{
			Subject: robotUser,
			TTL:     -1,
		})
	require.NoError(t, err)
	token1 := resp.Token
	robotClient1 := getPachClient(t, "")
	robotClient1.SetAuthToken(token1)

	// Confirm identity tied to 'token1'
	who, err := robotClient1.WhoAmI(robotClient1.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, robotUser, who.Username)
	require.False(t, who.IsAdmin)
	require.Equal(t, int64(-1), who.TTL)
}

// TestRobotUserWhoAmI tests that robot users can call WhoAmI and get a response
// with the right prefix
func TestRobotUserWhoAmI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	defer deleteAll(t)
	adminClient := getPachClient(t, admin)

	// Generate a robot user auth credential, and create a client for that user
	robotUser := robot(tu.UniqueString("r2d2"))
	resp, err := adminClient.GetAuthToken(adminClient.Ctx(),
		&auth.GetAuthTokenRequest{Subject: robotUser})
	require.NoError(t, err)
	// copy client & use resp token
	robotClient := adminClient.WithCtx(context.Background())
	robotClient.SetAuthToken(resp.Token)

	who, err := robotClient.WhoAmI(robotClient.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, robotUser, who.Username)
	require.True(t, strings.HasPrefix(who.Username, auth.RobotPrefix))
	require.Equal(t, 0, len(who.AdminRoles.Roles))
}

// TestRobotUserACL tests that a robot user can create a repo, add github users
// to their repo, and be added to github user's repo.
func TestRobotUserACL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	alice := tu.UniqueString("alice")
	aliceClient, adminClient := getPachClient(t, alice), getPachClient(t, admin)

	// Generate a robot user auth credential, and create a client for that user
	robotUser := robot(tu.UniqueString("voltron"))
	resp, err := adminClient.GetAuthToken(adminClient.Ctx(),
		&auth.GetAuthTokenRequest{Subject: robotUser})
	require.NoError(t, err)
	// copy client & use resp token
	robotClient := adminClient.WithCtx(context.Background())
	robotClient.SetAuthToken(resp.Token)

	// robotUser creates a repo and adds alice as a writer
	repo := tu.UniqueString("TestRobotUserACL")
	require.NoError(t, robotClient.CreateRepo(repo))
	require.Equal(t, entries(robotUser, "owner"), getACL(t, robotClient, repo))
	robotClient.SetScope(robotClient.Ctx(), &auth.SetScopeRequest{
		Repo:     repo,
		Scope:    auth.Scope_WRITER,
		Username: auth.GitHubPrefix + alice,
	})
	require.ElementsEqual(t,
		entries(alice, "writer", robotUser, "owner"), getACL(t, robotClient, repo))

	// test that alice can commit to the robot user's repo
	commit, err := aliceClient.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.ID))

	// Now alice creates a repo, and adds robotUser as a writer
	repo2 := tu.UniqueString("TestRobotUserACL")
	require.NoError(t, aliceClient.CreateRepo(repo2))
	require.Equal(t, entries(alice, "owner"), getACL(t, aliceClient, repo2))
	aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Repo:     repo2,
		Scope:    auth.Scope_WRITER,
		Username: robotUser,
	})
	require.ElementsEqual(t,
		entries(alice, "owner", robotUser, "writer"), getACL(t, aliceClient, repo2))

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
	deleteAll(t)
	defer deleteAll(t)
	alice := tu.UniqueString("alice")
	adminClient := getPachClient(t, admin)
	aliceClient := getPachClient(t, alice)

	// Generate a robot user auth credential, and create a client for that user
	robotUser := robot(tu.UniqueString("bender"))
	resp, err := adminClient.GetAuthToken(adminClient.Ctx(),
		&auth.GetAuthTokenRequest{Subject: robotUser})
	require.NoError(t, err)
	// copy client & use resp token
	robotClient := adminClient.WithCtx(context.Background())
	robotClient.SetAuthToken(resp.Token)

	// make robotUser an admin
	_, err = adminClient.ModifyAdmins(adminClient.Ctx(), &auth.ModifyAdminsRequest{
		Principal: robotUser,
		Roles:     superAdminRole(),
	})
	require.NoError(t, err)
	// wait until robotUser shows up in admin list
	require.NoError(t, backoff.Retry(func() error {
		resp, err := adminClient.GetAdmins(adminClient.Ctx(), &auth.GetAdminsRequest{})
		require.NoError(t, err)
		return require.EqualOrErr(admins(admin, robotUser)(), resp.Admins)
	}, backoff.NewTestingBackOff()))

	// robotUser mints a token for robotUser2
	robotUser2 := robot(tu.UniqueString("robocop"))
	resp, err = robotClient.GetAuthToken(robotClient.Ctx(), &auth.GetAuthTokenRequest{
		Subject: robotUser2,
	})
	require.NoError(t, err)
	require.NotEqual(t, "", resp.Token)
	robotClient2 := adminClient.WithCtx(context.Background())
	robotClient2.SetAuthToken(resp.Token)

	// robotUser2 creates a repo, and robotUser commits to it
	repo := tu.UniqueString("TestRobotUserAdmin")
	require.NoError(t, robotClient2.CreateRepo(repo))
	commit, err := robotClient.StartCommit(repo, "master")
	require.NoError(t, err) // admin privs means robotUser can commit
	require.NoError(t, robotClient.FinishCommit(repo, commit.ID))

	// robotUser adds alice to the repo, and checks that the ACL is updated
	require.ElementsEqual(t,
		entries(robotUser2, "owner"), getACL(t, robotClient, repo))
	_, err = robotClient.SetScope(robotClient.Ctx(), &auth.SetScopeRequest{
		Repo:     repo,
		Scope:    auth.Scope_WRITER,
		Username: alice,
	})
	require.NoError(t, err)
	require.ElementsEqual(t,
		entries(robotUser2, "owner", alice, "writer"), getACL(t, robotClient, repo))
	commit, err = aliceClient.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.ID))

	robotClient.Deactivate(robotClient.Ctx(), &auth.DeactivateRequest{})
	require.NoError(t, err)

	deactivateResp, err := robotClient.AuthAPIClient.Activate(context.Background(),
		&auth.ActivateRequest{Subject: admin})
	require.NoError(t, err)
	tokenMap[admin] = deactivateResp.PachToken
	robotClient.SetAuthToken(deactivateResp.PachToken)
}

// TestTokenTTL tests that an admin can create a token with a TTL for a user,
// and that that token will expire
func TestTokenTTL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	defer deleteAll(t)
	adminClient := getPachClient(t, admin)

	// Create repo (so alice has something to list)
	repo := tu.UniqueString("TestTokenTTL")
	require.NoError(t, adminClient.CreateRepo(repo))

	// Create auth token for alice
	alice := tu.UniqueString("alice")
	resp, err := adminClient.GetAuthToken(adminClient.Ctx(), &auth.GetAuthTokenRequest{
		Subject: alice,
		TTL:     5, // seconds
	})
	require.NoError(t, err)
	aliceClient := adminClient.WithCtx(context.Background())
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
	deleteAll(t)
	defer deleteAll(t)
	adminClient := getPachClient(t, admin)

	// Create repo (so alice has something to list)
	repo := tu.UniqueString("TestTokenTTLExtend")
	require.NoError(t, adminClient.CreateRepo(repo))

	alice := tu.UniqueString("alice")
	resp, err := adminClient.GetAuthToken(adminClient.Ctx(), &auth.GetAuthTokenRequest{
		Subject: alice,
		TTL:     5, // seconds
	})
	require.NoError(t, err)
	aliceClient := adminClient.WithCtx(context.Background())
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
	resp, err = adminClient.GetAuthToken(adminClient.Ctx(), &auth.GetAuthTokenRequest{
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
	_, err = adminClient.ExtendAuthToken(adminClient.Ctx(), &auth.ExtendAuthTokenRequest{
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
	deleteAll(t)
	defer deleteAll(t)
	adminClient := getPachClient(t, admin)

	// Create repo (so alice has something to list)
	repo := tu.UniqueString("TestTokenRevoke")
	require.NoError(t, adminClient.CreateRepo(repo))

	alice := tu.UniqueString("alice")
	resp, err := adminClient.GetAuthToken(adminClient.Ctx(), &auth.GetAuthTokenRequest{
		Subject: alice,
	})
	require.NoError(t, err)
	aliceClient := adminClient.WithCtx(context.Background())
	aliceClient.SetAuthToken(resp.Token)

	// alice's token is valid
	repos, err := aliceClient.ListRepo()
	require.NoError(t, err)
	require.ElementsEqualUnderFn(t, []string{repo}, repos, RepoInfoToName)

	// admin revokes token
	_, err = adminClient.RevokeAuthToken(adminClient.Ctx(), &auth.RevokeAuthTokenRequest{
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
	deleteAll(t)
	defer deleteAll(t)

	alice := tu.UniqueString("alice")
	aliceClient := getPachClient(t, alice)
	resp, err := aliceClient.GetAuthToken(aliceClient.Ctx(), &auth.GetAuthTokenRequest{
		Subject: robot(tu.UniqueString("t-1000")),
	})
	require.Nil(t, resp)
	require.YesError(t, err)
	require.Matches(t, "must be an admin", err.Error())
}

// TestGetAuthTokenErrorFSAdminUser tests that FS admin users can't call
// GetAuthToken on behalf of another user
func TestGetAuthTokenErrorFSAdminUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	defer deleteAll(t)

	alice := tu.UniqueString("alice")
	aliceClient := getPachClient(t, alice)
	adminClient := getPachClient(t, admin)

	// 'admin' makes alice an fs admin
	_, err := adminClient.ModifyAdmins(adminClient.Ctx(),
		&auth.ModifyAdminsRequest{Principal: gh(alice), Roles: fsAdminRole()})
	require.NoError(t, err)

	// wait until alice shows up in admin list
	require.NoError(t, backoff.Retry(func() error {
		resp, err := aliceClient.GetAdmins(aliceClient.Ctx(), &auth.GetAdminsRequest{})
		require.NoError(t, err)
		return require.EqualOrErr(
			admins(admin)(gh(alice)), resp.Admins,
		)
	}, backoff.NewTestingBackOff()))

	// Try to get a token for a robot as alice
	resp, err := aliceClient.GetAuthToken(aliceClient.Ctx(), &auth.GetAuthTokenRequest{
		Subject: robot(tu.UniqueString("t-1000")),
	})
	require.Nil(t, resp)
	require.YesError(t, err)
	require.Matches(t, "must be an admin", err.Error())
}

// TestGetAuthTokenDefaultTTL tests the default TTL of a token returned from
// GetAuthToken
func TestGetAuthTokenDefaultTTL(t *testing.T) {
	// if testing.Short() {
	// 	t.Skip("Skipping integration tests in short mode")
	// }
	// deleteAll(t)
	// defer deleteAll(t)

	// alice := tu.UniqueString("alice")
	// aliceClient := getPachClient(t, alice)
	// resp, err := aliceClient.GetAuthToken(aliceClient.Ctx(), &auth.GetAuthTokenRequest{
	// 	Subject: robot(tu.UniqueString("t-1000")),
	// })
	// require.Nil(t, resp)
	// require.YesError(t, err)
	// require.Matches(t, "must be an admin", err.Error())
}

// TestGetAuthTokenPermanent tests the TTL of a token returned from GetAuthToken
// when request.TTL is -1 (i.e. a permanent token)
func TestGetAuthTokenPermanent(t *testing.T) {
	// if testing.Short() {
	// 	t.Skip("Skipping integration tests in short mode")
	// }
	// deleteAll(t)
	// defer deleteAll(t)

	// alice := tu.UniqueString("alice")
	// aliceClient := getPachClient(t, alice)
	// resp, err := aliceClient.GetAuthToken(aliceClient.Ctx(), &auth.GetAuthTokenRequest{
	// 	Subject: robot(tu.UniqueString("t-1000")),
	// })
	// require.Nil(t, resp)
	// require.YesError(t, err)
	// require.Matches(t, "must be an admin", err.Error())
}

// TestActivateAsRobotUser tests that Pachyderm can be activated such that the
// initial admin is a robot user (i.e. without any human intervention)
func TestActivateAsRobotUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	defer deleteAll(t)

	// Activate Pachyderm Enterprise (if it's not already active)
	require.NoError(t, tu.ActivateEnterprise(t, seedClient))

	client := seedClient.WithCtx(context.Background())
	resp, err := client.Activate(client.Ctx(), &auth.ActivateRequest{
		Subject: robot("deckard"),
	})
	require.NoError(t, err)
	client.SetAuthToken(resp.PachToken)
	who, err := client.WhoAmI(client.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, robot("deckard"), who.Username)

	// Make sure the robot token has no TTL
	require.Equal(t, int64(-1), who.TTL)

	client.Deactivate(client.Ctx(), &auth.DeactivateRequest{})
	require.NoError(t, err)

	resp, err = client.AuthAPIClient.Activate(context.Background(),
		&auth.ActivateRequest{Subject: admin})
	require.NoError(t, err)
	tokenMap[admin] = resp.PachToken
}

func TestActivateMismatchedUsernames(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	defer deleteAll(t)

	// Activate Pachyderm Enterprise (if it's not already active)
	require.NoError(t, tu.ActivateEnterprise(t, seedClient))

	client := seedClient.WithCtx(context.Background())
	_, err := client.Activate(client.Ctx(), &auth.ActivateRequest{
		Subject:     "alice",
		GitHubToken: "bob",
	})
	require.YesError(t, err)
	require.Matches(t, "github:alice", err.Error())
	require.Matches(t, "github:bob", err.Error())
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
	deleteAll(t)
	defer deleteAll(t)

	alice := tu.UniqueString("alice")
	aliceClient, adminClient := getPachClient(t, alice), getPachClient(t, admin)

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
	_, err = aliceClient.PutFile(repo, commit.ID, "/file1", strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.ID))

	// make sure the pipeline runs
	iter, err := aliceClient.FlushCommit(
		[]*pfs.Commit{commit},
		[]*pfs.Repo{{Name: pipeline}},
	)
	require.NoError(t, err)
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := iter.Next()
		return err
	})

	// Deactivate auth
	_, err = adminClient.Deactivate(adminClient.Ctx(), &auth.DeactivateRequest{})
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
	deleteAll(t)
	defer deleteAll(t)
	alice := tu.UniqueString("alice")
	c := getPachClient(t, alice)

	// Create input repo w/ initial commit
	repo := tu.UniqueString(t.Name())
	require.NoError(t, c.CreateRepo(repo))
	_, err := c.PutFile(repo, "master", "/file.1", strings.NewReader("1"))
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
	iter, err := c.FlushCommit(
		[]*pfs.Commit{client.NewCommit(repo, "master")},
		[]*pfs.Repo{client.NewRepo(pipeline)})
	require.NoError(t, err)
	require.NoErrorWithinT(t, 30*time.Second, func() error {
		_, err := iter.Next()
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
	_, err = c.PutFile(repo, "master", "/file.2", strings.NewReader("1"))
	require.NoError(t, err)
	iter, err = c.FlushCommit(
		[]*pfs.Commit{client.NewCommit(repo, "master")},
		[]*pfs.Repo{client.NewRepo(pipeline)})
	require.NoError(t, err)
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := iter.Next()
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
	deleteAll(t)
	defer deleteAll(t)
	alice := tu.UniqueString("alice")
	aliceClient := getPachClient(t, alice)

	// Create input repo w/ initial commit
	repo := tu.UniqueString(t.Name())
	require.NoError(t, aliceClient.CreateRepo(repo))
	_, err := aliceClient.PutFile(repo, "master", "/file.1", strings.NewReader("1"))
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
	_, err = aliceClient.PutFile(repo, "master", "/file.2", strings.NewReader("2"))
	require.NoError(t, err)
	iter, err := aliceClient.FlushCommit(
		[]*pfs.Commit{client.NewCommit(repo, "master")},
		[]*pfs.Repo{client.NewRepo(pipeline)})
	require.NoError(t, err)
	require.NoErrorWithinT(t, 30*time.Second, func() error {
		_, err := iter.Next()
		// TODO(msteffen): While not currently possible, PFS could return
		// CommitDeleted here. This should detect that error, but first:
		// - src/server/pfs/pfs.go should be moved to src/client/pfs (w/ other err
		//   handling code)
		// - packages depending on that code should be migrated
		// Then this could add "|| pfs.IsCommitDeletedErr(err)" and satisfy the todo
		if errors.Is(err, io.EOF) {
			return nil // expected--with no output repo, FlushCommit can't return anything
		}
		return errors.Wrapf(err, "unexpected error value")
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
	iter, err = aliceClient.FlushCommit(
		[]*pfs.Commit{client.NewCommit(repo, "master")},
		[]*pfs.Repo{client.NewRepo(pipeline2)})
	require.NoError(t, err)
	require.NoErrorWithinT(t, 30*time.Second, func() error {
		_, err := iter.Next()
		return err
	})
	buf := &bytes.Buffer{}
	require.NoError(t, aliceClient.GetFile(pipeline2, "master", "/file.1", 0, 0, buf))
	require.Equal(t, "1", buf.String())
	buf.Reset()
	require.NoError(t, aliceClient.GetFile(pipeline2, "master", "/file.2", 0, 0, buf))
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
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	defer deleteAll(t)
	alice := tu.UniqueString("alice")
	aliceClient, adminClient := getPachClient(t, alice), getPachClient(t, admin)

	// Create input repo w/ initial commit
	repo := tu.UniqueString(t.Name())
	require.NoError(t, aliceClient.CreateRepo(repo))
	_, err := aliceClient.PutFile(repo, "master", "/file.1", strings.NewReader("1"))
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
	_, err = adminClient.SetScope(adminClient.Ctx(), &auth.SetScopeRequest{
		Username: fmt.Sprintf("pipeline:%s", pipeline),
		Repo:     pipeline,
		Scope:    auth.Scope_NONE,
	})
	require.NoError(t, err)

	// make sure flush-commit returns (pipeline either
	// fails or restarts RC & finishes)
	iter, err := aliceClient.FlushCommit(
		[]*pfs.Commit{client.NewCommit(repo, "master")},
		[]*pfs.Repo{client.NewRepo(pipeline)})
	require.NoError(t, err)
	require.NoErrorWithinT(t, 30*time.Second, func() error {
		_, err := iter.Next()
		return err
	})

	// make sure the pipeline is failed
	pi, err := adminClient.InspectPipeline(pipeline)
	require.NoError(t, err)
	require.Equal(t, pps.PipelineState_PIPELINE_FAILURE, pi.State)
}

// TestOneTimePasswordOtherUser tests that it's possible for an admin to
// generate an OTP on behalf of another user (similar to how an admin can
// generate a Pachyderm token for another user).
func TestOneTimePasswordOtherUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)

	adminClient, anonClient := getPachClient(t, admin), getPachClient(t, "")
	otpResp, err := adminClient.GetOneTimePassword(adminClient.Ctx(),
		&auth.GetOneTimePasswordRequest{
			Subject: gh("alice"),
		})
	require.NoError(t, err)

	authResp, err := anonClient.Authenticate(anonClient.Ctx(),
		&auth.AuthenticateRequest{
			OneTimePassword: otpResp.Code,
		})
	require.NoError(t, err)
	anonClient.SetAuthToken(authResp.PachToken)
	who, err := anonClient.WhoAmI(anonClient.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, 0, len(who.AdminRoles.Roles))
	require.Equal(t, gh("alice"), who.Username)
	require.True(t, who.TTL > 0)
}

// TestFSAdminOneTimePasswordOtherUser tests that it's not possible for an FS admin to
// generate an OTP on behalf of another user.
func TestFSAdminOneTimePasswordOtherUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)

	alice := tu.UniqueString("alice")
	aliceClient := getPachClient(t, alice)
	adminClient := getPachClient(t, admin)

	// 'admin' makes alice an fs admin
	_, err := adminClient.ModifyAdmins(adminClient.Ctx(),
		&auth.ModifyAdminsRequest{Principal: gh(alice), Roles: fsAdminRole()})
	require.NoError(t, err)

	// wait until alice shows up in fs admin list
	require.NoError(t, backoff.Retry(func() error {
		resp, err := aliceClient.GetAdmins(aliceClient.Ctx(), &auth.GetAdminsRequest{})
		require.NoError(t, err)
		return require.EqualOrErr(
			admins(admin)(gh(alice)), resp.Admins,
		)
	}, backoff.NewTestingBackOff()))

	// Fail to get an OTP for another subject as alice
	otpResp, err := aliceClient.GetOneTimePassword(aliceClient.Ctx(),
		&auth.GetOneTimePasswordRequest{
			Subject: gh("bob"),
		})
	require.Nil(t, otpResp)
	require.YesError(t, err)
	require.Matches(t, "must be an admin", err.Error())
}

// TestInitialRobotUserGetOTP tests that Pachyderm can be activated with a robot
// user, which will have no session expiration, and that when this user gets an
// OTP, the token associated with their OTP will expire
func TestInitialRobotUserGetOTP(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	// adminClient will use 'admin's initial Pachyderm token, which does not
	// expire
	adminClient := getPachClient(t, admin)
	// make sure the admin client is a robot user initial admin
	who, err := adminClient.WhoAmI(adminClient.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, superAdminRole(), who.AdminRoles)
	require.True(t, strings.HasPrefix(who.Username, auth.RobotPrefix))
	require.Equal(t, int64(-1), who.TTL)

	// Get an OTP as the initial admin
	otpResp, err := adminClient.GetOneTimePassword(adminClient.Ctx(),
		&auth.GetOneTimePasswordRequest{})
	require.NoError(t, err)

	authResp, err := adminClient.Authenticate(adminClient.Ctx(), &auth.AuthenticateRequest{
		OneTimePassword: otpResp.Code,
	})
	require.NoError(t, err)
	adminClient.SetAuthToken(authResp.PachToken)
	who, err = adminClient.WhoAmI(adminClient.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, superAdminRole(), who.AdminRoles)
	require.True(t, strings.HasPrefix(who.Username, auth.RobotPrefix))
	require.True(t, who.TTL > 0)
}
