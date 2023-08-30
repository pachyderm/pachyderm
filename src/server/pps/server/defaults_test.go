package server_test

import (
	"bytes"
	"html/template"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"
)

func TestAPIServer_CreatePipelineV2_noDefaults(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

	dr, err := env.PPSServer.GetClusterDefaults(ctx, &pps.GetClusterDefaultsRequest{})
	require.NoError(t, err, "GetClusterDefaults must succeed")
	require.NotEqual(t, "", dr.ClusterDefaultsJson, "baseline cluster defaults must not be missing")
	var defaults pps.ClusterDefaults
	err = protojson.Unmarshal([]byte(dr.ClusterDefaultsJson), &defaults)
	require.NoError(t, err, "baseline cluster defaults must unmarshal")

	repo := "input"
	pipeline := "pipeline"
	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
	var pipelineTemplate = `{
		"pipeline": {
			"project": {
				"name": "{{.ProjectName | js}}"
			},
			"name": "{{.PipelineName | js}}"
		},
		"transform": {
			"cmd": ["cp", "r", "/pfs/in", "/pfs/out"]
		},
		"input": {
			"pfs": {
				"project": "default",
				"repo": "{{.RepoName | js}}",
				"glob": "/*",
				"name": "in"
			}
		},
                "resource_requests": {
	                "cpu": null,
	                "disk": "187Mi"
                },
		"autoscaling": false
	}`
	tmpl, err := template.New("pipeline").Parse(pipelineTemplate)
	require.NoError(t, err, "template must parse")
	var buf bytes.Buffer
	require.NoError(t, tmpl.Execute(&buf, struct {
		ProjectName, PipelineName, RepoName string
	}{pfs.DefaultProjectName, pipeline, repo}), "template must execute")
	resp, err := env.PachClient.PpsAPIClient.CreatePipelineV2(ctx, &pps.CreatePipelineV2Request{
		CreatePipelineRequestJson: buf.String(),
	})
	require.NoError(t, err, "CreatePipelineV2 must succeed")
	require.False(t, resp.EffectiveCreatePipelineRequestJson == "", "response includes effective JSON")
	var req pps.CreatePipelineRequest
	err = protojson.Unmarshal([]byte(resp.EffectiveCreatePipelineRequestJson), &req)
	require.NoError(t, err, "unmarshalling effective JSON must not error")
	require.False(t, req.Autoscaling, "spec must override default")
	require.Equal(t, float32(0), req.ResourceRequests.Cpu, "CPU request must be deleted")
	require.Equal(t, "187Mi", req.ResourceRequests.Disk, "disk request must be overridden")
	require.Equal(t, defaults.CreatePipelineRequest.ResourceRequests.Memory, req.ResourceRequests.Memory, "memory request must be default")
	require.NotNil(t, req.SidecarResourceRequests, "unspecified object must be default")
}

func TestAPIServer_CreatePipelineV2_defaults(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

	_, err := env.PPSServer.SetClusterDefaults(ctx, &pps.SetClusterDefaultsRequest{
		ClusterDefaultsJson: `{"create_pipeline_request": {"datum_tries": 17, "autoscaling": true}}`,
	})
	require.NoError(t, err, "SetClusterDefaults must succeed")

	repo := "input"
	pipeline := "pipeline"
	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
	var pipelineTemplate = `{
		"pipeline": {
			"project": {
				"name": "{{.ProjectName | js}}"
			},
			"name": "{{.PipelineName | js}}"
		},
		"transform": {
			"cmd": ["cp", "r", "/pfs/in", "/pfs/out"]
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
	}`
	tmpl, err := template.New("pipeline").Parse(pipelineTemplate)
	require.NoError(t, err, "template must parse")
	var buf bytes.Buffer
	require.NoError(t, tmpl.Execute(&buf, struct {
		ProjectName, PipelineName, RepoName string
	}{pfs.DefaultProjectName, pipeline, repo}), "template must execute")
	resp, err := env.PachClient.PpsAPIClient.CreatePipelineV2(ctx, &pps.CreatePipelineV2Request{
		CreatePipelineRequestJson: buf.String(),
	})
	require.NoError(t, err, "CreatePipelineV2 must succeed")
	require.False(t, resp.EffectiveCreatePipelineRequestJson == "", "response includes effective JSON")
	var req pps.CreatePipelineRequest
	require.NoError(t, protojson.Unmarshal([]byte(resp.EffectiveCreatePipelineRequestJson), &req), "unmarshalling effective JSON must not error")
	require.Equal(t, int64(17), req.DatumTries, "cluster default is effective")
	require.False(t, req.Autoscaling, "spec must override default")
	require.NotNil(t, req.Transform)
	require.Len(t, req.Transform.Cmd, 4)
}

