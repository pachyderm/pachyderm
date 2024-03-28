package server

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"runtime/coverage"
	runtimedebug "runtime/debug"
	"runtime/pprof"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wcharczuk/go-chart"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	describe "k8s.io/kubectl/pkg/describe"

	"github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	loki "github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/server/debug/server/debugstar"
	workerserver "github.com/pachyderm/pachyderm/v2/src/server/worker/server"
)

const (
	defaultDuration = time.Minute
	pachdPrefix     = "pachd"
	pipelinePrefix  = "pipelines"
	podPrefix       = "pods"
	databasePrefix  = "database"
	defaultsPrefix  = "defaults"
)

type Env struct {
	DB                   *pachsql.DB
	SidecarClient        *client.APIClient
	Name                 string
	GetLokiClient        func() (*loki.Client, error)
	GetKubeClient        func() kubernetes.Interface
	GetDynamicKubeClient func() dynamic.Interface
	Config               pachconfig.Configuration
	TaskService          task.Service

	GetPachClient func(context.Context) *client.APIClient
}

type debugServer struct {
	debug.UnsafeDebugServer

	env                 Env
	name                string
	sidecarClient       *client.APIClient
	marshaller          *protojson.MarshalOptions
	database            *pachsql.DB
	logLevel, grpcLevel log.LevelChanger
}

// NewDebugServer creates a new server that serves the debug api over GRPC
func NewDebugServer(env Env) debug.DebugServer {
	s := &debugServer{
		env:           env,
		name:          env.Name,
		sidecarClient: env.SidecarClient,
		marshaller:    &protojson.MarshalOptions{Indent: "  "},
		database:      env.DB,
		logLevel:      log.LogLevel,
		grpcLevel:     log.GRPCLevel,
	}
	ticker := time.NewTicker(3 * time.Second)
	query := `SELECT current_timestamp - query_start as runtime,
                         datname,
                         usename,
                         client_addr,
                         backend_xid,
                         query
		  FROM pg_stat_activity WHERE state = 'idle' AND current_timestamp - query_start > interval '10 seconds' ORDER BY runtime DESC;`
	ctx := pctx.Background("debug server")
	go func() {
		for {
			select {
			case <-ticker.C:
				rows, err := s.database.QueryContext(ctx, query)
				if err != nil {
					log.Error(ctx, "query idle txns", zap.Error(err))
				}
				var buf bytes.Buffer
				if err := s.writeRowsToJSON(rows, &buf); err != nil {
					log.Error(ctx, "write idle txns query results", zap.Error(err))
				}
				log.Info(ctx, "idle txns", zap.String("txns", buf.String()))
			}
		}
	}()
	return s
}

func (s *debugServer) logDBState() {
	ticker := time.NewTicker(3 * time.Second)
	queries := map[string]string{
		"idle queries": `
                  SELECT current_timestamp - query_start as runtime,
                         datname,
                         usename,
                         client_addr,
                         backend_xid,
                         query
		  FROM pg_stat_activity WHERE state = 'idle' AND current_timestamp - query_start > interval '10 seconds' ORDER BY runtime DESC;`,
		"active queries": `
                  SELECT current_timestamp - query_start as runtime,
                         datname,
                         usename,
                         client_addr,
                         backend_xid,
                         query
		  FROM pg_stat_activity WHERE state != 'idle' AND current_timestamp - query_start > interval '10 seconds' ORDER BY runtime DESC;`,
	}
	ctx := pctx.Background("debug server")
	for {
		select {
		case <-ticker.C:
			for msg, query := range queries {
				rows, err := s.database.QueryContext(ctx, query)
				if err != nil {
					log.Error(ctx, fmt.Sprintf("query %s", msg), zap.Error(err))
				}
				var buf bytes.Buffer
				if err := s.writeRowsToJSON(rows, &buf); err != nil {
					log.Error(ctx, fmt.Sprintf("write results for  %s", msg), zap.Error(err))
				}
				log.Info(ctx, msg, zap.String("txns", buf.String()))
			}
		}
	}

}

// the returned taskPath gets deleted by the caller after its contents are streamed, and is the location of any associated error file
type taskFunc func(ctx context.Context, dfs DumpFS) error

func (s *debugServer) GetDumpV2Template(ctx context.Context, request *debug.GetDumpV2TemplateRequest) (*debug.GetDumpV2TemplateResponse, error) {
	var ps []*debug.Pipeline
	pis, err := s.env.GetPachClient(ctx).ListPipeline()
	if err != nil {
		return nil, errors.Wrap(err, "list pipelines")
	}
	for _, pi := range pis {
		ps = append(ps, &debug.Pipeline{Name: pi.Pipeline.Name, Project: pi.Pipeline.Project.Name})
	}
	var pipelines []*pps.Pipeline
	for _, p := range pis {
		pipelines = append(pipelines, p.GetPipeline())
	}
	apps, possibleApps, err := s.listApps(ctx, pipelines)
	if err != nil {
		return nil, err
	}
	// App could have oneof label selector or list of specifics
	var pachApps []*debug.App
	for _, app := range apps {
		app.Pods = nil // clear pods
		if app.Pipeline != nil || app.Name == "pachd" {
			pachApps = append(pachApps, app)
		}
	}
	var scripts []*debug.Starlark
	for _, name := range maps.Keys(debugstar.BuiltinScripts) {
		scripts = append(scripts, &debug.Starlark{
			Source: &debug.Starlark_Builtin{
				Builtin: name,
			},
			Timeout: durationpb.New(time.Minute),
		})
	}
	return &debug.GetDumpV2TemplateResponse{
		Request: &debug.DumpV2Request{
			System: &debug.System{
				Helm:      true,
				Database:  true,
				Version:   true,
				Describes: apps,
				Logs:      apps,
				LokiLogs:  possibleApps,
				Profiles:  pachApps,
			},
			InputRepos: true,
			Pipelines:  ps,
			Defaults: &debug.DumpV2Request_Defaults{
				ClusterDefaults: true,
			},
			StarlarkScripts: scripts,
		},
	}, nil
}

// list apps returns a list of running apps, and a list of apps which may possibly exist.  The
// intent is to use the first result for things like "kubectl describe" and the second for getting
// loki logs.
func (s *debugServer) listApps(ctx context.Context, pipelines []*pps.Pipeline) (running []*debug.App, possible []*debug.App, retErr error) {
	ctx, end := log.SpanContext(ctx, "listApps")
	defer end(log.Errorp(&retErr))
	ctx, c := context.WithTimeout(ctx, 10*time.Minute)
	defer c()
	pods, err := s.env.GetKubeClient().CoreV1().Pods(s.env.Config.Namespace).List(ctx, metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		},
		LabelSelector: metav1.FormatLabelSelector(&metav1.LabelSelector{
			MatchLabels: map[string]string{
				"suite": "pachyderm",
			},
		}),
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "list apps")
	}

	pipelineMap := make(map[string]*pps.Pipeline)
	for _, p := range pipelines {
		pipelineMap[p.String()] = p
	}

	apps := make(map[string]*debug.App)
	for _, p := range pods.Items {
		var appName string
		// pipelines are represented as their own apps here even though they all share the label app=pipeline
		if p.Labels["app"] == "pipeline" {
			appName = fmt.Sprintf("%s/%s", p.Labels["pipelineProject"], p.Labels["pipelineName"])
			delete(pipelineMap, appName)
		} else {
			appName = p.Labels["app"]
		}
		app, ok := apps[appName]
		if !ok {
			app = &debug.App{Name: appName}
			apps[appName] = app
		}
		var containers []string
		for _, c := range p.Spec.Containers {
			containers = append(containers, c.Name)
		}
		pod := &debug.Pod{Name: p.Name, Ip: p.Status.PodIP, Containers: containers}
		if p.Labels["app"] == "pipeline" {
			app.Pipeline = &debug.Pipeline{Name: p.Labels["pipelineName"], Project: p.Labels["pipelineProject"]}
		}
		app.Pods = append(app.Pods, pod)
	}
	for _, app := range apps {
		sort.Slice(app.Pods, func(i, j int) bool {
			return app.Pods[i].Name < app.Pods[j].Name
		})
		running = append(running, app)
		possible = append(possible, app)
	}
	for name, pipeline := range pipelineMap {
		possible = append(possible, &debug.App{
			Name: name,
			Pipeline: &debug.Pipeline{
				Project: pipeline.GetProject().GetName(),
				Name:    pipeline.GetName(),
			},
		})
	}
	sort.Slice(running, func(i, j int) bool {
		return running[i].GetName() < running[j].GetName()
	})
	sort.Slice(possible, func(i, j int) bool {
		return possible[i].GetName() < possible[j].GetName()
	})
	return running, possible, nil
}

