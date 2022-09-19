//go:build k8s

package server

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
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
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/license"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

// TODO(Fahad): make these tests work with realEnv

// TestDeleteAll tests that you must be a cluster admin to call DeleteAll
func TestDeleteAll(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	alice := tu.Robot(tu.UniqueString("alice"))
	aliceClient, adminClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, auth.RootUser)

	// admin creates a repo
	repo := tu.UniqueString(t.Name())
	require.NoError(t, adminClient.CreateRepo(repo))

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

// TestGetPermissions tests that GetPermissions and GetPermissionsForPrincipal work for repos and the cluster itself
func TestGetPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	alice, bob := tu.Robot(tu.UniqueString("alice")), tu.Robot(tu.UniqueString("bob"))
	aliceClient, rootClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, auth.RootUser)

	// alice creates a repo and makes Bob a writer
	repo := tu.UniqueString(t.Name())
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(repo, bob, []string{auth.RepoWriterRole}))

	// alice can get her own permissions on the cluster (none) and on the repo (repoOwner)
	permissions, err := aliceClient.GetPermissions(aliceClient.Ctx(), &auth.GetPermissionsRequest{Resource: &auth.Resource{Type: auth.ResourceType_CLUSTER}})
	require.NoError(t, err)
	require.Nil(t, permissions.Roles)

	permissions, err = aliceClient.GetPermissions(aliceClient.Ctx(), &auth.GetPermissionsRequest{Resource: &auth.Resource{Type: auth.ResourceType_REPO, Name: repo}})
	require.NoError(t, err)
	require.Equal(t, []string{"repoOwner"}, permissions.Roles)

	// the root user can get bob's permissions
	permissions, err = rootClient.GetPermissionsForPrincipal(rootClient.Ctx(), &auth.GetPermissionsForPrincipalRequest{Resource: &auth.Resource{Type: auth.ResourceType_CLUSTER}, Principal: bob})
	require.NoError(t, err)
	require.Nil(t, permissions.Roles)

	permissions, err = rootClient.GetPermissionsForPrincipal(rootClient.Ctx(), &auth.GetPermissionsForPrincipalRequest{Resource: &auth.Resource{Type: auth.ResourceType_REPO, Name: repo}, Principal: bob})
	require.NoError(t, err)
	require.Equal(t, []string{"repoWriter"}, permissions.Roles)

	// alice cannot get bob's permissions
	_, err = aliceClient.GetPermissionsForPrincipal(aliceClient.Ctx(), &auth.GetPermissionsForPrincipalRequest{Resource: &auth.Resource{Type: auth.ResourceType_CLUSTER}, Principal: bob})
	require.YesError(t, err)
	require.Matches(t, "is not authorized to perform this operation - needs permissions", err.Error())

	_, err = aliceClient.GetPermissionsForPrincipal(aliceClient.Ctx(), &auth.GetPermissionsForPrincipalRequest{Resource: &auth.Resource{Type: auth.ResourceType_REPO, Name: repo}, Principal: bob})
	require.YesError(t, err)
	require.Matches(t, "is not authorized to perform this operation - needs permissions", err.Error())
}

