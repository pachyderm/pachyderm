//go:build k8s

package server

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/minio/minio-go/v6"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/license"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

// works with global.postgresql.postgresqlAuthType: md5, which is default in values file.
var defaultTestOptions = minikubetestenv.WithValueOverrides(map[string]string{
	"global.postgresql.postgresqlPassword": "-strong-password!@#$%^&*(){}[]|\\:\"'<>?,./_+=0123A",
})

// TODO(Fahad): Potentially convert these to unit tests by using Jon's fake kube apiserver

// TestListDatum tests that you must have READER access to all of job's
// input repos to call ListDatum on that job
func TestListDatum(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t, defaultTestOptions)
	tu.ActivateAuthClient(t, c)
	alice, bob := tu.Robot(tu.UniqueString("alice")), tu.Robot(tu.UniqueString("bob"))
	aliceClient, bobClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, bob)

	// alice creates a repo
	repoA := tu.UniqueString(t.Name())
	require.NoError(t, aliceClient.CreateRepo(pfs.DefaultProjectName, repoA))
	repoB := tu.UniqueString(t.Name())
	require.NoError(t, aliceClient.CreateRepo(pfs.DefaultProjectName, repoB))

	// alice creates a pipeline
	pipeline := tu.UniqueString("alice-pipeline")
	input := client.NewCrossInput(
		client.NewPFSInput(pfs.DefaultProjectName, repoA, "/*"),
		client.NewPFSInput(pfs.DefaultProjectName, repoB, "/*"),
	)
	require.NoError(t, aliceClient.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"", // default image: DefaultUserImage
		[]string{"bash"},
		[]string{"ls /pfs/*/*; cp /pfs/*/* /pfs/out/"},
		&pps.ParallelismSpec{Constant: 1},
		input,
		"", // default output branch: master
		false,
	))

	// alice commits to the input repos, and the pipeline runs successfully
	for i, repo := range []string{repoA, repoB} {
		var err error
		file := fmt.Sprintf("/file%d", i+1)
		err = aliceClient.PutFile(client.NewCommit(pfs.DefaultProjectName, repo, "master", ""), file, strings.NewReader("test"))
		require.NoError(t, err)
	}
	require.NoErrorWithinT(t, 90*time.Second, func() error {
		_, err := aliceClient.WaitCommit(pfs.DefaultProjectName, pipeline, "master", "")
		return err
	})
	jobs, err := aliceClient.ListJob(pfs.DefaultProjectName, pipeline, nil /*inputs*/, -1 /*history*/, true /* full */)
	require.NoError(t, err)
	require.Equal(t, 3, len(jobs))
	jobID := jobs[0].Job.Id

	// bob cannot call ListDatum
	_, err = bobClient.ListDatumAll(pfs.DefaultProjectName, pipeline, jobID)
	require.YesError(t, err)
	require.True(t, auth.IsErrNotAuthorized(err), err.Error())

	// alice adds bob to repoA, but bob still can't call GetLogs
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, repoA, bob, []string{auth.RepoReaderRole}))
	_, err = bobClient.ListDatumAll(pfs.DefaultProjectName, pipeline, jobID)
	require.YesError(t, err)
	require.True(t, auth.IsErrNotAuthorized(err), err.Error())

	// alice removes bob from repoA and adds bob to repoB, but bob still can't
	// call ListDatum
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, repoA, bob, []string{}))
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, repoB, bob, []string{auth.RepoReaderRole}))
	_, err = bobClient.ListDatumAll(pfs.DefaultProjectName, pipeline, jobID)
	require.YesError(t, err)
	require.True(t, auth.IsErrNotAuthorized(err), err.Error())

	// alice adds bob to repoA, and now bob can call ListDatum
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, repoA, bob, []string{auth.RepoReaderRole}))
	_, err = bobClient.ListDatumAll(pfs.DefaultProjectName, pipeline, jobID)
	require.YesError(t, err)
	require.True(t, auth.IsErrNotAuthorized(err), err.Error())

	// Finally, alice adds bob to the output repo, and now bob can call ListDatum
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, pipeline, bob, []string{auth.RepoReaderRole}))
	dis, err := bobClient.ListDatumAll(pfs.DefaultProjectName, pipeline, jobID)
	require.NoError(t, err)
	files := make(map[string]struct{})
	for _, di := range dis {
		for _, f := range di.Data {
			files[path.Base(f.File.Path)] = struct{}{}
		}
	}
	require.Equal(t, map[string]struct{}{
		"file1": {},
		"file2": {},
	}, files)

	// Test list datum input.
	disInput, err := bobClient.ListDatumInputAll(input)
	require.NoError(t, err)
	require.Equal(t, len(dis), len(disInput))
	for i, di := range dis {
		require.Equal(t, di.Datum.Id, disInput[i].Datum.Id)
	}
}

