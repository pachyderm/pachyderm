package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/enterprise"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/pkg/testutil"
)

var (
	clientMapMut sync.Mutex
	clientMap    = make(map[string]*client.APIClient)
)

// getPachClient creates a seed client with a grpc connection to a pachyderm
// cluster, and then enable the auth service in that cluster
func getPachClient(t testing.TB, u string) *client.APIClient {
	t.Helper()
	// Check if "u" already has a client -- if not create one
	func() {
		// Client creation is wrapped in an anonymous function to make locking and
		// releasing clientMapMut easier, and keep concurrent tests from racing with
		// initialization
		clientMapMut.Lock()
		defer clientMapMut.Unlock()

		// Check if seed client exists -- if not, create it
		seedClient, ok := clientMap[""]
		if !ok {
			var err error
			if _, ok := os.LookupEnv("PACHD_PORT_650_TCP_ADDR"); ok {
				seedClient, err = client.NewInCluster()
			} else {
				seedClient, err = client.NewOnUserMachine(false, "user")
			}
			require.NoError(t, err)
			seedClient.SetAuthToken("") // anonymous client
			clientMap[""] = seedClient
		}

		// Activate Pachyderm Enterprise (if it's not already active)
		require.NoError(t, backoff.Retry(func() error {
			resp, err := seedClient.Enterprise.GetState(context.Background(),
				&enterprise.GetStateRequest{})
			if err != nil {
				return err
			}
			if resp.State == enterprise.State_ACTIVE {
				return nil
			}
			_, err = seedClient.Enterprise.Activate(context.Background(),
				&enterprise.ActivateRequest{
					ActivationCode: testutil.GetTestEnterpriseCode(),
				})

			return err
		}, backoff.NewTestingBackOff()))

		// Activate Pachyderm auth (if it's not already active)
		require.NoError(t, backoff.Retry(func() error {
			if _, err := seedClient.GetAdmins(context.Background(),
				&auth.GetAdminsRequest{},
			); err == nil {
				return nil // auth already active
			}
			// Auth not active -- clear existing cached clients (as their auth tokens
			// are no longer valid)
			clientMap = map[string]*client.APIClient{"": clientMap[""]}
			if _, err := seedClient.AuthAPIClient.Activate(context.Background(),
				&auth.ActivateRequest{GithubUsername: "admin"},
			); err != nil && !strings.HasSuffix(err.Error(), "already activated") {
				return fmt.Errorf("could not activate auth service: %s", err.Error())
			}
			return nil
		}, backoff.NewTestingBackOff()))

		// Wait for the Pachyderm Auth system to activate
		require.NoError(t, backoff.Retry(func() error {
			if _, err := seedClient.AuthAPIClient.WhoAmI(seedClient.Ctx(),
				&auth.WhoAmIRequest{},
			); err != nil && auth.IsNotActivatedError(err) {
				return err
			}
			return nil
		}, backoff.NewTestingBackOff()))

		// Re-use old client for 'u', or create a new one if none exists
		if _, ok := clientMap[u]; !ok {
			userClient := *clientMap[""]
			resp, err := userClient.Authenticate(context.Background(),
				&auth.AuthenticateRequest{GithubUsername: string(u)})
			require.NoError(t, err)
			userClient.SetAuthToken(resp.PachToken)
			clientMap[u] = &userClient
		}
	}()
	return clientMap[u]
}

// entries constructs an auth.ACL struct from a list of the form
// [ user_1, scope_1, user_2, scope_2, ... ]
func entries(items ...string) []auth.ACLEntry {
	if len(items)%2 != 0 {
		panic("cannot create an ACL from an odd number of items")
	}
	if len(items) == 0 {
		return []auth.ACLEntry{}
	}
	result := make([]auth.ACLEntry, 0, len(items)/2)
	for i := 0; i < len(items); i += 2 {
		scope, err := auth.ParseScope(items[i+1])
		if err != nil {
			panic(fmt.Sprintf("could not parse scope: %v", err))
		}
		result = append(result, auth.ACLEntry{Username: items[i], Scope: scope})
	}
	return result
}

// GetACL uses the client 'c' to get the ACL protecting the repo 'repo'
// TODO(msteffen) create an auth client?
func GetACL(t *testing.T, c *client.APIClient, repo string) []auth.ACLEntry {
	t.Helper()
	resp, err := c.AuthAPIClient.GetACL(c.Ctx(), &auth.GetACLRequest{
		Repo: repo,
	})
	require.NoError(t, err)
	result := make([]auth.ACLEntry, 0, len(resp.Entries))
	for _, p := range resp.Entries {
		result = append(result, *p)
	}
	return result
}

// CommitCnt uses 'c' to get the number of commits made to the repo 'repo'
func CommitCnt(t *testing.T, c *client.APIClient, repo string) int {
	t.Helper()
	commitList, err := c.ListCommitByRepo(repo)
	require.NoError(t, err)
	return len(commitList)
}

// PipelineNames returns the names of all pipelines that 'c' gets from
// ListPipeline
func PipelineNames(t *testing.T, c *client.APIClient) []string {
	t.Helper()
	ps, err := c.ListPipeline()
	require.NoError(t, err)
	result := make([]string, len(ps))
	for i, p := range ps {
		result[i] = p.Pipeline.Name
	}
	return result
}