func (s *debugServer) fillApps(ctx context.Context, reqApps []*debug.App) error {
	if len(reqApps) == 0 {
		return nil
	}
	apps, _, err := s.listApps(ctx, nil)
	if err != nil {
		return err
	}
	appMap := make(map[string]*debug.App)
	for _, a := range apps {
		appMap[a.Name] = a
	}
	for _, reqApp := range reqApps {
		if app, ok := appMap[reqApp.Name]; ok {
			if len(reqApp.Pods) == 0 {
				reqApp.Pods = app.Pods
			} else {
				appPods := make(map[string]*debug.Pod)
				for _, p := range app.Pods {
					appPods[p.Name] = p
				}
				for _, reqPod := range reqApp.Pods {
					if pod, ok := appPods[reqPod.Name]; ok {
						reqPod.Ip = pod.Ip
					} else {
						return errors.Errorf("Requested pod %q of app %q not found", reqPod.Name, reqApp.Name)
					}
				}
			}
		}
	}
	return nil
}

func (s *debugServer) DumpV2(request *debug.DumpV2Request, server debug.Debug_DumpV2Server) error {
	if request == nil {
		return errors.New("nil debug.DumpV2Request")
	}
	var apps []*debug.App
	for _, l := range [][]*debug.App{
		request.GetSystem().GetBinaries(),
		request.GetSystem().GetDescribes(),
		request.GetSystem().GetLogs(),
		request.GetSystem().GetLokiLogs(),
		request.GetSystem().GetProfiles(),
	} {
		if len(l) > 0 {
			apps = append(apps, l...)
		}
	}
	// fill the pod info for all the pods referenced in request.System Apps
	if err := s.fillApps(server.Context(), apps); err != nil {
		return err
	}
	c := s.env.GetPachClient(server.Context())
	tasks := s.makeTasks(server.Context(), request, server)
	var d time.Duration
	if request.Timeout != nil {
		d = request.Timeout.AsDuration()
	}
	return s.dump(c, server, tasks, d)
}

func (s *debugServer) makeTasks(ctx context.Context, request *debug.DumpV2Request, server debug.Debug_DumpV2Server) []taskFunc {
	var ts []taskFunc
	if request.InputRepos {
		ts = append(ts, func(ctx context.Context, dfs DumpFS) error {
			return s.collectInputRepos(ctx, dfs.WithPrefix("source-repos"), s.env.GetPachClient(ctx), int64(0), server)
		})
	}
	if len(request.Pipelines) > 0 {
		ts = append(ts, s.makePipelinesTask(server, request.Pipelines))
	}
	if sys := request.System; sys != nil {
		if sys.Helm {
			ts = append(ts, func(ctx context.Context, dfs DumpFS) error {
				return s.collectHelm(ctx, dfs, server)
			})
		}
		if sys.Database {
			ts = append(ts, func(ctx context.Context, dfs DumpFS) error {
				return s.collectDatabaseDump(ctx, dfs.WithPrefix(databasePrefix), server)
			})
		}
		if sys.Version {
			// this section needs some work to work for more than just this pachd
			rp := recordProgress(server, "version", 1)
			ts = append(ts, func(ctx context.Context, dfs DumpFS) error {
				return s.collectPachdVersion(ctx, dfs.WithPrefix(fmt.Sprintf("/pachd/%s/pachd", s.name)), s.env.GetPachClient(ctx), rp)
			})
		}
		if len(sys.Describes) > 0 {
			rp := recordProgress(server, "describes", len(sys.Logs))
			for _, app := range sys.Describes {
				func(app *debug.App) {
					ts = append(ts, s.makeDescribesTask(app, rp))
				}(app)
			}
		}
		if len(sys.Logs) > 0 {
			rp := recordProgress(server, "logs", len(sys.Logs))
			for _, app := range sys.Logs {
				func(app *debug.App) {
					ts = append(ts, s.makeLogsTask(app, rp))
				}(app)
			}
		}
		if len(sys.LokiLogs) > 0 {
			rp := recordProgress(server, "loki-logs", len(sys.LokiLogs))
			for _, app := range sys.LokiLogs {
				func(app *debug.App) {
					ts = append(ts, s.makeLokiTask(app, rp))
				}(app)
			}
		}
		if len(sys.Profiles) > 0 {
			ts = append(ts, s.makeProfilesTask(server, sys.Profiles))
		}
		if len(sys.Binaries) > 0 {
			ts = append(ts, s.makeBinariesTask(server, sys.Binaries))
		}
	}
	if request.Defaults.GetClusterDefaults() {
		rp := recordProgress(server, "defaults", 1)
		ts = append(ts, func(ctx context.Context, dfs DumpFS) error {
			return s.collectDefaults(ctx, dfs.WithPrefix(defaultsPrefix), server, rp)
		})
	}
	if sls := request.GetStarlarkScripts(); len(sls) > 0 {
		rp := recordProgress(server, "starlark", len(sls))
		for _, sl := range sls {
			ts = append(ts, s.makeStarlarkScriptTask(sl, rp))
		}
	}
	return ts
}

func (s *debugServer) makePipelinesTask(server debug.Debug_DumpV2Server, pipelines []*debug.Pipeline) taskFunc {
	return func(ctx context.Context, dfs DumpFS) error {
		rp := recordProgress(server, "pipelines", len(pipelines))
		var errs error
		for _, p := range pipelines {
			pip := &pps.Pipeline{Name: p.Name, Project: &pfs.Project{Name: p.Project}}
			if err := s.collectPipeline(ctx, dfs.WithPrefix(pipelinePrefix), pip, 0); err != nil {
				errors.JoinInto(&errs, errors.Wrapf(err, "collect pipeline %q", pip))
			}
			rp(ctx)
		}
		return errs
	}
}

func (s *debugServer) makeDescribesTask(app *debug.App, rp incProgressFunc) taskFunc {
	return func(ctx context.Context, dfs DumpFS) error {
		defer rp(ctx)
		var errs error
		if app.Timeout != nil {
			var cf context.CancelFunc
			ctx, cf = context.WithTimeout(ctx, app.Timeout.AsDuration())
			defer cf()
		}
		for _, pod := range app.Pods {
			if err := s.collectDescribe(ctx, dfs, app, pod); err != nil {
				errors.JoinInto(&errs, errors.Wrapf(err, "describe pod %q", pod.Name))
			}
		}
		return errs
	}
}

func (s *debugServer) makeLogsTask(app *debug.App, rp incProgressFunc) taskFunc {
	return func(ctx context.Context, dfs DumpFS) error {
		defer rp(ctx)
		if len(app.Pods) == 0 {
			return nil
		}
		if app.Timeout != nil {
			var cf context.CancelFunc
			ctx, cf = context.WithTimeout(ctx, app.Timeout.AsDuration())
			defer cf()
		}
		var errs error
		for _, pod := range app.Pods {
			if err := s.kubeLogs(ctx, dfs, app, pod); err != nil {
				errors.JoinInto(&errs, errors.Wrapf(err, "collectLogs(%s)", pod.Name))
			}
		}
		return errs
	}
}

