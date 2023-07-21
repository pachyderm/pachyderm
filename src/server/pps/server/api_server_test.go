package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	ppsserver "github.com/pachyderm/pachyderm/v2/src/server/pps/server"
)

func TestListDatum(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
	ctx = env.Context
	repo := "TestListDatum"
	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
	commit1, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
	require.NoError(t, err)
	for i := 0; i < 9; i++ {
		require.NoError(t, env.PachClient.PutFile(commit1, fmt.Sprintf("/file%d", i), &bytes.Buffer{}))
	}
	require.NoError(t, env.PachClient.FinishCommit(pfs.DefaultProjectName, repo, "master", commit1.Id))
	_, err = env.PachClient.WaitCommit(pfs.DefaultProjectName, repo, "master", commit1.Id)
	require.NoError(t, err)

	input := &pps.Input{Pfs: &pps.PFSInput{Repo: repo, Glob: "/*"}}
	request := &pps.ListDatumRequest{Input: input}
	listDatumClient, err := env.PachClient.PpsAPIClient.ListDatum(ctx, request)
	require.NoError(t, err)
	dis, err := grpcutil.Collect[*pps.DatumInfo](listDatumClient, 1000)
	require.NoError(t, err)
	require.Equal(t, 9, len(dis))
	var datumIDs []string
	for _, di := range dis {
		datumIDs = append(datumIDs, di.Datum.Id)
	}
	// Test getting the datums in three pages of 3
	var pagedDatumIDs []string
	request = &pps.ListDatumRequest{Input: input, Number: 3}
	listDatumClient, err = env.PachClient.PpsAPIClient.ListDatum(ctx, request)
	require.NoError(t, err)
	dis, err = grpcutil.Collect[*pps.DatumInfo](listDatumClient, 1000)
	require.NoError(t, err)
	require.Equal(t, 3, len(dis))
	for _, di := range dis {
		pagedDatumIDs = append(pagedDatumIDs, di.Datum.Id)
	}
	// get next two pages
	for i := 0; i < 2; i++ {
		request = &pps.ListDatumRequest{Input: input, Number: 3, PaginationMarker: dis[2].Datum.Id}
		listDatumClient, err = env.PachClient.PpsAPIClient.ListDatum(ctx, request)
		require.NoError(t, err)
		dis, err = grpcutil.Collect[*pps.DatumInfo](listDatumClient, 1000)
		require.NoError(t, err)
		require.Equal(t, 3, len(dis))
		for _, di := range dis {
			pagedDatumIDs = append(pagedDatumIDs, di.Datum.Id)
		}
	}
	// we should have gotten all the datums
	require.ElementsEqual(t, datumIDs, pagedDatumIDs)
	request = &pps.ListDatumRequest{Input: input, Number: 1, PaginationMarker: dis[2].Datum.Id}
	listDatumClient, err = env.PachClient.PpsAPIClient.ListDatum(ctx, request)
	require.NoError(t, err)
	dis, err = grpcutil.Collect[*pps.DatumInfo](listDatumClient, 1000)
	require.NoError(t, err)
	require.Equal(t, 0, len(dis))
	// Test getting the datums in three pages of 3 in reverse order
	var reverseDatumIDs []string
	// get last page
	request = &pps.ListDatumRequest{Input: input, Number: 3, Reverse: true}
	listDatumClient, err = env.PachClient.PpsAPIClient.ListDatum(ctx, request)
	require.NoError(t, err)
	dis, err = grpcutil.Collect[*pps.DatumInfo](listDatumClient, 1000)
	require.NoError(t, err)
	require.Equal(t, 3, len(dis))
	for _, di := range dis {
		reverseDatumIDs = append(reverseDatumIDs, di.Datum.Id)
	}
	// get previous two pages
	for i := 0; i < 2; i++ {
		request = &pps.ListDatumRequest{Input: input, Number: 3, PaginationMarker: dis[2].Datum.Id, Reverse: true}
		listDatumClient, err = env.PachClient.PpsAPIClient.ListDatum(ctx, request)
		require.NoError(t, err)
		dis, err = grpcutil.Collect[*pps.DatumInfo](listDatumClient, 1000)
		require.NoError(t, err)
		require.Equal(t, 3, len(dis))
		for _, di := range dis {
			reverseDatumIDs = append(reverseDatumIDs, di.Datum.Id)
		}
	}
	request = &pps.ListDatumRequest{Input: input, Number: 1, PaginationMarker: dis[2].Datum.Id, Reverse: true}
	listDatumClient, err = env.PachClient.PpsAPIClient.ListDatum(ctx, request)
	require.NoError(t, err)
	dis, err = grpcutil.Collect[*pps.DatumInfo](listDatumClient, 1000)
	require.NoError(t, err)
	require.Equal(t, 0, len(dis))
	for i, di := range datumIDs {
		require.Equal(t, di, reverseDatumIDs[len(reverseDatumIDs)-1-i])
	}
}

