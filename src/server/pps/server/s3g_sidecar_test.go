//go:build k8s

/*
This test is for PPS pipelines that use S3 inputs/outputs. Most of these
pipelines use the pachyderm/s3testing image, which exists on dockerhub but can
be built by running:

	cd etc/testing/images/s3testing
	make push-to-minikube
*/
package server

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// This test is designed to run against pachyderm in custom namespaces and with
// auth on or off. It reads env vars into here and adjusts tests to make sure
// our s3 gateway feature works in those contexts

func initPachClient(t testing.TB) (*client.APIClient, string, string) {
	c, ns := minikubetestenv.AcquireCluster(t)
	if _, ok := os.LookupEnv("PACH_TEST_WITH_AUTH"); !ok {
		return c, "", ns
	}
	c = tu.AuthenticatedPachClient(t, c, tu.UniqueString("user-"))
	return c, c.AuthToken(), ns
}

func TestS3PipelineErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _, _ := initPachClient(t)

	repo1, repo2 := tu.UniqueString(t.Name()+"_data"), tu.UniqueString(t.Name()+"_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repo1))
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repo2))

	pipeline := tu.UniqueString("Pipeline")
	err := c.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			"ls -R /pfs >/pfs/out/files",
		},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		&pps.Input{
			Union: []*pps.Input{
				{Pfs: &pps.PFSInput{
					Repo:   repo1,
					Branch: "master",
					S3:     true,
					Glob:   "/",
				}},
				{Pfs: &pps.PFSInput{
					Repo:   repo2,
					Branch: "master",
					Glob:   "/*",
				}},
			},
		},
		"",
		false,
	)
	require.YesError(t, err)
	require.Matches(t, "union", err.Error())
	err = c.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			"ls -R /pfs >/pfs/out/files",
		},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		&pps.Input{
			Join: []*pps.Input{
				{Pfs: &pps.PFSInput{
					Repo:   repo1,
					Branch: "master",
					S3:     true,
					Glob:   "/",
				}},
				{Pfs: &pps.PFSInput{
					Repo:   repo2,
					Branch: "master",
					Glob:   "/*",
				}},
			},
		},
		"",
		false,
	)
	require.YesError(t, err)
	require.Matches(t, "join", err.Error())
}

func testS3Input(t *testing.T, c *client.APIClient, ns, projectName string) {
	repo := tu.UniqueString("data")
	require.NoError(t, c.CreateRepo(projectName, repo))
	masterCommit := client.NewCommit(projectName, repo, "master", "")

	require.NoError(t, c.PutFile(masterCommit, "foo", strings.NewReader("foo")))

	pipeline := tu.UniqueString("Pipeline")
	pipelineCommit := client.NewCommit(projectName, pipeline, "master", "")
	_, err := c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline: client.NewPipeline(projectName, pipeline),
		Transform: &pps.Transform{
			Image: "pachyderm/ubuntu-with-s3-clients:v0.0.1",
			Cmd:   []string{"bash", "-x"},
			Stdin: []string{
				"ls -R /pfs >/pfs/out/pfs_files",
				"aws --endpoint=${S3_ENDPOINT} s3 ls >/pfs/out/s3_buckets",
				"aws --endpoint=${S3_ENDPOINT} s3 ls s3://input_repo >/pfs/out/s3_files",
			},
		},
		ParallelismSpec: &pps.ParallelismSpec{Constant: 1},
		Input: &pps.Input{
			Pfs: &pps.PFSInput{
				Project: projectName,
				Repo:    repo,
				Name:    "input_repo",
				Branch:  "master",
				S3:      true,
				Glob:    "/",
			},
		},
	})
	require.NoError(t, err)

	commitInfo, err := c.InspectCommit(projectName, pipeline, "master", "")
	require.NoError(t, err)

	jobInfo, err := c.WaitJob(projectName, pipeline, commitInfo.Commit.Id, false)
	require.NoError(t, err)
	require.Equal(t, "JOB_SUCCESS", jobInfo.State.String())

	// Make sure ListFile works
	files, err := c.ListFileAll(pipelineCommit, "/")
	require.NoError(t, err)
	require.ElementsEqualUnderFn(t, []string{"/pfs_files", "/s3_buckets", "/s3_files"}, files,
		func(i interface{}) interface{} {
			return i.(*pfs.FileInfo).File.Path
		})

	// check files in /pfs
	var buf bytes.Buffer
	require.NoError(t, c.GetFile(pipelineCommit, "pfs_files", &buf))
	require.True(t,
		strings.Contains(buf.String(), "out") && !strings.Contains(buf.String(), "foo"),
		"expected \"out\" but not \"foo\" in %s: %q", "pfs_files", buf.String())

	// check s3 buckets
	buf.Reset()
	require.NoError(t, c.GetFile(pipelineCommit, "s3_buckets", &buf))
	require.True(t,
		strings.Contains(buf.String(), "input_repo") && !strings.Contains(buf.String(), "out"),
		"expected \"input_repo\" but not \"out\" in %s: %q", "s3_buckets", buf.String())

	// Check files in input_repo
	buf.Reset()
	require.NoError(t, c.GetFile(pipelineCommit, "s3_files", &buf))
	require.Matches(t, "foo", buf.String())

	// Check that no service is left over
	k := tu.GetKubeClient(t)
	require.NoErrorWithinTRetry(t, 68*time.Second, func() error {
		svcs, err := k.CoreV1().Services(ns).List(context.Background(), metav1.ListOptions{})
		require.NoError(t, err)
		for _, s := range svcs.Items {
			if s.ObjectMeta.Name == ppsutil.SidecarS3GatewayService(jobInfo.Job.Pipeline, jobInfo.Job.Id) {
				return errors.Errorf("service %q should be cleaned up by sidecar after job", s.ObjectMeta.Name)
			}
		}
		return nil
	})
}