func (s *debugServer) makeLokiTask(app *debug.App, rp incProgressFunc) taskFunc {
	return func(ctx context.Context, dfs DumpFS) (retErr error) {
		if !s.hasLoki() {
			return nil
		}
		if app.Timeout != nil {
			var cf context.CancelFunc
			ctx, cf = context.WithTimeout(ctx, app.Timeout.AsDuration())
			defer cf()
		}
		ctx, end := log.SpanContext(ctx, "collectLokiLogs")
		defer end(log.Errorp(&retErr))
		defer rp(ctx)
		var errs error
		ctx, cf := context.WithTimeout(ctx, 10*time.Minute)
		defer cf()
		if app.Pipeline != nil {
			if err := s.forEachWorkerLoki(ctx, client.NewPipeline(app.Pipeline.Project, app.Pipeline.Name), func(pod string) (retErr error) {
				ctx, end := log.SpanContext(ctx, "forEachWorkerLoki.worker", zap.String("pod", pod))
				defer end(log.Errorp(&retErr))
				if err := s.collectLogsLoki(ctx, dfs, app, pod, client.PPSWorkerUserContainerName); err != nil {
					return errors.Wrapf(err, "collect user container loki logs for pod %q", pod)
				}
				if err := s.collectLogsLoki(ctx, dfs, app, pod, client.PPSWorkerSidecarContainerName); err != nil {
					return errors.Wrapf(err, "collect storage container loki logs for pod %q", pod)
				}
				return nil
			}); err != nil {
				return err
			}
		} else {
			for _, pod := range app.Pods {
				for _, c := range pod.Containers {
					if err := s.collectLogsLoki(ctx, dfs, app, pod.Name, c); err != nil {
						return errors.Wrapf(err, "collect pachd container loki logs for pod %q", pod.Name)
					}
				}
			}
			return nil
		}
		return errs
	}
}

func (s *debugServer) makeProfilesTask(server debug.Debug_DumpV2Server, apps []*debug.App) taskFunc {
	return func(ctx context.Context, dfs DumpFS) error {
		var errs error
		rp := recordProgress(server, "profiles", len(apps))
		for _, app := range apps {
			func() {
				if app.Timeout != nil {
					var cf context.CancelFunc
					ctx, cf = context.WithTimeout(ctx, app.Timeout.AsDuration())
					defer cf()
				}
				defer rp(ctx)
				for _, pod := range app.Pods {
					for _, c := range pod.Containers {
						for _, profile := range []string{"heap", "goroutine"} {
							if err := s.collectProfile(ctx, dfs, app, pod, c, profile); err != nil {
								errors.JoinInto(&errs, err)
							}
						}
					}
				}
			}()
		}
		log.Error(ctx, "profile errors", zap.Error(errs))
		return errs

	}
}

func (s *debugServer) makeBinariesTask(server debug.Debug_DumpV2Server, apps []*debug.App) taskFunc {
	return func(ctx context.Context, dfs DumpFS) error {
		var errs error
		rp := recordProgress(server, "binaries", len(apps))
		for _, app := range apps {
			func() {
				if app.Timeout != nil {
					var cf context.CancelFunc
					ctx, cf = context.WithTimeout(ctx, app.Timeout.AsDuration())
					defer cf()
				}
				defer rp(ctx)
				for _, pod := range app.Pods {
					if pod.Ip == "" {
						log.Info(ctx, "skipping debug binary for pod with empty IP", zap.String("pod", pod.Name))
						continue
					}
					if err := dfs.Write(filepath.Join(appDir(app), "pods", pod.Name, "binary"), func(w io.Writer) (retErr error) {
						if pod.Name == s.name {
							f, err := os.Open(os.Args[0])
							if err != nil {
								return errors.Wrap(err, "open file")
							}
							defer func() {
								if err := f.Close(); err != nil {
									errors.JoinInto(&retErr, err)
								}
							}()
							_, err = io.Copy(w, f)
							return errors.Wrap(err, "write local binary")
						}
						c, err := workerserver.NewClient(pod.Ip)
						if err != nil {
							return errors.Wrapf(err, "create worker client for IP %q", pod.Ip)
						}
						binaryC, err := c.Binary(ctx, &debug.BinaryRequest{})
						if err != nil {
							return errors.Wrap(err, "binary client")
						}
						r := grpcutil.NewStreamingBytesReader(binaryC, nil)
						defer func() {
							if err := r.Close(); err != nil {
								errors.JoinInto(&retErr, errors.Wrapf(err, "close binary writer for pod %q", pod.Name))
							}
						}()
						_, err = io.Copy(w, r)
						return errors.Wrap(err, "write binary")
					}); err != nil {
						errors.JoinInto(&errs, errors.Wrapf(err, "close binary writer for pod %q", pod.Name))
					}
				}
			}()
		}
		return errs
	}
}

func appDir(app *debug.App) string {
	if app.Pipeline != nil {
		return filepath.Join("pipelines", app.Pipeline.Project, app.Pipeline.Name)
	}
	return app.Name
}

type dumpContentServer struct {
	server debug.Debug_DumpV2Server
}

func (dcs *dumpContentServer) Send(bytesValue *wrapperspb.BytesValue) error {
	return dcs.server.Send(&debug.DumpChunk{Chunk: &debug.DumpChunk_Content{Content: &debug.DumpContent{Content: bytesValue.Value}}})
}

type incProgressFunc func(context.Context)

// TODO should this be sending in a goro?
func recordProgress(server debug.Debug_DumpV2Server, task string, total int) incProgressFunc {
	if server == nil {
		return func(ctx context.Context) {}
	}
	inc := 0
	return func(ctx context.Context) {
		inc++
		if err := server.Send(
			&debug.DumpChunk{
				Chunk: &debug.DumpChunk_Progress{
					Progress: &debug.DumpProgress{
						Task:     task,
						Progress: int64(inc),
						Total:    int64(total),
					},
				},
			},
		); err != nil {
			log.Error(ctx, fmt.Sprintf("error sending progress for task %q", task), zap.Error(err))
		}
	}
}

// TODO: sharpen timeouts
func (s *debugServer) dump(c *client.APIClient, server debug.Debug_DumpV2Server, tasks []taskFunc, timeout time.Duration) (retErr error) {
	ctx := c.Ctx() // this context has authorization credentials we need
	dumpRoot := filepath.Join(os.TempDir(), uuid.NewWithoutDashes())
	if err := os.Mkdir(dumpRoot, 0744); err != nil {
		return errors.Wrap(err, "create dump root directory")
	}
	defer func() {
		errors.JoinInto(&retErr, os.RemoveAll(dumpRoot))
	}()
	to := 25 * time.Minute
	if timeout != 0 {
		to = timeout
	}
	if deadline, ok := server.Context().Deadline(); ok {
		if d := time.Until(deadline); d < to {
			to = d
		}
	}
	var cf context.CancelFunc
	// Time out our own operations at ~90% of the client-set deadline. (22 minutes of
	// server-side processing for every 25 minutes of debug dumping.)
	ctx, cf = context.WithTimeout(ctx, time.Duration(22)*to/time.Duration(25))
	defer cf()
	eg, ctx := errgroup.WithContext(ctx)
	var mu sync.Mutex // to synchronize writes to the tar stream
	dumpContent := &dumpContentServer{server: server}
	return grpcutil.WithStreamingBytesWriter(dumpContent, func(w io.Writer) error {
		return withDebugWriter(w, func(tw *tar.Writer) error {
			for _, f := range tasks {
				func(f taskFunc) {
					eg.Go(func() (retErr error) {
						taskCtx := ctx
						defer func() {
							if retErr != nil {
								log.Error(taskCtx, fmt.Sprintf("failed dump task %q", "name"), zap.Error(retErr))
								retErr = nil
							}
						}()
						taskDir := filepath.Join(dumpRoot, uuid.NewWithoutDashes()) + string(filepath.Separator)
						if err := os.Mkdir(taskDir, 0744); err != nil {
							return errors.Wrap(err, "create dump task sub-directory")
						}
						defer func() {
							if err := os.RemoveAll(taskDir); err != nil {
								errors.JoinInto(&retErr, err)
							}
						}()
						dfs := NewDumpFS(taskDir)
						if err := f(ctx, dfs); err != nil {
							if writeErr := writeErrorFile(dfs, err, ""); writeErr != nil {
								log.Error(taskCtx, "write error", zap.Error(writeErr))
							}
						}
						mu.Lock()
						defer mu.Unlock()
						err := filepath.Walk(taskDir, func(path string, fi fs.FileInfo, err error) error {
							if err != nil {
								return errors.Wrapf(err, "walk path %q", path)
							}
							if fi.IsDir() {
								return nil
							}
							dest := strings.TrimPrefix(path, taskDir)
							return errors.Wrapf(writeTarFile(tw, dest, path, fi), "write tar file %q", dest)
						})
						return errors.Wrapf(err, "walk task directory %q", taskDir)
					})
				}(f)
			}
			return errors.Wrap(eg.Wait(), "run dump tasks")
		})
	})
}