func TestRenderTemplate(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
	client := env.PachClient.PpsAPIClient
	res, err := client.RenderTemplate(ctx, &pps.RenderTemplateRequest{
		Args: map[string]string{
			"arg1": "value1",
		},
		Template: `
			function (arg1) {
				pipeline: {name: arg1},
			}
		`,
	})
	require.NoError(t, err)
	require.Len(t, res.Specs, 1)
}

func TestParseLokiLine(t *testing.T) {
	testData := []struct {
		name        string
		line        string
		wantMessage string
		wantErr     bool
	}{
		{
			name:    "empty",
			line:    "",
			wantErr: true,
		},
		{
			name:    "invalid json",
			line:    "{this is not json}",
			wantErr: true,
		},
		{
			name:    "useless json",
			line:    "{}",
			wantErr: true,
		},
		{
			name:        "docker json",
			line:        `{"log":"{\"message\":\"ok\"}"}`,
			wantMessage: "ok",
		},
		{
			name:        "docker json with extra fields",
			line:        `{"log":"{\"message\":\"ok\",\"extraField\":42}"}`,
			wantMessage: "ok",
		},
		{
			name:    "docker with invalid json inside",
			line:    `{"log":"{this is not json}"}`,
			wantErr: true,
		},
		{
			name:        "native json",
			line:        `{"message":"ok"}`,
			wantMessage: "ok",
		},
		{
			name:        "native json with extra fields",
			line:        `{"message":"ok","extraField":42}`,
			wantMessage: "ok",
		},
		{
			name:        "native json with duplicate field",
			line:        `{"message":"ok","message":"ok"}`,
			wantMessage: "ok",
		},
		{
			name:        "CRI with duplicate field",
			line:        `2022-01-01T00:00:00.1234 stdout F {"message":"ok","message":"ok"}`,
			wantMessage: "ok",
		},
		{
			name:        "docker json with duplicate field",
			line:        `{"log":"{\"message\":\"ok\",\"message\":\"ok\"}"}`,
			wantMessage: "ok",
		},
		{
			name:        "mostly empty native json",
			line:        `{"master":false,"user":true}`,
			wantMessage: "",
		},
		{
			name:        "cri format with flags and valid message",
			line:        `2022-01-01T00:00:00.1234 stdout F {"message":"ok"}`,
			wantMessage: "ok",
		},
		{
			name:        "cri format without flags and valid message",
			line:        `2022-01-01T00:00:00.1234 stdout {"message":"ok"}`,
			wantMessage: "ok",
		},
		{
			name:        "cri format with flags and valid message and extra fields",
			line:        `2022-01-01T00:00:00.1234 stdout F {"message":"ok","extraField":42}`,
			wantMessage: "ok",
		},
		{
			name:        "cri format without flags and valid message and extra fields",
			line:        `2022-01-01T00:00:00.1234 stdout {"message":"ok","extraField":42}`,
			wantMessage: "ok",
		},
		{
			name:    "cri format with flags and EOF",
			line:    `2022-01-01T00:00:00.1234 stdout F`,
			wantErr: true,
		},
		{
			name:    "cri format without flags and EOF",
			line:    `2022-01-01T00:00:00.1234 stdout`,
			wantErr: true,
		},
		{
			name:    "cri format with flags and invalid json",
			line:    `2022-01-01T00:00:00.1234 stdout F this is not json`,
			wantErr: true,
		},
		{
			name:    "cri format without flags and invalid json",
			line:    `2022-01-01T00:00:00.1234 stdout this is not json`,
			wantErr: true,
		},
		{
			name:    "cri format with flags and EOF right after {",
			line:    `2022-01-01T00:00:00.1234 stdout F {`,
			wantErr: true,
		},
		{
			name:    "cri format without flags and EOF right after {",
			line:    `2022-01-01T00:00:00.1234 stdout {`,
			wantErr: true,
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			var msg pps.LogMessage
			err := ppsserver.ParseLokiLine(test.line, &msg)
			t.Logf("err: %v", err)
			if test.wantErr && err == nil {
				t.Fatal("parse: got success, want error")
			} else if !test.wantErr && err != nil {
				t.Fatalf("parse: unexpected error: %v", err)
			}
			if got, want := msg.Message, test.wantMessage; got != want {
				t.Fatalf("parse: message:\n  got: %v\n want: %v", got, want)
			}
		})
	}
}