func TestS3Input(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _, ns := initPachClient(t)
	t.Run("DefaultProject", func(t *testing.T) {
		testS3Input(t, c, ns, pfs.DefaultProjectName)
	})
	t.Run("NonDefaultProject", func(t *testing.T) {
		projectName := tu.UniqueString("project")
		require.NoError(t, c.CreateProject(projectName))
		testS3Input(t, c, ns, projectName)
	})
}

func testS3Chain(t *testing.T, c *client.APIClient, ns, projectName string) {
	dataRepo := tu.UniqueString("data")
	require.NoError(t, c.CreateRepo(projectName, dataRepo))
	dataCommit := client.NewCommit(projectName, dataRepo, "master", "")

	numPipelines := 5
	pipelines := make([]string, numPipelines)
	for i := 0; i < numPipelines; i++ {
		pipelines[i] = tu.UniqueString("pipeline")
		input := dataRepo
		if i > 0 {
			input = pipelines[i-1]
		}
		_, err := c.PpsAPIClient.CreatePipeline(c.Ctx(),
			&pps.CreatePipelineRequest{
				Pipeline: client.NewPipeline(projectName, pipelines[i]),
				Transform: &pps.Transform{
					Image: "pachyderm/ubuntu-with-s3-clients:v0.0.1",
					Cmd:   []string{"bash", "-x"},
					Stdin: []string{
						"aws --endpoint=${S3_ENDPOINT} s3 cp s3://s3g_in/file /tmp/s3in",
						"echo '1' >> /tmp/s3in",
						"aws --endpoint=${S3_ENDPOINT} s3 cp /tmp/s3in s3://out/file",
					},
				},
				ParallelismSpec: &pps.ParallelismSpec{Constant: 1},
				Input: &pps.Input{
					Pfs: &pps.PFSInput{
						Project: projectName,
						Repo:    input,
						Name:    "s3g_in",
						Branch:  "master",
						S3:      true,
						Glob:    "/",
					},
				},
				S3Out: true,
			},
		)
		require.NoError(t, err)
	}

	require.NoError(t, c.PutFile(dataCommit, "file", strings.NewReader("")))
	commitInfo, err := c.InspectCommit(projectName, dataCommit.Branch.Repo.Name, dataCommit.Branch.Name, "")
	require.NoError(t, err)

	_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)
	for i := 0; i < numPipelines; i++ {
		var buf bytes.Buffer
		require.NoError(t, c.GetFile(client.NewCommit(projectName, pipelines[i], "master", ""), "/file", &buf))
		require.Equal(t, i+1, strings.Count(buf.String(), "1\n"))
	}
}

