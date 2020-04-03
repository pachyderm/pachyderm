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
	"io"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/enterprise"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsutil"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// This test is designed to run against pachyderm in custom namespaces and with
// auth on or off. It reads env vars into here and adjusts tests to make sure
// our s3 gateway feature works in those contexts
var Namespace string
var adminToken string

func init() {
	var ok bool
	Namespace, ok = os.LookupEnv("PACH_NAMESPACE")
	if !ok {
		Namespace = v1.NamespaceDefault
	}
}

// TODO(msteffen) much of the point of factoring GetPachClient into testutil was
// to avoid test-local helpers like this. For expediency, and because I don't
// have a clear vision of how to generalize this yet, I'm subverting that vision
// and creating a helper specific to this test that works with auth both on or
// off. It's a simpler but more brittle implementation of the code in
// s/s/auth/server/testing/auth_test.go. CI will not run this suite with auth
// activated.
//
// At some point, we might want to factor a lot of the auth activation code out
// of src/server/auth/server/testing/auth_test.go into a hypothetical auth
// activation library and then factor this code into testutil (and make
// PACH_TEST_WITH_AUTH a standard env variable across tests). Then it would be
// easier to run all tests with auth either on or off, and we could possibly
// wrap our auth tests in calls to this hypothetical auth activation library but
// have them rely on the hypothetical future implementation of
// tu.GetAuthenticatedPachClient()
func initPachClient(t testing.TB) (*client.APIClient, string) {
	// t.Helper()
	c := tu.GetPachClient(t)
	if adminToken != "" {
		// Use the admin token to clear the cluster and get a non-admin user token,
		// which will be returned. Note that c is shared across tests, but the pach
		// admin token shouldn't change.
		c.SetAuthToken(adminToken)
	}
	require.NoError(t, c.DeleteAll())
	if _, ok := os.LookupEnv("PACH_TEST_WITH_AUTH"); !ok {
		return c, ""
	}
	// Activate Pachyderm Enterprise (if it's not already active)
	_, err := c.Enterprise.Activate(context.Background(),
		&enterprise.ActivateRequest{
			ActivationCode: tu.GetTestEnterpriseCode(),
		})
	require.NoError(t, err)
	activateResp, err := c.AuthAPIClient.Activate(context.Background(),
		&auth.ActivateRequest{Subject: "robot:admin"},
	)
	require.NoError(t, err)
	c.SetAuthToken(activateResp.PachToken)
	adminToken = activateResp.PachToken

	// Create new user auth token
	user := tu.UniqueString("user-")
	resp, err := c.GetAuthToken(c.Ctx(), &auth.GetAuthTokenRequest{
		Subject: user,
	})
	require.NoError(t, err)
	userClient := c.WithCtx(context.Background())
	userClient.SetAuthToken(resp.Token)
	return userClient, resp.Token
}

func TestS3PipelineErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c, _ := initPachClient(t)

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