// TestActivate tests the Activate API (in particular, verifying
// that Activate() also authenticates). Even though GetClient also activates
// auth, this makes sure the code path is exercised (as auth may already be
// active when the test starts)
func TestActivate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	// Get client, but don't use getPachClient, as that will call Activate()
	var c *client.APIClient
	var err error
	if _, ok := os.LookupEnv("PACHD_PORT_650_TCP_ADDR"); ok {
		c, err = client.NewInCluster()
	} else {
		c, err = client.NewOnUserMachine(false, "user")
	}
	require.NoError(t, err)

	// Deactivate auth (if it's activated)
	require.NoError(t, func() error {
		adminClient := &client.APIClient{}
		*adminClient = *c
		resp, err := adminClient.Authenticate(adminClient.Ctx(),
			&auth.AuthenticateRequest{
				GithubUsername: "admin",
			})
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
	resp, err := c.Activate(c.Ctx(), &auth.ActivateRequest{
		GithubUsername: "admin",
	})
	require.NoError(t, err)
	c.SetAuthToken(resp.PachToken)

	// Check that the token 'c' recieved from PachD authenticates them as "admin"
	who, err := c.WhoAmI(c.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.True(t, who.IsAdmin)
	require.Equal(t, "admin", who.Username)
}

// TestGetSetBasic creates two users, alice and bob, and gives bob gradually
// escalating privileges, checking what bob can and can't do after each change
func TestGetSetBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	alice, bob := tu.UniqueString("alice"), tu.uniqueString("bob")
	aliceClient, bobClient := getPachClient(t, alice), getPachClient(t, bob)

	// create repo, and check that alice is the owner of the new repo
	dataRepo := tu.UniqueString("TestGetSetBasic")
	require.NoError(t, aliceClient.CreateRepo(dataRepo))
	require.NoError(t, ElementsEqual(
		entries(alice, "owner"), GetACL(t, aliceClient, dataRepo)))

	// Add data to repo (alice can write). Make sure alice can read also.
	commit, err := aliceClient.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = aliceClient.PutFile(dataRepo, commit.ID, "/file", strings.NewReader("lorem ipsum"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(dataRepo, commit.ID)) // # commits = 1
	buf := &bytes.Buffer{}
	require.NoError(t, aliceClient.GetFile(dataRepo, "master", "/file", 0, 0, buf))
	require.Equal(t, "lorem ipsum", buf.String())

	//////////
	/// Initially, bob has no privileges
	// bob can't read
	err = bobClient.GetFile(dataRepo, "master", "/file", 0, 0, buf)
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	// bob can't write
	_, err = bobClient.StartCommit(dataRepo, "master")
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.Equal(t, 1, CommitCnt(t, aliceClient, dataRepo)) // check that no commits were created
	// bob can't update the ACL
	_, err = bobClient.SetScope(bobClient.Ctx(), &auth.SetScopeRequest{
		Repo:     dataRepo,
		Username: "carol",
		Scope:    auth.Scope_READER,
	})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	// check that ACL wasn't updated)
	require.NoError(t, ElementsEqual(
		entries(alice, "owner"), GetACL(t, aliceClient, dataRepo)))

	//////////
	/// alice adds bob to the ACL as a reader (alice can modify ACL)
	_, err = aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Repo:     dataRepo,
		Username: bob,
		Scope:    auth.Scope_READER,
	})
	require.NoError(t, err)
	// bob can read
	buf.Reset()
	require.NoError(t, bobClient.GetFile(dataRepo, "master", "/file", 0, 0, buf))
	require.Equal(t, "lorem ipsum", buf.String())
	// bob can't write
	_, err = bobClient.StartCommit(dataRepo, "master")
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.Equal(t, 1, CommitCnt(t, aliceClient, dataRepo)) // check that no commits were created
	// bob can't update the ACL
	_, err = bobClient.SetScope(bobClient.Ctx(), &auth.SetScopeRequest{
		Repo:     dataRepo,
		Username: "carol",
		Scope:    auth.Scope_READER,
	})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	// check that ACL wasn't updated)
	require.NoError(t, ElementsEqual(
		entries(alice, "owner", bob, "reader"), GetACL(t, aliceClient, dataRepo)))

	//////////
	/// alice adds bob to the ACL as a writer
	_, err = aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Repo:     dataRepo,
		Username: bob,
		Scope:    auth.Scope_WRITER,
	})
	require.NoError(t, err)
	// bob can read
	buf.Reset()
	require.NoError(t, bobClient.GetFile(dataRepo, "master", "/file", 0, 0, buf))
	require.Equal(t, "lorem ipsum", buf.String())
	// bob can write
	commit, err = bobClient.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, bobClient.FinishCommit(dataRepo, commit.ID))
	require.Equal(t, 2, CommitCnt(t, aliceClient, dataRepo)) // check that a new commit was created
	// bob can't update the ACL
	_, err = bobClient.SetScope(bobClient.Ctx(), &auth.SetScopeRequest{
		Repo:     dataRepo,
		Username: "carol",
		Scope:    auth.Scope_READER,
	})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	// check that ACL wasn't updated)
	require.NoError(t, ElementsEqual(
		entries(alice, "owner", bob, "writer"), GetACL(t, aliceClient, dataRepo)))

	//////////
	/// alice adds bob to the ACL as an owner
	_, err = aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Repo:     dataRepo,
		Username: bob,
		Scope:    auth.Scope_OWNER,
	})
	require.NoError(t, err)
	// bob can read
	buf.Reset()
	require.NoError(t, bobClient.GetFile(dataRepo, "master", "/file", 0, 0, buf))
	require.Equal(t, "lorem ipsum", buf.String())
	// bob can write
	commit, err = bobClient.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, bobClient.FinishCommit(dataRepo, commit.ID))
	require.Equal(t, 3, CommitCnt(t, aliceClient, dataRepo)) // check that a new commit was created
	// bob can update the ACL
	_, err = bobClient.SetScope(bobClient.Ctx(), &auth.SetScopeRequest{
		Repo:     dataRepo,
		Username: "carol",
		Scope:    auth.Scope_READER,
	})
	require.NoError(t, err)
	// check that ACL was updated)
	require.NoError(t, ElementsEqual(
		entries(alice, "owner", bob, "owner", "carol", "reader"),
		GetACL(t, aliceClient, dataRepo)))
}

// TestGetSetReverse creates two users, alice and bob, and gives bob gradually
// shrinking privileges, checking what bob can and can't do after each change
func TestGetSetReverse(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	alice, bob := tu.UniqueString("alice"), tu.uniqueString("bob")
	aliceClient, bobClient := getPachClient(t, alice), getPachClient(t, bob)

	// create repo, and check that alice is the owner of the new repo
	dataRepo := tu.UniqueString("TestGetSetReverse")
	require.NoError(t, aliceClient.CreateRepo(dataRepo))
	require.NoError(t, ElementsEqual(
		entries(alice, "owner"), GetACL(t, aliceClient, dataRepo)))

	// Add data to repo (alice can write). Make sure alice can read also.
	commit, err := aliceClient.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = aliceClient.PutFile(dataRepo, commit.ID, "/file", strings.NewReader("lorem ipsum"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(dataRepo, commit.ID)) // # commits = 1
	buf := &bytes.Buffer{}
	require.NoError(t, aliceClient.GetFile(dataRepo, "master", "/file", 0, 0, buf))
	require.Equal(t, "lorem ipsum", buf.String())

	//////////
	/// alice adds bob to the ACL as an owner
	_, err = aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Repo:     dataRepo,
		Username: bob,
		Scope:    auth.Scope_OWNER,
	})
	require.NoError(t, err)
	// bob can read
	buf.Reset()
	require.NoError(t, bobClient.GetFile(dataRepo, "master", "/file", 0, 0, buf))
	require.Equal(t, "lorem ipsum", buf.String())
	// bob can write
	commit, err = bobClient.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, bobClient.FinishCommit(dataRepo, commit.ID))
	require.Equal(t, 2, CommitCnt(t, aliceClient, dataRepo)) // check that a new commit was created
	// bob can update the ACL
	_, err = bobClient.SetScope(bobClient.Ctx(), &auth.SetScopeRequest{
		Repo:     dataRepo,
		Username: "carol",
		Scope:    auth.Scope_READER,
	})
	require.NoError(t, err)
	// check that ACL was updated)
	require.NoError(t, ElementsEqual(
		entries(alice, "owner", bob, "owner", "carol", "reader"),
		GetACL(t, aliceClient, dataRepo)))

	// clear carol
	aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Repo:     dataRepo,
		Username: "carol",
		Scope:    auth.Scope_NONE,
	})
	require.NoError(t, ElementsEqual(
		entries(alice, "owner", bob, "owner"), GetACL(t, aliceClient, dataRepo)))

	//////////
	/// alice adds bob to the ACL as a writer
	_, err = aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Repo:     dataRepo,
		Username: bob,
		Scope:    auth.Scope_WRITER,
	})
	require.NoError(t, err)
	// bob can read
	buf.Reset()
	require.NoError(t, bobClient.GetFile(dataRepo, "master", "/file", 0, 0, buf))
	require.Equal(t, "lorem ipsum", buf.String())
	// bob can write
	commit, err = bobClient.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, bobClient.FinishCommit(dataRepo, commit.ID))
	require.Equal(t, 3, CommitCnt(t, aliceClient, dataRepo)) // check that a new commit was created
	// bob can't update the ACL
	_, err = bobClient.SetScope(bobClient.Ctx(), &auth.SetScopeRequest{
		Repo:     dataRepo,
		Username: "carol",
		Scope:    auth.Scope_READER,
	})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	// check that ACL wasn't updated)
	require.NoError(t, ElementsEqual(
		entries(alice, "owner", bob, "writer"), GetACL(t, aliceClient, dataRepo)))

	//////////
	/// alice adds bob to the ACL as a reader (alice can modify ACL)
	_, err = aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Repo:     dataRepo,
		Username: bob,
		Scope:    auth.Scope_READER,
	})
	require.NoError(t, err)
	// bob can read
	buf.Reset()
	require.NoError(t, bobClient.GetFile(dataRepo, "master", "/file", 0, 0, buf))
	require.Equal(t, "lorem ipsum", buf.String())
	// bob can't write
	_, err = bobClient.StartCommit(dataRepo, "master")
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.Equal(t, 3, CommitCnt(t, aliceClient, dataRepo)) // check that no commits were created
	// bob can't update the ACL
	_, err = bobClient.SetScope(bobClient.Ctx(), &auth.SetScopeRequest{
		Repo:     dataRepo,
		Username: "carol",
		Scope:    auth.Scope_READER,
	})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	// check that ACL wasn't updated)
	require.NoError(t, ElementsEqual(
		entries(alice, "owner", bob, "reader"), GetACL(t, aliceClient, dataRepo)))

	//////////
	/// alice revokes all of bob's privileges
	_, err = aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Repo:     dataRepo,
		Username: bob,
		Scope:    auth.Scope_NONE,
	})
	require.NoError(t, err)
	// bob can't read
	err = bobClient.GetFile(dataRepo, "master", "/file", 0, 0, buf)
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	// bob can't write
	_, err = bobClient.StartCommit(dataRepo, "master")
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.Equal(t, 3, CommitCnt(t, aliceClient, dataRepo)) // check that no commits were created
	// bob can't update the ACL
	_, err = bobClient.SetScope(bobClient.Ctx(), &auth.SetScopeRequest{
		Repo:     dataRepo,
		Username: "carol",
		Scope:    auth.Scope_READER,
	})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	// check that ACL wasn't updated)
	require.NoError(t, ElementsEqual(
		entries(alice, "owner"), GetACL(t, aliceClient, dataRepo)))
}

func TestCreateAndUpdatePipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	type createArgs struct {
		client     *client.APIClient
		name, repo string
		update     bool
	}
	createPipeline := func(args createArgs) error {
		return args.client.CreatePipeline(
			args.name,
			"", // default image: ubuntu:16.04
			[]string{"bash"},
			[]string{"cp /pfs/*/* /pfs/out/"},
			&pps.ParallelismSpec{Constant: 1},
			client.NewAtomInput(args.repo, "/*"),
			"", // default output branch: master
			args.update,
		)
	}
	alice, bob := tu.UniqueString("alice"), tu.uniqueString("bob")
	aliceClient, bobClient := getPachClient(t, alice), getPachClient(t, bob)

	// create repo, and check that alice is the owner of the new repo
	dataRepo := tu.UniqueString("TestCreateAndUpdatePipeline")
	require.NoError(t, aliceClient.CreateRepo(dataRepo))
	require.NoError(t, ElementsEqual(
		entries(alice, "owner"), GetACL(t, aliceClient, dataRepo)))

	// alice can create a pipeline (she owns the input repo)
	pipelineName := tu.UniqueString("alice-pipeline")
	require.NoError(t, createPipeline(createArgs{
		client: aliceClient,
		name:   pipelineName,
		repo:   dataRepo,
	}))
	require.OneOfEquals(t, pipelineName, PipelineNames(t, aliceClient))
	// check that alice owns the output repo too)
	require.NoError(t, ElementsEqual(
		entries(alice, "owner"), GetACL(t, aliceClient, pipelineName)))

	// Make sure alice's pipeline runs successfully
	commit, err := aliceClient.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = aliceClient.PutFile(dataRepo, commit.ID, tu.UniqueString("/file"),
		strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(dataRepo, commit.ID))
	iter, err := aliceClient.FlushCommit(
		[]*pfs.Commit{commit},
		[]*pfs.Repo{{Name: pipelineName}},
	)
	require.NoError(t, err)
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := iter.Next()
		return err
	})

	// bob can't create a pipeline
	badPipeline := tu.UniqueString("bob-bad")
	err = createPipeline(createArgs{
		client: bobClient,
		name:   badPipeline,
		repo:   dataRepo,
	})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.NoneEquals(t, badPipeline, PipelineNames(t, aliceClient))

	// alice adds bob as a reader of the input repo
	_, err = aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Repo:     dataRepo,
		Username: bob,
		Scope:    auth.Scope_READER,
	})
	require.NoError(t, err)

	// now bob can create a pipeline
	goodPipeline := tu.UniqueString("bob-good")
	require.NoError(t, createPipeline(createArgs{
		client: bobClient,
		name:   goodPipeline,
		repo:   dataRepo,
	}))
	require.OneOfEquals(t, goodPipeline, PipelineNames(t, aliceClient))
	// check that bob owns the output repo too)
	require.NoError(t, ElementsEqual(
		entries(bob, "owner"), GetACL(t, bobClient, goodPipeline)))

	// Make sure bob's pipeline runs successfully
	commit, err = aliceClient.StartCommit(dataRepo, "master")
	_, err = aliceClient.PutFile(dataRepo, commit.ID, tu.UniqueString("/file"),
		strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(dataRepo, commit.ID))
	iter, err = bobClient.FlushCommit(
		[]*pfs.Commit{commit},
		[]*pfs.Repo{{Name: goodPipeline}},
	)
	require.NoError(t, err)
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := iter.Next()
		return err
	})
	require.NoError(t, err)

	// bob can't update alice's pipeline
	infoBefore, err := aliceClient.InspectPipeline(pipelineName)
	require.NoError(t, err)
	err = createPipeline(createArgs{
		client: bobClient,
		name:   pipelineName,
		repo:   dataRepo,
		update: true,
	})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	infoAfter, err := aliceClient.InspectPipeline(pipelineName)
	require.NoError(t, err)
	require.Equal(t, infoBefore.Version, infoAfter.Version)

	// alice adds bob as a writer of the output repo, and removes him as a reader
	// of the input repo
	_, err = aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Repo:     pipelineName,
		Username: bob,
		Scope:    auth.Scope_WRITER,
	})
	require.NoError(t, err)
	require.NoError(t, ElementsEqual(
		entries(alice, "owner", bob, "writer"),
		GetACL(t, aliceClient, pipelineName)))

	_, err = aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Repo:     dataRepo,
		Username: bob,
		Scope:    auth.Scope_NONE,
	})
	require.NoError(t, err)
	require.NoError(t, ElementsEqual(
		entries(alice, "owner"), GetACL(t, aliceClient, dataRepo)))

	// bob still can't update alice's pipeline
	infoBefore, err = aliceClient.InspectPipeline(pipelineName)
	require.NoError(t, err)
	err = createPipeline(createArgs{
		client: bobClient,
		name:   pipelineName,
		repo:   dataRepo,
		update: true,
	})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	infoAfter, err = aliceClient.InspectPipeline(pipelineName)
	require.NoError(t, err)
	require.Equal(t, infoBefore.Version, infoAfter.Version)

	// alice re-adds bob as a reader of the input repo
	_, err = aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Repo:     dataRepo,
		Username: bob,
		Scope:    auth.Scope_READER,
	})
	require.NoError(t, err)
	require.NoError(t, ElementsEqual(
		entries(alice, "owner", bob, "reader"), GetACL(t, aliceClient, dataRepo)))

	// now bob can update alice's pipeline
	infoBefore, err = aliceClient.InspectPipeline(pipelineName)
	require.NoError(t, err)
	err = createPipeline(createArgs{
		client: bobClient,
		name:   pipelineName,
		repo:   dataRepo,
		update: true,
	})
	require.NoError(t, err)
	infoAfter, err = aliceClient.InspectPipeline(pipelineName)
	require.NoError(t, err)
	require.NotEqual(t, infoBefore.Version, infoAfter.Version)

	// Make sure the updated pipeline runs successfully
	commit, err = aliceClient.StartCommit(dataRepo, "master")
	_, err = aliceClient.PutFile(dataRepo, commit.ID, tu.UniqueString("/file"),
		strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(dataRepo, commit.ID))
	iter, err = bobClient.FlushCommit(
		[]*pfs.Commit{commit},
		[]*pfs.Repo{{Name: pipelineName}},
	)
	require.NoError(t, err)
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := iter.Next()
		return err
	})
}

