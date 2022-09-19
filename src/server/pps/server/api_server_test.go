package server

import (
	"context"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

func newClient(t testing.TB) pps.APIClient {
	srv := newServer(t)
	gc := grpcutil.NewTestClient(t, func(gs *grpc.Server) {
		pps.RegisterAPIServer(gs, srv)
	})
	return pps.NewAPIClient(gc)
}

func newServer(t testing.TB) pps.APIServer {
	txnEnv := transactionenv.New()
	db := dockertestenv.NewTestDB(t)
	etcdEnv := testetcd.NewEnv(t)
	env := Env{
		BackgroundContext: context.Background(),
		Logger:            logrus.StandardLogger(),

		DB:         db,
		EtcdClient: etcdEnv.EtcdClient,
		KubeClient: nil,

		AuthServer: nil,
		PFSServer:  nil,

		TxnEnv:        txnEnv,
		GetPachClient: nil,
		Config:        newConfig(t),
	}
	srv, err := NewAPIServerNoMaster(env)
	require.NoError(t, err)
	return srv
}

func newConfig(testing.TB) serviceenv.Configuration {
	return *serviceenv.ConfigFromOptions()
}

func TestRenderTemplate(t *testing.T) {
	ctx := context.Background()
	client := newClient(t)
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
			name:        "useless json",
			line:        "{}",
			wantMessage: "",
		},
		{
			name:        "docker json",
			line:        `{"log":"{\"message\":\"ok\"}"}`,
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
			name:        "empty native json",
			line:        `{"master":false}`,
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
			err := parseLokiLine(test.line, &msg)
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
