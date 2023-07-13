//go:build unit_test

package server_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/authdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	internalauth "github.com/pachyderm/pachyderm/v2/src/internal/middleware/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/protoutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/license"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	authserver "github.com/pachyderm/pachyderm/v2/src/server/auth/server"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
)

func envWithAuth(t *testing.T) *realenv.RealEnv {
	t.Helper()
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	tu.ActivateLicense(t, env.PachClient, peerPort)
	_, err := env.PachClient.Enterprise.Activate(env.PachClient.Ctx(),
		&enterprise.ActivateRequest{
			LicenseServer: "grpc://localhost:" + peerPort,
			Id:            "localhost",
			Secret:        "localhost",
		})
	require.NoError(t, err, "activate client should work")
	_, err = env.AuthServer.Activate(env.PachClient.Ctx(), &auth.ActivateRequest{RootToken: tu.RootToken})
	require.NoError(t, err, "activate server should work")
	env.PachClient.SetAuthToken(tu.RootToken)
	require.NoError(t, config.WritePachTokenToConfig(tu.RootToken, false))
	client := env.PachClient.WithCtx(context.Background())
	_, err = client.PfsAPIClient.ActivateAuth(client.Ctx(), &pfs.ActivateAuthRequest{})
	require.NoError(t, err, "should be able to activate auth")
	_, err = client.PpsAPIClient.ActivateAuth(client.Ctx(), &pps.ActivateAuthRequest{})
	require.NoError(t, err, "should be able to activate auth")
	return env
}

// TestGetSetBasic creates two users, alice and bob, and gives bob gradually
// escalating privileges, checking what bob can and can't do after each change
func TestGetSetBasic(t *testing.T) {
	t.Parallel()
	env := envWithAuth(t)
	c := env.PachClient
	alice, bob := tu.Robot(tu.UniqueString("alice")), tu.Robot(tu.UniqueString("bob"))
	aliceClient, bobClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, bob)

	// create repo, and check that alice is the owner of the new repo
	repoName := tu.UniqueString(t.Name())
	repo := client.NewRepo(pfs.DefaultProjectName, repoName)
	require.NoError(t, aliceClient.CreateRepo(pfs.DefaultProjectName, repoName))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, repoName))
	dataCommit := client.NewCommit(pfs.DefaultProjectName, repoName, "master", "")

	// Add data to repo (alice can write). Make sure alice can read also.
	err := aliceClient.PutFile(dataCommit, "/file", strings.NewReader("1"), client.WithAppendPutFile())
	require.NoError(t, err)
	buf := &bytes.Buffer{}
	require.NoError(t, aliceClient.GetFile(dataCommit, "/file", buf))
	require.Equal(t, "1", buf.String())

	//////////
	/// Initially, bob has no privileges
	// bob can't read
	err = bobClient.GetFile(dataCommit, "/file", buf)
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	// bob can't write (check both the standalone form of PutFile and StartCommit)
	err = bobClient.PutFile(dataCommit, "/file", strings.NewReader("lorem ipsum"), client.WithAppendPutFile())
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.Equal(t, 1, tu.CommitCnt(t, aliceClient, repo)) // check that no commits were created
	_, err = bobClient.StartCommit(pfs.DefaultProjectName, repoName, "master")
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.Equal(t, 1, tu.CommitCnt(t, aliceClient, repo)) // check that no commits were created
	// bob can't update the ACL
	err = bobClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, repoName, tu.Robot("carol"), []string{auth.RepoReaderRole})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	// check that ACL wasn't updated)
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, repoName))

	//////////
	/// alice adds bob to the ACL as a reader (alice can modify ACL)
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, repoName, bob, []string{auth.RepoReaderRole}))
	// bob can read
	buf.Reset()
	require.NoError(t, bobClient.GetFile(dataCommit, "/file", buf))
	require.Equal(t, "1", buf.String())
	// bob can't write
	err = bobClient.PutFile(dataCommit, "/file", strings.NewReader("2"), client.WithAppendPutFile())
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.Equal(t, 1, tu.CommitCnt(t, aliceClient, repo)) // check that no commits were created
	_, err = bobClient.StartCommit(pfs.DefaultProjectName, repoName, "master")
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.Equal(t, 1, tu.CommitCnt(t, aliceClient, repo)) // check that no commits were created
	// bob can't update the ACL
	err = bobClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, repoName, tu.Robot("carol"), []string{auth.RepoReaderRole})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	// check that ACL wasn't updated)
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, bob, auth.RepoReaderRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, repoName))

	//////////
	/// alice adds bob to the ACL as a writer
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, repoName, bob, []string{auth.RepoWriterRole}))
	// bob can read
	buf.Reset()
	require.NoError(t, bobClient.GetFile(dataCommit, "/file", buf))
	require.Equal(t, "1", buf.String())
	// bob can write
	err = bobClient.PutFile(dataCommit, "/file", strings.NewReader("2"), client.WithAppendPutFile())
	require.NoError(t, err)
	require.Equal(t, 2, tu.CommitCnt(t, aliceClient, repo)) // check that a new commit was created
	commit, err := bobClient.StartCommit(pfs.DefaultProjectName, repoName, "master")
	require.NoError(t, err)
	require.NoError(t, bobClient.FinishCommit(pfs.DefaultProjectName, repoName, commit.Branch.Name, commit.Id))
	require.Equal(t, 3, tu.CommitCnt(t, aliceClient, repo)) // check that a new commit was created
	// bob can't update the ACL
	err = bobClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, repoName, tu.Robot("carol"), []string{auth.RepoReaderRole})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	// check that ACL wasn't updated)
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, bob, auth.RepoWriterRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, repoName))

	//////////
	/// alice adds bob to the ACL as an owner
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, repoName, bob, []string{auth.RepoOwnerRole}))
	// bob can read
	buf.Reset()
	require.NoError(t, bobClient.GetFile(dataCommit, "/file", buf))
	require.Equal(t, "12", buf.String())
	// bob can write
	err = bobClient.PutFile(dataCommit, "/file", strings.NewReader("3"), client.WithAppendPutFile())
	require.NoError(t, err)
	require.Equal(t, 4, tu.CommitCnt(t, aliceClient, repo)) // check that a new commit was created
	commit, err = bobClient.StartCommit(pfs.DefaultProjectName, repoName, "master")
	require.NoError(t, err)
	require.NoError(t, bobClient.FinishCommit(pfs.DefaultProjectName, repoName, commit.Branch.Name, commit.Id))
	require.Equal(t, 5, tu.CommitCnt(t, aliceClient, repo)) // check that a new commit was created
	// bob can update the ACL
	require.NoError(t, bobClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, repoName, tu.Robot("carol"), []string{auth.RepoReaderRole}))
	// check that ACL was updated)
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, bob, auth.RepoOwnerRole, tu.Robot("carol"), auth.RepoReaderRole),
		tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, repoName))
}

// TestGetSetReverse creates two users, alice and bob, and gives bob gradually
// shrinking privileges, checking what bob can and can't do after each change
func TestGetSetReverse(t *testing.T) {
	t.Parallel()
	env := envWithAuth(t)
	c := env.PachClient
	alice, bob := tu.Robot(tu.UniqueString("alice")), tu.Robot(tu.UniqueString("bob"))
	aliceClient, bobClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, bob)

	// create repo, and check that alice is the owner of the new repo
	repoName := tu.UniqueString(t.Name())
	repo := client.NewRepo(pfs.DefaultProjectName, repoName)
	require.NoError(t, aliceClient.CreateRepo(pfs.DefaultProjectName, repoName))
	require.Equal(t, tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, repoName))
	dataCommit := client.NewCommit(pfs.DefaultProjectName, repoName, "master", "")

	// Add data to repo (alice can write). Make sure alice can read also.
	commit, err := aliceClient.StartCommit(pfs.DefaultProjectName, repoName, "master")
	require.NoError(t, err)
	err = aliceClient.PutFile(commit, "/file", strings.NewReader("1"), client.WithAppendPutFile())
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(pfs.DefaultProjectName, repoName, commit.Branch.Name, commit.Id)) // # commits = 1
	buf := &bytes.Buffer{}
	require.NoError(t, aliceClient.GetFile(dataCommit, "/file", buf))
	require.Equal(t, "1", buf.String())

	//////////
	/// alice adds bob to the ACL as an owner
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, repoName, bob, []string{auth.RepoOwnerRole}))
	// bob can read
	buf.Reset()
	require.NoError(t, bobClient.GetFile(dataCommit, "/file", buf))
	require.Equal(t, "1", buf.String())
	// bob can write
	err = bobClient.PutFile(dataCommit, "/file", strings.NewReader("2"), client.WithAppendPutFile())
	require.NoError(t, err)
	require.Equal(t, 2, tu.CommitCnt(t, aliceClient, repo)) // check that a new commit was created
	commit, err = bobClient.StartCommit(pfs.DefaultProjectName, repoName, "master")
	require.NoError(t, err)
	require.NoError(t, bobClient.FinishCommit(pfs.DefaultProjectName, repoName, commit.Branch.Name, commit.Id))
	require.Equal(t, 3, tu.CommitCnt(t, aliceClient, repo)) // check that a new commit was created
	// bob can update the ACL
	require.NoError(t, bobClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, repoName, tu.Robot("carol"), []string{auth.RepoReaderRole}))
	// check that ACL was updated)
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, bob, auth.RepoOwnerRole, tu.Robot("carol"), auth.RepoReaderRole),
		tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, repoName))

	// clear carol
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, repoName, tu.Robot("carol"), []string{}))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, bob, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, repoName))

	//////////
	/// alice adds bob to the ACL as a writer
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, repoName, bob, []string{auth.RepoWriterRole}))
	// bob can read
	buf.Reset()
	require.NoError(t, bobClient.GetFile(dataCommit, "/file", buf))
	require.Equal(t, "12", buf.String())
	// bob can write
	err = bobClient.PutFile(dataCommit, "/file", strings.NewReader("3"), client.WithAppendPutFile())
	require.NoError(t, err)
	require.Equal(t, 4, tu.CommitCnt(t, aliceClient, repo)) // check that a new commit was created
	commit, err = bobClient.StartCommit(pfs.DefaultProjectName, repoName, "master")
	require.NoError(t, err)
	require.NoError(t, bobClient.FinishCommit(pfs.DefaultProjectName, repoName, commit.Branch.Name, commit.Id))
	require.Equal(t, 5, tu.CommitCnt(t, aliceClient, repo)) // check that a new commit was created
	// bob can't update the ACL
	err = bobClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, repoName, tu.Robot("carol"), []string{auth.RepoReaderRole})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	// check that ACL wasn't updated)
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, bob, auth.RepoWriterRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, repoName))

	//////////
	/// alice adds bob to the ACL as a reader (alice can modify ACL)
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, repoName, bob, []string{auth.RepoReaderRole}))
	// bob can read
	buf.Reset()
	require.NoError(t, bobClient.GetFile(dataCommit, "/file", buf))
	require.Equal(t, "123", buf.String())
	// bob can't write
	err = bobClient.PutFile(dataCommit, "/file", strings.NewReader("4"), client.WithAppendPutFile())
	require.YesError(t, err)
	require.Equal(t, 5, tu.CommitCnt(t, aliceClient, repo)) // check that a new commit was created
	_, err = bobClient.StartCommit(pfs.DefaultProjectName, repoName, "master")
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.Equal(t, 5, tu.CommitCnt(t, aliceClient, repo)) // check that no commits were created
	// bob can't update the ACL
	err = bobClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, repoName, tu.Robot("carol"), []string{auth.RepoReaderRole})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	// check that ACL wasn't updated)
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, bob, auth.RepoReaderRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, repoName))

	//////////
	/// alice revokes all of bob's privileges
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, repoName, bob, []string{}))
	// bob can't read
	err = bobClient.GetFile(dataCommit, "/file", buf)
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	// bob can't write
	err = bobClient.PutFile(dataCommit, "/file", strings.NewReader("4"), client.WithAppendPutFile())
	require.YesError(t, err)
	require.Equal(t, 5, tu.CommitCnt(t, aliceClient, repo)) // check that a new commit was created
	_, err = bobClient.StartCommit(pfs.DefaultProjectName, repoName, "master")
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.Equal(t, 5, tu.CommitCnt(t, aliceClient, repo)) // check that no commits were created
	// bob can't update the ACL
	err = bobClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, repoName, tu.Robot("carol"), []string{auth.RepoReaderRole})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	// check that ACL wasn't updated)
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, repoName))
}

// TestCreateAndUpdateRepo tests that if CreateRepo(foo, update=true) is
// called, and foo exists, then the ACL for foo won't be modified.
func TestCreateAndUpdateRepo(t *testing.T) {
	t.Parallel()
	env := envWithAuth(t)
	c := env.PachClient
	alice, bob := tu.Robot(tu.UniqueString("alice")), tu.Robot(tu.UniqueString("bob"))
	aliceClient, bobClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, bob)

	// create repo, and check that alice is the owner of the new repo
	dataRepo := tu.UniqueString(t.Name())
	require.NoError(t, aliceClient.CreateRepo(pfs.DefaultProjectName, dataRepo))
	require.Equal(t, tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, dataRepo))
	dataCommit := client.NewCommit(pfs.DefaultProjectName, dataRepo, "master", "")

	// Add data to repo (alice can write). Make sure alice can read also.
	err := aliceClient.PutFile(dataCommit, "/file", strings.NewReader("1"))
	require.NoError(t, err)
	buf := &bytes.Buffer{}
	require.NoError(t, aliceClient.GetFile(dataCommit, "/file", buf))
	require.Equal(t, "1", buf.String())

	/// alice adds bob to the ACL as a reader (alice can modify ACL)
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, dataRepo, bob, []string{auth.RepoReaderRole}))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, bob, auth.RepoReaderRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, dataRepo))
	// bob can read
	buf.Reset()
	require.NoError(t, bobClient.GetFile(dataCommit, "/file", buf))
	require.Equal(t, "1", buf.String())

	/// alice updates the repo
	description := "This request updates the description to force a write"
	_, err = aliceClient.PfsAPIClient.CreateRepo(aliceClient.Ctx(), &pfs.CreateRepoRequest{
		Repo:        client.NewRepo(pfs.DefaultProjectName, dataRepo),
		Description: description,
		Update:      true,
	})
	require.NoError(t, err)
	repoInfo, err := aliceClient.InspectRepo(pfs.DefaultProjectName, dataRepo)
	require.NoError(t, err)
	require.Equal(t, description, repoInfo.Description)
	// buildBindings haven't changed
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, bob, auth.RepoReaderRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, dataRepo))
	// bob can still read
	buf.Reset()
	require.NoError(t, bobClient.GetFile(dataCommit, "/file", buf))
	require.Equal(t, "1", buf.String())
}

// TestCreateRepoWithUpdateFlag tests that if CreateRepo(foo, update=true) is
// called, and foo doesn't exist, then the ACL for foo will still be created and
// initialized to the correct value
func TestCreateRepoWithUpdateFlag(t *testing.T) {
	t.Parallel()
	env := envWithAuth(t)
	c := env.PachClient
	alice := tu.Robot(tu.UniqueString("alice"))
	aliceClient := tu.AuthenticateClient(t, c, alice)

	// create repo, and check that alice is the owner of the new repo
	dataRepo := tu.UniqueString(t.Name())
	/// alice creates the repo with Update set
	_, err := aliceClient.PfsAPIClient.CreateRepo(aliceClient.Ctx(), &pfs.CreateRepoRequest{
		Repo:   client.NewRepo(pfs.DefaultProjectName, dataRepo),
		Update: true,
	})
	require.NoError(t, err)
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, dataRepo))
	dataCommit := client.NewCommit(pfs.DefaultProjectName, dataRepo, "master", "")
	// Add data to repo (alice can write). Make sure alice can read also.
	err = aliceClient.PutFile(dataCommit, "/file", strings.NewReader("1"))
	require.NoError(t, err)
	buf := &bytes.Buffer{}
	require.NoError(t, aliceClient.GetFile(dataCommit, "/file", buf))
	require.Equal(t, "1", buf.String())
}