// TestListDatum tests that you must have READER access to all of job's
// input repos to call ListDatum on that job
func TestListDatum(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	alice, bob := tu.Robot(tu.UniqueString("alice")), tu.Robot(tu.UniqueString("bob"))
	aliceClient, bobClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, bob)

	// alice creates a repo
	repoA := tu.UniqueString(t.Name())
	require.NoError(t, aliceClient.CreateRepo(repoA))
	repoB := tu.UniqueString(t.Name())
	require.NoError(t, aliceClient.CreateRepo(repoB))

	// alice creates a pipeline
	pipeline := tu.UniqueString("alice-pipeline")
	require.NoError(t, aliceClient.CreatePipeline(
		pipeline,
		"", // default image: DefaultUserImage
		[]string{"bash"},
		[]string{"ls /pfs/*/*; cp /pfs/*/* /pfs/out/"},
		&pps.ParallelismSpec{Constant: 1},
		client.NewCrossInput(
			client.NewPFSInput(repoA, "/*"),
			client.NewPFSInput(repoB, "/*"),
		),
		"", // default output branch: master
		false,
	))

	// alice commits to the input repos, and the pipeline runs successfully
	for i, repo := range []string{repoA, repoB} {
		var err error
		file := fmt.Sprintf("/file%d", i+1)
		err = aliceClient.PutFile(client.NewCommit(repo, "master", ""), file, strings.NewReader("test"))
		require.NoError(t, err)
	}
	require.NoErrorWithinT(t, 90*time.Second, func() error {
		_, err := aliceClient.WaitCommit(pipeline, "master", "")
		return err
	})
	jobs, err := aliceClient.ListJob(pipeline, nil /*inputs*/, -1 /*history*/, true /* full */)
	require.NoError(t, err)
	require.Equal(t, 3, len(jobs))
	jobID := jobs[0].Job.ID

	// bob cannot call ListDatum
	_, err = bobClient.ListDatumAll(pipeline, jobID)
	require.YesError(t, err)
	require.True(t, auth.IsErrNotAuthorized(err), err.Error())

	// alice adds bob to repoA, but bob still can't call GetLogs
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(repoA, bob, []string{auth.RepoReaderRole}))
	_, err = bobClient.ListDatumAll(pipeline, jobID)
	require.YesError(t, err)
	require.True(t, auth.IsErrNotAuthorized(err), err.Error())

	// alice removes bob from repoA and adds bob to repoB, but bob still can't
	// call ListDatum
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(repoA, bob, []string{}))
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(repoB, bob, []string{auth.RepoReaderRole}))
	_, err = bobClient.ListDatumAll(pipeline, jobID)
	require.YesError(t, err)
	require.True(t, auth.IsErrNotAuthorized(err), err.Error())

	// alice adds bob to repoA, and now bob can call ListDatum
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(repoA, bob, []string{auth.RepoReaderRole}))
	_, err = bobClient.ListDatumAll(pipeline, jobID)
	require.YesError(t, err)
	require.True(t, auth.IsErrNotAuthorized(err), err.Error())

	// Finally, alice adds bob to the output repo, and now bob can call ListDatum
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(pipeline, bob, []string{auth.RepoReaderRole}))
	dis, err := bobClient.ListDatumAll(pipeline, jobID)
	require.NoError(t, err)
	files := make(map[string]struct{})
	for _, di := range dis {
		for _, f := range di.Data {
			files[path.Base(f.File.Path)] = struct{}{}
		}
	}
	require.Equal(t, map[string]struct{}{
		"file1": struct{}{},
		"file2": struct{}{},
	}, files)
}

func TestS3GatewayAuthRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
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

// TestDeactivateFSAdmin tests that users with the FS admin role can't call Deactivate
func TestDeactivateFSAdmin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
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
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	alice := tu.Robot(tu.UniqueString("alice"))
	aliceClient, adminClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, auth.RootUser)

	// alice can't extract auth tokens because she's not an admin
	_, err := aliceClient.ExtractAuthTokens(aliceClient.Ctx(), &auth.ExtractAuthTokensRequest{})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())

	// Create a token with a TTL and confirm it is extracted with an expiration
	tokenResp, err := adminClient.GetRobotToken(adminClient.Ctx(), &auth.GetRobotTokenRequest{Robot: "other", TTL: 1000})
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
					require.True(t, token.Expiration.After(time.Now()))
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
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
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
	req.Token.Expiration = &pastExpiration

	_, err = adminClient.RestoreAuthToken(adminClient.Ctx(), req)
	require.YesError(t, err)
	require.True(t, auth.IsErrExpiredToken(err))

	// restore a token with an expiration date in the future
	req.Token.HashedToken = fmt.Sprintf("%x", sha256.Sum256([]byte("expiring-token")))
	futureExpiration := time.Now().Add(10 * time.Minute)
	req.Token.Expiration = &futureExpiration

	_, err = adminClient.RestoreAuthToken(adminClient.Ctx(), req)
	require.NoError(t, err)

	aliceClient.SetAuthToken("expiring-token")
	whoAmIResp, err = aliceClient.WhoAmI(aliceClient.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)

	// Relying on time.Now is gross but the token should have a TTL in the
	// next 10 minutes
	require.True(t, whoAmIResp.Expiration.After(time.Now()))
	require.True(t, whoAmIResp.Expiration.Before(time.Now().Add(time.Duration(600)*time.Second)))
}

