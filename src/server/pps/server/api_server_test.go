package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"strconv"
	"strings"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	ppsserver "github.com/pachyderm/pachyderm/v2/src/server/pps/server"

	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

func TestListDatum(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
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

func TestCreateDatum(t *testing.T) {
	ctx := pctx.TestContext(t)
	pc := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption).PachClient
	repo := tu.UniqueString("TestCreateDatum")
	require.NoError(t, pc.CreateRepo(pfs.DefaultProjectName, repo))
	require.NoError(t, pc.CreateBranch(pfs.DefaultProjectName, repo, "master", "", "", nil))
	input := client.NewPFSInput(pfs.DefaultProjectName, repo, "/*")

	t.Run("EmptyRepo", func(t *testing.T) {
		datumClient, err := pc.PpsAPIClient.CreateDatum(ctx)
		require.NoError(t, err)
		require.NoError(t, datumClient.Send(&pps.CreateDatumRequest{Body: &pps.CreateDatumRequest_Start{Start: &pps.StartCreateDatumRequest{Input: input}}}))
		_, err = datumClient.Recv()
		require.ErrorIs(t, err, io.EOF)
	})

	commit, err := pc.StartCommit(pfs.DefaultProjectName, repo, "master")
	require.NoError(t, err)
	numFiles := ppsserver.DefaultDatumBatchSize + 50
	for i := 0; i < numFiles; i++ {
		require.NoError(t, pc.PutFile(commit, fmt.Sprintf("file%d", i), strings.NewReader(fmt.Sprintf("file%d", i))))
	}
	require.NoError(t, pc.FinishCommit(pfs.DefaultProjectName, repo, "master", commit.Id))

	t.Run("SingleBatch", func(t *testing.T) {
		datumClient, err := pc.PpsAPIClient.CreateDatum(ctx)
		require.NoError(t, err)
		// Requesting more datums than exist should return all datums without erroring
		require.NoError(t, datumClient.Send(&pps.CreateDatumRequest{Body: &pps.CreateDatumRequest_Start{Start: &pps.StartCreateDatumRequest{Input: input, Number: int32(numFiles)}}}))
		n, err := grpcutil.Read[*pps.DatumInfo](datumClient, make([]*pps.DatumInfo, numFiles+1))
		require.True(t, stream.IsEOS(err))
		require.Equal(t, numFiles, n)
	})
	t.Run("MultipleBatches", func(t *testing.T) {
		datumClient, err := pc.PpsAPIClient.CreateDatum(ctx)
		require.NoError(t, err)
		// Not specifying number of datums should return DefaultDatumBatchSize datums
		require.NoError(t, datumClient.Send(&pps.CreateDatumRequest{Body: &pps.CreateDatumRequest_Start{Start: &pps.StartCreateDatumRequest{Input: input}}}))
		dis := make([]*pps.DatumInfo, ppsserver.DefaultDatumBatchSize)
		n, err := grpcutil.Read[*pps.DatumInfo](datumClient, dis)
		require.NoError(t, err)
		require.Equal(t, ppsserver.DefaultDatumBatchSize, n)
		receivedDatums := make(map[string]bool)
		for i := range n {
			require.Equal(t, 1, len(dis[i].Data))
			receivedDatums[dis[i].Data[0].File.Path] = true
		}

		remaining := numFiles - ppsserver.DefaultDatumBatchSize
		require.NoError(t, datumClient.Send(&pps.CreateDatumRequest{Body: &pps.CreateDatumRequest_Continue{Continue: &pps.ContinueCreateDatumRequest{Number: int32(remaining)}}}))
		dis = make([]*pps.DatumInfo, remaining+1)
		n, err = grpcutil.Read[*pps.DatumInfo](datumClient, dis)
		require.True(t, stream.IsEOS(err))
		require.Equal(t, remaining, n)
		for i := range n {
			require.Equal(t, 1, len(dis[i].Data))
			receivedDatums[dis[i].Data[0].File.Path] = true
		}

		for i := range numFiles {
			_, ok := receivedDatums[fmt.Sprintf("/file%d", i)]
			require.True(t, ok, fmt.Sprintf("datum with path /file%d not received", i))
		}
	})
	t.Run("UnionInput", func(t *testing.T) {
		datumClient, err := pc.PpsAPIClient.CreateDatum(ctx)
		require.NoError(t, err)
		input := client.NewUnionInput(
			client.NewPFSInput(pfs.DefaultProjectName, repo, "/*"),
			client.NewPFSInput(pfs.DefaultProjectName, repo, "/*"),
		)
		require.NoError(t, datumClient.Send(&pps.CreateDatumRequest{Body: &pps.CreateDatumRequest_Start{Start: &pps.StartCreateDatumRequest{Input: input}}}))
		dis := make([]*pps.DatumInfo, ppsserver.DefaultDatumBatchSize)
		n, err := grpcutil.Read[*pps.DatumInfo](datumClient, dis)
		require.NoError(t, err)
		require.Equal(t, ppsserver.DefaultDatumBatchSize, n)
		receivedDatums := make(map[string]int)
		for i := range n {
			require.Equal(t, 1, len(dis[i].Data))
			receivedDatums[dis[i].Data[0].File.Path]++
		}

		remaining := 2*numFiles - ppsserver.DefaultDatumBatchSize
		require.NoError(t, datumClient.Send(&pps.CreateDatumRequest{Body: &pps.CreateDatumRequest_Continue{Continue: &pps.ContinueCreateDatumRequest{Number: int32(remaining)}}}))
		dis = make([]*pps.DatumInfo, remaining+1)
		n, err = grpcutil.Read[*pps.DatumInfo](datumClient, dis)
		require.True(t, stream.IsEOS(err))
		require.Equal(t, remaining, n)
		for i := range n {
			require.Equal(t, 1, len(dis[i].Data))
			receivedDatums[dis[i].Data[0].File.Path]++
		}

		for i := range numFiles {
			count, ok := receivedDatums[fmt.Sprintf("/file%d", i)]
			require.True(t, ok, fmt.Sprintf("datum with path /file%d not received", i))
			require.Equal(t, 2, count, fmt.Sprintf("datum with path /file%d received %d times; expected 2", i, count))
		}
	})
	t.Run("CrossInput", func(t *testing.T) {
		datumClient, err := pc.PpsAPIClient.CreateDatum(ctx)
		require.NoError(t, err)
		input := client.NewCrossInput(
			client.NewPFSInput(pfs.DefaultProjectName, repo, "/file?"),
			client.NewPFSInput(pfs.DefaultProjectName, repo, "/file?"),
		)
		require.NoError(t, datumClient.Send(&pps.CreateDatumRequest{Body: &pps.CreateDatumRequest_Start{Start: &pps.StartCreateDatumRequest{Input: input, Number: 50}}}))
		receivedDatums := make(map[string]int)
		dis := make([]*pps.DatumInfo, 50)
		n, err := grpcutil.Read[*pps.DatumInfo](datumClient, dis)
		require.NoError(t, err)
		require.Equal(t, 50, n)
		for i := range n {
			require.Equal(t, 2, len(dis[i].Data))
			receivedDatums[dis[i].Data[0].File.Path]++
			receivedDatums[dis[i].Data[1].File.Path]++
		}

		require.NoError(t, datumClient.Send(&pps.CreateDatumRequest{Body: &pps.CreateDatumRequest_Continue{Continue: &pps.ContinueCreateDatumRequest{}}}))
		dis = make([]*pps.DatumInfo, 51)
		n, err = grpcutil.Read[*pps.DatumInfo](datumClient, dis)
		require.True(t, stream.IsEOS(err))
		require.Equal(t, 50, n)
		for i := range n {
			require.Equal(t, 2, len(dis[i].Data))
			receivedDatums[dis[i].Data[0].File.Path]++
			receivedDatums[dis[i].Data[1].File.Path]++
		}

		for i := range 10 {
			count, ok := receivedDatums[fmt.Sprintf("/file%d", i)]
			require.True(t, ok, fmt.Sprintf("datum with path /file%d not received", i))
			require.Equal(t, 20, count, fmt.Sprintf("datum with path /file%d received %d times; expected 20", i, count))
		}
	})
	t.Run("JoinInput", func(t *testing.T) {
		datumClient, err := pc.PpsAPIClient.CreateDatum(ctx)
		require.NoError(t, err)
		input := client.NewJoinInput(
			client.NewPFSInputOpts(repo, pfs.DefaultProjectName, repo, "master", "/file(?)", "$1", "", false, false, nil),
			client.NewPFSInputOpts(repo, pfs.DefaultProjectName, repo, "master", "/file1(?)", "$1", "", false, false, nil),
		)
		require.NoError(t, datumClient.Send(&pps.CreateDatumRequest{Body: &pps.CreateDatumRequest_Start{Start: &pps.StartCreateDatumRequest{Input: input}}}))
		dis := make([]*pps.DatumInfo, 11)
		n, err := grpcutil.Read[*pps.DatumInfo](datumClient, dis)
		require.True(t, stream.IsEOS(err))
		require.Equal(t, 10, n)
		receivedDatums := make(map[string]bool)
		for i := range n {
			require.Equal(t, 2, len(dis[i].Data))
			var shorter, longer string
			if len(dis[i].Data[0].File.Path) < len(dis[i].Data[1].File.Path) {
				shorter = dis[i].Data[0].File.Path
				longer = dis[i].Data[1].File.Path
			} else {
				shorter = dis[i].Data[1].File.Path
				longer = dis[i].Data[0].File.Path
			}
			require.Equal(t, shorter[5], longer[6])
			receivedDatums[string(shorter[5])] = true
		}
		for i := range 10 {
			require.True(t, receivedDatums[strconv.Itoa(i)])
		}
	})
	t.Run("GroupInput", func(t *testing.T) {
		datumClient, err := pc.PpsAPIClient.CreateDatum(ctx)
		require.NoError(t, err)
		input := client.NewGroupInput(
			client.NewPFSInputOpts(repo, pfs.DefaultProjectName, repo, "master", "/file(?)*", "", "$1", false, false, nil),
			client.NewPFSInputOpts(repo, pfs.DefaultProjectName, repo, "master", "/file(?)*", "", "$1", false, false, nil),
		)
		require.NoError(t, datumClient.Send(&pps.CreateDatumRequest{Body: &pps.CreateDatumRequest_Start{Start: &pps.StartCreateDatumRequest{Input: input}}}))
		dis := make([]*pps.DatumInfo, 11)
		n, err := grpcutil.Read[*pps.DatumInfo](datumClient, dis)
		require.True(t, stream.IsEOS(err))
		require.Equal(t, 10, n)
		receivedDatums := make(map[string]bool)
		for i := range n {
			require.True(t, len(dis[i].Data) >= 2)
			for j := range len(dis[i].Data) {
				require.Equal(t, dis[i].Data[0].File.Path[5], dis[i].Data[j].File.Path[5])
			}
			receivedDatums[string(dis[i].Data[0].File.Path[5])] = true
		}
		t.Log(receivedDatums)
		for i := range 10 {
			require.True(t, receivedDatums[strconv.Itoa(i)], fmt.Sprintf("datum with prefix /file%d not received", i))
		}
	})
	t.Run("WrongRequestMessage", func(t *testing.T) {
		datumClient, err := pc.PpsAPIClient.CreateDatum(ctx)
		require.NoError(t, err)
		require.NoError(t, datumClient.Send(&pps.CreateDatumRequest{Body: &pps.CreateDatumRequest_Continue{Continue: &pps.ContinueCreateDatumRequest{}}}))
		_, err = datumClient.Recv()
		require.ErrorContains(t, err, "first message must be a StartCreateDatumRequest message")

		datumClient, err = pc.PpsAPIClient.CreateDatum(ctx)
		require.NoError(t, err)
		require.NoError(t, datumClient.Send(&pps.CreateDatumRequest{Body: &pps.CreateDatumRequest_Start{Start: &pps.StartCreateDatumRequest{Input: input, Number: 1}}}))
		_, err = datumClient.Recv()
		require.NoError(t, err)
		require.NoError(t, datumClient.Send(&pps.CreateDatumRequest{Body: &pps.CreateDatumRequest_Start{Start: &pps.StartCreateDatumRequest{Input: input}}}))
		_, err = datumClient.Recv()
		require.ErrorContains(t, err, "subsequent messages must be a ContinueCreateDatumRequest message")
	})
}