func TestPipelineMultipleInputs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	type createArgs struct {
		client *client.APIClient
		name   string
		input  *pps.Input
		update bool
	}
	createPipeline := func(args createArgs) error {
		return args.client.CreatePipeline(
			args.name,
			"", // default image: ubuntu:16.04
			[]string{"bash"},
			[]string{"echo \"work\" >/pfs/out/x"},
			&pps.ParallelismSpec{Constant: 1},
			args.input,
			"", // default output branch: master
			args.update,
		)
	}
	alice, bob := tu.UniqueString("alice"), tu.uniqueString("bob")
	aliceClient, bobClient := getPachClient(t, alice), getPachClient(t, bob)

	// create two repos, and check that alice is the owner of the new repos
	dataRepo1 := tu.UniqueString("TestPipelineMultipleInputs")
	dataRepo2 := tu.UniqueString("TestPipelineMultipleInputs")
	require.NoError(t, aliceClient.CreateRepo(dataRepo1))
	require.NoError(t, aliceClient.CreateRepo(dataRepo2))
	require.NoError(t, ElementsEqual(
		entries(alice, "owner"), GetACL(t, aliceClient, dataRepo1)))
	require.NoError(t, ElementsEqual(
		entries(alice, "owner"), GetACL(t, aliceClient, dataRepo2)))

	// alice can create a cross-pipeline with both inputs
	aliceCrossPipeline := tu.UniqueString("alice-pipeline-cross")
	require.NoError(t, createPipeline(createArgs{
		client: aliceClient,
		name:   aliceCrossPipeline,
		input: client.NewCrossInput(
			client.NewAtomInput(dataRepo1, "/*"),
			client.NewAtomInput(dataRepo2, "/*"),
		),
	}))
	require.OneOfEquals(t, aliceCrossPipeline, PipelineNames(t, aliceClient))
	// check that alice owns the output repo too)
	require.NoError(t, ElementsEqual(
		entries(alice, "owner"), GetACL(t, aliceClient, aliceCrossPipeline)))

	// alice can create a union-pipeline with both inputs
	aliceUnionPipeline := tu.UniqueString("alice-pipeline-union")
	require.NoError(t, createPipeline(createArgs{
		client: aliceClient,
		name:   aliceUnionPipeline,
		input: client.NewUnionInput(
			client.NewAtomInput(dataRepo1, "/*"),
			client.NewAtomInput(dataRepo2, "/*"),
		),
	}))
	require.OneOfEquals(t, aliceUnionPipeline, PipelineNames(t, aliceClient))
	// check that alice owns the output repo too)
	require.NoError(t, ElementsEqual(
		entries(alice, "owner"), GetACL(t, aliceClient, aliceUnionPipeline)))

	// alice adds bob as a reader of one of the input repos, but not the other
	_, err := aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Repo:     dataRepo1,
		Username: bob,
		Scope:    auth.Scope_READER,
	})
	require.NoError(t, err)

	// bob cannot create a cross-pipeline with both inputs
	bobCrossPipeline := tu.UniqueString("bob-pipeline-cross")
	err = createPipeline(createArgs{
		client: bobClient,
		name:   bobCrossPipeline,
		input: client.NewCrossInput(
			client.NewAtomInput(dataRepo1, "/*"),
			client.NewAtomInput(dataRepo2, "/*"),
		),
	})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.NoneEquals(t, bobCrossPipeline, PipelineNames(t, aliceClient))

	// bob cannot create a union-pipeline with both inputs
	bobUnionPipeline := tu.UniqueString("bob-pipeline-union")
	err = createPipeline(createArgs{
		client: bobClient,
		name:   bobUnionPipeline,
		input: client.NewUnionInput(
			client.NewAtomInput(dataRepo1, "/*"),
			client.NewAtomInput(dataRepo2, "/*"),
		),
	})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.NoneEquals(t, bobUnionPipeline, PipelineNames(t, aliceClient))

	// alice adds bob as a writer of her pipeline's output
	_, err = aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Repo:     aliceCrossPipeline,
		Username: bob,
		Scope:    auth.Scope_WRITER,
	})
	require.NoError(t, err)

	// bob can update alice's pipeline if he removes one of the inputs
	infoBefore, err := aliceClient.InspectPipeline(aliceCrossPipeline)
	require.NoError(t, createPipeline(createArgs{
		client: bobClient,
		name:   aliceCrossPipeline,
		input: client.NewCrossInput(
			// This cross input deliberately only has one element, to make sure it's
			// not simply rejected for having a cross input
			client.NewAtomInput(dataRepo1, "/*"),
		),
		update: true,
	}))
	infoAfter, err := aliceClient.InspectPipeline(aliceCrossPipeline)
	require.NoError(t, err)
	require.NotEqual(t, infoBefore.Version, infoAfter.Version)

	// bob cannot update alice's to put the second input back
	infoBefore, err = aliceClient.InspectPipeline(aliceCrossPipeline)
	err = createPipeline(createArgs{
		client: bobClient,
		name:   aliceCrossPipeline,
		input: client.NewCrossInput(
			client.NewAtomInput(dataRepo1, "/*"),
			client.NewAtomInput(dataRepo2, "/*"),
		),
		update: true,
	})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	infoAfter, err = aliceClient.InspectPipeline(aliceCrossPipeline)
	require.NoError(t, err)
	require.Equal(t, infoBefore.Version, infoAfter.Version)

	// alice adds bob as a reader of the second input
	_, err = aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Repo:     dataRepo2,
		Username: bob,
		Scope:    auth.Scope_READER,
	})
	require.NoError(t, err)

	// bob can now update alice's to put the second input back
	infoBefore, err = aliceClient.InspectPipeline(aliceCrossPipeline)
	require.NoError(t, createPipeline(createArgs{
		client: bobClient,
		name:   aliceCrossPipeline,
		input: client.NewCrossInput(
			client.NewAtomInput(dataRepo1, "/*"),
			client.NewAtomInput(dataRepo2, "/*"),
		),
		update: true,
	}))
	infoAfter, err = aliceClient.InspectPipeline(aliceCrossPipeline)
	require.NoError(t, err)
	require.NotEqual(t, infoBefore.Version, infoAfter.Version)

	// bob can create a cross-pipeline with both inputs
	require.NoError(t, createPipeline(createArgs{
		client: bobClient,
		name:   bobCrossPipeline,
		input: client.NewCrossInput(
			client.NewAtomInput(dataRepo1, "/*"),
			client.NewAtomInput(dataRepo2, "/*"),
		),
	}))
	require.OneOfEquals(t, bobCrossPipeline, PipelineNames(t, aliceClient))

	// bob can create a union-pipeline with both inputs
	require.NoError(t, createPipeline(createArgs{
		client: bobClient,
		name:   bobUnionPipeline,
		input: client.NewUnionInput(
			client.NewAtomInput(dataRepo1, "/*"),
			client.NewAtomInput(dataRepo2, "/*"),
		),
	}))
	require.OneOfEquals(t, bobUnionPipeline, PipelineNames(t, aliceClient))

}

