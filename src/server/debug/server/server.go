package server

import (
	"archive/tar"
	"context"
	"io"
	"io/ioutil"
	"os"
	"path"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/debug"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	workerserver "github.com/pachyderm/pachyderm/src/server/worker/server"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube "k8s.io/client-go/kubernetes"
)

const (
	defaultDuration = time.Minute
)

// NewDebugServer creates a new server that serves the debug api over GRPC
// TODO: Move over to functional options before this gets too unwieldy.
func NewDebugServer(name string, kubeClient *kube.Clientset, namespace string, sidecarClient *client.APIClient) debug.DebugServer {
	return &debugServer{
		name:          name,
		kubeClient:    kubeClient,
		namespace:     namespace,
		sidecarClient: sidecarClient,
	}
}

type debugServer struct {
	name          string
	kubeClient    *kube.Clientset
	namespace     string
	sidecarClient *client.APIClient
}

func (s *debugServer) Profile(request *debug.ProfileRequest, server debug.Debug_ProfileServer) error {
	return withTarWriter(grpcutil.NewStreamingBytesWriter(server), func(tw *tar.Writer) error {
		// Handle filter.
		if request.Filter != nil {
			switch filter := request.Filter.Filter.(type) {
			case *debug.Filter_Pachd:
				return collectProfile(tw, path.Join(s.name, "pachd"), request.Profile)
			case *debug.Filter_Pipeline:
				return s.collectPipeline(filter.Pipeline, func(pod *v1.Pod) error {
					return profileRedirect(server.Context(), tw, request.Profile, pod)
				})
			case *debug.Filter_Worker:
				if filter.Worker.Redirected {
					// Collect the storage container profile.
					if s.sidecarClient == nil {
						return collectProfile(tw, client.PPSWorkerSidecarContainerName, request.Profile)
					}
					// Collect the user container profile.
					if err := collectProfile(tw, client.PPSWorkerUserContainerName, request.Profile); err != nil {
						return err
					}
					// Collect the profile from the storage container.
					profileC, err := s.sidecarClient.DebugClient.Profile(server.Context(), request)
					if err != nil {
						return err
					}
					return copyTar(tw, tar.NewReader(grpcutil.NewStreamingBytesReader(profileC, nil)))
				}
				pod, err := s.kubeClient.CoreV1().Pods(s.namespace).Get(filter.Worker.Pod, metav1.GetOptions{})
				if err != nil {
					return err
				}
				return profileRedirect(server.Context(), tw, request.Profile, pod)
			}
		}
		// No filter, collect everything.
		if err := collectProfile(tw, path.Join(s.name, "pachd"), request.Profile); err != nil {
			return err
		}
		return s.collectPipeline(&debug.Pipeline{Name: ""}, func(pod *v1.Pod) error {
			return profileRedirect(server.Context(), tw, request.Profile, pod)
		})
	})
}

func (s *debugServer) collectPipeline(pipeline *debug.Pipeline, cb func(pod *v1.Pod) error) error {
	pods, err := s.getWorkerPods(pipeline)
	if err != nil {
		return err
	}
	if len(pods) == 0 && pipeline.Name != "" {
		return errors.Errorf("no worker pods found for pipeline %v", pipeline.Name)
	}
	for _, pod := range pods {
		if err := cb(&pod); err != nil {
			return err
		}
	}
	return nil
}

func (s *debugServer) getWorkerPods(pipeline *debug.Pipeline) ([]v1.Pod, error) {
	podList, err := s.kubeClient.CoreV1().Pods(s.namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		},
		LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(getLabel(pipeline))),
	})
	if err != nil {
		return nil, err
	}
	return podList.Items, nil
}

func getLabel(pipeline *debug.Pipeline) map[string]string {
	if pipeline.Name == "" {
		return map[string]string{"component": "worker"}
	}
	return map[string]string{"app": pipeline.Name}
}

func profileRedirect(ctx context.Context, tw *tar.Writer, profile *debug.Profile, pod *v1.Pod) (retErr error) {
	defer func() {
		if retErr != nil {
			retErr = writeErrorFile(tw, pod.Name, retErr)
		}
	}()
	c, err := workerserver.NewClient(pod.Status.PodIP)
	if err != nil {
		return err
	}
	request := &debug.ProfileRequest{
		Profile: profile,
		Filter: &debug.Filter{
			Filter: &debug.Filter_Worker{
				Worker: &debug.Worker{
					Pod:        pod.Name,
					Redirected: true,
				},
			},
		},
	}
	profileC, err := c.Profile(ctx, request)
	if err != nil {
		return err
	}
	tr := tar.NewReader(grpcutil.NewStreamingBytesReader(profileC, nil))
	return copyTar(tw, tr, func(name string) string {
		return strings.TrimPrefix(path.Join(pod.Name, name), "/")
	})
}