func TestS3GatewayAuthRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()

	c, _ := minikubetestenv.AcquireCluster(t, defaultTestOptions)
	tu.ActivateAuthClient(t, c)
	// generate auth credentials
	adminClient := tu.AuthenticateClient(t, c, auth.RootUser)
	alice := tu.UniqueString("alice")
	authResp, err := adminClient.GetRobotToken(adminClient.Ctx(), &auth.GetRobotTokenRequest{
		Robot: alice,
	})
	require.NoError(t, err)
	authToken := authResp.Token

	ip := os.Getenv("VM_IP")
	if ip == "" {
		ip = "127.0.0.1"
	}
	// Port set dynamically in src/internal/minikubetestenv/deploy.go
	address := net.JoinHostPort(ip, fmt.Sprint(c.GetAddress().Port+3))

	// anon login via V2 - should fail
	minioClientV2, err := minio.NewV2(address, "", "", false)
	require.NoError(t, err)
	_, err = minioClientV2.ListBuckets()
	require.YesError(t, err)

	// anon login via V4 - should fail
	minioClientV4, err := minio.NewV4(address, "", "", false)
	require.NoError(t, err)
	_, err = minioClientV4.ListBuckets()
	require.YesError(t, err)

	// proper login via V2 - should succeed
	minioClientV2, err = minio.NewV2(address, authToken, authToken, false)
	require.NoError(t, err)
	_, err = minioClientV2.ListBuckets()
	require.NoError(t, err)

	// proper login via V4 - should succeed
	minioClientV2, err = minio.NewV4(address, authToken, authToken, false)
	require.NoError(t, err)
	_, err = minioClientV2.ListBuckets()
	require.NoError(t, err)
}

// Need to restructure testing such that we have the implementation of this
// test in one place while still being able to test auth enabled and disabled clusters.
func testDebug(t *testing.T, c *client.APIClient, projectName, repoName string) {
	t.Helper()
	// Get all the authenticated clients at the beginning of the test.
	// GetAuthenticatedPachClient will always re-activate auth, which
	// causes PPS to rotate all the pipeline tokens. This makes the RCs
	// change and recreates all the pods, which used to race with collecting logs.
	alice := tu.Robot(tu.UniqueString("alice"))
	aliceClient, adminClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, auth.RootUser)
	if projectName != "default" {
		require.NoError(t, aliceClient.CreateProject(projectName))
	}

	require.NoError(t, aliceClient.CreateRepo(projectName, repoName))

	expectedFiles, pipelines := tu.DebugFiles(t, projectName, repoName)

	for _, p := range pipelines {
		require.NoError(t, aliceClient.CreatePipeline(projectName,
			p,
			"",
			[]string{"bash"},
			[]string{
				fmt.Sprintf("cp /pfs/%s/* /pfs/out/", repoName),
				"sleep 45",
			},
			&pps.ParallelismSpec{
				Constant: 1,
			},
			client.NewPFSInput(projectName, repoName, "/*"),
			"",
			false,
		))
	}

	commit1, err := aliceClient.StartCommit(projectName, repoName, "master")
	require.NoError(t, err)
	err = aliceClient.PutFile(commit1, "file", strings.NewReader("foo"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(projectName, repoName, commit1.Branch.Name, commit1.Id))

	jobInfos, err := aliceClient.WaitJobSetAll(commit1.Id, false)
	require.NoError(t, err)
	require.Equal(t, 3, len(jobInfos))

	require.YesError(t, aliceClient.Dump(nil, 0, &bytes.Buffer{}))

	require.NoErrorWithinT(t, time.Minute, func() error {
		// Only admins can collect a debug dump.
		buf := &bytes.Buffer{}
		require.NoError(t, adminClient.Dump(nil, 0, buf))
		gr, err := gzip.NewReader(buf)
		if err != nil {
			return err //nolint:wrapcheck
		}
		defer func() {
			require.NoError(t, gr.Close())
		}()
		// Check that all of the expected files were returned.
		tr := tar.NewReader(gr)
		for {
			hdr, err := tr.Next()
			if err != nil {
				if err == io.EOF {
					break
				}
				return err //nolint:wrapcheck
			}
			for pattern, g := range expectedFiles {
				if g.Match(hdr.Name) {
					delete(expectedFiles, pattern)
					break
				}
			}
		}
		if len(expectedFiles) > 0 {
			return errors.Errorf("Debug dump hasn't produced the exepcted files: %v", expectedFiles)
		}
		return nil
	})
}

func TestDebug(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t, defaultTestOptions)
	tu.ActivateAuthClient(t, c)
	for _, projectName := range []string{pfs.DefaultProjectName, tu.UniqueString("project")} {
		t.Run(projectName, func(t *testing.T) {
			testDebug(t, c, projectName, tu.UniqueString("repo"))
		})
	}
}

