package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"strconv"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	v1 "k8s.io/api/core/v1"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/ancestry"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachd"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	at "github.com/pachyderm/pachyderm/v2/src/server/auth/server/testing"
	ppsserver "github.com/pachyderm/pachyderm/v2/src/server/pps/server"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/datum"

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

// TestCreatePipeline tests pipeline creation.
func TestCreatePipeline(t *testing.T) {
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
	})
	require.NoError(t, err, "CreatePipelineV2 must succeed")
	require.False(t, resp.EffectiveCreatePipelineRequestJson == "", "response includes effective JSON")
	var req pps.CreatePipelineRequest
	require.NoError(t, protojson.Unmarshal([]byte(resp.EffectiveCreatePipelineRequestJson), &req), "unmarshalling effective JSON must not error")
	require.Equal(t, int64(4), req.DatumTries, "default and spec names map")
	require.False(t, req.Autoscaling, "spec must override default")

	ir, err := env.PachClient.PpsAPIClient.InspectPipeline(ctx, &pps.InspectPipelineRequest{
		Pipeline: &pps.Pipeline{Project: &pfs.Project{Name: pfs.DefaultProjectName}, Name: pipeline},
		Details:  true,
	})
	require.NoError(t, err, "InspectPipeline")
	require.NotNil(t, ir)
	require.NotNil(t, ir.Details)
	require.NotNil(t, ir.Details.UpdatedAt)
	require.True(t, !ir.Details.UpdatedAt.AsTime().Before(ir.Details.CreatedAt.AsTime()))
}

func testCreatedBy(t testing.TB, c *client.APIClient, username string) (string, string) {
	repo := tu.UniqueString("input")
	pipeline := tu.UniqueString("pipeline")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repo))
	resp, err := c.PpsAPIClient.CreatePipelineV2(c.Ctx(), &pps.CreatePipelineV2Request{
		CreatePipelineRequestJson: fmt.Sprintf(`{
		"pipeline": {
			"project": {
				"name": %q
			},
			"name": %q
		},
		"transform": {
			"cmd": ["cp", "-r", "/pfs/in", "/pfs/out"]
		},
		"input": {
			"pfs": {
				"project": "default",
				"repo": %q,
				"glob": "/*",
				"name": "in"
			}
		},
		"datumTries": 4,
		"autoscaling": false
	}`, pfs.DefaultProjectName, pipeline, repo),
	})
	require.NoError(t, err, "CreatePipelineV2 must succeed")
	require.False(t, resp.EffectiveCreatePipelineRequestJson == "", "response includes effective JSON")
	var req pps.CreatePipelineRequest
	require.NoError(t, protojson.Unmarshal([]byte(resp.EffectiveCreatePipelineRequestJson), &req), "unmarshalling effective JSON must not error")
	require.Equal(t, int64(4), req.DatumTries, "default and spec names map")
	require.False(t, req.Autoscaling, "spec must override default")

	ir, err := c.PpsAPIClient.InspectPipeline(c.Ctx(), &pps.InspectPipelineRequest{
		Pipeline: &pps.Pipeline{Project: &pfs.Project{Name: pfs.DefaultProjectName}, Name: pipeline},
		Details:  true,
	})
	require.NoError(t, err, "InspectPipeline")
	require.NotNil(t, ir)
	require.NotNil(t, ir.Details)
	require.Equal(t, username, ir.Details.CreatedBy)
	return repo, pipeline
}