func withTarWriter(w io.Writer, cb func(*tar.Writer) error) (retErr error) {
	tw := tar.NewWriter(w)
	defer func() {
		if retErr == nil {
			retErr = tw.Close()
		}
	}()
	return cb(tw)
}

func collectProfile(tw *tar.Writer, container string, profile *debug.Profile) (retErr error) {
	defer func() {
		if retErr != nil {
			retErr = writeErrorFile(tw, path.Join(container, profile.Name), retErr)
		}
	}()
	return withTmpFile(func(f *os.File) error {
		if err := writeProfile(f, profile); err != nil {
			return err
		}
		return writeTarFile(tw, path.Join(container, profile.Name), f)
	})
}

func writeErrorFile(tw *tar.Writer, file string, err error) error {
	return withTmpFile(func(f *os.File) error {
		if _, err := io.Copy(f, strings.NewReader(err.Error())); err != nil {
			return err
		}
		return writeTarFile(tw, file+"-error", f)
	})
}

func withTmpFile(cb func(*os.File) error) (retErr error) {
	if err := os.MkdirAll(os.TempDir(), 0700); err != nil {
		return err
	}
	f, err := ioutil.TempFile(os.TempDir(), "pachyderm_debug")
	if err != nil {
		return err
	}
	defer func() {
		if err := os.Remove(f.Name()); retErr == nil {
			retErr = err
		}
		if err := f.Close(); retErr == nil {
			retErr = err
		}
	}()
	return cb(f)
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
	return p.WriteTo(w, 2)
}

func writeTarFile(tw *tar.Writer, name string, f *os.File) error {
	fi, err := os.Stat(f.Name())
	if err != nil {
		return err
	}
	hdr := &tar.Header{
		Name: strings.TrimPrefix(name, "/"),
		Size: fi.Size(),
		Mode: 0777,
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		return err
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	_, err = io.Copy(tw, f)
	return err
}

func copyTar(tw *tar.Writer, tr *tar.Reader, transform ...func(string) string) error {
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if len(transform) > 0 {
			hdr.Name = transform[0](hdr.Name)
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		_, err = io.Copy(tw, tr)
		if err != nil {
			return err
		}
	}
}

func (s *debugServer) Binary(request *debug.BinaryRequest, server debug.Debug_BinaryServer) error {
	return withTarWriter(grpcutil.NewStreamingBytesWriter(server), func(tw *tar.Writer) error {
		// Handle filter.
		if request.Filter != nil {
			switch filter := request.Filter.Filter.(type) {
			case *debug.Filter_Pachd:
				return collectBinary(tw, path.Join(s.name, "pachd"))
			case *debug.Filter_Pipeline:
				return s.collectPipeline(filter.Pipeline, func(pod *v1.Pod) error {
					return binaryRedirect(server.Context(), tw, pod)
				})
			case *debug.Filter_Worker:
				if filter.Worker.Redirected {
					// Collect the storage container binary.
					if s.sidecarClient == nil {
						return collectBinary(tw, client.PPSWorkerSidecarContainerName)
					}
					// Collect the user container binary.
					if err := collectBinary(tw, client.PPSWorkerUserContainerName); err != nil {
						return err
					}
					// Collect the binary from the storage container.
					binaryC, err := s.sidecarClient.DebugClient.Binary(server.Context(), request)
					if err != nil {
						return err
					}
					return copyTar(tw, tar.NewReader(grpcutil.NewStreamingBytesReader(binaryC, nil)))
				}
				pod, err := s.kubeClient.CoreV1().Pods(s.namespace).Get(filter.Worker.Pod, metav1.GetOptions{})
				if err != nil {
					return err
				}
				return binaryRedirect(server.Context(), tw, pod)
			}
		}
		// No filter, collect everything.
		if err := collectBinary(tw, path.Join(s.name, "pachd")); err != nil {
			return err
		}
		return s.collectPipeline(&debug.Pipeline{Name: ""}, func(pod *v1.Pod) error {
			return binaryRedirect(server.Context(), tw, pod)
		})
	})
}

func binaryRedirect(ctx context.Context, tw *tar.Writer, pod *v1.Pod) (retErr error) {
	defer func() {
		if retErr != nil {
			retErr = writeErrorFile(tw, pod.Name, retErr)
		}
	}()
	c, err := workerserver.NewClient(pod.Status.PodIP)
	if err != nil {
		return err
	}
	request := &debug.BinaryRequest{
		Filter: &debug.Filter{
			Filter: &debug.Filter_Worker{
				Worker: &debug.Worker{
					Pod:        pod.Name,
					Redirected: true,
				},
			},
		},
	}
	binaryC, err := c.Binary(ctx, request)
	if err != nil {
		return err
	}
	tr := tar.NewReader(grpcutil.NewStreamingBytesReader(binaryC, nil))
	return copyTar(tw, tr, func(name string) string {
		return strings.TrimPrefix(path.Join(pod.Name, name), "/")
	})
}

func collectBinary(tw *tar.Writer, container string) (retErr error) {
	f, err := os.Open(os.Args[0])
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); retErr == nil {
			retErr = err
		}
	}()
	return writeTarFile(tw, path.Join(container, "binary"), f)
}