func TestCreateAndUpdatePipeline(t *testing.T) {
	t.Parallel()
	env := envWithAuth(t)
	c := env.PachClient
	type createArgs struct {
		client     *client.APIClient
		name, repo string
		update     bool
	}
	createPipeline := func(args createArgs) error {
		return args.client.CreatePipeline(pfs.DefaultProjectName,
			args.name,
			"", // default image: DefaultUserImage
			[]string{"bash"},
			[]string{"cp /pfs/*/* /pfs/out/"},
			&pps.ParallelismSpec{Constant: 1},
			client.NewPFSInput(pfs.DefaultProjectName, args.repo, "/*"),
			"", // default output branch: master
			args.update,
		)
	}
	alice, bob := tu.Robot(tu.UniqueString("alice")), tu.Robot(tu.UniqueString("bob"))
	aliceClient, bobClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, bob)

	// create repo, and check that alice is the owner of the new repo
	dataRepo := tu.UniqueString(t.Name())
	require.NoError(t, aliceClient.CreateRepo(pfs.DefaultProjectName, dataRepo))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, dataRepo))
	dataCommit := client.NewCommit(pfs.DefaultProjectName, dataRepo, "master", "")

	// alice can create a pipeline (she owns the input repo)
	pipeline := tu.UniqueString("alice-pipeline")
	require.NoError(t, createPipeline(createArgs{
		client: aliceClient,
		name:   pipeline,
		repo:   dataRepo,
	}))
	require.OneOfEquals(t, pipeline, tu.PipelineNames(t, aliceClient, pfs.DefaultProjectName))
	// check that alice owns the output repo too)
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Pl(pfs.DefaultProjectName, pipeline), auth.RepoWriterRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, pipeline))

	// Make sure alice's pipeline runs successfully
	err := aliceClient.PutFile(dataCommit, tu.UniqueString("/file"),
		strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := aliceClient.WaitCommit(pfs.DefaultProjectName, pipeline, "master", "")
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
	require.NoneEquals(t, badPipeline, tu.PipelineNames(t, aliceClient, pfs.DefaultProjectName))

	// alice adds bob as a reader of the input repo
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, dataRepo, bob, []string{auth.RepoReaderRole}))

	// now bob can create a pipeline
	goodPipeline := tu.UniqueString("bob-good")
	require.NoError(t, createPipeline(createArgs{
		client: bobClient,
		name:   goodPipeline,
		repo:   dataRepo,
	}))
	require.OneOfEquals(t, goodPipeline, tu.PipelineNames(t, aliceClient, pfs.DefaultProjectName))
	// check that bob owns the output repo too)
	require.Equal(t,
		tu.BuildBindings(bob, auth.RepoOwnerRole, tu.Pl(pfs.DefaultProjectName, goodPipeline), auth.RepoWriterRole), tu.GetRepoRoleBinding(t, bobClient, pfs.DefaultProjectName, goodPipeline))

	// Make sure bob's pipeline runs successfully
	err = aliceClient.PutFile(dataCommit, tu.UniqueString("/file"),
		strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoErrorWithinT(t, 4*time.Minute, func() error {
		_, err := bobClient.WaitCommit(pfs.DefaultProjectName, goodPipeline, "master", "")
		return err
	})

	// bob can't update alice's pipeline
	infoBefore, err := aliceClient.InspectPipeline(pfs.DefaultProjectName, pipeline, true)
	require.NoError(t, err)
	err = createPipeline(createArgs{
		client: bobClient,
		name:   pipeline,
		repo:   dataRepo,
		update: true,
	})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	infoAfter, err := aliceClient.InspectPipeline(pfs.DefaultProjectName, pipeline, true)
	require.NoError(t, err)
	require.Equal(t, infoBefore.Version, infoAfter.Version)

	// alice adds bob as a writer of the output repo, and removes him as a reader
	// of the input repo
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, pipeline, bob, []string{auth.RepoWriterRole}))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, bob, auth.RepoWriterRole, tu.Pl(pfs.DefaultProjectName, pipeline), auth.RepoWriterRole),
		tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, pipeline))

	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, dataRepo, bob, []string{}))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Pl(pfs.DefaultProjectName, pipeline), auth.RepoReaderRole, tu.Pl(pfs.DefaultProjectName, goodPipeline), auth.RepoReaderRole),
		tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, dataRepo))

	// bob still can't update alice's pipeline
	infoBefore, err = aliceClient.InspectPipeline(pfs.DefaultProjectName, pipeline, true)
	require.NoError(t, err)
	err = createPipeline(createArgs{
		client: bobClient,
		name:   pipeline,
		repo:   dataRepo,
		update: true,
	})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	infoAfter, err = aliceClient.InspectPipeline(pfs.DefaultProjectName, pipeline, true)
	require.NoError(t, err)
	require.Equal(t, infoBefore.Version, infoAfter.Version)

	// alice re-adds bob as a reader of the input repo
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, dataRepo, bob, []string{auth.RepoReaderRole}))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, bob, auth.RepoReaderRole, tu.Pl(pfs.DefaultProjectName, pipeline), auth.RepoReaderRole, tu.Pl(pfs.DefaultProjectName, goodPipeline), auth.RepoReaderRole),
		tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, dataRepo))

	// now bob can update alice's pipeline
	infoBefore, err = aliceClient.InspectPipeline(pfs.DefaultProjectName, pipeline, true)
	require.NoError(t, err)
	err = createPipeline(createArgs{
		client: bobClient,
		name:   pipeline,
		repo:   dataRepo,
		update: true,
	})
	require.NoError(t, err)
	infoAfter, err = aliceClient.InspectPipeline(pfs.DefaultProjectName, pipeline, true)
	require.NoError(t, err)
	require.NotEqual(t, infoBefore.Version, infoAfter.Version)

	// Make sure that we don't get an auth token returned by the inspect
	require.Equal(t, "", infoAfter.AuthToken)
	infoAfter, err = aliceClient.InspectPipeline(pfs.DefaultProjectName, pipeline, false)
	require.NoError(t, err)
	require.Equal(t, "", infoAfter.AuthToken)

	// ListPipeline without details should list all repos
	pipelineInfos, err := aliceClient.ListPipeline(false)
	require.NoError(t, err)
	require.Equal(t, 2, len(pipelineInfos))
	for _, pipelineInfo := range pipelineInfos {
		require.Equal(t, "", pipelineInfo.AuthToken)
	}

	// Users can access a spec commit even if they can't list the repo itself,
	// so the details should be populated for every repo
	pipelineInfos, err = aliceClient.ListPipeline(true)
	require.NoError(t, err)
	require.Equal(t, 2, len(pipelineInfos))
	for _, pipelineInfo := range pipelineInfos {
		require.Equal(t, "", pipelineInfo.AuthToken)
		require.NotNil(t, pipelineInfo.Details)
	}

	pipelineInfos, err = bobClient.ListPipeline(true)
	require.NoError(t, err)
	require.Equal(t, 2, len(pipelineInfos))
	for _, pipelineInfo := range pipelineInfos {
		require.Equal(t, "", pipelineInfo.AuthToken)
		require.NotNil(t, pipelineInfo.Details)
	}

	// Make sure the updated pipeline runs successfully
	err = aliceClient.PutFile(dataCommit, tu.UniqueString("/file"),
		strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := bobClient.WaitCommit(pfs.DefaultProjectName, pipeline, "master", "")
		return err
	})
}

func TestPipelineMultipleInputs(t *testing.T) {
	t.Parallel()
	env := envWithAuth(t)
	c := env.PachClient
	type createArgs struct {
		client *client.APIClient
		name   string
		input  *pps.Input
		update bool
	}
	createPipeline := func(args createArgs) error {
		return args.client.CreatePipeline(pfs.DefaultProjectName,
			args.name,
			"", // default image: DefaultUserImage
			[]string{"bash"},
			[]string{"echo \"work\" >/pfs/out/x"},
			&pps.ParallelismSpec{Constant: 1},
			args.input,
			"", // default output branch: master
			args.update,
		)
	}
	alice, bob := tu.Robot(tu.UniqueString("alice")), tu.Robot(tu.UniqueString("bob"))
	aliceClient, bobClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, bob)

	// create two repos, and check that alice is the owner of the new repos
	dataRepo1 := tu.UniqueString(t.Name())
	dataRepo2 := tu.UniqueString(t.Name())
	require.NoError(t, aliceClient.CreateRepo(pfs.DefaultProjectName, dataRepo1))
	require.NoError(t, aliceClient.CreateRepo(pfs.DefaultProjectName, dataRepo2))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, dataRepo1))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, dataRepo2))

	// alice can create a cross-pipeline with both inputs
	aliceCrossPipeline := tu.UniqueString("alice-cross")
	require.NoError(t, createPipeline(createArgs{
		client: aliceClient,
		name:   aliceCrossPipeline,
		input: client.NewCrossInput(
			client.NewPFSInput(pfs.DefaultProjectName, dataRepo1, "/*"),
			client.NewPFSInput(pfs.DefaultProjectName, dataRepo2, "/*"),
		),
	}))
	require.OneOfEquals(t, aliceCrossPipeline, tu.PipelineNames(t, aliceClient, pfs.DefaultProjectName))
	// check that alice owns the output repo too)
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Pl(pfs.DefaultProjectName, aliceCrossPipeline), auth.RepoWriterRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, aliceCrossPipeline))

	// alice can create a union-pipeline with both inputs
	aliceUnionPipeline := tu.UniqueString("alice-union")
	require.NoError(t, createPipeline(createArgs{
		client: aliceClient,
		name:   aliceUnionPipeline,
		input: client.NewUnionInput(
			client.NewPFSInput(pfs.DefaultProjectName, dataRepo1, "/*"),
			client.NewPFSInput(pfs.DefaultProjectName, dataRepo2, "/*"),
		),
	}))
	require.OneOfEquals(t, aliceUnionPipeline, tu.PipelineNames(t, aliceClient, pfs.DefaultProjectName))
	// check that alice owns the output repo too)
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Pl(pfs.DefaultProjectName, aliceUnionPipeline), auth.RepoWriterRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, aliceUnionPipeline))

	// alice adds bob as a reader of one of the input repos, but not the other
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, dataRepo1, bob, []string{auth.RepoReaderRole}))

	// bob cannot create a cross-pipeline with both inputs
	bobCrossPipeline := tu.UniqueString("bob-cross")
	err := createPipeline(createArgs{
		client: bobClient,
		name:   bobCrossPipeline,
		input: client.NewCrossInput(
			client.NewPFSInput(pfs.DefaultProjectName, dataRepo1, "/*"),
			client.NewPFSInput(pfs.DefaultProjectName, dataRepo2, "/*"),
		),
	})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.NoneEquals(t, bobCrossPipeline, tu.PipelineNames(t, aliceClient, pfs.DefaultProjectName))

	// bob cannot create a union-pipeline with both inputs
	bobUnionPipeline := tu.UniqueString("bob-union")
	err = createPipeline(createArgs{
		client: bobClient,
		name:   bobUnionPipeline,
		input: client.NewUnionInput(
			client.NewPFSInput(pfs.DefaultProjectName, dataRepo1, "/*"),
			client.NewPFSInput(pfs.DefaultProjectName, dataRepo2, "/*"),
		),
	})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.NoneEquals(t, bobUnionPipeline, tu.PipelineNames(t, aliceClient, pfs.DefaultProjectName))

	// alice adds bob as a writer of her pipeline's output
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, aliceCrossPipeline, bob, []string{auth.RepoWriterRole}))

	// bob can update alice's pipeline if he removes one of the inputs
	infoBefore, err := aliceClient.InspectPipeline(pfs.DefaultProjectName, aliceCrossPipeline, true)
	require.NoError(t, err)
	require.NoError(t, createPipeline(createArgs{
		client: bobClient,
		name:   aliceCrossPipeline,
		input: client.NewCrossInput(
			// This cross input deliberately only has one element, to make sure it's
			// not simply rejected for having a cross input
			client.NewPFSInput(pfs.DefaultProjectName, dataRepo1, "/*"),
		),
		update: true,
	}))
	infoAfter, err := aliceClient.InspectPipeline(pfs.DefaultProjectName, aliceCrossPipeline, true)
	require.NoError(t, err)
	require.NotEqual(t, infoBefore.Version, infoAfter.Version)

	// bob cannot update alice's to put the second input back
	infoBefore, err = aliceClient.InspectPipeline(pfs.DefaultProjectName, aliceCrossPipeline, true)
	require.NoError(t, err)
	err = createPipeline(createArgs{
		client: bobClient,
		name:   aliceCrossPipeline,
		input: client.NewCrossInput(
			client.NewPFSInput(pfs.DefaultProjectName, dataRepo1, "/*"),
			client.NewPFSInput(pfs.DefaultProjectName, dataRepo2, "/*"),
		),
		update: true,
	})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	infoAfter, err = aliceClient.InspectPipeline(pfs.DefaultProjectName, aliceCrossPipeline, true)
	require.NoError(t, err)
	require.Equal(t, infoBefore.Version, infoAfter.Version)

	// alice adds bob as a reader of the second input
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, dataRepo2, bob, []string{auth.RepoReaderRole}))

	// bob can now update alice's to put the second input back
	infoBefore, err = aliceClient.InspectPipeline(pfs.DefaultProjectName, aliceCrossPipeline, true)
	require.NoError(t, err)
	require.NoError(t, createPipeline(createArgs{
		client: bobClient,
		name:   aliceCrossPipeline,
		input: client.NewCrossInput(
			client.NewPFSInput(pfs.DefaultProjectName, dataRepo1, "/*"),
			client.NewPFSInput(pfs.DefaultProjectName, dataRepo2, "/*"),
		),
		update: true,
	}))
	infoAfter, err = aliceClient.InspectPipeline(pfs.DefaultProjectName, aliceCrossPipeline, true)
	require.NoError(t, err)
	require.NotEqual(t, infoBefore.Version, infoAfter.Version)

	// bob can create a cross-pipeline with both inputs
	require.NoError(t, createPipeline(createArgs{
		client: bobClient,
		name:   bobCrossPipeline,
		input: client.NewCrossInput(
			client.NewPFSInput(pfs.DefaultProjectName, dataRepo1, "/*"),
			client.NewPFSInput(pfs.DefaultProjectName, dataRepo2, "/*"),
		),
	}))
	require.OneOfEquals(t, bobCrossPipeline, tu.PipelineNames(t, aliceClient, pfs.DefaultProjectName))

	// bob can create a union-pipeline with both inputs
	require.NoError(t, createPipeline(createArgs{
		client: bobClient,
		name:   bobUnionPipeline,
		input: client.NewUnionInput(
			client.NewPFSInput(pfs.DefaultProjectName, dataRepo1, "/*"),
			client.NewPFSInput(pfs.DefaultProjectName, dataRepo2, "/*"),
		),
	}))
	require.OneOfEquals(t, bobUnionPipeline, tu.PipelineNames(t, aliceClient, pfs.DefaultProjectName))

}

