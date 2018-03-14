// admin_test.go tests various features related to pachyderm's auth admins.
// Because the cluster has one global set of admins, these tests can't be run in
// parallel

package server

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/enterprise"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
)

func RepoInfoToName(repoInfo interface{}) interface{} {
	return repoInfo.(*pfs.RepoInfo).Repo.Name
}

func TSProtoOrDie(t *testing.T, ts time.Time) *types.Timestamp {
	proto, err := types.TimestampProto(ts)
	require.NoError(t, err)
	return proto
}

// helper function that prepends auth.GitHubPrefix to 'user'--useful for validating
// responses
func gh(user string) string {
	return auth.GitHubPrefix + user
}

// TestActivate tests the Activate API (in particular, verifying
// that Activate() also authenticates). Even though GetClient also activates
// auth, this makes sure the code path is exercised (as auth may already be
// active when the test starts)
func TestActivate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	// Get anonymous client (this will activate auth, which is about to be
	// deactivated, but it also activates Pacyderm enterprise, which is needed for
	// this test to pass)
	c := getPachClient(t, "")

	// Deactivate auth (if it's activated)
	require.NoError(t, func() error {
		adminClient := &client.APIClient{}
		*adminClient = *c
		resp, err := adminClient.Authenticate(adminClient.Ctx(),
			&auth.AuthenticateRequest{GitHubToken: "admin"})
		if err != nil {
			if auth.IsNotActivatedError(err) {
				return nil
			}
			return err
		}
		adminClient.SetAuthToken(resp.PachToken)
		_, err = adminClient.Deactivate(adminClient.Ctx(), &auth.DeactivateRequest{})
		return err
	}())

	// Activate auth
	resp, err := c.AuthAPIClient.Activate(c.Ctx(), &auth.ActivateRequest{GitHubToken: "admin"})
	require.NoError(t, err)
	c.SetAuthToken(resp.PachToken)

	// Check that the token 'c' received from PachD authenticates them as "admin"
	who, err := c.WhoAmI(c.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.True(t, who.IsAdmin)
	require.Equal(t, admin, who.Username)
}