// asserts that retrieval of Pachd logs requires additional permissions granted to the PachdLogReader role
func TestGetPachdLogsRequiresPerm(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t, defaultTestOptions)
	tu.ActivateAuthClient(t, c)
	adminClient := tu.AuthenticateClient(t, c, auth.RootUser)

	alice := tu.UniqueString("robot:alice")
	aliceClient := tu.AuthenticateClient(t, c, alice)

	aliceRepo := tu.UniqueString("alice_repo")
	err := aliceClient.CreateRepo(pfs.DefaultProjectName, aliceRepo)
	require.NoError(t, err)

	// create pipeline
	alicePipeline := tu.UniqueString("pipeline_for_logs")
	err = aliceClient.CreatePipeline(pfs.DefaultProjectName,
		alicePipeline,
		"", // default image: DefaultUserImage
		[]string{"bash"},
		[]string{"cp /pfs/*/* /pfs/out/"},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(pfs.DefaultProjectName, aliceRepo, "/*"),
		"", // default output branch: master
		false,
	)
	require.NoError(t, err)

	// wait for the pipeline pod to come up
	time.Sleep(time.Second * 10)

	// must be authorized to access pachd logs
	pachdLogsIter := aliceClient.GetLogs(pfs.DefaultProjectName, "", "", nil, "", false, false, 0)
	pachdLogsIter.Next()
	require.YesError(t, pachdLogsIter.Err())
	require.True(t, strings.Contains(pachdLogsIter.Err().Error(), "is not authorized to perform this operation"))

	// alice can view the pipeline logs
	require.NoErrorWithinTRetry(t, time.Minute, func() error {
		pipelineLogsIter := aliceClient.GetLogs(pfs.DefaultProjectName, alicePipeline, "", nil, "", false, false, 0)
		pipelineLogsIter.Next()
		return pipelineLogsIter.Err()
	}, "alice can view the pipeline logs")

	// PachdLogReaderRole grants authorized retrieval of pachd logs
	_, err = adminClient.AuthAPIClient.ModifyRoleBinding(adminClient.Ctx(),
		&auth.ModifyRoleBindingRequest{
			Principal: alice,
			Roles:     []string{auth.PachdLogReaderRole},
			Resource:  &auth.Resource{Type: auth.ResourceType_CLUSTER},
		})
	require.NoError(t, err)
	pachdLogsIter = aliceClient.GetLogs(pfs.DefaultProjectName, "", "", nil, "", false, false, 0)
	pachdLogsIter.Next()
	require.NoError(t, pachdLogsIter.Err())
}