func TestCreatedByAndAt(t *testing.T) {
	t.Parallel()
	c := at.EnvWithAuth(t).PachClient

	t.Run("root", func(t *testing.T) {
		testCreatedBy(t, tu.AuthenticateClient(t, c, auth.RootUser), "pach:root")
	})

	t.Run("user", func(t *testing.T) {

		aliceName, alice := tu.RandomRobot(t, c, "alice")
		repo, pipeline := testCreatedBy(t, alice, aliceName)
		ir, err := c.PpsAPIClient.InspectPipeline(c.Ctx(), &pps.InspectPipelineRequest{
			Pipeline: &pps.Pipeline{Project: &pfs.Project{Name: pfs.DefaultProjectName}, Name: pipeline},
			Details:  true,
		})
		require.NoError(t, err, "InspectPipeline")
		require.NotNil(t, ir)
		require.NotNil(t, ir.Details)
		createdAt := ir.Details.CreatedAt.AsTime()

		// Updating the pipeline does not change CreatedBy.
		c := tu.AuthenticateClient(t, c, auth.RootUser)
		_, err = c.PpsAPIClient.CreatePipelineV2(c.Ctx(), &pps.CreatePipelineV2Request{
			CreatePipelineRequestJson: fmt.Sprintf(`{
				"pipeline": {
					"project": {
						"name": %q
					},
					"name": %q
				},
				"transform": {
					"cmd": ["cp", "-r", "/pfs/in", "/pfs/out"]
				},
				"input": {
					"pfs": {
						"project": "default",
						"repo": %q,
						"glob": "/*",
						"name": "in"
					}
				},
				"datumTries": 3,
				"autoscaling": false
			}`, pfs.DefaultProjectName, pipeline, repo),
			Update: true,
		})

		require.NoError(t, err, "CreatePipelineV2 must succeed")
		ir, err = c.PpsAPIClient.InspectPipeline(c.Ctx(), &pps.InspectPipelineRequest{
			Pipeline: &pps.Pipeline{Project: &pfs.Project{Name: pfs.DefaultProjectName}, Name: pipeline},
			Details:  true,
		})
		require.NoError(t, err, "InspectPipeline")
		require.NotNil(t, ir)
		require.NotNil(t, ir.Details)
		require.Equal(t, aliceName, ir.Details.CreatedBy)
		require.Equal(t, createdAt, ir.Details.CreatedAt.AsTime())
	})
}

func TestCreatedByInTransaction(t *testing.T) {
	t.Parallel()
	c := at.EnvWithAuth(t).PachClient
	aliceName, alice := tu.RandomRobot(t, c, "alice")
	repo := tu.UniqueString("input")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repo))
	require.NoError(t, c.ModifyRepoRoleBinding(c.Ctx(), pfs.DefaultProjectName, repo, aliceName, []string{"repoReader"}), "ModifyRepoRoleBinding")

	txn, err := c.StartTransaction()
	require.NoError(t, err, "StartTransaction")
	tc := c.WithTransaction(txn)
	pipeline := tu.UniqueString("pipeline")
	_, err = tc.PpsAPIClient.CreatePipelineV2(tc.Ctx(), &pps.CreatePipelineV2Request{
		CreatePipelineRequestJson: fmt.Sprintf(`{
		"pipeline": {
			"project": {
				"name": %q
			},
			"name": %q
		},
		"transform": {
			"cmd": ["cp", "-r", "/pfs/in", "/pfs/out"]
		},
		"input": {
			"pfs": {
				"project": "default",
				"repo": %q,
				"glob": "/*",
				"name": "in"
			}
		},
		"datumTries": 4,
		"autoscaling": false
	}`, pfs.DefaultProjectName, pipeline, repo),
	})
	require.NoError(t, err, "CreatePipelineV2")

	ac := alice.WithTransaction(txn)
	_, err = ac.FinishTransaction(txn)
	require.NoError(t, err, "FinishTransaction")

	ir, err := c.PpsAPIClient.InspectPipeline(c.Ctx(), &pps.InspectPipelineRequest{
		Pipeline: &pps.Pipeline{Project: &pfs.Project{Name: pfs.DefaultProjectName}, Name: pipeline},
		Details:  true,
	})
	require.NoError(t, err, "InspectPipeline")
	require.NotNil(t, ir)
	require.NotNil(t, ir.Details)
	require.Equal(t, "pach:root", ir.Details.CreatedBy)
}

