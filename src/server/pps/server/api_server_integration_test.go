//go:build k8s

package server_test

import (
	"bytes"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"google.golang.org/protobuf/encoding/protojson"
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