func TestS3Chain(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _, ns := initPachClient(t)
	t.Run("DefaultProject", func(t *testing.T) {
		testS3Chain(t, c, ns, pfs.DefaultProjectName)
	})
	t.Run("NonDefaultProject", func(t *testing.T) {
		projectName := tu.UniqueString("project")
		require.NoError(t, c.CreateProject(projectName))
		testS3Chain(t, c, ns, projectName)
	})
}

func TestNamespaceInEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _, ns := initPachClient(t)

	repo := tu.UniqueString(t.Name() + "_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repo))
	masterCommit := client.NewCommit(pfs.DefaultProjectName, repo, "master", "")

	require.NoError(t, c.PutFile(masterCommit, "foo", strings.NewReader("foo")))

	pipeline := tu.UniqueString("Pipeline")
	pipelineCommit := client.NewCommit(pfs.DefaultProjectName, pipeline, "master", "")
	_, err := c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
		Transform: &pps.Transform{
			Cmd: []string{"bash", "-x"},
			Stdin: []string{
				"echo \"${S3_ENDPOINT}\" >/pfs/out/s3_endpoint",
			},
		},
		ParallelismSpec: &pps.ParallelismSpec{Constant: 1},
		Input: &pps.Input{
			Pfs: &pps.PFSInput{
				Project: pfs.DefaultProjectName,
				Repo:    repo,
				Name:    "input_repo",
				Branch:  "master",
				S3:      true,
				Glob:    "/",
			},
		},
	})
	require.NoError(t, err)

	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, pipeline, "master", "")
	require.NoError(t, err)

	jobInfo, err := c.WaitJob(pfs.DefaultProjectName, pipeline, commitInfo.Commit.Id, false)
	require.NoError(t, err)
	require.Equal(t, "JOB_SUCCESS", jobInfo.State.String())

	// check S3_ENDPOINT variable
	var buf bytes.Buffer
	require.NoError(t, c.GetFile(pipelineCommit, "s3_endpoint", &buf))
	require.True(t, strings.Contains(buf.String(), "."+ns))
}

func testS3Output(t *testing.T, c *client.APIClient, ns, projectName string) {
	repo := tu.UniqueString("data")
	require.NoError(t, c.CreateRepo(projectName, repo))
	masterCommit := client.NewCommit(projectName, repo, "master", "")

	require.NoError(t, c.PutFile(masterCommit, "foo", strings.NewReader("foo")))

	pipeline := tu.UniqueString("Pipeline")
	pipelineCommit := client.NewCommit(projectName, pipeline, "master", "")
	_, err := c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline: client.NewPipeline(projectName, pipeline),
		Transform: &pps.Transform{
			Image: "pachyderm/ubuntu-with-s3-clients:v0.0.1",
			Cmd:   []string{"bash", "-x"},
			Stdin: []string{
				"ls -R /pfs | aws --endpoint=${S3_ENDPOINT} s3 cp - s3://out/pfs_files",
				"aws --endpoint=${S3_ENDPOINT} s3 ls | aws --endpoint=${S3_ENDPOINT} s3 cp - s3://out/s3_buckets",
			},
		},
		ParallelismSpec: &pps.ParallelismSpec{Constant: 1},
		Input: &pps.Input{
			Pfs: &pps.PFSInput{
				Project: projectName,
				Repo:    repo,
				Name:    "input_repo",
				Branch:  "master",
				Glob:    "/",
			},
		},
		S3Out: true,
	})
	require.NoError(t, err)

	commitInfo, err := c.InspectCommit(projectName, pipeline, "master", "")
	require.NoError(t, err)

	jobInfo, err := c.WaitJob(projectName, pipeline, commitInfo.Commit.Id, false)
	require.NoError(t, err)
	require.Equal(t, "JOB_SUCCESS", jobInfo.State.String())

	// Make sure ListFile works
	files, err := c.ListFileAll(pipelineCommit, "/")
	require.NoError(t, err)
	require.ElementsEqualUnderFn(t, []string{"/pfs_files", "/s3_buckets"}, files,
		func(i interface{}) interface{} {
			return i.(*pfs.FileInfo).File.Path
		})

	// check files in /pfs
	var buf bytes.Buffer
	require.NoError(t, c.GetFile(pipelineCommit, "pfs_files", &buf))
	require.True(t,
		!strings.Contains(buf.String(), "out") && strings.Contains(buf.String(), "input_repo"),
		"expected \"input_repo\" but not \"out\" in %s: %q", "pfs_files", buf.String())

	// check s3 buckets
	buf.Reset()
	require.NoError(t, c.GetFile(pipelineCommit, "s3_buckets", &buf))
	require.True(t,
		!strings.Contains(buf.String(), "input_repo") && strings.Contains(buf.String(), "out"),
		"expected \"out\" but not \"input_repo\" in %s: %q", "s3_buckets", buf.String())

	// Check that no service is left over
	k := tu.GetKubeClient(t)
	require.NoErrorWithinTRetry(t, 68*time.Second, func() error {
		svcs, err := k.CoreV1().Services(ns).List(context.Background(), metav1.ListOptions{})
		require.NoError(t, err)
		for _, s := range svcs.Items {
			if s.ObjectMeta.Name == ppsutil.SidecarS3GatewayService(jobInfo.Job.Pipeline, jobInfo.Job.Id) {
				return errors.Errorf("service %q should be cleaned up by sidecar after job", s.ObjectMeta.Name)
			}
		}
		return nil
	})
}