func TestPipelineUniqueness(t *testing.T) {
	t.Parallel()
	c := pachd.NewTestPachd(t)

	repo := tu.UniqueString("data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repo))
	pipelineName := tu.UniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipelineName,
		"",
		[]string{"bash"},
		[]string{""},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, repo, "/"),
		"",
		false,
	))
	err := c.CreatePipeline(pfs.DefaultProjectName,
		pipelineName,
		"",
		[]string{"bash"},
		[]string{""},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, repo, "/"),
		"",
		false,
	)
	require.YesError(t, err)
	require.Matches(t, "pipeline .*? already exists", err.Error())
}

// Checks that "time to first datum" for CreateDatum is faster than ListDatum
func BenchmarkCreateDatum(b *testing.B) {
	c := pachd.NewTestPachd(b)
	repo := tu.UniqueString("BenchmarkCreateDatum")
	require.NoError(b, c.CreateRepo(pfs.DefaultProjectName, repo))

	// Need multiple shards worth of files to see benefit of CreateDatum
	// compared to ListDatum
	numFiles := 5 * datum.ShardNumFiles
	commit, err := c.StartCommit(pfs.DefaultProjectName, repo, "master")
	require.NoError(b, err)
	require.NoError(b, c.WithModifyFileClient(commit, func(mfc client.ModifyFile) error {
		for i := 0; i < numFiles; i++ {
			require.NoError(b, mfc.PutFile(fmt.Sprintf("file-%d", i), strings.NewReader(""), client.WithAppendPutFile()))
		}
		return nil
	}))
	require.NoError(b, c.FinishCommit(pfs.DefaultProjectName, repo, "master", ""))

	inputs := []struct {
		name  string
		input *pps.Input
	}{
		{"PFS", client.NewPFSInput(pfs.DefaultProjectName, repo, "/*")},
		{"Union", client.NewUnionInput(
			client.NewPFSInput(pfs.DefaultProjectName, repo, "/*"),
			client.NewPFSInput(pfs.DefaultProjectName, repo, "/*"),
		)},
		{"Cross", client.NewCrossInput(
			client.NewPFSInput(pfs.DefaultProjectName, repo, "/file-??"),
			client.NewPFSInput(pfs.DefaultProjectName, repo, "/file-???"),
		)},
		{"Join", client.NewJoinInput(
			client.NewPFSInputOpts(repo, pfs.DefaultProjectName, repo, "master", "/file-?*(??)0", "$1", "", false, false, nil),
			client.NewPFSInputOpts(repo, pfs.DefaultProjectName, repo, "master", "/file-?0(??)0", "$1", "", false, false, nil),
		)},
		// No entry for Group because CreateDatum's streaming can't improve
		// time to first datum
	}

	for _, input := range inputs {
		b.Run(input.name+"-ListDatum", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				req := &pps.ListDatumRequest{Input: input.input}
				client, err := c.PpsAPIClient.ListDatum(c.Ctx(), req)
				require.NoError(b, err)
				_, err = client.Recv()
				require.NoError(b, err)
			}
		})
	}

	for _, input := range inputs {
		b.Run(input.name+"-CreateDatum", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				client, err := c.PpsAPIClient.CreateDatum(c.Ctx())
				require.NoError(b, err)
				req := &pps.CreateDatumRequest{Body: &pps.CreateDatumRequest_Start{Start: &pps.StartCreateDatumRequest{Input: input.input}}}
				require.NoError(b, client.Send(req))
				_, err = client.Recv()
				require.NoError(b, err)
			}
		})
	}
}

func TestSelfReferentialPipeline(t *testing.T) {
	t.Parallel()
	c := pachd.NewTestPachd(t)
	pipeline := tu.UniqueString("pipeline")
	require.YesError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"",
		[]string{"true"},
		nil,
		nil,
		client.NewPFSInput(pfs.DefaultProjectName, pipeline, "/"),
		"",
		false,
	))
}