func TestS3Input(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c, userToken := initPachClient(t)

	repo := tu.UniqueString(t.Name() + "_data")
	require.NoError(t, c.CreateRepo(repo))

	_, err := c.PutFile(repo, "master", "foo", strings.NewReader("foo"))
	require.NoError(t, err)

	pipeline := tu.UniqueString("Pipeline")
	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline: client.NewPipeline(pipeline),
		Transform: &pps.Transform{
			Image: "pachyderm/ubuntu-with-s3-clients:v0.0.1",
			Cmd:   []string{"bash", "-x"},
			Stdin: []string{
				"ls -R /pfs >/pfs/out/pfs_files",
				"aws --endpoint=${S3_ENDPOINT} s3 ls >/pfs/out/s3_buckets",
				"aws --endpoint=${S3_ENDPOINT} s3 ls s3://input_repo >/pfs/out/s3_files",
			},
			Env: map[string]string{
				"AWS_ACCESS_KEY_ID":     userToken,
				"AWS_SECRET_ACCESS_KEY": userToken,
			},
		},
		ParallelismSpec: &pps.ParallelismSpec{Constant: 1},
		Input: &pps.Input{
			Pfs: &pps.PFSInput{
				Name:   "input_repo",
				Repo:   repo,
				Branch: "master",
				S3:     true,
				Glob:   "/",
			},
		},
	})
	require.NoError(t, err)

	jis, err := c.FlushJobAll([]*pfs.Commit{client.NewCommit(repo, "master")}, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(jis))
	jobInfo := jis[0]
	require.Equal(t, "JOB_SUCCESS", jobInfo.State.String())

	// Make sure ListFile works
	files, err := c.ListFile(pipeline, "master", "/")
	require.NoError(t, err)
	require.ElementsEqualUnderFn(t, []string{"/pfs_files", "/s3_buckets", "/s3_files"}, files,
		func(i interface{}) interface{} {
			return i.(*pfs.FileInfo).File.Path
		})

	// check files in /pfs
	var buf bytes.Buffer
	c.GetFile(pipeline, "master", "pfs_files", 0, 0, &buf)
	require.True(t,
		strings.Contains(buf.String(), "out") && !strings.Contains(buf.String(), "input_repo"),
		"expected \"out\" but not \"input_repo\" in %s: %q", "pfs_files", buf.String())

	// check s3 buckets
	buf.Reset()
	c.GetFile(pipeline, "master", "s3_buckets", 0, 0, &buf)
	require.True(t,
		strings.Contains(buf.String(), "input_repo") && !strings.Contains(buf.String(), "out"),
		"expected \"input_repo\" but not \"out\" in %s: %q", "s3_buckets", buf.String())

	// Check files in input_repo
	buf.Reset()
	c.GetFile(pipeline, "master", "s3_files", 0, 0, &buf)
	require.Matches(t, "foo", buf.String())

	// Check that no service is left over
	k := tu.GetKubeClient(t)
	require.NoErrorWithinTRetry(t, 68*time.Second, func() error {
		svcs, err := k.CoreV1().Services(Namespace).List(metav1.ListOptions{})
		require.NoError(t, err)
		for _, s := range svcs.Items {
			if s.ObjectMeta.Name == ppsutil.SidecarS3GatewayService(jobInfo.Job.ID) {
				return errors.Errorf("service %q should be cleaned up by sidecar after job", s.ObjectMeta.Name)
			}
		}
		return nil
	})
}

func TestNamespaceInEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c, _ := initPachClient(t)

	repo := tu.UniqueString(t.Name() + "_data")
	require.NoError(t, c.CreateRepo(repo))

	_, err := c.PutFile(repo, "master", "foo", strings.NewReader("foo"))
	require.NoError(t, err)

	pipeline := tu.UniqueString("Pipeline")
	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline: client.NewPipeline(pipeline),
		Transform: &pps.Transform{
			Cmd: []string{"bash", "-x"},
			Stdin: []string{
				"echo \"${S3_ENDPOINT}\" >/pfs/out/s3_endpoint",
			},
		},
		ParallelismSpec: &pps.ParallelismSpec{Constant: 1},
		Input: &pps.Input{
			Pfs: &pps.PFSInput{
				Name:   "input_repo",
				Repo:   repo,
				Branch: "master",
				S3:     true,
				Glob:   "/",
			},
		},
	})
	require.NoError(t, err)

	jis, err := c.FlushJobAll([]*pfs.Commit{client.NewCommit(repo, "master")}, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(jis))
	jobInfo := jis[0]
	require.Equal(t, "JOB_SUCCESS", jobInfo.State.String())

	// check S3_ENDPOINT variable
	var buf bytes.Buffer
	c.GetFile(pipeline, "master", "s3_endpoint", 0, 0, &buf)
	require.True(t, strings.Contains(buf.String(), ".default"))
}

