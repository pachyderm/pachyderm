package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
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

func uniqueString(prefix string) string {
	return prefix + uuid.NewWithoutDashes()[0:12]
}

var (
	clientMapMut sync.Mutex
	clientMap    = make(map[string]*client.APIClient)
)

// getPachClient creates a seed client with a grpc connection to a pachyderm
// cluster, and then enable the auth service in that cluster
func getPachClient(t testing.TB, u string) *client.APIClient {
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
			resp, err := seedClient.AuthAPIClient.Activate(context.Background(),
				&auth.ActivateRequest{GithubUsername: "admin"})
			if err == nil {
				// Activate success -- create "admin" client with response token
				clientMap["admin"] = *clientMap[""]
				clientMap["admin"].SetAuthToken(resp.PachToken)
			} else if !strings.HasSuffix(err.Error(), "already activated") {
				_, ok := clientMap["admin"]
				require.True(t, ok)
				return fmt.Errorf("could not activate auth service: %s", err.Error())
			}
			return nil
		}, backoff.NewTestingBackOff()))

		// Wait for the Pachyderm Auth system to activate
		adminClient := clientMap["admin"]
		resp, err := adminClient.AuthAPIClient.WhoAmI(adminClient.Ctx(),
			&auth.WhoAmIRequest{},
		)
		require.NoError(t, backoff.Retry(func() error {
			if auth.IsNotActivatedError(err) {
				return err
			}
			require.Equal(t, "admin", resp.Username)
			require.True(t, resp.IsAdmin)
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

// acl constructs an auth.ACL struct from a list of the form
// [ user_1, scope_1, user_2, scope_2, ... ]
func acl(items ...string) *auth.ACL {
	if len(items)%2 != 0 {
		panic("cannot create an ACL from an odd number of items")
	}
	if len(items) == 0 {
		return &auth.ACL{}
	}
	result := &auth.ACL{Entries: make(map[string]auth.Scope)}
	for i := 0; i < len(items); i += 2 {
		scope, err := auth.ParseScope(items[i+1])
		if err != nil {
			panic(fmt.Sprintf("could not parse scope: %v", err))
		}
		result.Entries[items[i]] = scope
	}
	return result
}

// GetACL uses the client 'c' to get the ACL protecting the repo 'repo'
// TODO(msteffen) create an auth client?
func GetACL(t *testing.T, c *client.APIClient, repo string) *auth.ACL {
	resp, err := c.AuthAPIClient.GetACL(c.Ctx(), &auth.GetACLRequest{
		Repo: repo,
	})
	require.NoError(t, err)
	return resp.ACL
}

// CommitCnt uses 'c' to get the number of commits made to the repo 'repo'
func CommitCnt(t *testing.T, c *client.APIClient, repo string) int {
	commitList, err := c.ListCommitByRepo(repo)
	require.NoError(t, err)
	return len(commitList)
}

// PipelineNames returns the names of all pipelines that 'c' gets from
// ListPipeline
func PipelineNames(t *testing.T, c *client.APIClient) []string {
	ps, err := c.ListPipeline()
	require.NoError(t, err)
	result := make([]string, len(ps))
	for i, p := range ps {
		result[i] = p.Pipeline.Name
	}
	return result
}

// TestGetSetBasic creates two users, alice and bob, and gives bob gradually
// escalating privileges, checking what bob can and can't do after each change
func TestGetSetBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	alice, bob := uniqueString("alice"), uniqueString("bob")
	aliceClient, bobClient := getPachClient(t, alice), getPachClient(t, bob)

	// create repo, and check that alice is the owner of the new repo
	dataRepo := uniqueString("TestGetSetBasic")
	require.NoError(t, aliceClient.CreateRepo(dataRepo))
	require.Equal(t, acl(alice, "owner"), GetACL(t, aliceClient, dataRepo))

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
	require.Equal(t, acl(alice, "owner"), GetACL(t, aliceClient, dataRepo)) // check that ACL wasn't updated

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
	require.Equal(t, acl(alice, "owner", bob, "reader"),
		GetACL(t, aliceClient, dataRepo)) // check that ACL wasn't updated

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
	require.Equal(t, acl(alice, "owner", bob, "writer"),
		GetACL(t, aliceClient, dataRepo)) // check that ACL wasn't updated

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
	require.Equal(t, acl(alice, "owner", bob, "owner", "carol", "reader"),
		GetACL(t, aliceClient, dataRepo)) // check that ACL was updated
}

// TestGetSetReverse creates two users, alice and bob, and gives bob gradually
// shrinking privileges, checking what bob can and can't do after each change
func TestGetSetReverse(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	alice, bob := uniqueString("alice"), uniqueString("bob")
	aliceClient, bobClient := getPachClient(t, alice), getPachClient(t, bob)

	// create repo, and check that alice is the owner of the new repo
	dataRepo := uniqueString("TestGetSetReverse")
	require.NoError(t, aliceClient.CreateRepo(dataRepo))
	require.Equal(t, acl(alice, "owner"), GetACL(t, aliceClient, dataRepo))

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
	require.Equal(t, acl(alice, "owner", bob, "owner", "carol", "reader"),
		GetACL(t, aliceClient, dataRepo)) // check that ACL was updated

	// clear carol
	aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Repo:     dataRepo,
		Username: "carol",
		Scope:    auth.Scope_NONE,
	})
	require.Equal(t, acl(alice, "owner", bob, "owner"),
		GetACL(t, aliceClient, dataRepo))

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
	require.Equal(t, acl(alice, "owner", bob, "writer"),
		GetACL(t, aliceClient, dataRepo)) // check that ACL wasn't updated

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
	require.Equal(t, acl(alice, "owner", bob, "reader"),
		GetACL(t, aliceClient, dataRepo)) // check that ACL wasn't updated

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
	require.Equal(t, acl(alice, "owner"), GetACL(t, aliceClient, dataRepo)) // check that ACL wasn't updated
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
			"", // default image: ubuntu:14.04
			[]string{"bash"},
			[]string{fmt.Sprintf("cp /pfs/%s/* /pfs/out/", args.repo)},
			&pps.ParallelismSpec{Constant: 1},
			client.NewAtomInput(args.repo, "/*"),
			"", // default output branch: master
			args.update,
		)
	}
	alice, bob := uniqueString("alice"), uniqueString("bob")
	aliceClient, bobClient := getPachClient(t, alice), getPachClient(t, bob)

	// create repo, and check that alice is the owner of the new repo
	dataRepo := uniqueString("TestCreateAndUpdatePipeline")
	require.NoError(t, aliceClient.CreateRepo(dataRepo))
	require.Equal(t, acl(alice, "owner"), GetACL(t, aliceClient, dataRepo))

	// alice can create a pipeline (she owns the input repo)
	pipelineName := uniqueString("alice-pipeline")
	require.NoError(t, createPipeline(createArgs{
		client: aliceClient,
		name:   pipelineName,
		repo:   dataRepo,
	}))
	require.OneOfEquals(t, pipelineName, PipelineNames(t, aliceClient))
	require.Equal(t, acl(alice, "owner"), GetACL(t, aliceClient, pipelineName)) // check that alice owns the output repo too

	// Make sure alice's pipeline runs successfully
	commit, err := aliceClient.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = aliceClient.PutFile(dataRepo, commit.ID, uniqueString("/file"),
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
	badPipeline := uniqueString("bob-bad")
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
	goodPipeline := uniqueString("bob-good")
	require.NoError(t, createPipeline(createArgs{
		client: bobClient,
		name:   goodPipeline,
		repo:   dataRepo,
	}))
	require.OneOfEquals(t, goodPipeline, PipelineNames(t, aliceClient))
	require.Equal(t, acl(bob, "owner"), GetACL(t, bobClient, goodPipeline)) // check that bob owns the output repo too

	// Make sure bob's pipeline runs successfully
	commit, err = aliceClient.StartCommit(dataRepo, "master")
	_, err = aliceClient.PutFile(dataRepo, commit.ID, uniqueString("/file"),
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
	require.Equal(t, acl(alice, "owner", bob, "writer"), GetACL(t, aliceClient, pipelineName))
	_, err = aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Repo:     dataRepo,
		Username: bob,
		Scope:    auth.Scope_NONE,
	})
	require.NoError(t, err)
	require.Equal(t, acl(alice, "owner"), GetACL(t, aliceClient, dataRepo))

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
	require.Equal(t, acl(alice, "owner", bob, "reader"), GetACL(t, aliceClient, dataRepo))

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
	_, err = aliceClient.PutFile(dataRepo, commit.ID, uniqueString("/file"),
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
			"", // default image: ubuntu:14.04
			[]string{"bash"},
			[]string{"echo \"work\" >/pfs/out/x"},
			&pps.ParallelismSpec{Constant: 1},
			args.input,
			"", // default output branch: master
			args.update,
		)
	}
	alice, bob := uniqueString("alice"), uniqueString("bob")
	aliceClient, bobClient := getPachClient(t, alice), getPachClient(t, bob)

	// create two repos, and check that alice is the owner of the new repos
	dataRepo1 := uniqueString("TestPipelineMultipleInputs")
	dataRepo2 := uniqueString("TestPipelineMultipleInputs")
	require.NoError(t, aliceClient.CreateRepo(dataRepo1))
	require.NoError(t, aliceClient.CreateRepo(dataRepo2))
	require.Equal(t, acl(alice, "owner"), GetACL(t, aliceClient, dataRepo1))
	require.Equal(t, acl(alice, "owner"), GetACL(t, aliceClient, dataRepo2))

	// alice can create a cross-pipeline with both inputs
	aliceCrossPipeline := uniqueString("alice-pipeline-cross")
	require.NoError(t, createPipeline(createArgs{
		client: aliceClient,
		name:   aliceCrossPipeline,
		input: client.NewCrossInput(
			client.NewAtomInput(dataRepo1, "/*"),
			client.NewAtomInput(dataRepo2, "/*"),
		),
	}))
	require.OneOfEquals(t, aliceCrossPipeline, PipelineNames(t, aliceClient))
	require.Equal(t, acl(alice, "owner"), GetACL(t, aliceClient, aliceCrossPipeline)) // check that alice owns the output repo too

	// alice can create a union-pipeline with both inputs
	aliceUnionPipeline := uniqueString("alice-pipeline-union")
	require.NoError(t, createPipeline(createArgs{
		client: aliceClient,
		name:   aliceUnionPipeline,
		input: client.NewUnionInput(
			client.NewAtomInput(dataRepo1, "/*"),
			client.NewAtomInput(dataRepo2, "/*"),
		),
	}))
	require.OneOfEquals(t, aliceUnionPipeline, PipelineNames(t, aliceClient))
	require.Equal(t, acl(alice, "owner"), GetACL(t, aliceClient, aliceUnionPipeline)) // check that alice owns the output repo too

	// alice adds bob as a reader of one of the input repos, but not the other
	_, err := aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Repo:     dataRepo1,
		Username: bob,
		Scope:    auth.Scope_READER,
	})
	require.NoError(t, err)

	// bob cannot create a cross-pipeline with both inputs
	bobCrossPipeline := uniqueString("bob-pipeline-cross")
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
	bobUnionPipeline := uniqueString("bob-pipeline-union")
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
	alice, bob := uniqueString("alice"), uniqueString("bob")
	aliceClient, bobClient := getPachClient(t, alice), getPachClient(t, bob)

	// alice creates a repo, and adds bob as a reader
	repo := uniqueString("TestPipelineRevoke")
	require.NoError(t, aliceClient.CreateRepo(repo))
	_, err := aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Repo:     repo,
		Username: bob,
		Scope:    auth.Scope_READER,
	})
	require.NoError(t, err)
	require.Equal(t, acl(alice, "owner", bob, "reader"), GetACL(t, aliceClient, repo))

	// bob creates a pipeline
	pipeline := uniqueString("bob-pipeline")
	require.NoError(t, bobClient.CreatePipeline(
		pipeline,
		"", // default image: ubuntu:14.04
		[]string{"bash"},
		[]string{fmt.Sprintf("cp /pfs/%s/* /pfs/out/", repo)},
		&pps.ParallelismSpec{Constant: 1},
		client.NewAtomInput(repo, "/*"),
		"", // default output branch: master
		false,
	))
	require.Equal(t, acl(bob, "owner"), GetACL(t, bobClient, pipeline))

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
	require.Equal(t, acl(alice, "owner"), GetACL(t, aliceClient, repo))

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
	require.Equal(t, acl(bob, "owner", alice, "writer"), GetACL(t, bobClient, pipeline))
	require.NoError(t, aliceClient.CreatePipeline(
		pipeline,
		"", // default image: ubuntu:14.04
		[]string{"bash"},
		[]string{fmt.Sprintf("cp /pfs/%s/* /pfs/out/", repo)},
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
	alice, bob := uniqueString("alice"), uniqueString("bob")
	aliceClient, bobClient := getPachClient(t, alice), getPachClient(t, bob)

	// alice creates a repo
	repo := uniqueString("TestPipelineRevoke")
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
		"", // default output branch: master
		false,
	))

	// alice deletes the pipeline
	require.NoError(t, aliceClient.DeletePipeline(pipeline, true))

	// alice deletes the output repo
	require.NoError(t, aliceClient.DeleteRepo(pipeline, false))

	// Make sure the output repo's ACL is gone
	_, err := aliceClient.AuthAPIClient.GetACL(aliceClient.Ctx(), &auth.GetACLRequest{
		Repo: pipeline,
	})
	require.YesError(t, err)

	// alice deletes the input repo
	require.NoError(t, aliceClient.DeleteRepo(repo, false))

	// Make sure the input repo's ACL is gone
	_, err = aliceClient.AuthAPIClient.GetACL(aliceClient.Ctx(), &auth.GetACLRequest{
		Repo: repo,
	})
	require.YesError(t, err)

	// alice creates another repo
	repo = uniqueString("TestPipelineRevoke")
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.Equal(t, acl(alice, "owner"), GetACL(t, aliceClient, repo))

	// alice creates another pipeline
	pipeline = uniqueString("alice-pipeline")
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

	// bob can't delete alice's pipeline
	err = bobClient.DeletePipeline(pipeline, true)
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())

	// alice adds bob as a reader of the input repo
	_, err = aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Repo:     repo,
		Username: bob,
		Scope:    auth.Scope_READER,
	})
	require.NoError(t, err)
	require.Equal(t, acl(alice, "owner", bob, "reader"), GetACL(t, aliceClient, repo))

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
	require.Equal(t, acl(alice, "owner"), GetACL(t, aliceClient, repo))
	_, err = aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Repo:     pipeline,
		Username: bob,
		Scope:    auth.Scope_WRITER,
	})
	require.NoError(t, err)
	require.Equal(t, acl(alice, "owner", bob, "writer"), GetACL(t, aliceClient, pipeline))

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
	require.Equal(t, acl(alice, "owner", bob, "reader"), GetACL(t, aliceClient, repo))

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
	require.Equal(t, acl(alice, "owner", bob, "owner"), GetACL(t, aliceClient, pipeline))

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
	alice, bob := uniqueString("alice"), uniqueString("bob")
	aliceClient, bobClient := getPachClient(t, alice), getPachClient(t, bob)

	// alice creates a repo and makes Bob a writer
	repoWriter := uniqueString("TestListRepo")
	require.NoError(t, aliceClient.CreateRepo(repoWriter))
	_, err := aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Repo:     repoWriter,
		Username: bob,
		Scope:    auth.Scope_WRITER,
	})
	require.NoError(t, err)
	require.Equal(t, acl(alice, "owner", bob, "writer"), GetACL(t, aliceClient, repoWriter))

	// alice creates a repo and makes Bob a reader
	repoReader := uniqueString("TestListRepo")
	require.NoError(t, aliceClient.CreateRepo(repoReader))
	_, err = aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Repo:     repoReader,
		Username: bob,
		Scope:    auth.Scope_READER,
	})
	require.NoError(t, err)
	require.Equal(t, acl(alice, "owner", bob, "reader"), GetACL(t, aliceClient, repoReader))

	// alice creates a repo and gives Bob no access privileges
	repoNone := uniqueString("TestListRepo")
	require.NoError(t, aliceClient.CreateRepo(repoNone))
	require.Equal(t, acl(alice, "owner"), GetACL(t, aliceClient, repoNone))

	// bob creates a repo
	repoOwner := uniqueString("TestListRepo")
	require.NoError(t, bobClient.CreateRepo(repoOwner))
	require.Equal(t, acl(bob, "owner"), GetACL(t, bobClient, repoOwner))

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
	alice, bob := uniqueString("alice"), uniqueString("bob")
	aliceClient, bobClient := getPachClient(t, alice), getPachClient(t, bob)

	// alice creates a repo
	repo := uniqueString("TestUnprivilegedUserCannotMakeSelfOwner")
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.Equal(t, acl(alice, "owner"), GetACL(t, aliceClient, repo))

	// bob calls SetScope(bob, OWNER) on alice's repo. This should fail
	_, err := bobClient.SetScope(bobClient.Ctx(), &auth.SetScopeRequest{
		Repo:     repo,
		Scope:    auth.Scope_OWNER,
		Username: bob,
	})
	require.YesError(t, err)
	// make sure ACL wasn't updated
	require.Equal(t, acl(alice, "owner"), GetACL(t, aliceClient, repo))
}