func TestPipelineDescription(t *testing.T) {
	t.Parallel()
	c := pachd.NewTestPachd(t)

	dataRepo := tu.UniqueString("TestPipelineDescription_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	description := "pipeline description"
	pipeline := tu.UniqueString("TestPipelineDescription")
	_, err := c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline:    client.NewPipeline(pfs.DefaultProjectName, pipeline),
			Transform:   &pps.Transform{Cmd: []string{"true"}},
			Description: description,
			Input:       client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/"),
		})
	require.NoError(t, err)
	pi, err := c.InspectPipeline(pfs.DefaultProjectName, pipeline, true)
	require.NoError(t, err)
	require.Equal(t, description, pi.Details.Description)
}

func TestPipelineVersions(t *testing.T) {
	t.Parallel()
	c := pachd.NewTestPachd(t)

	dataRepo := tu.UniqueString("TestPipelineVersions_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	pipeline := tu.UniqueString("TestPipelineVersions")
	nVersions := 5
	for i := 0; i < nVersions; i++ {
		require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
			pipeline,
			"",
			[]string{fmt.Sprintf("%d", i)}, // an obviously illegal command, but the pipeline will never run
			nil,
			&pps.ParallelismSpec{
				Constant: 1,
			},
			client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
			"",
			i != 0,
		))
	}

	for i := 0; i < nVersions; i++ {
		pi, err := c.InspectPipeline(pfs.DefaultProjectName, ancestry.Add(pipeline, nVersions-1-i), true)
		require.NoError(t, err)
		require.Equal(t, fmt.Sprintf("%d", i), pi.Details.Transform.Cmd[0])
	}
}

// TestCreatePipelineErrorNoPipeline tests that sending a CreatePipeline requests to pachd with no
// 'pipeline' field doesn't kill pachd
func TestCreatePipelineErrorNoPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c := pachd.NewTestPachd(t)

	// Create input repo
	dataRepo := tu.UniqueString(t.Name() + "-data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	// Create pipeline w/ no pipeline field--make sure we get a response
	_, err := c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: nil,
			Transform: &pps.Transform{
				Cmd:   []string{"/bin/bash"},
				Stdin: []string{`cat foo >/pfs/out/file`},
			},
			Input: client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		})
	require.YesError(t, err)
	require.Matches(t, "request.Pipeline", err.Error())
}

// TestCreatePipelineError tests that sending a CreatePipeline requests to pachd with no 'transform'
// or 'pipeline' field doesn't kill pachd
func TestCreatePipelineError(t *testing.T) {
	t.Parallel()
	c := pachd.NewTestPachd(t)

	// Create input repo
	dataRepo := tu.UniqueString(t.Name() + "-data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	// Create pipeline w/ no transform--make sure we get a response (& make sure
	// it explains the problem)
	pipeline := tu.UniqueString("no-transform-")
	_, err := c.PpsAPIClient.CreatePipelineV2(
		context.Background(),
		&pps.CreatePipelineV2Request{
			CreatePipelineRequestJson: fmt.Sprintf(`{"pipeline": {"project": {"name": %q}, "name": %q}, "transform": null, "input": {"pfs": {"project": %q, "repo": %q, "glob": %q}}}`, pfs.DefaultProjectName, pipeline, pfs.DefaultProjectName, dataRepo, "/*"),
		})
	require.YesError(t, err)
	require.Matches(t, "transform", err.Error())
}

// TestCreatePipelineErrorNoCmd tests that sending a CreatePipeline request to
// pachd with no 'transform.cmd' field doesn't kill pachd
func TestCreatePipelineErrorNoCmd(t *testing.T) {
	t.Parallel()
	c := pachd.NewTestPachd(t)

	// Create input data
	dataRepo := tu.UniqueString(t.Name() + "-data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
	require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, dataRepo, "master", ""), "file", strings.NewReader("foo"), client.WithAppendPutFile()))

	// create pipeline
	pipeline := tu.UniqueString("no-cmd-")
	_, err := c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
			Transform: &pps.Transform{
				Cmd:   nil,
				Stdin: []string{`cat foo >/pfs/out/file`},
			},
			Input: client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		})
	require.NoError(t, err)
	time.Sleep(5 * time.Second) // give pipeline time to start

	require.NoErrorWithinTRetry(t, 30*time.Second, func() error {
		pipelineInfo, err := c.InspectPipeline(pfs.DefaultProjectName, pipeline, false)
		if err != nil {
			return err
		}
		if pipelineInfo.State != pps.PipelineState_PIPELINE_FAILURE {
			return errors.Errorf("pipeline should be in state FAILURE, not: %s", pipelineInfo.State.String())
		}
		return nil
	})
}