func writeTar(ctx context.Context, w io.Writer, f func(ctx context.Context, dfs DumpFS) error) (retErr error) {
	dumpRoot := filepath.Join(os.TempDir(), uuid.NewWithoutDashes()) + string(filepath.Separator)
	if err := os.Mkdir(dumpRoot, 0744); err != nil {
		return errors.Wrap(err, "create dump root directory")
	}
	defer func() {
		errors.JoinInto(&retErr, os.RemoveAll(dumpRoot))
	}()
	tw := tar.NewWriter(w)
	defer errors.Close(&retErr, tw, "close tar writer")
	if err := f(ctx, NewDumpFS(dumpRoot)); err != nil {
		return errors.Wrap(err, "write dfs to tar")
	}
	err := filepath.Walk(dumpRoot, func(path string, fi fs.FileInfo, err error) error {
		if err != nil {
			return errors.Wrapf(err, "walk path %q", path)
		}
		if fi.IsDir() {
			return nil
		}
		dest := strings.TrimPrefix(path, dumpRoot)
		return errors.Wrapf(writeTarFile(tw, dest, path, fi), "write tar file %q", dest)
	})
	return errors.Wrapf(err, "walk dump directory %q", dumpRoot)
}

func writeTarGz(ctx context.Context, w io.Writer, f func(ctx context.Context, dfs DumpFS) error) (retErr error) {
	gw := gzip.NewWriter(w)
	defer errors.Close(&retErr, gw, "close gzip writer")
	return writeTar(ctx, w, f)
}

func (s *debugServer) kubeLogs(ctx context.Context, dfs DumpFS, app *debug.App, pod *debug.Pod) (retErr error) {
	ctx, end := log.SpanContext(ctx, "collectKubeLogs")
	defer end(log.Errorp(&retErr))
	ctx, c := context.WithTimeout(ctx, 10*time.Minute)
	defer c()
	var errs error
	for _, c := range pod.Containers {
		if err := s.collectLogs(ctx, dfs, app, pod, c); err != nil {
			errors.JoinInto(&errs, errors.Wrapf(err, "collectLogsLoki(%s.%s)", pod.Name, c))
		}
	}
	return errs
}

func (s *debugServer) Profile(request *debug.ProfileRequest, server debug.Debug_ProfileServer) error {
	return grpcutil.WithStreamingBytesWriter(server, func(w io.Writer) error {
		return errors.Wrap(writeTarGz(server.Context(), w, func(ctx context.Context, dfs DumpFS) error {
			return writeProfile(ctx, dfs, request.Profile)
		}), "collect profile")
	})
}

func (s *debugServer) collectProfile(ctx context.Context, dfs DumpFS, app *debug.App, pod *debug.Pod, container, profile string) error {
	if err := dfs.Write(filepath.Join(appDir(app), "pods", pod.Name, container, profile), func(w io.Writer) (retErr error) {
		req := &debug.ProfileRequest{Profile: &debug.Profile{Name: profile}}
		if pod.Name == s.name {
			if err := dfs.Write(filepath.Join(appDir(app), "pods", pod.Name, container, "go_info.txt"), func(w io.Writer) error {
				fmt.Fprintf(w, "build info: ")
				info, ok := runtimedebug.ReadBuildInfo()
				if ok {
					fmt.Fprintf(w, "%s", info.String())
				} else {
					fmt.Fprint(w, "<no build info>")
				}
				fmt.Fprintf(w, "GOOS: %v\nGOARCH: %v\nGOMAXPROCS: %v\nNumCPU: %v\n", runtime.GOOS, runtime.GOARCH, runtime.GOMAXPROCS(0), runtime.NumCPU())
				return nil
			}); err != nil {
				return errors.Wrap(err, "go info")
			}
			return errors.Wrap(writeProfile(ctx, dfs.WithPrefix(filepath.Join(appDir(app), "pods", pod.Name, container)), req.Profile), "write profile")
		}
		c, err := workerserver.NewClient(pod.Ip)
		if err != nil {
			return errors.Wrapf(err, "worker client for %q at %q", pod.Name, pod.Ip)
		}
		prof, err := c.Profile(ctx, req)
		if err != nil {
			return errors.Wrapf(err, "collect profile %q for %q at %q", profile, pod.Name, pod.Ip)
		}
		r := grpcutil.NewStreamingBytesReader(prof, nil)
		defer func() {
			if err := r.Close(); err != nil {
				errors.JoinInto(&retErr, errors.Wrapf(err, "close profile writer for pod %q; profile %q", pod.Name, profile))
			}
		}()
		_, err = io.Copy(w, r)
		return errors.Wrap(err, "write profile")
	}); err != nil {
		return errors.Wrapf(err, "collect profile %q", profile)
	}
	return nil
}

func writeProfile(ctx context.Context, dfs DumpFS, profile *debug.Profile) (retErr error) {
	defer log.Span(ctx, "writeProfile", zap.String("profile", profile.GetName()))(log.Errorp(&retErr))
	switch profile.GetName() {
	case "cover":
		var errs error
		// "go tool covdata" relies on these files being named the same as what is
		// written out automatically. See go/src/runtime/coverage/emit.go.
		// (covcounters.<16 byte hex-encoded hash that nothing checks>.<pid>.<unix
		// nanos>, covmeta.<same hash>.)
		hash := [16]byte{}
		if _, err := rand.Read(hash[:]); err != nil {
			return errors.Wrap(err, "generate random coverage id")
		}
		if err := dfs.Write(fmt.Sprintf("cover/covcounters.%x.%d.%d", hash, os.Getpid(), time.Now().UnixNano()), func(w io.Writer) error {
			return errors.Wrap(coverage.WriteCounters(w), "counters")
		}); err != nil {
			errors.JoinInto(&errs, err)
		}
		if err := dfs.Write(fmt.Sprintf("cover/covmeta.%x", hash), func(w io.Writer) error {
			return errors.Wrap(coverage.WriteMeta(w), "meta")
		}); err != nil {
			errors.JoinInto(&errs, err)
		}
		if errs == nil {
			log.Debug(ctx, "clearing coverage counters")
			if err := coverage.ClearCounters(); err != nil {
				log.Debug(ctx, "problem clearing coverage counters", zap.Error(err))
			}
		}
		if errs == nil {
			log.Debug(ctx, "clearing coverage counters")
			if err := coverage.ClearCounters(); err != nil {
				log.Debug(ctx, "problem clearing coverage counters", zap.Error(err))
			}
		}
		return errs
	case "cpu":
		return dfs.Write(profile.Name, func(w io.Writer) error {
			if err := pprof.StartCPUProfile(w); err != nil {
				return errors.EnsureStack(err)
			}
			defer pprof.StopCPUProfile()
			duration := defaultDuration
			if profile.Duration != nil {
				duration = profile.Duration.AsDuration()
			}
			// Wait for either the defined duration, or until the context is
			// done.
			t := time.NewTimer(duration)
			select {
			case <-ctx.Done():
				t.Stop()
				return errors.EnsureStack(context.Cause(ctx))
			case <-t.C:
				return nil
			}
		})
	default:
		return dfs.Write(profile.Name, func(w io.Writer) error {
			p := pprof.Lookup(profile.Name)
			if p == nil {
				return errors.Errorf("unable to find profile %q", profile.Name)
			}
			return errors.EnsureStack(p.WriteTo(w, 0))
		})
	}
}