// TODO: Make logs work with V2.
// TestGetLogs tests that you must have READER access to all of a job's input
// repos and READER access to its output repo to call GetLogs()
//func TestGetLogs(t *testing.T) {
//	if testing.Short() {
//		t.Skip("Skipping integration tests in short mode")
//	}
//	tu.DeleteAll(t)
//	defer tu.DeleteAll(t)
//	alice, bob := tu.Robot(tu.UniqueString("alice")), tu.Robot(tu.UniqueString("bob"))
//	aliceClient, bobClient := tu.GetAuthenticatedPachClient(t, alice), tu.GetAuthenticatedPachClient(t, bob)
//
//	// alice creates a repo
//	repo := tu.UniqueString(t.Name())
//	require.NoError(t, aliceClient.CreateProjectRepo(pfs.DefaultProjectName,repo))
//
//	// alice creates a pipeline
//	pipeline := tu.UniqueString("pipeline")
//	require.NoError(t, aliceClient.CreateProjectPipeline(pfs.DefaultProjectName,
//		pipeline,
//		"", // default image: DefaultUserImage
//		[]string{"bash"},
//		[]string{"cp /pfs/*/* /pfs/out/"},
//		&pps.ParallelismSpec{Constant: 1},
//		client.NewProjectPFSInput(pfs.DefaultProjectName,repo, "/*"),
//		"", // default output branch: master
//		false,
//	))
//
//	// alice commits to the input repos, and the pipeline runs successfully
//	err := aliceClient.PutFile(repo, "master", "/file1", strings.NewReader("test"))
//	require.NoError(t, err)
//	commitIter, err := aliceClient.WaitProjectCommit(pfs.DefaultProjectName,pipeline, "master", "")
//	require.NoError(t, err)
//	require.NoErrorWithinT(t, 60*time.Second, func() error {
//		_, err := commitIter.Next()
//		return err
//	})
//
//	// bob cannot call GetLogs
//	iter := bobClient.GetProjectLogs(pfs.DefaultProjectName,pipeline, "", nil, "", false, false, 0)
//	require.False(t, iter.Next())
//	require.YesError(t, iter.Err())
//	require.True(t, auth.IsErrNotAuthorized(iter.Err()), iter.Err().Error())
//
//	// bob also can't call GetLogs for the master process
//	iter = bobClient.GetProjectLogs(pfs.DefaultProjectName,pipeline, "", nil, "", true, false, 0)
//	require.False(t, iter.Next())
//	require.YesError(t, iter.Err())
//	require.True(t, auth.IsErrNotAuthorized(iter.Err()), iter.Err().Error())
//
//	// alice adds bob to the input repo, but bob still can't call GetLogs
//	aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
//		Username: bob,
//		Scope:    auth.Scope_READER,
//		Repo:     repo,
//	})
//	iter = bobClient.GetProjectLogs(pfs.DefaultProjectName,pipeline, "", nil, "", false, false, 0)
//	require.False(t, iter.Next())
//	require.YesError(t, iter.Err())
//	require.True(t, auth.IsErrNotAuthorized(iter.Err()), iter.Err().Error())
//
//	// alice removes bob from the input repo and adds bob to the output repo, but
//	// bob still can't call GetLogs
//	aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
//		Username: bob,
//		Scope:    auth.Scope_NONE,
//		Repo:     repo,
//	})
//	aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
//		Username: bob,
//		Scope:    auth.Scope_READER,
//		Repo:     pipeline,
//	})
//	iter = bobClient.GetProjectLogs(pfs.DefaultProjectName,pipeline, "", nil, "", false, false, 0)
//	require.False(t, iter.Next())
//	require.YesError(t, iter.Err())
//	require.True(t, auth.IsErrNotAuthorized(iter.Err()), iter.Err().Error())
//
//	// alice adds bob to the output repo, and now bob can call GetLogs
//	aliceClient.SetScope(aliceClient.Ctx(), &auth.SetScopeRequest{
//		Username: bob,
//		Scope:    auth.Scope_READER,
//		Repo:     repo,
//	})
//	iter = bobClient.GetProjectLogs(pfs.DefaultProjectName,pipeline, "", nil, "", false, false, 0)
//	iter.Next()
//	require.NoError(t, iter.Err())
//
//	// bob can also call GetLogs for the master process
//	iter = bobClient.GetProjectLogs(pfs.DefaultProjectName,pipeline, "", nil, "", true, false, 0)
//	iter.Next()
//	require.NoError(t, iter.Err())
//}
//
//// TestGetLogsFromStats tests that GetLogs still works even when stats are
//// enabled
//func TestGetLogsFromStats(t *testing.T) {
//	if testing.Short() {
//		t.Skip("Skipping integration tests in short mode")
//	}
//	tu.DeleteAll(t)
//	defer tu.DeleteAll(t)
//	alice := tu.Robot(tu.UniqueString("alice"))
//	aliceClient := tu.GetAuthenticatedPachClient(t, alice)
//
//	// alice creates a repo
//	repo := tu.UniqueString(t.Name())
//	require.NoError(t, aliceClient.CreateProjectRepo(pfs.DefaultProjectName,repo))
//
//	// alice creates a pipeline (we must enable stats for InspectDatum, which
//	// means calling the grpc client function directly)
//	pipeline := tu.UniqueString("alice")
//	_, err := aliceClient.PpsAPIClient.CreateProjectPipeline(pfs.DefaultProjectName,aliceClient.Ctx(),
//		&pps.CreatePipelineRequest{
//			Pipeline: &pps.Pipeline{Name: pipeline},
//			Transform: &pps.Transform{
//				Cmd:   []string{"bash"},
//				Stdin: []string{"cp /pfs/*/* /pfs/out/"},
//			},
//			ParallelismSpec: &pps.ParallelismSpec{Constant: 1},
//			Input:           client.NewProjectPFSInput(pfs.DefaultProjectName,repo, "/*"),
//		})
//	require.NoError(t, err)
//
//	// alice commits to the input repo, and the pipeline runs successfully
//	err = aliceClient.PutFile(repo, "master", "/file1", strings.NewReader("test"))
//	require.NoError(t, err)
//	commitItr, err := aliceClient.WaitProjectCommit(pfs.DefaultProjectName,pipeline, "master", "")
//	require.NoError(t, err)
//	require.NoErrorWithinT(t, 3*time.Minute, func() error {
//		_, err := commitItr.Next()
//		return err
//	})
//	jobs, err := aliceClient.ListJob(pipeline, nil /*inputs*/, -1 /*history*/, true)
//	require.NoError(t, err)
//	require.Equal(t, 1, len(jobs))
//	jobID := jobs[0].Job.ID
//
//	iter := aliceClient.GetProjectLogs(pfs.DefaultProjectName,pipeline, jobID, nil, "", false, false, 0)
//	require.True(t, iter.Next())
//	require.NoError(t, iter.Err())
//
//	iter = aliceClient.GetProjectLogs(pfs.DefaultProjectName,pipeline, jobID, nil, "", true, false, 0)
//	iter.Next()
//	require.NoError(t, iter.Err())
//}