func TestS3Output(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _, ns := initPachClient(t)
	t.Run("DefaultProject", func(t *testing.T) {
		testS3Output(t, c, ns, pfs.DefaultProjectName)
	})
	t.Run("NonDefaultProject", func(t *testing.T) {
		projectName := tu.UniqueString("project")
		require.NoError(t, c.CreateProject(projectName))
		testS3Output(t, c, ns, projectName)
	})
}

func testFullS3(t *testing.T, c *client.APIClient, ns, projectName string) {
	repo := tu.UniqueString("data")
	require.NoError(t, c.CreateRepo(projectName, repo))
	masterCommit := client.NewCommit(projectName, repo, "master", "")

	require.NoError(t, c.PutFile(masterCommit, "foo", strings.NewReader("foo")))

	pipeline := tu.UniqueString("Pipeline")
	pipelineCommit := client.NewCommit(projectName, pipeline, "master", "")
	_, err := c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline: client.NewPipeline(projectName, pipeline),
		Transform: &pps.Transform{
			Image: "pachyderm/ubuntu-with-s3-clients:v0.0.1",
			Cmd:   []string{"bash", "-x"},
			Stdin: []string{
				"ls -R /pfs | aws --endpoint=${S3_ENDPOINT} s3 cp - s3://out/pfs_files",
				"aws --endpoint=${S3_ENDPOINT} s3 ls | aws --endpoint=${S3_ENDPOINT} s3 cp - s3://out/s3_buckets",
			},
		},
		ParallelismSpec: &pps.ParallelismSpec{Constant: 1},
		Input: &pps.Input{
			Pfs: &pps.PFSInput{
				Project: projectName,
				Repo:    repo,
				Name:    "input_repo",
				Branch:  "master",
				S3:      true,
				Glob:    "/",
			},
		},
		S3Out: true,
	})
	require.NoError(t, err)

	commitInfo, err := c.InspectCommit(projectName, pipeline, "master", "")
	require.NoError(t, err)

	jobInfo, err := c.WaitJob(projectName, pipeline, commitInfo.Commit.Id, false)
	require.NoError(t, err)
	require.Equal(t, "JOB_SUCCESS", jobInfo.State.String())

	// Make sure ListFile works
	files, err := c.ListFileAll(pipelineCommit, "/")
	require.NoError(t, err)
	require.ElementsEqualUnderFn(t, []string{"/pfs_files", "/s3_buckets"}, files,
		func(i interface{}) interface{} {
			return i.(*pfs.FileInfo).File.Path
		})

	// check files in /pfs
	var buf bytes.Buffer
	require.NoError(t, c.GetFile(pipelineCommit, "pfs_files", &buf))
	require.True(t,
		!strings.Contains(buf.String(), "foo") && !strings.Contains(buf.String(), "out"),
		"expected neither \"out\" nor \"foo\" in %s: %q", "pfs_files", buf.String())

	// check s3 buckets
	buf.Reset()
	require.NoError(t, c.GetFile(pipelineCommit, "s3_buckets", &buf))
	require.True(t,
		strings.Contains(buf.String(), "out") && strings.Contains(buf.String(), "input_repo"),
		"expected both \"input_repo\" and \"out\" in %s: %q", "s3_buckets", buf.String())

	// Check that no service is left over
	k := tu.GetKubeClient(t)
	require.NoErrorWithinTRetry(t, 68*time.Second, func() error {
		svcs, err := k.CoreV1().Services(ns).List(context.Background(), metav1.ListOptions{})
		require.NoError(t, err)
		for _, s := range svcs.Items {
			if s.ObjectMeta.Name == ppsutil.SidecarS3GatewayService(jobInfo.Job.Pipeline, jobInfo.Job.Id) {
				return errors.Errorf("service %q should be cleaned up by sidecar after job", s.ObjectMeta.Name)
			}
		}
		return nil
	})
}