func (s *debugServer) Binary(request *debug.BinaryRequest, server debug.Debug_BinaryServer) (retErr error) {
	return grpcutil.WithStreamingBytesWriter(server, func(w io.Writer) error {
		defer log.Span(server.Context(), "collectBinary", zap.String("binary", os.Args[0]))(log.Errorp(&retErr))
		f, err := os.Open(os.Args[0])
		if err != nil {
			return errors.Wrap(err, "open file")
		}
		defer func() {
			if err := f.Close(); retErr == nil {
				retErr = err
			}
		}()
		_, err = io.Copy(w, f)
		return errors.Wrap(err, "collect binary")
	})
}

func (s *debugServer) Dump(request *debug.DumpRequest, server debug.Debug_DumpServer) error {
	return status.Error(codes.Unimplemented, "Dump gRPC is deprecated in favor of DumpV2")
}

func (s *debugServer) collectInputRepos(ctx context.Context, dfs DumpFS, pachClient *client.APIClient, limit int64, dumpServer debug.Debug_DumpV2Server) (retErr error) {
	repoInfos, err := pachClient.ListRepo()
	if err != nil {
		return errors.Wrap(err, "list repo")
	}
	rp := recordProgress(dumpServer, "source-repos", len(repoInfos))
	var errs error
	for i, repoInfo := range repoInfos {
		func() {
			defer rp(ctx)
			if err := validateRepoInfo(repoInfo); err != nil {
				errors.JoinInto(&errs, errors.Wrapf(err, "invalid repo info %d (%s) from ListRepo", i, repoInfo.String()))
				return
			}
			if _, err := pachClient.InspectPipeline(repoInfo.Repo.Project.Name, repoInfo.Repo.Name, true); err != nil {
				if errutil.IsNotFoundError(err) {
					if err := s.collectCommits(ctx, dfs, pachClient, repoInfo.Repo, limit); err != nil {
						errors.JoinInto(&errs, errors.Wrapf(err, "collectCommits(%s)", repoInfo.GetRepo()))
					}
					return
				}
				errors.JoinInto(&errs, errors.Wrapf(err, "inspectPipeline(%s)", repoInfo.GetRepo()))
			}
		}()
	}
	return errs
}

// TODO(acohen4): there's a bug in here when there's only one commit
func (s *debugServer) collectCommits(rctx context.Context, dfs DumpFS, pachClient *client.APIClient, repo *pfs.Repo, limit int64) error {
	compacting := chart.ContinuousSeries{
		Name: "compacting",
		Style: chart.Style{
			Show:        true,
			StrokeColor: chart.GetDefaultColor(0).WithAlpha(255),
		},
	}
	validating := chart.ContinuousSeries{
		Name: "validating",
		Style: chart.Style{
			Show:        true,
			StrokeColor: chart.GetDefaultColor(1).WithAlpha(255),
		},
	}
	finishing := chart.ContinuousSeries{
		Name: "finishing",
		Style: chart.Style{
			Show:        true,
			StrokeColor: chart.GetDefaultColor(2).WithAlpha(255),
		},
	}
	if err := dfs.Write(filepath.Join(repo.Project.Name, repo.Name, "commits.json"), func(w io.Writer) error {
		ctx, cancel := context.WithCancel(rctx)
		defer cancel()
		client, err := pachClient.PfsAPIClient.ListCommit(ctx, &pfs.ListCommitRequest{
			Repo:   repo,
			Number: limit,
			All:    true,
		})
		if err != nil {
			return errors.EnsureStack(err)
		}
		return grpcutil.ForEach[*pfs.CommitInfo](client, func(ci *pfs.CommitInfo) error {
			if ci.Finished != nil && ci.Details != nil && ci.Details.CompactingTime != nil && ci.Details.ValidatingTime != nil {
				compactingDuration := ci.Details.CompactingTime.AsDuration()
				compacting.XValues = append(compacting.XValues, float64(len(compacting.XValues)+1))
				compacting.YValues = append(compacting.YValues, float64(compactingDuration))
				validatingDuration := ci.Details.ValidatingTime.AsDuration()
				validating.XValues = append(validating.XValues, float64(len(validating.XValues)+1))
				validating.YValues = append(validating.YValues, float64(validatingDuration))
				finishingTime := ci.Finishing.AsTime()
				finishedTime := ci.Finished.AsTime()
				finishingDuration := finishedTime.Sub(finishingTime)
				finishing.XValues = append(finishing.XValues, float64(len(finishing.XValues)+1))
				finishing.YValues = append(finishing.YValues, float64(finishingDuration))
			}
			b, err := s.marshaller.Marshal(ci)
			if err != nil {
				return errors.Wrap(err, "marshal")
			}
			if _, err := w.Write(b); err != nil {
				return errors.Wrap(err, "write")
			}
			return nil
		})
	}); err != nil {
		return err
	}
	// Reverse the x values since we collect them from newest to oldest.
	// TODO: It would probably be better to support listing jobs in reverse order.
	reverseContinuousSeries(compacting, validating, finishing)
	graphPath := filepath.Join(repo.Project.Name, repo.Name, "commits-chart.png")
	if err := collectGraph(dfs, graphPath, "number of commits", []chart.Series{compacting, validating, finishing}); err != nil {
		if writeErr := writeErrorFile(dfs, err, filepath.Join(repo.Project.Name, repo.Name, "commits-chart")); writeErr != nil {
			errors.JoinInto(&err, writeErr)
			return err
		}
	}
	return nil
}

func reverseContinuousSeries(series ...chart.ContinuousSeries) {
	for _, s := range series {
		for i := 0; i < len(s.XValues)/2; i++ {
			s.XValues[i], s.XValues[len(s.XValues)-1-i] = s.XValues[len(s.XValues)-1-i], s.XValues[i]
		}
	}
}

func collectGraph(dfs DumpFS, path, XAxisName string, series []chart.Series) error {
	return dfs.Write(path, func(w io.Writer) error {
		graph := chart.Chart{
			Title: filepath.Base(path),
			TitleStyle: chart.Style{
				Show: true,
			},
			XAxis: chart.XAxis{
				Name: XAxisName,
				NameStyle: chart.Style{
					Show: true,
				},
				Style: chart.Style{
					Show: true,
				},
				ValueFormatter: func(v interface{}) string {
					return fmt.Sprintf("%.0f", v.(float64))
				},
			},
			YAxis: chart.YAxis{
				Name: "time (seconds)",
				NameStyle: chart.Style{
					Show: true,
				},
				Style: chart.Style{
					Show:                true,
					TextHorizontalAlign: chart.TextHorizontalAlignLeft,
				},
				ValueFormatter: func(v interface{}) string {
					seconds := v.(float64) / float64(time.Second)
					return fmt.Sprintf("%.3f", seconds)
				},
			},
			Series: series,
		}
		graph.Elements = []chart.Renderable{
			chart.Legend(&graph),
		}
		return errors.EnsureStack(graph.Render(chart.PNG, w))
	})
}

func (s *debugServer) collectPachdVersion(ctx context.Context, dfs DumpFS, pachClient *client.APIClient, rp incProgressFunc) error {
	defer rp(ctx)
	return dfs.Write("version.txt", func(w io.Writer) error {
		version, err := pachClient.Version()
		if err != nil {
			return err
		}
		_, err = io.Copy(w, strings.NewReader(version+"\n"))
		return errors.EnsureStack(err)
	})
}

