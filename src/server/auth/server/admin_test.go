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
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
)

// ElementsEqual returns nil if the elements of the slice "expecteds" are
// exactly the elements of the slice "actuals", ignoring order (i.e.
// setwise-equal), and an error otherwise.
func ElementsEqual(expecteds interface{}, actuals interface{}) error {
	es := reflect.ValueOf(expecteds)
	as := reflect.ValueOf(actuals)
	if es.Kind() != reflect.Slice {
		return fmt.Errorf("ElementsEqual must be called with a slice, but \"expected\" was %s", es.Type().String())
	}
	if as.Kind() != reflect.Slice {
		return fmt.Errorf("ElementsEqual must be called with a slice, but \"actual\" was %s", as.Type().String())
	}

	// Make sure expecteds and actuals are slices of the same type, modulo
	// pointers (*T ~= T in this function)
	esArePtrs := es.Type().Elem().Kind() == reflect.Ptr
	asArePtrs := as.Type().Elem().Kind() == reflect.Ptr
	esElemType, asElemType := es.Type().Elem(), as.Type().Elem()
	if esArePtrs {
		esElemType = es.Type().Elem().Elem()
	}
	if asArePtrs {
		asElemType = as.Type().Elem().Elem()
	}
	if esElemType != asElemType {
		return fmt.Errorf("Expected []%s but got []%s", es.Type().Elem(), as.Type().Elem())
	}

	// Count up elements of expecteds
	intType := reflect.TypeOf(int(0))
	expectedCt := reflect.MakeMap(reflect.MapOf(esElemType, intType))
	for i := 0; i < es.Len(); i++ {
		v := es.Index(i)
		if esArePtrs {
			v = v.Elem()
		}
		if !expectedCt.MapIndex(v).IsValid() {
			expectedCt.SetMapIndex(v, reflect.ValueOf(1))
		} else {
			newCt := expectedCt.MapIndex(v).Int() + 1
			expectedCt.SetMapIndex(v, reflect.ValueOf(newCt))
		}
	}

	// Count up elements of actuals
	actualCt := reflect.MakeMap(reflect.MapOf(asElemType, intType))
	for i := 0; i < as.Len(); i++ {
		v := as.Index(i)
		if asArePtrs {
			v = v.Elem()
		}
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
	alice, bob := tu.UniqueString("alice"), tu.UniqueString("bob")
	aliceClient, bobClient := getPachClient(t, alice), getPachClient(t, bob)
	adminClient := getPachClient(t, "admin")

	// The initial set of admins is just the user "admin"
	resp, err := aliceClient.GetAdmins(aliceClient.Ctx(), &auth.GetAdminsRequest{})
	require.NoError(t, err)
	require.NoError(t, ElementsEqual([]string{"admin"}, resp.Admins))

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
		return ElementsEqual([]string{"admin", bob}, resp.Admins)
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
	require.NoError(t, ElementsEqual(
		entries(alice, "owner", "carol", "reader"), GetACL(t, aliceClient, repo)))

	// 'admin' revokes bob's admin status
	_, err = adminClient.ModifyAdmins(adminClient.Ctx(),
		&auth.ModifyAdminsRequest{Remove: []string{bob}})
	require.NoError(t, err)

	// wait until bob is not in admin list
	backoff.Retry(func() error {
		resp, err := aliceClient.GetAdmins(aliceClient.Ctx(), &auth.GetAdminsRequest{})
		require.NoError(t, err)
		return ElementsEqual([]string{"admin"}, resp.Admins)
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
	require.NoError(t, ElementsEqual(
		entries(alice, "owner", "carol", "reader"), GetACL(t, aliceClient, repo)))
}

func TestAdminFixBrokenRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	alice := tu.UniqueString("alice")
	aliceClient, adminClient := getPachClient(t, alice), getPachClient(t, "admin")

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

func TestCannotRemoveAllClusterAdmins(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	alice := tu.UniqueString("alice")
	aliceClient, adminClient := getPachClient(t, alice), getPachClient(t, "admin")

	// Check that the initial set of admins is just "admin"
	resp, err := adminClient.GetAdmins(adminClient.Ctx(), &auth.GetAdminsRequest{})
	require.NoError(t, err)
	require.NoError(t, ElementsEqual([]string{"admin"}, resp.Admins))

	// admin cannot remove themselves from the list of cluster admins (otherwise
	// there would be no admins)
	_, err = adminClient.ModifyAdmins(adminClient.Ctx(),
		&auth.ModifyAdminsRequest{Remove: []string{"admin"}})
	require.YesError(t, err)
	require.NoError(t, backoff.Retry(func() error {
		resp, err = adminClient.GetAdmins(adminClient.Ctx(), &auth.GetAdminsRequest{})
		require.NoError(t, err)
		return ElementsEqual([]string{"admin"}, resp.Admins)
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
		return ElementsEqual([]string{"admin", alice}, resp.Admins)
	}, backoff.NewTestingBackOff()))

	// Now admin can remove themselves as a cluster admin
	_, err = adminClient.ModifyAdmins(adminClient.Ctx(),
		&auth.ModifyAdminsRequest{Remove: []string{"admin"}})
	require.NoError(t, err)
	require.NoError(t, backoff.Retry(func() error {
		resp, err = adminClient.GetAdmins(adminClient.Ctx(), &auth.GetAdminsRequest{})
		require.NoError(t, err)
		return ElementsEqual([]string{alice}, resp.Admins)
	}, backoff.NewTestingBackOff()))

	// now alice is the only admin, and she cannot remove herself as a cluster
	// administrator
	_, err = aliceClient.ModifyAdmins(aliceClient.Ctx(),
		&auth.ModifyAdminsRequest{Remove: []string{alice}})
	require.YesError(t, err)
	require.NoError(t, backoff.Retry(func() error {
		resp, err = aliceClient.GetAdmins(aliceClient.Ctx(), &auth.GetAdminsRequest{})
		require.NoError(t, err)
		return ElementsEqual([]string{alice}, resp.Admins)
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
		return ElementsEqual([]string{"admin"}, resp.Admins)
	}, backoff.NewTestingBackOff()))
}

func TestPreActivationPipelinesRunAsAdmin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	alice := tu.UniqueString("alice")
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
		GithubUsername: "admin",
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
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := iter.Next()
		return err
	})
}

