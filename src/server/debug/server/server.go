package server

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/internal/clientsdk"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	workerserver "github.com/pachyderm/pachyderm/v2/src/server/worker/server"
	log "github.com/sirupsen/logrus"
	"github.com/wcharczuk/go-chart"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO: Figure out how pipeline versions should come into play with this.
// Right now we just collect the spec and jobs for the current pipeline version.

const (
	defaultDuration = time.Minute
	pachdPrefix     = "pachd"
	pipelinePrefix  = "pipelines"
	podPrefix       = "pods"
)

type debugServer struct {
	env           serviceenv.ServiceEnv
	name          string
	sidecarClient *client.APIClient
	marshaller    *jsonpb.Marshaler
}

// NewDebugServer creates a new server that serves the debug api over GRPC
func NewDebugServer(env serviceenv.ServiceEnv, name string, sidecarClient *client.APIClient) debug.DebugServer {
	return &debugServer{
		env:           env,
		name:          name,
		sidecarClient: sidecarClient,
		marshaller:    &jsonpb.Marshaler{Indent: "  "},
	}
}

type collectPipelineFunc func(*tar.Writer, *pps.PipelineInfo, ...string) error
type collectWorkerFunc func(*tar.Writer, *v1.Pod, ...string) error
type redirectFunc func(debug.DebugClient, *debug.Filter) (io.Reader, error)
type collectFunc func(*tar.Writer, ...string) error

func (s *debugServer) handleRedirect(
	pachClient *client.APIClient,
	server grpcutil.StreamingBytesServer,
	filter *debug.Filter,
	collectPachd collectFunc,
	collectPipeline collectPipelineFunc,
	collectWorker collectWorkerFunc,
	redirect redirectFunc,
	collect collectFunc,
	extraApps ...string,
) error {
	return grpcutil.WithStreamingBytesWriter(server, func(w io.Writer) error {
		return withDebugWriter(w, func(tw *tar.Writer) error {
			// Handle filter.
			pachdContainerPrefix := join(pachdPrefix, s.name, "pachd")
			if filter != nil {
				switch f := filter.Filter.(type) {
				case *debug.Filter_Pachd:
					return collectPachd(tw, pachdContainerPrefix)
				case *debug.Filter_Pipeline:
					pipelineInfo, err := pachClient.InspectPipeline(f.Pipeline.Name, true)
					if err != nil {
						return err
					}
					return s.handlePipelineRedirect(tw, pipelineInfo, collectPipeline, collectWorker, redirect)
				case *debug.Filter_Worker:
					if f.Worker.Redirected {
						// Collect the storage container.
						if s.sidecarClient == nil {
							return collect(tw, client.PPSWorkerSidecarContainerName)
						}
						// Collect the user container.
						if err := collect(tw, client.PPSWorkerUserContainerName); err != nil {
							return err
						}
						// Redirect to the storage container.
						r, err := redirect(s.sidecarClient.DebugClient, filter)
						if err != nil {
							return err
						}
						return collectDebugStream(tw, r)

					}
					pod, err := s.env.GetKubeClient().CoreV1().Pods(s.env.Config().Namespace).Get(f.Worker.Pod, metav1.GetOptions{})
					if err != nil {
						return err
					}
					return s.handleWorkerRedirect(tw, pod, collectWorker, redirect)
				}
			}
			// No filter, collect everything.
			if err := collectPachd(tw, pachdContainerPrefix); err != nil {
				return err
			}
			pipelineInfos, err := pachClient.ListPipeline(true)
			if err != nil {
				return err
			}
			for _, pipelineInfo := range pipelineInfos {
				if err := s.handlePipelineRedirect(tw, pipelineInfo, collectPipeline, collectWorker, redirect); err != nil {
					return err
				}
			}
			if len(extraApps) > 0 {
				return s.appLogs(tw, extraApps)
			}
			return nil
		})
	})
}

func (s *debugServer) appLogs(tw *tar.Writer, apps []string) error {
	pods, err := s.env.GetKubeClient().CoreV1().Pods(s.env.Config().Namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		},
		LabelSelector: metav1.FormatLabelSelector(&metav1.LabelSelector{
			MatchLabels: map[string]string{
				"suite": "pachyderm",
			},
			MatchExpressions: []metav1.LabelSelectorRequirement{{
				Key:      "app",
				Operator: metav1.LabelSelectorOpIn,
				Values:   apps,
			}},
		}),
	})
	if err != nil {
		return err
	}
	for _, pod := range pods.Items {
		if err := s.collectLogs(tw, pod.Name, "", join(pod.Labels["app"], pod.Name)); err != nil {
			return err
		}
	}
	return nil
}

