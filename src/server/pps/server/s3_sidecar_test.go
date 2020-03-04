/*
This test is for PPS pipelines that use S3 inputs/outputs. Most of these
pipelines use the pachyderm/s3testing image, which exists on dockerhub but can
be built by running:
  cd etc/testing/images/s3testing
  make push-to-minikube
*/
package server

import (
	"bytes"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsutil"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestS3PipelineErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := tu.GetPachClient(t)
	require.NoError(t, c.DeleteAll())

	repo1, repo2 := tu.UniqueString(t.Name()+"_data"), tu.UniqueString(t.Name()+"_data")
	require.NoError(t, c.CreateRepo(repo1))
	require.NoError(t, c.CreateRepo(repo2))

	pipeline := tu.UniqueString("Pipeline")
	err := c.CreatePipeline(
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
	err = c.CreatePipeline(
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

func TestS3Inputs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := tu.GetPachClient(t)
	require.NoError(t, c.DeleteAll())

	repo := tu.UniqueString(t.Name() + "_data")
	require.NoError(t, c.CreateRepo(repo))

	_, err := c.PutFile(repo, "master", "foo", strings.NewReader("foo"))
	require.NoError(t, err)

	pipeline := tu.UniqueString("Pipeline")
	err = c.CreatePipeline(
		pipeline,
		"pachyderm/ubuntus3clients:v0.0.1",
		[]string{"bash", "-x"},
		[]string{
			"ls -R /pfs >/pfs/out/pfs_files",
			"aws --endpoint=${S3_ENDPOINT} s3 ls >/pfs/out/s3_buckets",
			"aws --endpoint=${S3_ENDPOINT} s3 ls s3://input_repo >/pfs/out/s3_files",
		},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		&pps.Input{
			Pfs: &pps.PFSInput{
				Name:   "input_repo",
				Repo:   repo,
				Branch: "master",
				S3:     true,
			},
		},
		"",
		false,
	)

	jis, err := c.FlushJobAll([]*pfs.Commit{client.NewCommit(repo, "master")}, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(jis))
	jobInfo := jis[0]
	require.Equal(t, "JOB_SUCCESS", jobInfo.State.String())

	// check files in /pfs
	var buf bytes.Buffer
	c.GetFile(pipeline, "master", "pfs_files", 0, 0, &buf)
	require.True(t,
		strings.Contains(buf.String(), "out") && !strings.Contains(buf.String(), "input_repo"),
		"expected \"out\" but not \"input_repo\" in %q", buf.String())

	// check s3 buckets
	buf.Reset()
	c.GetFile(pipeline, "master", "s3_buckets", 0, 0, &buf)
	require.True(t,
		strings.Contains(buf.String(), "input_repo") && !strings.Contains(buf.String(), "out"),
		"expected \"input_repo\" but not \"out\" in %q", buf.String())

	// Check files in input_repo
	buf.Reset()
	c.GetFile(pipeline, "master", "s3_files", 0, 0, &buf)
	require.Matches(t, "foo", buf.String())

	// Check that no service is left over
	k := tu.GetKubeClient(t)
	svcs, err := k.CoreV1().Services(v1.NamespaceDefault).List(metav1.ListOptions{})
	require.NoError(t, err)
	for _, s := range svcs.Items {
		require.NotEqual(t, s.ObjectMeta.Name, ppsutil.SidecarS3GatewayService(jobInfo.Job.ID))
	}
}