func TestStopAndDeletePipeline(t *testing.T) {
	t.Parallel()
	env := envWithAuth(t)
	c := env.PachClient
	alice, bob := tu.Robot(tu.UniqueString("alice")), tu.Robot(tu.UniqueString("bob"))
	aliceClient, bobClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, bob)

	// alice creates a repo
	repo := tu.UniqueString(t.Name())
	require.NoError(t, aliceClient.CreateRepo(pfs.DefaultProjectName, repo))
	require.Equal(t, tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, repo))

	// alice creates a pipeline
	pipeline := tu.UniqueString("alice-pipeline")
	require.NoError(t, aliceClient.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"", // default image: DefaultUserImage
		[]string{"bash"},
		[]string{"cp /pfs/*/* /pfs/out/"},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(pfs.DefaultProjectName, repo, "/*"),
		"", // default output branch: master
		false,
	))
	// Make sure the input and output repos have non-empty ACLs
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Pl(pfs.DefaultProjectName, pipeline), auth.RepoReaderRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, repo))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Pl(pfs.DefaultProjectName, pipeline), auth.RepoWriterRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, pipeline))

	// alice stops the pipeline (owner of the input and output repos can stop)
	require.NoError(t, aliceClient.StopPipeline(pfs.DefaultProjectName, pipeline))

	// Make sure the remaining input and output repos *still* have non-empty ACLs
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Pl(pfs.DefaultProjectName, pipeline), auth.RepoReaderRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, repo))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Pl(pfs.DefaultProjectName, pipeline), auth.RepoWriterRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, pipeline))

	// alice deletes the pipeline (owner of the input and output repos can delete)
	require.NoError(t, aliceClient.DeletePipeline(pfs.DefaultProjectName, pipeline, false))
	require.Nil(t, tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, pipeline).Entries)

	// alice deletes the input repo (make sure the input repo's ACL is gone)
	require.NoError(t, aliceClient.DeleteRepo(pfs.DefaultProjectName, repo, false))
	require.Nil(t, tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, repo).Entries)

	// alice creates another repo
	repo = tu.UniqueString(t.Name())
	require.NoError(t, aliceClient.CreateRepo(pfs.DefaultProjectName, repo))
	require.Equal(t, tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, repo))

	// alice creates another pipeline
	pipeline = tu.UniqueString("alice-pipeline")
	require.NoError(t, aliceClient.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"", // default image: DefaultUserImage
		[]string{"bash"},
		[]string{"cp /pfs/*/* /pfs/out/"},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(pfs.DefaultProjectName, repo, "/*"),
		"", // default output branch: master
		false,
	))

	// bob can't stop or delete alice's pipeline
	err := bobClient.StopPipeline(pfs.DefaultProjectName, pipeline)
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	err = bobClient.DeletePipeline(pfs.DefaultProjectName, pipeline, false)
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())

	// alice adds bob as a reader of the input repo
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, repo, bob, []string{auth.RepoReaderRole}))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, bob, auth.RepoReaderRole, tu.Pl(pfs.DefaultProjectName, pipeline), auth.RepoReaderRole),
		tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, repo))

	// bob still can't stop or delete alice's pipeline
	err = bobClient.StopPipeline(pfs.DefaultProjectName, pipeline)
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	err = bobClient.DeletePipeline(pfs.DefaultProjectName, pipeline, false)
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())

	// alice removes bob as a reader of the input repo and adds bob as a writer of
	// the output repo
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, repo, bob, []string{}))

	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Pl(pfs.DefaultProjectName, pipeline), auth.RepoReaderRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, repo))
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, pipeline, bob, []string{auth.RepoWriterRole}))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, bob, auth.RepoWriterRole, tu.Pl(pfs.DefaultProjectName, pipeline), auth.RepoWriterRole),
		tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, pipeline))

	// bob can now start and stop the pipeline, but can't delete it
	require.NoError(t, bobClient.StopPipeline(pfs.DefaultProjectName, pipeline))
	require.NoError(t, bobClient.StartPipeline(pfs.DefaultProjectName, pipeline))
	err = bobClient.DeletePipeline(pfs.DefaultProjectName, pipeline, false)
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	// alice re-adds bob as a reader of the input repo
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, repo, bob, []string{auth.RepoReaderRole}))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, bob, auth.RepoReaderRole, tu.Pl(pfs.DefaultProjectName, pipeline), auth.RepoReaderRole),
		tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, repo))

	// no change to bob's capabilities
	require.NoError(t, bobClient.StopPipeline(pfs.DefaultProjectName, pipeline))
	require.NoError(t, bobClient.StartPipeline(pfs.DefaultProjectName, pipeline))
	err = bobClient.DeletePipeline(pfs.DefaultProjectName, pipeline, false)
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())

	// alice adds bob as an owner of the output repo
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, pipeline, bob, []string{auth.RepoOwnerRole}))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, bob, auth.RepoOwnerRole, tu.Pl(pfs.DefaultProjectName, pipeline), auth.RepoWriterRole),
		tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, pipeline))

	// finally bob can delete alice's pipeline
	err = bobClient.DeletePipeline(pfs.DefaultProjectName, pipeline, false)
	require.NoError(t, err)
}

// TestStopJob just confirms that the StopJob API works when auth is on
func TestStopJob(t *testing.T) {
	t.Parallel()
	env := envWithAuth(t)
	c := env.PachClient
	alice := tu.Robot(tu.UniqueString("alice"))
	aliceClient := tu.AuthenticateClient(t, c, alice)

	// alice creates a repo
	repo := tu.UniqueString(t.Name())
	commit := client.NewCommit(pfs.DefaultProjectName, repo, "master", "")
	require.NoError(t, aliceClient.CreateRepo(pfs.DefaultProjectName, repo))
	require.Equal(t, tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, repo))
	err := aliceClient.PutFile(commit, "/file", strings.NewReader("test"))
	require.NoError(t, err)

	// alice creates a pipeline
	pipeline := tu.UniqueString("alice-pipeline")
	require.NoError(t, aliceClient.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"", // default image: DefaultUserImage
		[]string{"bash"},
		[]string{"sleep 600"},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(pfs.DefaultProjectName, repo, "/*"),
		"", // default output branch: master
		false,
	))
	// Make sure the input and output repos have non-empty ACLs
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Pl(pfs.DefaultProjectName, pipeline), auth.RepoReaderRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, repo))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Pl(pfs.DefaultProjectName, pipeline), auth.RepoWriterRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, pipeline))

	// Stop the first job in 'pipeline'
	var jobID string
	require.NoErrorWithinTRetry(t, 30*time.Second, func() error {
		jobs, err := aliceClient.ListJob(pfs.DefaultProjectName, pipeline, nil /*inputs*/, -1 /*history*/, true /* full */)
		if err != nil {
			return err
		}
		if len(jobs) != 1 {
			return errors.Errorf("expected one job but got %d", len(jobs))
		}
		jobID = jobs[0].Job.Id
		return nil
	})

	require.NoError(t, aliceClient.StopJob(pfs.DefaultProjectName, pipeline, jobID))
	require.NoErrorWithinTRetry(t, 30*time.Second, func() error {
		ji, err := aliceClient.InspectJob(pfs.DefaultProjectName, pipeline, jobID, false)
		if err != nil {
			return errors.Wrapf(err, "could not inspect job %q", jobID)
		}
		if ji.State != pps.JobState_JOB_KILLED {
			return errors.Errorf("expected job %q to be in JOB_KILLED but was in %s", jobID, ji.State.String())
		}
		return nil
	})
}

// Test ListRepo checks that the auth information returned by ListRepo and
// InspectRepo is correct.
// TODO(msteffen): This should maybe go in pachyderm_test, since ListRepo isn't
// an auth API call
func TestListAndInspectRepo(t *testing.T) {
	t.Parallel()
	env := envWithAuth(t)
	c := env.PachClient
	alice, bob := tu.Robot(tu.UniqueString("alice")), tu.Robot(tu.UniqueString("bob"))
	aliceClient, bobClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, bob)

	// alice creates a repo and makes Bob a writer
	repoWriter := tu.UniqueString("repoWriter")
	require.NoError(t, aliceClient.CreateRepo(pfs.DefaultProjectName, repoWriter))
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, repoWriter, bob, []string{auth.RepoWriterRole}))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, bob, auth.RepoWriterRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, repoWriter))

	// alice creates a repo and makes Bob a reader
	repoReader := tu.UniqueString("repoReader")
	require.NoError(t, aliceClient.CreateRepo(pfs.DefaultProjectName, repoReader))
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, repoReader, bob, []string{auth.RepoReaderRole}))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, bob, auth.RepoReaderRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, repoReader))

	// alice creates a repo and gives Bob no access privileges
	repoNone := tu.UniqueString("repoNone")
	require.NoError(t, aliceClient.CreateRepo(pfs.DefaultProjectName, repoNone))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, repoNone))

	// put a file in the repo Bob can't access - we need to be able to get the size of the commits
	err := aliceClient.PutFile(client.NewCommit(pfs.DefaultProjectName, repoNone, "master", ""), "/test", strings.NewReader("test"))
	require.NoError(t, err)

	// bob creates a repo, and becomes its owner
	repoOwner := tu.UniqueString("repoOwner")
	require.NoError(t, bobClient.CreateRepo(pfs.DefaultProjectName, repoOwner))
	require.Equal(t, tu.BuildBindings(bob, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, bobClient, pfs.DefaultProjectName, repoOwner))

	// Bob calls ListRepo, and the response must indicate the correct access scope
	// for each repo (because other tests have run, we may see repos besides the
	// above. Bob's access to those should be NONE
	lrClient, err := bobClient.PfsAPIClient.ListRepo(bobClient.Ctx(), &pfs.ListRepoRequest{})
	require.NoError(t, err)
	repoInfos, err := grpcutil.Collect[*pfs.RepoInfo](lrClient, 1000)
	require.NoError(t, err)
	expectedPermissions := map[string][]auth.Permission{
		repoOwner: {
			auth.Permission_PIPELINE_LIST_JOB,
			auth.Permission_REPO_ADD_PIPELINE_READER,
			auth.Permission_REPO_ADD_PIPELINE_WRITER,
			auth.Permission_REPO_CREATE_BRANCH,
			auth.Permission_REPO_DELETE_BRANCH,
			auth.Permission_REPO_DELETE_COMMIT,
			auth.Permission_REPO_DELETE,
			auth.Permission_REPO_INSPECT_COMMIT,
			auth.Permission_REPO_INSPECT_FILE,
			auth.Permission_REPO_LIST_BRANCH,
			auth.Permission_REPO_LIST_COMMIT,
			auth.Permission_REPO_LIST_FILE,
			auth.Permission_REPO_MODIFY_BINDINGS,
			auth.Permission_REPO_READ,
			auth.Permission_REPO_REMOVE_PIPELINE_READER,
			auth.Permission_REPO_WRITE,
		},
		repoWriter: {
			auth.Permission_PIPELINE_LIST_JOB,
			auth.Permission_REPO_ADD_PIPELINE_READER,
			auth.Permission_REPO_ADD_PIPELINE_WRITER,
			auth.Permission_REPO_CREATE_BRANCH,
			auth.Permission_REPO_DELETE_BRANCH,
			auth.Permission_REPO_DELETE_COMMIT,
			auth.Permission_REPO_INSPECT_COMMIT,
			auth.Permission_REPO_INSPECT_FILE,
			auth.Permission_REPO_LIST_BRANCH,
			auth.Permission_REPO_LIST_COMMIT,
			auth.Permission_REPO_LIST_FILE,
			auth.Permission_REPO_READ,
			auth.Permission_REPO_REMOVE_PIPELINE_READER,
			auth.Permission_REPO_WRITE,
		},
		repoReader: {
			auth.Permission_PIPELINE_LIST_JOB,
			auth.Permission_REPO_ADD_PIPELINE_READER,
			auth.Permission_REPO_INSPECT_COMMIT,
			auth.Permission_REPO_INSPECT_FILE,
			auth.Permission_REPO_LIST_BRANCH,
			auth.Permission_REPO_LIST_COMMIT,
			auth.Permission_REPO_LIST_FILE,
			auth.Permission_REPO_READ,
			auth.Permission_REPO_REMOVE_PIPELINE_READER,
		},
		repoNone: {},
	}
	for _, info := range repoInfos {
		require.ElementsEqual(t, expectedPermissions[info.Repo.Name], info.AuthInfo.Permissions)
	}

	for _, name := range []string{repoOwner, repoWriter, repoReader, repoNone} {
		inspectResp, err := bobClient.PfsAPIClient.InspectRepo(bobClient.Ctx(),
			&pfs.InspectRepoRequest{
				Repo: client.NewRepo(pfs.DefaultProjectName, name),
			})
		require.NoError(t, err)
		require.ElementsEqual(t, expectedPermissions[name], inspectResp.AuthInfo.Permissions)
	}
}

func TestUnprivilegedUserCannotMakeSelfOwner(t *testing.T) {
	t.Parallel()
	env := envWithAuth(t)
	c := env.PachClient
	alice, bob := tu.Robot(tu.UniqueString("alice")), tu.Robot(tu.UniqueString("bob"))
	aliceClient, bobClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, bob)

	// alice creates a repo
	repo := tu.UniqueString(t.Name())
	require.NoError(t, aliceClient.CreateRepo(pfs.DefaultProjectName, repo))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, repo))

	// bob calls SetScope(bob, OWNER) on alice's repo. This should fail
	err := bobClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, repo, bob, []string{auth.RepoOwnerRole})
	require.YesError(t, err)
	// make sure ACL wasn't updated
	require.Equal(t, tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, repo))
}

// TestListRepoNotLoggedInError makes sure that if a user isn't logged in, and
// they call ListRepo(), they get an error.
func TestListRepoNotLoggedInError(t *testing.T) {
	t.Parallel()
	env := envWithAuth(t)
	client := env.PachClient
	alice := tu.Robot(tu.UniqueString("alice"))
	aliceClient, anonClient := tu.AuthenticateClient(t, client, alice), tu.UnauthenticatedPachClient(t, client)

	// alice creates a repo
	repo := tu.UniqueString(t.Name())
	require.NoError(t, aliceClient.CreateRepo(pfs.DefaultProjectName, repo))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, repo))

	// Anon (non-logged-in user) calls ListRepo, and must receive an error
	c, err := anonClient.PfsAPIClient.ListRepo(anonClient.Ctx(),
		&pfs.ListRepoRequest{})
	require.NoError(t, err)
	_, err = grpcutil.Collect[*pfs.RepoInfo](c, 1000)
	require.YesError(t, err)
	require.Matches(t, "no authentication token", err.Error())
}

// TestListRepoNoAuthInfoIfDeactivated tests that if auth isn't activated, then
// ListRepo returns RepoInfos where AuthInfo isn't set (i.e. is nil)
func TestListRepoNoAuthInfoIfDeactivated(t *testing.T) {
	env := envWithAuth(t)
	c := env.PachClient
	// Dont't run this test in parallel, since it deactivates the auth system
	// globally, so any tests running concurrently will fail
	alice, bob := tu.Robot(tu.UniqueString("alice")), tu.Robot(tu.UniqueString("bob"))
	aliceClient, bobClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, bob)
	adminClient := tu.AuthenticateClient(t, c, auth.RootUser)

	// alice creates a repo
	repo := tu.UniqueString(t.Name())
	require.NoError(t, aliceClient.CreateRepo(pfs.DefaultProjectName, repo))

	infos, err := bobClient.ListRepo()
	require.NoError(t, err)
	for _, info := range infos {
		require.ElementsEqual(t, []string{}, info.AuthInfo.Roles)
	}

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

	// bob calls ListRepo, now AuthInfo isn't set anywhere
	infos, err = bobClient.ListRepo()
	require.NoError(t, err)
	for _, info := range infos {
		require.Nil(t, info.AuthInfo)
	}
}

