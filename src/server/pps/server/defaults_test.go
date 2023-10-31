package server_test

import (
	"bytes"
	"context"
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
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)

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
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)

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

func TestAPIServer_CreatePipelineV2_projectDefaults(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)

	_, err := env.PPSServer.SetClusterDefaults(ctx, &pps.SetClusterDefaultsRequest{
		ClusterDefaultsJson: `{"create_pipeline_request": {"datum_tries": 17, "autoscaling": true}}`,
	})
	require.NoError(t, err, "SetClusterDefaults must succeed")
	type projectDefaultsSetter interface {
		SetProjectDefaults(ctx context.Context, req *pps.SetProjectDefaultsRequest) (*pps.SetProjectDefaultsResponse, error)
	}

	_, err = env.PPSServer.(projectDefaultsSetter).SetProjectDefaults(ctx, &pps.SetProjectDefaultsRequest{
		Project:             &pfs.Project{Name: "default"},
		ProjectDefaultsJson: `{"create_pipeline_request": {"datum_tries": 13}}`,
	})
	require.NoError(t, err, "SetProjectDefaults must succeed")

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
	require.Equal(t, int64(13), req.DatumTries, "cluster default is effective")
	require.False(t, req.Autoscaling, "spec must override default")
	require.NotNil(t, req.Transform)
	require.Len(t, req.Transform.Cmd, 4)
}

func TestAPIServer_SetClusterDefaults_regenerate(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)

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

func TestAPIServer_SetProjectDefaults_regenerate(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)

	_, err := env.PPSServer.SetClusterDefaults(ctx, &pps.SetClusterDefaultsRequest{
		ClusterDefaultsJson: `{"create_pipeline_request": {"datum_tries": 8, "autoscaling": false}}`,
	})
	require.NoError(t, err, "SetClusterDefaults must succeed")

	_, err = env.PPSServer.SetProjectDefaults(ctx, &pps.SetProjectDefaultsRequest{
		ProjectDefaultsJson: `{"create_pipeline_request": {"datum_tries": 9, "autoscaling": true}}`,
	})
	require.NoError(t, err, "SetProjectDefaults must succeed")

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
	require.Equal(t, int64(9), req.DatumTries, "project default is effective")
	require.False(t, req.Autoscaling, "spec must override default")

	_, err = env.PPSServer.SetClusterDefaults(ctx, &pps.SetClusterDefaultsRequest{
		ClusterDefaultsJson: `{"create_pipeline_request": {"datum_tries": 6, "autoscaling": true}}`,
		Regenerate:          true,
	})
	require.NoError(t, err, "SetClusterDefaults must succeed")
	ir, err := env.PPSServer.InspectPipeline(ctx, &pps.InspectPipelineRequest{Pipeline: &pps.Pipeline{Project: &pfs.Project{Name: pfs.DefaultProjectName}, Name: pipeline}, Details: true})
	require.NoError(t, err, "InspectPipeline must succeed")
	require.NoError(t, protojson.Unmarshal([]byte(ir.EffectiveSpecJson), &req), "unmarshalling effective spec JSON must not error")
	require.Equal(t, int64(9), req.DatumTries, "cluster default is ineffective")
	require.Equal(t, false, req.Autoscaling, "autoscaling is still false")

	_, err = env.PPSServer.SetProjectDefaults(ctx, &pps.SetProjectDefaultsRequest{
		ProjectDefaultsJson: `{"create_pipeline_request": {"datum_tries": 12, "autoscaling": true}}`,
		Regenerate:          true,
	})
	require.NoError(t, err, "SetProjectDefaults must succeed")
	ir, err = env.PPSServer.InspectPipeline(ctx, &pps.InspectPipelineRequest{Pipeline: &pps.Pipeline{Project: &pfs.Project{Name: pfs.DefaultProjectName}, Name: pipeline}, Details: true})
	require.NoError(t, err, "InspectPipeline must succeed")
	require.NoError(t, protojson.Unmarshal([]byte(ir.EffectiveSpecJson), &req), "unmarshalling effective spec JSON must not error")
	require.Equal(t, int64(12), req.DatumTries, "project default is effective")
	require.Equal(t, false, req.Autoscaling, "autoscaling is still false")

	_, err = env.PPSServer.SetProjectDefaults(ctx, &pps.SetProjectDefaultsRequest{
		ProjectDefaultsJson: `{"create_pipeline_request": {"autoscaling": true}}`,
		Regenerate:          true,
	})
	require.NoError(t, err, "SetProjectDefaults must succeed")
	ir, err = env.PPSServer.InspectPipeline(ctx, &pps.InspectPipelineRequest{Pipeline: &pps.Pipeline{Project: &pfs.Project{Name: pfs.DefaultProjectName}, Name: pipeline}, Details: true})
	require.NoError(t, err, "InspectPipeline must succeed")
	require.NoError(t, protojson.Unmarshal([]byte(ir.EffectiveSpecJson), &req), "unmarshalling effective spec JSON must not error")
	require.Equal(t, int64(6), req.DatumTries, "cluster default is effective")
	require.Equal(t, false, req.Autoscaling, "autoscaling is still false")
}