// TODO(Fahad): think we need to register the debug server api functions.
// TODO: This test mirrors TestDebug in src/server/pachyderm_test.go.
// Need to restructure testing such that we have the implementation of this
// test in one place while still being able to test auth enabled and disabled clusters.
func TestDebug(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	// Get all the authenticated clients at the beginning of the test.
	// GetAuthenticatedPachClient will always re-activate auth, which
	// causes PPS to rotate all the pipeline tokens. This makes the RCs
	// change and recreates all the pods, which used to race with collecting logs.
	alice := tu.Robot(tu.UniqueString("alice"))
	aliceClient, adminClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, auth.RootUser)

	dataRepo := tu.UniqueString("TestDebug_data")
	require.NoError(t, aliceClient.CreateRepo(dataRepo))

	expectedFiles, pipelines := tu.DebugFiles(t, dataRepo)

	for _, p := range pipelines {
		require.NoError(t, aliceClient.CreatePipeline(
			p,
			"",
			[]string{"bash"},
			[]string{
				fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
				"sleep 45",
			},
			&pps.ParallelismSpec{
				Constant: 1,
			},
			client.NewPFSInput(dataRepo, "/*"),
			"",
			false,
		))
	}

	commit1, err := aliceClient.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	err = aliceClient.PutFile(commit1, "file", strings.NewReader("foo"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(dataRepo, commit1.Branch.Name, commit1.ID))

	jobInfos, err := aliceClient.WaitJobSetAll(commit1.ID, false)
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

// asserts that retrieval of Pachd logs requires additional permissions granted to the PachdLogReader role
func TestGetPachdLogsRequiresPerm(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	adminClient := tu.AuthenticateClient(t, c, auth.RootUser)

	alice := tu.UniqueString("robot:alice")
	aliceClient := tu.AuthenticateClient(t, c, alice)

	aliceRepo := tu.UniqueString("alice_repo")
	err := aliceClient.CreateRepo(aliceRepo)
	require.NoError(t, err)

	// create pipeline
	alicePipeline := tu.UniqueString("pipeline_for_logs")
	err = aliceClient.CreatePipeline(
		alicePipeline,
		"", // default image: DefaultUserImage
		[]string{"bash"},
		[]string{"cp /pfs/*/* /pfs/out/"},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(aliceRepo, "/*"),
		"", // default output branch: master
		false,
	)
	require.NoError(t, err)

	// wait for the pipeline pod to come up
	time.Sleep(time.Second * 10)

	// must be authorized to access pachd logs
	pachdLogsIter := aliceClient.GetLogs("", "", nil, "", false, false, 0)
	pachdLogsIter.Next()
	require.YesError(t, pachdLogsIter.Err())
	require.True(t, strings.Contains(pachdLogsIter.Err().Error(), "is not authorized to perform this operation"))

	// alice can view the pipeline logs
	require.NoErrorWithinTRetry(t, time.Minute, func() error {
		pipelineLogsIter := aliceClient.GetLogs(alicePipeline, "", nil, "", false, false, 0)
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
	pachdLogsIter = aliceClient.GetLogs("", "", nil, "", false, false, 0)
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
//	require.NoError(t, aliceClient.CreateRepo(repo))
//
//	// alice creates a pipeline
//	pipeline := tu.UniqueString("pipeline")
//	require.NoError(t, aliceClient.CreatePipeline(
//		pipeline,
//		"", // default image: DefaultUserImage
//		[]string{"bash"},
//		[]string{"cp /pfs/*/* /pfs/out/"},
//		&pps.ParallelismSpec{Constant: 1},
//		client.NewPFSInput(repo, "/*"),
//		"", // default output branch: master
//		false,
//	))
//
//	// alice commits to the input repos, and the pipeline runs successfully
//	err := aliceClient.PutFile(repo, "master", "/file1", strings.NewReader("test"))
//	require.NoError(t, err)
//	commitIter, err := aliceClient.WaitCommit(pipeline, "master", "")
//	require.NoError(t, err)
//	require.NoErrorWithinT(t, 60*time.Second, func() error {
//		_, err := commitIter.Next()
//		return err
//	})
//
//	// bob cannot call GetLogs
//	iter := bobClient.GetLogs(pipeline, "", nil, "", false, false, 0)
//	require.False(t, iter.Next())
//	require.YesError(t, iter.Err())
//	require.True(t, auth.IsErrNotAuthorized(iter.Err()), iter.Err().Error())
//
//	// bob also can't call GetLogs for the master process
//	iter = bobClient.GetLogs(pipeline, "", nil, "", true, false, 0)
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
//	iter = bobClient.GetLogs(pipeline, "", nil, "", false, false, 0)
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
//	iter = bobClient.GetLogs(pipeline, "", nil, "", false, false, 0)
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
//	iter = bobClient.GetLogs(pipeline, "", nil, "", false, false, 0)
//	iter.Next()
//	require.NoError(t, iter.Err())
//
//	// bob can also call GetLogs for the master process
//	iter = bobClient.GetLogs(pipeline, "", nil, "", true, false, 0)
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
//	require.NoError(t, aliceClient.CreateRepo(repo))
//
//	// alice creates a pipeline (we must enable stats for InspectDatum, which
//	// means calling the grpc client function directly)
//	pipeline := tu.UniqueString("alice")
//	_, err := aliceClient.PpsAPIClient.CreatePipeline(aliceClient.Ctx(),
//		&pps.CreatePipelineRequest{
//			Pipeline: &pps.Pipeline{Name: pipeline},
//			Transform: &pps.Transform{
//				Cmd:   []string{"bash"},
//				Stdin: []string{"cp /pfs/*/* /pfs/out/"},
//			},
//			ParallelismSpec: &pps.ParallelismSpec{Constant: 1},
//			Input:           client.NewPFSInput(repo, "/*"),
//		})
//	require.NoError(t, err)
//
//	// alice commits to the input repo, and the pipeline runs successfully
//	err = aliceClient.PutFile(repo, "master", "/file1", strings.NewReader("test"))
//	require.NoError(t, err)
//	commitItr, err := aliceClient.WaitCommit(pipeline, "master", "")
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
//	iter := aliceClient.GetLogs(pipeline, jobID, nil, "", false, false, 0)
//	require.True(t, iter.Next())
//	require.NoError(t, iter.Err())
//
//	iter = aliceClient.GetLogs(pipeline, jobID, nil, "", true, false, 0)
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
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	alice, bob := tu.Robot(tu.UniqueString("alice")), tu.Robot(tu.UniqueString("bob"))
	aliceClient, bobClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, bob)

	// alice creates a repo, and adds bob as a reader
	repo := tu.UniqueString(t.Name())
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(repo, bob, []string{auth.RepoReaderRole}))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, bob, auth.RepoReaderRole), tu.GetRepoRoleBinding(t, aliceClient, repo))
	commit := client.NewCommit(repo, "master", "")

	// bob creates a pipeline
	pipeline := tu.UniqueString("bob-pipeline")
	require.NoError(t, bobClient.CreatePipeline(
		pipeline,
		"", // default image: DefaultUserImage
		[]string{"bash"},
		[]string{"cp /pfs/*/* /pfs/out/"},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(repo, "/*"),
		"", // default output branch: master
		false,
	))
	require.Equal(t,
		tu.BuildBindings(bob, auth.RepoOwnerRole, tu.Pl(pipeline), auth.RepoWriterRole), tu.GetRepoRoleBinding(t, bobClient, pipeline))
	// bob adds alice as a reader of the pipeline's output repo, so alice can
	// flush input commits (which requires her to inspect commits in the output)
	// and update the pipeline
	require.NoError(t, bobClient.ModifyRepoRoleBinding(pipeline, alice, []string{auth.RepoWriterRole}))
	require.Equal(t,
		tu.BuildBindings(bob, auth.RepoOwnerRole, alice, auth.RepoWriterRole, tu.Pl(pipeline), auth.RepoWriterRole),
		tu.GetRepoRoleBinding(t, bobClient, pipeline))

	// alice commits to the input repo, and the pipeline runs successfully
	require.NoError(t, aliceClient.PutFile(commit, "/file", strings.NewReader("test")))
	require.NoErrorWithinT(t, 45*time.Second, func() error {
		_, err := bobClient.WaitCommit(pipeline, "master", commit.ID)
		return err
	})

	// alice removes bob as a reader of her repo, and then commits to the input
	// repo, but bob's pipeline still runs (it has its own principal--it doesn't
	// inherit bob's privileges)
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(repo, bob, []string{}))
	require.Equal(t,
		tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Pl(pipeline), auth.RepoReaderRole), tu.GetRepoRoleBinding(t, aliceClient, repo))
	require.NoError(t, aliceClient.PutFile(commit, "/file", strings.NewReader("test")))
	require.NoErrorWithinT(t, 45*time.Second, func() error {
		_, err := aliceClient.WaitCommit(pipeline, "master", commit.ID)
		return err
	})

	// alice revokes the pipeline's access to 'repo' directly, and the pipeline
	// stops running
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(repo, tu.Pl(pipeline), []string{}))
	require.NoError(t, aliceClient.PutFile(commit, "/file", strings.NewReader("test")))
	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		_, err := aliceClient.WaitCommit(pipeline, "master", commit.ID)
		require.NoError(t, err)
	}()
	select {
	case <-doneCh:
		t.Fatal("pipeline should not be able to finish with no access")
	case <-time.After(45 * time.Second):
	}

	// alice updates bob's pipline, but the pipeline still doesn't run
	require.NoError(t, aliceClient.CreatePipeline(
		pipeline,
		"", // default image: DefaultUserImage
		[]string{"bash"},
		[]string{"cp /pfs/*/* /pfs/out/"},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(repo, "/*"),
		"", // default output branch: master
		true,
	))
	require.NoError(t, aliceClient.PutFile(commit, "/file", strings.NewReader("test")))
	doneCh = make(chan struct{})
	go func() {
		defer close(doneCh)
		_, err := aliceClient.WaitCommit(pipeline, "master", commit.ID)
		require.NoError(t, err)
	}()
	select {
	case <-doneCh:
		t.Fatal("pipeline should not be able to finish with no access")
	case <-time.After(45 * time.Second):
	}

	// alice restores the pipeline's access to its input repo, and now the
	// pipeline runs successfully
	require.NoError(t, aliceClient.ModifyRepoRoleBinding(repo, tu.Pl(pipeline), []string{auth.RepoReaderRole}))
	require.NoErrorWithinT(t, 45*time.Second, func() error {
		_, err := aliceClient.WaitCommit(pipeline, "master", commit.ID)
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
	c, ns := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	alice := tu.Robot(tu.UniqueString("alice"))
	c = tu.AuthenticateClient(t, c, alice)

	// Create input repo w/ initial commit
	repo := tu.UniqueString(t.Name())
	require.NoError(t, c.CreateRepo(repo))
	err := c.PutFile(client.NewCommit(repo, "master", ""), "/file.1", strings.NewReader("1"))
	require.NoError(t, err)

	// Create pipeline
	pipeline := tu.UniqueString("pipeline")
	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pipeline),
			Transform: &pps.Transform{
				Image: "", // default image: DefaultUserImage
				Cmd:   []string{"bash"},
				Stdin: []string{"cp /pfs/*/* /pfs/out"},
			},
			ParallelismSpec: &pps.ParallelismSpec{Constant: 1},
			Input:           client.NewPFSInput(repo, "/*"),
			Autoscaling:     true,
		})
	require.NoError(t, err)

	// Wait for pipeline to process input commit & go into standby
	require.NoErrorWithinT(t, 30*time.Second, func() error {
		_, err := c.WaitCommit(pipeline, "master", "")
		return err
	})
	require.NoErrorWithinTRetry(t, 30*time.Second, func() error {
		pi, err := c.InspectPipeline(pipeline, false)
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
	err = c.PutFile(client.NewCommit(repo, "master", ""), "/file.2", strings.NewReader("1"))
	require.NoError(t, err)
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := c.WaitCommit(pipeline, "master", "")
		return err
	})
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
	// TODO: Reenable when finishing job state is transactional.
	t.Skip("Job state does not get finished in a transaction, so stats commit is left open")
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	alice := tu.Robot(tu.UniqueString("alice"))
	aliceClient, rootClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, auth.RootUser)

	// Create input repo w/ initial commit
	repo := tu.UniqueString(t.Name())
	commit := client.NewCommit(repo, "master", "")
	require.NoError(t, aliceClient.CreateRepo(repo))
	err := aliceClient.PutFile(commit, "/file.1", strings.NewReader("1"))
	require.NoError(t, err)

	// Create pipeline
	pipeline := tu.UniqueString("pipeline")
	require.NoError(t, aliceClient.CreatePipeline(
		pipeline,
		"", // default image: DefaultUserImage
		[]string{"bash"},
		[]string{
			"sleep 10",
			"cp /pfs/*/* /pfs/out/",
		},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(repo, "/*"),
		"", // default output branch: master
		false,
	))

	// Revoke pipeline's access to output repo while 'sleep 10' is running (so
	// that it fails)
	require.NoError(t, rootClient.ModifyRepoRoleBinding(repo, fmt.Sprintf("pipeline:%s", pipeline), []string{}))

	// make sure the pipeline either fails or restarts RC & finishes
	require.NoErrorWithinT(t, 30*time.Second, func() error {
		_, err := aliceClient.WaitCommit(pipeline, "master", commit.ID)
		return err
	})

	// make sure the pipeline is failed
	pi, err := rootClient.InspectPipeline(pipeline, false)
	require.NoError(t, err)
	require.Equal(t, pps.PipelineState_PIPELINE_FAILURE, pi.State)
}