// TestCreateRepoAlreadyExistsError tests that creating a repo that already
// exists gives you an error to that effect, even when auth is already
// activated (rather than "access denied")
func TestCreateRepoAlreadyExistsError(t *testing.T) {
	t.Parallel()
	env := envWithAuth(t)
	c := env.PachClient
	alice, bob := tu.Robot(tu.UniqueString("alice")), tu.Robot(tu.UniqueString("bob"))
	aliceClient, bobClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, bob)

	// alice creates a repo
	repo := tu.UniqueString(t.Name())
	require.NoError(t, aliceClient.CreateRepo(pfs.DefaultProjectName, repo))

	// bob creates the same repo, and should get an error to the effect that the
	// repo already exists (rather than "access denied")
	err := bobClient.CreateRepo(pfs.DefaultProjectName, repo)
	require.YesError(t, err)
	require.Matches(t, "already exists", err.Error())
}

// TestCreateRepoNotLoggedInError makes sure that if a user isn't logged in, and
// they call CreateRepo(), they get an error.
func TestCreateRepoNotLoggedInError(t *testing.T) {
	t.Parallel()
	env := envWithAuth(t)
	c := env.PachClient
	anonClient := tu.UnauthenticatedPachClient(t, c)

	// anonClient tries and fails to create a repo
	repo := tu.UniqueString(t.Name())
	err := anonClient.CreateRepo(pfs.DefaultProjectName, repo)
	require.YesError(t, err)
	require.Matches(t, "no authentication token", err.Error())
}

// TestProjectWriter tests the access control related to the ProjectWriter role.
func TestProjectWriter(t *testing.T) {
	t.Parallel()
	env := envWithAuth(t)
	c := env.PachClient

	admin := tu.AuthenticateClient(t, c, auth.RootUser)
	aliceName, alice := tu.RandomRobot(t, c, "alice")

	project, repo := tu.UniqueString("project"), tu.UniqueString("repo")

	// Without ProjectWriter's PROJECT_CREATE_REPO permission, Alice cannot create a repo in project
	require.NoError(t, admin.CreateProject(project))
	require.ErrorContains(t, alice.CreateRepo(project, repo), "not authorized")

	// Admin grants Alice the ProjectWriter role, which allows Alice to create a repo in the project.
	_, err := admin.ModifyRoleBinding(admin.Ctx(), &auth.ModifyRoleBindingRequest{
		Principal: aliceName,
		Roles:     []string{auth.ProjectWriterRole},
		Resource:  &auth.Resource{Type: auth.ResourceType_PROJECT, Name: project},
	})
	require.NoError(t, err)
	require.NoError(t, alice.CreateRepo(project, repo))

	// Pipeline creation depends on Repo creation. Bob cannot create repos in the project, but Alice can.
	_, bob := tu.RandomRobot(t, c, "bob")
	require.ErrorContains(t, bob.CreatePipeline(project, "pipeline", "", []string{"cp", "/pfs/in/*", "/pfs/out"}, nil, nil, &pps.Input{Pfs: &pps.PFSInput{Project: project, Repo: repo, Glob: "/*", Name: "in"}}, "", false), "not authorized")
	require.NoError(t, alice.CreatePipeline(project, "pipeline", "", []string{"cp", "/pfs/in/*", "/pfs/out"}, nil, nil, &pps.Input{Pfs: &pps.PFSInput{Project: project, Repo: repo, Glob: "/*", Name: "in"}}, "", false))
}

// Creating a pipeline when the output repo already exists gives is not allowed
func TestCreatePipelineRepoAlreadyExists(t *testing.T) {
	t.Parallel()
	env := envWithAuth(t)
	c := env.PachClient
	alice, bob := tu.Robot(tu.UniqueString("alice")), tu.Robot(tu.UniqueString("bob"))
	aliceClient, bobClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, bob)

	// alice creates a repo
	inputRepo := tu.UniqueString(t.Name())
	require.NoError(t, aliceClient.CreateRepo(pfs.DefaultProjectName, inputRepo))
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, inputRepo, bob, []string{auth.RepoReaderRole}))
	pipeline := tu.UniqueString("pipeline")
	require.NoError(t, aliceClient.CreateRepo(pfs.DefaultProjectName, pipeline))

	// bob creates a pipeline, and should get an "access denied" error
	err := bobClient.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"", // default image: DefaultUserImage
		[]string{"bash"},
		[]string{"cp /pfs/*/* /pfs/out/"},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(pfs.DefaultProjectName, inputRepo, "/*"),
		"",    // default output branch: master
		false, // Don't update -- we want an error
	)
	require.YesError(t, err)
	require.Matches(t, "already exists", err.Error())

	// alice gives bob writer scope on pipeline output repo, but nothing changes
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, pipeline, bob, []string{auth.RepoWriterRole}))
	err = bobClient.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"", // default image: DefaultUserImage
		[]string{"bash"},
		[]string{"cp /pfs/*/* /pfs/out/"},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(pfs.DefaultProjectName, inputRepo, "/*"),
		"",    // default output branch: master
		false, // Don't update -- we want an error
	)
	require.YesError(t, err)
	require.Matches(t, "already exists", err.Error())
}

// TestAuthorizedEveryone tests that Authorized(user, repo, NONE) tests that the
// `allClusterUsers` binding  for an ACL sets the minimum authorized scope
func TestAuthorizedEveryone(t *testing.T) {
	t.Parallel()
	env := envWithAuth(t)
	c := env.PachClient
	alice, bob := tu.Robot(tu.UniqueString("alice")), tu.Robot(tu.UniqueString("bob"))
	aliceClient, bobClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, bob)

	// alice creates a repo
	repo := tu.UniqueString(t.Name())
	require.NoError(t, aliceClient.CreateRepo(pfs.DefaultProjectName, repo))

	// alice is authorized as `OWNER`
	resp, err := aliceClient.Authorize(aliceClient.Ctx(), &auth.AuthorizeRequest{
		Resource: &auth.Resource{Type: auth.ResourceType_REPO, Name: pfs.DefaultProjectName + "/" + repo},
		Permissions: []auth.Permission{
			auth.Permission_REPO_MODIFY_BINDINGS,
			auth.Permission_REPO_WRITE,
			auth.Permission_REPO_READ,
		},
	})
	require.NoError(t, err)
	require.True(t, resp.Authorized)

	// bob is not authorized
	resp, err = bobClient.Authorize(bobClient.Ctx(), &auth.AuthorizeRequest{
		Resource: &auth.Resource{Type: auth.ResourceType_REPO, Name: pfs.DefaultProjectName + "/" + repo},
		Permissions: []auth.Permission{
			auth.Permission_REPO_READ,
		},
	})
	require.NoError(t, err)
	require.False(t, resp.Authorized)

	// alice grants everybody WRITER access
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, repo, auth.AllClusterUsersSubject, []string{auth.RepoWriterRole}))

	// alice is still authorized as `OWNER`
	resp, err = aliceClient.Authorize(aliceClient.Ctx(), &auth.AuthorizeRequest{
		Resource: &auth.Resource{Type: auth.ResourceType_REPO, Name: pfs.DefaultProjectName + "/" + repo},
		Permissions: []auth.Permission{
			auth.Permission_REPO_MODIFY_BINDINGS,
			auth.Permission_REPO_WRITE,
			auth.Permission_REPO_READ,
		},
	})
	require.NoError(t, err)
	require.True(t, resp.Authorized)

	// bob is now authorized as WRITER
	resp, err = bobClient.Authorize(bobClient.Ctx(), &auth.AuthorizeRequest{
		Resource: &auth.Resource{Type: auth.ResourceType_REPO, Name: pfs.DefaultProjectName + "/" + repo},
		Permissions: []auth.Permission{
			auth.Permission_REPO_WRITE,
			auth.Permission_REPO_READ,
		},
	})
	require.NoError(t, err)
	require.True(t, resp.Authorized)
}

// TestDeleteAllRepos tests that when you delete all repos,
// only the repos you are authorized to delete are deleted
func TestDeleteAllRepos(t *testing.T) {
	t.Parallel()
	env := envWithAuth(t)
	c := env.PachClient
	alice := tu.Robot(tu.UniqueString("alice"))
	aliceClient, adminClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, auth.RootUser)

	// admin creates a repo
	adminRepo := tu.UniqueString(t.Name())
	require.NoError(t, adminClient.CreateRepo(pfs.DefaultProjectName, adminRepo))

	// alice creates a repo
	aliceRepo := tu.UniqueString(t.Name())
	require.NoError(t, aliceClient.CreateRepo(pfs.DefaultProjectName, aliceRepo))

	// alice tries to delete all repos, but is not allowed to delete admin's repo
	_, err := aliceClient.PfsAPIClient.DeleteAll(aliceClient.Ctx(), &emptypb.Empty{})
	require.ErrorContains(t, err, "not authorized")

	// admin can delete all repos
	_, err = adminClient.PfsAPIClient.DeleteAll(adminClient.Ctx(), &emptypb.Empty{})
	require.NoError(t, err)

	listResp, err := aliceClient.ListRepo()
	require.NoError(t, err)
	require.Equal(t, 0, len(listResp))
}

// TestListJob tests that you must have READER access to a pipeline's output
// repo to call ListJob on that pipeline, but a blank ListJob always succeeds
// (but doesn't return a given job if you don't have access to the job's output
// repo)
func TestListJob(t *testing.T) {
	t.Parallel()
	env := envWithAuth(t)
	c := env.PachClient
	alice, bob := tu.Robot(tu.UniqueString("alice")), tu.Robot(tu.UniqueString("bob"))
	aliceClient, bobClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, bob)

	// alice creates a repo
	repo := tu.UniqueString(t.Name())
	require.NoError(t, aliceClient.CreateRepo(pfs.DefaultProjectName, repo))

	// alice creates a pipeline
	pipeline := tu.UniqueString("alice-pipeline")
	require.NoError(t, aliceClient.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"", // default image: DefaultUserImage
		[]string{"bash"},
		[]string{"ls /pfs/*/*; cp /pfs/*/* /pfs/out/"},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(pfs.DefaultProjectName, repo, "/*"),
		"", // default output branch: master
		false,
	))

	// alice commits to the input repos, and the pipeline runs successfully
	var err error
	err = aliceClient.PutFile(client.NewCommit(pfs.DefaultProjectName, repo, "master", ""), "/file", strings.NewReader("test"))
	require.NoError(t, err)
	require.NoErrorWithinT(t, 4*time.Minute, func() error {
		_, err := aliceClient.WaitCommit(pfs.DefaultProjectName, pipeline, "master", "")
		return err
	})
	jobs, err := aliceClient.ListJob(pfs.DefaultProjectName, pipeline, nil /*inputs*/, -1 /*history*/, true)
	require.NoError(t, err)
	require.Equal(t, 2, len(jobs))
	jobID := jobs[0].Job.Id

	// bob cannot call ListJob on 'pipeline'
	_, err = bobClient.ListJob(pfs.DefaultProjectName, pipeline, nil, -1 /*history*/, true)
	require.YesError(t, err)
	require.True(t, auth.IsErrNotAuthorized(err), err.Error())
	// bob can call blank ListJob, but gets no results
	jobs, err = bobClient.ListJob(pfs.DefaultProjectName, "", nil, -1 /*history*/, true)
	require.NoError(t, err)
	require.Equal(t, 0, len(jobs))

	// alice adds bob to repo, but bob still can't call ListJob on 'pipeline' or
	// get any output
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, repo, bob, []string{auth.RepoReaderRole}))
	_, err = bobClient.ListJob(pfs.DefaultProjectName, pipeline, nil, -1 /*history*/, true)
	require.YesError(t, err)
	require.True(t, auth.IsErrNotAuthorized(err), err.Error())
	jobs, err = bobClient.ListJob(pfs.DefaultProjectName, "", nil, -1 /*history*/, true)
	require.NoError(t, err)
	require.Equal(t, 0, len(jobs))

	// alice removes bob from repo and adds bob to 'pipeline', and now bob can
	// call listJob on 'pipeline', and gets results back from blank listJob
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, repo, bob, []string{}))
	err = aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, pipeline, bob, []string{auth.RepoReaderRole})
	require.NoError(t, err)
	jobs, err = bobClient.ListJob(pfs.DefaultProjectName, pipeline, nil, -1 /*history*/, true)
	require.NoError(t, err)
	require.Equal(t, 2, len(jobs))
	require.Equal(t, jobID, jobs[0].Job.Id)
	jobs, err = bobClient.ListJob(pfs.DefaultProjectName, "", nil, -1 /*history*/, true)
	require.NoError(t, err)
	require.Equal(t, 2, len(jobs))
	require.Equal(t, jobID, jobs[0].Job.Id)
}

// TestInspectDatum tests InspectDatum runs even when auth is activated
func TestInspectDatum(t *testing.T) {
	t.Parallel()
	env := envWithAuth(t)
	c := env.PachClient
	alice := tu.Robot(tu.UniqueString("alice"))
	aliceClient := tu.AuthenticateClient(t, c, alice)

	// alice creates a repo
	repo := tu.UniqueString(t.Name())
	require.NoError(t, aliceClient.CreateRepo(pfs.DefaultProjectName, repo))

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
			Input:           client.NewPFSInput(pfs.DefaultProjectName, repo, "/*"),
		})
	require.NoError(t, err)

	// alice commits to the input repo, and the pipeline runs successfully
	err = aliceClient.PutFile(client.NewCommit(pfs.DefaultProjectName, repo, "master", ""), "/file", strings.NewReader("test"))
	require.NoError(t, err)
	require.NoErrorWithinT(t, 2*time.Minute, func() error {
		_, err := aliceClient.WaitCommit(pfs.DefaultProjectName, pipeline, "master", "")
		return err
	})
	jobs, err := aliceClient.ListJob(pfs.DefaultProjectName, pipeline, nil /*inputs*/, -1 /*history*/, true)
	require.NoError(t, err)
	require.Equal(t, 2, len(jobs))
	jobID := jobs[0].Job.Id

	// ListDatum seems like it may return inconsistent results, so sleep until
	// the /stats branch is written
	// TODO(msteffen): verify if this is true, and if so, why
	time.Sleep(5 * time.Second)
	dis, err := aliceClient.ListDatumAll(pfs.DefaultProjectName, pipeline, jobID)
	require.NoError(t, err)
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		for _, di := range dis {
			if _, err := aliceClient.InspectDatum(pfs.DefaultProjectName, pipeline, jobID, di.Datum.Id); err != nil {
				continue
			}
		}
		return nil
	})
}