func TestAPIServer_CreatePipelineV2_delete(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)

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
	require.Equal(t, int64(3), req.DatumTries, "built-in default")

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

func TestAPIServer_CreatePipelineV2_zero_value(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)

	_, err := env.PPSServer.SetClusterDefaults(ctx, &pps.SetClusterDefaultsRequest{
		ClusterDefaultsJson: `{"create_pipeline_request": {"datum_tries": 17, "autoscaling": true, "resource_requests": {"memory": "200Mi", "cpu": 0}}}`,
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

func TestAPIServer_CreatePipelineV2_reprocessSpec(t *testing.T) {
	var testCases = map[string]struct {
		defaults string
		wantErr  bool
	}{
		"missing reprocess spec":  {defaults: `{"create_pipeline_request": {}}`},
		"reprocess until success": {defaults: `{"create_pipeline_request": {"reprocessSpec": "until_success"}}`},
		"reprocess every job":     {defaults: `{"create_pipeline_request": {"reprocessSpec": "every_job"}}`},
		"bad reprocess spec":      {defaults: `{"create_pipeline_request": {"reprocessSpec": "#<not a reprocess spec>"}}`, wantErr: true},
	}
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			_, err := env.PPSServer.SetClusterDefaults(ctx, &pps.SetClusterDefaultsRequest{
				ClusterDefaultsJson: testCase.defaults,
			})
			if err != nil {
				if testCase.wantErr {
					return
				}
				t.Errorf("%s errored: %v", testCase.defaults, err)
			}
			if testCase.wantErr {
				t.Errorf("%s should have errored, but didn’t", testCase.defaults)
			}
		})
	}
}

func TestAPIServer_GetProjectDefaults(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	_, err := env.PPSServer.GetProjectDefaults(ctx, &pps.GetProjectDefaultsRequest{})
	require.NoError(t, err, "empty request must not be an error")
	_, err = env.PPSServer.GetProjectDefaults(ctx, &pps.GetProjectDefaultsRequest{Project: &pfs.Project{}})
	require.NoError(t, err, "empty project must not be an error")
	_, err = env.PPSServer.GetProjectDefaults(ctx, &pps.GetProjectDefaultsRequest{Project: &pfs.Project{Name: "#<invalid project>"}})
	require.YesError(t, err, "invalid project must be an error")
	resp, err := env.PPSServer.GetProjectDefaults(ctx, &pps.GetProjectDefaultsRequest{Project: &pfs.Project{Name: "default"}})
	require.NoError(t, err, "default project must have defaults")
	require.Equal(t, resp.ProjectDefaultsJson, "{}", "default project defaults must be empty")
}
