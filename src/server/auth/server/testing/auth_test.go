package server_test

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"

	"github.com/pachyderm/pachyderm/v2/src/license"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/clientsdk"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

func envWithAuth(t *testing.T) *realenv.RealEnv {
	env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
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
	dataRepo := tu.UniqueString(t.Name())
	require.NoError(t, aliceClient.CreateRepo(dataRepo))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, dataRepo))
	dataCommit := client.NewCommit(dataRepo, "master", "")

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
	require.Equal(t, 1, tu.CommitCnt(t, aliceClient, dataRepo)) // check that no commits were created
	_, err = bobClient.StartCommit(dataRepo, "master")
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.Equal(t, 1, tu.CommitCnt(t, aliceClient, dataRepo)) // check that no commits were created
	// bob can't update the ACL
	err = bobClient.ModifyRepoRoleBinding(dataRepo, tu.Robot("carol"), []string{auth.RepoReaderRole})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	// check that ACL wasn't updated)
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, dataRepo))

	//////////
	/// alice adds bob to the ACL as a reader (alice can modify ACL)
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(dataRepo, bob, []string{auth.RepoReaderRole}))
	// bob can read
	buf.Reset()
	require.NoError(t, bobClient.GetFile(dataCommit, "/file", buf))
	require.Equal(t, "1", buf.String())
	// bob can't write
	err = bobClient.PutFile(dataCommit, "/file", strings.NewReader("2"), client.WithAppendPutFile())
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.Equal(t, 1, tu.CommitCnt(t, aliceClient, dataRepo)) // check that no commits were created
	_, err = bobClient.StartCommit(dataRepo, "master")
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.Equal(t, 1, tu.CommitCnt(t, aliceClient, dataRepo)) // check that no commits were created
	// bob can't update the ACL
	err = bobClient.ModifyRepoRoleBinding(dataRepo, tu.Robot("carol"), []string{auth.RepoReaderRole})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	// check that ACL wasn't updated)
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, bob, auth.RepoReaderRole), tu.GetRepoRoleBinding(t, aliceClient, dataRepo))

	//////////
	/// alice adds bob to the ACL as a writer
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(dataRepo, bob, []string{auth.RepoWriterRole}))
	// bob can read
	buf.Reset()
	require.NoError(t, bobClient.GetFile(dataCommit, "/file", buf))
	require.Equal(t, "1", buf.String())
	// bob can write
	err = bobClient.PutFile(dataCommit, "/file", strings.NewReader("2"), client.WithAppendPutFile())
	require.NoError(t, err)
	require.Equal(t, 2, tu.CommitCnt(t, aliceClient, dataRepo)) // check that a new commit was created
	commit, err := bobClient.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, bobClient.FinishCommit(dataRepo, commit.Branch.Name, commit.ID))
	require.Equal(t, 3, tu.CommitCnt(t, aliceClient, dataRepo)) // check that a new commit was created
	// bob can't update the ACL
	err = bobClient.ModifyRepoRoleBinding(dataRepo, tu.Robot("carol"), []string{auth.RepoReaderRole})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	// check that ACL wasn't updated)
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, bob, auth.RepoWriterRole), tu.GetRepoRoleBinding(t, aliceClient, dataRepo))

	//////////
	/// alice adds bob to the ACL as an owner
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(dataRepo, bob, []string{auth.RepoOwnerRole}))
	// bob can read
	buf.Reset()
	require.NoError(t, bobClient.GetFile(dataCommit, "/file", buf))
	require.Equal(t, "12", buf.String())
	// bob can write
	err = bobClient.PutFile(dataCommit, "/file", strings.NewReader("3"), client.WithAppendPutFile())
	require.NoError(t, err)
	require.Equal(t, 4, tu.CommitCnt(t, aliceClient, dataRepo)) // check that a new commit was created
	commit, err = bobClient.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, bobClient.FinishCommit(dataRepo, commit.Branch.Name, commit.ID))
	require.Equal(t, 5, tu.CommitCnt(t, aliceClient, dataRepo)) // check that a new commit was created
	// bob can update the ACL
	require.NoError(t, bobClient.ModifyRepoRoleBinding(dataRepo, tu.Robot("carol"), []string{auth.RepoReaderRole}))
	// check that ACL was updated)
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, bob, auth.RepoOwnerRole, tu.Robot("carol"), auth.RepoReaderRole),
		tu.GetRepoRoleBinding(t, aliceClient, dataRepo))
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
	dataRepo := tu.UniqueString(t.Name())
	require.NoError(t, aliceClient.CreateRepo(dataRepo))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, dataRepo))
	dataCommit := client.NewCommit(dataRepo, "master", "")

	// Add data to repo (alice can write). Make sure alice can read also.
	commit, err := aliceClient.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	err = aliceClient.PutFile(commit, "/file", strings.NewReader("1"), client.WithAppendPutFile())
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(dataRepo, commit.Branch.Name, commit.ID)) // # commits = 1
	buf := &bytes.Buffer{}
	require.NoError(t, aliceClient.GetFile(dataCommit, "/file", buf))
	require.Equal(t, "1", buf.String())

	//////////
	/// alice adds bob to the ACL as an owner
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(dataRepo, bob, []string{auth.RepoOwnerRole}))
	// bob can read
	buf.Reset()
	require.NoError(t, bobClient.GetFile(dataCommit, "/file", buf))
	require.Equal(t, "1", buf.String())
	// bob can write
	err = bobClient.PutFile(dataCommit, "/file", strings.NewReader("2"), client.WithAppendPutFile())
	require.NoError(t, err)
	require.Equal(t, 2, tu.CommitCnt(t, aliceClient, dataRepo)) // check that a new commit was created
	commit, err = bobClient.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, bobClient.FinishCommit(dataRepo, commit.Branch.Name, commit.ID))
	require.Equal(t, 3, tu.CommitCnt(t, aliceClient, dataRepo)) // check that a new commit was created
	// bob can update the ACL
	require.NoError(t, bobClient.ModifyRepoRoleBinding(dataRepo, tu.Robot("carol"), []string{auth.RepoReaderRole}))
	// check that ACL was updated)
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, bob, auth.RepoOwnerRole, tu.Robot("carol"), auth.RepoReaderRole),
		tu.GetRepoRoleBinding(t, aliceClient, dataRepo))

	// clear carol
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(dataRepo, tu.Robot("carol"), []string{}))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, bob, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, dataRepo))

	//////////
	/// alice adds bob to the ACL as a writer
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(dataRepo, bob, []string{auth.RepoWriterRole}))
	// bob can read
	buf.Reset()
	require.NoError(t, bobClient.GetFile(dataCommit, "/file", buf))
	require.Equal(t, "12", buf.String())
	// bob can write
	err = bobClient.PutFile(dataCommit, "/file", strings.NewReader("3"), client.WithAppendPutFile())
	require.NoError(t, err)
	require.Equal(t, 4, tu.CommitCnt(t, aliceClient, dataRepo)) // check that a new commit was created
	commit, err = bobClient.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, bobClient.FinishCommit(dataRepo, commit.Branch.Name, commit.ID))
	require.Equal(t, 5, tu.CommitCnt(t, aliceClient, dataRepo)) // check that a new commit was created
	// bob can't update the ACL
	err = bobClient.ModifyRepoRoleBinding(dataRepo, tu.Robot("carol"), []string{auth.RepoReaderRole})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	// check that ACL wasn't updated)
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, bob, auth.RepoWriterRole), tu.GetRepoRoleBinding(t, aliceClient, dataRepo))

	//////////
	/// alice adds bob to the ACL as a reader (alice can modify ACL)
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(dataRepo, bob, []string{auth.RepoReaderRole}))
	// bob can read
	buf.Reset()
	require.NoError(t, bobClient.GetFile(dataCommit, "/file", buf))
	require.Equal(t, "123", buf.String())
	// bob can't write
	err = bobClient.PutFile(dataCommit, "/file", strings.NewReader("4"), client.WithAppendPutFile())
	require.YesError(t, err)
	require.Equal(t, 5, tu.CommitCnt(t, aliceClient, dataRepo)) // check that a new commit was created
	_, err = bobClient.StartCommit(dataRepo, "master")
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.Equal(t, 5, tu.CommitCnt(t, aliceClient, dataRepo)) // check that no commits were created
	// bob can't update the ACL
	err = bobClient.ModifyRepoRoleBinding(dataRepo, tu.Robot("carol"), []string{auth.RepoReaderRole})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	// check that ACL wasn't updated)
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, bob, auth.RepoReaderRole), tu.GetRepoRoleBinding(t, aliceClient, dataRepo))

	//////////
	/// alice revokes all of bob's privileges
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(dataRepo, bob, []string{}))
	// bob can't read
	err = bobClient.GetFile(dataCommit, "/file", buf)
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	// bob can't write
	err = bobClient.PutFile(dataCommit, "/file", strings.NewReader("4"), client.WithAppendPutFile())
	require.YesError(t, err)
	require.Equal(t, 5, tu.CommitCnt(t, aliceClient, dataRepo)) // check that a new commit was created
	_, err = bobClient.StartCommit(dataRepo, "master")
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.Equal(t, 5, tu.CommitCnt(t, aliceClient, dataRepo)) // check that no commits were created
	// bob can't update the ACL
	err = bobClient.ModifyRepoRoleBinding(dataRepo, tu.Robot("carol"), []string{auth.RepoReaderRole})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	// check that ACL wasn't updated)
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, dataRepo))
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
	require.NoError(t, aliceClient.CreateRepo(dataRepo))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, dataRepo))
	dataCommit := client.NewCommit(dataRepo, "master", "")

	// Add data to repo (alice can write). Make sure alice can read also.
	err := aliceClient.PutFile(dataCommit, "/file", strings.NewReader("1"))
	require.NoError(t, err)
	buf := &bytes.Buffer{}
	require.NoError(t, aliceClient.GetFile(dataCommit, "/file", buf))
	require.Equal(t, "1", buf.String())

	/// alice adds bob to the ACL as a reader (alice can modify ACL)
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(dataRepo, bob, []string{auth.RepoReaderRole}))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, bob, auth.RepoReaderRole), tu.GetRepoRoleBinding(t, aliceClient, dataRepo))
	// bob can read
	buf.Reset()
	require.NoError(t, bobClient.GetFile(dataCommit, "/file", buf))
	require.Equal(t, "1", buf.String())

	/// alice updates the repo
	description := "This request updates the description to force a write"
	_, err = aliceClient.PfsAPIClient.CreateRepo(aliceClient.Ctx(), &pfs.CreateRepoRequest{
		Repo:        client.NewRepo(dataRepo),
		Description: description,
		Update:      true,
	})
	require.NoError(t, err)
	repoInfo, err := aliceClient.InspectRepo(dataRepo)
	require.NoError(t, err)
	require.Equal(t, description, repoInfo.Description)
	// buildBindings haven't changed
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, bob, auth.RepoReaderRole), tu.GetRepoRoleBinding(t, aliceClient, dataRepo))
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
		Repo:   client.NewRepo(dataRepo),
		Update: true,
	})
	require.NoError(t, err)
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, dataRepo))
	dataCommit := client.NewCommit(dataRepo, "master", "")
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
		return args.client.CreatePipeline(
			args.name,
			"", // default image: DefaultUserImage
			[]string{"bash"},
			[]string{"cp /pfs/*/* /pfs/out/"},
			&pps.ParallelismSpec{Constant: 1},
			client.NewPFSInput(args.repo, "/*"),
			"", // default output branch: master
			args.update,
		)
	}
	alice, bob := tu.Robot(tu.UniqueString("alice")), tu.Robot(tu.UniqueString("bob"))
	aliceClient, bobClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, bob)

	// create repo, and check that alice is the owner of the new repo
	dataRepo := tu.UniqueString(t.Name())
	require.NoError(t, aliceClient.CreateRepo(dataRepo))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, dataRepo))
	dataCommit := client.NewCommit(dataRepo, "master", "")

	// alice can create a pipeline (she owns the input repo)
	pipeline := tu.UniqueString("alice-pipeline")
	require.NoError(t, createPipeline(createArgs{
		client: aliceClient,
		name:   pipeline,
		repo:   dataRepo,
	}))
	require.OneOfEquals(t, pipeline, tu.PipelineNames(t, aliceClient))
	// check that alice owns the output repo too)
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Pl(pipeline), auth.RepoWriterRole), tu.GetRepoRoleBinding(t, aliceClient, pipeline))

	// Make sure alice's pipeline runs successfully
	err := aliceClient.PutFile(dataCommit, tu.UniqueString("/file"),
		strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := aliceClient.WaitCommit(pipeline, "master", "")
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
	require.NoneEquals(t, badPipeline, tu.PipelineNames(t, aliceClient))

	// alice adds bob as a reader of the input repo
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(dataRepo, bob, []string{auth.RepoReaderRole}))

	// now bob can create a pipeline
	goodPipeline := tu.UniqueString("bob-good")
	require.NoError(t, createPipeline(createArgs{
		client: bobClient,
		name:   goodPipeline,
		repo:   dataRepo,
	}))
	require.OneOfEquals(t, goodPipeline, tu.PipelineNames(t, aliceClient))
	// check that bob owns the output repo too)
	require.Equal(t,
		tu.BuildBindings(bob, auth.RepoOwnerRole, tu.Pl(goodPipeline), auth.RepoWriterRole), tu.GetRepoRoleBinding(t, bobClient, goodPipeline))

	// Make sure bob's pipeline runs successfully
	err = aliceClient.PutFile(dataCommit, tu.UniqueString("/file"),
		strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoErrorWithinT(t, 4*time.Minute, func() error {
		_, err := bobClient.WaitCommit(goodPipeline, "master", "")
		return err
	})

	// bob can't update alice's pipeline
	infoBefore, err := aliceClient.InspectPipeline(pipeline, true)
	require.NoError(t, err)
	err = createPipeline(createArgs{
		client: bobClient,
		name:   pipeline,
		repo:   dataRepo,
		update: true,
	})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	infoAfter, err := aliceClient.InspectPipeline(pipeline, true)
	require.NoError(t, err)
	require.Equal(t, infoBefore.Version, infoAfter.Version)

	// alice adds bob as a writer of the output repo, and removes him as a reader
	// of the input repo
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pipeline, bob, []string{auth.RepoWriterRole}))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, bob, auth.RepoWriterRole, tu.Pl(pipeline), auth.RepoWriterRole),
		tu.GetRepoRoleBinding(t, aliceClient, pipeline))

	require.NoError(t, aliceClient.ModifyRepoRoleBinding(dataRepo, bob, []string{}))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Pl(pipeline), auth.RepoReaderRole, tu.Pl(goodPipeline), auth.RepoReaderRole),
		tu.GetRepoRoleBinding(t, aliceClient, dataRepo))

	// bob still can't update alice's pipeline
	infoBefore, err = aliceClient.InspectPipeline(pipeline, true)
	require.NoError(t, err)
	err = createPipeline(createArgs{
		client: bobClient,
		name:   pipeline,
		repo:   dataRepo,
		update: true,
	})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	infoAfter, err = aliceClient.InspectPipeline(pipeline, true)
	require.NoError(t, err)
	require.Equal(t, infoBefore.Version, infoAfter.Version)

	// alice re-adds bob as a reader of the input repo
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(dataRepo, bob, []string{auth.RepoReaderRole}))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, bob, auth.RepoReaderRole, tu.Pl(pipeline), auth.RepoReaderRole, tu.Pl(goodPipeline), auth.RepoReaderRole),
		tu.GetRepoRoleBinding(t, aliceClient, dataRepo))

	// now bob can update alice's pipeline
	infoBefore, err = aliceClient.InspectPipeline(pipeline, true)
	require.NoError(t, err)
	err = createPipeline(createArgs{
		client: bobClient,
		name:   pipeline,
		repo:   dataRepo,
		update: true,
	})
	require.NoError(t, err)
	infoAfter, err = aliceClient.InspectPipeline(pipeline, true)
	require.NoError(t, err)
	require.NotEqual(t, infoBefore.Version, infoAfter.Version)

	// Make sure that we don't get an auth token returned by the inspect
	require.Equal(t, "", infoAfter.AuthToken)
	infoAfter, err = aliceClient.InspectPipeline(pipeline, false)
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
		_, err := bobClient.WaitCommit(pipeline, "master", "")
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
		return args.client.CreatePipeline(
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
	require.NoError(t, aliceClient.CreateRepo(dataRepo1))
	require.NoError(t, aliceClient.CreateRepo(dataRepo2))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, dataRepo1))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, dataRepo2))

	// alice can create a cross-pipeline with both inputs
	aliceCrossPipeline := tu.UniqueString("alice-cross")
	require.NoError(t, createPipeline(createArgs{
		client: aliceClient,
		name:   aliceCrossPipeline,
		input: client.NewCrossInput(
			client.NewPFSInput(dataRepo1, "/*"),
			client.NewPFSInput(dataRepo2, "/*"),
		),
	}))
	require.OneOfEquals(t, aliceCrossPipeline, tu.PipelineNames(t, aliceClient))
	// check that alice owns the output repo too)
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Pl(aliceCrossPipeline), auth.RepoWriterRole), tu.GetRepoRoleBinding(t, aliceClient, aliceCrossPipeline))

	// alice can create a union-pipeline with both inputs
	aliceUnionPipeline := tu.UniqueString("alice-union")
	require.NoError(t, createPipeline(createArgs{
		client: aliceClient,
		name:   aliceUnionPipeline,
		input: client.NewUnionInput(
			client.NewPFSInput(dataRepo1, "/*"),
			client.NewPFSInput(dataRepo2, "/*"),
		),
	}))
	require.OneOfEquals(t, aliceUnionPipeline, tu.PipelineNames(t, aliceClient))
	// check that alice owns the output repo too)
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Pl(aliceUnionPipeline), auth.RepoWriterRole), tu.GetRepoRoleBinding(t, aliceClient, aliceUnionPipeline))

	// alice adds bob as a reader of one of the input repos, but not the other
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(dataRepo1, bob, []string{auth.RepoReaderRole}))

	// bob cannot create a cross-pipeline with both inputs
	bobCrossPipeline := tu.UniqueString("bob-cross")
	err := createPipeline(createArgs{
		client: bobClient,
		name:   bobCrossPipeline,
		input: client.NewCrossInput(
			client.NewPFSInput(dataRepo1, "/*"),
			client.NewPFSInput(dataRepo2, "/*"),
		),
	})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.NoneEquals(t, bobCrossPipeline, tu.PipelineNames(t, aliceClient))

	// bob cannot create a union-pipeline with both inputs
	bobUnionPipeline := tu.UniqueString("bob-union")
	err = createPipeline(createArgs{
		client: bobClient,
		name:   bobUnionPipeline,
		input: client.NewUnionInput(
			client.NewPFSInput(dataRepo1, "/*"),
			client.NewPFSInput(dataRepo2, "/*"),
		),
	})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	require.NoneEquals(t, bobUnionPipeline, tu.PipelineNames(t, aliceClient))

	// alice adds bob as a writer of her pipeline's output
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(aliceCrossPipeline, bob, []string{auth.RepoWriterRole}))

	// bob can update alice's pipeline if he removes one of the inputs
	infoBefore, err := aliceClient.InspectPipeline(aliceCrossPipeline, true)
	require.NoError(t, err)
	require.NoError(t, createPipeline(createArgs{
		client: bobClient,
		name:   aliceCrossPipeline,
		input: client.NewCrossInput(
			// This cross input deliberately only has one element, to make sure it's
			// not simply rejected for having a cross input
			client.NewPFSInput(dataRepo1, "/*"),
		),
		update: true,
	}))
	infoAfter, err := aliceClient.InspectPipeline(aliceCrossPipeline, true)
	require.NoError(t, err)
	require.NotEqual(t, infoBefore.Version, infoAfter.Version)

	// bob cannot update alice's to put the second input back
	infoBefore, err = aliceClient.InspectPipeline(aliceCrossPipeline, true)
	require.NoError(t, err)
	err = createPipeline(createArgs{
		client: bobClient,
		name:   aliceCrossPipeline,
		input: client.NewCrossInput(
			client.NewPFSInput(dataRepo1, "/*"),
			client.NewPFSInput(dataRepo2, "/*"),
		),
		update: true,
	})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	infoAfter, err = aliceClient.InspectPipeline(aliceCrossPipeline, true)
	require.NoError(t, err)
	require.Equal(t, infoBefore.Version, infoAfter.Version)

	// alice adds bob as a reader of the second input
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(dataRepo2, bob, []string{auth.RepoReaderRole}))

	// bob can now update alice's to put the second input back
	infoBefore, err = aliceClient.InspectPipeline(aliceCrossPipeline, true)
	require.NoError(t, err)
	require.NoError(t, createPipeline(createArgs{
		client: bobClient,
		name:   aliceCrossPipeline,
		input: client.NewCrossInput(
			client.NewPFSInput(dataRepo1, "/*"),
			client.NewPFSInput(dataRepo2, "/*"),
		),
		update: true,
	}))
	infoAfter, err = aliceClient.InspectPipeline(aliceCrossPipeline, true)
	require.NoError(t, err)
	require.NotEqual(t, infoBefore.Version, infoAfter.Version)

	// bob can create a cross-pipeline with both inputs
	require.NoError(t, createPipeline(createArgs{
		client: bobClient,
		name:   bobCrossPipeline,
		input: client.NewCrossInput(
			client.NewPFSInput(dataRepo1, "/*"),
			client.NewPFSInput(dataRepo2, "/*"),
		),
	}))
	require.OneOfEquals(t, bobCrossPipeline, tu.PipelineNames(t, aliceClient))

	// bob can create a union-pipeline with both inputs
	require.NoError(t, createPipeline(createArgs{
		client: bobClient,
		name:   bobUnionPipeline,
		input: client.NewUnionInput(
			client.NewPFSInput(dataRepo1, "/*"),
			client.NewPFSInput(dataRepo2, "/*"),
		),
	}))
	require.OneOfEquals(t, bobUnionPipeline, tu.PipelineNames(t, aliceClient))

}