func (s *debugServer) handlePipelineRedirect(
	tw *tar.Writer,
	pipelineInfo *pps.PipelineInfo,
	collectPipeline collectPipelineFunc,
	collectWorker collectWorkerFunc,
	redirect redirectFunc,
) (retErr error) {
	prefix := join(pipelinePrefix, pipelineInfo.Pipeline.Name)
	defer func() {
		if retErr != nil {
			retErr = writeErrorFile(tw, retErr, prefix)
		}
	}()
	if collectPipeline != nil {
		if err := collectPipeline(tw, pipelineInfo, prefix); err != nil {
			return err
		}
	}
	return s.forEachWorker(tw, pipelineInfo, func(pod *v1.Pod) error {
		return s.handleWorkerRedirect(tw, pod, collectWorker, redirect, prefix)
	})
}

func (s *debugServer) forEachWorker(tw *tar.Writer, pipelineInfo *pps.PipelineInfo, cb func(*v1.Pod) error) error {
	pods, err := s.getWorkerPods(pipelineInfo)
	if err != nil {
		return err
	}
	if len(pods) == 0 {
		return errors.Errorf("no worker pods found for pipeline %v", pipelineInfo.Pipeline.Name)
	}
	for _, pod := range pods {
		if err := cb(&pod); err != nil {
			return err
		}
	}
	return nil
}

func (s *debugServer) getWorkerPods(pipelineInfo *pps.PipelineInfo) ([]v1.Pod, error) {
	podList, err := s.env.GetKubeClient().CoreV1().Pods(s.env.Config().Namespace).List(
		metav1.ListOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ListOptions",
				APIVersion: "v1",
			},
			LabelSelector: metav1.FormatLabelSelector(
				metav1.SetAsLabelSelector(
					map[string]string{
						"app": ppsutil.PipelineRcName(pipelineInfo.Pipeline.Name, pipelineInfo.Version),
					},
				),
			),
		},
	)
	if err != nil {
		return nil, err
	}
	return podList.Items, nil
}

func (s *debugServer) handleWorkerRedirect(tw *tar.Writer, pod *v1.Pod, collectWorker collectWorkerFunc, cb redirectFunc, prefix ...string) (retErr error) {
	workerPrefix := join(podPrefix, pod.Name)
	if len(prefix) > 0 {
		workerPrefix = join(prefix[0], workerPrefix)
	}
	defer func() {
		if retErr != nil {
			retErr = writeErrorFile(tw, retErr, workerPrefix)
		}
	}()
	if collectWorker != nil {
		if err := collectWorker(tw, pod, workerPrefix); err != nil {
			return err
		}
	}
	if pod.Status.Phase != v1.PodRunning {
		return errors.Errorf("pod in phase %v, must be in phase %v to collect debug information", pod.Status.Phase, v1.PodRunning)
	}
	c, err := workerserver.NewClient(pod.Status.PodIP)
	if err != nil {
		return err
	}
	defer func() {
		if err := c.Close(); err != nil {
			log.Errorf("errored closing worker client: %v", err)
		}
	}()
	r, err := cb(c.DebugClient, &debug.Filter{
		Filter: &debug.Filter_Worker{
			Worker: &debug.Worker{
				Pod:        pod.Name,
				Redirected: true,
			},
		},
	})
	if err != nil {
		return err
	}
	return collectDebugStream(tw, r, workerPrefix)
}

func (s *debugServer) Profile(request *debug.ProfileRequest, server debug.Debug_ProfileServer) error {
	pachClient := s.env.GetPachClient(server.Context())
	return s.handleRedirect(
		pachClient,
		server,
		request.Filter,
		collectProfileFunc(request.Profile),
		nil,
		nil,
		redirectProfileFunc(pachClient.Ctx(), request.Profile),
		collectProfileFunc(request.Profile),
	)
}

func collectProfileFunc(profile *debug.Profile) collectFunc {
	return func(tw *tar.Writer, prefix ...string) error {
		return collectProfile(tw, profile, prefix...)
	}
}