func TestS3Output(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c, userToken := initPachClient(t)

	repo := tu.UniqueString(t.Name() + "_data")
	require.NoError(t, c.CreateRepo(repo))

	_, err := c.PutFile(repo, "master", "foo", strings.NewReader("foo"))
	require.NoError(t, err)

	pipeline := tu.UniqueString("Pipeline")
	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline: client.NewPipeline(pipeline),
		Transform: &pps.Transform{
			Image: "pachyderm/ubuntu-with-s3-clients:v0.0.1",
			Cmd:   []string{"bash", "-x"},
			Stdin: []string{
				"ls -R /pfs | aws --endpoint=${S3_ENDPOINT} s3 cp - s3://out/pfs_files",
				"aws --endpoint=${S3_ENDPOINT} s3 ls | aws --endpoint=${S3_ENDPOINT} s3 cp - s3://out/s3_buckets",
			},
			Env: map[string]string{
				"AWS_ACCESS_KEY_ID":     userToken,
				"AWS_SECRET_ACCESS_KEY": userToken,
			},
		},
		ParallelismSpec: &pps.ParallelismSpec{Constant: 1},
		Input: &pps.Input{
			Pfs: &pps.PFSInput{
				Name:   "input_repo",
				Repo:   repo,
				Branch: "master",
				Glob:   "/",
			},
		},
		S3Out: true,
	})
	require.NoError(t, err)

	jis, err := c.FlushJobAll([]*pfs.Commit{client.NewCommit(repo, "master")}, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(jis))
	jobInfo := jis[0]
	require.Equal(t, "JOB_SUCCESS", jobInfo.State.String())

	// Make sure ListFile works
	files, err := c.ListFile(pipeline, "master", "/")
	require.NoError(t, err)
	require.ElementsEqualUnderFn(t, []string{"/pfs_files", "/s3_buckets"}, files,
		func(i interface{}) interface{} {
			return i.(*pfs.FileInfo).File.Path
		})

	// check files in /pfs
	var buf bytes.Buffer
	c.GetFile(pipeline, "master", "pfs_files", 0, 0, &buf)
	require.True(t,
		!strings.Contains(buf.String(), "out") && strings.Contains(buf.String(), "input_repo"),
		"expected \"input_repo\" but not \"out\" in %s: %q", "pfs_files", buf.String())

	// check s3 buckets
	buf.Reset()
	c.GetFile(pipeline, "master", "s3_buckets", 0, 0, &buf)
	require.True(t,
		!strings.Contains(buf.String(), "input_repo") && strings.Contains(buf.String(), "out"),
		"expected \"out\" but not \"input_repo\" in %s: %q", "s3_buckets", buf.String())

	// Check that no service is left over
	k := tu.GetKubeClient(t)
	require.NoErrorWithinTRetry(t, 68*time.Second, func() error {
		svcs, err := k.CoreV1().Services(Namespace).List(metav1.ListOptions{})
		require.NoError(t, err)
		for _, s := range svcs.Items {
			if s.ObjectMeta.Name == ppsutil.SidecarS3GatewayService(jobInfo.Job.ID) {
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

	c, userToken := initPachClient(t)

	repo := tu.UniqueString(t.Name() + "_data")
	require.NoError(t, c.CreateRepo(repo))

	_, err := c.PutFile(repo, "master", "foo", strings.NewReader("foo"))
	require.NoError(t, err)

	pipeline := tu.UniqueString("Pipeline")
	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline: client.NewPipeline(pipeline),
		Transform: &pps.Transform{
			Image: "pachyderm/ubuntu-with-s3-clients:v0.0.1",
			Cmd:   []string{"bash", "-x"},
			Stdin: []string{
				"ls -R /pfs | aws --endpoint=${S3_ENDPOINT} s3 cp - s3://out/pfs_files",
				"aws --endpoint=${S3_ENDPOINT} s3 ls | aws --endpoint=${S3_ENDPOINT} s3 cp - s3://out/s3_buckets",
			},
			Env: map[string]string{
				"AWS_ACCESS_KEY_ID":     userToken,
				"AWS_SECRET_ACCESS_KEY": userToken,
			},
		},
		ParallelismSpec: &pps.ParallelismSpec{Constant: 1},
		Input: &pps.Input{
			Pfs: &pps.PFSInput{
				Name:   "input_repo",
				Repo:   repo,
				Branch: "master",
				S3:     true,
				Glob:   "/",
			},
		},
		S3Out: true,
	})
	require.NoError(t, err)

	jis, err := c.FlushJobAll([]*pfs.Commit{client.NewCommit(repo, "master")}, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(jis))
	jobInfo := jis[0]
	require.Equal(t, "JOB_SUCCESS", jobInfo.State.String())

	// Make sure ListFile works
	files, err := c.ListFile(pipeline, "master", "/")
	require.NoError(t, err)
	require.ElementsEqualUnderFn(t, []string{"/pfs_files", "/s3_buckets"}, files,
		func(i interface{}) interface{} {
			return i.(*pfs.FileInfo).File.Path
		})

	// check files in /pfs
	var buf bytes.Buffer
	c.GetFile(pipeline, "master", "pfs_files", 0, 0, &buf)
	require.True(t,
		!strings.Contains(buf.String(), "input_repo") && !strings.Contains(buf.String(), "out"),
		"expected neither \"out\" nor \"input_repo\" in %s: %q", "pfs_files", buf.String())

	// check s3 buckets
	buf.Reset()
	c.GetFile(pipeline, "master", "s3_buckets", 0, 0, &buf)
	require.True(t,
		strings.Contains(buf.String(), "out") && strings.Contains(buf.String(), "input_repo"),
		"expected both \"input_repo\" and \"out\" in %s: %q", "s3_buckets", buf.String())

	// Check that no service is left over
	k := tu.GetKubeClient(t)
	require.NoErrorWithinTRetry(t, 68*time.Second, func() error {
		svcs, err := k.CoreV1().Services(Namespace).List(metav1.ListOptions{})
		require.NoError(t, err)
		for _, s := range svcs.Items {
			if s.ObjectMeta.Name == ppsutil.SidecarS3GatewayService(jobInfo.Job.ID) {
				return errors.Errorf("service %q should be cleaned up by sidecar after job", s.ObjectMeta.Name)
			}
		}
		return nil
	})
}

func TestS3SkippedDatums(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	name := t.Name()

	c, userToken := initPachClient(t)

	t.Run("S3Inputs", func(t *testing.T) {
		s3in := tu.UniqueString(name + "_s3_data")
		require.NoError(t, c.CreateRepo(s3in))
		pfsin := tu.UniqueString(name + "_pfs_data")
		require.NoError(t, c.CreateRepo(pfsin))
		// Pipelines with S3 inputs should still skip datums, as long as the S3 input
		// hasn't changed. We'll check this by reading from a repo that isn't a
		// pipeline input
		background := tu.UniqueString(name + "_bg_data")
		require.NoError(t, c.CreateRepo(background))

		_, err := c.PutFile(s3in, "master", "file", strings.NewReader("foo"))
		require.NoError(t, err)

		pipeline := tu.UniqueString("Pipeline")
		_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pipeline),
			Transform: &pps.Transform{
				Image: "pachyderm/ubuntu-with-s3-clients:v0.0.1",
				Cmd:   []string{"bash", "-x"},
				Stdin: []string{
					fmt.Sprintf(
						// access background repo via regular s3g (not S3_ENDPOINT, which
						// can only access inputs)
						"aws --endpoint=http://pachd.%s:600 s3 cp s3://master.%s/round /tmp/bg",
						Namespace, background,
					),
					"aws --endpoint=${S3_ENDPOINT} s3 cp s3://s3g_in/file /tmp/s3in",
					"cat /pfs/pfs_in/* >/tmp/pfsin",
					"echo \"$(cat /tmp/bg) $(cat /tmp/pfsin) $(cat /tmp/s3in)\" >/pfs/out/out",
				},
				Env: map[string]string{
					"AWS_ACCESS_KEY_ID":     userToken,
					"AWS_SECRET_ACCESS_KEY": userToken,
				},
			},
			ParallelismSpec: &pps.ParallelismSpec{Constant: 1},
			Input: &pps.Input{
				Cross: []*pps.Input{
					{Pfs: &pps.PFSInput{
						Name:   "pfs_in",
						Repo:   pfsin,
						Branch: "master",
						Glob:   "/*",
					}},
					{Pfs: &pps.PFSInput{
						Name:   "s3g_in",
						Repo:   s3in,
						Branch: "master",
						S3:     true,
						Glob:   "/",
					}},
				}},
		})
		require.NoError(t, err)

		jis, err := c.FlushJobAll([]*pfs.Commit{client.NewCommit(s3in, "master")}, nil)
		require.NoError(t, err)
		require.Equal(t, 1, len(jis))

		// Part 1: add files in pfs input w/o changing s3 input. Old files in
		// 'pfsin' should be skipped datums
		// ----------------------------------------------------------------------
		for i := 0; i < 10; i++ {
			// Increment "/round" in 'background'
			iS := fmt.Sprintf("%d", i)
			bgc, err := c.StartCommit(background, "master")
			require.NoError(t, err)
			c.DeleteFile(background, bgc.ID, "/round")
			_, err = c.PutFile(background, bgc.ID, "/round", strings.NewReader(iS))
			require.NoError(t, err)
			c.FinishCommit(background, bgc.ID)

			//  Put new file in 'pfsin' to create a new datum and trigger a job
			_, err = c.PutFile(pfsin, "master", iS, strings.NewReader(iS))
			require.NoError(t, err)

			_, err = c.FlushJobAll([]*pfs.Commit{client.NewCommit(s3in, "master")}, nil)
			require.NoError(t, err)
			jis, err = c.ListJob(pipeline, nil, nil, 0, false)
			require.NoError(t, err)
			require.Equal(t, i+2, len(jis)) // one empty job w/ initial s3in commit
			for j := 0; j < len(jis); j++ {
				require.Equal(t, "JOB_SUCCESS", jis[j].State.String())
			}

			// check output
			var buf bytes.Buffer
			c.GetFile(pipeline, "master", "out", 0, 0, &buf)
			s := bufio.NewScanner(&buf)
			for s.Scan() {
				// [0] = bg, [1] = pfsin, [2] = s3in
				p := strings.Split(s.Text(), " ")
				// Check that bg is the same as the datum; this implies that bg is not
				// being re-read during each job and the datum is being skipped (what we
				// want)
				require.Equal(t, p[0], p[1], "line: "+s.Text())
				require.Equal(t, p[2], "foo", "line: "+s.Text()) // s3 input is being read but not changing
			}
		}

		// Part 2: change s3 input. All old datums should get reprocessed
		// --------------------------------------------------------------
		// Increment "/round" in 'background'
		bgc, err := c.StartCommit(background, "master")
		require.NoError(t, err)
		c.DeleteFile(background, bgc.ID, "/round")
		_, err = c.PutFile(background, bgc.ID, "/round", strings.NewReader("10"))
		require.NoError(t, err)
		c.FinishCommit(background, bgc.ID)

		//  Put new file in 's3in' to create a new datum and trigger a job
		s3c, err := c.StartCommit(s3in, "master")
		require.NoError(t, err)
		c.DeleteFile(s3in, s3c.ID, "/file")
		_, err = c.PutFile(s3in, s3c.ID, "/file", strings.NewReader("bar"))
		require.NoError(t, err)
		c.FinishCommit(s3in, s3c.ID)

		_, err = c.FlushJobAll([]*pfs.Commit{client.NewCommit(s3in, "master")}, nil)
		require.NoError(t, err)
		jis, err = c.ListJob(pipeline, nil, nil, 0, false)
		require.NoError(t, err)
		require.Equal(t, 12, len(jis))
		for j := 0; j < len(jis); j++ {
			require.Equal(t, "JOB_SUCCESS", jis[j].State.String())

			// Check that no service is left over
			k := tu.GetKubeClient(t)
			require.NoErrorWithinTRetry(t, 68*time.Second, func() error {
				svcs, err := k.CoreV1().Services(Namespace).List(metav1.ListOptions{})
				require.NoError(t, err)
				for _, s := range svcs.Items {
					if s.ObjectMeta.Name == ppsutil.SidecarS3GatewayService(jis[j].Job.ID) {
						return errors.Errorf("service %q should be cleaned up by sidecar after job", s.ObjectMeta.Name)
					}
				}
				return nil
			})
		}

		// check output
		var buf bytes.Buffer
		// Flush commit so that GetFile doesn't accidentally run after the job
		// finishes but before the commit finishes
		ciIter, err := c.FlushCommit(
			[]*pfs.Commit{client.NewCommit(s3in, "master")},
			[]*pfs.Repo{client.NewRepo(pipeline)})
		require.NoError(t, err)
		for err != io.EOF {
			_, err = ciIter.Next()
		}
		c.GetFile(pipeline, "master", "out", 0, 0, &buf)
		s := bufio.NewScanner(&buf)
		var seen [10]bool // One per file in 'pfsin'
		for s.Scan() {
			// [0] = bg, [1] = pfsin, [2] = s3in
			p := strings.Split(s.Text(), " ")
			// Updating the S3 input should've altered all datums, forcing all files
			// to be reprocessed and changing p[0] in every line of "out"
			require.Equal(t, p[0], "10")
			pfsinI, err := strconv.Atoi(p[1])
			require.NoError(t, err)
			require.False(t, seen[pfsinI])
			seen[pfsinI] = true
			require.Equal(t, p[2], "bar") // s3 input is now "bar" everywhere
		}
		for j := 0; j < len(seen); j++ {
			require.True(t, seen[j], j) // all datums from pfsin were reprocessed
		}
	})

	t.Run("S3Output", func(t *testing.T) {
		repo := tu.UniqueString(name + "_pfs_data")
		require.NoError(t, c.CreateRepo(repo))
		// Pipelines with S3 output should not skip datums, as they have no way of
		// tracking which output data should be associated with which input data.
		// We'll check this by reading from a repo that isn't a pipeline input
		background := tu.UniqueString(name + "_bg_data")
		require.NoError(t, c.CreateRepo(background))

		pipeline := tu.UniqueString("Pipeline")
		_, err := c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pipeline),
			Transform: &pps.Transform{
				Image: "pachyderm/ubuntu-with-s3-clients:v0.0.1",
				Cmd:   []string{"bash", "-x"},
				Stdin: []string{
					fmt.Sprintf(
						// access background repo via regular s3g (not S3_ENDPOINT, which
						// can only access inputs)
						"aws --endpoint=http://pachd.%s:600 s3 cp s3://master.%s/round /tmp/bg",
						Namespace, background,
					),
					"cat /pfs/in/* >/tmp/pfsin",
					// Write the "background" value to a new file in every datum. As
					// 'aws s3 cp' is destructive, datums will overwrite each others
					// values unless each datum writes to a unique key. This way, we
					// should see a file for every datum processed written in every job
					"aws --endpoint=${S3_ENDPOINT} s3 cp /tmp/bg s3://out/\"$(cat /tmp/pfsin)\"",
					// Also write a file that is itself named for the background value.
					// Because S3-out jobs don't merge their outputs with their parent's
					// outputs, we should only see one such file in every job's output.
					"aws --endpoint=${S3_ENDPOINT} s3 cp /tmp/bg s3://out/bg/\"$(cat /tmp/bg)\"",
				},
				Env: map[string]string{
					"AWS_ACCESS_KEY_ID":     userToken,
					"AWS_SECRET_ACCESS_KEY": userToken,
				},
			},
			ParallelismSpec: &pps.ParallelismSpec{Constant: 1},
			Input: &pps.Input{
				Pfs: &pps.PFSInput{
					Name:   "in",
					Repo:   repo,
					Branch: "master",
					Glob:   "/*",
				},
			},
			S3Out: true,
		})
		require.NoError(t, err)

		// Add files to 'repo'. Old files in 'repo' should be reprocessed in every
		// job, changing the 'background' field in the output
		for i := 0; i < 10; i++ {
			// Increment "/round" in 'background'
			iS := strconv.Itoa(i)
			bgc, err := c.StartCommit(background, "master")
			require.NoError(t, err)
			c.DeleteFile(background, bgc.ID, "/round")
			_, err = c.PutFile(background, bgc.ID, "/round", strings.NewReader(iS))
			require.NoError(t, err)
			c.FinishCommit(background, bgc.ID)

			//  Put new file in 'repo' to create a new datum and trigger a job
			_, err = c.PutFile(repo, "master", iS, strings.NewReader(iS))
			require.NoError(t, err)

			_, err = c.FlushJobAll([]*pfs.Commit{client.NewCommit(repo, "master")}, nil)
			require.NoError(t, err)
			jis, err := c.ListJob(pipeline, nil, nil, 0, false)
			require.NoError(t, err)
			require.Equal(t, i+1, len(jis))
			for j := 0; j < len(jis); j++ {
				require.Equal(t, "JOB_SUCCESS", jis[j].State.String())
			}

			// check output
			// ------------
			// Flush commit so that GetFile doesn't accidentally run after the job
			// finishes but before the commit finishes
			ciIter, err := c.FlushCommit(
				[]*pfs.Commit{client.NewCommit(repo, "master")},
				[]*pfs.Repo{client.NewRepo(pipeline)})
			require.NoError(t, err)
			for err != io.EOF {
				_, err = ciIter.Next()
			}
			for j := 0; j <= i; j++ {
				var buf bytes.Buffer
				require.NoError(t, c.GetFile(pipeline, "master", strconv.Itoa(j), 0, 0, &buf))
				// buf contains the background value; this should be updated in every
				// datum by every job, because this is an S3Out pipeline
				require.Equal(t, iS, buf.String())
			}

			// Make sure that there's exactly one file in 'bg/'--this confirms that
			// the output commit isn't merged with its parent, it only contains the
			// result of processing all datums
			// -------------------------------
			// check no extra files
			fis, err := c.ListFile(pipeline, "master", "/bg")
			require.NoError(t, err)
			require.Equal(t, 1, len(fis))
			// check that expected file exists
			fis, err = c.ListFile(pipeline, "master", fmt.Sprintf("/bg/%d", i))
			require.NoError(t, err)
			require.Equal(t, 1, len(fis))

		}
	})
}