func TestGetScopeRequiresReader(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	alice, bob := uniqueString("alice"), uniqueString("bob")
	aliceClient, bobClient := getPachClient(t, alice), getPachClient(t, bob)

	// alice creates a repo
	repo := uniqueString("TestUnprivilegedUserCannotMakeSelfOwner")
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.Equal(t, acl(alice, "owner"), GetACL(t, aliceClient, repo))

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
	alice := uniqueString("alice")
	aliceClient, anonClient := getPachClient(t, alice), getPachClient(t, "")

	// alice creates a repo
	repoWriter := uniqueString("TestListRepo")
	require.NoError(t, aliceClient.CreateRepo(repoWriter))
	require.Equal(t, acl(alice, "owner"), GetACL(t, aliceClient, repoWriter))

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
	alice, bob := uniqueString("alice"), uniqueString("bob")
	aliceClient, bobClient := getPachClient(t, alice), getPachClient(t, bob)
	adminClient := getPachClient(t, "admin")

	// alice creates a repo
	repoWriter := uniqueString("TestListRepo")
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
	alice, bob := uniqueString("alice"), uniqueString("bob")
	aliceClient, bobClient := getPachClient(t, alice), getPachClient(t, bob)

	// alice creates a repo
	repo := uniqueString("TestCreateRepoAlreadyExistsError")
	require.NoError(t, aliceClient.CreateRepo(repo))

	// bob creates the same repo, and should get an error to the effect that the
	// repo already exists (rather than "access denied")
	err := bobClient.CreateRepo(repo)
	require.YesError(t, err)
	require.Matches(t, "already exists", err.Error())
}