func TestSecrets(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c := pachd.NewTestPachd(t)

	b := []byte(
		`{
			"kind": "Secret",
			"apiVersion": "v1",
			"type":"Opaque",
			"metadata": {
				"name": "test-secret",
				"creationTimestamp": null
			},
			"data": {
				"mykey": "bXktdmFsdWU="
			}
		}`)
	require.NoError(t, c.CreateSecret(b))

	secretInfo, err := c.InspectSecret("test-secret")
	secretInfo.CreationTimestamp = nil
	require.NoError(t, err)
	require.Equal(t, &pps.SecretInfo{
		Secret: &pps.Secret{
			Name: "test-secret",
		},
		Type:              "Opaque",
		CreationTimestamp: nil,
	}, secretInfo)

	secretInfos, err := c.ListSecret()
	require.NoError(t, err)
	initialLength := len(secretInfos)

	require.NoError(t, c.DeleteSecret("test-secret"))

	secretInfos, err = c.ListSecret()
	require.NoError(t, err)
	require.Equal(t, initialLength-1, len(secretInfos))

	_, err = c.InspectSecret("test-secret")
	require.YesError(t, err)
}

// Test that an unauthenticated user can't call secrets APIS
func TestSecretsUnauthenticated(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	// Get an unauthenticated client
	c := pachd.NewTestPachd(t, pachd.ActivateAuthOption("root"))
	c.SetAuthToken("")

	b := []byte(
		`{
			"kind": "Secret",
			"apiVersion": "v1",
			"metadata": {
				"name": "test-secret",
				"creationTimestamp": null
			},
			"data": {
				"mykey": "bXktdmFsdWU="
			}
		}`)

	err := c.CreateSecret(b)
	require.YesError(t, err)
	require.Matches(t, "no authentication token", err.Error())

	_, err = c.InspectSecret("test-secret")
	require.YesError(t, err)
	require.Matches(t, "no authentication token", err.Error())

	_, err = c.ListSecret()
	require.YesError(t, err)
	require.Matches(t, "no authentication token", err.Error())

	err = c.DeleteSecret("test-secret")
	require.YesError(t, err)
	require.Matches(t, "no authentication token", err.Error())
}

