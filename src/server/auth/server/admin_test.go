// admin_test.go tests various features related to pachyderm's auth admins.
// Because the cluster has one global set of admins, these tests can't be run in
// parallel

package auth

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
)

// adminsEqual returns nil if the elements of the slice "expecteds" are
// exactly the elements of the slice "actuals", ignoring order (i.e.
// setwise-equal), and an error otherwise.
func adminsEqual(expecteds interface{}, actuals interface{}) error {
	es := reflect.ValueOf(expecteds)
	as := reflect.ValueOf(actuals)
	if es.Kind() != reflect.Slice {
		return fmt.Errorf("ElementsEqual must be called with a slice, but \"expected\" was %s", es.Type().String())
	}
	if as.Kind() != reflect.Slice {
		return fmt.Errorf("ElementsEqual must be called with a slice, but \"actual\" was %s", as.Type().String())
	}
	if es.Type().Elem() != as.Type().Elem() {
		return fmt.Errorf("Expected []%s but got []%s", es.Type().Elem(), as.Type().Elem())
	}
	expectedCt := reflect.MakeMap(reflect.MapOf(es.Type().Elem(), reflect.TypeOf(int(0))))
	actualCt := reflect.MakeMap(reflect.MapOf(as.Type().Elem(), reflect.TypeOf(int(0))))
	for i := 0; i < es.Len(); i++ {
		v := es.Index(i)
		if !expectedCt.MapIndex(v).IsValid() {
			expectedCt.SetMapIndex(v, reflect.ValueOf(1))
		} else {
			newCt := expectedCt.MapIndex(v).Int() + 1
			expectedCt.SetMapIndex(v, reflect.ValueOf(newCt))
		}
	}
	for i := 0; i < as.Len(); i++ {
		v := as.Index(i)
		if !actualCt.MapIndex(v).IsValid() {
			actualCt.SetMapIndex(v, reflect.ValueOf(1))
		} else {
			newCt := actualCt.MapIndex(v).Int() + 1
			actualCt.SetMapIndex(v, reflect.ValueOf(newCt))
		}
	}
	if expectedCt.Len() != actualCt.Len() {
		return fmt.Errorf("expected %d distinct elements, but got %d", expectedCt.Len(), actualCt.Len())
	}
	for _, key := range expectedCt.MapKeys() {
		ec := expectedCt.MapIndex(key)
		ac := actualCt.MapIndex(key)
		if !ec.IsValid() || !ac.IsValid() || ec.Int() != ac.Int() {
			ecInt, acInt := int64(0), int64(0)
			if ec.IsValid() {
				ecInt = ec.Int()
			}
			if ac.IsValid() {
				acInt = ac.Int()
			}
			return fmt.Errorf("expected %d copies of %s, but got %d copies", ecInt, key.String(), acInt)
		}
	}
	return nil
}

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
	require.NoError(t, adminsEqual([]string{"admin"}, resp.Admins))

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
		if err := adminsEqual([]string{"admin", bob}, resp.Admins); err != nil {
			return err
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
		if err := adminsEqual([]string{"admin"}, resp.Admins); err != nil {
			return err
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
	require.NoError(t, adminsEqual([]string{"admin"}, resp.Admins))

	// admin cannot remove themselves from the list of cluster admins (otherwise
	// there would be no admins)
	_, err = adminClient.ModifyAdmins(adminClient.Ctx(),
		&auth.ModifyAdminsRequest{Remove: []string{"admin"}})
	require.YesError(t, err)
	require.NoError(t, backoff.Retry(func() error {
		resp, err = adminClient.GetAdmins(adminClient.Ctx(), &auth.GetAdminsRequest{})
		require.NoError(t, err)
		if err := adminsEqual([]string{"admin"}, resp.Admins); err != nil {
			return err
		}
		return nil
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
		if err := adminsEqual([]string{"admin", alice}, resp.Admins); err != nil {
			return err
		}
		return nil
	}, backoff.NewTestingBackOff()))

	// Now admin can remove themselves as a cluster admin
	_, err = adminClient.ModifyAdmins(adminClient.Ctx(),
		&auth.ModifyAdminsRequest{Remove: []string{"admin"}})
	require.NoError(t, err)
	require.NoError(t, backoff.Retry(func() error {
		resp, err = adminClient.GetAdmins(adminClient.Ctx(), &auth.GetAdminsRequest{})
		require.NoError(t, err)
		if err := adminsEqual([]string{alice}, resp.Admins); err != nil {
			return err
		}
		return nil
	}, backoff.NewTestingBackOff()))

	// now alice is the only admin, and she cannot remove herself as a cluster
	// administrator
	_, err = aliceClient.ModifyAdmins(aliceClient.Ctx(),
		&auth.ModifyAdminsRequest{Remove: []string{alice}})
	require.YesError(t, err)
	require.NoError(t, backoff.Retry(func() error {
		resp, err = aliceClient.GetAdmins(aliceClient.Ctx(), &auth.GetAdminsRequest{})
		require.NoError(t, err)
		if err := adminsEqual([]string{alice}, resp.Admins); err != nil {
			return err
		}
		return nil
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
		if err := adminsEqual([]string{"admin"}, resp.Admins); err != nil {
			return err
		}
		return nil
	}, backoff.NewTestingBackOff()))
}