func (s *debugServer) Dump(request *debug.DumpRequest, server debug.Debug_DumpServer) error {
	return withTarWriter(grpcutil.NewStreamingBytesWriter(server), func(tw *tar.Writer) error {
		// Handle filter.
		if request.Filter != nil {
			switch filter := request.Filter.Filter.(type) {
			case *debug.Filter_Pachd:
				return s.collectPachdDump(tw)
			case *debug.Filter_Pipeline:
				return s.collectPipeline(filter.Pipeline, func(pod *v1.Pod) error {
					return s.dumpRedirect(server.Context(), tw, pod)
				})
			case *debug.Filter_Worker:
				if filter.Worker.Redirected {
					// Collect the storage container dump.
					if s.sidecarClient == nil {
						return collectDump(tw, client.PPSWorkerSidecarContainerName)
					}
					// Collect the user container dump.
					if err := collectDump(tw, client.PPSWorkerUserContainerName); err != nil {
						return err
					}
					// Collect the dump from the storage container.
					dumpC, err := s.sidecarClient.DebugClient.Dump(server.Context(), request)
					if err != nil {
						return err
					}
					return copyTar(tw, tar.NewReader(grpcutil.NewStreamingBytesReader(dumpC, nil)))
				}
				pod, err := s.kubeClient.CoreV1().Pods(s.namespace).Get(filter.Worker.Pod, metav1.GetOptions{})
				if err != nil {
					return err
				}
				return s.dumpRedirect(server.Context(), tw, pod)
			}
		}
		// No filter, collect everything.
		if err := s.collectPachdDump(tw); err != nil {
			return err
		}
		return s.collectPipeline(&debug.Pipeline{Name: ""}, func(pod *v1.Pod) error {
			return s.dumpRedirect(server.Context(), tw, pod)
		})
	})
}

func (s *debugServer) collectPachdDump(tw *tar.Writer) error {
	// Collect the pachd container logs.
	if err := s.collectLogs(tw, s.name, "pachd"); err != nil {
		return err
	}
	// Collect the pachd container dump.
	return collectDump(tw, path.Join(s.name, "pachd"))
}

func (s *debugServer) collectLogs(tw *tar.Writer, pod, container string) (retErr error) {
	defer func() {
		if retErr != nil {
			retErr = writeErrorFile(tw, path.Join(pod, container, "logs"), retErr)
		}
	}()
	stream, err := s.kubeClient.CoreV1().Pods(s.namespace).GetLogs(pod, &v1.PodLogOptions{Container: container}).Stream()
	if err != nil {
		return err
	}
	defer func() {
		if err := stream.Close(); retErr == nil {
			retErr = err
		}
	}()
	return withTmpFile(func(f *os.File) error {
		if _, err := io.Copy(f, stream); err != nil {
			return err
		}
		return writeTarFile(tw, path.Join(pod, container, "logs"), f)
	})
}

func collectDump(tw *tar.Writer, container string) error {
	if err := collectProfile(tw, container, &debug.Profile{Name: "goroutine"}); err != nil {
		return err
	}
	return collectProfile(tw, container, &debug.Profile{Name: "heap"})
}

func (s *debugServer) dumpRedirect(ctx context.Context, tw *tar.Writer, pod *v1.Pod) (retErr error) {
	defer func() {
		if retErr != nil {
			retErr = writeErrorFile(tw, pod.Name, retErr)
		}
	}()
	// Collect the worker user and storage container logs.
	if err := s.collectLogs(tw, pod.Name, client.PPSWorkerUserContainerName); err != nil {
		return err
	}
	if err := s.collectLogs(tw, pod.Name, client.PPSWorkerSidecarContainerName); err != nil {
		return err
	}
	// Redirect the dump request to the worker pod (user container).
	c, err := workerserver.NewClient(pod.Status.PodIP)
	if err != nil {
		return err
	}
	request := &debug.DumpRequest{
		Filter: &debug.Filter{
			Filter: &debug.Filter_Worker{
				Worker: &debug.Worker{
					Pod:        pod.Name,
					Redirected: true,
				},
			},
		},
	}
	dumpC, err := c.Dump(ctx, request)
	if err != nil {
		return err
	}
	tr := tar.NewReader(grpcutil.NewStreamingBytesReader(dumpC, nil))
	return copyTar(tw, tr, func(name string) string {
		return strings.TrimPrefix(path.Join(pod.Name, name), "/")
	})
}