// Regression test to make sure that pipeline creation doesn't crash pachd due to missing fields
func TestMalformedPipeline(t *testing.T) {
	t.Parallel()
	c := pachd.NewTestPachd(t)

	pipelineName := tu.UniqueString("MalformedPipeline")

	var err error
	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{})
	require.YesError(t, err)
	require.Matches(t, "request.Pipeline cannot be nil", err.Error())

	_, err = c.PpsAPIClient.CreatePipelineV2(c.Ctx(), &pps.CreatePipelineV2Request{
		CreatePipelineRequestJson: fmt.Sprintf(`{"pipeline": {"project": {"name": %q}, "name": %q}, "transform": null}`, pfs.DefaultProjectName, pipelineName)},
	)
	require.YesError(t, err)
	require.Matches(t, "must specify a transform", err.Error())

	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline:  client.NewPipeline(pfs.DefaultProjectName, pipelineName),
		Transform: &pps.Transform{},
		Input:     &pps.Input{},
	})
	require.YesError(t, err)
	require.Matches(t, "no input set", err.Error())

	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline:  client.NewPipeline(pfs.DefaultProjectName, pipelineName),
		Transform: &pps.Transform{},
		Service: &pps.Service{
			Type: string(v1.ServiceTypeNodePort),
		},
		ParallelismSpec: &pps.ParallelismSpec{},
	})
	require.YesError(t, err)
	require.Matches(t, "services can only be run with a constant parallelism of 1", err.Error())

	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline:   client.NewPipeline(pfs.DefaultProjectName, pipelineName),
		Transform:  &pps.Transform{},
		SpecCommit: &pfs.Commit{},
	})
	require.YesError(t, err)
	require.ErrorContains(t, err, "cannot pick commit")

	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline:   client.NewPipeline(pfs.DefaultProjectName, pipelineName),
		Transform:  &pps.Transform{},
		SpecCommit: &pfs.Commit{Branch: &pfs.Branch{}},
	})
	require.YesError(t, err)
	require.ErrorContains(t, err, "cannot pick commit")

	dataRepo := tu.UniqueString("TestMalformedPipeline_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	dataCommit := client.NewCommit(pfs.DefaultProjectName, dataRepo, "master", "")
	require.NoError(t, c.PutFile(dataCommit, "file", strings.NewReader("foo"), client.WithAppendPutFile()))

	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline:  client.NewPipeline(pfs.DefaultProjectName, pipelineName),
		Transform: &pps.Transform{},
		Input:     &pps.Input{Pfs: &pps.PFSInput{}},
	})
	require.YesError(t, err)
	require.Matches(t, "input must specify a name", err.Error())

	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline:  client.NewPipeline(pfs.DefaultProjectName, pipelineName),
		Transform: &pps.Transform{},
		Input:     &pps.Input{Pfs: &pps.PFSInput{Name: "data"}},
	})
	require.YesError(t, err)
	require.Matches(t, "input must specify a repo", err.Error())

	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline:  client.NewPipeline(pfs.DefaultProjectName, pipelineName),
		Transform: &pps.Transform{},
		Input:     &pps.Input{Pfs: &pps.PFSInput{Repo: dataRepo}},
	})
	require.YesError(t, err)
	require.Matches(t, "input must specify a glob", err.Error())

	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline:  client.NewPipeline(pfs.DefaultProjectName, pipelineName),
		Transform: &pps.Transform{},
		Input:     client.NewPFSInput(pfs.DefaultProjectName, "out", "/*"),
	})
	require.YesError(t, err)
	require.Matches(t, "input cannot be named out", err.Error())

	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline:  client.NewPipeline(pfs.DefaultProjectName, pipelineName),
		Transform: &pps.Transform{},
		Input:     &pps.Input{Pfs: &pps.PFSInput{Name: "out", Repo: dataRepo, Glob: "/*"}},
	})
	require.YesError(t, err)
	require.Matches(t, "input cannot be named out", err.Error())

	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline:  client.NewPipeline(pfs.DefaultProjectName, pipelineName),
		Transform: &pps.Transform{},
		Input:     &pps.Input{Pfs: &pps.PFSInput{Name: "data", Repo: "dne", Glob: "/*"}},
	})
	require.YesError(t, err)
	require.Matches(t, "dne[^ ]* not found", err.Error())

	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline:  client.NewPipeline(pfs.DefaultProjectName, pipelineName),
		Transform: &pps.Transform{},
		Input: client.NewCrossInput(
			client.NewPFSInput(pfs.DefaultProjectName, "foo", "/*"),
			client.NewPFSInput(pfs.DefaultProjectName, "foo", "/*"),
		),
	})
	require.YesError(t, err)
	require.Matches(t, "name \"foo\" was used more than once", err.Error())

	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline:  client.NewPipeline(pfs.DefaultProjectName, pipelineName),
		Transform: &pps.Transform{},
		Input:     &pps.Input{Cron: &pps.CronInput{}},
	})
	require.YesError(t, err)
	require.Matches(t, "input must specify a name", err.Error())

	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline:  client.NewPipeline(pfs.DefaultProjectName, pipelineName),
		Transform: &pps.Transform{},
		Input:     &pps.Input{Cron: &pps.CronInput{Name: "cron"}},
	})
	require.YesError(t, err)
	require.Matches(t, "Empty spec string", err.Error())

	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline:  client.NewPipeline(pfs.DefaultProjectName, pipelineName),
		Transform: &pps.Transform{},
		Input:     &pps.Input{Cross: []*pps.Input{}},
	})
	require.YesError(t, err)
	require.Matches(t, "no input set", err.Error())

	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline:  client.NewPipeline(pfs.DefaultProjectName, pipelineName),
		Transform: &pps.Transform{},
		Input:     &pps.Input{Union: []*pps.Input{}},
	})
	require.YesError(t, err)
	require.Matches(t, "no input set", err.Error())

	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline:  client.NewPipeline(pfs.DefaultProjectName, pipelineName),
		Transform: &pps.Transform{},
		Input:     &pps.Input{Join: []*pps.Input{}},
	})
	require.YesError(t, err)
	require.Matches(t, "no input set", err.Error())
}