func TestStopAndDeletePipeline(t *testing.T) {
	t.Parallel()
	env := envWithAuth(t)
	c := env.PachClient
	alice, bob := tu.Robot(tu.UniqueString("alice")), tu.Robot(tu.UniqueString("bob"))
	aliceClient, bobClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, bob)

	// alice creates a repo
	repo := tu.UniqueString(t.Name())
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.Equal(t, tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, repo))

	// alice creates a pipeline
	pipeline := tu.UniqueString("alice-pipeline")
	require.NoError(t, aliceClient.CreatePipeline(
		pipeline,
		"", // default image: DefaultUserImage
		[]string{"bash"},
		[]string{"cp /pfs/*/* /pfs/out/"},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(repo, "/*"),
		"", // default output branch: master
		false,
	))
	// Make sure the input and output repos have non-empty ACLs
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Pl(pipeline), auth.RepoReaderRole), tu.GetRepoRoleBinding(t, aliceClient, repo))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Pl(pipeline), auth.RepoWriterRole), tu.GetRepoRoleBinding(t, aliceClient, pipeline))

	// alice stops the pipeline (owner of the input and output repos can stop)
	require.NoError(t, aliceClient.StopPipeline(pipeline))

	// Make sure the remaining input and output repos *still* have non-empty ACLs
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Pl(pipeline), auth.RepoReaderRole), tu.GetRepoRoleBinding(t, aliceClient, repo))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Pl(pipeline), auth.RepoWriterRole), tu.GetRepoRoleBinding(t, aliceClient, pipeline))

	// alice deletes the pipeline (owner of the input and output repos can delete)
	require.NoError(t, aliceClient.DeletePipeline(pipeline, false))
	require.Nil(t, tu.GetRepoRoleBinding(t, aliceClient, pipeline).Entries)

	// alice deletes the input repo (make sure the input repo's ACL is gone)
	require.NoError(t, aliceClient.DeleteRepo(repo, false))
	require.Nil(t, tu.GetRepoRoleBinding(t, aliceClient, repo).Entries)

	// alice creates another repo
	repo = tu.UniqueString(t.Name())
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.Equal(t, tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, repo))

	// alice creates another pipeline
	pipeline = tu.UniqueString("alice-pipeline")
	require.NoError(t, aliceClient.CreatePipeline(
		pipeline,
		"", // default image: DefaultUserImage
		[]string{"bash"},
		[]string{"cp /pfs/*/* /pfs/out/"},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(repo, "/*"),
		"", // default output branch: master
		false,
	))

	// bob can't stop or delete alice's pipeline
	err := bobClient.StopPipeline(pipeline)
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	err = bobClient.DeletePipeline(pipeline, false)
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())

	// alice adds bob as a reader of the input repo
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(repo, bob, []string{auth.RepoReaderRole}))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, bob, auth.RepoReaderRole, tu.Pl(pipeline), auth.RepoReaderRole),
		tu.GetRepoRoleBinding(t, aliceClient, repo))

	// bob still can't stop or delete alice's pipeline
	err = bobClient.StopPipeline(pipeline)
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	err = bobClient.DeletePipeline(pipeline, false)
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())

	// alice removes bob as a reader of the input repo and adds bob as a writer of
	// the output repo
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(repo, bob, []string{}))

	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Pl(pipeline), auth.RepoReaderRole), tu.GetRepoRoleBinding(t, aliceClient, repo))
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pipeline, bob, []string{auth.RepoWriterRole}))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, bob, auth.RepoWriterRole, tu.Pl(pipeline), auth.RepoWriterRole),
		tu.GetRepoRoleBinding(t, aliceClient, pipeline))

	// bob can now start and stop the pipeline, but can't delete it
	require.NoError(t, bobClient.StopPipeline(pipeline))
	require.NoError(t, bobClient.StartPipeline(pipeline))
	err = bobClient.DeletePipeline(pipeline, false)
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
	// alice re-adds bob as a reader of the input repo
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(repo, bob, []string{auth.RepoReaderRole}))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, bob, auth.RepoReaderRole, tu.Pl(pipeline), auth.RepoReaderRole),
		tu.GetRepoRoleBinding(t, aliceClient, repo))

	// no change to bob's capabilities
	require.NoError(t, bobClient.StopPipeline(pipeline))
	require.NoError(t, bobClient.StartPipeline(pipeline))
	err = bobClient.DeletePipeline(pipeline, false)
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())

	// alice adds bob as an owner of the output repo
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pipeline, bob, []string{auth.RepoOwnerRole}))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, bob, auth.RepoOwnerRole, tu.Pl(pipeline), auth.RepoWriterRole),
		tu.GetRepoRoleBinding(t, aliceClient, pipeline))

	// finally bob can delete alice's pipeline
	err = bobClient.DeletePipeline(pipeline, false)
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
	commit := client.NewCommit(repo, "master", "")
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.Equal(t, tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, repo))
	err := aliceClient.PutFile(commit, "/file", strings.NewReader("test"))
	require.NoError(t, err)

	// alice creates a pipeline
	pipeline := tu.UniqueString("alice-pipeline")
	require.NoError(t, aliceClient.CreatePipeline(
		pipeline,
		"", // default image: DefaultUserImage
		[]string{"bash"},
		[]string{"sleep 600"},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(repo, "/*"),
		"", // default output branch: master
		false,
	))
	// Make sure the input and output repos have non-empty ACLs
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Pl(pipeline), auth.RepoReaderRole), tu.GetRepoRoleBinding(t, aliceClient, repo))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Pl(pipeline), auth.RepoWriterRole), tu.GetRepoRoleBinding(t, aliceClient, pipeline))

	// Stop the first job in 'pipeline'
	var jobID string
	require.NoErrorWithinTRetry(t, 30*time.Second, func() error {
		jobs, err := aliceClient.ListJob(pipeline, nil /*inputs*/, -1 /*history*/, true /* full */)
		if err != nil {
			return err
		}
		if len(jobs) != 1 {
			return errors.Errorf("expected one job but got %d", len(jobs))
		}
		jobID = jobs[0].Job.ID
		return nil
	})

	require.NoError(t, aliceClient.StopJob(pipeline, jobID))
	require.NoErrorWithinTRetry(t, 30*time.Second, func() error {
		ji, err := aliceClient.InspectJob(pipeline, jobID, false)
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
	repoWriter := tu.UniqueString(t.Name())
	require.NoError(t, aliceClient.CreateRepo(repoWriter))
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(repoWriter, bob, []string{auth.RepoWriterRole}))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, bob, auth.RepoWriterRole), tu.GetRepoRoleBinding(t, aliceClient, repoWriter))

	// alice creates a repo and makes Bob a reader
	repoReader := tu.UniqueString(t.Name())
	require.NoError(t, aliceClient.CreateRepo(repoReader))
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(repoReader, bob, []string{auth.RepoReaderRole}))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, bob, auth.RepoReaderRole), tu.GetRepoRoleBinding(t, aliceClient, repoReader))

	// alice creates a repo and gives Bob no access privileges
	repoNone := tu.UniqueString(t.Name())
	require.NoError(t, aliceClient.CreateRepo(repoNone))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, repoNone))

	// put a file in the repo Bob can't access - we need to be able to get the size of the commits
	err := aliceClient.PutFile(client.NewCommit(repoNone, "master", ""), "/test", strings.NewReader("test"))
	require.NoError(t, err)

	// bob creates a repo
	repoOwner := tu.UniqueString(t.Name())
	require.NoError(t, bobClient.CreateRepo(repoOwner))
	require.Equal(t, tu.BuildBindings(bob, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, bobClient, repoOwner))

	// Bob calls ListRepo, and the response must indicate the correct access scope
	// for each repo (because other tests have run, we may see repos besides the
	// above. Bob's access to those should be NONE
	lrClient, err := bobClient.PfsAPIClient.ListRepo(bobClient.Ctx(),
		&pfs.ListRepoRequest{})
	require.NoError(t, err)
	repoInfos, err := clientsdk.ListRepoInfo(lrClient)
	require.NoError(t, err)
	expectedPermissions := map[string][]auth.Permission{
		repoOwner: []auth.Permission{
			auth.Permission_REPO_READ,
			auth.Permission_REPO_WRITE,
			auth.Permission_REPO_MODIFY_BINDINGS,
			auth.Permission_REPO_DELETE,
			auth.Permission_REPO_INSPECT_COMMIT,
			auth.Permission_REPO_LIST_COMMIT,
			auth.Permission_REPO_DELETE_COMMIT,
			auth.Permission_REPO_CREATE_BRANCH,
			auth.Permission_REPO_LIST_BRANCH,
			auth.Permission_REPO_DELETE_BRANCH,
			auth.Permission_REPO_LIST_FILE,
			auth.Permission_REPO_ADD_PIPELINE_READER,
			auth.Permission_REPO_REMOVE_PIPELINE_READER,
			auth.Permission_REPO_ADD_PIPELINE_WRITER,
			auth.Permission_REPO_INSPECT_FILE,
			auth.Permission_PIPELINE_LIST_JOB,
		},
		repoWriter: []auth.Permission{
			auth.Permission_REPO_READ,
			auth.Permission_REPO_WRITE,
			auth.Permission_REPO_INSPECT_COMMIT,
			auth.Permission_REPO_LIST_COMMIT,
			auth.Permission_REPO_DELETE_COMMIT,
			auth.Permission_REPO_CREATE_BRANCH,
			auth.Permission_REPO_LIST_BRANCH,
			auth.Permission_REPO_DELETE_BRANCH,
			auth.Permission_REPO_LIST_FILE,
			auth.Permission_REPO_ADD_PIPELINE_READER,
			auth.Permission_REPO_REMOVE_PIPELINE_READER,
			auth.Permission_REPO_ADD_PIPELINE_WRITER,
			auth.Permission_REPO_INSPECT_FILE,
			auth.Permission_PIPELINE_LIST_JOB,
		},
		repoReader: []auth.Permission{
			auth.Permission_REPO_READ,
			auth.Permission_REPO_INSPECT_COMMIT,
			auth.Permission_REPO_LIST_COMMIT,
			auth.Permission_REPO_LIST_BRANCH,
			auth.Permission_REPO_LIST_FILE,
			auth.Permission_REPO_ADD_PIPELINE_READER,
			auth.Permission_REPO_REMOVE_PIPELINE_READER,
			auth.Permission_REPO_INSPECT_FILE,
			auth.Permission_PIPELINE_LIST_JOB,
		},
	}
	for _, info := range repoInfos {
		require.ElementsEqual(t, expectedPermissions[info.Repo.Name], info.AuthInfo.Permissions)
	}

	for _, name := range []string{repoOwner, repoWriter, repoReader, repoNone} {
		inspectResp, err := bobClient.PfsAPIClient.InspectRepo(bobClient.Ctx(),
			&pfs.InspectRepoRequest{
				Repo: client.NewRepo(name),
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
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, repo))

	// bob calls SetScope(bob, OWNER) on alice's repo. This should fail
	err := bobClient.ModifyRepoRoleBinding(repo, bob, []string{auth.RepoOwnerRole})
	require.YesError(t, err)
	// make sure ACL wasn't updated
	require.Equal(t, tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, repo))
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
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, repo))

	// Anon (non-logged-in user) calls ListRepo, and must receive an error
	c, err := anonClient.PfsAPIClient.ListRepo(anonClient.Ctx(),
		&pfs.ListRepoRequest{})
	require.NoError(t, err)
	_, err = clientsdk.ListRepoInfo(c)
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
	require.NoError(t, aliceClient.CreateRepo(repo))

	// bob calls ListRepo, but has NONE access to all repos
	infos, err := bobClient.ListRepo()
	require.NoError(t, err)
	for _, info := range infos {
		require.Nil(t, info.AuthInfo.Permissions)
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
	require.NoError(t, aliceClient.CreateRepo(repo))

	// bob creates the same repo, and should get an error to the effect that the
	// repo already exists (rather than "access denied")
	err := bobClient.CreateRepo(repo)
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
	err := anonClient.CreateRepo(repo)
	require.YesError(t, err)
	require.Matches(t, "no authentication token", err.Error())
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
	require.NoError(t, aliceClient.CreateRepo(inputRepo))
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(inputRepo, bob, []string{auth.RepoReaderRole}))
	pipeline := tu.UniqueString("pipeline")
	require.NoError(t, aliceClient.CreateRepo(pipeline))

	// bob creates a pipeline, and should get an "access denied" error
	err := bobClient.CreatePipeline(
		pipeline,
		"", // default image: DefaultUserImage
		[]string{"bash"},
		[]string{"cp /pfs/*/* /pfs/out/"},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(inputRepo, "/*"),
		"",    // default output branch: master
		false, // Don't update -- we want an error
	)
	require.YesError(t, err)
	require.Matches(t, "already exists", err.Error())

	// alice gives bob writer scope on pipeline output repo, but nothing changes
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pipeline, bob, []string{auth.RepoWriterRole}))
	err = bobClient.CreatePipeline(
		pipeline,
		"", // default image: DefaultUserImage
		[]string{"bash"},
		[]string{"cp /pfs/*/* /pfs/out/"},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(inputRepo, "/*"),
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
	require.NoError(t, aliceClient.CreateRepo(repo))

	// alice is authorized as `OWNER`
	resp, err := aliceClient.Authorize(aliceClient.Ctx(), &auth.AuthorizeRequest{
		Resource: &auth.Resource{Type: auth.ResourceType_REPO, Name: repo},
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
		Resource: &auth.Resource{Type: auth.ResourceType_REPO, Name: repo},
		Permissions: []auth.Permission{
			auth.Permission_REPO_READ,
		},
	})
	require.NoError(t, err)
	require.False(t, resp.Authorized)

	// alice grants everybody WRITER access
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(repo, "allClusterUsers", []string{auth.RepoWriterRole}))

	// alice is still authorized as `OWNER`
	resp, err = aliceClient.Authorize(aliceClient.Ctx(), &auth.AuthorizeRequest{
		Resource: &auth.Resource{Type: auth.ResourceType_REPO, Name: repo},
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
		Resource: &auth.Resource{Type: auth.ResourceType_REPO, Name: repo},
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
	require.NoError(t, adminClient.CreateRepo(adminRepo))

	// alice creates a repo
	aliceRepo := tu.UniqueString(t.Name())
	require.NoError(t, aliceClient.CreateRepo(aliceRepo))

	// alice calls DeleteAll. It passes, but only deletes the repos she was authorized to delete
	_, err := aliceClient.PfsAPIClient.DeleteAll(aliceClient.Ctx(), &types.Empty{})
	require.NoError(t, err)

	listResp, err := aliceClient.ListRepo()
	require.NoError(t, err)

	require.Equal(t, 1, len(listResp))
	require.Equal(t, adminRepo, listResp[0].Repo.Name)
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
	require.NoError(t, aliceClient.CreateRepo(repo))

	// alice creates a pipeline
	pipeline := tu.UniqueString("alice-pipeline")
	require.NoError(t, aliceClient.CreatePipeline(
		pipeline,
		"", // default image: DefaultUserImage
		[]string{"bash"},
		[]string{"ls /pfs/*/*; cp /pfs/*/* /pfs/out/"},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(repo, "/*"),
		"", // default output branch: master
		false,
	))

	// alice commits to the input repos, and the pipeline runs successfully
	var err error
	err = aliceClient.PutFile(client.NewCommit(repo, "master", ""), "/file", strings.NewReader("test"))
	require.NoError(t, err)
	require.NoErrorWithinT(t, 4*time.Minute, func() error {
		_, err := aliceClient.WaitCommit(pipeline, "master", "")
		return err
	})
	jobs, err := aliceClient.ListJob(pipeline, nil /*inputs*/, -1 /*history*/, true)
	require.NoError(t, err)
	require.Equal(t, 2, len(jobs))
	jobID := jobs[0].Job.ID

	// bob cannot call ListJob on 'pipeline'
	_, err = bobClient.ListJob(pipeline, nil, -1 /*history*/, true)
	require.YesError(t, err)
	require.True(t, auth.IsErrNotAuthorized(err), err.Error())
	// bob can call blank ListJob, but gets no results
	jobs, err = bobClient.ListJob("", nil, -1 /*history*/, true)
	require.NoError(t, err)
	require.Equal(t, 0, len(jobs))

	// alice adds bob to repo, but bob still can't call ListJob on 'pipeline' or
	// get any output
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(repo, bob, []string{auth.RepoReaderRole}))
	_, err = bobClient.ListJob(pipeline, nil, -1 /*history*/, true)
	require.YesError(t, err)
	require.True(t, auth.IsErrNotAuthorized(err), err.Error())
	jobs, err = bobClient.ListJob("", nil, -1 /*history*/, true)
	require.NoError(t, err)
	require.Equal(t, 0, len(jobs))

	// alice removes bob from repo and adds bob to 'pipeline', and now bob can
	// call listJob on 'pipeline', and gets results back from blank listJob
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(repo, bob, []string{}))
	err = aliceClient.ModifyRepoRoleBinding(pipeline, bob, []string{auth.RepoReaderRole})
	require.NoError(t, err)
	jobs, err = bobClient.ListJob(pipeline, nil, -1 /*history*/, true)
	require.NoError(t, err)
	require.Equal(t, 2, len(jobs))
	require.Equal(t, jobID, jobs[0].Job.ID)
	jobs, err = bobClient.ListJob("", nil, -1 /*history*/, true)
	require.NoError(t, err)
	require.Equal(t, 2, len(jobs))
	require.Equal(t, jobID, jobs[0].Job.ID)
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
			Input:           client.NewPFSInput(repo, "/*"),
		})
	require.NoError(t, err)

	// alice commits to the input repo, and the pipeline runs successfully
	err = aliceClient.PutFile(client.NewCommit(repo, "master", ""), "/file", strings.NewReader("test"))
	require.NoError(t, err)
	require.NoErrorWithinT(t, 2*time.Minute, func() error {
		_, err := aliceClient.WaitCommit(pipeline, "master", "")
		return err
	})
	jobs, err := aliceClient.ListJob(pipeline, nil /*inputs*/, -1 /*history*/, true)
	require.NoError(t, err)
	require.Equal(t, 2, len(jobs))
	jobID := jobs[0].Job.ID

	// ListDatum seems like it may return inconsistent results, so sleep until
	// the /stats branch is written
	// TODO(msteffen): verify if this is true, and if so, why
	time.Sleep(5 * time.Second)
	dis, err := aliceClient.ListDatumAll(pipeline, jobID)
	require.NoError(t, err)
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		for _, di := range dis {
			if _, err := aliceClient.InspectDatum(pipeline, jobID, di.Datum.ID); err != nil {
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
		require.NoError(t, aliceClient.CreateRepo(repo[i]))
		require.Equal(t, tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, repo[i]))

		// Commit to repo
		err := aliceClient.PutFile(
			client.NewCommit(repo[i], "master", ""), "/"+repo[i], strings.NewReader(repo[i]))
		require.NoError(t, err)
	}

	// alice creates a pipeline
	pipeline := tu.UniqueString("alice-pipeline")
	require.NoError(t, aliceClient.CreatePipeline(
		pipeline,
		"", // default image: DefaultUserImage
		[]string{"bash"},
		[]string{"cp /pfs/*/* /pfs/out/"},
		&pps.ParallelismSpec{Constant: 1},
		client.NewUnionInput(
			client.NewPFSInput(repo[0], "/*"),
			client.NewPFSInput(repo[1], "/*"),
		),
		"", // default output branch: master
		false,
	))
	// Make sure the input and output repos have appropriate ACLs
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Pl(pipeline), auth.RepoReaderRole), tu.GetRepoRoleBinding(t, aliceClient, repo[0]))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Pl(pipeline), auth.RepoReaderRole), tu.GetRepoRoleBinding(t, aliceClient, repo[1]))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Pl(pipeline), auth.RepoWriterRole), tu.GetRepoRoleBinding(t, aliceClient, pipeline))
	// repo[2] is not on pipeline -- doesn't include 'pipeline'
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, repo[2]))

	// make sure the pipeline runs
	require.NoErrorWithinT(t, 4*time.Minute, func() error {
		_, err := aliceClient.WaitCommit(pipeline, "master", "")
		return err
	})

	// alice updates the pipeline to replace repo[0] with repo[2]
	require.NoError(t, aliceClient.CreatePipeline(
		pipeline,
		"", // default image: DefaultUserImage
		[]string{"bash"},
		[]string{"cp /pfs/*/* /pfs/out/"},
		&pps.ParallelismSpec{Constant: 1},
		client.NewUnionInput(
			client.NewPFSInput(repo[1], "/*"),
			client.NewPFSInput(repo[2], "/*"),
		),
		"", // default output branch: master
		true,
	))
	// Make sure the input and output repos have appropriate ACLs
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Pl(pipeline), auth.RepoReaderRole), tu.GetRepoRoleBinding(t, aliceClient, repo[1]))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Pl(pipeline), auth.RepoReaderRole), tu.GetRepoRoleBinding(t, aliceClient, repo[2]))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Pl(pipeline), auth.RepoWriterRole), tu.GetRepoRoleBinding(t, aliceClient, pipeline))
	// repo[0] is not on pipeline -- doesn't include 'pipeline'
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, repo[0]))

	// make sure the pipeline still runs
	require.NoErrorWithinT(t, 2*time.Minute, func() error {
		_, err := aliceClient.WaitCommit(pipeline, "master", "")
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
				&auth.ModifyMembersRequest{
					Add:   []string{alice},
					Group: organization,
				},
				&auth.ModifyMembersRequest{
					Add:   []string{alice},
					Group: organization,
				},
			},
			map[string][]string{
				alice: []string{organization},
			},
		},
		{
			[]*auth.ModifyMembersRequest{
				&auth.ModifyMembersRequest{
					Add:   []string{bob},
					Group: organization,
				},
				&auth.ModifyMembersRequest{
					Add:   []string{alice, bob},
					Group: engineering,
				},
				&auth.ModifyMembersRequest{
					Add:   []string{bob},
					Group: security,
				},
			},
			map[string][]string{
				alice: []string{organization, engineering},
				bob:   []string{organization, engineering, security},
			},
		},
		{
			[]*auth.ModifyMembersRequest{
				&auth.ModifyMembersRequest{
					Add:    []string{alice},
					Remove: []string{bob},
					Group:  security,
				},
				&auth.ModifyMembersRequest{
					Remove: []string{bob},
					Group:  engineering,
				},
			},
			map[string][]string{
				alice: []string{organization, engineering, security},
				bob:   []string{organization},
			},
		},
		{
			[]*auth.ModifyMembersRequest{
				&auth.ModifyMembersRequest{
					Remove: []string{alice, bob},
					Group:  organization,
				},
				&auth.ModifyMembersRequest{
					Remove: []string{alice, bob},
					Group:  security,
				},
				&auth.ModifyMembersRequest{
					Add:    []string{alice},
					Remove: []string{alice},
					Group:  organization,
				},
				&auth.ModifyMembersRequest{
					Add:    []string{},
					Remove: []string{},
					Group:  organization,
				},
			},
			map[string][]string{
				alice: []string{engineering},
				bob:   []string{},
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
	commit := client.NewCommit(repo, "master", "")
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.Equal(t, tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, repo))
	err := aliceClient.PutFile(commit, "/file", strings.NewReader("lorem ipsum"))
	require.NoError(t, err)

	// alice creates a pipeline
	pipeline := tu.UniqueString("alice-pipeline")
	require.NoError(t, aliceClient.CreatePipeline(
		pipeline,
		"", // default image: DefaultUserImage
		[]string{"bash"},
		[]string{"cp /pfs/*/* /pfs/out/"},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(repo, "/*"),
		"", // default output branch: master
		false,
	))

	// Wait for pipeline to finish
	_, err = aliceClient.WaitCommit(pipeline, "master", commit.ID)
	require.NoError(t, err)

	// alice calls 'list job'
	jobs, err := aliceClient.ListJob("", nil, -1 /*history*/, true)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobs))

	// anonClient calls 'list job'
	_, err = anonClient.ListJob("", nil, -1 /*history*/, true)
	require.YesError(t, err)
	require.Matches(t, "no authentication token", err.Error())

	// alice calls 'list job' again, and the existing job must still be present
	jobs2, err := aliceClient.ListJob("", nil, -1 /*history*/, true)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobs2))
	require.Equal(t, jobs[0].Job.ID, jobs2[0].Job.ID)
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
	commit := client.NewCommit(repo, "master", "")
	require.NoError(t, aliceClient.CreateRepo(repo))
	err := aliceClient.PutFile(commit, "/file", strings.NewReader("1"))
	require.NoError(t, err)

	// Create pipeline
	pipeline := tu.UniqueString("pipeline")
	require.NoError(t, aliceClient.CreatePipeline(
		pipeline,
		"does-not-exist", // nonexistant image
		[]string{"true"}, nil,
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(repo, "/*"),
		"", // default output branch: master
		false,
	))
	require.NoError(t, aliceClient.DeletePipeline(pipeline, true))

	// Get the latest commit from the input repo (which should be an alias from
	// when the pipeline was created)
	commitInfo, err := aliceClient.InspectCommit(repo, "master", "")
	require.NoError(t, err)

	// make sure the pipeline failure doesn't cause waits to block indefinitely
	require.NoErrorWithinT(t, 30*time.Second, func() error {
		_, err := aliceClient.WaitCommitSetAll(commitInfo.Commit.ID)
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
	commit := client.NewCommit(repo, "master", "")
	require.NoError(t, aliceClient.CreateRepo(repo))
	err := aliceClient.PutFile(commit, "/file", strings.NewReader("1"))
	require.NoError(t, err)

	// Create pipeline
	pipeline := tu.UniqueString("pipeline")
	require.NoError(t, aliceClient.CreatePipeline(
		pipeline,
		"does-not-exist", // nonexistant image
		[]string{"true"}, nil,
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(repo, "/*"),
		"", // default output branch: master
		false,
	))

	// force-delete input and output repos
	require.NoError(t, aliceClient.DeleteRepo(repo, true))
	require.NoError(t, aliceClient.DeleteRepo(pipeline, true))

	// Attempt to delete the pipeline--must succeed
	require.NoError(t, aliceClient.DeletePipeline(pipeline, true))
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

	fastExpirationResp, fastExpErr := adminClient.GetRobotToken(adminClient.Ctx(), &auth.GetRobotTokenRequest{Robot: "robot:alice", TTL: 1})
	require.NoError(t, fastExpErr)

	slowExpirationResp, slowExpErr := adminClient.GetRobotToken(adminClient.Ctx(), &auth.GetRobotTokenRequest{Robot: "robot:alice", TTL: 1000})
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
	require.NoError(t, aliceClient.CreateRepo(repo))

	// Admin can list repos
	repoInfo, err := adminClient.ListRepo()
	require.NoError(t, err)
	require.Equal(t, 1, len(repoInfo))

	// Alice can list repos
	repoInfo, err = aliceClient.ListRepo()
	require.NoError(t, err)
	require.Equal(t, 1, len(repoInfo))

	// set Enterprise Token value to have expired
	ts := &types.Timestamp{Seconds: time.Now().Unix() - 100}
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
	require.Equal(t, []string{"clusterAdmin", "repoOwner", "repoReader", "repoWriter"}, names)
}

// TODO: This test mirrors TestLoad in src/server/pfs/server/testing/load_test.go.
// Need to restructure testing such that we have the implementation of this
// test in one place while still being able to test auth enabled and disabled clusters.
func TestLoad(t *testing.T) {
	t.Parallel()
	env := envWithAuth(t)
	c := env.PachClient
	alice := tu.UniqueString("robot:alice")
	aliceClient := tu.AuthenticateClient(t, c, alice)
	resp, err := aliceClient.PfsAPIClient.RunLoadTestDefault(aliceClient.Ctx(), &types.Empty{})
	require.NoError(t, err)
	buf := &bytes.Buffer{}
	require.NoError(t, cmdutil.Encoder("", buf).EncodeProto(resp))
	require.Equal(t, "", resp.Error, buf.String())
}