func TestPreActivationCronPipelinesKeepRunningAfterActivation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
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
	require.NoError(t, aliceClient.CreatePipeline(
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
	require.NoError(t, aliceClient.CreatePipeline(
		pipeline2,
		"",
		[]string{"/bin/bash"},
		[]string{"cp " + fmt.Sprintf("/pfs/%s/*", pipeline1) + " /pfs/out/"},
		nil,
		client.NewPFSInput(pipeline1, "/*"),
		"",
		false,
	))

	// subscribe to the pipeline2 cron repo and wait for inputs
	repo := client.NewRepo(pipeline2)
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
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	alice := tu.Robot(tu.UniqueString("alice"))
	aliceClient, rootClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, auth.RootUser)

	// alice creates a repo
	repo := tu.UniqueString("TestPipelinesRunAfterExpiration")
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.Equal(t, tu.BuildBindings(alice, auth.RepoOwnerRole), tu.GetRepoRoleBinding(t, aliceClient, repo))

	// alice creates a pipeline
	pipeline := tu.UniqueString("alice-pipeline")
	require.NoError(t, aliceClient.CreatePipeline(
		pipeline,
		"", // default image: DefaultUserImage
		[]string{"bash"},
		[]string{fmt.Sprintf("cp /pfs/%s/* /pfs/out/", repo)},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(repo, "/*"),
		"",    // default output branch: master
		false, // no update
	))
	require.OneOfEquals(t, pipeline, tu.PipelineNames(t, aliceClient))
	// check that alice owns the output repo too,
	require.Equal(t, tu.BuildBindings(alice, auth.RepoOwnerRole, tu.Pl(pipeline), auth.RepoWriterRole), tu.GetRepoRoleBinding(t, aliceClient, pipeline))

	// Make sure alice's pipeline runs successfully
	commit, err := aliceClient.StartCommit(repo, "master")
	require.NoError(t, err)
	err = aliceClient.PutFile(commit, tu.UniqueString("/file1"),
		strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.Branch.Name, commit.ID))
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := aliceClient.WaitCommit(pipeline, "master", commit.ID)
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
	commit, err = rootClient.StartCommit(repo, "master")
	require.NoError(t, err)
	err = rootClient.PutFile(commit, tu.UniqueString("/file2"),
		strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, rootClient.FinishCommit(repo, commit.Branch.Name, commit.ID))
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := rootClient.WaitCommit(pipeline, "master", commit.ID)
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
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	alice := tu.Robot(tu.UniqueString("alice"))
	aliceClient, rootClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, auth.RootUser)

	// alice creates a pipeline
	repo := tu.UniqueString("TestDeleteAllAfterDeactivate")
	pipeline := tu.UniqueString("pipeline")
	require.NoError(t, aliceClient.CreateRepo(repo))
	require.NoError(t, aliceClient.CreatePipeline(
		pipeline,
		"", // default image: DefaultUserImage
		[]string{"bash"},
		[]string{fmt.Sprintf("cp /pfs/%s/* /pfs/out/", repo)},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(repo, "/*"),
		"", // default output branch: master
		false,
	))

	// alice makes an input commit
	commit, err := aliceClient.StartCommit(repo, "master")
	require.NoError(t, err)
	err = aliceClient.PutFile(commit, "/file1", strings.NewReader("test data"))
	require.NoError(t, err)
	require.NoError(t, aliceClient.FinishCommit(repo, commit.Branch.Name, commit.ID))

	// make sure the pipeline runs
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		_, err := aliceClient.WaitCommit(pipeline, "master", commit.ID)
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

// TestGetRobotTokenErrorNonAdminUser tests that non-admin users can't call
// GetRobotToken
func TestGetRobotTokenErrorNonAdminUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	alice := tu.Robot(tu.UniqueString("alice"))
	aliceClient := tu.AuthenticateClient(t, c, alice)
	resp, err := aliceClient.GetRobotToken(aliceClient.Ctx(), &auth.GetRobotTokenRequest{
		Robot: tu.UniqueString("t-1000"),
	})
	require.Nil(t, resp)
	require.YesError(t, err)
	require.Matches(t, "needs permissions \\[CLUSTER_AUTH_GET_ROBOT_TOKEN\\] on CLUSTER", err.Error())
}
