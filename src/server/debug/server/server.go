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

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/debug"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	workerserver "github.com/pachyderm/pachyderm/src/server/worker/server"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube "k8s.io/client-go/kubernetes"
)

const (
	defaultDuration = time.Minute
	user            = "user"
	storage         = "storage"
)

// NewDebugServer creates a new server that serves the debug api over GRPC
// TODO: Move over to functional options before this gets too unwieldy.
func NewDebugServer(name string, etcdClient *etcd.Client, etcdPrefix string, workerGrpcPort uint16, clusterID string, sidecarClient *client.APIClient, kubeClient *kube.Clientset, namespace string) debug.DebugServer {
	return &debugServer{
		name:           name,
		etcdClient:     etcdClient,
		etcdPrefix:     etcdPrefix,
		workerGrpcPort: workerGrpcPort,
		clusterID:      clusterID,
		sidecarClient:  sidecarClient,
		kubeClient:     kubeClient,
		namespace:      namespace,
	}
}

type debugServer struct {
	name           string
	etcdClient     *etcd.Client
	etcdPrefix     string
	workerGrpcPort uint16
	clusterID      string
	sidecarClient  *client.APIClient
	kubeClient     *kube.Clientset
	namespace      string
}

func (s *debugServer) Profile(request *debug.ProfileRequest, server debug.Debug_ProfileServer) error {
	if request.Worker != nil {
		if err := validateWorker(request.Worker); err != nil {
			return err
		}
		// Redirect the request to the worker pod (user container).
		if request.Worker.Pod != s.name {
			return s.workerRedirect(server.Context(), request.Worker, func(c workerserver.Client) error {
				profileC, err := c.Profile(server.Context(), request)
				if err != nil {
					return err
				}
				return grpcutil.WriteFromStreamingBytesClient(profileC, grpcutil.NewStreamingBytesWriter(server))
			})
		}
		// Setup a tar writer, then write the user and storage container profile.
		return withTarWriter(grpcutil.NewStreamingBytesWriter(server), func(tw *tar.Writer) error {
			// Collect the user container profile.
			if err := collectProfile(tw, user, request.Profile); err != nil {
				return err
			}
			// Collect the storage container profile.
			request.Worker = nil
			profileC, err := s.sidecarClient.DebugClient.Profile(server.Context(), request)
			if err != nil {
				return err
			}
			return copyTar(tw, tar.NewReader(grpcutil.NewStreamingBytesReader(profileC, nil)))
		})
	}
	// Collect pachd / storage container profile.
	return withTarWriter(grpcutil.NewStreamingBytesWriter(server), func(tw *tar.Writer) error {
		return collectProfile(tw, storage, request.Profile)
	})
}

func validateWorker(worker *debug.Worker) error {
	if worker.Pod == "" {
		return errors.Errorf("worker pod name cannot be empty")
	}
	return nil
}

func (s *debugServer) workerRedirect(ctx context.Context, worker *debug.Worker, cb func(workerserver.Client) error) error {
	pod, err := s.kubeClient.CoreV1().Pods(s.namespace).Get(worker.Pod, metav1.GetOptions{})
	if err != nil {
		return err
	}
	workerClients, err := workerserver.Clients(ctx, "", s.etcdClient, s.etcdPrefix, s.workerGrpcPort, pod.Status.PodIP)
	if err != nil {
		return err
	}
	if len(workerClients) == 0 {
		return errors.Errorf("unable to find worker with pod name %v and IP address %v", worker.Pod, pod.Status.PodIP)
	}
	return cb(workerClients[0])
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

func collectProfile(tw *tar.Writer, container string, profile *debug.Profile) error {
	return withTmpFile(func(f *os.File) error {
		if err := writeProfile(f, profile); err != nil {
			return err
		}
		return writeTarFile(tw, path.Join(container, profile.Name), f)
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

func copyTar(tw *tar.Writer, tr *tar.Reader) error {
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
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
	if request.Worker != nil {
		if err := validateWorker(request.Worker); err != nil {
			return err
		}
		// Redirect the request to the worker pod (user container).
		if request.Worker.Pod != s.name {
			return s.workerRedirect(server.Context(), request.Worker, func(c workerserver.Client) error {
				binaryC, err := c.Binary(server.Context(), request)
				if err != nil {
					return err
				}
				return grpcutil.WriteFromStreamingBytesClient(binaryC, grpcutil.NewStreamingBytesWriter(server))
			})
		}
		// Setup a tar writer, then write the user and storage container binary.
		return withTarWriter(grpcutil.NewStreamingBytesWriter(server), func(tw *tar.Writer) error {
			// Collect the user container binary.
			if err := collectBinary(tw, user); err != nil {
				return err
			}
			// Collect the storage container binary.
			request.Worker = nil
			binaryC, err := s.sidecarClient.DebugClient.Binary(server.Context(), request)
			if err != nil {
				return err
			}
			return copyTar(tw, tar.NewReader(grpcutil.NewStreamingBytesReader(binaryC, nil)))
		})
	}
	// Collect pachd / storage container binary.
	return withTarWriter(grpcutil.NewStreamingBytesWriter(server), func(tw *tar.Writer) error {
		return collectBinary(tw, storage)
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

func (s *debugServer) SOS(request *debug.SOSRequest, server debug.Debug_SOSServer) error {
	if request.Worker != nil {
		if err := validateWorker(request.Worker); err != nil {
			return err
		}
		// Redirect the request to the worker pod (user container).
		if request.Worker.Pod != s.name {
			return s.workerRedirect(server.Context(), request.Worker, func(c workerserver.Client) error {
				sosC, err := c.SOS(server.Context(), request)
				if err != nil {
					return err
				}
				return grpcutil.WriteFromStreamingBytesClient(sosC, grpcutil.NewStreamingBytesWriter(server))
			})
		}
		// Setup a tar writer, then write the user and storage container sos.
		return withTarWriter(grpcutil.NewStreamingBytesWriter(server), func(tw *tar.Writer) error {
			// Collect the user container sos.
			if err := collectSOS(tw, user); err != nil {
				return err
			}
			// Collect the storage container sos.
			request.Worker = nil
			sosC, err := s.sidecarClient.DebugClient.SOS(server.Context(), request)
			if err != nil {
				return err
			}
			return copyTar(tw, tar.NewReader(grpcutil.NewStreamingBytesReader(sosC, nil)))
		})
	}
	// Collect pachd / storage container sos.
	return withTarWriter(grpcutil.NewStreamingBytesWriter(server), func(tw *tar.Writer) error {
		return collectSOS(tw, storage)
	})
}

func collectSOS(tw *tar.Writer, container string) error {
	if err := collectProfile(tw, container, &debug.Profile{Name: "goroutine"}); err != nil {
		return err
	}
	return collectProfile(tw, container, &debug.Profile{Name: "heap"})
}