func TestPipelineRevoke(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	alice, bob := tu.UniqueString("alice"), tu.uniqueString("bob")
	aliceClient, bobClient := getPachClient(t, alice), getPachClient(t, bob)

	// alice creates a repo, and adds bob as a reader
	repo := tu.UniqueString("TestPipelineRevoke")
	require.NoError(t, aliceClient.CreateRepo(repo))
	_, err := aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Repo:     repo,
		Username: bob,
		Scope:    auth.Scope_READER,
	})
	require.NoError(t, err)
	require.NoError(t, ElementsEqual(
		entries(alice, "owner", bob, "reader"), GetACL(t, aliceClient, repo)))

	// bob creates a pipeline
	pipeline := tu.UniqueString("bob-pipeline")
	require.NoError(t, bobClient.CreatePipeline(
		pipeline,
		"", // default image: ubuntu:16.04
		[]string{"bash"},
		[]string{"cp /pfs/*/* /pfs/out/"},
		&pps.ParallelismSpec{Constant: 1},
		client.NewAtomInput(repo, "/*"),
		"", // default output branch: master
		false,
	))
	require.NoError(t, ElementsEqual(
		entries(bob, "owner"), GetACL(t, bobClient, pipeline)))

	// alice commits to the input repo, and the pipeline runs successfully
	commit, err := aliceClient.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = aliceClient.PutFile(repo, commit.ID, "/file1", strings.NewReader("test"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.ID))
	iter, err := bobClient.FlushCommit(
		[]*pfs.Commit{commit},
		[]*pfs.Repo{{Name: pipeline}},
	)
	require.NoError(t, err)
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := iter.Next()
		return err
	})

	// alice removes bob as a reader of her repo
	_, err = aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Repo:     repo,
		Username: bob,
		Scope:    auth.Scope_NONE,
	})
	require.NoError(t, err)
	require.NoError(t, ElementsEqual(
		entries(alice, "owner"), GetACL(t, aliceClient, repo)))

	// alice commits to the input repo, and bob's pipeline does not run
	commit, err = aliceClient.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = aliceClient.PutFile(repo, commit.ID, "/file1", strings.NewReader("test"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.ID))

	doneCh := make(chan struct{})
	go func() {
		iter, err = aliceClient.FlushCommit(
			[]*pfs.Commit{commit},
			[]*pfs.Repo{{Name: pipeline}},
		)
		require.NoError(t, err)
		_, err = iter.Next()
		require.NoError(t, err)
		close(doneCh)
	}()
	select {
	case <-doneCh:
		t.Fatal("pipeline should not be able to finish with no access")
	case <-time.After(30 * time.Second):
	}

	// alice updates bob's pipline, and now it runs
	_, err = bobClient.SetScope(bobClient.Ctx(), &auth.SetScopeRequest{
		Repo:     pipeline,
		Username: alice,
		Scope:    auth.Scope_WRITER,
	})
	require.NoError(t, err)
	require.NoError(t, ElementsEqual(
		entries(bob, "owner", alice, "writer"), GetACL(t, bobClient, pipeline)))
	require.NoError(t, aliceClient.CreatePipeline(
		pipeline,
		"", // default image: ubuntu:16.04
		[]string{"bash"},
		[]string{"cp /pfs/*/* /pfs/out/"},
		&pps.ParallelismSpec{Constant: 1},
		client.NewAtomInput(repo, "/*"),
		"", // default output branch: master
		true,
	))

	// Pipeline now finishes successfully
	iter, err = aliceClient.FlushCommit(
		[]*pfs.Commit{commit},
		[]*pfs.Repo{{Name: pipeline}},
	)
	require.NoError(t, err)
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := iter.Next()
		return err
	})
}

func TestDeletePipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	alice, bob := tu.UniqueString("alice"), tu.uniqueString("bob")
	aliceClient, bobClient := getPachClient(t, alice), getPachClient(t, bob)

	// alice creates a repo
	repo := tu.UniqueString("TestDeletePipeline")
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.NoError(t, ElementsEqual(
		entries(alice, "owner"), GetACL(t, aliceClient, repo)))

	// alice creates a pipeline
	pipeline := tu.UniqueString("alice-pipeline")
	require.NoError(t, aliceClient.CreatePipeline(
		pipeline,
		"", // default image: ubuntu:16.04
		[]string{"bash"},
		[]string{"cp /pfs/*/* /pfs/out/"},
		&pps.ParallelismSpec{Constant: 1},
		client.NewAtomInput(repo, "/*"),
		"", // default output branch: master
		false,
	))
	// Make sure the input and output repos have non-empty ACLs
	require.NoError(t, ElementsEqual(
		entries(alice, "owner"), GetACL(t, aliceClient, repo)))
	require.NoError(t, ElementsEqual(
		entries(alice, "owner"), GetACL(t, aliceClient, pipeline)))

	// alice deletes the pipeline (owner of the input and output repos can delete)
	require.NoError(t, aliceClient.DeletePipeline(pipeline, true))

	// Make sure the remaining input and output repos *still* have non-empty ACLs
	require.NoError(t, ElementsEqual(
		entries(alice, "owner"), GetACL(t, aliceClient, repo)))
	require.NoError(t, ElementsEqual(
		entries(alice, "owner"), GetACL(t, aliceClient, pipeline)))

	// alice deletes the output repo (make sure the output repo's ACL is gone)
	require.NoError(t, aliceClient.DeleteRepo(pipeline, false))
	require.NoError(t, ElementsEqual(entries(), GetACL(t, aliceClient, pipeline)))

	// alice deletes the input repo (make sure the input repo's ACL is gone)
	require.NoError(t, aliceClient.DeleteRepo(repo, false))
	require.NoError(t, ElementsEqual(entries(), GetACL(t, aliceClient, repo)))

	// alice creates another repo
	repo = tu.UniqueString("TestDeletePipeline")
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.NoError(t, ElementsEqual(
		entries(alice, "owner"), GetACL(t, aliceClient, repo)))

	// alice creates another pipeline
	pipeline = tu.UniqueString("alice-pipeline")
	require.NoError(t, aliceClient.CreatePipeline(
		pipeline,
		"", // default image: ubuntu:16.04
		[]string{"bash"},
		[]string{"cp /pfs/*/* /pfs/out/"},
		&pps.ParallelismSpec{Constant: 1},
		client.NewAtomInput(repo, "/*"),
		"", // default output branch: master
		false,
	))

	// bob can't delete alice's pipeline
	err := bobClient.DeletePipeline(pipeline, true)
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())

	// alice adds bob as a reader of the input repo
	_, err = aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Repo:     repo,
		Username: bob,
		Scope:    auth.Scope_READER,
	})
	require.NoError(t, err)
	require.NoError(t, ElementsEqual(
		entries(alice, "owner", bob, "reader"), GetACL(t, aliceClient, repo)))

	// bob still can't delete alice's pipeline
	err = bobClient.DeletePipeline(pipeline, true)
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())

	// alice removes bob as a reader of the input repo and adds bob as a writer of
	// the output repo
	_, err = aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Repo:     repo,
		Username: bob,
		Scope:    auth.Scope_NONE,
	})
	require.NoError(t, err)

	require.NoError(t, ElementsEqual(
		entries(alice, "owner"), GetACL(t, aliceClient, repo)))
	_, err = aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Repo:     pipeline,
		Username: bob,
		Scope:    auth.Scope_WRITER,
	})
	require.NoError(t, err)
	require.NoError(t, ElementsEqual(
		entries(alice, "owner", bob, "writer"), GetACL(t, aliceClient, pipeline)))

	// bob still can't delete alice's pipeline
	err = bobClient.DeletePipeline(pipeline, true)
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())

	// alice re-adds bob as a reader of the input repo
	_, err = aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Repo:     repo,
		Username: bob,
		Scope:    auth.Scope_READER,
	})
	require.NoError(t, err)
	require.NoError(t, ElementsEqual(
		entries(alice, "owner", bob, "reader"), GetACL(t, aliceClient, repo)))

	// bob still can't delete alice's pipeline
	err = bobClient.DeletePipeline(pipeline, true)
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())

	// alice adds bob as an owner of the output repo
	_, err = aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Repo:     pipeline,
		Username: bob,
		Scope:    auth.Scope_OWNER,
	})
	require.NoError(t, err)
	require.NoError(t, ElementsEqual(
		entries(alice, "owner", bob, "owner"), GetACL(t, aliceClient, pipeline)))

	// finally bob can delete alice's pipeline
	require.NoError(t, bobClient.DeletePipeline(pipeline, true))
}

