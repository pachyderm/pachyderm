package pjs_test

import (
	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/pjs"
	"github.com/pachyderm/pachyderm/v2/src/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"io"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	clientlog_interceptor "github.com/pachyderm/pachyderm/v2/src/internal/middleware/logging/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachd"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func rootPachClientFromClient(t *testing.T, otherClient *client.APIClient) *client.APIClient {
	ctx := pctx.TestContext(t)
	c, err := client.NewFromPachdAddress(ctx, otherClient.GetAddress(),
		client.WithAdditionalUnaryClientInterceptors(
			clientlog_interceptor.LogUnary,
		),
		client.WithAdditionalStreamClientInterceptors(
			clientlog_interceptor.LogStream,
		),
	)
	require.NoError(t, err)
	c.SetAuthToken("iamroot")
	return c
}

func TestCheckPermissions(t *testing.T) {
	c := pachd.NewTestPachd(t, pachd.ActivateAuthOption(""))
	cfc, err := c.CreateFileset(c.Ctx())
	require.NoError(t, err)
	err = cfc.Send(&storage.CreateFilesetRequest{
		Modification: &storage.CreateFilesetRequest_AppendFile{
			AppendFile: &storage.AppendFile{
				Path: "/foo",
				Data: wrapperspb.Bytes([]byte("bar")),
			},
		},
	})
	require.NoError(t, err)
	createFileSetResp, err := cfc.CloseAndRecv()
	require.NoError(t, err)
	jobResp, err := c.CreateJob(c.Ctx(), &pjs.CreateJobRequest{
		Context: "",
		Program: createFileSetResp.FilesetId,
		Input:   []string{createFileSetResp.FilesetId}},
	)
	require.NoError(t, err)
	expected := jobResp.Job.Id
	t.Run("invalid/projectCreator", func(t *testing.T) {
		c := rootPachClientFromClient(t, c)

		robot, err := c.AuthAPIClient.GetRobotToken(c.Ctx(), &auth.GetRobotTokenRequest{
			Robot: "wall-e",
		})
		require.NoError(t, err)
		c.SetAuthToken(robot.Token)

		wc, err := c.WalkJob(c.Ctx(), &pjs.WalkJobRequest{Job: &pjs.Job{Id: expected}, Algorithm: pjs.WalkAlgorithm_LEVEL_ORDER})
		require.NoError(t, err)
		_, err = wc.Recv()
		require.YesError(t, err)
		s := status.Convert(err)
		require.Equal(t, codes.PermissionDenied, s.Code())
	})
	t.Run("valid/clusterAdmin", func(t *testing.T) {
		c := rootPachClientFromClient(t, c)

		wc, err := c.WalkJob(c.Ctx(), &pjs.WalkJobRequest{Job: &pjs.Job{Id: expected}, Algorithm: pjs.WalkAlgorithm_LEVEL_ORDER})
		require.NoError(t, err)
		resp, err := wc.Recv()
		require.NoError(t, err)
		require.NoDiff(t, expected, resp.Job.Id, nil)
		_, err = wc.Recv()
		require.YesError(t, err)
		require.ErrorIs(t, err, io.EOF)
	})
}