func TestExpirationRepoOnlyAccessibleToAdmins(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	alice := tu.UniqueString("alice")
	aliceClient, adminClient := getPachClient(t, alice), getPachClient(t, "admin")

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
	// check that ACL was updated
	require.NoError(t, ElementsEqual(
		entries(alice, "owner", "carol", "reader"), GetACL(t, adminClient, repo)))

	// admin can re-authenticate
	resp, err := adminClient.Authenticate(context.Background(),
		&auth.AuthenticateRequest{GithubUsername: "admin"})
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
	// check that ACL was updated
	require.NoError(t, ElementsEqual(
		entries(alice, "owner", "carol", "writer"), GetACL(t, adminClient, repo)))
}

func TestPipelinesRunAfterExpiration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	alice := tu.UniqueString("alice")
	aliceClient, adminClient := getPachClient(t, alice), getPachClient(t, "admin")

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
	alice := tu.UniqueString("alice")
	aliceClient, adminClient := getPachClient(t, alice), getPachClient(t, "admin")

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
	require.NoError(t, ElementsEqual(
		entries(alice, "owner", "carol", "reader"), GetACL(t, adminClient, repo)))

	// admin can call GetAcl on repo
	aclResp, err := adminClient.GetACL(adminClient.Ctx(), &auth.GetACLRequest{
		Repo: repo,
	})
	require.NoError(t, err)
	require.NoError(t, ElementsEqual(
		entries(alice, "owner", "carol", "reader"), aclResp.Entries))

	// admin can call SetAcl on repo
	_, err = adminClient.SetACL(adminClient.Ctx(), &auth.SetACLRequest{
		Repo: repo,
		Entries: []*auth.ACLEntry{
			{alice, auth.Scope_OWNER},
			{"carol", auth.Scope_WRITER},
		},
	})
	require.NoError(t, err)
	require.NoError(t, ElementsEqual(
		entries(alice, "owner", "carol", "writer"), GetACL(t, adminClient, repo)))
}

// TestAdminWhoAmI tests that when an admin calls WhoAmI(), the IsAdmin field
// in the result is true (and if a non-admin calls WhoAmI(), IsAdmin is false)
func TestAdminWhoAmI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	alice := tu.UniqueString("alice")
	aliceClient, adminClient := getPachClient(t, alice), getPachClient(t, "admin")

	// Make sure admin WhoAmI indicates that they're an admin, and non-admin
	// WhoAmI indicates that they're not
	resp, err := aliceClient.WhoAmI(aliceClient.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, alice, resp.Username)
	require.False(t, resp.IsAdmin)
	resp, err = adminClient.WhoAmI(adminClient.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, "admin", resp.Username)
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