func TestRenderTemplate(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
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
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
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
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
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
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	resp, err := env.PPSServer.GetClusterDefaults(ctx, &pps.GetClusterDefaultsRequest{})
	require.NoError(t, err, "GetClusterDefaults failed")
	require.NotEqual(t, "", resp.ClusterDefaultsJson)
	var cd pps.ClusterDefaults
	require.NoError(t, json.Unmarshal([]byte(resp.ClusterDefaultsJson), &cd), "cluster defaults JSON must unmarshal")
}

func TestSetClusterDefaults(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)

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

	t.Run("DryRun", func(t *testing.T) {
		getResp, err := env.PPSServer.GetClusterDefaults(ctx, &pps.GetClusterDefaultsRequest{})
		require.NoError(t, err, "GetClusterDefaults failed")

		var originalDefaults pps.ClusterDefaults
		err = json.Unmarshal([]byte(getResp.GetClusterDefaultsJson()), &originalDefaults)
		require.NoError(t, err, "unmarshal retrieved cluster defaults")

		repo := tu.UniqueString("input")
		pipeline := tu.UniqueString("pipeline")
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
		}
	}`
		tmpl, err := template.New("pipeline").Parse(pipelineTemplate)
		require.NoError(t, err, "template must parse")
		var buf bytes.Buffer
		err = tmpl.Execute(&buf, struct {
			ProjectName, PipelineName, RepoName string
		}{pfs.DefaultProjectName, pipeline, repo})
		require.NoError(t, err, "template must execute")
		cpr, err := env.PachClient.PpsAPIClient.CreatePipelineV2(ctx, &pps.CreatePipelineV2Request{
			CreatePipelineRequestJson: buf.String(),
		})
		require.NoError(t, err, "CreatePipelineV2 must succeed")
		require.False(t, cpr.EffectiveCreatePipelineRequestJson == "", "response includes effective JSON")
		var req pps.CreatePipelineRequest
		require.NoError(t, protojson.Unmarshal([]byte(cpr.EffectiveCreatePipelineRequestJson), &req), "unmarshalling effective JSON must not error")
		require.False(t, req.Autoscaling, "autoscaling defaults to false")

		resp, err := env.PPSServer.SetClusterDefaults(ctx, &pps.SetClusterDefaultsRequest{
			ClusterDefaultsJson: `{"create_pipeline_request": {"autoscaling": true, "datumTries": 1234}}`,
			Regenerate:          true,
			DryRun:              true,
		})
		require.NoError(t, err, "SetClusterDefaults failed")
		require.Len(t, resp.AffectedPipelines, 1, "pipelines should be affected by setting defaults")
		getResp, err = env.PPSServer.GetClusterDefaults(ctx, &pps.GetClusterDefaultsRequest{})
		require.NoError(t, err, "GetClusterDefaults failed")

		var defaults pps.ClusterDefaults
		err = json.Unmarshal([]byte(getResp.GetClusterDefaultsJson()), &defaults)
		require.NoError(t, err, "unmarshal retrieved cluster defaults")
		require.True(t, proto.Equal(&originalDefaults, &defaults), "defaults should not have actually changed")

		pr, err := env.PachClient.PpsAPIClient.InspectPipeline(ctx, &pps.InspectPipelineRequest{Pipeline: &pps.Pipeline{Project: &pfs.Project{Name: pfs.DefaultProjectName}, Name: pipeline}})
		require.NoError(t, err, "InspectPipeline should succeed")
		require.False(t, pr.EffectiveSpecJson == "", "effective spec should not be empty")
		require.NoError(t, protojson.Unmarshal([]byte(pr.EffectiveSpecJson), &req), "effective spec should unmarshal")
		require.False(t, req.Autoscaling, "defaults did not actually change")
	})

	t.Run("ValidDetails", func(t *testing.T) {
		err := env.PachClient.DeleteAll(env.PachClient.Ctx())
		require.NoError(t, err)
		repo := tu.UniqueString("input")
		pipeline := tu.UniqueString("pipeline")
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
		}
	}`
		tmpl, err := template.New("pipeline").Parse(pipelineTemplate)
		require.NoError(t, err, "template must parse")
		var buf bytes.Buffer
		require.NoError(t, tmpl.Execute(&buf, struct {
			ProjectName, PipelineName, RepoName string
		}{pfs.DefaultProjectName, pipeline, repo}), "template must execute")
		cpr, err := env.PachClient.PpsAPIClient.CreatePipelineV2(ctx, &pps.CreatePipelineV2Request{
			CreatePipelineRequestJson: buf.String(),
		})
		require.NoError(t, err, "CreatePipelineV2 must succeed")
		require.False(t, cpr.EffectiveCreatePipelineRequestJson == "", "response includes effective JSON")
		var req pps.CreatePipelineRequest
		require.NoError(t, protojson.Unmarshal([]byte(cpr.EffectiveCreatePipelineRequestJson), &req), "unmarshalling effective JSON must not error")
		require.False(t, req.Autoscaling, "autoscaling defaults to false")

		resp, err := env.PPSServer.SetClusterDefaults(ctx, &pps.SetClusterDefaultsRequest{
			ClusterDefaultsJson: `{"create_pipeline_request": {"autoscaling": true}}`,
			Regenerate:          true,
		})
		require.NoError(t, err, "SetClusterDefaults failed")
		require.Len(t, resp.AffectedPipelines, 1, "pipelines should be affected by setting defaults")
		getResp, err := env.PPSServer.GetClusterDefaults(ctx, &pps.GetClusterDefaultsRequest{})
		require.NoError(t, err, "GetClusterDefaults failed")

		var defaults pps.ClusterDefaults
		err = json.Unmarshal([]byte(getResp.GetClusterDefaultsJson()), &defaults)
		require.NoError(t, err, "unmarshal retrieved cluster defaults")
		require.NotNil(t, defaults.CreatePipelineRequest, "Create Pipeline Request should not be nil after SetClusterDefaults")
		require.True(t, defaults.CreatePipelineRequest.Autoscaling, "default autoscaling should be true after SetClusterDefaults")

		pr, err := env.PachClient.PpsAPIClient.InspectPipeline(ctx, &pps.InspectPipelineRequest{Pipeline: &pps.Pipeline{Project: &pfs.Project{Name: pfs.DefaultProjectName}, Name: pipeline}})
		require.NoError(t, err, "InspectPipeline should succeed")
		require.False(t, pr.EffectiveSpecJson == "", "effective spec should not be empty")
		require.NoError(t, protojson.Unmarshal([]byte(pr.EffectiveSpecJson), &req), "effective spec should unmarshal")
		require.True(t, req.Autoscaling, "defaults propagated")
	})
}