func TestInterruptedUpdatePipelineInTransaction(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c := pachd.NewTestPachd(t)
	inputA := tu.UniqueString("A")
	inputB := tu.UniqueString("B")
	inputC := tu.UniqueString("C")
	pipeline := tu.UniqueString("pipeline")

	createPipeline := func(c *client.APIClient, input string, update bool) error {
		return c.CreatePipeline(pfs.DefaultProjectName,
			pipeline,
			"",
			[]string{"bash"},
			[]string{fmt.Sprintf("cp /pfs/%s/* /pfs/out/", input)},
			&pps.ParallelismSpec{
				Constant: 1,
			},
			client.NewPFSInput(pfs.DefaultProjectName, input, "/*"),
			"",
			update,
		)
	}

	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, inputA))
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, inputB))
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, inputC))
	require.NoError(t, createPipeline(c, inputA, false))

	txn, err := c.StartTransaction()
	require.NoError(t, err)

	require.NoError(t, createPipeline(c.WithTransaction(txn), inputB, true))
	require.NoError(t, createPipeline(c, inputC, true))

	_, err = c.FinishTransaction(txn)
	require.NoError(t, err)
	// make sure the final pipeline is the third version, with the input from in the transaction
	pipelineInfo, err := c.InspectPipeline(pfs.DefaultProjectName, pipeline, true)
	require.NoError(t, err)
	require.Equal(t, uint64(3), pipelineInfo.Version)
	require.NotNil(t, pipelineInfo.Details.Input.Pfs)
	require.Equal(t, inputB, pipelineInfo.Details.Input.Pfs.Repo)
}

func basicPipelineReq(name, input string) *pps.CreatePipelineRequest {
	return &pps.CreatePipelineRequest{
		Pipeline: client.NewPipeline(pfs.DefaultProjectName, name),
		Transform: &pps.Transform{
			Cmd: []string{"bash"},
			Stdin: []string{
				fmt.Sprintf("cp /pfs/%s/* /pfs/out/", input),
			},
		},
		ParallelismSpec: &pps.ParallelismSpec{
			Constant: 1,
		},
		Input: client.NewPFSInput(pfs.DefaultProjectName, input, "/*"),
	}
}