func collectProfile(tw *tar.Writer, profile *debug.Profile, prefix ...string) error {
	return collectDebugFile(tw, profile.Name, "", func(w io.Writer) error {
		return writeProfile(w, profile)
	}, prefix...)
}

func writeProfile(w io.Writer, profile *debug.Profile) error {
	if profile.Name == "cpu" {
		if err := pprof.StartCPUProfile(w); err != nil {
			return err
		}
		duration := defaultDuration
		if profile.Duration != nil {
			var err error
			duration, err = types.DurationFromProto(profile.Duration)
			if err != nil {
				return err
			}
		}
		time.Sleep(duration)
		pprof.StopCPUProfile()
		return nil
	}
	p := pprof.Lookup(profile.Name)
	if p == nil {
		return errors.Errorf("unable to find profile %q", profile.Name)
	}
	if profile.Name == "goroutine" {
		return p.WriteTo(w, 2)
	}
	return p.WriteTo(w, 0)
}

func redirectProfileFunc(ctx context.Context, profile *debug.Profile) redirectFunc {
	return func(c debug.DebugClient, filter *debug.Filter) (io.Reader, error) {
		profileC, err := c.Profile(ctx, &debug.ProfileRequest{
			Profile: profile,
			Filter:  filter,
		})
		if err != nil {
			return nil, err
		}
		return grpcutil.NewStreamingBytesReader(profileC, nil), nil
	}
}

func (s *debugServer) Binary(request *debug.BinaryRequest, server debug.Debug_BinaryServer) error {
	pachClient := s.env.GetPachClient(server.Context())
	return s.handleRedirect(
		pachClient,
		server,
		request.Filter,
		collectBinary,
		nil,
		nil,
		redirectBinaryFunc(pachClient.Ctx()),
		collectBinary,
	)
}

func collectBinary(tw *tar.Writer, prefix ...string) error {
	return collectDebugFile(tw, "binary", "", func(w io.Writer) (retErr error) {
		f, err := os.Open(os.Args[0])
		if err != nil {
			return err
		}
		defer func() {
			if err := f.Close(); retErr == nil {
				retErr = err
			}
		}()
		_, err = io.Copy(w, f)
		return err
	}, prefix...)
}

func redirectBinaryFunc(ctx context.Context) redirectFunc {
	return func(c debug.DebugClient, filter *debug.Filter) (io.Reader, error) {
		binaryC, err := c.Binary(ctx, &debug.BinaryRequest{Filter: filter})
		if err != nil {
			return nil, err
		}
		return grpcutil.NewStreamingBytesReader(binaryC, nil), nil
	}
}

func (s *debugServer) Dump(request *debug.DumpRequest, server debug.Debug_DumpServer) error {
	if request.Limit == 0 {
		request.Limit = math.MaxInt64
	}
	pachClient := s.env.GetPachClient(server.Context())
	return s.handleRedirect(
		pachClient,
		server,
		request.Filter,
		s.collectPachdDumpFunc(pachClient, request.Limit),
		s.collectPipelineDumpFunc(pachClient, request.Limit),
		s.collectWorkerDump,
		redirectDumpFunc(pachClient.Ctx()),
		collectDump,
		"pg-bouncer", "etcd",
	)
}

func (s *debugServer) collectPachdDumpFunc(pachClient *client.APIClient, limit int64) collectFunc {
	return func(tw *tar.Writer, prefix ...string) error {
		// Collect input repos.
		if err := s.collectInputRepos(tw, pachClient, limit); err != nil {
			return err
		}
		// Collect the pachd version.
		if err := s.collectPachdVersion(tw, pachClient, prefix...); err != nil {
			return err
		}
		// Collect the pachd container logs.
		if err := s.collectLogs(tw, s.name, "pachd", prefix...); err != nil {
			return err
		}
		// Collect the pachd container dump.
		return collectDump(tw, prefix...)
	}
}

func (s *debugServer) collectInputRepos(tw *tar.Writer, pachClient *client.APIClient, limit int64) error {
	repoInfos, err := pachClient.ListRepo()
	if err != nil {
		return err
	}
	for _, repoInfo := range repoInfos {
		if _, err := pachClient.InspectPipeline(repoInfo.Repo.Name, true); err != nil {
			if errutil.IsNotFoundError(err) {
				repoPrefix := join("source-repos", repoInfo.Repo.Name)
				return s.collectCommits(tw, pachClient, repoInfo.Repo, limit, repoPrefix)
			}
			return err
		}
	}
	return nil
}