func TestDeletePipelines(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
	inputRepo := tu.UniqueString("repo")
	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, inputRepo))
	// pipeline1 is in default project and takes inputRepo as input
	// pipeline2 is in a non-default project, and takes pipeline1 as input
	project := tu.UniqueString("project-")
	pipeline1, pipeline2 := tu.UniqueString("pipeline1-"), tu.UniqueString("pipeline2-")
	require.NoError(t, env.PachClient.CreateProject(project))
	require.NoError(t, env.PachClient.CreatePipeline(
		pfs.DefaultProjectName,
		pipeline1,
		"", /* default image*/
		[]string{"cp", "-r", "/pfs/in", "/pfs/out"},
		nil, /* stdin */
		nil, /* spec */
		&pps.Input{Pfs: &pps.PFSInput{Project: pfs.DefaultProjectName, Repo: inputRepo, Glob: "/*", Name: "in"}},
		"",   /* output */
		true, /* update */
	))
	require.NoError(t, env.PachClient.CreatePipeline(
		project,
		pipeline2,
		"", /* default image*/
		[]string{"cp", "-r", "/pfs/in", "/pfs/out"},
		nil, /* stdin */
		nil, /* spec */
		&pps.Input{Pfs: &pps.PFSInput{Project: pfs.DefaultProjectName, Repo: pipeline1, Glob: "/*", Name: "in"}},
		"",   /* output */
		true, /* update */
	))
	// update pipeline 1; this helps verify that internally, we delete pipelines topologically
	require.NoError(t, env.PachClient.CreatePipeline(
		pfs.DefaultProjectName,
		pipeline1,
		"", /* default image*/
		[]string{"cp", "-r", "/pfs/in", "/pfs/out"},
		nil, /* stdin */
		nil, /* spec */
		&pps.Input{Pfs: &pps.PFSInput{Project: pfs.DefaultProjectName, Repo: inputRepo, Glob: "/*", Name: "in"}},
		"",   /* output */
		true, /* update */
	))
	inspectResp, err := env.PachClient.InspectPipeline(pfs.DefaultProjectName, pipeline1, false)
	require.NoError(t, err)
	require.Equal(t, uint64(2), inspectResp.Version)
	deleteResp, err := env.PachClient.DeletePipelines(ctx, &pps.DeletePipelinesRequest{All: true})
	require.NoError(t, err)
	require.Equal(t, 2, len(deleteResp.Pipelines))
}