func (s *debugServer) collectHelm(ctx context.Context, dfs DumpFS, server debug.Debug_DumpV2Server) error {
	secrets, err := s.env.GetKubeClient().CoreV1().Secrets(s.env.Config.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "owner=helm",
	})
	if err != nil {
		return errors.Wrap(err, "list helm secrets")
	}
	rp := recordProgress(server, "helm", len(secrets.Items))
	var writeErrs error
	for _, secret := range secrets.Items {
		func() {
			defer rp(ctx)
			if err := handleHelmSecret(ctx, dfs.WithPrefix(filepath.Join("helm", secret.Name)), secret); err != nil {
				errors.JoinInto(&writeErrs, errors.Wrapf(err, "%v: write error.txt", secret.Name))
			}
		}()
	}
	return writeErrs
}

func (s *debugServer) collectDescribe(ctx context.Context, dfs DumpFS, app *debug.App, pod *debug.Pod) error {
	return dfs.Write(filepath.Join(appDir(app), "pods", pod.Name, "describe.txt"), func(output io.Writer) (retErr error) {
		// Gate the total time of "describe".
		ctx, c := context.WithTimeout(ctx, 2*time.Minute)
		defer c()
		defer log.Span(ctx, "collectDescribe")(log.Errorp(&retErr))
		r, w := io.Pipe()
		go func() {
			defer log.Span(ctx, "collectDescribe.backgroundDescribe")()
			// Do the "describe" in the background, because the k8s client doesn't take
			// a context here, and we want the ability to abandon the request.  We leak
			// memory if this runs forever but we return, but that's better than a debug
			// dump that doesn't return.
			pd := describe.PodDescriber{
				Interface: s.env.GetKubeClient(),
			}
			output, err := pd.Describe(s.env.Config.Namespace, pod.Name, describe.DescriberSettings{ShowEvents: true})
			if err != nil {
				w.CloseWithError(errors.EnsureStack(err))
				return
			}
			_, err = w.Write([]byte(output))
			w.CloseWithError(errors.EnsureStack(err))
		}()
		go func() {
			// Close the pipe when the context times out; bounding the time on the
			// io.Copy operation below.
			<-ctx.Done()
			w.CloseWithError(errors.EnsureStack(context.Cause(ctx)))
		}()
		if _, err := io.Copy(output, r); err != nil {
			return errors.EnsureStack(err)
		}
		return nil
	})
}

func (s *debugServer) collectLogs(ctx context.Context, dfs DumpFS, app *debug.App, pod *debug.Pod, container string) (retErr error) {
	dir := filepath.Join(appDir(app), "pods", pod.Name, container)
	if err := dfs.Write(filepath.Join(dir, "logs.txt"), func(w io.Writer) (retErr error) {
		defer log.Span(ctx, "collectLogs", zap.String("pod", pod.Name), zap.String("container", container))(log.Errorp(&retErr))
		stream, err := s.env.GetKubeClient().CoreV1().Pods(s.env.Config.Namespace).GetLogs(pod.Name, &v1.PodLogOptions{Container: container}).Stream(ctx)
		if err != nil {
			return errors.EnsureStack(err)
		}
		defer func() {
			if err := stream.Close(); err != nil {
				errors.JoinInto(&retErr, err)
			}
		}()
		_, err = io.Copy(w, stream)
		return errors.EnsureStack(err)
	}); err != nil {
		return err
	}
	if err := dfs.Write(filepath.Join(dir, "logs-previous.txt"), func(w io.Writer) (retErr error) {
		defer log.Span(ctx, "collectLogs.previous", zap.String("pod", pod.Name), zap.String("container", container))(log.Errorp(&retErr))
		stream, err := s.env.GetKubeClient().CoreV1().Pods(s.env.Config.Namespace).GetLogs(pod.Name, &v1.PodLogOptions{Container: container, Previous: true}).Stream(ctx)
		if err != nil {
			return errors.EnsureStack(err)
		}
		defer func() {
			if err := stream.Close(); err != nil {
				errors.JoinInto(&retErr, err)
			}
		}()
		_, err = io.Copy(w, stream)
		return errors.EnsureStack(err)
	}); err != nil {
		if writeErr := writeErrorFile(dfs, err, filepath.Join(dir, "logs-previous")); writeErr != nil {
			errors.JoinInto(&err, writeErr)
			return err
		}
	}
	return nil
}

func (s *debugServer) hasLoki() bool {
	_, err := s.env.GetLokiClient()
	return err == nil
}

func (s *debugServer) collectLogsLoki(ctx context.Context, dfs DumpFS, app *debug.App, pod, container string) error {
	if !s.hasLoki() {
		return nil
	}
	return dfs.Write(filepath.Join(appDir(app), "pods", pod, container, "logs-loki.txt"), func(w io.Writer) (retErr error) {
		defer log.Span(ctx, "collectLogsLoki", zap.String("pod", pod), zap.String("container", container))(log.Errorp(&retErr))
		queryStr := `{pod="` + pod
		if container != "" {
			queryStr += `", container="` + container
		}
		queryStr += `"}`
		logs, err := s.queryLoki(ctx, queryStr)
		if err != nil {
			return errors.EnsureStack(err)
		}
		var cursor *loki.LabelSet
		for _, entry := range logs {
			// Print the stream labels in %v format whenever they are different from the
			// previous line. The pointer comparison is a fast path to avoid
			// reflect.DeepEqual when both log lines are from the same chunk of logs
			// returned by Loki.
			if cursor != entry.Labels && !reflect.DeepEqual(cursor, entry.Labels) {
				if _, err := fmt.Fprintf(w, "%v\n", entry.Labels); err != nil {
					return errors.EnsureStack(err)
				}
			}
			cursor = entry.Labels
			// Then the line itself.
			if _, err := fmt.Fprintf(w, "%s\n", entry.Entry.Line); err != nil {
				return errors.EnsureStack(err)
			}
		}
		return nil
	})
}

func (s *debugServer) collectPipeline(ctx context.Context, dfs DumpFS, p *pps.Pipeline, limit int64) (retErr error) {
	ctx, end := log.SpanContext(ctx, "collectPipelineDump", zap.Stringer("pipeline", p))
	defer end(log.Errorp(&retErr))
	var errs error
	if err := dfs.Write(filepath.Join(p.Project.Name, p.Name, "spec.json"), func(w io.Writer) error {
		fullPipelineInfos, err := s.env.GetPachClient(ctx).ListPipelineHistory(p.Project.Name, p.Name, -1)
		if err != nil {
			return err
		}
		var pipelineErrs error
		for _, fullPipelineInfo := range fullPipelineInfos {
			bs, err := s.marshaller.Marshal(fullPipelineInfo)
			if err != nil {
				errors.JoinInto(&pipelineErrs, errors.Wrapf(err, "marshalFullPipelineInfo(%s)", fullPipelineInfo.GetPipeline()))
			}
			if _, err := w.Write(bs); err != nil {
				errors.JoinInto(&pipelineErrs, errors.Wrapf(err, "marshalFullPipelineInfo(%s).Write", fullPipelineInfo.GetPipeline()))
			}
		}
		return nil
	}); err != nil {
		errors.JoinInto(&errs, errors.Wrap(err, "listProjectPipelineHistory"))
	}
	if err := s.collectCommits(ctx, dfs, s.env.GetPachClient(ctx), client.NewRepo(p.Project.GetName(), p.Name), limit); err != nil {
		errors.JoinInto(&errs, errors.Wrap(err, "collectCommits"))
	}
	if err := s.collectJobs(dfs, s.env.GetPachClient(ctx), p, limit); err != nil {
		errors.JoinInto(&errs, errors.Wrap(err, "collectJobs"))
	}
	return errs
}

func (s *debugServer) forEachWorkerLoki(ctx context.Context, p *pps.Pipeline, cb func(string) error) error {
	pods, err := s.getWorkerPodsLoki(ctx, p)
	if err != nil {
		return err
	}
	var errs error
	for pod := range pods {
		if err := cb(pod); err != nil {
			errors.JoinInto(&errs, errors.Wrapf(err, "forEachWorkersLoki(%s)", pod))
		}
	}
	return errs
}