func (s *debugServer) collectCommits(tw *tar.Writer, pachClient *client.APIClient, repo *pfs.Repo, limit int64, prefix ...string) error {
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
	if err := collectDebugFile(tw, "commits", "json", func(w io.Writer) error {
		ctx, cancel := context.WithCancel(pachClient.Ctx())
		defer cancel()
		client, err := pachClient.PfsAPIClient.ListCommit(ctx, &pfs.ListCommitRequest{
			Repo:   repo,
			Number: limit,
			All:    true,
		})
		if err != nil {
			return err
		}
		return clientsdk.ForEachCommit(client, func(ci *pfs.CommitInfo) error {
			if ci.Finished != nil && ci.Details.CompactingTime != nil && ci.Details.ValidatingTime != nil {
				compactingDuration, err := types.DurationFromProto(ci.Details.CompactingTime)
				if err != nil {
					return err
				}
				compacting.XValues = append(compacting.XValues, float64(len(compacting.XValues)+1))
				compacting.YValues = append(compacting.YValues, float64(compactingDuration))
				validatingDuration, err := types.DurationFromProto(ci.Details.ValidatingTime)
				if err != nil {
					return err
				}
				validating.XValues = append(validating.XValues, float64(len(validating.XValues)+1))
				validating.YValues = append(validating.YValues, float64(validatingDuration))
				finishingTime, err := types.TimestampFromProto(ci.Finishing)
				if err != nil {
					return err
				}
				finishedTime, err := types.TimestampFromProto(ci.Finished)
				if err != nil {
					return err
				}
				finishingDuration := finishedTime.Sub(finishingTime)
				finishing.XValues = append(finishing.XValues, float64(len(finishing.XValues)+1))
				finishing.YValues = append(finishing.YValues, float64(finishingDuration))
			}
			return s.marshaller.Marshal(w, ci)
		})
	}, prefix...); err != nil {
		return err
	}
	// Reverse the x values since we collect them from newest to oldest.
	// TODO: It would probably be better to support listing jobs in reverse order.
	reverseContinuousSeries(compacting, validating, finishing)
	return collectGraph(tw, "commits-chart", "number of commits", []chart.Series{compacting, validating, finishing}, prefix...)
}

func reverseContinuousSeries(series ...chart.ContinuousSeries) {
	for _, s := range series {
		for i := 0; i < len(s.XValues)/2; i++ {
			s.XValues[i], s.XValues[len(s.XValues)-1-i] = s.XValues[len(s.XValues)-1-i], s.XValues[i]
		}
	}
}