func TestPipelineNewInput(t *testing.T) {
	t.Parallel()
	env := envWithAuth(t)
	c := env.PachClient
	alice := tu.Robot(tu.UniqueString("alice"))
	aliceClient := tu.AuthenticateClient(t, c, alice)

	// alice creates three repos and commits to them
	var repo []string
	for i := 0; i < 3; i++ {
		repo = append(repo, tu.UniqueString(fmt.Sprint("TestPipelineNewInput-", i, "-")))
		require.NoError(t, aliceClient.CreateRepo(pfs.DefaultProjectName, repo[i]))
		require.Equal(t, tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, repo[i]))

		// Commit to repo
		err := aliceClient.PutFile(
			client.NewCommit(pfs.DefaultProjectName, repo[i], "master", ""), "/"+repo[i], strings.NewReader(repo[i]))
		require.NoError(t, err)
	}

	// alice creates a pipeline
	pipeline := tu.UniqueString("alice-pipeline")
	require.NoError(t, aliceClient.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"", // default image: DefaultUserImage
		[]string{"bash"},
		[]string{"cp /pfs/*/* /pfs/out/"},
		&pps.ParallelismSpec{Constant: 1},
		client.NewUnionInput(
			client.NewPFSInput(pfs.DefaultProjectName, repo[0], "/*"),
			client.NewPFSInput(pfs.DefaultProjectName, repo[1], "/*"),
		),
		"", // default output branch: master
		false,
	))
	// Make sure the input and output repos have appropriate ACLs
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Pl(pfs.DefaultProjectName, pipeline), auth.RepoReaderRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, repo[0]))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Pl(pfs.DefaultProjectName, pipeline), auth.RepoReaderRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, repo[1]))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Pl(pfs.DefaultProjectName, pipeline), auth.RepoWriterRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, pipeline))
	// repo[2] is not on pipeline -- doesn't include 'pipeline'
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, repo[2]))

	// make sure the pipeline runs
	require.NoErrorWithinT(t, 4*time.Minute, func() error {
		_, err := aliceClient.WaitCommit(pfs.DefaultProjectName, pipeline, "master", "")
		return err
	})

	// alice updates the pipeline to replace repo[0] with repo[2]
	require.NoError(t, aliceClient.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"", // default image: DefaultUserImage
		[]string{"bash"},
		[]string{"cp /pfs/*/* /pfs/out/"},
		&pps.ParallelismSpec{Constant: 1},
		client.NewUnionInput(
			client.NewPFSInput(pfs.DefaultProjectName, repo[1], "/*"),
			client.NewPFSInput(pfs.DefaultProjectName, repo[2], "/*"),
		),
		"", // default output branch: master
		true,
	))
	// Make sure the input and output repos have appropriate ACLs
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Pl(pfs.DefaultProjectName, pipeline), auth.RepoReaderRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, repo[1]))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Pl(pfs.DefaultProjectName, pipeline), auth.RepoReaderRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, repo[2]))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Pl(pfs.DefaultProjectName, pipeline), auth.RepoWriterRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, pipeline))
	// repo[0] is not on pipeline -- doesn't include 'pipeline'
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, repo[0]))

	// make sure the pipeline still runs
	require.NoErrorWithinT(t, 2*time.Minute, func() error {
		_, err := aliceClient.WaitCommit(pfs.DefaultProjectName, pipeline, "master", "")
		return err
	})
}

func TestModifyMembers(t *testing.T) {
	t.Parallel()
	env := envWithAuth(t)
	c := env.PachClient
	alice := tu.Robot(tu.UniqueString("alice"))
	bob := tu.Robot(tu.UniqueString("bob"))
	organization := tu.UniqueString("organization")
	engineering := tu.UniqueString("engineering")
	security := tu.UniqueString("security")

	adminClient := tu.AuthenticateClient(t, c, auth.RootUser)

	// This is a sequence dependent list of tests
	tests := []struct {
		Requests []*auth.ModifyMembersRequest
		Expected map[string][]string
	}{
		{
			[]*auth.ModifyMembersRequest{
				{
					Add:   []string{alice},
					Group: organization,
				},
				{
					Add:   []string{alice},
					Group: organization,
				},
			},
			map[string][]string{
				alice: {organization},
			},
		},
		{
			[]*auth.ModifyMembersRequest{
				{
					Add:   []string{bob},
					Group: organization,
				},
				{
					Add:   []string{alice, bob},
					Group: engineering,
				},
				{
					Add:   []string{bob},
					Group: security,
				},
			},
			map[string][]string{
				alice: {organization, engineering},
				bob:   {organization, engineering, security},
			},
		},
		{
			[]*auth.ModifyMembersRequest{
				{
					Add:    []string{alice},
					Remove: []string{bob},
					Group:  security,
				},
				{
					Remove: []string{bob},
					Group:  engineering,
				},
			},
			map[string][]string{
				alice: {organization, engineering, security},
				bob:   {organization},
			},
		},
		{
			[]*auth.ModifyMembersRequest{
				{
					Remove: []string{alice, bob},
					Group:  organization,
				},
				{
					Remove: []string{alice, bob},
					Group:  security,
				},
				{
					Add:    []string{alice},
					Remove: []string{alice},
					Group:  organization,
				},
				{
					Add:    []string{},
					Remove: []string{},
					Group:  organization,
				},
			},
			map[string][]string{
				alice: {engineering},
				bob:   {},
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			for _, req := range test.Requests {
				_, err := adminClient.ModifyMembers(adminClient.Ctx(), req)
				require.NoError(t, err)
			}

			for username, groups := range test.Expected {
				groupsActual, err := adminClient.GetGroupsForPrincipal(adminClient.Ctx(), &auth.GetGroupsForPrincipalRequest{
					Principal: username,
				})
				require.NoError(t, err)
				require.ElementsEqual(t, groups, groupsActual.Groups)

				for _, group := range groups {
					users, err := adminClient.GetUsers(adminClient.Ctx(), &auth.GetUsersRequest{
						Group: group,
					})
					require.NoError(t, err)
					require.OneOfEquals(t, username, users.Usernames)
				}
			}
		})
	}
}

func TestSetGroupsForUser(t *testing.T) {
	t.Parallel()
	env := envWithAuth(t)
	c := env.PachClient
	alice := tu.Robot(tu.UniqueString("alice"))
	organization := tu.UniqueString("organization")
	engineering := tu.UniqueString("engineering")
	security := tu.UniqueString("security")

	adminClient := tu.AuthenticateClient(t, c, auth.RootUser)

	groups := []string{organization, engineering}
	_, err := adminClient.SetGroupsForUser(adminClient.Ctx(), &auth.SetGroupsForUserRequest{
		Username: alice,
		Groups:   groups,
	})
	require.NoError(t, err)
	groupsActual, err := adminClient.GetGroupsForPrincipal(adminClient.Ctx(), &auth.GetGroupsForPrincipalRequest{
		Principal: alice,
	})
	require.NoError(t, err)
	require.ElementsEqual(t, groups, groupsActual.Groups)
	for _, group := range groups {
		users, err := adminClient.GetUsers(adminClient.Ctx(), &auth.GetUsersRequest{
			Group: group,
		})
		require.NoError(t, err)
		require.OneOfEquals(t, alice, users.Usernames)
	}

	groups = append(groups, security)
	_, err = adminClient.SetGroupsForUser(adminClient.Ctx(), &auth.SetGroupsForUserRequest{
		Username: alice,
		Groups:   groups,
	})
	require.NoError(t, err)
	groupsActual, err = adminClient.GetGroupsForPrincipal(adminClient.Ctx(), &auth.GetGroupsForPrincipalRequest{
		Principal: alice,
	})
	require.NoError(t, err)
	require.ElementsEqual(t, groups, groupsActual.Groups)
	for _, group := range groups {
		users, err := adminClient.GetUsers(adminClient.Ctx(), &auth.GetUsersRequest{
			Group: group,
		})
		require.NoError(t, err)
		require.OneOfEquals(t, alice, users.Usernames)
	}

	groups = groups[:1]
	_, err = adminClient.SetGroupsForUser(adminClient.Ctx(), &auth.SetGroupsForUserRequest{
		Username: alice,
		Groups:   groups,
	})
	require.NoError(t, err)
	groupsActual, err = adminClient.GetGroupsForPrincipal(adminClient.Ctx(), &auth.GetGroupsForPrincipalRequest{
		Principal: alice,
	})
	require.NoError(t, err)
	require.ElementsEqual(t, groups, groupsActual.Groups)
	for _, group := range groups {
		users, err := adminClient.GetUsers(adminClient.Ctx(), &auth.GetUsersRequest{
			Group: group,
		})
		require.NoError(t, err)
		require.OneOfEquals(t, alice, users.Usernames)
	}

	groups = []string{}
	_, err = adminClient.SetGroupsForUser(adminClient.Ctx(), &auth.SetGroupsForUserRequest{
		Username: alice,
		Groups:   groups,
	})
	require.NoError(t, err)
	groupsActual, err = adminClient.GetGroupsForPrincipal(adminClient.Ctx(), &auth.GetGroupsForPrincipalRequest{
		Principal: alice,
	})
	require.NoError(t, err)
	require.ElementsEqual(t, groups, groupsActual.Groups)
	for _, group := range groups {
		users, err := adminClient.GetUsers(adminClient.Ctx(), &auth.GetUsersRequest{
			Group: group,
		})
		require.NoError(t, err)
		require.OneOfEquals(t, alice, users.Usernames)
	}
}

func TestGetOwnGroups(t *testing.T) {
	t.Parallel()
	env := envWithAuth(t)
	c := env.PachClient
	alice := tu.Robot(tu.UniqueString("alice"))
	organization := tu.UniqueString("organization")
	engineering := tu.UniqueString("engineering")
	security := tu.UniqueString("security")

	adminClient := tu.AuthenticateClient(t, c, auth.RootUser)

	_, err := adminClient.SetGroupsForUser(adminClient.Ctx(), &auth.SetGroupsForUserRequest{
		Username: alice,
		Groups:   []string{organization, engineering, security},
	})
	require.NoError(t, err)

	aliceClient := tu.AuthenticateClient(t, c, alice)
	groups, err := aliceClient.GetGroups(aliceClient.Ctx(), &auth.GetGroupsRequest{})
	require.NoError(t, err)
	require.ElementsEqual(t, []string{organization, engineering, security}, groups.Groups)

	groups, err = adminClient.GetGroups(adminClient.Ctx(), &auth.GetGroupsRequest{})
	require.NoError(t, err)
	require.Equal(t, 0, len(groups.Groups))
}

// TestGetJobsBugFix tests the fix for https://github.com/pachyderm/pachyderm/v2/issues/2879
// where calling pps.ListJob when not logged in would delete all old jobs
func TestGetJobsBugFix(t *testing.T) {
	t.Parallel()
	env := envWithAuth(t)
	c := env.PachClient
	alice := tu.Robot(tu.UniqueString("alice"))
	aliceClient, anonClient := tu.AuthenticateClient(t, c, alice), tu.UnauthenticatedPachClient(t, c)

	// alice creates a repo
	repo := tu.UniqueString(t.Name())
	commit := client.NewCommit(pfs.DefaultProjectName, repo, "master", "")
	require.NoError(t, aliceClient.CreateRepo(pfs.DefaultProjectName, repo))
	require.Equal(t, tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, repo))
	err := aliceClient.PutFile(commit, "/file", strings.NewReader("lorem ipsum"))
	require.NoError(t, err)

	// alice creates a pipeline
	pipeline := tu.UniqueString("alice-pipeline")
	require.NoError(t, aliceClient.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"", // default image: DefaultUserImage
		[]string{"bash"},
		[]string{"cp /pfs/*/* /pfs/out/"},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(pfs.DefaultProjectName, repo, "/*"),
		"", // default output branch: master
		false,
	))

	// Wait for pipeline to finish
	_, err = aliceClient.WaitCommit(pfs.DefaultProjectName, pipeline, "master", commit.Id)
	require.NoError(t, err)

	// alice calls 'list job'
	jobs, err := aliceClient.ListJob(pfs.DefaultProjectName, "", nil, -1 /*history*/, true)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobs))

	// anonClient calls 'list job'
	_, err = anonClient.ListJob(pfs.DefaultProjectName, "", nil, -1 /*history*/, true)
	require.YesError(t, err)
	require.Matches(t, "no authentication token", err.Error())

	// alice calls 'list job' again, and the existing job must still be present
	jobs2, err := aliceClient.ListJob(pfs.DefaultProjectName, "", nil, -1 /*history*/, true)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobs2))
	require.Equal(t, jobs[0].Job.Id, jobs2[0].Job.Id)
}

// TestDeleteFailedPipeline creates a pipeline with an invalid image and then
// tries to delete it (which shouldn't be blocked by the auth system)
func TestDeleteFailedPipeline(t *testing.T) {
	t.Parallel()
	env := envWithAuth(t)
	c := env.PachClient
	alice := tu.Robot(tu.UniqueString("alice"))
	aliceClient := tu.AuthenticateClient(t, c, alice)

	// Create input repo w/ initial commit
	repo := tu.UniqueString(t.Name())
	commit := client.NewCommit(pfs.DefaultProjectName, repo, "master", "")
	require.NoError(t, aliceClient.CreateRepo(pfs.DefaultProjectName, repo))
	err := aliceClient.PutFile(commit, "/file", strings.NewReader("1"))
	require.NoError(t, err)

	// Create pipeline
	pipeline := tu.UniqueString("pipeline")
	require.NoError(t, aliceClient.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"does-not-exist", // nonexistant image
		[]string{"true"}, nil,
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(pfs.DefaultProjectName, repo, "/*"),
		"", // default output branch: master
		false,
	))
	require.NoError(t, aliceClient.DeletePipeline(pfs.DefaultProjectName, pipeline, true))

	// Get the latest commit from the input repo (which should be an alias from
	// when the pipeline was created)
	commitInfo, err := aliceClient.InspectCommit(pfs.DefaultProjectName, repo, "master", "")
	require.NoError(t, err)

	// make sure the pipeline failure doesn't cause waits to block indefinitely
	require.NoErrorWithinT(t, 30*time.Second, func() error {
		_, err := aliceClient.WaitCommitSetAll(commitInfo.Commit.Id)
		return err
	})
}

// TestDeletePipelineMissingRepos creates a pipeline, force-deletes its input
// and output repos, and then confirms that DeletePipeline still works
// (i.e. the missing repos/ACLs don't cause an auth error).
func TestDeletePipelineMissingRepos(t *testing.T) {
	t.Parallel()
	env := envWithAuth(t)
	c := env.PachClient
	alice := tu.Robot(tu.UniqueString("alice"))
	aliceClient := tu.AuthenticateClient(t, c, alice)

	// Create input repo w/ initial commit
	repo := tu.UniqueString(t.Name())
	commit := client.NewCommit(pfs.DefaultProjectName, repo, "master", "")
	require.NoError(t, aliceClient.CreateRepo(pfs.DefaultProjectName, repo))
	err := aliceClient.PutFile(commit, "/file", strings.NewReader("1"))
	require.NoError(t, err)

	// Create pipeline
	pipeline := tu.UniqueString("pipeline")
	require.NoError(t, aliceClient.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"does-not-exist", // nonexistant image
		[]string{"true"}, nil,
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(pfs.DefaultProjectName, repo, "/*"),
		"", // default output branch: master
		false,
	))

	// force-delete input and output repos
	require.NoError(t, aliceClient.DeleteRepo(pfs.DefaultProjectName, repo, true))

	// Attempt to delete the pipeline--must succeed
	require.NoError(t, aliceClient.DeletePipeline(pfs.DefaultProjectName, pipeline, false))
	pis, err := aliceClient.ListPipeline(false)
	require.NoError(t, err)
	for _, pi := range pis {
		if pi.Pipeline.Name == pipeline {
			t.Fatalf("Expected %q to be deleted, but still present", pipeline)
		}
	}
}