func TestAPIServer_CreatePipelineV2_regenerate(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

	_, err := env.PPSServer.SetClusterDefaults(ctx, &pps.SetClusterDefaultsRequest{
		ClusterDefaultsJson: `{"create_pipeline_request": {"datum_tries": 17, "autoscaling": true}}`,
	})
	require.NoError(t, err, "SetClusterDefaults must succeed")

	repo := "input"
	pipeline := "pipeline"
	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
	var pipelineTemplate = `{
		"pipeline": {
			"project": {
				"name": "{{.ProjectName | js}}"
			},
			"name": "{{.PipelineName | js}}"
		},
		"transform": {
			"cmd": ["cp", "r", "/pfs/in", "/pfs/out"]
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
	}`
	tmpl, err := template.New("pipeline").Parse(pipelineTemplate)
	require.NoError(t, err, "template must parse")
	var buf bytes.Buffer
	require.NoError(t, tmpl.Execute(&buf, struct {
		ProjectName, PipelineName, RepoName string
	}{pfs.DefaultProjectName, pipeline, repo}), "template must execute")
	resp, err := env.PachClient.PpsAPIClient.CreatePipelineV2(ctx, &pps.CreatePipelineV2Request{
		CreatePipelineRequestJson: buf.String(),
	})
	require.NoError(t, err, "CreatePipelineV2 must succeed")
	require.False(t, resp.EffectiveCreatePipelineRequestJson == "", "response includes effective JSON")
	var req pps.CreatePipelineRequest
	require.NoError(t, protojson.Unmarshal([]byte(resp.EffectiveCreatePipelineRequestJson), &req), "unmarshalling effective JSON must not error")
	require.Equal(t, int64(17), req.DatumTries, "cluster default is effective")
	require.False(t, req.Autoscaling, "spec must override default")

	_, err = env.PPSServer.SetClusterDefaults(ctx, &pps.SetClusterDefaultsRequest{
		ClusterDefaultsJson: `{"create_pipeline_request": {"datum_tries": 4, "autoscaling": true}}`,
		Regenerate:          true,
	})
	require.NoError(t, err, "SetClusterDefaults failed")
	ir, err := env.PPSServer.InspectPipeline(ctx, &pps.InspectPipelineRequest{Pipeline: &pps.Pipeline{Project: &pfs.Project{Name: pfs.DefaultProjectName}, Name: pipeline}, Details: true})
	require.NoError(t, err, "InspectPipeline must succeed")
	require.NoError(t, protojson.Unmarshal([]byte(ir.EffectiveSpecJson), &req), "unmarshalling effective spec JSON must not error")
	require.Equal(t, int64(4), req.DatumTries, "cluster default is effective")
	require.Equal(t, false, req.Autoscaling, "autoscaling is still false")
}

func TestAPIServer_CreatePipelineV2_delete(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

	_, err := env.PPSServer.SetClusterDefaults(ctx, &pps.SetClusterDefaultsRequest{
		ClusterDefaultsJson: "{}",
	})
	require.NoError(t, err, "deleting SetClusterDefaults must succeed")

	repo := "input"
	pipeline := "pipeline"
	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
	var pipelineTemplate = `{
		"pipeline": {
			"project": {
				"name": "{{.ProjectName | js}}"
			},
			"name": "{{.PipelineName | js}}"
		},
		"transform": {
			"cmd": ["cp", "r", "/pfs/in", "/pfs/out"]
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
	}`
	tmpl, err := template.New("pipeline").Parse(pipelineTemplate)
	require.NoError(t, err, "template must parse")
	var buf bytes.Buffer
	require.NoError(t, tmpl.Execute(&buf, struct {
		ProjectName, PipelineName, RepoName string
	}{pfs.DefaultProjectName, pipeline, repo}), "template must execute")
	resp, err := env.PachClient.PpsAPIClient.CreatePipelineV2(ctx, &pps.CreatePipelineV2Request{
		CreatePipelineRequestJson: buf.String(),
	})
	require.NoError(t, err, "CreatePipelineV2 must succeed")
	require.False(t, resp.EffectiveCreatePipelineRequestJson == "", "response includes effective JSON")
	var req pps.CreatePipelineRequest
	require.NoError(t, protojson.Unmarshal([]byte(resp.EffectiveCreatePipelineRequestJson), &req), "unmarshalling effective JSON must not error")
	require.Equal(t, int64(0), req.DatumTries, "no default")

	_, err = env.PPSServer.SetClusterDefaults(ctx, &pps.SetClusterDefaultsRequest{
		ClusterDefaultsJson: `{"create_pipeline_request": {"datum_tries": 4, "autoscaling": true}}`,
		Regenerate:          true,
	})
	require.NoError(t, err, "SetClusterDefaults failed")
	ir, err := env.PPSServer.InspectPipeline(ctx, &pps.InspectPipelineRequest{Pipeline: &pps.Pipeline{Project: &pfs.Project{Name: pfs.DefaultProjectName}, Name: pipeline}, Details: true})
	require.NoError(t, err, "InspectPipeline must succeed")
	require.NoError(t, protojson.Unmarshal([]byte(ir.EffectiveSpecJson), &req), "unmarshalling effective spec JSON must not error")
	require.Equal(t, int64(4), req.DatumTries, "cluster default is effective")
	require.Equal(t, false, req.Autoscaling, "autoscaling is still false")
}