// Creating a pipeline when the output repo already exists gives you an error to
// that effect, even when auth is already activated (rather than "access
// denied")
func TestCreatePipelineRepoAlreadyExistsError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	alice, bob := uniqueString("alice"), uniqueString("bob")
	aliceClient, bobClient := getPachClient(t, alice), getPachClient(t, bob)

	// alice creates a repo
	inputRepo := uniqueString("TestCreatePipelineRepoAlreadyExistsError")
	require.NoError(t, aliceClient.CreateRepo(inputRepo))
	aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Username: bob,
		Scope:    auth.Scope_READER,
		Repo:     inputRepo,
	})
	pipeline := uniqueString("pipeline")
	require.NoError(t, aliceClient.CreateRepo(pipeline))

	// bob creates a pipeline, and should get an error to the effect that the
	// repo already exists (rather than "access denied")
	err := bobClient.CreatePipeline(
		pipeline,
		"", // default image: ubuntu:14.04
		[]string{"bash"},
		[]string{fmt.Sprintf("cp /pfs/%s/* /pfs/out/", inputRepo)},
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
	repo := uniqueString("TestAuthorizedNoneRole")
	require.NoError(t, adminClient.CreateRepo(repo))

	// Get new pach clients, re-activating auth
	alice := uniqueString("alice")
	aliceClient, adminClient := getPachClient(t, alice), getPachClient(t, "admin")

	// Check that the repo has no ACL
	require.Equal(t, acl(), GetACL(t, adminClient, repo))

	// alice authorizes against it with the 'NONE' scope
	resp, err := aliceClient.Authorize(aliceClient.Ctx(), &auth.AuthorizeRequest{
		Repo:  repo,
		Scope: auth.Scope_NONE,
	})
	require.NoError(t, err)
	require.True(t, resp.Authorized)
}