// TestCreatePipelineMultipleNames tests that camelCase and snake_case names map
// to the same thing.
func TestCreatePipelineMultipleNames(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)

	_, err := env.PPSServer.SetClusterDefaults(ctx, &pps.SetClusterDefaultsRequest{
		ClusterDefaultsJson: `{"create_pipeline_request": {"datum_tries": 17, "autoscaling": true, "salt": "mysalt"}}`,
	})
	require.NoError(t, err, "SetClusterDefaults failed")

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
		"datumTries": 4,
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
	require.Equal(t, int64(4), req.DatumTries, "default and spec names map")
	require.False(t, req.Autoscaling, "spec must override default")
	require.Equal(t, req.Salt, "mysalt", "default must apply if not overridden")
	// validate that the user and effective specs are correct
	r, err := env.PachClient.PpsAPIClient.InspectPipeline(ctx, &pps.InspectPipelineRequest{Pipeline: &pps.Pipeline{Name: pipeline}})
	require.NoError(t, err, "InspectPipeline must succeed")

	require.NotEqual(t, r.UserSpecJson, "", "user spec must not be blank")
	var spec map[string]any
	err = unmarshalJSON(r.UserSpecJson, &spec)
	require.NoError(t, err, "user spec %s must unmarshal", r.UserSpecJson)
	_, ok := spec["salt"]
	require.False(t, ok, "salt must not be found in the user spec")
	require.False(t, spec["autoscaling"].(bool), "user spec must have autoscaling set to false")

	require.NotEqual(t, r.EffectiveSpecJson, "", "effective spec must not be blank")
	err = unmarshalJSON(r.EffectiveSpecJson, &spec)
	require.NoError(t, err, "effective spec %s must unmarshal", r.EffectiveSpecJson)
	require.Equal(t, spec["salt"], "mysalt", "salt must be set in the effective spec")
	require.False(t, spec["autoscaling"].(bool), "effective spec must have autoscaling set to false")
	require.Equal(t, spec["datumTries"], json.Number("4"), "effective spec must have datumTries = 4")
}