func TestPipelineAncestry(t *testing.T) {
	t.Parallel()
	c := pachd.NewTestPachd(t)

	dataRepo := tu.UniqueString(t.Name())
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	pipeline := tu.UniqueString("pipeline")
	base := basicPipelineReq(pipeline, dataRepo)
	base.Autoscaling = true

	// create three versions of the pipeline differing only in the transform user field
	for i := 1; i <= 3; i++ {
		base.Transform.User = fmt.Sprintf("user:%d", i)
		base.Update = i != 1
		_, err := c.PpsAPIClient.CreatePipeline(c.Ctx(), base)
		require.NoError(t, err)
	}

	for i := 1; i <= 3; i++ {
		info, err := c.InspectPipeline(pfs.DefaultProjectName, fmt.Sprintf("%s^%d", pipeline, 3-i), true)
		require.NoError(t, err)
		require.Equal(t, fmt.Sprintf("user:%d", i), info.Details.Transform.User)
	}

	infos, err := c.ListPipeline()
	require.NoError(t, err)
	require.Equal(t, 1, len(infos))
	require.Equal(t, pipeline, infos[0].Pipeline.Name)
	require.Equal(t, fmt.Sprintf("user:%d", 3), infos[0].Details.Transform.User)

	checkInfos := func(infos []*pps.PipelineInfo) {
		// make sure versions are sorted and have the correct details
		for i, info := range infos {
			require.Equal(t, 3-i, int(info.Version))
			require.Equal(t, fmt.Sprintf("user:%d", 3-i), info.Details.Transform.User)
		}
	}

	// get all pipelines
	infos, err = c.ListPipelineHistory(pfs.DefaultProjectName, pipeline, -1)
	require.NoError(t, err)
	require.Equal(t, 3, len(infos))
	checkInfos(infos)

	// get all pipelines by asking for too many
	infos, err = c.ListPipelineHistory(pfs.DefaultProjectName, pipeline, 3)
	require.NoError(t, err)
	require.Equal(t, 3, len(infos))
	checkInfos(infos)

	// get only the later two pipelines
	infos, err = c.ListPipelineHistory(pfs.DefaultProjectName, pipeline, 1)
	require.NoError(t, err)
	require.Equal(t, 2, len(infos))
	checkInfos(infos)
}

func TestProjectDefaultsCreatedBy(t *testing.T) {
	t.Parallel()
	c := at.EnvWithAuth(t).PachClient

	admin := tu.AuthenticateClient(t, c, auth.RootUser)
	aliceName, alice := tu.RandomRobot(t, c, "alice")
	adminProject := tu.UniqueString("adminProject")
	require.NoError(t, admin.CreateProject(adminProject))
	_, err := admin.PpsAPIClient.SetProjectDefaults(admin.Ctx(), &pps.SetProjectDefaultsRequest{
		Project:             &pfs.Project{Name: adminProject},
		ProjectDefaultsJson: `{"createPipelineRequest": {"autoscaling": true}}`,
	})
	require.NoError(t, err)
	resp, err := admin.PpsAPIClient.GetProjectDefaults(c.Ctx(), &pps.GetProjectDefaultsRequest{Project: &pfs.Project{Name: adminProject}})
	require.NoError(t, err)
	require.Equal(t, "pach:root", resp.CreatedBy)

	aliceProject := tu.UniqueString("aliceProject")
	require.NoError(t, alice.CreateProject(aliceProject))
	_, err = alice.PpsAPIClient.SetProjectDefaults(alice.Ctx(), &pps.SetProjectDefaultsRequest{
		Project:             &pfs.Project{Name: aliceProject},
		ProjectDefaultsJson: `{"createPipelineRequest": {"autoscaling": true}}`,
	})
	require.NoError(t, err)
	resp, err = alice.PpsAPIClient.GetProjectDefaults(c.Ctx(), &pps.GetProjectDefaultsRequest{Project: &pfs.Project{Name: aliceProject}})
	require.NoError(t, err)
	require.Equal(t, aliceName, resp.CreatedBy)

	pi, err := admin.InspectProject(adminProject)
	require.NoError(t, err)
	require.Equal(t, "pach:root", pi.CreatedBy)
	pi, err = alice.InspectProject(aliceProject)
	require.NoError(t, err)
	require.Equal(t, aliceName, pi.CreatedBy)

	pp, err := admin.ListProject()
	require.NoError(t, err)
	for _, p := range pp {
		switch p.Project.Name {
		case "default":
			require.Equal(t, "", p.CreatedBy)
		case adminProject:
			require.Equal(t, "pach:root", p.CreatedBy)
		case aliceProject:
			require.Equal(t, aliceName, p.CreatedBy)
		default:
			t.Fatalf("unexpected project %v", p.Project)
		}
	}
}
