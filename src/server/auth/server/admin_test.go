// admin_test.go tests various features related to pachyderm's auth admins.
// Because the cluster has one global set of admins, these tests can't be run in
// parallel

package auth

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
)

// TestAdmins tests adding and removing cluster admins, as well as admins
// reading, writing, and moderating clusters.
func TestAdmin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	alice, bob := uniqueString("alice"), uniqueString("bob")
	aliceClient, bobClient := getPachClient(t, alice), getPachClient(t, bob)
	adminClient := getPachClient(t, "admin")

	// The initial set of admins is just the user "admin"
	resp, err := aliceClient.GetAdmins(aliceClient.Ctx(), &auth.GetAdminsRequest{})
	require.NoError(t, err)
	require.Equal(t, len(resp.Admins), 1)
	require.Equal(t, resp.Admins[0], "admin")

	// alice creates a repo (that only she owns) and puts a file
	repo := uniqueString("TestAdmin")
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.Equal(t, acl(alice, "owner"), GetACL(t, aliceClient, repo))
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
	require.Equal(t, 1, CommitCnt(t, aliceClient, repo)) // check that no commits were created

	// bob can't update the ACL
	_, err = bobClient.SetScope(bobClient.Ctx(), &auth.SetScopeRequest{
		Repo:     repo,
		Username: "carol",
		Scope:    auth.Scope_READER,
	})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.Equal(t, acl(alice, "owner"), GetACL(t, aliceClient, repo)) // check that ACL wasn't updated

	// 'admin' makes bob an admin
	_, err = adminClient.ModifyAdmins(adminClient.Ctx(),
		&auth.ModifyAdminsRequest{Add: []string{bob}})
	require.NoError(t, err)

	// wait until bob shows up in admin list
	require.NoError(t, backoff.Retry(func() error {
		resp, err := aliceClient.GetAdmins(aliceClient.Ctx(), &auth.GetAdminsRequest{})
		require.NoError(t, err)
		if len(resp.Admins) != 2 {
			return fmt.Errorf("admins are \"%v\", but expected [\"bob\", \"admin\"]", resp.Admins)
		}
		if resp.Admins[0] != bob && resp.Admins[1] != bob {
			return fmt.Errorf("admins are \"%v\", but expected at least one to be \"%s\"", resp.Admins, bob)
		}
		return nil
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
	require.Equal(t, acl(alice, "owner", "carol", "reader"),
		GetACL(t, aliceClient, repo)) // check that ACL was updated

	// 'admin' revokes bob's admin status
	_, err = adminClient.ModifyAdmins(adminClient.Ctx(),
		&auth.ModifyAdminsRequest{Remove: []string{bob}})
	require.NoError(t, err)

	// wait until bob is not in admin list
	backoff.Retry(func() error {
		resp, err := aliceClient.GetAdmins(aliceClient.Ctx(), &auth.GetAdminsRequest{})
		require.NoError(t, err)
		if len(resp.Admins) != 1 || resp.Admins[0] != "admin" {
			return fmt.Errorf("admins are \"%v\", but expected [\"admin\"]", resp.Admins)
		}
		return nil
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
	require.Equal(t, acl(alice, "owner", "carol", "reader"),
		GetACL(t, aliceClient, repo)) // check that ACL wasn't updated
}

func TestAdminFixBrokenRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	alice := uniqueString("alice")
	aliceClient, adminClient := getPachClient(t, alice), getPachClient(t, "admin")

	// alice creates a repo (that only she owns) and puts a file
	repo := uniqueString("TestAdmin")
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.Equal(t, acl(alice, "owner"), GetACL(t, aliceClient, repo))

	// admin deletes the repo's ACL
	_, err := adminClient.AuthAPIClient.SetACL(adminClient.Ctx(),
		&auth.SetACLRequest{
			Repo:   repo,
			NewACL: &auth.ACL{},
		})
	require.NoError(t, err)

	// Check that the ACL is empty
	require.Equal(t, acl(), GetACL(t, adminClient, repo))

	// admin can update the ACL to put Alice back, even though reading the ACL
	// will fail
	_, err = adminClient.SetACL(adminClient.Ctx(),
		&auth.SetACLRequest{
			Repo: repo,
			NewACL: &auth.ACL{
				Entries: map[string]auth.Scope{alice: auth.Scope_OWNER},
			},
		})
	require.NoError(t, err)
	require.Equal(t, acl(alice, "owner"), GetACL(t, aliceClient, repo))
}

func TestCannotRemoveAllClusterAdmins(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	alice := uniqueString("alice")
	aliceClient, adminClient := getPachClient(t, alice), getPachClient(t, "admin")

	// Check that the initial set of admins is just "admin"
	resp, err := adminClient.GetAdmins(adminClient.Ctx(), &auth.GetAdminsRequest{})
	require.NoError(t, err)
	require.ElementsEqual(t, []string{"admin"}, resp.Admins)

	// admin cannot remove themselves from the list of cluster admins (otherwise
	// there would be no admins)
	_, err = adminClient.ModifyAdmins(adminClient.Ctx(),
		&auth.ModifyAdminsRequest{Remove: []string{"admin"}})
	require.YesError(t, err)
	resp, err = adminClient.GetAdmins(adminClient.Ctx(), &auth.GetAdminsRequest{})
	require.NoError(t, err)
	require.ElementsEqual(t, []string{"admin"}, resp.Admins)

	// admin can make alice a cluster administrator
	_, err = adminClient.ModifyAdmins(adminClient.Ctx(),
		&auth.ModifyAdminsRequest{
			Add: []string{alice},
		})
	require.NoError(t, err)
	resp, err = adminClient.GetAdmins(adminClient.Ctx(), &auth.GetAdminsRequest{})
	require.NoError(t, err)
	require.ElementsEqual(t, []string{"admin", alice}, resp.Admins)

	// Now admin can remove themselves as a cluster admin
	_, err = adminClient.ModifyAdmins(adminClient.Ctx(),
		&auth.ModifyAdminsRequest{Remove: []string{"admin"}})
	require.NoError(t, err)
	resp, err = adminClient.GetAdmins(adminClient.Ctx(), &auth.GetAdminsRequest{})
	require.NoError(t, err)
	require.ElementsEqual(t, []string{alice}, resp.Admins)

	// now alice cannot remove herself as a cluster administrator
	_, err = aliceClient.ModifyAdmins(aliceClient.Ctx(),
		&auth.ModifyAdminsRequest{Remove: []string{alice}})
	require.YesError(t, err)
	resp, err = aliceClient.GetAdmins(aliceClient.Ctx(), &auth.GetAdminsRequest{})
	require.NoError(t, err)
	require.ElementsEqual(t, []string{alice}, resp.Admins)

	// alice can make admin the cluster administrator again (swapping herself and
	// admin)
	_, err = aliceClient.ModifyAdmins(aliceClient.Ctx(),
		&auth.ModifyAdminsRequest{
			Add:    []string{"admin"},
			Remove: []string{alice},
		})
	require.NoError(t, err)
	resp, err = adminClient.GetAdmins(adminClient.Ctx(), &auth.GetAdminsRequest{})
	require.NoError(t, err)
	require.ElementsEqual(t, []string{"admin"}, resp.Admins)
}