// TestDefaultPropagation tests that changed defaults propagate to a pipeline.
func TestDefaultPropagation(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)

	_, err := env.PPSServer.SetClusterDefaults(ctx, &pps.SetClusterDefaultsRequest{
		ClusterDefaultsJson: `{"create_pipeline_request": {"datum_tries": 17, "autoscaling": true, "metadata": {"annotations": {"foo": "bar"}}}}`})
	require.NoError(t, err, "SetClusterDefaults failed")

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
		"datumTries": 4
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
	require.Equal(t, int64(4), req.DatumTries, "default and spec names map")
	require.True(t, req.Autoscaling, "default must apply to spec")
	require.Equal(t, "bar", req.Metadata.Annotations["foo"], "default must apply to spec")

	// validate that the user and effective specs are correct
	r, err := env.PachClient.PpsAPIClient.InspectPipeline(ctx, &pps.InspectPipelineRequest{Pipeline: &pps.Pipeline{Name: pipeline}, Details: true})
	require.NoError(t, err, "InspectPipeline must succeed")

	require.NotEqual(t, r.UserSpecJson, "", "user spec must not be blank")
	var spec map[string]any
	err = unmarshalJSON(r.UserSpecJson, &spec)
	require.NoError(t, err, "user spec %s must unmarshal", r.UserSpecJson)
	_, ok := spec["metadata"]
	require.False(t, ok, "metadata must not be found in the user spec")
	_, ok = spec["autoscaling"]
	require.False(t, ok, "autoscaling must not be found in the user spec")

	require.NotEqual(t, r.EffectiveSpecJson, "", "user spec must not be blank")
	err = unmarshalJSON(r.EffectiveSpecJson, &spec)
	require.NoError(t, err, "effective spec %s must unmarshal", r.EffectiveSpecJson)
	_, ok = spec["metadata"]
	require.True(t, ok, "metadata must  be found in the effective spec")
	require.True(t, spec["autoscaling"].(bool), "autoscaling must be true in the effective spec")
	require.Equal(t, spec["datumTries"], json.Number("4"), "effective spec must have datumTries = 4")

	version := r.Version
	salt := r.Details.Salt
	// changing the defaults without regenerate set should have no effect on the pipeline
	_, err = env.PPSServer.SetClusterDefaults(ctx, &pps.SetClusterDefaultsRequest{
		ClusterDefaultsJson: `{"create_pipeline_request": {"datum_tries": 17, "autoscaling": false, "metadata": {"annotations": {"baz": "quux"}}}}`,
	})
	require.NoError(t, err, "SetClusterDefaults failed")
	r, err = env.PPSServer.InspectPipeline(ctx, &pps.InspectPipelineRequest{Pipeline: r.Pipeline, Details: true})
	require.NoError(t, err, "InspectPipeline must succeed")
	require.Equal(t, version, r.Version, "SetClusterDefaults with regenerate=false should have no effect on pipelines")
	require.Equal(t, true, r.Details.Autoscaling, "SetClusterDefaults with regenerate=false should have no effect on pipelines")
	require.Equal(t, "bar", r.Details.Metadata.Annotations["foo"], "SetClusterDefaults with regenerate=false should have no effect on pipelines")
	require.Equal(t, "", r.Details.Metadata.Annotations["baz"], "SetClusterDefaults with regenerate=false should have no effect on pipelines")

	// changing the defaults with regenerate set should have an effect, but without reprocess set the salt should remain the same
	_, err = env.PPSServer.SetClusterDefaults(ctx, &pps.SetClusterDefaultsRequest{
		ClusterDefaultsJson: `{"create_pipeline_request": {"datum_tries": 17, "autoscaling": true, "metadata": {"annotations": {"baz": "quux"}}, "salt": "mysalt2"}}`,
		Regenerate:          true,
	})
	require.NoError(t, err, "SetClusterDefaults failed")
	r, err = env.PPSServer.InspectPipeline(ctx, &pps.InspectPipelineRequest{Pipeline: r.Pipeline, Details: true})
	require.NoError(t, err, "InspectPipeline must succeed")
	require.NotEqual(t, version, r.Version, "SetClusterDefaults with regenerate=true should have an effect on pipelines")
	require.Equal(t, true, r.Details.Autoscaling, "SetClusterDefaults with regenerate=true should have an effect on pipelines")
	require.Equal(t, "", r.Details.Metadata.Annotations["foo"], "SetClusterDefaults with regenerate=true should have an effect on pipelines")
	require.Equal(t, "quux", r.Details.Metadata.Annotations["baz"], "SetClusterDefaults with regenerate=true should have an effect on pipelines")
	require.Equal(t, salt, r.Details.Salt, "SetClusterDefaults with regenerate=true should have an effect, but not on salt")

	version = r.Version
	// changing the defaults with regenerate set but no effective changes should have … no effect on the pipeline
	_, err = env.PPSServer.SetClusterDefaults(ctx, &pps.SetClusterDefaultsRequest{
		ClusterDefaultsJson: `{"create_pipeline_request": {"datum_tries": 12, "autoscaling": true, "metadata": {"annotations": {"baz": "quux"}}, "salt": "mysalt2"}}`,
		Regenerate:          true,
	})
	require.NoError(t, err, "SetClusterDefaults failed")
	r, err = env.PPSServer.InspectPipeline(ctx, &pps.InspectPipelineRequest{Pipeline: r.Pipeline, Details: true})
	require.NoError(t, err, "InspectPipeline must succeed")
	require.Equal(t, r.Version, version, "SetClusterDefaults which does not change an effective spec should have no effect")
	require.Equal(t, int64(4), r.Details.DatumTries, "SetClusterDefaults which does not change an effective spec should have no effect")
	require.Equal(t, true, r.Details.Autoscaling, "SetClusterDefaults which does not change an effective spec should have no effect")
	require.Equal(t, salt, r.Details.Salt, "SetClusterDefaults which does not change an effective spec should have no effect")

	// changing the defaults and setting reprocess should update the salt
	_, err = env.PPSServer.SetClusterDefaults(ctx, &pps.SetClusterDefaultsRequest{
		ClusterDefaultsJson: `{"create_pipeline_request": {"datum_tries": 17, "autoscaling": true, "metadata": {"annotations": {"foobar": "bazquux"}}, "salt": "mysalt2"}}`,
		Regenerate:          true,
		Reprocess:           true,
	})
	require.NoError(t, err, "SetClusterDefaults failed")
	r, err = env.PPSServer.InspectPipeline(ctx, &pps.InspectPipelineRequest{Pipeline: r.Pipeline, Details: true})
	require.NoError(t, err, "InspectPipeline must succeed")
	require.NotEqual(t, version, r.Version, "SetClusterDefaults with regenerate=true should have an effect on pipelines")
	require.Equal(t, true, r.Details.Autoscaling, "SetClusterDefaults with regenerate=true should have an effect on pipelines")
	require.Equal(t, "", r.Details.Metadata.Annotations["foo"], "SetClusterDefaults with regenerate=true should have an effect on pipelines")
	require.Equal(t, "", r.Details.Metadata.Annotations["baz"], "SetClusterDefaults with regenerate=true should have an effect on pipelines")
	require.Equal(t, "bazquux", r.Details.Metadata.Annotations["foobar"], "SetClusterDefaults with regenerate=true should have an effect on pipelines")
	require.NotEqual(t, salt, r.Details.Salt, "SetClusterDefaults with regenerate=true and reprocess=true should update salt")

}

