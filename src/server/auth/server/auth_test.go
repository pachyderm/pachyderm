package auth

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
)

const testActivationCode = `eyJ0b2tlbiI6IntcImV4cGlyeVwiOlwiMjAyNy0wNy0xMlQwM` +
	`zowOTowNi4xODBaXCIsXCJzY29wZXNcIjp7XCJiYXNpY1wiOnRydWV9LFwibmFtZVwiOlwicGF` +
	`jaHlkZXJtRW5naW5lZXJpbmdcIn0iLCJzaWduYXR1cmUiOiJWdjZQbEkrL3RJamlWYUNHMGw0T` +
	`Ud6WldDS2YrUFMyWS9WUzFkZ3plcVdjS3RETlJvdkRnSnd3TXFXbWdCOUs5a2lPemVQRlh4eTh` +
	`3U2dMbTJ4dnBlTHN2bGlsTlc5MEhKbGxxcjhKWEVTbVV4R2tKQldMTHZHak5mYUlHZ0IvZTFEM` +
	`zQzMi95eUVnSW1LZDlpZ3J3RXZsRCtGdW0wa1hqS3Rrb2pPRmhkMDR6RHFEMSt5ZWpsTmRtUzB` +
	`TaDJKWHRTMnFqWk0zTE5lWlpTRldLcEVJTmlXa2dhOTdTNUw2ZVlCdXFZcFJLMTkwd1pXNTVCO` +
	`VFJSHJNNWtDWGQrWUN5aTh0QU9kcFY2a3FMSDNoVGgxVDIwVjYveFNZNUVheHZObm8yRmFYbDU` +
	`yQzRFSWIvZ05RWW8xVExDd1hJN0FYL2lpL0VTckVBQmYzdDlYZmlwWGxleE9OMmhJaWY5dDROZ` +
	`FBaQ1pmYlErbW8vSlQ3Um5VTGpTb2J3alNWVk1qMUozLzZKbmhQRFpFSWNDdlVvUnMyL2M2WUZ` +
	`xOVo1TFRJNkUxV2Q0bE1RczRJYXVsTHVQOEFVa3R3ejBiQmY2dUhPd3VvTlk4UjJ3ZTA1MmUxW` +
	`VVGbmNyUE4wd2ZJVHo5Vm51M1dNcktpaDhhRzNmMzRLb2x0R3hpWXJHL2JZQjgweUFaTytCbzF` +
	`mTTJwaDB0emRXejFLR0lNQUlEbjBFWHU2V0duSUFFUWN1NHVFc1pSVXRzNFhuYk5PTC9vYU1NK` +
	`3RLV3UzdnFMdEhMWWlPaWZHNHpEcUxwYnNNN2NhZGNXWjJ3QzNoZVh6Y1loaUwzMHJlOGJ4MFc` +
	`3Vm1FOSt4elJHZisyNEdvRjFaS1BvaDNhY3hCS0dsZzRxN2JQd0c3QWJESmxkak1HbkVEdz0if` +
	`Q==`

func uniqueString(prefix string) string {
	return prefix + uuid.NewWithoutDashes()[0:12]
}

var clientMapMut sync.Mutex
var clientMap = make(map[string]*client.APIClient)

// getPachClient creates a seed client with a grpc connection to a pachyderm
// cluster, and then activates the auth service in that cluster
func getPachClient(t testing.TB, u string) *client.APIClient {
	// Check if "u" already has a client -- if not create one, and block other
	// concurrent tests from interfering
	func() {
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
			clientMap[""] = seedClient

			// Since this is the first auth client, also activate the auth service
			require.NoError(t, backoff.Retry(func() error {
				_, err = seedClient.Activate(context.Background(),
					&auth.ActivateRequest{
						ActivationCode: testActivationCode,
						Admins:         []string{"admin"},
					})
				if err != nil && !strings.HasSuffix(err.Error(), "already activated") {
					return fmt.Errorf("could not activate auth service: %s", err.Error())
				}
				return nil
			}, backoff.NewTestingBackOff()))
		}

		// Re-use old client for 'u', or create a new one if none exists
		if _, ok := clientMap[u]; !ok {
			userClient := *clientMap[""]
			resp, err := userClient.AuthAPIClient.Authenticate(
				context.Background(),
				&auth.AuthenticateRequest{
					GithubUsername: string(u),
				})
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

func TestValidateActivationCode(t *testing.T) {
	require.NoError(t, validateActivationCode(testActivationCode))
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

func TestCreatePipeline(t *testing.T) {
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
	dataRepo := uniqueString("TestCreatePipeline")
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
	_, err = aliceClient.PutFile(dataRepo, commit.ID, uniqueString("/file"),
		strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(dataRepo, commit.ID))
	iter, err := aliceClient.FlushCommit(
		[]*pfs.Commit{commit},
		[]*pfs.Repo{{Name: pipelineName}},
	)
	require.NoError(t, err)
	_, err = iter.Next()
	require.NoError(t, err)

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
	_, err = iter.Next()
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

	// alice adds bob as a writer of the output repo
	_, err = aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
		Repo:     pipelineName,
		Username: bob,
		Scope:    auth.Scope_WRITER,
	})
	require.NoError(t, err)

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
	_, err = iter.Next()
	require.NoError(t, err)
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
	_, err = iter.Next()
	require.NoError(t, err)

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
	case <-time.Tick(30 * time.Second):
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
	_, err = iter.Next()
	require.NoError(t, err)
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

// TestAdmins tests adding and removing cluster admins. Note that cluster admins
// are global, unlike pipelines and repos, so while this test can run in
// parallel with other auth tests, it cannot run in parallel with other cluster
// admin tests. If any are added, t.Parallel() should be removed here.
func TestAdmin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
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

	// bob cannot get the file
	buf := &bytes.Buffer{}
	err = bobClient.GetFile(repo, "master", "/file", 0, 0, buf)
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())

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

	// now bob can get the file
	buf.Reset()
	require.NoError(t, bobClient.GetFile(repo, "master", "/file", 0, 0, buf))
	require.Matches(t, "test data", buf.String())

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

	// bob can no longer get the file
	buf.Reset()
	err = bobClient.GetFile(repo, "master", "/file", 0, 0, buf)
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())

	// admin cannot remove themselves from the list of cluster admins (otherwise
	// there would be no admins)
	_, err = adminClient.ModifyAdmins(adminClient.Ctx(),
		&auth.ModifyAdminsRequest{Remove: []string{"admin"}})
	require.YesError(t, err)
}