// Test ListRepo checks that the auth information returned by ListRepo and
// InspectRepo is correct.
// TODO(msteffen): This should maybe go in pachyderm_test, since ListRepo isn't
// an auth API call
func TestListAndInspectRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	alice, bob := tu.UniqueString("alice"), tu.uniqueString("bob")
	aliceClient, bobClient := getPachClient(t, alice), getPachClient(t, bob)

	// alice creates a repo and makes Bob a writer
	repoWriter := tu.UniqueString("TestListRepo")
	require.NoError(t, aliceClient.CreateRepo(repoWriter))
	_, err := aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Repo:     repoWriter,
		Username: bob,
		Scope:    auth.Scope_WRITER,
	})
	require.NoError(t, err)
	require.NoError(t, ElementsEqual(
		entries(alice, "owner", bob, "writer"), GetACL(t, aliceClient, repoWriter)))

	// alice creates a repo and makes Bob a reader
	repoReader := tu.UniqueString("TestListRepo")
	require.NoError(t, aliceClient.CreateRepo(repoReader))
	_, err = aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Repo:     repoReader,
		Username: bob,
		Scope:    auth.Scope_READER,
	})
	require.NoError(t, err)
	require.NoError(t, ElementsEqual(
		entries(alice, "owner", bob, "reader"), GetACL(t, aliceClient, repoReader)))

	// alice creates a repo and gives Bob no access privileges
	repoNone := tu.UniqueString("TestListRepo")
	require.NoError(t, aliceClient.CreateRepo(repoNone))
	require.NoError(t, ElementsEqual(
		entries(alice, "owner"), GetACL(t, aliceClient, repoNone)))

	// bob creates a repo
	repoOwner := tu.UniqueString("TestListRepo")
	require.NoError(t, bobClient.CreateRepo(repoOwner))
	require.NoError(t, ElementsEqual(
		entries(bob, "owner"), GetACL(t, bobClient, repoOwner)))

	// Bob calls ListRepo, and the response must indicate the correct access scope
	// for each repo (because other tests have run, we may see repos besides the
	// above. Bob's access to those should be NONE
	listResp, err := bobClient.PfsAPIClient.ListRepo(bobClient.Ctx(),
		&pfs.ListRepoRequest{})
	require.NoError(t, err)
	expectedAccess := map[string]auth.Scope{
		repoOwner:  auth.Scope_OWNER,
		repoWriter: auth.Scope_WRITER,
		repoReader: auth.Scope_READER,
	}
	for _, info := range listResp.RepoInfo {
		require.Equal(t, expectedAccess[info.Repo.Name], info.AuthInfo.AccessLevel)
	}

	for _, name := range []string{repoOwner, repoWriter, repoReader, repoNone} {
		inspectResp, err := bobClient.PfsAPIClient.InspectRepo(bobClient.Ctx(),
			&pfs.InspectRepoRequest{
				Repo: &pfs.Repo{Name: name},
			})
		require.NoError(t, err)
		require.Equal(t, expectedAccess[name], inspectResp.AuthInfo.AccessLevel)
	}
}

func TestUnprivilegedUserCannotMakeSelfOwner(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	alice, bob := tu.UniqueString("alice"), tu.uniqueString("bob")
	aliceClient, bobClient := getPachClient(t, alice), getPachClient(t, bob)

	// alice creates a repo
	repo := tu.UniqueString("TestUnprivilegedUserCannotMakeSelfOwner")
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.NoError(t, ElementsEqual(
		entries(alice, "owner"), GetACL(t, aliceClient, repo)))

	// bob calls SetScope(bob, OWNER) on alice's repo. This should fail
	_, err := bobClient.SetScope(bobClient.Ctx(), &auth.SetScopeRequest{
		Repo:     repo,
		Scope:    auth.Scope_OWNER,
		Username: bob,
	})
	require.YesError(t, err)
	// make sure ACL wasn't updated
	require.NoError(t, ElementsEqual(
		entries(alice, "owner"), GetACL(t, aliceClient, repo)))
}

func TestGetScopeRequiresReader(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	alice, bob := tu.UniqueString("alice"), tu.uniqueString("bob")
	aliceClient, bobClient := getPachClient(t, alice), getPachClient(t, bob)

	// alice creates a repo
	repo := tu.UniqueString("TestUnprivilegedUserCannotMakeSelfOwner")
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.NoError(t, ElementsEqual(
		entries(alice, "owner"), GetACL(t, aliceClient, repo)))

	// bob calls GetScope(repo). This should succeed
	resp, err := bobClient.GetScope(bobClient.Ctx(), &auth.GetScopeRequest{
		Repos: []string{repo},
	})
	require.NoError(t, err)
	require.Equal(t, []auth.Scope{auth.Scope_NONE}, resp.Scopes)

	// bob calls GetScope(repo, alice). This should fail because bob isn't a
	// READER
	_, err = bobClient.GetScope(bobClient.Ctx(),
		&auth.GetScopeRequest{
			Repos:    []string{repo},
			Username: alice,
		})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
}

// TestListRepoNotLoggedInError makes sure that if a user isn't logged in, and
// they call ListRepo(), they get an error.
func TestListRepoNotLoggedInError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	alice := tu.UniqueString("alice")
	aliceClient, anonClient := getPachClient(t, alice), getPachClient(t, "")

	// alice creates a repo
	repoWriter := tu.UniqueString("TestListRepo")
	require.NoError(t, aliceClient.CreateRepo(repoWriter))
	require.NoError(t, ElementsEqual(
		entries(alice, "owner"), GetACL(t, aliceClient, repoWriter)))

	// Anon (non-logged-in user) calls ListRepo, and must recieve an error
	_, err := anonClient.PfsAPIClient.ListRepo(anonClient.Ctx(),
		&pfs.ListRepoRequest{})
	require.YesError(t, err)
	require.Matches(t, "auth token not found in context", err.Error())
}

// TestListRepoNoAuthInfoIfDeactivated tests that if auth isn't activated, then
// ListRepo returns RepoInfos where AuthInfo isn't set (i.e. is nil)
func TestListRepoNoAuthInfoIfDeactivated(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	// Dont't run this test in parallel, since it deactivates the auth system
	// globally, so any tests running concurrently will fail
	alice, bob := tu.UniqueString("alice"), tu.uniqueString("bob")
	aliceClient, bobClient := getPachClient(t, alice), getPachClient(t, bob)
	adminClient := getPachClient(t, "admin")

	// alice creates a repo
	repoWriter := tu.UniqueString("TestListRepo")
	require.NoError(t, aliceClient.CreateRepo(repoWriter))

	// bob calls ListRepo, but has NONE access to all repos
	infos, err := bobClient.ListRepo([]string{})
	require.NoError(t, err)
	for _, info := range infos {
		require.Equal(t, auth.Scope_NONE, info.AuthInfo.AccessLevel)
	}

	// Deactivate auth
	_, err = adminClient.Deactivate(adminClient.Ctx(), &auth.DeactivateRequest{})
	require.NoError(t, err)

	// Wait for auth to be deactivated
	require.NoError(t, backoff.Retry(func() error {
		_, err := aliceClient.WhoAmI(aliceClient.Ctx(), &auth.WhoAmIRequest{})
		if err != nil && auth.IsNotActivatedError(err) {
			return nil // WhoAmI should fail when auth is deactivated
		}
		return errors.New("auth is not yet deactivated")
	}, backoff.NewTestingBackOff()))

	// bob calls ListRepo, now AuthInfo isn't set anywhere
	infos, err = bobClient.ListRepo([]string{})
	require.NoError(t, err)
	for _, info := range infos {
		require.Nil(t, info.AuthInfo)
	}
}