func TestDeleteExpiredAuthTokens(t *testing.T) {
	t.Parallel()
	env := envWithAuth(t)
	c := env.PachClient
	// generate auth credentials
	adminClient := tu.AuthenticateClient(t, c, auth.RootUser)

	// create a token that will instantly expire, a token that will expire later, and a token that will never expire
	noExpirationResp, noExpErr := adminClient.GetRobotToken(adminClient.Ctx(), &auth.GetRobotTokenRequest{Robot: "robot:alice"})
	require.NoError(t, noExpErr)

	fastExpirationResp, fastExpErr := adminClient.GetRobotToken(adminClient.Ctx(), &auth.GetRobotTokenRequest{Robot: "robot:alice", Ttl: 1})
	require.NoError(t, fastExpErr)

	slowExpirationResp, slowExpErr := adminClient.GetRobotToken(adminClient.Ctx(), &auth.GetRobotTokenRequest{Robot: "robot:alice", Ttl: 1000})
	require.NoError(t, slowExpErr)

	contains := func(tokens []*auth.TokenInfo, hashedToken string) bool {
		for _, v := range tokens {
			if v.HashedToken == hashedToken {
				return true
			}
		}
		return false
	}

	// query all tokens to show that the instantly expired one is expired
	extractTokensResp, firstExtractErr := adminClient.ExtractAuthTokens(adminClient.Ctx(), &auth.ExtractAuthTokensRequest{})
	require.NoError(t, firstExtractErr)

	preDeleteTokens := extractTokensResp.Tokens
	require.Equal(t, 3, len(preDeleteTokens), "all three tokens should be returned")
	require.True(t, contains(preDeleteTokens, auth.HashToken(noExpirationResp.Token)), "robot token without expiration should be extracted")
	require.True(t, contains(preDeleteTokens, auth.HashToken(fastExpirationResp.Token)), "robot token without expiration should be extracted")
	require.True(t, contains(preDeleteTokens, auth.HashToken(slowExpirationResp.Token)), "robot token without expiration should be extracted")

	// wait for the one token to expire
	time.Sleep(time.Duration(2) * time.Second)

	// record admin token
	adminToken := adminClient.AuthToken()

	// before deleting, check that WhoAmI call still fails for existing & expired token
	adminClient.SetAuthToken(fastExpirationResp.Token)
	_, whoAmIErr := adminClient.WhoAmI(adminClient.Ctx(), &auth.WhoAmIRequest{})
	require.True(t, auth.IsErrExpiredToken(whoAmIErr))

	// run DeleteExpiredAuthTokens RPC and verify that only the instantly expired token is inaccessible
	adminClient.SetAuthToken(adminToken)
	_, deleteErr := adminClient.DeleteExpiredAuthTokens(adminClient.Ctx(), &auth.DeleteExpiredAuthTokensRequest{})
	require.NoError(t, deleteErr)

	extractTokensAfterDeleteResp, sndExtractErr := adminClient.ExtractAuthTokens(adminClient.Ctx(), &auth.ExtractAuthTokensRequest{})
	require.NoError(t, sndExtractErr)

	postDeleteTokens := extractTokensAfterDeleteResp.Tokens

	require.Equal(t, 2, len(postDeleteTokens), "only the two unexpired tokens should be returned.")
	require.True(t, contains(postDeleteTokens, auth.HashToken(noExpirationResp.Token)), "robot token without expiration should be extracted")
	require.True(t, contains(postDeleteTokens, auth.HashToken(slowExpirationResp.Token)), "robot token without expiration should be extracted")
}

func TestExpiredClusterLocksOutUsers(t *testing.T) {
	t.Parallel()
	env := envWithAuth(t)
	c := env.PachClient
	adminClient := tu.AuthenticateClient(t, c, auth.RootUser)

	alice := tu.UniqueString("robot:alice")
	aliceClient := tu.AuthenticateClient(t, c, alice)

	repo := tu.UniqueString("TestRotateAuthToken")
	require.NoError(t, aliceClient.CreateRepo(pfs.DefaultProjectName, repo))

	// Admin can list repos
	repoInfo, err := adminClient.ListRepo()
	require.NoError(t, err)
	require.Equal(t, 1, len(repoInfo))

	// Alice can list repos
	repoInfo, err = aliceClient.ListRepo()
	require.NoError(t, err)
	require.Equal(t, 1, len(repoInfo))

	// set Enterprise Token value to have expired
	ts := &timestamppb.Timestamp{Seconds: time.Now().Unix() - 100}
	resp, err := adminClient.License.Activate(adminClient.Ctx(),
		&license.ActivateRequest{
			ActivationCode: tu.GetTestEnterpriseCode(t),
			Expires:        ts,
		})
	require.NoError(t, err)
	require.True(t, resp.GetInfo().Expires.Seconds == ts.Seconds)

	// Heartbeat forces Enterprise Service to refresh it's view of the LicenseRecord
	_, err = adminClient.Enterprise.Heartbeat(adminClient.Ctx(), &enterprise.HeartbeatRequest{})
	require.NoError(t, err)

	// verify Alice can no longer operate on the system
	_, err = aliceClient.ListRepo()
	require.YesError(t, err)
	require.True(t, strings.Contains(err.Error(), "Pachyderm Enterprise is not active"))

	// verify that admin can still complete an operation (ex. ListRepo)
	repoInfo, err = adminClient.ListRepo()
	require.NoError(t, err)
	require.Equal(t, 1, len(repoInfo))

	// admin grants alice cluster admin role
	_, err = adminClient.AuthAPIClient.ModifyRoleBinding(adminClient.Ctx(),
		&auth.ModifyRoleBindingRequest{
			Principal: alice,
			Roles:     []string{auth.ClusterAdminRole},
			Resource:  &auth.Resource{Type: auth.ResourceType_CLUSTER},
		})
	require.NoError(t, err)

	// verify that the Alice can now operate on cluster again
	repoInfo, err = aliceClient.ListRepo()
	require.NoError(t, err)
	require.Equal(t, 1, len(repoInfo))
}

// TestRolesForPermission tests all users can look up the roles that correspond to
// a given permission.
func TestRolesForPermission(t *testing.T) {
	t.Parallel()
	env := envWithAuth(t)
	c := env.PachClient
	alice := tu.UniqueString("robot:alice")
	aliceClient := tu.AuthenticateClient(t, c, alice)
	resp, err := aliceClient.GetRolesForPermission(aliceClient.Ctx(), &auth.GetRolesForPermissionRequest{Permission: auth.Permission_REPO_READ})
	require.NoError(t, err)

	names := make([]string, len(resp.Roles))
	for i, r := range resp.Roles {
		names[i] = r.Name
	}
	sort.Strings(names)
	require.Equal(t, []string{"clusterAdmin", "projectOwner", "repoOwner", "repoReader", "repoWriter"}, names)
}

// TODO: This test mirrors TestLoad in src/server/pfs/server/testing/load_test.go.
// Need to restructure testing such that we have the implementation of this
// test in one place while still being able to test auth enabled and disabled clusters.
// Other tests that follow this pattern: TestPutFileURL, TestGetFileURL.
func TestLoad(t *testing.T) {
	t.Parallel()
	env := envWithAuth(t)
	c := env.PachClient
	alice := tu.UniqueString("robot:alice")
	aliceClient := tu.AuthenticateClient(t, c, alice)
	resp, err := aliceClient.PfsAPIClient.RunLoadTestDefault(aliceClient.Ctx(), &emptypb.Empty{})
	require.NoError(t, err)
	buf := &bytes.Buffer{}
	require.NoError(t, cmdutil.Encoder("", buf).EncodeProto(resp))
	require.Equal(t, "", resp.Error, buf.String())
}

// TODO: Refer to TestLoad
func TestPutFileURL(t *testing.T) {
	t.Parallel()
	env := envWithAuth(t)
	c := env.PachClient
	alice := tu.UniqueString("robot:alice")
	aliceClient := tu.AuthenticateClient(t, c, alice)
	ctx := aliceClient.Ctx()
	repo := "repo"
	require.NoError(t, aliceClient.CreateRepo(pfs.DefaultProjectName, repo))
	commit, err := aliceClient.StartCommit(pfs.DefaultProjectName, repo, "master")
	require.NoError(t, err)
	bucket, url := dockertestenv.NewTestBucket(ctx, t)
	paths := []string{"files/foo", "files/bar", "files/fizz"}
	for _, path := range paths {
		require.NoError(t, bucket.WriteAll(ctx, path, []byte(path), nil))
	}
	for _, p := range paths {
		objURL := url + p
		require.NoError(t, aliceClient.PutFileURL(commit, p, objURL, false))
	}
	srcURL := url + "files"
	require.NoError(t, aliceClient.PutFileURL(commit, "recursive", srcURL, true))
	check := func() {
		cis, err := aliceClient.ListCommit(client.NewRepo(pfs.DefaultProjectName, repo), nil, nil, 0)
		require.NoError(t, err)
		require.Equal(t, 1, len(cis))
		for _, path := range paths {
			var b bytes.Buffer
			require.NoError(t, aliceClient.GetFile(commit, path, &b))
			require.Equal(t, path, b.String())
			b.Reset()
			require.NoError(t, aliceClient.GetFile(commit, filepath.Join("recursive", filepath.Base(path)), &b))
			require.Equal(t, path, b.String())
		}
	}
	check()
	require.NoError(t, finishCommit(aliceClient, repo, commit.Branch.Name, commit.Id))
	check()
}

func finishCommit(pachClient *client.APIClient, repo, branch, id string) error {
	return finishProjectCommit(pachClient, pfs.DefaultProjectName, repo, branch, id)
}

func finishProjectCommit(pachClient *client.APIClient, project, repo, branch, id string) error {
	if err := pachClient.FinishCommit(project, repo, branch, id); err != nil {
		if !pfsserver.IsCommitFinishedErr(err) {
			return err
		}
	}
	_, err := pachClient.WaitCommit(project, repo, branch, id)
	return err
}

// TODO: Refer to TestLoad.
func TestGetFileURL(t *testing.T) {
	t.Parallel()
	env := envWithAuth(t)
	c := env.PachClient
	alice := tu.UniqueString("robot:alice")
	aliceClient := tu.AuthenticateClient(t, c, alice)
	ctx := aliceClient.Ctx()
	repo := "repo"
	require.NoError(t, aliceClient.CreateRepo(pfs.DefaultProjectName, repo))
	commit, err := aliceClient.StartCommit(pfs.DefaultProjectName, repo, "master")
	require.NoError(t, err)
	paths := []string{"files/foo", "files/bar", "files/fizz"}
	for _, path := range paths {
		require.NoError(t, aliceClient.PutFile(commit, path, strings.NewReader(path)))
	}
	check := func() {
		bucket, url := dockertestenv.NewTestBucket(ctx, t)
		for _, path := range paths {
			require.NoError(t, env.PachClient.GetFileURL(commit, path, url))
		}
		for _, path := range paths {
			data, err := bucket.ReadAll(ctx, path)
			require.NoError(t, err)
			require.True(t, bytes.Equal([]byte(path), data))
			require.NoError(t, bucket.Delete(ctx, path))
		}
		require.NoError(t, env.PachClient.GetFileURL(commit, "files/*", url))
		for _, path := range paths {
			data, err := bucket.ReadAll(ctx, path)
			require.NoError(t, err)
			require.True(t, bytes.Equal([]byte(path), data))
			require.NoError(t, bucket.Delete(ctx, path))
		}
		prefix := "/prefix"
		require.NoError(t, env.PachClient.GetFileURL(commit, "files/*", url+prefix))
		for _, path := range paths {
			data, err := bucket.ReadAll(ctx, prefix+"/"+path)
			require.NoError(t, err)
			require.True(t, bytes.Equal([]byte(path), data))
		}
	}
	check()
	require.NoError(t, finishCommit(aliceClient, repo, commit.Branch.Name, commit.Id))
	check()
}

// TestGetPermissions tests that GetPermissions and GetPermissionsForPrincipal work for repos and the cluster itself
func TestGetPermissions(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	c := env.PachClient
	tu.ActivateAuthClient(t, c, peerPort)
	alice, bob := tu.Robot(tu.UniqueString("alice")), tu.Robot(tu.UniqueString("bob"))
	aliceClient, rootClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, auth.RootUser)

	// alice creates a repo and makes Bob a writer
	repo := tu.UniqueString(t.Name())
	require.NoError(t, aliceClient.CreateRepo(pfs.DefaultProjectName, repo))
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, repo, bob, []string{auth.RepoWriterRole}))

	// alice can get her own permissions on the cluster (none) and on the repo (repoOwner)
	permissions, err := aliceClient.GetPermissions(aliceClient.Ctx(), &auth.GetPermissionsRequest{Resource: &auth.Resource{Type: auth.ResourceType_CLUSTER}})
	require.NoError(t, err)
	require.ElementsEqual(t, []string{auth.ProjectCreatorRole}, permissions.Roles)

	permissions, err = aliceClient.GetPermissions(aliceClient.Ctx(), &auth.GetPermissionsRequest{Resource: &auth.Resource{Type: auth.ResourceType_REPO, Name: pfs.DefaultProjectName + "/" + repo}})
	require.NoError(t, err)
	require.Equal(t, []string{auth.RepoOwnerRole}, permissions.Roles)

	// the root user can get bob's permissions
	permissions, err = rootClient.GetPermissionsForPrincipal(rootClient.Ctx(), &auth.GetPermissionsForPrincipalRequest{Resource: &auth.Resource{Type: auth.ResourceType_CLUSTER}, Principal: bob})
	require.NoError(t, err)
	require.ElementsEqual(t, []string{auth.ProjectCreatorRole}, permissions.Roles)

	permissions, err = rootClient.GetPermissionsForPrincipal(rootClient.Ctx(), &auth.GetPermissionsForPrincipalRequest{Resource: &auth.Resource{Type: auth.ResourceType_REPO, Name: pfs.DefaultProjectName + "/" + repo}, Principal: bob})
	require.NoError(t, err)
	require.Equal(t, []string{auth.RepoWriterRole}, permissions.Roles)

	// alice cannot get bob's permissions
	_, err = aliceClient.GetPermissionsForPrincipal(aliceClient.Ctx(), &auth.GetPermissionsForPrincipalRequest{Resource: &auth.Resource{Type: auth.ResourceType_CLUSTER}, Principal: bob})
	require.YesError(t, err)
	require.Matches(t, "is not authorized to perform this operation - needs permissions", err.Error())

	_, err = aliceClient.GetPermissionsForPrincipal(aliceClient.Ctx(), &auth.GetPermissionsForPrincipalRequest{Resource: &auth.Resource{Type: auth.ResourceType_REPO, Name: pfs.DefaultProjectName + "/" + repo}, Principal: bob})
	require.YesError(t, err)
	require.Matches(t, "is not authorized to perform this operation - needs permissions", err.Error())
}