func collectGraph(tw *tar.Writer, name, XAxisName string, series []chart.Series, prefix ...string) error {
	return collectDebugFile(tw, name, "", func(w io.Writer) error {
		graph := chart.Chart{
			Title: name,
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
		return graph.Render(chart.PNG, w)
	}, prefix...)
}

func (s *debugServer) collectPachdVersion(tw *tar.Writer, pachClient *client.APIClient, prefix ...string) error {
	return collectDebugFile(tw, "version", "", func(w io.Writer) error {
		version, err := pachClient.Version()
		if err != nil {
			return err
		}
		_, err = io.Copy(w, strings.NewReader(version+"\n"))
		return err
	}, prefix...)
}

func (s *debugServer) collectLogs(tw *tar.Writer, pod, container string, prefix ...string) error {
	if err := collectDebugFile(tw, "logs", "txt", func(w io.Writer) (retErr error) {
		stream, err := s.env.GetKubeClient().CoreV1().Pods(s.env.Config().Namespace).GetLogs(pod, &v1.PodLogOptions{Container: container}).Stream()
		if err != nil {
			return err
		}
		defer func() {
			if err := stream.Close(); retErr == nil {
				retErr = err
			}
		}()
		_, err = io.Copy(w, stream)
		return err
	}, prefix...); err != nil {
		return err
	}
	return collectDebugFile(tw, "logs-previous", "txt", func(w io.Writer) (retErr error) {
		stream, err := s.env.GetKubeClient().CoreV1().Pods(s.env.Config().Namespace).GetLogs(pod, &v1.PodLogOptions{Container: container, Previous: true}).Stream()
		if err != nil {
			return err
		}
		defer func() {
			if err := stream.Close(); retErr == nil {
				retErr = err
			}
		}()
		_, err = io.Copy(w, stream)
		return err
	}, prefix...)
}

func collectDump(tw *tar.Writer, prefix ...string) error {
	if err := collectProfile(tw, &debug.Profile{Name: "goroutine"}, prefix...); err != nil {
		return err
	}
	return collectProfile(tw, &debug.Profile{Name: "heap"}, prefix...)
}

func (s *debugServer) collectPipelineDumpFunc(pachClient *client.APIClient, limit int64) collectPipelineFunc {
	return func(tw *tar.Writer, pipelineInfo *pps.PipelineInfo, prefix ...string) error {
		if err := collectDebugFile(tw, "spec", "json", func(w io.Writer) error {
			fullPipelineInfos, err := pachClient.ListPipelineHistory(pipelineInfo.Pipeline.Name, -1, true)
			if err != nil {
				return err
			}
			for _, fullPipelineInfo := range fullPipelineInfos {
				if err := s.marshaller.Marshal(w, fullPipelineInfo); err != nil {
					return err
				}
			}
			return nil
		}, prefix...); err != nil {
			return err
		}
		if err := s.collectCommits(tw, pachClient, client.NewRepo(pipelineInfo.Pipeline.Name), limit, prefix...); err != nil {
			return err
		}
		return s.collectJobs(tw, pachClient, pipelineInfo.Pipeline.Name, limit, prefix...)
	}
}

func (s *debugServer) collectJobs(tw *tar.Writer, pachClient *client.APIClient, pipelineName string, limit int64, prefix ...string) error {
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
	if err := collectDebugFile(tw, "jobs", "json", func(w io.Writer) error {
		// TODO: The limiting should eventually be a feature of list job.
		var count int64
		return pachClient.ListJobF(pipelineName, nil, -1, false, func(ji *pps.JobInfo) error {
			if count >= limit {
				return errutil.ErrBreak
			}
			count++
			if ji.Stats.DownloadTime != nil && ji.Stats.ProcessTime != nil && ji.Stats.UploadTime != nil {
				downloadDuration, err := types.DurationFromProto(ji.Stats.DownloadTime)
				if err != nil {
					return err
				}
				processDuration, err := types.DurationFromProto(ji.Stats.ProcessTime)
				if err != nil {
					return err
				}
				uploadDuration, err := types.DurationFromProto(ji.Stats.UploadTime)
				if err != nil {
					return err
				}
				download.XValues = append(download.XValues, float64(len(download.XValues)+1))
				download.YValues = append(download.YValues, float64(downloadDuration))
				process.XValues = append(process.XValues, float64(len(process.XValues)+1))
				process.YValues = append(process.YValues, float64(processDuration))
				upload.XValues = append(upload.XValues, float64(len(upload.XValues)+1))
				upload.YValues = append(upload.YValues, float64(uploadDuration))
			}
			return s.marshaller.Marshal(w, ji)
		})
	}, prefix...); err != nil {
		return err
	}
	// Reverse the x values since we collect them from newest to oldest.
	// TODO: It would probably be better to support listing jobs in reverse order.
	reverseContinuousSeries(download, process, upload)
	return collectGraph(tw, "jobs-chart", "number of jobs", []chart.Series{download, process, upload}, prefix...)
}

func (s *debugServer) collectWorkerDump(tw *tar.Writer, pod *v1.Pod, prefix ...string) error {
	// Collect the worker user and storage container logs.
	userPrefix := client.PPSWorkerUserContainerName
	sidecarPrefix := client.PPSWorkerSidecarContainerName
	if len(prefix) > 0 {
		userPrefix = join(prefix[0], userPrefix)
		sidecarPrefix = join(prefix[0], sidecarPrefix)
	}
	if err := s.collectLogs(tw, pod.Name, client.PPSWorkerUserContainerName, userPrefix); err != nil {
		return err
	}
	return s.collectLogs(tw, pod.Name, client.PPSWorkerSidecarContainerName, sidecarPrefix)
}

func redirectDumpFunc(ctx context.Context) redirectFunc {
	return func(c debug.DebugClient, filter *debug.Filter) (io.Reader, error) {
		dumpC, err := c.Dump(ctx, &debug.DumpRequest{Filter: filter})
		if err != nil {
			return nil, err
		}
		return grpcutil.NewStreamingBytesReader(dumpC, nil), nil
	}
}