// TestPipelineRevoke tests revoking the privileges of a pipeline's creator as
// well as revoking the pipeline itself.
//
// When pipelines inherited privileges from their creator, revoking the owner's
// access to the pipeline's inputs would cause pipelines to stop running, but
// now it does not. In general, this should actually be more secure--it used to
// be that any pipeline Bob created could access any repo that Bob could, even
// if the repo was unrelated to the pipeline (making pipelines a powerful
// vector for privilege escalation). Now pipelines are their own principals,
// and they can only read from their inputs and write to their outputs.
//
// Ideally both would be required: if either the pipeline's access to its inputs
// or bob's access to the pipeline's inputs are revoked, the pipeline should
// stop, but for now it's required to revoke the pipeline's access directly
func TestPipelineRevoke(t *testing.T) {
	t.Skip("TestPipelineRevoke is broken")
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t, defaultTestOptions)
	tu.ActivateAuthClient(t, c)
	alice, bob := tu.Robot(tu.UniqueString("alice")), tu.Robot(tu.UniqueString("bob"))
	aliceClient, bobClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, bob)

	// alice creates a repo, and adds bob as a reader
	repo := tu.UniqueString(t.Name())
	require.NoError(t, aliceClient.CreateRepo(pfs.DefaultProjectName, repo))
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, repo, bob, []string{auth.RepoReaderRole}))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, bob, auth.RepoReaderRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, repo))
	commit := client.NewCommit(pfs.DefaultProjectName, repo, "master", "")

	// bob creates a pipeline
	pipeline := tu.UniqueString("bob-pipeline")
	require.NoError(t, bobClient.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"", // default image: DefaultUserImage
		[]string{"bash"},
		[]string{"cp /pfs/*/* /pfs/out/"},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(pfs.DefaultProjectName, repo, "/*"),
		"", // default output branch: master
		false,
	))
	require.Equal(t,
		tu.BuildBindings(bob, auth.RepoOwnerRole, tu.Pl(pfs.DefaultProjectName, pipeline), auth.RepoWriterRole), tu.GetRepoRoleBinding(t, bobClient, pfs.DefaultProjectName, pipeline))
	// bob adds alice as a reader of the pipeline's output repo, so alice can
	// flush input commits (which requires her to inspect commits in the output)
	// and update the pipeline
	require.NoError(t, bobClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, pipeline, alice, []string{auth.RepoWriterRole}))
	require.Equal(t,
		tu.BuildBindings(bob, auth.RepoOwnerRole, alice, auth.RepoWriterRole, tu.Pl(pfs.DefaultProjectName, pipeline), auth.RepoWriterRole),
		tu.GetRepoRoleBinding(t, bobClient, pfs.DefaultProjectName, pipeline))

	// alice commits to the input repo, and the pipeline runs successfully
	require.NoError(t, aliceClient.PutFile(commit, "/file", strings.NewReader("test")))
	require.NoErrorWithinT(t, 45*time.Second, func() error {
		_, err := bobClient.WaitCommit(pfs.DefaultProjectName, pipeline, "master", commit.Id)
		return err
	})

	// alice removes bob as a reader of her repo, and then commits to the input
	// repo, but bob's pipeline still runs (it has its own principal--it doesn't
	// inherit bob's privileges)
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, repo, bob, []string{}))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Pl(pfs.DefaultProjectName, pipeline), auth.RepoReaderRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, repo))
	require.NoError(t, aliceClient.PutFile(commit, "/file", strings.NewReader("test")))
	require.NoErrorWithinT(t, 45*time.Second, func() error {
		_, err := aliceClient.WaitCommit(pfs.DefaultProjectName, pipeline, "master", commit.Id)
		return err
	})

	// alice revokes the pipeline's access to 'repo' directly, and the pipeline
	// stops running
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, repo, tu.Pl(pfs.DefaultProjectName, pipeline), []string{}))
	require.NoError(t, aliceClient.PutFile(commit, "/file", strings.NewReader("test")))
	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		_, err := aliceClient.WaitCommit(pfs.DefaultProjectName, pipeline, "master", commit.Id)
		require.NoError(t, err)
	}()
	select {
	case <-doneCh:
		t.Fatal("pipeline should not be able to finish with no access")
	case <-time.After(45 * time.Second):
	}

	// alice updates bob's pipline, but the pipeline still doesn't run
	require.NoError(t, aliceClient.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"", // default image: DefaultUserImage
		[]string{"bash"},
		[]string{"cp /pfs/*/* /pfs/out/"},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(pfs.DefaultProjectName, repo, "/*"),
		"", // default output branch: master
		true,
	))
	require.NoError(t, aliceClient.PutFile(commit, "/file", strings.NewReader("test")))
	doneCh = make(chan struct{})
	go func() {
		defer close(doneCh)
		_, err := aliceClient.WaitCommit(pfs.DefaultProjectName, pipeline, "master", commit.Id)
		require.NoError(t, err)
	}()
	select {
	case <-doneCh:
		t.Fatal("pipeline should not be able to finish with no access")
	case <-time.After(45 * time.Second):
	}

	// alice restores the pipeline's access to its input repo, and now the
	// pipeline runs successfully
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pfs.DefaultProjectName, repo, tu.Pl(pfs.DefaultProjectName, pipeline), []string{auth.RepoReaderRole}))
	require.NoErrorWithinT(t, 45*time.Second, func() error {
		_, err := aliceClient.WaitCommit(pfs.DefaultProjectName, pipeline, "master", commit.Id)
		return err
	})
}