func unmarshalJSON(s string, v any) error {
	d := json.NewDecoder(strings.NewReader(s))
	d.UseNumber()
	return errors.Wrapf(d.Decode(v), "could not unmarshal %q as JSON", s)
}

// TestCreatePipelineDryRun tests that creating a pipeline with dry run set to
// true does not create a pipeline, but does return a correct effective spec.
func TestCreatePipelineDryRun(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)

	_, err := env.PPSServer.SetClusterDefaults(ctx, &pps.SetClusterDefaultsRequest{
		ClusterDefaultsJson: `{"create_pipeline_request": {"datum_tries": 17, "autoscaling": true}}`,
	})
	require.NoError(t, err, "SetClusterDefaults failed")

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
		"datumTries": 4,
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
		DryRun:                    true,
	})
	require.NoError(t, err, "CreatePipelineV2 must succeed")
	require.False(t, resp.EffectiveCreatePipelineRequestJson == "", "response includes effective JSON")
	var req pps.CreatePipelineRequest
	require.NoError(t, protojson.Unmarshal([]byte(resp.EffectiveCreatePipelineRequestJson), &req), "unmarshalling effective JSON must not error")
	require.Equal(t, int64(4), req.DatumTries, "default and spec names map")
	require.False(t, req.Autoscaling, "spec must override default")

	if _, err = env.PachClient.PpsAPIClient.InspectPipeline(ctx, &pps.InspectPipelineRequest{Pipeline: &pps.Pipeline{Project: &pfs.Project{Name: pfs.DefaultProjectName}, Name: pipeline}}); err == nil {
		t.Error("InspectPipeline should fail if pipeline was not created")
	}
}

