//go:build k8s

package server_test

import (
	"bytes"
	"strings"
	"testing"
	"text/template"

	"google.golang.org/protobuf/encoding/protojson"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

func TestRerunPipeline(t *testing.T) {
	ctx := pctx.TestContext(t)
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	repo := tu.UniqueString("input")
	pipeline := tu.UniqueString("pipeline")

	// commit data to the repo so there will be datums to process
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repo))
	commit1, err := c.StartCommit(pfs.DefaultProjectName, repo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit1, "/file.txt", &bytes.Buffer{}))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, repo, "master", commit1.Id))
	_, err = c.WaitCommit(pfs.DefaultProjectName, repo, "master", commit1.Id)
	require.NoError(t, err)

	// create the pipelne
	spec := &pps.CreatePipelineRequest{
		Pipeline: &pps.Pipeline{
			Project: &pfs.Project{
				Name: pfs.DefaultProjectName,
			},
			Name: pipeline,
		},
		Transform: &pps.Transform{Cmd: []string{"bash"}, Stdin: []string{"date > /pfs/out/date.txt"}},
		Input: &pps.Input{
			Pfs: &pps.PFSInput{
				Repo: repo,
				Glob: "/",
			},
		},
	}
	js, err := protojson.Marshal(spec)
	require.NoError(t, err, "marshalling JSON must not error")
	_, err = c.PpsAPIClient.CreatePipelineV2(ctx, &pps.CreatePipelineV2Request{
		CreatePipelineRequestJson: string(js),
	})
	require.NoError(t, err, "CreatePipelineV2 must succeed")

	resp, err := c.PpsAPIClient.InspectPipeline(ctx, &pps.InspectPipelineRequest{
		Pipeline: &pps.Pipeline{
			Project: &pfs.Project{
				Name: pfs.DefaultProjectName,
			},
			Name: pipeline,
		}})
	require.NoError(t, err, "InspectPipeline must succeed")
	createdEffectiveSpecJSON := resp.EffectiveSpecJson

	// wait for job to finish and get output commit data
	jobs, err := c.ListJob(pfs.DefaultProjectName, pipeline, nil, 0, true)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobs))
	_, err = c.WaitJob(pfs.DefaultProjectName, pipeline, jobs[0].Job.GetId(), false)
	require.NoError(t, err)
	var date1 bytes.Buffer
	require.NoError(t, c.GetFile(client.NewCommit(pfs.DefaultProjectName, pipeline, "master", ""), "/date.txt", &date1))
	require.NotEqual(t, date1.String(), "")

	// rerun pipeline
	_, err = c.PpsAPIClient.RerunPipeline(ctx, &pps.RerunPipelineRequest{
		Pipeline: &pps.Pipeline{
			Project: &pfs.Project{Name: pfs.DefaultProjectName},
			Name:    pipeline,
		},
	})
	require.NoError(t, err, "RerunPipeline must succeed")
	r, err := c.PpsAPIClient.InspectPipeline(ctx, &pps.InspectPipelineRequest{Pipeline: &pps.Pipeline{Name: pipeline}})
	require.NoError(t, err, "InspectPipeline must succeed")
	require.Equal(t, createdEffectiveSpecJSON, r.EffectiveSpecJson)
	require.Equal(t, r.Version, uint64(2), "pipeline version should = 2")
	// wait for job to finish and get output commit data
	jobs, err = c.ListJob(pfs.DefaultProjectName, pipeline, nil, 0, true)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobs))
	_, err = c.WaitJob(pfs.DefaultProjectName, pipeline, jobs[0].Job.GetId(), false)
	require.NoError(t, err)
	var date2 bytes.Buffer
	require.NoError(t, c.GetFile(client.NewCommit(pfs.DefaultProjectName, pipeline, "master", ""), "/date.txt", &date2))
	require.NotEqual(t, date2.String(), "")
	require.Equal(t, date1.String(), date2.String())
	jobInfo, err := c.InspectJob(pfs.DefaultProjectName, pipeline, jobs[0].Job.GetId(), false)
	require.NoError(t, err)
	require.Equal(t, pps.JobState_JOB_SUCCESS, jobInfo.State)
	require.Equal(t, int64(0), jobInfo.DataProcessed)
	require.Equal(t, int64(1), jobInfo.DataSkipped)
	require.Equal(t, int64(0), jobInfo.DataRecovered)
	require.Equal(t, int64(0), jobInfo.DataFailed)

	// rerun pipeline with reprocess
	_, err = c.PpsAPIClient.RerunPipeline(ctx, &pps.RerunPipelineRequest{
		Pipeline: &pps.Pipeline{
			Project: &pfs.Project{Name: pfs.DefaultProjectName},
			Name:    pipeline,
		},
		Reprocess: true,
	})
	require.NoError(t, err, "RerunPipeline must succeed")
	r, err = c.PpsAPIClient.InspectPipeline(ctx, &pps.InspectPipelineRequest{Pipeline: &pps.Pipeline{Name: pipeline}})
	require.NoError(t, err, "InspectPipeline must succeed")
	require.Equal(t, r.Version, uint64(3), "pipeline version should = 3")
	// wait for job to finish and get output commit data
	jobs, err = c.ListJob(pfs.DefaultProjectName, pipeline, nil, 0, true)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobs))
	_, err = c.WaitJob(pfs.DefaultProjectName, pipeline, jobs[0].Job.GetId(), false)
	require.NoError(t, err)
	var date3 bytes.Buffer
	require.NoError(t, c.GetFile(client.NewCommit(pfs.DefaultProjectName, pipeline, "master", ""), "/date.txt", &date3))
	require.NotEqual(t, date3.String(), "")
	require.NotEqual(t, date1.String(), date3.String())
	require.NotEqual(t, date2.String(), date3.String())
	jobInfo, err = c.InspectJob(pfs.DefaultProjectName, pipeline, jobs[0].Job.GetId(), false)
	require.NoError(t, err)
	require.Equal(t, pps.JobState_JOB_SUCCESS, jobInfo.State)
	require.Equal(t, int64(1), jobInfo.DataProcessed)
	require.Equal(t, int64(0), jobInfo.DataSkipped)
	require.Equal(t, int64(0), jobInfo.DataRecovered)
	require.Equal(t, int64(0), jobInfo.DataFailed)
}