// TestDeleteRCInStandby creates a pipeline, waits for it to enter standby, and
// then deletes its RC. This should not crash the PPS master, and the
// commit should eventually finish (though the pipeline may fail rather than
// processing anything in this state)
//
// Note: Like 'TestNoOutputRepoDoesntCrashPPSMaster', this test doesn't use the
// admin client at all, but it uses the kubernetes client, so out of prudence it
// shouldn't be run in parallel with any other test
func TestDeleteRCInStandby(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, ns := minikubetestenv.AcquireCluster(t, defaultTestOptions)
	tu.ActivateAuthClient(t, c)
	alice := tu.Robot(tu.UniqueString("alice"))
	c = tu.AuthenticateClient(t, c, alice)

	// Create input repo w/ initial commit
	repo := tu.UniqueString(t.Name())
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repo))
	err := c.PutFile(client.NewCommit(pfs.DefaultProjectName, repo, "master", ""), "/file.1", strings.NewReader("1"))
	require.NoError(t, err)

	// Create pipeline
	pipeline := tu.UniqueString("pipeline")
	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
			Transform: &pps.Transform{
				Image: "", // default image: DefaultUserImage
				Cmd:   []string{"bash"},
				Stdin: []string{"cp /pfs/*/* /pfs/out"},
			},
			ParallelismSpec: &pps.ParallelismSpec{Constant: 1},
			Input:           client.NewPFSInput(pfs.DefaultProjectName, repo, "/*"),
			Autoscaling:     true,
		})
	require.NoError(t, err)

	// Wait for pipeline to process input commit & go into standby
	require.NoErrorWithinT(t, 30*time.Second, func() error {
		_, err := c.WaitCommit(pfs.DefaultProjectName, pipeline, "master", "")
		return err
	})
	require.NoErrorWithinTRetry(t, 30*time.Second, func() error {
		pi, err := c.InspectPipeline(pfs.DefaultProjectName, pipeline, false)
		if err != nil {
			return err
		}
		if pi.State != pps.PipelineState_PIPELINE_STANDBY {
			return errors.Errorf("pipeline should be in standby, but is in %s", pi.State.String())
		}
		return nil
	})

	// delete pipeline RC
	tu.DeletePipelineRC(t, pipeline, ns)

	// Create new input commit (to force pipeline out of standby) & make sure
	// the pipeline either fails or restarts RC & finishes
	err = c.PutFile(client.NewCommit(pfs.DefaultProjectName, repo, "master", ""), "/file.2", strings.NewReader("1"))
	require.NoError(t, err)
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := c.WaitCommit(pfs.DefaultProjectName, pipeline, "master", "")
		return err
	})
}