func TestUpdatePipelineInputBranch(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
	repo := "input"
	pipeline := "pipeline"
	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
	require.NoError(t, env.PachClient.CreatePipeline(
		pfs.DefaultProjectName,
		pipeline,
		"", /* default image*/
		[]string{"cp", "-r", "/pfs/in", "/pfs/out"},
		nil, /* stdin */
		nil, /* spec */
		&pps.Input{Pfs: &pps.PFSInput{Project: pfs.DefaultProjectName, Repo: repo, Glob: "/*", Name: "in"}},
		"",   /* output */
		true, /* update */
	))
	commit1, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
	require.NoError(t, err)
	require.NoError(t, env.PachClient.PutFile(commit1, "/foo", strings.NewReader("foo")))
	require.NoError(t, env.PachClient.FinishCommit(pfs.DefaultProjectName, repo, "master", commit1.Id))
	require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, repo, "pin", "master", "", nil))
	require.NoError(t, env.PachClient.CreatePipeline(
		pfs.DefaultProjectName,
		pipeline,
		"", /* default image*/
		[]string{"cp", "-r", "/pfs/in", "/pfs/out"},
		nil, /* stdin */
		nil, /* spec */
		&pps.Input{Pfs: &pps.PFSInput{Project: pfs.DefaultProjectName, Repo: repo, Branch: "pin", Glob: "/*", Name: "in"}},
		"",   /* output */
		true, /* update */
	))
}

func TestGetClusterDefaults(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
	resp, err := env.PPSServer.GetClusterDefaults(ctx, &pps.GetClusterDefaultsRequest{})
	require.NoError(t, err, "GetClusterDefaults failed")
	require.NotEqual(t, "", resp.ClusterDefaultsJson)
	var cd pps.ClusterDefaults
	require.NoError(t, json.Unmarshal([]byte(resp.ClusterDefaultsJson), &cd), "cluster defaults JSON must unmarshal")
}

func TestSetClusterDefaults(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

	t.Run("BadJSON", func(t *testing.T) {
		_, err := env.PPSServer.SetClusterDefaults(ctx, &pps.SetClusterDefaultsRequest{
			ClusterDefaultsJson: `#<this is not JSON>`,
		})
		require.YesError(t, err, "syntactically-invalid JSON is an error")
		s, ok := status.FromError(err)
		require.True(t, ok, "syntactically-invalid JSON returns a status")
		require.Equal(t, codes.InvalidArgument, s.Code(), "syntactically-invalid JSON is an invalid argument")
	})

	t.Run("InvalidDetails", func(t *testing.T) {
		_, err := env.PPSServer.SetClusterDefaults(ctx, &pps.SetClusterDefaultsRequest{
			ClusterDefaultsJson: `{"not an valid spec field":123}`,
		})
		require.YesError(t, err, "invalid details are an error")
		s, ok := status.FromError(err)
		require.True(t, ok, "semantically-invalid JSON returns a status")
		require.Equal(t, codes.InvalidArgument, s.Code(), "semantically-invalid JSON is an invalid argument")
	})

	t.Run("ValidDetails", func(t *testing.T) {
		resp, err := env.PPSServer.SetClusterDefaults(ctx, &pps.SetClusterDefaultsRequest{
			ClusterDefaultsJson: `{"create_pipeline_request": {"autoscaling": true}}`,
		})
		require.NoError(t, err, "SetClusterDefaults failed")
		// FIXME: this will change once CORE-1708 is implemented
		require.Len(t, resp.AffectedPipelines, 0, "pipelines should not yet be affected by setting defaults")
		getResp, err := env.PPSServer.GetClusterDefaults(ctx, &pps.GetClusterDefaultsRequest{})
		require.NoError(t, err, "GetClusterDefaults failed")

		var defaults pps.ClusterDefaults
		err = json.Unmarshal([]byte(getResp.GetClusterDefaultsJson()), &defaults)
		require.NoError(t, err, "unmarshal retrieved cluster defaults")
		require.NotNil(t, defaults.CreatePipelineRequest, "Create Pipeline Request should not be nil after SetClusterDefaults")
		require.True(t, defaults.CreatePipelineRequest.Autoscaling, "default autoscaling should be true after SetClusterDefaults")
	})
}
