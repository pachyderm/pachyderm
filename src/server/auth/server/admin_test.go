// admin_test.go tests various features related to pachyderm's auth admins.
// Because the cluster has one global set of admins, these tests can't be run in
// parallel

package server

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
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

func TSProtoOrDie(t *testing.T, ts time.Time) *types.Timestamp {
	proto, err := types.TimestampProto(ts)
	require.NoError(t, err)
	return proto
}

// TestAdminRWO tests adding and removing cluster admins, as well as admins
// reading, writing, and moderating (owning) all repos in the cluster.
func TestAdminRWO(t *testing.T) {
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
	repo := uniqueString("TestAdminRWO")
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
			NewACL: &auth.ACL{
				Entries: map[string]auth.Scope{alice: auth.Scope_OWNER},
			},
		})
	require.NoError(t, err)
	require.Equal(t, acl(alice, "owner"), GetACL(t, aliceClient, repo))

	// now alice can write to the repo
	commit, err := aliceClient.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = aliceClient.PutFile(repo, commit.ID, "/file", strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.ID))
	require.Equal(t, 1, CommitCnt(t, adminClient, repo)) // check that a new commit was created
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

func TestPreActivationPipelinesRunAsAdmin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	alice := uniqueString("alice")
	aliceClient, adminClient := getPachClient(t, alice), getPachClient(t, "admin")

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
	repo := uniqueString("TestPreActivationPipelinesRunAsAdmin")
	pipeline := uniqueString("alice-pipeline")
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
	require.NoErrorWithinT(t, 30*time.Second, func() error {
		_, err := iter.Next()
		return err
	})

	// activate auth
	_, err = adminClient.Activate(adminClient.Ctx(), &auth.ActivateRequest{
		Admins: []string{"admin"},
	})
	require.NoError(t, err)

	// re-authenticate, as old tokens were deleted
	require.NoError(t, backoff.Retry(func() error {
		for i, client := range []*client.APIClient{adminClient, aliceClient} {
			resp, err := client.Authenticate(context.Background(),
				&auth.AuthenticateRequest{
					GithubUsername: []string{"admin", alice}[i],
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
	require.NoErrorWithinT(t, 30*time.Second, func() error {
		_, err := iter.Next()
		return err
	})
}

func TestExpirationRepoOnlyAccessibleToAdmins(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	alice := uniqueString("alice")
	aliceClient, adminClient := getPachClient(t, alice), getPachClient(t, "admin")

	// alice creates a repo
	repo := uniqueString("TestExpirationRepoOnlyAccessibleToAdmins")
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
			ActivationCode: testActivationCode,
			Expires:        TSProtoOrDie(t, time.Now().Add(-30*time.Second)),
		})

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
	require.Equal(t, acl(alice, "owner"), GetACL(t, adminClient, repo)) // check that ACL wasn't updated

	// alice also can't re-authenticate
	_, err = aliceClient.Authenticate(context.Background(),
		&auth.AuthenticateRequest{GithubUsername: alice})
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
	require.Equal(t, acl(alice, "owner", "carol", "reader"),
		GetACL(t, adminClient, repo)) // check that ACL was updated

	// admin can re-authenticate
	resp, err := adminClient.Authenticate(context.Background(),
		&auth.AuthenticateRequest{GithubUsername: "admin"})
	require.NoError(t, err)
	adminClient.SetAuthToken(resp.PachToken)

	// Re-enable enterprise
	year := 365 * 24 * time.Hour
	adminClient.Enterprise.Activate(adminClient.Ctx(),
		&enterprise.ActivateRequest{
			ActivationCode: testActivationCode,
			// This will stop working some time in 2026
			Expires: TSProtoOrDie(t, time.Now().Add(year)),
		})

	// alice can now re-authenticate
	resp, err = aliceClient.Authenticate(context.Background(),
		&auth.AuthenticateRequest{GithubUsername: alice})
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
	require.Equal(t, acl(alice, "owner", "carol", "writer"),
		GetACL(t, adminClient, repo)) // check that ACL was updated
}

func TestPipelinesRunAfterExpiration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	alice := uniqueString("alice")
	aliceClient, adminClient := getPachClient(t, alice), getPachClient(t, "admin")

	// alice creates a repo
	repo := uniqueString("TestPipelinesRunAfterExpiration")
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.Equal(t, acl(alice, "owner"), GetACL(t, aliceClient, repo))

	// alice creates a pipeline
	pipeline := uniqueString("alice-pipeline")
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
	require.Equal(t, acl(alice, "owner"), GetACL(t, aliceClient, pipeline)) // check that alice owns the output repo too

	// Make sure alice's pipeline runs successfully
	commit, err := aliceClient.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = aliceClient.PutFile(repo, commit.ID, uniqueString("/file1"),
		strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.ID))
	iter, err := aliceClient.FlushCommit(
		[]*pfs.Commit{commit},
		[]*pfs.Repo{{Name: pipeline}},
	)
	require.NoError(t, err)
	require.NoErrorWithinT(t, 30*time.Second, func() error {
		_, err := iter.Next()
		return err
	})

	// Make current enterprise token expire
	adminClient.Enterprise.Activate(adminClient.Ctx(),
		&enterprise.ActivateRequest{
			ActivationCode: testActivationCode,
			Expires:        TSProtoOrDie(t, time.Now().Add(-30*time.Second)),
		})

	// Make sure alice's pipeline still runs successfully
	commit, err = adminClient.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = adminClient.PutFile(repo, commit.ID, uniqueString("/file2"),
		strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, adminClient.FinishCommit(repo, commit.ID))
	iter, err = adminClient.FlushCommit(
		[]*pfs.Commit{commit},
		[]*pfs.Repo{{Name: pipeline}},
	)
	require.NoError(t, err)
	require.NoErrorWithinT(t, 30*time.Second, func() error {
		_, err := iter.Next()
		return err
	})
}

func TestGetSetScopeAclWithExpiredToken(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	alice := uniqueString("alice")
	aliceClient, adminClient := getPachClient(t, alice), getPachClient(t, "admin")

	// alice creates a repo
	repo := uniqueString("TestPipelinesRunAfterExpiration")
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.Equal(t, acl(alice, "owner"), GetACL(t, aliceClient, repo))

	// Make current enterprise token expire
	adminClient.Enterprise.Activate(adminClient.Ctx(),
		&enterprise.ActivateRequest{
			ActivationCode: testActivationCode,
			Expires:        TSProtoOrDie(t, time.Now().Add(-30*time.Second)),
		})

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
	require.Equal(t, acl(alice, "owner"), GetACL(t, adminClient, repo))

	// alice can't call GetAcl on repo
	_, err = aliceClient.GetACL(aliceClient.Ctx(), &auth.GetACLRequest{
		Repo: repo,
	})
	require.YesError(t, err)
	require.Matches(t, "not active", err.Error())

	// alice can't call GetAcl on repo
	_, err = aliceClient.SetACL(aliceClient.Ctx(), &auth.SetACLRequest{
		Repo: repo,
		NewACL: &auth.ACL{
			Entries: map[string]auth.Scope{
				alice:   auth.Scope_OWNER,
				"carol": auth.Scope_READER,
			},
		},
	})
	require.YesError(t, err)
	require.Matches(t, "not active", err.Error())
	require.Equal(t, acl(alice, "owner"), GetACL(t, adminClient, repo))

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
	require.Equal(t, acl(alice, "owner", "carol", "reader"),
		GetACL(t, adminClient, repo))

	// admin can call GetAcl on repo
	aclResp, err := adminClient.GetACL(adminClient.Ctx(), &auth.GetACLRequest{
		Repo: repo,
	})
	require.NoError(t, err)
	require.Equal(t, acl(alice, "owner", "carol", "reader"), aclResp.ACL)

	// admin can call GetAcl on repo
	_, err = adminClient.SetACL(adminClient.Ctx(), &auth.SetACLRequest{
		Repo: repo,
		NewACL: &auth.ACL{
			Entries: map[string]auth.Scope{
				alice:   auth.Scope_OWNER,
				"carol": auth.Scope_WRITER,
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, acl(alice, "owner", "carol", "writer"),
		GetACL(t, adminClient, repo))
}