func TestPreActivationCronPipelinesKeepRunningAfterActivation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t, defaultTestOptions)
	tu.ActivateAuthClient(t, c)
	alice := tu.Robot(tu.UniqueString("alice"))
	aliceClient, rootClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, auth.RootUser)

	// Deactivate auth
	_, err := rootClient.Deactivate(rootClient.Ctx(), &auth.DeactivateRequest{})
	require.NoError(t, err)

	// Wait for auth to be deactivated
	require.NoError(t, backoff.Retry(func() error {
		_, err := aliceClient.WhoAmI(aliceClient.Ctx(), &auth.WhoAmIRequest{})
		if err != nil && auth.IsErrNotActivated(err) {
			return nil // WhoAmI should fail when auth is deactivated
		}
		return errors.New("auth is not yet deactivated")
	}, backoff.NewTestingBackOff()))

	// alice creates a pipeline
	pipeline1 := tu.UniqueString("cron1-")
	require.NoError(t, aliceClient.CreatePipeline(pfs.DefaultProjectName,
		pipeline1,
		"",
		[]string{"/bin/bash"},
		[]string{"cp /pfs/time/* /pfs/out/"},
		nil,
		client.NewCronInput("time", "@every 3s"),
		"",
		false,
	))
	pipeline2 := tu.UniqueString("cron2-")
	require.NoError(t, aliceClient.CreatePipeline(pfs.DefaultProjectName,
		pipeline2,
		"",
		[]string{"/bin/bash"},
		[]string{"cp " + fmt.Sprintf("/pfs/%s/*", pipeline1) + " /pfs/out/"},
		nil,
		client.NewPFSInput(pfs.DefaultProjectName, pipeline1, "/*"),
		"",
		false,
	))

	// subscribe to the pipeline2 cron repo and wait for inputs
	repo := client.NewRepo(pfs.DefaultProjectName, pipeline2)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel() //cleanup resources

	checkCronCommits := func(n int) error {
		count := 0
		return aliceClient.WithCtx(ctx).SubscribeCommit(repo, "master", "", pfs.CommitState_FINISHED, func(ci *pfs.CommitInfo) error {
			files, err := aliceClient.ListFileAll(ci.Commit, "")
			require.NoError(t, err)

			require.Equal(t, count, len(files))
			count++
			if count > n {
				return errutil.ErrBreak
			}
			return nil
		})
	}
	// make sure the cron is working
	require.NoError(t, checkCronCommits(1))

	// activate auth
	resp, err := rootClient.Activate(rootClient.Ctx(), &auth.ActivateRequest{RootToken: tu.RootToken})
	require.NoError(t, err)
	rootClient.SetAuthToken(resp.PachToken)

	// activate auth in PFS
	_, err = rootClient.PfsAPIClient.ActivateAuth(rootClient.Ctx(), &pfs.ActivateAuthRequest{})
	require.NoError(t, err)

	// activate auth in PPS
	_, err = rootClient.PpsAPIClient.ActivateAuth(rootClient.Ctx(), &pps.ActivateAuthRequest{})
	require.NoError(t, err)

	// re-authenticate, as old tokens were deleted
	aliceClient = tu.AuthenticateClient(t, c, alice)
	require.NoError(t, rootClient.ModifyClusterRoleBinding(alice, []string{auth.RepoWriterRole}))

	// make sure the cron is working
	require.NoError(t, checkCronCommits(5))
}

func TestPipelinesRunAfterExpiration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t, defaultTestOptions)
	tu.ActivateAuthClient(t, c)
	alice := tu.Robot(tu.UniqueString("alice"))
	aliceClient, rootClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, auth.RootUser)

	// alice creates a repo
	repo := tu.UniqueString("TestPipelinesRunAfterExpiration")
	require.NoError(t, aliceClient.CreateRepo(pfs.DefaultProjectName, repo))
	require.Equal(t, tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, repo))

	// alice creates a pipeline
	pipeline := tu.UniqueString("alice-pipeline")
	require.NoError(t, aliceClient.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"", // default image: DefaultUserImage
		[]string{"bash"},
		[]string{fmt.Sprintf("cp /pfs/%s/* /pfs/out/", repo)},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(pfs.DefaultProjectName, repo, "/*"),
		"",    // default output branch: master
		false, // no update
	))
	require.OneOfEquals(t, pipeline, tu.PipelineNames(t, aliceClient, pfs.DefaultProjectName))
	// check that alice owns the output repo too,
	require.Equal(t, tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Pl(pfs.DefaultProjectName, pipeline), auth.RepoWriterRole), tu.GetRepoRoleBinding(t, aliceClient, pfs.DefaultProjectName, pipeline))

	// Make sure alice's pipeline runs successfully
	commit, err := aliceClient.StartCommit(pfs.DefaultProjectName, repo, "master")
	require.NoError(t, err)
	err = aliceClient.PutFile(commit, tu.UniqueString("/file1"),
		strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(pfs.DefaultProjectName, repo, commit.Branch.Name, commit.Id))
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := aliceClient.WaitCommit(pfs.DefaultProjectName, pipeline, "master", commit.Id)
		return err
	})

	// Make current enterprise token expire
	_, err = rootClient.License.Activate(rootClient.Ctx(),
		&license.ActivateRequest{
			ActivationCode: tu.GetTestEnterpriseCode(t),
			Expires:        tu.TSProtoOrDie(t, time.Now().Add(-30*time.Second)),
		})
	require.NoError(t, err)
	_, err = rootClient.Enterprise.Activate(rootClient.Ctx(),
		&enterprise.ActivateRequest{
			LicenseServer: "localhost:1650",
			Id:            "localhost",
			Secret:        "localhost",
		})
	require.NoError(t, err)

	// wait for Enterprise token to expire
	require.NoError(t, backoff.Retry(func() error {
		resp, err := rootClient.Enterprise.GetState(rootClient.Ctx(),
			&enterprise.GetStateRequest{})
		if err != nil {
			return errors.EnsureStack(err)
		}
		if resp.State == enterprise.State_ACTIVE {
			return errors.New("Pachyderm Enterprise is still active")
		}
		return nil
	}, backoff.NewTestingBackOff()))

	// Make sure alice's pipeline still runs successfully
	commit, err = rootClient.StartCommit(pfs.DefaultProjectName, repo, "master")
	require.NoError(t, err)
	err = rootClient.PutFile(commit, tu.UniqueString("/file2"),
		strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, rootClient.FinishCommit(pfs.DefaultProjectName, repo, commit.Branch.Name, commit.Id))
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := rootClient.WaitCommit(pfs.DefaultProjectName, pipeline, "master", commit.Id)
		return err
	})
}

