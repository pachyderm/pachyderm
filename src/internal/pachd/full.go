package pachd

import (
	"context"
	"io"
	"net/http"
	"os"
	"path"

	"google.golang.org/grpc"

	"github.com/dustin/go-humanize"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	eprsserver "github.com/pachyderm/pachyderm/v2/src/server/enterprise/server"
	"github.com/sirupsen/logrus"
)

// A fullBuilder builds a full-mode pachd.
type fullBuilder struct {
	builder
}

func (fb *fullBuilder) maybeRegisterIdentityServer(ctx context.Context) error {
	if fb.env.Config().EnterpriseMember {
		return nil
	}
	return fb.builder.registerIdentityServer(ctx)
}

// registerEnterpriseServer registers a FULL-mode enterprise server.  This
// differs from enterprise mode in that the mode & unpaused-mode options are
// passed; it differs from sidecar mode in that the mode & unpaused-mode options
// are passed, the heartbeat is enabled and the license environment’s enterprise
// server is set; it differs from paused mode in that the mode option is in full
// mode.
//
// TODO: refactor the four modes to have a cleaner license/enterprise server
// abstraction.
func (fb *fullBuilder) registerEnterpriseServer(ctx context.Context) error {
	fb.enterpriseEnv = eprsserver.EnvFromServiceEnv(
		fb.env,
		path.Join(fb.env.Config().EtcdPrefix, fb.env.Config().EnterpriseEtcdPrefix),
		fb.txnEnv,
		eprsserver.WithMode(eprsserver.FullMode),
		eprsserver.WithUnpausedMode(os.Getenv("UNPAUSED_MODE")),
	)
	apiServer, err := eprsserver.NewEnterpriseServer(
		fb.enterpriseEnv,
		true,
	)
	if err != nil {
		return err
	}
	fb.forGRPCServer(func(s *grpc.Server) {
		enterprise.RegisterAPIServer(s, apiServer)
	})
	fb.bootstrappers = append(fb.bootstrappers, apiServer)
	fb.env.SetEnterpriseServer(apiServer)
	fb.licenseEnv.EnterpriseServer = apiServer
	return nil
}

// newFullBuilder returns a new initialized FullBuilder.
func newFullBuilder(config any) *fullBuilder {
	return &fullBuilder{newBuilder(config, "pachyderm-pachd-full")}
}

// buildAndRun builds and starts a full-mode pachd.
func (fb *fullBuilder) buildAndRun(ctx context.Context) error {
	fb.daemon.criticalServersOnly = fb.env.Config().RequireCriticalServersOnly
	return fb.apply(ctx,
		fb.setupDB,
		fb.maybeInitDexDB,
		fb.maybeInitReporter,
		fb.initInternalServer,
		fb.initExternalServer,
		fb.registerLicenseServer,
		fb.registerEnterpriseServer,
		fb.maybeRegisterIdentityServer,
		fb.registerAuthServer,
		fb.registerPFSServer,
		fb.registerPPSServer,
		fb.registerTransactionServer,
		fb.registerAdminServer,
		fb.registerHealthServer,
		fb.registerVersionServer,
		fb.registerDebugServer,
		fb.registerProxyServer,
		fb.initS3Server,
		fb.initPrometheusServer,

		fb.initTransaction,
		fb.internallyListen,
		fb.bootstrap,
		fb.externallyListen,
		fb.resumeHealth,
		func(ctx context.Context) error {
			logrus.Infof("spawning server")
			go func() {
				logrus.Infof("server starting")
				err := http.ListenAndServe("0.0.0.0:2000", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					logrus.Infof("upload request started")
					defer logrus.Infof("upload request ended")
					var badusage bool
					if req.Method != "PUT" {
						badusage = true
					}
					dir, file := path.Split(req.URL.Path)
					if dir != "/upload/" {
						badusage = true
					}
					if badusage {
						logrus.Error("bad upload usage")
						http.Error(w, "usage: PUT /upload/filename", http.StatusBadRequest)
						return
					}
					c := fb.env.GetPachClient(req.Context())
					c.SetAuthToken("V4Ptmxchcj04TY5vmngJAD0RiuJ3JYc6")
					defer req.Body.Close()
					var err error
					if req.URL.Query().Has("fake") {
						var n int64
						n, err = io.Copy(io.Discard, req.Body)
						logrus.Infof("fake upload: read %v", humanize.IBytes(uint64(n)))
					} else {
						err = c.PutFile(client.NewProjectCommit("default", "benchmark-upload", "master", ""), file, req.Body)
					}
					if err != nil {
						logrus.Error("putfile failed: %v", err)
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}

					w.WriteHeader(http.StatusAccepted)
				}))
				if err != nil {
					logrus.Errorf("could not start http server: %v", err)
				}
			}()
			return nil
		},
		fb.daemon.serve,
	)
}

// FullMode runs a full-mode pachd.
//
// Full mode is that standard pachd which users interact with using pachctl and
// which manages pipelines, files and so forth.
func FullMode(ctx context.Context, config any) error {
	return newFullBuilder(config).buildAndRun(ctx)
}