func TestGetPermissions_Project(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	c := env.PachClient
	tu.ActivateAuthClient(t, c, peerPort)

	rootClient := tu.AuthenticateClient(t, c, auth.RootUser)
	project := tu.UniqueString("project-")
	err := rootClient.CreateProject(project)
	require.NoError(t, err, "should create project")

	repo := tu.UniqueString("repo-")
	err = rootClient.CreateRepo(project, repo)
	require.NoError(t, err, "should create repo")

	err = rootClient.ModifyProjectRoleBinding(project, "user:alice", []string{"projectOwner"})
	require.NoError(t, err, "should make alice a project owner")

	perms, err := rootClient.GetPermissionsForPrincipal(rootClient.Ctx(), &auth.GetPermissionsForPrincipalRequest{
		Resource: &auth.Resource{
			Type: auth.ResourceType_REPO,
			Name: (&pfs.Repo{Name: repo, Project: &pfs.Project{Name: project}, Type: pfs.UserRepoType}).String(),
		},
		Principal: "user:alice",
	})
	require.NoError(t, err, "should get permissions for alice on repo %v", repo)

	require.OneOfEquals(t, "projectOwner", perms.Roles, "alice should be projectOwner on the repo")
}

// TestDeactivateFSAdmin tests that users with the FS admin role can't call Deactivate
func TestDeactivateFSAdmin(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	c := env.PachClient
	tu.ActivateAuthClient(t, c, peerPort)
	alice := tu.Robot(tu.UniqueString("alice"))
	aliceClient, adminClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, auth.RootUser)

	// admin makes alice an fs admin
	require.NoError(t, adminClient.ModifyClusterRoleBinding(alice, []string{auth.RepoOwnerRole}))

	// wait until alice shows up in admin list
	resp, err := aliceClient.GetClusterRoleBinding()
	require.NoError(t, err)
	require.Equal(t, tu.BuildClusterBindings(alice, auth.RepoOwnerRole), resp)

	// alice tries to deactivate, but doesn't have permission as an FS admin
	_, err = aliceClient.Deactivate(aliceClient.Ctx(), &auth.DeactivateRequest{})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
}

// TestExtractAuthToken tests that admins can extract hashed robot auth tokens
func TestExtractAuthToken(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	c := env.PachClient
	tu.ActivateAuthClient(t, c, peerPort)
	alice := tu.Robot(tu.UniqueString("alice"))
	aliceClient, adminClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, auth.RootUser)

	// alice can't extract auth tokens because she's not an admin
	_, err := aliceClient.ExtractAuthTokens(aliceClient.Ctx(), &auth.ExtractAuthTokensRequest{})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())

	// Create a token with a TTL and confirm it is extracted with an expiration
	tokenResp, err := adminClient.GetRobotToken(adminClient.Ctx(), &auth.GetRobotTokenRequest{Robot: "other", Ttl: 1000})
	require.NoError(t, err)

	// Create a token without a TTL and confirm it is extracted
	tokenRespTwo, err := adminClient.GetRobotToken(adminClient.Ctx(), &auth.GetRobotTokenRequest{Robot: "otherTwo"})
	require.NoError(t, err)

	// admins can extract auth tokens
	resp, err := adminClient.ExtractAuthTokens(adminClient.Ctx(), &auth.ExtractAuthTokensRequest{})
	require.NoError(t, err)

	// only robot tokens are extracted, so only the admin token (not the alice one) should be included
	containsToken := func(plaintext, subject string, expires bool) error {
		hash := auth.HashToken(plaintext)
		for _, token := range resp.Tokens {
			if token.HashedToken == hash {
				require.Equal(t, subject, token.Subject)
				if expires {
					require.True(t, protoutil.MustTime(token.Expiration).After(time.Now()))
				} else {
					require.Nil(t, token.Expiration)
				}
				return nil
			}
		}
		return errors.Errorf("didn't find a token with hash %q", hash)
	}

	require.NoError(t, containsToken(tokenResp.Token, "robot:other", true))
	require.NoError(t, containsToken(tokenRespTwo.Token, "robot:otherTwo", false))
}

// TestRestoreAuthToken tests that admins can restore hashed auth tokens that have been extracted
func TestRestoreAuthToken(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	c := env.PachClient
	tu.ActivateAuthClient(t, c, peerPort)
	// Create a request to restore a token with known plaintext
	req := &auth.RestoreAuthTokenRequest{
		Token: &auth.TokenInfo{
			HashedToken: fmt.Sprintf("%x", sha256.Sum256([]byte("an-auth-token"))),
			Subject:     "robot:restored",
		},
	}

	alice := tu.Robot(tu.UniqueString("alice"))
	aliceClient, adminClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, auth.RootUser)

	// alice can't restore auth tokens because she's not an admin
	_, err := aliceClient.RestoreAuthToken(aliceClient.Ctx(), req)
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())

	// admins can restore auth tokens
	_, err = adminClient.RestoreAuthToken(adminClient.Ctx(), req)
	require.NoError(t, err)

	req.Token.Subject = "robot:overwritten"
	_, err = adminClient.RestoreAuthToken(adminClient.Ctx(), req)
	require.YesError(t, err)
	require.Equal(t, "rpc error: code = Unknown desc = error restoring auth token: cannot overwrite existing token with same hash", err.Error())

	// now we can authenticate with the restored token
	aliceClient.SetAuthToken("an-auth-token")
	whoAmIResp, err := aliceClient.WhoAmI(aliceClient.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, "robot:restored", whoAmIResp.Username)
	require.Nil(t, whoAmIResp.Expiration)

	// restore a token with an expiration date in the past
	req.Token.HashedToken = fmt.Sprintf("%x", sha256.Sum256([]byte("expired-token")))
	pastExpiration := time.Now().Add(-1 * time.Minute)
	req.Token.Expiration = protoutil.MustTimestamp(pastExpiration)

	_, err = adminClient.RestoreAuthToken(adminClient.Ctx(), req)
	require.YesError(t, err)
	require.True(t, auth.IsErrExpiredToken(err))

	// restore a token with an expiration date in the future
	req.Token.HashedToken = fmt.Sprintf("%x", sha256.Sum256([]byte("expiring-token")))
	futureExpiration := time.Now().Add(10 * time.Minute)
	req.Token.Expiration = protoutil.MustTimestamp(futureExpiration)

	_, err = adminClient.RestoreAuthToken(adminClient.Ctx(), req)
	require.NoError(t, err)

	aliceClient.SetAuthToken("expiring-token")
	whoAmIResp, err = aliceClient.WhoAmI(aliceClient.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)

	// Relying on time.Now is gross but the token should have a TTL in the
	// next 10 minutes
	require.True(t, protoutil.MustTime(whoAmIResp.Expiration).After(time.Now()))
	require.True(t, protoutil.MustTime(whoAmIResp.Expiration).Before(time.Now().Add(time.Duration(600)*time.Second)))
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
	t.Parallel()
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	c := env.PachClient
	tu.ActivateAuthClient(t, c, peerPort)
	alice := tu.Robot(tu.UniqueString("alice"))
	aliceClient, rootClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, auth.RootUser)

	// Create input repo w/ initial commit
	repo := tu.UniqueString(t.Name())
	commit := client.NewCommit(pfs.DefaultProjectName, repo, "master", "")
	require.NoError(t, aliceClient.CreateRepo(pfs.DefaultProjectName, repo))
	err := aliceClient.PutFile(commit, "/file.1", strings.NewReader("1"))
	require.NoError(t, err)

	// Create pipeline
	pipeline := tu.UniqueString("pipeline")
	require.NoError(t, aliceClient.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"", // default image: DefaultUserImage
		[]string{"bash"},
		[]string{
			"sleep 10",
			"cp /pfs/*/* /pfs/out/",
		},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(pfs.DefaultProjectName, repo, "/*"),
		"", // default output branch: master
		false,
	))

	// Revoke pipeline's access to output repo while 'sleep 10' is running (so
	// that it fails)
	require.NoError(t, rootClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, repo, tu.Pl(pfs.DefaultProjectName, pipeline), []string{}))

	// make sure the pipeline either fails or restarts RC & finishes
	require.NoErrorWithinT(t, 30*time.Second, func() error {
		_, err := aliceClient.WaitCommit(pfs.DefaultProjectName, pipeline, "master", commit.Id)
		return err
	})

	// make sure the pipeline is failed
	pi, err := rootClient.InspectPipeline(pfs.DefaultProjectName, pipeline, false)
	require.NoError(t, err)
	require.Equal(t, pps.PipelineState_PIPELINE_FAILURE, pi.State)
}

// TestGetRobotTokenErrorNonAdminUser tests that non-admin users can't call
// GetRobotToken
func TestGetRobotTokenErrorNonAdminUser(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	c := env.PachClient
	tu.ActivateAuthClient(t, c, peerPort)
	alice := tu.Robot(tu.UniqueString("alice"))
	aliceClient := tu.AuthenticateClient(t, c, alice)
	resp, err := aliceClient.GetRobotToken(aliceClient.Ctx(), &auth.GetRobotTokenRequest{
		Robot: tu.UniqueString("t-1000"),
	})
	require.Nil(t, resp)
	require.YesError(t, err)
	require.Matches(t, "needs permissions \\[CLUSTER_AUTH_GET_ROBOT_TOKEN\\] on CLUSTER", err.Error())
}

// TestDeleteAll tests that you must be a cluster admin to call DeleteAll
func TestDeleteAll(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnvWithIdentity(ctx, t, dockertestenv.NewTestDBConfig(t))
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	c := env.PachClient
	tu.ActivateAuthClient(t, c, peerPort)
	alice := tu.Robot(tu.UniqueString("alice"))
	aliceClient, adminClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, auth.RootUser)

	// admin creates a repo
	repo := tu.UniqueString(t.Name())
	require.NoError(t, adminClient.CreateRepo(pfs.DefaultProjectName, repo))

	// alice calls DeleteAll, but it fails
	err := aliceClient.DeleteAll()
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())

	// admin makes alice an fs admin
	require.NoError(t, adminClient.ModifyClusterRoleBinding(alice, []string{auth.RepoOwnerRole}))

	// wait until alice shows up in admin list
	resp, err := aliceClient.GetClusterRoleBinding()
	require.NoError(t, err)
	require.Equal(t, tu.BuildClusterBindings(alice, auth.RepoOwnerRole), resp)

	// alice calls DeleteAll but it fails because she's only an fs admin
	err = aliceClient.DeleteAll()
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())

	// admin calls DeleteAll and succeeds
	require.NoError(t, adminClient.DeleteAll())
}

func TestCreateProject(t *testing.T) {
	t.Parallel()
	client := envWithAuth(t).PachClient
	alice := tu.Robot(tu.UniqueString("alice"))
	aliceClient := tu.AuthenticateClient(t, client, alice)

	// create a project and check the caller is the owner
	projectName := tu.UniqueString("project" + t.Name())
	require.NoError(t, aliceClient.CreateProject(projectName))
	require.Equal(t, tu.BuildBindings(alice, auth.ProjectOwnerRole), tu.GetProjectRoleBinding(t, aliceClient, projectName))

	// revoke cluster level role binding that grants all users ProjectCreate role
	// and see if create project fails
	rootClient := tu.AuthenticateClient(t, client, auth.RootUser)
	require.NoError(t, rootClient.ModifyClusterRoleBinding(auth.AllClusterUsersSubject, []string{}))
	require.ErrorContains(t, aliceClient.CreateProject(projectName), "not authorized to perform this operation - needs permissions [PROJECT_CREATE] on CLUSTER")
}

func TestModifyRoleBindingAccess(t *testing.T) {
	t.Parallel()

	// setup
	c := envWithAuth(t).PachClient
	clusterAdmin := tu.AuthenticateClient(t, c, auth.RootUser)
	_, aliceClient := tu.RandomRobot(t, c, "alice")
	bob, bobClient := tu.RandomRobot(t, c, "bob")
	project1, project2 := tu.UniqueString("project1"), tu.UniqueString("project2")
	repo1, repo2 := tu.UniqueString("repo1"), tu.UniqueString("repo2")
	// alice owns project1 and project1/repo1
	// bob owns project2 and project1/repo2
	require.NoError(t, aliceClient.CreateProject(project1))
	require.NoError(t, aliceClient.CreateRepo(project1, repo1))
	require.NoError(t, aliceClient.ModifyProjectRoleBinding(project1, bob, []string{auth.ProjectWriterRole}))
	require.NoError(t, bobClient.CreateRepo(project1, repo2))
	require.NoError(t, bobClient.CreateProject(project2))
	// auth resources
	clusterResource := auth.Resource{Type: auth.ResourceType_CLUSTER}
	aliceProject := auth.Resource{Type: auth.ResourceType_PROJECT, Name: project1}
	bobProject := auth.Resource{Type: auth.ResourceType_PROJECT, Name: project2}
	aliceRepoInAliceProject := auth.Resource{Type: auth.ResourceType_REPO, Name: fmt.Sprintf("%s/%s", project1, repo1)}
	bobRepoInAliceProject := auth.Resource{Type: auth.ResourceType_REPO, Name: fmt.Sprintf("%s/%s", project1, repo2)}

	tests := map[string]struct {
		client         *client.APIClient
		resource       *auth.Resource
		expectedErrMsg string
	}{
		"ClusterAdminCanModifyCluster":               {clusterAdmin, &clusterResource, ""},
		"ClusterAdminCanModifyProject":               {clusterAdmin, &aliceProject, ""},
		"ClusterAdminCanModifyRepo":                  {clusterAdmin, &aliceRepoInAliceProject, ""},
		"ProjectOwnerCanModifyProject":               {aliceClient, &aliceProject, ""},
		"ProjectOwnerCanModifyAnyRepoWithinProject":  {aliceClient, &bobRepoInAliceProject, ""},
		"RepoOwnerCanModifyRepo":                     {bobClient, &bobRepoInAliceProject, ""},
		"ProjectOwnerCannotModifyProjectTheyDontOwn": {aliceClient, &bobProject, "needs permissions [PROJECT_MODIFY_BINDINGS]"},
		"ProjectOwnerCannotModifyCluster":            {aliceClient, &clusterResource, "needs permissions [CLUSTER_MODIFY_BINDINGS]"},
		"RepoOwnerCannotModifyCluster":               {aliceClient, &clusterResource, "needs permissions [CLUSTER_MODIFY_BINDINGS]"},
		"RepoOwnerCannotModifyRepoTheyDontOwn":       {bobClient, &aliceRepoInAliceProject, "needs permissions [REPO_MODIFY_BINDINGS]"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := tc.client.ModifyRoleBinding(tc.client.Ctx(), &auth.ModifyRoleBindingRequest{
				Principal: tu.Robot("marvin"),
				Roles:     []string{},
				Resource:  tc.resource,
			})
			if tc.expectedErrMsg == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.expectedErrMsg)
			}
		})
	}
}

func TestListRepoAfterAuthActivation(t *testing.T) {
	// This test ensures that CORE-1548 doesn't come back.
	t.Parallel()
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
	c := env.PachClient.WithCtx(ctx)
	ctx = c.Ctx()

	// Create a repo in the default project.
	err := c.CreateRepo("default", "test")
	require.NoError(t, err, "should create the default/test repo")

	// Create a pipeline, to ensure that PPS auth activation works.
	err = c.CreatePipeline("default", "pipeline", "", nil, nil, &pps.ParallelismSpec{}, client.NewPFSInput("default", "test", "*"), "", false)
	require.NoError(t, err, "should create the default/pipeline pipeline")

	// Ensure we can list repos, and that this one shows up.
	ensureTestRepoExists := func(c *client.APIClient) {
		t.Helper()
		repos, err := c.ListRepo()
		require.NoError(t, err, "should be able to list repos")

		got := make(map[string]struct{})
		for _, r := range repos {
			got[r.GetRepo().String()] = struct{}{}
		}
		require.Equal(t, map[string]struct{}{"default/test": {}, "default/pipeline": {}}, got, "should have one repo: default/test")
	}
	ensureTestRepoExists(c)

	// Prepare to activate auth.
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))

	// Enterprise needs a license.
	tu.ActivateLicense(t, c, peerPort)

	// Activate enterprise.
	_, err = env.PachClient.Enterprise.Activate(env.PachClient.Ctx(),
		&enterprise.ActivateRequest{
			LicenseServer: "grpc://localhost:" + peerPort,
			Id:            "localhost",
			Secret:        "localhost",
		})
	require.NoError(t, err, "should activate enterprise")

	// Then setup auth in a single transaction.
	err = env.AuthServer.(interface {
		ActivateAuthEverywhere(context.Context, []authserver.ActivationScope, string) error
	}).ActivateAuthEverywhere(internalauth.AsInternalUser(ctx, authdb.InternalUser), []authserver.ActivationScope{authserver.ActivationScopePFS, authserver.ActivationScopePPS}, tu.RootToken)
	require.NoError(t, err, "should activate auth everywhere")

	// Ensure we can still list repos.
	c.SetAuthToken(tu.RootToken)
	ensureTestRepoExists(c)
}