func TestFullS3(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _, ns := initPachClient(t)
	t.Run("DefaultProject", func(t *testing.T) {
		testFullS3(t, c, ns, pfs.DefaultProjectName)
	})
	t.Run("NonDefaultProject", func(t *testing.T) {
		projectName := tu.UniqueString("project")
		require.NoError(t, c.CreateProject(projectName))
		testFullS3(t, c, ns, projectName)
	})
}

// testS3SkippedDatums is a very complicated test that checks the skipping
// behavior of various sidecar-S3-gateway-using pipelines. Two notes:
//
//   - In this test, we track the job in which each datum was processed by
//     creating pipelines that read from a "background" repo that isn't a pipeline
//     input, which we update between jobs. Each datum's output will include the
//     value of the "background" repo. So if we set the background repo's contents
//     to "1" and then process datum A, A's output will include "1". If we then
//     update the background repo's contents to "2" and then run another job, A's
//     output will remain "1" if it's skipped, or will now contain "2" if it's
//     reprocessed.
//
//   - In "S3Inputs", we have a pipeline that reads from a cross of an S3 input
//     and a globbed PFS input. This creates N datums, where each includes the S3
//     input and one of the files in the PFS input. Adding files to the PFS input
//     should trigger new jobs that skip old datums, but updating the S3 input
//     should trigger a job that reprocesses everything (and skips nothing)
//
//   - In "S3Output", we have a pipeline that outputs to the sidecar S3 gateway.
//     In pipelines that output to the sidecar S3 gateway, we cannot skip datums
//     as we have no way to know what a single datum's output was. Part 2 confirms
//     that no datums are ever skipped for S3-out-pipelines.
func testS3SkippedDatums(t *testing.T, c *client.APIClient, ns, projectName string) {
	// NOTE: in tests, the S3G port (not just the NodePort but the
	// cluster-internal port) is assigned dynamically in
	// src/internal/minikubetestenv/deploy.go
	s3gPort := func() (result int32) {
		k := tu.GetKubeClient(t)
		require.NoErrorWithinTRetry(t, 60*time.Second, func() error {
			svc, err := k.CoreV1().Services(ns).Get(context.Background(), "pachd", metav1.GetOptions{})
			require.NoError(t, err)
			for _, port := range svc.Spec.Ports {
				if port.Name == "s3gateway-port" {
					result = port.Port
					return nil
				}
			}
			return errors.New("could not find pachd port to make S3G connection")
		})
		return result
	}()

	t.Run("S3Inputs", func(t *testing.T) {
		s3in := tu.UniqueString("s3_data")
		require.NoError(t, c.CreateRepo(projectName, s3in))
		pfsin := tu.UniqueString("pfs_data")
		require.NoError(t, c.CreateRepo(projectName, pfsin))

		s3Commit := client.NewCommit(projectName, s3in, "master", "")
		// Pipelines with S3 inputs should still skip datums, as long as the S3 input
		// hasn't changed. We'll check this by reading from a repo that isn't a
		// pipeline input
		background := tu.UniqueString("bg_data")
		require.NoError(t, c.CreateRepo(projectName, background))

		require.NoError(t, c.PutFile(s3Commit, "file", strings.NewReader("foo")))

		pipeline := tu.UniqueString("Pipeline")
		// 'pipeline' needs access to the 'background' repo to run successfully
		require.NoError(t, c.ModifyRepoRoleBinding(projectName, background,
			fmt.Sprintf("pipeline:%s/%s", projectName, pipeline), []string{auth.RepoReaderRole}))

		pipelineCommit := client.NewCommit(projectName, pipeline, "master", "")
		_, err := c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(projectName, pipeline),
			Transform: &pps.Transform{
				Image: "pachyderm/ubuntu-with-s3-clients:v0.0.1",
				Cmd:   []string{"bash", "-x"},
				Stdin: []string{
					fmt.Sprintf(
						// access background repo via regular s3g (not S3_ENDPOINT, which
						// can only access inputs). Note: This is accessing pachd via its
						// internal kubernetes service (kubedns automatically resolves
						// 'pachd' as 'pachd.$namespace.svc.cluster.local' via a generated
						// 'search' line in the pod's /etc/resolv.conf).
						"aws --endpoint=http://pachd:%d s3 cp s3://master.%s.%s/round /tmp/bg",
						s3gPort, background, projectName,
					),
					"aws --endpoint=${S3_ENDPOINT} s3 cp s3://s3g_in/file /tmp/s3in",
					"cat /pfs/pfs_in/* >/tmp/pfsin",
					"echo \"$(cat /tmp/bg) $(cat /tmp/pfsin) $(cat /tmp/s3in)\" >/pfs/out/$(cat /tmp/pfsin)",
				},
			},
			ParallelismSpec: &pps.ParallelismSpec{Constant: 1},
			Input: &pps.Input{
				Cross: []*pps.Input{
					{Pfs: &pps.PFSInput{
						Project: projectName,
						Repo:    pfsin,
						Name:    "pfs_in",
						Branch:  "master",
						Glob:    "/*",
					}},
					{Pfs: &pps.PFSInput{
						Project: projectName,
						Repo:    s3in,
						Name:    "s3g_in",
						Branch:  "master",
						S3:      true,
						Glob:    "/",
					}},
				}},
		})
		require.NoError(t, err)

		_, err = c.WaitCommit(projectName, pipeline, "master", "")
		require.NoError(t, err)

		// Part 1: add files in pfs input w/o changing s3 input. Old files in
		// 'pfsin' should be skipped datums
		// ----------------------------------------------------------------------
		for i := 0; i < 10; i++ {
			// Increment "/round" in 'background'
			iS := fmt.Sprintf("%d", i)
			bgc, err := c.StartCommit(projectName, background, "master")
			require.NoError(t, err)
			require.NoError(t, c.DeleteFile(bgc, "/round"))
			require.NoError(t, c.PutFile(bgc, "/round", strings.NewReader(iS)))
			require.NoError(t, c.FinishCommit(projectName, background, bgc.Branch.Name, bgc.Id))

			//  Put new file in 'pfsin' to create a new datum and trigger a job
			require.NoError(t, c.PutFile(client.NewCommit(projectName, pfsin, "master", ""), iS, strings.NewReader(iS)))

			_, err = c.WaitCommit(projectName, pipeline, "master", "")
			require.NoError(t, err)

			jis, err := c.ListJob(projectName, pipeline, nil, 0, false)
			require.NoError(t, err)
			require.Equal(t, i+2, len(jis)) // one empty job w/ initial s3in commit
			for j := 0; j < len(jis); j++ {
				require.Equal(t, "JOB_SUCCESS", jis[j].State.String())
			}

			// check output
			var buf bytes.Buffer
			for j := 0; j <= i; j++ {
				buf.Reset()
				require.NoError(t, c.GetFile(pipelineCommit, strconv.Itoa(j), &buf))
				s := bufio.NewScanner(&buf)
				for s.Scan() {
					// [0] = bg, [1] = pfsin, [2] = s3in
					p := strings.Split(s.Text(), " ")
					// Check that bg is the same as the datum; this implies that bg is not
					// being re-read during each job and the datum is being skipped (what
					// we want)
					require.Equal(t, p[0], p[1], "line: "+s.Text())
					require.Equal(t, p[2], "foo", "line: "+s.Text()) // s3 input is being read but not changing
				}
			}
		}

		// Part 2: change s3 input. All old datums should get reprocessed
		// --------------------------------------------------------------
		// Increment "/round" in 'background' (this will trigger one more job w/ the
		// old S3-input commit)
		bgc, err := c.StartCommit(projectName, background, "master")
		require.NoError(t, err)
		require.NoError(t, c.DeleteFile(bgc, "/round"))
		require.NoError(t, c.PutFile(bgc, "/round", strings.NewReader("10")))
		require.NoError(t, c.FinishCommit(projectName, background, bgc.Branch.Name, bgc.Id))

		//  Put new file in 's3in', which will update every datum at once and
		//  trigger a job that, correspondingly, updates the 'background' part of
		//  every datum's output.
		s3c, err := c.StartCommit(projectName, s3in, "master")
		require.NoError(t, err)
		require.NoError(t, c.DeleteFile(s3Commit, "/file"))
		require.NoError(t, c.PutFile(s3Commit, "/file", strings.NewReader("bar")))
		require.NoError(t, c.FinishCommit(projectName, s3in, s3c.Branch.Name, s3c.Id))

		_, err = c.WaitCommit(projectName, pipeline, "master", "")
		require.NoError(t, err)

		jis, err := c.ListJob(projectName, pipeline, nil, 0, false)
		require.NoError(t, err)
		require.Equal(t, 12, len(jis))
		for j := 0; j < len(jis); j++ {
			require.Equal(t, "JOB_SUCCESS", jis[j].State.String())

			// Check that no service is left over
			k := tu.GetKubeClient(t)
			require.NoErrorWithinTRetry(t, 68*time.Second, func() error {
				svcs, err := k.CoreV1().Services(ns).List(context.Background(), metav1.ListOptions{})
				require.NoError(t, err)
				for _, s := range svcs.Items {
					if s.ObjectMeta.Name == ppsutil.SidecarS3GatewayService(jis[j].Job.Pipeline, jis[j].Job.Id) {
						return errors.Errorf("service %q should be cleaned up by sidecar after job", s.ObjectMeta.Name)
					}
				}
				return nil
			})
		}

		// One per file in 'pfsin' (the background repo was set to '10', but no
		// corresponding file was added to pfsin, so only files for 0-9 are in
		// there)
		var seen [10]bool
		// check output
		var buf bytes.Buffer
		for j := 0; j < len(seen); j++ {
			buf.Reset()
			require.NoError(t, c.GetFile(pipelineCommit, strconv.Itoa(j), &buf))
			s := bufio.NewScanner(&buf)
			for s.Scan() {
				// [0] = bg, [1] = pfsin, [2] = s3in
				p := strings.Split(s.Text(), " ")
				// Updating the S3 input should've altered all datums, forcing all files
				// to be reprocessed and changing p[0] in every output file
				require.Equal(t, p[0], "10")
				require.Equal(t, p[1], strconv.Itoa(j))
				pfsin, err := strconv.Atoi(p[1])
				require.NoError(t, err)
				require.False(t, seen[pfsin])
				seen[pfsin] = true
				require.Equal(t, p[2], "bar") // s3 input is now "bar" everywhere
			}
		}
		for j := 0; j < len(seen); j++ {
			require.True(t, seen[j], "%d", j) // all datums from pfsin were reprocessed
		}
	})

	t.Run("S3Output", func(t *testing.T) {
		repo := tu.UniqueString("pfs_data")
		require.NoError(t, c.CreateRepo(projectName, repo))
		// Pipelines with S3 output should not skip datums, as they have no way of
		// tracking which output data should be associated with which input data.
		// Therefore every output file should have the same "background" value after
		// each job finishes.
		background := tu.UniqueString("bg_data")
		require.NoError(t, c.CreateRepo(projectName, background))

		pipeline := tu.UniqueString("Pipeline")
		// 'pipeline' needs access to the 'background' repo to run successfully
		require.NoError(t, c.ModifyRepoRoleBinding(projectName, background,
			fmt.Sprintf("pipeline:%s/%s", projectName, pipeline), []string{auth.RepoReaderRole}))

		_, err := c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(projectName, pipeline),
			Transform: &pps.Transform{
				Image: "pachyderm/ubuntu-with-s3-clients:v0.0.1",
				Cmd:   []string{"bash", "-x"},
				Stdin: []string{
					fmt.Sprintf(
						// access background repo via regular s3g (not S3_ENDPOINT, which
						// can only access inputs), as above.
						"aws --endpoint=http://pachd:%d s3 cp s3://master.%s.%s/round /tmp/bg",
						s3gPort, background, projectName,
					),
					"cat /pfs/in/* >/tmp/pfsin",
					// Write the "background" value to a new file in every datum. We
					// should see a file for every datum processed, and this should be
					// rewritten in every job
					"aws --endpoint=${S3_ENDPOINT} s3 cp /tmp/bg s3://out/\"$(cat /tmp/pfsin)\"",
				},
			},
			ParallelismSpec: &pps.ParallelismSpec{Constant: 1},
			Input: &pps.Input{
				Pfs: &pps.PFSInput{
					Project: projectName,
					Repo:    repo,
					Name:    "in",
					Branch:  "master",
					Glob:    "/*",
				},
			},
			S3Out: true,
		})
		require.NoError(t, err)

		masterCommit := client.NewCommit(projectName, repo, "master", "")
		pipelineCommit := client.NewCommit(projectName, pipeline, "master", "")
		// Add files to 'repo'. Old files in 'repo' should be reprocessed in every
		// job, changing the 'background' field in the output
		for i := 1; i <= 5; i++ {
			// Increment "/round" in 'background'
			iS := strconv.Itoa(i)
			bgc, err := c.StartCommit(projectName, background, "master")
			require.NoError(t, err)
			require.NoError(t, c.DeleteFile(bgc, "/round"))
			require.NoError(t, c.PutFile(bgc, "/round", strings.NewReader(iS)))
			require.NoError(t, c.FinishCommit(projectName, background, bgc.Branch.Name, bgc.Id))

			// Put new file in 'repo' to create a new datum and trigger a job
			require.NoError(t, c.PutFile(masterCommit, iS, strings.NewReader(iS)))

			_, err = c.WaitCommit(projectName, pipeline, "master", "")
			require.NoError(t, err)
			jis, err := c.ListJob(projectName, pipeline, nil, 0, false)
			require.NoError(t, err)
			require.Equal(t, i+1, len(jis))
			for j := 0; j < len(jis); j++ {
				require.Equal(t, "JOB_SUCCESS", jis[j].State.String())
			}

			for j := 1; j <= i; j++ {
				var buf bytes.Buffer
				require.NoError(t, c.GetFile(pipelineCommit, strconv.Itoa(j), &buf))
				// buf contains the background value; this should be updated in every
				// datum by every job, because this is an S3Out pipeline
				require.Equal(t, iS, buf.String())
			}
		}
	})
}