// TestAdminRWO tests adding and removing cluster admins, as well as admins
// reading, writing, and moderating (owning) all repos in the cluster.
func TestAdminRWO(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	alice, bob := tu.UniqueString("alice"), tu.UniqueString("bob")
	aliceClient, bobClient := getPachClient(t, alice), getPachClient(t, bob)
	adminClient := getPachClient(t, admin)

	// The initial set of admins is just the user "admin"
	resp, err := aliceClient.GetAdmins(aliceClient.Ctx(), &auth.GetAdminsRequest{})
	require.NoError(t, err)
	require.ElementsEqual(t, []string{admin}, resp.Admins)

	// alice creates a repo (that only she owns) and puts a file
	repo := tu.UniqueString("TestAdminRWO")
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.Equal(t, entries(alice, "owner"), GetACL(t, aliceClient, repo))
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
	require.Equal(t, entries(alice, "owner"), GetACL(t, aliceClient, repo)) // check that ACL wasn't updated

	// 'admin' makes bob an admin
	_, err = adminClient.ModifyAdmins(adminClient.Ctx(),
		&auth.ModifyAdminsRequest{Add: []string{bob}})
	require.NoError(t, err)

	// wait until bob shows up in admin list
	require.NoError(t, backoff.Retry(func() error {
		resp, err := aliceClient.GetAdmins(aliceClient.Ctx(), &auth.GetAdminsRequest{})
		require.NoError(t, err)
		return require.ElementsEqualOrErr(
			[]string{admin, gh(bob)}, resp.Admins,
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
		entries(alice, "owner", "carol", "reader"), GetACL(t, aliceClient, repo))

	// 'admin' revokes bob's admin status
	_, err = adminClient.ModifyAdmins(adminClient.Ctx(),
		&auth.ModifyAdminsRequest{Remove: []string{bob}})
	require.NoError(t, err)

	// wait until bob is not in admin list
	backoff.Retry(func() error {
		resp, err := aliceClient.GetAdmins(aliceClient.Ctx(), &auth.GetAdminsRequest{})
		require.NoError(t, err)
		return require.ElementsEqualOrErr([]string{admin}, resp.Admins)
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
		entries(alice, "owner", "carol", "reader"), GetACL(t, aliceClient, repo))
}

func TestAdminFixBrokenRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	alice := tu.UniqueString("alice")
	aliceClient, adminClient := getPachClient(t, alice), getPachClient(t, admin)

	// alice creates a repo (that only she owns) and puts a file
	repo := tu.UniqueString("TestAdmin")
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.Equal(t, entries(alice, "owner"), GetACL(t, aliceClient, repo))

	// admin deletes the repo's ACL
	_, err := adminClient.AuthAPIClient.SetACL(adminClient.Ctx(),
		&auth.SetACLRequest{
			Repo:    repo,
			Entries: nil,
		})
	require.NoError(t, err)

	// Check that the ACL is empty
	require.Equal(t, entries(), GetACL(t, adminClient, repo))

	// alice cannot write to the repo
	_, err = aliceClient.StartCommit(repo, "master")
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.Equal(t, 0, CommitCnt(t, adminClient, repo)) // check that no commits were created

	// admin can update the ACL to put Alice back, even though reading the ACL
	// will fail
	_, err = adminClient.SetACL(adminClient.Ctx(),
		&auth.SetACLRequest{
			Repo: repo,
			Entries: []*auth.ACLEntry{
				{Username: alice, Scope: auth.Scope_OWNER},
			},
		})
	require.NoError(t, err)
	require.Equal(t, entries(alice, "owner"), GetACL(t, aliceClient, repo))

	// now alice can write to the repo
	commit, err := aliceClient.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = aliceClient.PutFile(repo, commit.ID, "/file", strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.ID))
	require.Equal(t, 1, CommitCnt(t, adminClient, repo)) // check that a new commit was created
}

func TestModifyAdminsErrorMissingAdmin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	alice := tu.UniqueString("alice")
	adminClient := getPachClient(t, admin)

	// Check that the initial set of admins is just "admin"
	resp, err := adminClient.GetAdmins(adminClient.Ctx(), &auth.GetAdminsRequest{})
	require.NoError(t, err)
	require.ElementsEqual(t, []string{admin}, resp.Admins)

	// make alice a cluster administrator
	_, err = adminClient.ModifyAdmins(adminClient.Ctx(),
		&auth.ModifyAdminsRequest{
			Add: []string{alice},
		})
	require.NoError(t, err)
	require.NoError(t, backoff.Retry(func() error {
		resp, err = adminClient.GetAdmins(adminClient.Ctx(), &auth.GetAdminsRequest{})
		require.NoError(t, err)
		return require.ElementsEqualOrErr([]string{admin, gh(alice)}, resp.Admins)
	}, backoff.NewTestingBackOff()))

	// Try to remove alice and FakeUser
	_, err = adminClient.ModifyAdmins(adminClient.Ctx(),
		&auth.ModifyAdminsRequest{Remove: []string{alice, "FakeUser"}})
	require.YesError(t, err)
	require.NoError(t, backoff.Retry(func() error {
		resp, err = adminClient.GetAdmins(adminClient.Ctx(), &auth.GetAdminsRequest{})
		require.NoError(t, err)
		return require.ElementsEqualOrErr([]string{admin}, resp.Admins)
	}, backoff.NewTestingBackOff()))
}

func TestCannotRemoveAllClusterAdmins(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	alice := tu.UniqueString("alice")
	aliceClient, adminClient := getPachClient(t, alice), getPachClient(t, admin)

	// Check that the initial set of admins is just "admin"
	resp, err := adminClient.GetAdmins(adminClient.Ctx(), &auth.GetAdminsRequest{})
	require.NoError(t, err)
	require.ElementsEqual(t, []string{admin}, resp.Admins)

	// admin cannot remove themselves from the list of cluster admins (otherwise
	// there would be no admins)
	_, err = adminClient.ModifyAdmins(adminClient.Ctx(),
		&auth.ModifyAdminsRequest{Remove: []string{"admin"}})
	require.YesError(t, err)
	require.NoError(t, backoff.Retry(func() error {
		resp, err = adminClient.GetAdmins(adminClient.Ctx(), &auth.GetAdminsRequest{})
		require.NoError(t, err)
		return require.ElementsEqualOrErr([]string{admin}, resp.Admins)
	}, backoff.NewTestingBackOff()))

	// admin can make alice a cluster administrator
	_, err = adminClient.ModifyAdmins(adminClient.Ctx(),
		&auth.ModifyAdminsRequest{
			Add: []string{alice},
		})
	require.NoError(t, err)
	require.NoError(t, backoff.Retry(func() error {
		resp, err = adminClient.GetAdmins(adminClient.Ctx(), &auth.GetAdminsRequest{})
		require.NoError(t, err)
		return require.ElementsEqualOrErr([]string{admin, gh(alice)}, resp.Admins)
	}, backoff.NewTestingBackOff()))

	// Now admin can remove themselves as a cluster admin
	_, err = adminClient.ModifyAdmins(adminClient.Ctx(),
		&auth.ModifyAdminsRequest{Remove: []string{"admin"}})
	require.NoError(t, err)
	require.NoError(t, backoff.Retry(func() error {
		resp, err = adminClient.GetAdmins(adminClient.Ctx(), &auth.GetAdminsRequest{})
		require.NoError(t, err)
		return require.ElementsEqualOrErr([]string{gh(alice)}, resp.Admins)
	}, backoff.NewTestingBackOff()))

	// now alice is the only admin, and she cannot remove herself as a cluster
	// administrator
	_, err = aliceClient.ModifyAdmins(aliceClient.Ctx(),
		&auth.ModifyAdminsRequest{Remove: []string{alice}})
	require.YesError(t, err)
	require.NoError(t, backoff.Retry(func() error {
		resp, err = aliceClient.GetAdmins(aliceClient.Ctx(), &auth.GetAdminsRequest{})
		require.NoError(t, err)
		return require.ElementsEqualOrErr([]string{gh(alice)}, resp.Admins)
	}, backoff.NewTestingBackOff()))

	// alice *can* swap herself and "admin"
	_, err = aliceClient.ModifyAdmins(aliceClient.Ctx(),
		&auth.ModifyAdminsRequest{
			Add:    []string{"admin"},
			Remove: []string{alice},
		})
	require.NoError(t, err)
	require.NoError(t, backoff.Retry(func() error {
		resp, err = adminClient.GetAdmins(adminClient.Ctx(), &auth.GetAdminsRequest{})
		require.NoError(t, err)
		return require.ElementsEqualOrErr([]string{admin}, resp.Admins)
	}, backoff.NewTestingBackOff()))
}

func TestPreActivationPipelinesRunAsAdmin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	alice := tu.UniqueString("alice")
	aliceClient, adminClient := getPachClient(t, alice), getPachClient(t, admin)

	// Deactivate auth
	_, err := adminClient.Deactivate(adminClient.Ctx(), &auth.DeactivateRequest{})
	require.NoError(t, err)

	// Wait for auth to be deactivated
	require.NoError(t, backoff.Retry(func() error {
		_, err := aliceClient.WhoAmI(aliceClient.Ctx(), &auth.WhoAmIRequest{})
		if err != nil && auth.IsNotActivatedError(err) {
			return nil // WhoAmI should fail when auth is deactivated
		}
		return errors.New("auth is not yet deactivated")
	}, backoff.NewTestingBackOff()))

	// alice creates a pipeline
	repo := tu.UniqueString("TestPreActivationPipelinesRunAsAdmin")
	pipeline := tu.UniqueString("alice-pipeline")
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.NoError(t, aliceClient.CreatePipeline(
		pipeline,
		"", // default image: ubuntu:14.04
		[]string{"bash"},
		[]string{fmt.Sprintf("cp /pfs/%s/* /pfs/out/", repo)},
		&pps.ParallelismSpec{Constant: 1},
		client.NewAtomInput(repo, "/*"),
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
	_, err = adminClient.Activate(adminClient.Ctx(), &auth.ActivateRequest{
		GitHubToken: "admin",
	})
	require.NoError(t, err)

	// re-authenticate, as old tokens were deleted
	require.NoError(t, backoff.Retry(func() error {
		for i := 0; i < 2; i++ {
			client := []*client.APIClient{adminClient, aliceClient}[i]
			username := []string{"admin", alice}[i]
			resp, err := client.Authenticate(context.Background(),
				&auth.AuthenticateRequest{
					GitHubToken: username,
				})
			if err != nil {
				return err
			}
			client.SetAuthToken(resp.PachToken)
		}
		return nil
	}, backoff.NewTestingBackOff()))

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

	// make sure the pipeline runs (i.e. it's not running as alice)
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
			ActivationCode: tu.GetTestEnterpriseCode(),
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
	require.Equal(t, entries(alice, "owner"), GetACL(t, adminClient, repo)) // check that ACL wasn't updated

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
		entries(alice, "owner", "carol", "reader"), GetACL(t, adminClient, repo))

	// admin can re-authenticate
	resp, err := adminClient.Authenticate(context.Background(),
		&auth.AuthenticateRequest{GitHubToken: "admin"})
	require.NoError(t, err)
	adminClient.SetAuthToken(resp.PachToken)

	// Re-enable enterprise
	year := 365 * 24 * time.Hour
	adminClient.Enterprise.Activate(adminClient.Ctx(),
		&enterprise.ActivateRequest{
			ActivationCode: tu.GetTestEnterpriseCode(),
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
	resp, err = aliceClient.Authenticate(context.Background(),
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
		entries(alice, "owner", "carol", "writer"), GetACL(t, adminClient, repo))
}

func TestPipelinesRunAfterExpiration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	alice := tu.UniqueString("alice")
	aliceClient, adminClient := getPachClient(t, alice), getPachClient(t, admin)

	// alice creates a repo
	repo := tu.UniqueString("TestPipelinesRunAfterExpiration")
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.Equal(t, entries(alice, "owner"), GetACL(t, aliceClient, repo))

	// alice creates a pipeline
	pipeline := tu.UniqueString("alice-pipeline")
	require.NoError(t, aliceClient.CreatePipeline(
		pipeline,
		"", // default image: ubuntu:14.04
		[]string{"bash"},
		[]string{fmt.Sprintf("cp /pfs/%s/* /pfs/out/", repo)},
		&pps.ParallelismSpec{Constant: 1},
		client.NewAtomInput(repo, "/*"),
		"",    // default output branch: master
		false, // no update
	))
	require.OneOfEquals(t, pipeline, PipelineNames(t, aliceClient))
	require.Equal(t, entries(alice, "owner"), GetACL(t, aliceClient, pipeline)) // check that alice owns the output repo too

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
			ActivationCode: tu.GetTestEnterpriseCode(),
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
	alice := tu.UniqueString("alice")
	aliceClient, adminClient := getPachClient(t, alice), getPachClient(t, admin)

	// alice creates a repo
	repo := tu.UniqueString("TestGetSetScopeAndAclWithExpiredToken")
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.Equal(t, entries(alice, "owner"), GetACL(t, aliceClient, repo))

	// Make current enterprise token expire
	adminClient.Enterprise.Activate(adminClient.Ctx(),
		&enterprise.ActivateRequest{
			ActivationCode: tu.GetTestEnterpriseCode(),
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
	require.Equal(t, entries(alice, "owner"), GetACL(t, adminClient, repo))

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
			{alice, auth.Scope_OWNER},
			{"carol", auth.Scope_READER},
		},
	})
	require.YesError(t, err)
	require.Matches(t, "not active", err.Error())
	require.Equal(t, entries(alice, "owner"), GetACL(t, adminClient, repo))

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
		entries(alice, "owner", "carol", "reader"), GetACL(t, adminClient, repo))

	// admin can call GetAcl on repo
	aclResp, err := adminClient.GetACL(adminClient.Ctx(), &auth.GetACLRequest{
		Repo: repo,
	})
	require.NoError(t, err)
	require.ElementsEqual(t,
		entries(alice, "owner", "carol", "reader"), aclResp.Entries)

	// admin can call SetAcl on repo
	_, err = adminClient.SetACL(adminClient.Ctx(), &auth.SetACLRequest{
		Repo: repo,
		Entries: []*auth.ACLEntry{
			{alice, auth.Scope_OWNER},
			{"carol", auth.Scope_WRITER},
		},
	})
	require.NoError(t, err)
	require.ElementsEqual(t,
		entries(alice, "owner", "carol", "writer"), GetACL(t, adminClient, repo))
}