// TestCreateRepoAlreadyExistsError tests that creating a repo that already
// exists gives you an error to that effect, even when auth is already
// activated (rather than "access denied")
func TestCreateRepoAlreadyExistsError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	alice, bob := tu.UniqueString("alice"), tu.uniqueString("bob")
	aliceClient, bobClient := getPachClient(t, alice), getPachClient(t, bob)

	// alice creates a repo
	repo := tu.UniqueString("TestCreateRepoAlreadyExistsError")
	require.NoError(t, aliceClient.CreateRepo(repo))

	// bob creates the same repo, and should get an error to the effect that the
	// repo already exists (rather than "access denied")
	err := bobClient.CreateRepo(repo)
	require.YesError(t, err)
	require.Matches(t, "already exists", err.Error())
}

// TestDeleteRepoDoesntExistError tests that if a client calls DeleteRepo on a
// repo that doesn't exist, they get an error notifying them that the repo
// doesn't exist, rather than an auth error
func TestDeleteRepoDoesntExistError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	alice := tu.UniqueString("alice")
	aliceClient := getPachClient(t, alice)

	err := aliceClient.DeleteRepo("dOeSnOtExIsT", false)
	require.YesError(t, err)
	require.Matches(t, "does not exist", err.Error())

	err = aliceClient.DeleteRepo("dOeSnOtExIsT", true)
	require.YesError(t, err)
	require.Matches(t, "does not exist", err.Error())
}

// Creating a pipeline when the output repo already exists gives you an error to
// that effect, even when auth is already activated (rather than "access
// denied")
func TestCreatePipelineRepoAlreadyExistsError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	alice, bob := tu.UniqueString("alice"), tu.uniqueString("bob")
	aliceClient, bobClient := getPachClient(t, alice), getPachClient(t, bob)

	// alice creates a repo
	inputRepo := tu.UniqueString("TestCreatePipelineRepoAlreadyExistsError")
	require.NoError(t, aliceClient.CreateRepo(inputRepo))
	aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Username: bob,
		Scope:    auth.Scope_READER,
		Repo:     inputRepo,
	})
	pipeline := tu.UniqueString("pipeline")
	require.NoError(t, aliceClient.CreateRepo(pipeline))

	// bob creates a pipeline, and should get an error to the effect that the
	// repo already exists (rather than "access denied")
	err := bobClient.CreatePipeline(
		pipeline,
		"", // default image: ubuntu:16.04
		[]string{"bash"},
		[]string{"cp /pfs/*/* /pfs/out/"},
		&pps.ParallelismSpec{Constant: 1},
		client.NewAtomInput(inputRepo, "/*"),
		"",    // default output branch: master
		false, // Don't update -- we want an error
	)
	require.YesError(t, err)
	require.Matches(t, "cannot overwrite repo", err.Error())
}

// TestAuthorizedNoneRole tests that Authorized(user, repo, NONE) yields 'true',
// even for repos with no ACL
func TestAuthorizedNoneRole(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	adminClient := getPachClient(t, "admin")

	// Deactivate auth
	_, err := adminClient.Deactivate(adminClient.Ctx(), &auth.DeactivateRequest{})
	require.NoError(t, err)

	// Wait for auth to be deactivated
	require.NoError(t, backoff.Retry(func() error {
		_, err = adminClient.WhoAmI(adminClient.Ctx(), &auth.WhoAmIRequest{})
		if err != nil && auth.IsNotActivatedError(err) {
			return nil // WhoAmI should fail when auth is deactivated
		}
		return errors.New("auth is not yet deactivated")
	}, backoff.NewTestingBackOff()))

	// alice creates a repo
	repo := tu.UniqueString("TestAuthorizedNoneRole")
	require.NoError(t, adminClient.CreateRepo(repo))

	// Get new pach clients, re-activating auth
	alice := tu.UniqueString("alice")
	aliceClient, adminClient := getPachClient(t, alice), getPachClient(t, "admin")

	// Check that the repo has no ACL
	require.NoError(t, ElementsEqual(entries(), GetACL(t, adminClient, repo)))

	// alice authorizes against it with the 'NONE' scope
	resp, err := aliceClient.Authorize(aliceClient.Ctx(), &auth.AuthorizeRequest{
		Repo:  repo,
		Scope: auth.Scope_NONE,
	})
	require.NoError(t, err)
	require.True(t, resp.Authorized)
}

// TestDeleteAll tests that you must be a cluster admin to call DeleteAll
func TestDeleteAll(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	alice := tu.UniqueString("alice")
	aliceClient, adminClient := getPachClient(t, alice), getPachClient(t, "admin")

	// alice creates a repo
	repo := tu.UniqueString("TestAuthorizedNoneRole")
	require.NoError(t, adminClient.CreateRepo(repo))

	// alice calls DeleteAll, but it fails
	err := aliceClient.DeleteAll()
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())

	// admin calls DeleteAll and succeeds
	require.NoError(t, adminClient.DeleteAll())
}

// TestListDatum tests that you must have READER access to all of job's
// input repos to call ListDatum on that job
func TestListDatum(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	alice, bob := tu.UniqueString("alice"), tu.uniqueString("bob")
	aliceClient, bobClient := getPachClient(t, alice), getPachClient(t, bob)

	// alice creates a repo
	repoA := tu.UniqueString("TestListDatum")
	require.NoError(t, aliceClient.CreateRepo(repoA))
	repoB := tu.UniqueString("TestListDatum")
	require.NoError(t, aliceClient.CreateRepo(repoB))

	// alice creates a pipeline
	pipeline := tu.UniqueString("alice-pipeline")
	require.NoError(t, aliceClient.CreatePipeline(
		pipeline,
		"", // default image: ubuntu:16.04
		[]string{"bash"},
		[]string{"ls /pfs/*/*; cp /pfs/*/* /pfs/out/"},
		&pps.ParallelismSpec{Constant: 1},
		client.NewCrossInput(
			client.NewAtomInput(repoA, "/*"),
			client.NewAtomInput(repoB, "/*"),
		),
		"", // default output branch: master
		false,
	))

	// alice commits to the input repos, and the pipeline runs successfully
	var commit *pfs.Commit
	for i, repo := range []string{repoA, repoB} {
		var err error
		commit, err = aliceClient.StartCommit(repo, "master")
		require.NoError(t, err)
		file := fmt.Sprintf("/file%d", i+1)
		_, err = aliceClient.PutFile(repo, commit.ID, file, strings.NewReader("test"))
		require.NoError(t, err)
		require.NoError(t, aliceClient.FinishCommit(repo, commit.ID))
	}
	iter, err := aliceClient.FlushCommit(
		[]*pfs.Commit{commit},
		[]*pfs.Repo{{Name: pipeline}},
	)
	require.NoError(t, err)
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := iter.Next()
		return err
	})
	jobs, err := bobClient.ListJob(pipeline, nil /*inputs*/, nil /*output*/)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobs))
	jobID := jobs[0].Job.ID

	// bob cannot call ListDatum
	resp, err := bobClient.ListDatum(jobID, 0 /*pageSize*/, 0 /*page*/)
	require.YesError(t, err)
	require.True(t, auth.IsNotAuthorizedError(err), err.Error())

	// alice adds bob to repoA, but bob still can't call GetLogs
	_, err = aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Username: bob,
		Scope:    auth.Scope_READER,
		Repo:     repoA,
	})
	require.NoError(t, err)
	_, err = bobClient.ListDatum(jobID, 0 /*pageSize*/, 0 /*page*/)
	require.YesError(t, err)
	require.True(t, auth.IsNotAuthorizedError(err), err.Error())

	// alice removes bob from repoA and adds bob to repoB, but bob still can't
	// call ListDatum
	_, err = aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Username: bob,
		Scope:    auth.Scope_NONE,
		Repo:     repoA,
	})
	require.NoError(t, err)
	_, err = aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Username: bob,
		Scope:    auth.Scope_READER,
		Repo:     repoB,
	})
	require.NoError(t, err)
	_, err = bobClient.ListDatum(jobID, 0 /*pageSize*/, 0 /*page*/)
	require.YesError(t, err)
	require.True(t, auth.IsNotAuthorizedError(err), err.Error())

	// alice adds bob to repoA, and now bob can call ListDatum
	_, err = aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Username: bob,
		Scope:    auth.Scope_READER,
		Repo:     repoA,
	})
	require.NoError(t, err)
	resp, err = bobClient.ListDatum(jobID, 0 /*pageSize*/, 0 /*page*/)
	require.NoError(t, err)
	files := make(map[string]struct{})
	for _, di := range resp.DatumInfos {
		for _, f := range di.Data {
			files[path.Base(f.File.Path)] = struct{}{}
		}
	}
	require.Equal(t, map[string]struct{}{
		"file1": struct{}{},
		"file2": struct{}{},
	}, files)
}