func TestListPipelinePagination(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	collectPipelinePage := func(page *pps.PipelinePage, projects []string) []*pps.PipelineInfo {
		var prjs []*pfs.Project
		for _, p := range projects {
			prjs = append(prjs, client.NewProject(p))
		}
		piClient, err := env.PachClient.PpsAPIClient.ListPipeline(ctx, &pps.ListPipelineRequest{
			Page:     page,
			Projects: prjs,
		})
		require.NoError(t, err)
		pis, err := grpcutil.Collect[*pps.PipelineInfo](piClient, 1000)
		require.NoError(t, err)
		return pis
	}
	repo := "input"
	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
	createPipeline := func(project, name string) {
		require.NoError(t, env.PachClient.CreatePipeline(
			project,
			name,
			"", /* default image*/
			[]string{"cp", "-r", "/pfs/in", "/pfs/out"},
			nil, /* stdin */
			nil, /* spec */
			&pps.Input{Pfs: &pps.PFSInput{Project: pfs.DefaultProjectName, Repo: repo, Glob: "/*", Name: "in"}},
			"",   /* output */
			true, /* update */
		))
	}
	projects := []string{"a", "b"}
	for _, prj := range projects {
		require.NoError(t, env.PachClient.CreateProject(prj))
	}
	for i, p := range []string{"A", "B", "C", "D", "E"} {
		proj := projects[0]
		if i%2 == 1 {
			proj = projects[1]
		}
		createPipeline(proj, p)
	}
	// page larger than number of repos
	ps := collectPipelinePage(&pps.PipelinePage{
		PageSize:  20,
		PageIndex: 0,
	}, nil)
	require.Equal(t, 5, len(ps))
	ps = collectPipelinePage(&pps.PipelinePage{
		PageSize:  3,
		PageIndex: 0,
	}, nil)
	assertPipelineSequence(t, []string{"E", "D", "C"}, ps)
	ps = collectPipelinePage(&pps.PipelinePage{
		PageSize:  3,
		PageIndex: 1,
	}, nil)
	assertPipelineSequence(t, []string{"B", "A"}, ps)
	// page overbounds
	ps = collectPipelinePage(&pps.PipelinePage{
		PageSize:  3,
		PageIndex: 2,
	}, nil)
	assertPipelineSequence(t, []string{}, ps)
	ps = collectPipelinePage(&pps.PipelinePage{
		PageSize:  2,
		PageIndex: 1,
	}, nil)
	assertPipelineSequence(t, []string{"C", "B"}, ps)
	// filter by project & pagination
	ps = collectPipelinePage(&pps.PipelinePage{
		PageSize:  1,
		PageIndex: 1,
	}, []string{"b"})
	assertPipelineSequence(t, []string{"B"}, ps)
}

func assertPipelineSequence(t *testing.T, names []string, pipelines []*pps.PipelineInfo) {
	require.Equal(t, len(names), len(pipelines))
	for i, n := range names {
		require.Equal(t, n, pipelines[i].Pipeline.Name)
	}
}