// TestAdminWhoAmI tests that when an admin calls WhoAmI(), the IsAdmin field
// in the result is true (and if a non-admin calls WhoAmI(), IsAdmin is false)
func TestAdminWhoAmI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	alice := tu.UniqueString("alice")
	aliceClient, adminClient := getPachClient(t, alice), getPachClient(t, admin)

	// Make sure admin WhoAmI indicates that they're an admin, and non-admin
	// WhoAmI indicates that they're not
	resp, err := aliceClient.WhoAmI(aliceClient.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, gh(alice), resp.Username)
	require.False(t, resp.IsAdmin)
	resp, err = adminClient.WhoAmI(adminClient.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, admin, resp.Username)
	require.True(t, resp.IsAdmin)
}

// TestListRepoAdminIsOwnerOfAllRepos tests that when an admin calls ListRepo,
// the result indicates that they're an owner of every repo in the cluster
// (needed by the Pachyderm dashboard)
func TestListRepoAdminIsOwnerOfAllRepos(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	alice, bob := tu.UniqueString("alice"), tu.UniqueString("bob")
	aliceClient, bobClient := getPachClient(t, alice), getPachClient(t, bob)
	adminClient := getPachClient(t, "admin")

	// alice creates a repo
	repoWriter := tu.UniqueString("TestListRepoAdminIsOwnerOfAllRepos")
	require.NoError(t, aliceClient.CreateRepo(repoWriter))

	// bob calls ListRepo, but has NONE access to all repos
	infos, err := bobClient.ListRepo([]string{})
	require.NoError(t, err)
	for _, info := range infos {
		require.Equal(t, auth.Scope_NONE, info.AuthInfo.AccessLevel)
	}

	// admin calls ListRepo, and has OWNER access to all repos
	infos, err = adminClient.ListRepo([]string{})
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
	adminClient := getPachClient(t, admin)

	// Generate two auth credentials, and give them to two separate clients
	robotUser := auth.RobotPrefix + tu.UniqueString("optimus_prime")
	resp, err := adminClient.GetAuthToken(adminClient.Ctx(),
		&auth.GetAuthTokenRequest{Subject: robotUser})
	require.NoError(t, err)
	// copy client & use resp token
	robotClient1 := adminClient.WithCtx(context.Background())
	robotClient1.SetAuthToken(resp.Token)

	token1 := resp.Token
	resp, err = adminClient.GetAuthToken(adminClient.Ctx(),
		&auth.GetAuthTokenRequest{Subject: robotUser})
	require.NoError(t, err)
	require.NotEqual(t, token1, resp.Token)
	// copy client & use resp token
	robotClient2 := adminClient.WithCtx(context.Background())
	robotClient2.SetAuthToken(resp.Token)

	// robotClient1 creates a repo
	repo := tu.UniqueString("TestPipelinesRunAfterExpiration")
	require.NoError(t, robotClient1.CreateRepo(repo))
	require.Equal(t, entries(robotUser, "owner"), GetACL(t, robotClient1, repo))

	// robotClient1 creates a pipeline
	pipeline := tu.UniqueString("optimus-prime-line")
	require.NoError(t, robotClient1.CreatePipeline(
		pipeline,
		"", // default image: ubuntu:14.04
		[]string{"bash"},
		[]string{fmt.Sprintf("cp /pfs/%s/* /pfs/out/", repo)},
		&pps.ParallelismSpec{Constant: 1},
		client.NewAtomInput(repo, "/*"),
		"",    // default output branch: master
		false, // no update
	))
	require.OneOfEquals(t, pipeline, PipelineNames(t, robotClient1))
	// check that robotUser owns the output repo
	require.Equal(t, entries(robotUser, "owner"), GetACL(t, robotClient1, pipeline))

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
		client.NewAtomInput(repo, "/*"),
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

// TestRobotUserWhoAmI tests that robot users can call WhoAmI and get a response
// with the right prefix
func TestRobotUserWhoAmI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	adminClient := getPachClient(t, admin)

	// Generate a robot user auth credential, and create a client for that user
	robotUser := auth.RobotPrefix + tu.UniqueString("r2d2")
	resp, err := adminClient.GetAuthToken(adminClient.Ctx(),
		&auth.GetAuthTokenRequest{Subject: robotUser})
	require.NoError(t, err)
	// copy client & use resp token
	robotClient := adminClient.WithCtx(context.Background())
	robotClient.SetAuthToken(resp.Token)

	whoAmIResp, err := robotClient.WhoAmI(robotClient.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, robotUser, whoAmIResp.Username)
	require.True(t, strings.HasPrefix(whoAmIResp.Username, auth.RobotPrefix))
	require.False(t, whoAmIResp.IsAdmin)
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
	robotUser := auth.RobotPrefix + tu.UniqueString("voltron")
	resp, err := adminClient.GetAuthToken(adminClient.Ctx(),
		&auth.GetAuthTokenRequest{Subject: robotUser})
	require.NoError(t, err)
	// copy client & use resp token
	robotClient := adminClient.WithCtx(context.Background())
	robotClient.SetAuthToken(resp.Token)

	// robotUser creates a repo and adds alice as a writer
	repo := tu.UniqueString("TestRobotUserACL")
	require.NoError(t, robotClient.CreateRepo(repo))
	require.Equal(t, entries(robotUser, "owner"), GetACL(t, robotClient, repo))
	robotClient.SetScope(robotClient.Ctx(), &auth.SetScopeRequest{
		Repo:     repo,
		Scope:    auth.Scope_WRITER,
		Username: auth.GitHubPrefix + alice,
	})
	require.ElementsEqual(t,
		entries(alice, "writer", robotUser, "owner"), GetACL(t, robotClient, repo))

	// test that alice can commit to the robot user's repo
	commit, err := aliceClient.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.ID))

	// Now alice creates a repo, and adds robotUser as a writer
	repo2 := tu.UniqueString("TestRobotUserACL")
	require.NoError(t, aliceClient.CreateRepo(repo2))
	require.Equal(t, entries(alice, "owner"), GetACL(t, aliceClient, repo2))
	aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Repo:     repo2,
		Scope:    auth.Scope_WRITER,
		Username: robotUser,
	})
	require.ElementsEqual(t,
		entries(alice, "owner", robotUser, "writer"), GetACL(t, aliceClient, repo2))

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
	alice := tu.UniqueString("alice")
	aliceClient, adminClient := getPachClient(t, alice), getPachClient(t, admin)

	// Generate a robot user auth credential, and create a client for that user
	robotUser := auth.RobotPrefix + tu.UniqueString("bender")
	resp, err := adminClient.GetAuthToken(adminClient.Ctx(),
		&auth.GetAuthTokenRequest{Subject: robotUser})
	require.NoError(t, err)
	// copy client & use resp token
	robotClient := adminClient.WithCtx(context.Background())
	robotClient.SetAuthToken(resp.Token)

	// make robotUser an admin
	_, err = adminClient.ModifyAdmins(adminClient.Ctx(), &auth.ModifyAdminsRequest{
		Add: []string{robotUser},
	})
	require.NoError(t, err)
	// wait until robotUser shows up in admin list
	require.NoError(t, backoff.Retry(func() error {
		resp, err := adminClient.GetAdmins(adminClient.Ctx(), &auth.GetAdminsRequest{})
		require.NoError(t, err)
		return require.ElementsEqualOrErr([]string{admin, robotUser}, resp.Admins)
	}, backoff.NewTestingBackOff()))

	// robotUser mints a token for robotUser2
	robotUser2 := auth.RobotPrefix + tu.UniqueString("robocop")
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
		entries(robotUser2, "owner"), GetACL(t, robotClient, repo))
	_, err = robotClient.SetScope(robotClient.Ctx(), &auth.SetScopeRequest{
		Repo:     repo,
		Scope:    auth.Scope_WRITER,
		Username: alice,
	})
	require.NoError(t, err)
	require.ElementsEqual(t,
		entries(robotUser2, "owner", alice, "writer"), GetACL(t, robotClient, repo))
	commit, err = aliceClient.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.ID))
}