func applyTemplate(text string, data any) (string, error) {
	tmpl, err := template.New("main").Parse(text)
	if err != nil {
		return "", errors.Wrapf(err, "could not parse template %q", text)
	}
	var b bytes.Buffer
	if err := tmpl.Execute(&b, data); err != nil {
		return "", errors.Wrap(err, "could not execute template")
	}
	return b.String(), nil
}

// TestSidecarMetrics tests that sidecar metrics are available.
func TestSidecarMetrics(t *testing.T) {
	ctx := pctx.TestContext(t)
	t.Parallel()
	c, namespace := minikubetestenv.AcquireCluster(t)

	inputRepo := "input"
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, inputRepo))

	js, err := applyTemplate(`{
		"pipeline": {
			"project": {
				"name": "{{.ProjectName | js}}"
			},
			"name": "{{.PipelineName | js}}"
		},
		"transform": {
			"cmd": ["cp", "-r", "/pfs/in/foo", "/pfs/out"]
		},
		"input": {
			"pfs": {
				"project": "default",
				"repo": "{{.RepoName | js}}",
				"glob": "/*",
				"name": "in"
			}
		},
		"autoscaling": false
	}`,
		struct {
			ProjectName, PipelineName, RepoName string
		}{pfs.DefaultProjectName, "pipeline1", inputRepo})
	require.NoError(t, err, "must execute template for pipeline1")
	_, err = c.PpsAPIClient.CreatePipelineV2(ctx, &pps.CreatePipelineV2Request{
		CreatePipelineRequestJson: js,
	})
	require.NoError(t, err, "creating pipeline1 must succeed")

	// pipeline2 pulls metrics from pipeline1
	js, err = applyTemplate(`{
		"pipeline": {
			"project": {
				"name": "{{.ProjectName | js}}"
			},
			"name": "{{.PipelineName | js}}"
		},
		"transform": {
			"image": "curlimages/curl:8.5.0",
			"cmd": ["sh", "-c", "curl http://default-pipeline1-v1:9091/metrics > /pfs/out/output"]
		},
		"input": {
			"pfs": {
				"project": "default",
				"repo": "{{.RepoName | js}}",
				"glob": "/*",
				"name": "in"
			}
		},
		"autoscaling": false
	}`,
		struct {
			ProjectName, PipelineName, RepoName string
		}{pfs.DefaultProjectName, "pipeline2", "pipeline1"})
	require.NoError(t, err, "must execute template for pipeline2")
	_, err = c.PpsAPIClient.CreatePipelineV2(ctx, &pps.CreatePipelineV2Request{
		CreatePipelineRequestJson: js,
	})
	require.NoError(t, err, "creating pipeline2 must succeed")

	// put a file into the input repo, which should trigger pipeline1, whose completion should trigger pipeline2
	commit1, err := c.StartCommit(pfs.DefaultProjectName, inputRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit1, "/foo", strings.NewReader("foo")))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, inputRepo, "master", commit1.Id))

	_, err = c.WaitCommitSetAll(commit1.Id)
	require.NoError(t, err, "commit set must finish")

	branch, err := c.InspectBranch(pfs.DefaultProjectName, "pipeline2", "master")
	require.NoError(t, err, "pipeline2@master must be inspectable")

	var buf bytes.Buffer
	err = c.GetFile(branch.Head, "/output", &buf)
	require.NoError(t, err, "get file must succeed")
	// this is a bit brute force, but it does work
	require.True(t, strings.Contains(buf.String(), "\npachyderm_pachd_pps"))

	kc := testutil.GetKubeClient(t)
	pl, err := kc.CoreV1().Pods(namespace).List(ctx, v1.ListOptions{
		LabelSelector: "pipelineName = pipeline1",
	})
	require.NoError(t, err, "k8s pod list must succeed")
	require.Equal(t, 1, len(pl.Items), "there must be exactly one pipeline")
	var foundSidecar bool
	for _, c := range pl.Items[0].Spec.Containers {
		if c.Name != client.PPSWorkerSidecarContainerName {
			continue
		}
		foundSidecar = true
		var foundMetrics bool
		for _, p := range c.Ports {
			if p.Name != "metrics-storage" {
				continue
			}
			foundMetrics = true
		}
		require.True(t, foundMetrics, "metrics port must exist")
	}
	require.True(t, foundSidecar, "sidecar must exist")
}