func TestPreAuthProjects(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
	c := env.PachClient
	project := tu.UniqueString("project")
	require.NoError(t, c.CreateProject(project))

	// activate enterprise + auth
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	tu.ActivateLicense(t, c, peerPort)
	_, err := env.PachClient.Enterprise.Activate(env.PachClient.Ctx(),
		&enterprise.ActivateRequest{
			LicenseServer: "grpc://localhost:" + peerPort,
			Id:            "localhost",
			Secret:        "localhost",
		})
	require.NoError(t, err)
	_, err = c.Activate(c.Ctx(), &auth.ActivateRequest{RootToken: tu.RootToken})
	require.NoError(t, err)
	c.SetAuthToken(tu.RootToken)

	// activate pfs auth
	_, err = c.PfsAPIClient.ActivateAuth(c.Ctx(), &pfs.ActivateAuthRequest{})
	require.NoError(t, err)

	// default project's role binding should be created automatically via PFS auth activation
	_, err = c.ModifyRoleBinding(c.Ctx(), &auth.ModifyRoleBindingRequest{
		Principal: tu.Robot("marvin"),
		Roles:     []string{},
		Resource:  &auth.Resource{Type: auth.ResourceType_PROJECT, Name: pfs.DefaultProjectName},
	})
	require.NoError(t, err)

	// non-default projects also get their role bindings through pfs auth activation
	_, err = c.ModifyRoleBinding(c.Ctx(), &auth.ModifyRoleBindingRequest{
		Principal: tu.Robot("marvin"),
		Roles:     []string{},
		Resource:  &auth.Resource{Type: auth.ResourceType_PROJECT, Name: project},
	})
	require.NoError(t, err)
}

// TestDeleteProject tests whether only owners of a project can delete the project.
func TestDeleteProject(t *testing.T) {
	t.Parallel()

	env := envWithAuth(t)
	c := env.PachClient
	project := tu.UniqueString("project")
	require.NoError(t, c.CreateProject(project))
	alice := tu.AuthenticateClient(t, c, tu.Robot(tu.UniqueString("alice")))

	require.ErrorContains(t, alice.DeleteProject(project, false), "not authorized")
	require.NoError(t, c.DeleteProject(project, false))
}

// TestDeleteRepos tests that when a user requests to delete all repos in a
// project, only those repos he may delete are deleted.
func TestDeleteRepos(t *testing.T) {
	t.Parallel()
	env := envWithAuth(t)
	c := env.PachClient
	alice, bob := tu.Robot(tu.UniqueString("alice")), tu.Robot(tu.UniqueString("bob"))
	aliceClient, bobClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, bob)

	projectName := tu.UniqueString("project")
	require.NoError(t, aliceClient.CreateProject(projectName))

	// create repoA, and check that alice is its owner
	require.NoError(t, aliceClient.CreateRepo(projectName, "repoA"))
	require.Equal(t, tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, projectName, "repoA"))
	// repoB will not be given to bob and should not be deleted
	require.NoError(t, aliceClient.CreateRepo(projectName, "repoB"))

	//////////
	/// alice adds bob to the ACL of repo1 as an owner
	/// alice grants bob projectWriter so that bob can create repos later
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(projectName, "repoA", bob, []string{auth.RepoOwnerRole}))
	require.NoError(t, aliceClient.ModifyProjectRoleBinding(projectName, bob, []string{auth.ProjectWriterRole}))

	// repoC belongs to bob and should be deleted
	require.NoError(t, bobClient.CreateRepo(projectName, "repoC"))
	resp, err := bobClient.PfsAPIClient.DeleteRepos(bobClient.Ctx(), &pfs.DeleteReposRequest{Projects: []*pfs.Project{{Name: projectName}}})
	require.NoError(t, err)
	var deleted = make(map[string]bool)
	for _, repo := range resp.Repos {
		require.Equal(t, repo.Project.GetName(), projectName)
		deleted[repo.Name] = true
	}
	require.False(t, deleted["repoB"])
	for _, name := range []string{"repoA", "repoC"} {
		require.True(t, deleted[name])
	}

	// actually list repos and ensure that repoB is still present
	repoInfos, err := aliceClient.ListRepo()
	require.NoError(t, err)
	seen := make(map[string]bool)
	for _, repoInfo := range repoInfos {
		if repoInfo.Repo.Project.GetName() == projectName {
			seen[repoInfo.Repo.Name] = true
		}
	}
	require.True(t, seen["repoB"])
}

func TestListRepoWithProjectAccessControl(t *testing.T) {
	t.Parallel()
	c := envWithAuth(t).PachClient
	// create users
	admin := tu.AuthenticateClient(t, c, auth.RootUser)
	aliceName, alice := tu.RandomRobot(t, c, "alice")
	bobName, bob := tu.RandomRobot(t, c, "bob")
	// create projects and repos
	project1, project2 := tu.UniqueString("project1-"), tu.UniqueString("project2-")
	repo1, repo2 := "repo1", "repo2"
	require.NoError(t, admin.CreateRepo(pfs.DefaultProjectName, repo1))
	require.NoError(t, admin.CreateRepo(pfs.DefaultProjectName, repo2))
	require.NoError(t, admin.CreateProject(project1))
	require.NoError(t, admin.CreateProject(project2))
	require.NoError(t, admin.CreateRepo(project1, repo1))
	require.NoError(t, admin.CreateRepo(project1, repo2))
	require.NoError(t, admin.CreateRepo(project2, repo1))
	require.NoError(t, admin.CreateRepo(project2, repo2))
	// auth repo resource names need to be project-aware
	defaultRepo1, defaultRepo2 := pfs.DefaultProjectName+"/"+repo1, pfs.DefaultProjectName+"/"+repo2
	project1Repo1, project1Repo2 := project1+"/"+repo1, project1+"/"+repo2
	project2Repo1, project2Repo2 := project2+"/"+repo1, project2+"/"+repo2
	// make bob ProjectViewer for project1
	_, err := admin.ModifyRoleBinding(admin.Ctx(), &auth.ModifyRoleBindingRequest{
		Principal: bobName,
		Roles:     []string{auth.ProjectViewerRole},
		Resource:  &auth.Resource{Type: auth.ResourceType_PROJECT, Name: project1},
	})
	require.NoError(t, err)
	// make alice RepoReader for two repos from separate projects
	_, err = admin.ModifyRoleBinding(admin.Ctx(), &auth.ModifyRoleBindingRequest{
		Principal: aliceName,
		Roles:     []string{auth.RepoReaderRole},
		Resource:  &auth.Resource{Type: auth.ResourceType_REPO, Name: project1Repo1},
	})
	require.NoError(t, err)
	_, err = admin.ModifyRoleBinding(admin.Ctx(), &auth.ModifyRoleBindingRequest{
		Principal: aliceName,
		Roles:     []string{auth.RepoReaderRole},
		Resource:  &auth.Resource{Type: auth.ResourceType_REPO, Name: project2Repo2},
	})
	require.NoError(t, err)

	defaultRepos := []string{defaultRepo1, defaultRepo2}
	project1Repos := []string{project1Repo1, project1Repo2}
	project2Repos := []string{project2Repo1, project2Repo2}
	allRepos := append(defaultRepos, append(project1Repos, project2Repos...)...)

	tests := map[string]struct {
		user          *client.APIClient
		filterBy      []*pfs.Project
		expectedRepos []string
	}{
		"admin no filter":                            {admin, nil, allRepos},
		"admin filters by default project":           {admin, []*pfs.Project{{Name: pfs.DefaultProjectName}}, defaultRepos},
		"admin filters by non-default project":       {admin, []*pfs.Project{{Name: project1}}, project1Repos},
		"non-admin no filter":                        {bob, nil, append(defaultRepos, project1Repos...)},
		"projectViewer filters on their own project": {bob, []*pfs.Project{{Name: project1}}, project1Repos},
		"projectViewer filters on different project": {bob, []*pfs.Project{{Name: project2}}, []string{}},
		"repoReader can see their own repos":         {alice, nil, append(defaultRepos, project1Repo1, project2Repo2)},
		"repoReader filters on project1":             {alice, []*pfs.Project{{Name: project1}}, []string{project1Repo1}},
		"repoReader filers on project2":              {alice, []*pfs.Project{{Name: project2}}, []string{project2Repo2}},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := pctx.WithCancel(tc.user.Ctx())
			defer cancel()
			lrClient, err := tc.user.PfsAPIClient.ListRepo(ctx, &pfs.ListRepoRequest{Type: pfs.UserRepoType, Projects: tc.filterBy})
			require.NoError(t, err)
			repoInfos, err := grpcutil.Collect[*pfs.RepoInfo](lrClient, 1000)
			require.NoError(t, err)
			var repos []string
			for _, repoInfo := range repoInfos {
				repos = append(repos, repoInfo.Repo.AuthResource().Name)
			}
			require.ElementsEqual(t, tc.expectedRepos, repos)
		})
	}
}

func TestListPipelinesWithProjectAccessControl(t *testing.T) {
	createPipeline := func(client *client.APIClient, project string, name string, inputRepo *pps.PFSInput) {
		inputRepo.Name = "in"
		require.NoError(t, client.CreatePipeline(project, name, "", []string{"cp", "/pfs/in/*", "/pfs/out"}, nil, nil, &pps.Input{Pfs: inputRepo}, "", false))

	}
	t.Parallel()
	c := envWithAuth(t).PachClient
	// create uers
	admin := tu.AuthenticateClient(t, c, auth.RootUser)
	aliceName, alice := tu.RandomRobot(t, c, "alice")
	bobName, bob := tu.RandomRobot(t, c, "bob")
	// create projects and pipelines
	project1, project2 := tu.UniqueString("project1-"), tu.UniqueString("project2-")
	require.NoError(t, admin.CreateProject(project1))
	require.NoError(t, admin.CreateProject(project2))
	// create input repo in default project
	inputRepo := tu.UniqueString("input")
	require.NoError(t, admin.CreateRepo(pfs.DefaultProjectName, inputRepo))
	// create 3 pipelines, one in each project
	defaultPipeline, pipeline1, pipeline2 := tu.UniqueString("pipeline-default"), tu.UniqueString("pipeline1-"), tu.UniqueString("pipeline2-")
	createPipeline(admin, pfs.DefaultProjectName, defaultPipeline, &pps.PFSInput{Project: pfs.DefaultProjectName, Repo: inputRepo, Glob: "/*"})
	createPipeline(admin, project1, pipeline1, &pps.PFSInput{Project: pfs.DefaultProjectName, Repo: inputRepo, Glob: "/*"})
	createPipeline(admin, project2, pipeline2, &pps.PFSInput{Project: pfs.DefaultProjectName, Repo: inputRepo, Glob: "/*"})
	// give the appropriate access to Alice
	_, err := admin.ModifyRoleBinding(admin.Ctx(), &auth.ModifyRoleBindingRequest{
		Principal: aliceName,
		Roles:     []string{auth.ProjectViewerRole},
		Resource:  &auth.Resource{Type: auth.ResourceType_PROJECT, Name: project1},
	})
	require.NoError(t, err)
	_, err = admin.ModifyRoleBinding(admin.Ctx(), &auth.ModifyRoleBindingRequest{
		Principal: bobName,
		Roles:     []string{auth.RepoReaderRole},
		Resource:  &auth.Resource{Type: auth.ResourceType_REPO, Name: project2 + "/" + pipeline2},
	})
	require.NoError(t, err)

	tests := map[string]struct {
		c        *client.APIClient
		projects []*pfs.Project
		expected []string
	}{
		"admin list all":           {c: admin, projects: nil, expected: []string{defaultPipeline, pipeline1, pipeline2}},
		"alice list all":           {c: alice, projects: nil, expected: []string{defaultPipeline, pipeline1}},
		"alice filter by default":  {c: alice, projects: []*pfs.Project{{Name: pfs.DefaultProjectName}}, expected: []string{defaultPipeline}},
		"alice filter by project1": {c: alice, projects: []*pfs.Project{{Name: project1}}, expected: []string{pipeline1}},
		"alice filter by project2": {c: alice, projects: []*pfs.Project{{Name: project2}}, expected: nil},
		"bob list all":             {c: bob, projects: nil, expected: []string{defaultPipeline, pipeline2}},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := pctx.WithCancel(tc.c.Ctx())
			defer cancel()
			lpClient, err := tc.c.PpsAPIClient.ListPipeline(ctx, &pps.ListPipelineRequest{Projects: tc.projects})
			require.NoError(t, err)
			pipelineInfos, err := grpcutil.Collect[*pps.PipelineInfo](lpClient, 10)
			require.NoError(t, err)
			var pipelines []string
			for _, pipelineInfo := range pipelineInfos {
				pipelines = append(pipelines, pipelineInfo.Pipeline.Name)
			}
			require.ElementsEqual(t, tc.expected, pipelines)
		})
	}
}

func TestListProjectWithAuth(t *testing.T) {
	t.Parallel()
	c := envWithAuth(t).PachClient

	admin := tu.AuthenticateClient(t, c, auth.RootUser)
	_, alice := tu.RandomRobot(t, c, "alice")

	adminProject := tu.UniqueString("adminProject")
	aliceProject := tu.UniqueString("aliceProject")
	require.NoError(t, admin.CreateProject(adminProject))
	require.NoError(t, alice.CreateProject(aliceProject))

	// For each project, we have a list of expected roles given the client.
	tests := map[string]struct {
		client   *client.APIClient
		expected map[string]*pfs.AuthInfo
	}{
		"admin": {client: admin, expected: map[string]*pfs.AuthInfo{
			pfs.DefaultProjectName: {Roles: []string{auth.ClusterAdminRole, auth.ProjectWriterRole}},
			adminProject:           {Roles: []string{auth.ClusterAdminRole, auth.ProjectOwnerRole}},
			aliceProject:           {Roles: []string{auth.ClusterAdminRole}},
		}},
		"alice": {client: alice, expected: map[string]*pfs.AuthInfo{
			pfs.DefaultProjectName: {Roles: []string{auth.ProjectWriterRole}},
			adminProject:           {Roles: []string{}},
			aliceProject:           {Roles: []string{auth.ProjectOwnerRole}},
		}},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			projectInfos, err := tc.client.ListProject()
			require.NoError(t, err)
			for _, projectInfo := range projectInfos {
				require.ElementsEqual(t, tc.expected[projectInfo.Project.Name].Roles, projectInfo.AuthInfo.Roles)
			}
		})
	}
}