// TestDeleteAllAfterDeactivate tests that deleting repos and (particularly)
// pipelines works if auth was deactivated after they were created. Pipelines
// store a unique auth token after auth is activated, and if that auth token
// is used in the deletion process, DeletePipeline (and therefore DeleteAll)
// fails.
func TestDeleteAllAfterDeactivate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t, defaultTestOptions)
	tu.ActivateAuthClient(t, c)
	alice := tu.Robot(tu.UniqueString("alice"))
	aliceClient, rootClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, auth.RootUser)

	// alice creates a pipeline
	repo := tu.UniqueString("TestDeleteAllAfterDeactivate")
	pipeline := tu.UniqueString("pipeline")
	require.NoError(t, aliceClient.CreateRepo(pfs.DefaultProjectName, repo))
	require.NoError(t, aliceClient.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"", // default image: DefaultUserImage
		[]string{"bash"},
		[]string{fmt.Sprintf("cp /pfs/%s/* /pfs/out/", repo)},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(pfs.DefaultProjectName, repo, "/*"),
		"", // default output branch: master
		false,
	))

	// alice makes an input commit
	commit, err := aliceClient.StartCommit(pfs.DefaultProjectName, repo, "master")
	require.NoError(t, err)
	err = aliceClient.PutFile(commit, "/file1", strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(pfs.DefaultProjectName, repo, commit.Branch.Name, commit.Id))

	// make sure the pipeline runs
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := aliceClient.WaitCommit(pfs.DefaultProjectName, pipeline, "master", commit.Id)
		return err
	})

	// Deactivate auth
	_, err = rootClient.Deactivate(rootClient.Ctx(), &auth.DeactivateRequest{})
	require.NoError(t, err)

	// Wait for auth to be deactivated
	require.NoError(t, backoff.Retry(func() error {
		_, err := aliceClient.WhoAmI(aliceClient.Ctx(), &auth.WhoAmIRequest{})
		if err != nil && auth.IsErrNotActivated(err) {
			return nil // WhoAmI should fail when auth is deactivated
		}
		return errors.New("auth is not yet deactivated")
	}, backoff.NewTestingBackOff()))

	// Make sure DeleteAll() succeeds
	require.NoError(t, aliceClient.DeleteAll())
}

func TestListFileNils(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t, defaultTestOptions)
	tu.ActivateAuthClient(t, c)
	alice := tu.Robot(tu.UniqueString("alice"))
	aliceClient := tu.AuthenticateClient(t, c, alice)
	repo := "foo"
	require.NoError(t, aliceClient.CreateRepo(pfs.DefaultProjectName, repo))
	for _, test := range []*pfs.Commit{
		nil,
		{},
		{Branch: &pfs.Branch{}},
		{Branch: &pfs.Branch{Repo: &pfs.Repo{Name: repo}}},
		{Branch: &pfs.Branch{Repo: &pfs.Repo{Name: repo, Project: &pfs.Project{}}}},
		{
			Branch: &pfs.Branch{
				Repo: &pfs.Repo{Name: repo, Project: &pfs.Project{}},
				Name: "master",
			},
		},
	} {
		if err := aliceClient.ListFile(test, "/", func(fi *pfs.FileInfo) error {
			t.Errorf("dead code ran")
			return errors.New("should never get here")
		}); err == nil {
			t.Errorf("ListFile(nil, %q, …) succeeded where it should have failed", test)
		}
	}
	// this used to cause a core dump
	var commit *pfs.Commit = &pfs.Commit{
		Branch: &pfs.Branch{
			Repo: &pfs.Repo{Name: repo, Project: &pfs.Project{}},
			Name: "master",
		},
		Id: "0123456789ab40123456789abcdef012",
	}
	if err := aliceClient.ListFile(commit, "/", func(fi *pfs.FileInfo) error {
		return errors.New("should never get here")
	}); err == nil {
		t.Errorf("ListFile for a non-existent commit should always be an error")
	} else if strings.Contains(err.Error(), "upstream connect error") {
		t.Errorf("server error: %v", err)
	}
}