// quoteLogQLStreamSelector returns a string quoted as a LogQL stream selector.
// The rules for quoting a LogQL stream selector are documented at
// https://grafana.com/docs/loki/latest/logql/log_queries/#log-stream-selector
// to be the same as for a Prometheus string literal, as documented at
// https://prometheus.io/docs/prometheus/latest/querying/basics/#string-literals.
// This happens to be the same as a Go string, with single or double quotes or
// backticks allowed as enclosing characters.
func quoteLogQLStreamSelector(s string) string {
	return strconv.Quote(s)
}

func (s *debugServer) getWorkerPodsLoki(ctx context.Context, p *pps.Pipeline) (map[string]struct{}, error) {
	// This function uses the log querying API and not the label querying API, to bound the the
	// number of workers for a pipeline that we return.  We'll get 30,000 of the most recent
	// logs for each pipeline, and return the names of the workers that contributed to those
	// logs for further inspection.  The alternative would be to get every worker that existed
	// in some time interval, but that results in too much data to inspect.

	queryStr := fmt.Sprintf(`{pipelineProject=%s, pipelineName=%s}`, quoteLogQLStreamSelector(p.Project.Name), quoteLogQLStreamSelector(p.Name))
	pods := make(map[string]struct{})
	logs, err := s.queryLoki(ctx, queryStr)
	if err != nil {
		return nil, err
	}

	for _, l := range logs {
		if l.Labels == nil {
			continue
		}
		labels := *l.Labels
		pod, ok := labels["pod"]
		if !ok {
			log.Debug(ctx, "pod label missing from loki labelset", zap.Any("log", l))
			continue
		}
		pods[pod] = struct{}{}
	}
	return pods, nil
}

const (
	// maxLogs used to be 5,000 but a comment said this was too few logs, so now it's 30,000.
	// If it still seems too small, bump it up again.
	maxLogs = 30000
	// 5,000 is the maximum value of "limit" in queries that Loki seems to accept.
	// We set serverMaxLogs below the actual limit because of a bug in Loki where it hangs when you request too much data from it.
	serverMaxLogs = 1000
)

type lokiLog struct {
	Labels *loki.LabelSet
	Entry  *loki.Entry
}

func (s *debugServer) queryLoki(ctx context.Context, queryStr string) (retResult []lokiLog, retErr error) {
	ctx, finishSpan := log.SpanContext(ctx, "queryLoki", zap.String("queryStr", queryStr))
	defer func() {
		finishSpan(zap.Error(retErr), zap.Int("logs", len(retResult)))
	}()

	sortLogs := func(logs []lokiLog) {
		sort.Slice(logs, func(i, j int) bool {
			return logs[i].Entry.Timestamp.Before(logs[j].Entry.Timestamp)
		})
	}

	c, err := s.env.GetLokiClient()
	if err != nil {
		return nil, errors.EnsureStack(errors.Errorf("get loki client: %w", err))
	}
	// We used to just stream the output, but that results in logs from different streams
	// (stdout/stderr) being randomly interspersed with each other, which is hard to follow when
	// written in text form.  We also used "FORWARD" and basically got the first 30,000 logs
	// this month, instead of the most recent 30,000 logs.  To use "BACKWARD" to get chunks of
	// logs starting with the most recent (whose start time we can't know without asking), we
	// have to either output logs in reverse order, or collect them all and then reverse.  Thus,
	// we just buffer.  30k logs * (200 bytes each + 16 bytes of pointers per line + 24 bytes of
	// time.Time) = 6.8MiB.  That seems totally reasonable to keep in memory, with the benefit
	// of providing lines in chronological order even when the app uses both stdout and stderr.
	var result []lokiLog

	var end time.Time
	start := time.Now().Add(-30 * 24 * time.Hour) // 30 days.  (Loki maximum range is 30 days + 1 hour.)

	for numLogs := 0; (end.IsZero() || start.Before(end)) && numLogs < maxLogs; {
		// Loki requests can hang if the size of the log lines is too big
		ctx, cancel := context.WithTimeout(ctx, time.Minute)
		resp, err := c.QueryRange(ctx, queryStr, serverMaxLogs, start, end, "BACKWARD", 0 /* step */, 0 /* interval */, true /* quiet */)
		cancel()
		if err != nil {
			// Note: the error from QueryRange has a stack.
			if errors.Is(err, context.DeadlineExceeded) {
				log.Debug(ctx, "loki query range timed out", zap.Int("serverMaxLogs", serverMaxLogs), zap.Time("start", start), zap.Time("end", end), zap.Int("logs", len(result)))
				sortLogs(result)
				return result, nil
			}
			return nil, errors.Errorf("query range (query=%v, limit=%v, start=%v, end=%v): %+v", queryStr, serverMaxLogs, start.Format(time.RFC3339), end.Format(time.RFC3339), err)
		}

		streams, ok := resp.Data.Result.(loki.Streams)
		if !ok {
			return nil, errors.Errorf("resp.Data.Result must be of type loki.Streams")
		}

		var readThisIteration int
		for _, s := range streams {
			stream := s // Alias for pointer later.
			for _, e := range stream.Entries {
				entry := e
				numLogs++
				readThisIteration++
				if end.IsZero() || entry.Timestamp.Before(end) {
					// Because end is an "exclusive" range, if we read any logs
					// at all we are guaranteed to not read them in the next
					// iteration.  (If all 5000 logs are at the same time, we
					// still advance past them on the next iteration, which is
					// the best we can do under those circumstances.)
					end = entry.Timestamp
				}
				result = append(result, lokiLog{
					Labels: &stream.Labels,
					Entry:  &entry,
				})
			}
		}
		if readThisIteration == 0 {
			break
		}
		cancel()
	}
	sortLogs(result)
	return result, nil
}

func (s *debugServer) collectJobs(dfs DumpFS, pachClient *client.APIClient, pipeline *pps.Pipeline, limit int64) error {
	download := chart.ContinuousSeries{
		Name: "download",
		Style: chart.Style{
			Show:        true,
			StrokeColor: chart.GetDefaultColor(0).WithAlpha(255),
		},
	}
	process := chart.ContinuousSeries{
		Name: "process",
		Style: chart.Style{
			Show:        true,
			StrokeColor: chart.GetDefaultColor(1).WithAlpha(255),
		},
	}
	upload := chart.ContinuousSeries{
		Name: "upload",
		Style: chart.Style{
			Show:        true,
			StrokeColor: chart.GetDefaultColor(2).WithAlpha(255),
		},
	}
	if err := dfs.Write(filepath.Join(pipeline.Project.Name, pipeline.Name, "jobs.json"), func(w io.Writer) error {
		var count int64
		return pachClient.ListJobF(pipeline.Project.Name, pipeline.Name, nil, -1, false, func(ji *pps.JobInfo) error {
			if limit != 0 && count >= limit {
				return errutil.ErrBreak
			}
			count++
			if ji.Stats.DownloadTime != nil && ji.Stats.ProcessTime != nil && ji.Stats.UploadTime != nil {
				downloadDuration := ji.Stats.DownloadTime.AsDuration()
				processDuration := ji.Stats.ProcessTime.AsDuration()
				uploadDuration := ji.Stats.UploadTime.AsDuration()
				download.XValues = append(download.XValues, float64(len(download.XValues)+1))
				download.YValues = append(download.YValues, float64(downloadDuration))
				process.XValues = append(process.XValues, float64(len(process.XValues)+1))
				process.YValues = append(process.YValues, float64(processDuration))
				upload.XValues = append(upload.XValues, float64(len(upload.XValues)+1))
				upload.YValues = append(upload.YValues, float64(uploadDuration))
			}
			b, err := s.marshaller.Marshal(ji)
			if err != nil {
				return errors.Wrap(err, "marshal")
			}
			if _, err := w.Write(b); err != nil {
				return errors.Wrap(err, "write")
			}
			return nil
		})
	}); err != nil {
		return err
	}
	// Reverse the x values since we collect them from newest to oldest.
	// TODO: It would probably be better to support listing jobs in reverse order.
	reverseContinuousSeries(download, process, upload)
	path := filepath.Join(pipeline.Project.Name, pipeline.Name, "jobs-chart.png")
	if err := collectGraph(dfs, path, "number of jobs", []chart.Series{download, process, upload}); err != nil {
		if writeErr := writeErrorFile(dfs, err, filepath.Join(pipeline.Project.Name, pipeline.Name, "jobs-chart")); writeErr != nil {
			errors.JoinInto(&err, writeErr)
			return err
		}
	}
	return nil
}

