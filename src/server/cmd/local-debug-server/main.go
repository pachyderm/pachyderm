package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/pachyderm/pachyderm/v2/src/admin"
	"github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	loggingmw "github.com/pachyderm/pachyderm/v2/src/internal/middleware/logging"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	adminserver "github.com/pachyderm/pachyderm/v2/src/server/admin/server"
	debugserver "github.com/pachyderm/pachyderm/v2/src/server/debug/server"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func main() {
	log.InitPachctlLogger()
	log.SetLevel(log.DebugLevel)

	ctx, c := signal.NotifyContext(pctx.Background(""), os.Interrupt, syscall.SIGTERM)
	defer c()

	loggingInterceptor := loggingmw.NewLoggingInterceptor(ctx)
	gs, err := grpcutil.NewServer(
		ctx,
		false,
		grpc.ChainUnaryInterceptor(
			loggingInterceptor.UnaryServerInterceptor,
		),
		grpc.ChainStreamInterceptor(
			loggingInterceptor.StreamServerInterceptor,
		),
	)
	if err != nil {
		log.Exit(ctx, "problem creating grpc server", zap.Error(err))
	}

	env := &serviceenv.TestServiceEnv{
		Configuration: &pachconfig.Configuration{
			GlobalConfiguration:        &pachconfig.GlobalConfiguration{},
			PachdSpecificConfiguration: &pachconfig.PachdSpecificConfiguration{},
		},
		KubeClient: fake.NewSimpleClientset(
			&v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pachyderm-proxy-1",
					Labels: map[string]string{
						"suite": "pachyderm",
						"app":   "pachyderm-proxy",
					},
				},
				Status: v1.PodStatus{
					Phase: v1.PodRunning,
					PodIP: "127.0.0.1",
				},
			},
		),
	}
	as := adminserver.NewAPIServer(adminserver.EnvFromServiceEnv(env))
	ds := debugserver.NewDebugServer(env, "debug", nil, nil)
	debug.RegisterDebugServer(gs.Server, ds)
	admin.RegisterAPIServer(gs.Server, as)

	if _, err := gs.ListenTCP("", 1650); err != nil {
		log.Exit(ctx, "problem setting up TCP listener", zap.Error(err))
	}

	log.Info(ctx, "debug server listening on :1650")

	if err := gs.Wait(); err != nil {
		log.Exit(ctx, "problem waiting on server", zap.Error(err))
	}
}