// TestInspectDatum tests InspectDatum runs even when auth is activated
func TestInspectDatum(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	alice := tu.UniqueString("alice")
	aliceClient := getPachClient(t, alice)

	// alice creates a repo
	repo := tu.UniqueString("TestInspectDatum")
	require.NoError(t, aliceClient.CreateRepo(repo))

	// alice creates a pipeline (we must enable stats for InspectDatum, which
	// means calling the grpc client function directly)
	pipeline := tu.UniqueString("alice-pipeline")
	_, err := aliceClient.PpsAPIClient.CreatePipeline(aliceClient.Ctx(),
		&pps.CreatePipelineRequest{
			Pipeline: &pps.Pipeline{Name: pipeline},
			Transform: &pps.Transform{
				Cmd:   []string{"bash"},
				Stdin: []string{"cp /pfs/*/* /pfs/out/"},
			},
			ParallelismSpec: &pps.ParallelismSpec{Constant: 1},
			Input:           client.NewAtomInput(repo, "/*"),
			EnableStats:     true,
		})
	require.NoError(t, err)

	// alice commits to the input repo, and the pipeline runs successfully
	commit, err := aliceClient.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = aliceClient.PutFile(repo, commit.ID, "/file1", strings.NewReader("test"))
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
	jobs, err := aliceClient.ListJob(pipeline, nil /*inputs*/, nil /*output*/)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobs))
	jobID := jobs[0].Job.ID

	// ListDatum seems like it may return inconsistent results, so sleep until
	// the /stats branch is written
	// TODO(msteffen): verify if this is true, and if so, why
	time.Sleep(5 * time.Second)
	resp, err := aliceClient.ListDatum(jobID, 0 /*pageSize*/, 0 /*page*/)
	require.NoError(t, err)
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		for _, di := range resp.DatumInfos {
			if _, err := aliceClient.InspectDatum(jobID, di.Datum.ID); err != nil {
				continue
			}
		}
		return nil
	})
}

// TestGetLogs tests that you must have READER access to all of a job's input
// repos and READER access to its output repo to call GetLogs()
func TestGetLogs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	alice, bob := tu.UniqueString("alice"), tu.uniqueString("bob")
	aliceClient, bobClient := getPachClient(t, alice), getPachClient(t, bob)

	// alice creates a repo
	repo := tu.UniqueString("TestGetLogs")
	require.NoError(t, aliceClient.CreateRepo(repo))

	// alice creates a pipeline
	pipeline := tu.UniqueString("pipeline")
	require.NoError(t, aliceClient.CreatePipeline(
		pipeline,
		"", // default image: ubuntu:16.04
		[]string{"bash"},
		[]string{"cp /pfs/*/* /pfs/out/"},
		&pps.ParallelismSpec{Constant: 1},
		client.NewAtomInput(repo, "/*"),
		"", // default output branch: master
		false,
	))

	// alice commits to the input repos, and the pipeline runs successfully
	commit, err := aliceClient.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = aliceClient.PutFile(repo, commit.ID, "/file1", strings.NewReader("test"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.ID))
	commitIter, err := aliceClient.FlushCommit(
		[]*pfs.Commit{commit},
		[]*pfs.Repo{{Name: pipeline}},
	)
	require.NoError(t, err)
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := commitIter.Next()
		return err
	})

	// bob cannot call GetLogs
	iter := bobClient.GetLogs(pipeline, "", nil, "", false)
	require.False(t, iter.Next())
	require.YesError(t, iter.Err())
	require.True(t, auth.IsNotAuthorizedError(iter.Err()), iter.Err().Error())

	// bob also can't call GetLogs for the master process
	iter = bobClient.GetLogs(pipeline, "", nil, "", true)
	require.False(t, iter.Next())
	require.YesError(t, iter.Err())
	require.True(t, auth.IsNotAuthorizedError(iter.Err()), iter.Err().Error())

	// alice adds bob to the input repo, but bob still can't call GetLogs
	aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Username: bob,
		Scope:    auth.Scope_READER,
		Repo:     repo,
	})
	iter = bobClient.GetLogs(pipeline, "", nil, "", false)
	require.False(t, iter.Next())
	require.YesError(t, iter.Err())
	require.True(t, auth.IsNotAuthorizedError(iter.Err()), iter.Err().Error())

	// alice removes bob from the input repo and adds bob to the output repo, but
	// bob still can't call GetLogs
	aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Username: bob,
		Scope:    auth.Scope_NONE,
		Repo:     repo,
	})
	aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Username: bob,
		Scope:    auth.Scope_READER,
		Repo:     pipeline,
	})
	iter = bobClient.GetLogs(pipeline, "", nil, "", false)
	require.False(t, iter.Next())
	require.YesError(t, iter.Err())
	require.True(t, auth.IsNotAuthorizedError(iter.Err()), iter.Err().Error())

	// alice adds bob to the output repo, and now bob can call GetLogs
	aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Username: bob,
		Scope:    auth.Scope_READER,
		Repo:     repo,
	})
	iter = bobClient.GetLogs(pipeline, "", nil, "", false)
	iter.Next()
	require.NoError(t, iter.Err())

	// bob can also call GetLogs for the master process
	iter = bobClient.GetLogs(pipeline, "", nil, "", true)
	iter.Next()
	require.NoError(t, iter.Err())
}

// TestGetLogsFromStats tests that GetLogs still works even when stats are
// enabled
func TestGetLogsFromStats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	alice := tu.UniqueString("alice")
	aliceClient := getPachClient(t, alice)

	// alice creates a repo
	repo := tu.UniqueString("TestGetLogsFromStats")
	require.NoError(t, aliceClient.CreateRepo(repo))

	// alice creates a pipeline (we must enable stats for InspectDatum, which
	// means calling the grpc client function directly)
	pipeline := tu.UniqueString("alice")
	_, err := aliceClient.PpsAPIClient.CreatePipeline(aliceClient.Ctx(),
		&pps.CreatePipelineRequest{
			Pipeline: &pps.Pipeline{Name: pipeline},
			Transform: &pps.Transform{
				Cmd:   []string{"bash"},
				Stdin: []string{"cp /pfs/*/* /pfs/out/"},
			},
			ParallelismSpec: &pps.ParallelismSpec{Constant: 1},
			Input:           client.NewAtomInput(repo, "/*"),
			EnableStats:     true,
		})
	require.NoError(t, err)

	// alice commits to the input repo, and the pipeline runs successfully
	commit, err := aliceClient.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = aliceClient.PutFile(repo, commit.ID, "/file1", strings.NewReader("test"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.ID))
	commitItr, err := aliceClient.FlushCommit(
		[]*pfs.Commit{commit},
		[]*pfs.Repo{{Name: pipeline}},
	)
	require.NoError(t, err)
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := commitItr.Next()
		return err
	})
	jobs, err := aliceClient.ListJob(pipeline, nil /*inputs*/, nil /*output*/)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobs))
	jobID := jobs[0].Job.ID

	iter := aliceClient.GetLogs("", jobID, nil, "", false)
	require.True(t, iter.Next())
	require.NoError(t, iter.Err())

	iter = aliceClient.GetLogs("", jobID, nil, "", true)
	iter.Next()
	require.NoError(t, iter.Err())
}