func handleHelmSecret(ctx context.Context, dfs DumpFS, secret v1.Secret) error {
	var errs error
	// path := filepath.Join("helm", secret.Name)
	name := secret.Name
	if got, want := string(secret.Type), "helm.sh/release.v1"; got != want {
		err := errors.Errorf("helm-owned secret of unknown version; got %v want %v", got, want)
		if err := writeErrorFile(dfs, err, ""); err != nil {
			errors.JoinInto(&errs, errors.Wrapf(err, "%v: collect secret", secret.Name))
		}
		return errs
	}
	if secret.Data == nil {
		log.Info(ctx, "skipping helm secret with no data", zap.String("secretName", name))
		return nil
	}
	releaseData, ok := secret.Data["release"]
	if !ok {
		log.Info(ctx, "secret data doesn't have a release key", zap.String("secretName", name), zap.Strings("keys", maps.Keys(secret.Data)))
		return nil
	}
	// There are a few labels that helm adds that are interesting (the "release
	// status").  Grab those and write to metadata.json.
	if err := dfs.Write("metadata.json", func(w io.Writer) error {
		text, err := json.Marshal(secret.ObjectMeta)
		if err != nil {
			return errors.Wrap(err, "marshal ObjectMeta")
		}
		if _, err := w.Write(text); err != nil {
			return errors.Wrap(err, "write ObjectMeta json")
		}
		return nil
	}); err != nil {
		// We can still try to get the release JSON if the metadata doesn't marshal
		// or write cleanly.
		if err := writeErrorFile(dfs, err, ""); err != nil {
			errors.JoinInto(&errs, errors.Wrapf(err, "%v: collect metadata", secret.Name))
		}
	}
	// Get the text of the release and write it to release.json.
	var releaseJSON []byte
	if err := dfs.Write("release.json", func(w io.Writer) error {
		// The helm release data is base64-encoded gzipped JSON-marshalled protobuf.
		// The base64 encoding is IN ADDITION to the base64 encoding that k8s does
		// when serving secrets through the API; client-go removes that for us.
		var err error
		releaseJSON, err = base64.StdEncoding.DecodeString(string(releaseData))
		if err != nil {
			releaseJSON = nil
			return errors.Wrap(err, "decode base64")
		}
		// Older versions of helm do not compress the data; this check is for the
		// gzip header; if deteced, decompress.
		if len(releaseJSON) > 3 && bytes.Equal(releaseJSON[0:3], []byte{0x1f, 0x8b, 0x08}) {
			gz, err := gzip.NewReader(bytes.NewReader(releaseJSON))
			if err != nil {
				return errors.Wrap(err, "create gzip reader")
			}
			releaseJSON, err = io.ReadAll(gz)
			if err != nil {
				return errors.Wrap(err, "decompress")
			}
			if err := gz.Close(); err != nil {
				return errors.Wrap(err, "close gzip reader")
			}
		}

		// Write out the raw release JSON, so fields we don't pick out in the next
		// phase can be parsed out by the person analyzing the dump if necessary.
		if _, err := w.Write(releaseJSON); err != nil {
			return errors.Wrap(err, "write release")
		}
		return nil
	}); err != nil {
		if err := writeErrorFile(dfs, err, "release"); err != nil {
			errors.JoinInto(&errs, errors.Wrapf(err, "%v: collect release", secret.Name))
		}
		// The next steps need the release JSON, so we have to give up if any of
		// this failed.  Technically if the write fails, we could continue, but it
		// doesn't seem worth the effort because the next writes are also likely to
		// fail.
		return errs
	}

	// Unmarshal the release JSON, and start writing files for each interesting part.
	var release struct {
		// Config is the merged values.yaml that aren't in the chart.
		Config map[string]any `json:"config"`
		// Manifest is the rendered manifest that was applied to the cluster.
		Manifest string `json:"manifest"`
	}
	if err := json.Unmarshal(releaseJSON, &release); err != nil {
		if err := writeErrorFile(dfs, err, ""); err != nil {
			errors.JoinInto(&errs, errors.Wrapf(err, "%v: unmarshal release json", secret.Name))
		}
		return errs
	}
	// Write the manifest YAML.
	if err := dfs.Write("manifest.yaml", func(w io.Writer) error {
		// Helm adds a newline at the end of the manifest.
		if _, err := fmt.Fprintf(w, "%s", release.Manifest); err != nil {
			return errors.Wrap(err, "print manifest")
		}
		return nil
	}); err != nil {
		if err := writeErrorFile(dfs, err, "manifest"); err != nil {
			errors.JoinInto(&errs, errors.Wrapf(err, "%v: write manifest.yaml", secret.Name))
		}
		// We can try the next step if this fails.
	}
	// Write values.yaml.
	if err := dfs.Write("values.yaml", func(w io.Writer) error {
		b, err := yaml.Marshal(release.Config)
		if err != nil {
			return errors.Wrap(err, "marshal values to yaml")
		}
		if _, err := w.Write(b); err != nil {
			return errors.Wrap(err, "print values")
		}
		return nil
	}); err != nil {
		if err := writeErrorFile(dfs, err, "values"); err != nil {
			errors.JoinInto(&errs, errors.Wrapf(err, "%v: write values.yaml", secret.Name))
		}
	}
	return errs
}

func (s *debugServer) collectDefaults(ctx context.Context, dfs DumpFS, server debug.Debug_DumpV2Server, rp incProgressFunc) error {
	defer rp(ctx)
	resp, err := s.env.GetPachClient(ctx).PpsAPIClient.GetClusterDefaults(ctx, &pps.GetClusterDefaultsRequest{})
	if err != nil {
		return errors.Wrap(err, "get defaults")
	}
	if resp.ClusterDefaultsJson == "" {
		return nil
	}
	if err := dfs.Write("cluster-defaults.json", func(w io.Writer) error {
		var buf bytes.Buffer
		if err := json.Indent(&buf, []byte(resp.ClusterDefaultsJson), "", "  "); err != nil {
			return errors.Wrapf(err, "indent JSON %s", resp.ClusterDefaultsJson)
		}
		if _, err := io.Copy(w, &buf); err != nil {
			return errors.Wrap(err, "write indented JSON")
		}
		return err
	}); err != nil {
		return errors.Wrap(err, "write cluster-defaults.json")
	}
	return nil
}

func (s *debugServer) makeStarlarkScriptTask(st *debug.Starlark, rp incProgressFunc) taskFunc {
	var name, script string
	switch st.GetSource().(type) {
	case *debug.Starlark_Builtin:
		name = st.GetBuiltin()
		var ok bool
		script, ok = debugstar.BuiltinScripts[name]
		if !ok {
			return func(ctx context.Context, dfs DumpFS) error {
				return errors.Errorf("no builtin script %q", name)
			}
		}
	case *debug.Starlark_Literal:
		name = st.GetLiteral().GetName()
		script = st.GetLiteral().GetProgramText()
	}
	return func(rctx context.Context, dfs DumpFS) error {
		defer rp(rctx) // report progress even if timeout is exceeded

		timeout := time.Minute
		if t := st.GetTimeout().AsDuration(); t > 0 {
			timeout = t
		}
		ctx, c := context.WithTimeout(rctx, timeout)
		defer c()

		env := debugstar.Env{
			FS:                  dfs.WithPrefix(name),
			Kubernetes:          s.env.GetKubeClient(),
			KubernetesDynamic:   s.env.GetDynamicKubeClient(),
			KubernetesNamespace: s.env.Config.Namespace,
		}
		if err := env.RunStarlark(ctx, name, script); err != nil {
			return errors.Wrapf(err, "starlark script %q", name)
		}
		return nil
	}
}