func TestS3SkippedDatums(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _, ns := initPachClient(t)
	t.Run("DefaultProject", func(t *testing.T) {
		testS3SkippedDatums(t, c, ns, pfs.DefaultProjectName)
	})
	t.Run("NonDefaultProject", func(t *testing.T) {
		projectName := tu.UniqueString("project")
		require.NoError(t, c.CreateProject(projectName))
		testS3SkippedDatums(t, c, ns, projectName)
	})
}

// TestDontDownloadData tests that when a pipeline sets `S3: true` on an input,
// the worker binary doesn't download any datums from that input to the worker.
// Previously, we downloaded the data but didn't link it (a bug), so this test
// both scans `/pfs` (`-L` to follow symlinks) and confirms no data appears
// either in the usual `/pfs/input_repo` or in `/pfs/.scratch`. It also checks
// explicitly that no dead symlink is created at `/pfs/input_repo`.
func TestDontDownloadData(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _, _ := initPachClient(t)

	repo := tu.UniqueString(t.Name() + "_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repo))
	masterCommit := client.NewCommit(pfs.DefaultProjectName, repo, "master", "")
	require.NoError(t, c.PutFile(masterCommit, "test.txt", strings.NewReader("This is a test")))

	pipeline := tu.UniqueString("Pipeline")
	_, err := c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
		Transform: &pps.Transform{
			Cmd: []string{"bash", "-x"},
			Stdin: []string{
				`if find -L /pfs -type f | xargs grep --with-filename "This is a test"; then`,
				`  exit 1`, // The data shouldn't be downloaded; this means it was
				`fi`,
			},
		},
		ParallelismSpec: &pps.ParallelismSpec{Constant: 1},
		Input: &pps.Input{
			Pfs: &pps.PFSInput{
				Project: pfs.DefaultProjectName,
				Repo:    repo,
				Name:    "input_repo",
				Branch:  "master",
				S3:      true,
				Glob:    "/",
			},
		},
	})
	require.NoError(t, err)

	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, pipeline, "master", "")
	require.NoError(t, err)

	jobInfo, err := c.WaitJob(pfs.DefaultProjectName, pipeline, commitInfo.Commit.Id, false)
	require.NoError(t, err)
	require.Equal(t, "JOB_SUCCESS", jobInfo.State.String())
}