// TestTokenTTL tests that an admin can create a token with a TTL for a user,
// and that that token will expire
func TestTokenTTL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
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
	repos, err := aliceClient.ListRepo(nil)
	require.NoError(t, err)
	require.ElementsEqualUnderFn(t, []string{repo}, repos, RepoInfoToName)
	require.NoError(t, backoff.Retry(func() error {
		repos, err = aliceClient.ListRepo(nil)
		if err == nil {
			return errors.New("alice still has access to ListRepo")
		}
		require.True(t, auth.IsBadTokenError(err), err.Error())
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
	repos, err := aliceClient.ListRepo(nil)
	require.NoError(t, err)
	require.ElementsEqualUnderFn(t, []string{repo}, repos, RepoInfoToName)
	time.Sleep(10 * time.Second)
	repos, err = aliceClient.ListRepo(nil)
	require.True(t, auth.IsBadTokenError(err), err.Error())
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
	repos, err = aliceClient.ListRepo(nil)
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
	repos, err = aliceClient.ListRepo(nil)
	require.NoError(t, err)
	require.ElementsEqualUnderFn(t, []string{repo}, repos, RepoInfoToName)
	// wait longer, now token is expired
	time.Sleep(10 * time.Second)
	repos, err = aliceClient.ListRepo(nil)
	require.True(t, auth.IsBadTokenError(err), err.Error())
	require.Equal(t, 0, len(repos))
}

// TestTokenRevoke tests that an admin can revoke that token and it no longer works
func TestTokenRevoke(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
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
	repos, err := aliceClient.ListRepo(nil)
	require.NoError(t, err)
	require.ElementsEqualUnderFn(t, []string{repo}, repos, RepoInfoToName)

	// admin revokes token
	_, err = adminClient.RevokeAuthToken(adminClient.Ctx(), &auth.RevokeAuthTokenRequest{
		Token: resp.Token,
	})
	require.NoError(t, err)

	// alice's token is no longer valid
	repos, err = aliceClient.ListRepo(nil)
	require.True(t, auth.IsBadTokenError(err), err.Error())
	require.Equal(t, 0, len(repos))
}

// TestGetAuthTokenErrorNonAdminUser tests that non-admin users can't call
// GetAuthToken
func TestGetAuthTokenErrorNonAdminUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)

	alice := tu.UniqueString("alice")
	aliceClient := getPachClient(t, alice)
	resp, err := aliceClient.GetAuthToken(aliceClient.Ctx(), &auth.GetAuthTokenRequest{
		Subject: auth.RobotPrefix + tu.UniqueString("terminator"),
	})
	require.Nil(t, resp)
	require.YesError(t, err)
	require.Matches(t, "must be an admin", err.Error())
}
