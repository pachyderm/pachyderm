package server

import (
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"go.uber.org/zap/zapcore"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/durationpb"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

type level struct {
	sync.Mutex
	level zapcore.Level
}

func (l *level) SetLevelFor(newLevel zapcore.Level, d time.Duration, notify func(string, string)) {
	l.Lock()
	l.level = newLevel
	l.Unlock()
}

func (l *level) Level() zapcore.Level {
	l.Lock()
	defer l.Unlock()
	return l.level
}

var _ log.LevelChanger = new(level)

func TestSetLogLevel(t *testing.T) {
	testData := []struct {
		name          string
		req           *debug.SetLogLevelRequest
		wantResponse  *debug.SetLogLevelResponse
		wantLevel     zapcore.Level
		wantGRPCLevel zapcore.Level
		wantErr       bool
	}{
		{
			name:         "empty request",
			req:          &debug.SetLogLevelRequest{},
			wantResponse: &debug.SetLogLevelResponse{},
			wantErr:      true,
		},
		{
			name: "change log level",
			req: &debug.SetLogLevelRequest{
				Level: &debug.SetLogLevelRequest_Pachyderm{
					Pachyderm: debug.SetLogLevelRequest_DEBUG,
				},
				Duration: durationpb.New(time.Minute),
			},
			wantResponse: &debug.SetLogLevelResponse{
				AffectedPods: []string{
					"the-tests",
				},
			},
			wantLevel: zapcore.DebugLevel,
		},
		{
			name: "change grpc level",
			req: &debug.SetLogLevelRequest{
				Level: &debug.SetLogLevelRequest_Grpc{
					Grpc: debug.SetLogLevelRequest_DEBUG,
				},
				Duration: durationpb.New(time.Minute),
			},
			wantResponse: &debug.SetLogLevelResponse{
				AffectedPods: []string{
					"the-tests",
				},
			},
			wantGRPCLevel: zapcore.DebugLevel,
		},
		{
			name: "change log level recursively",
			req: &debug.SetLogLevelRequest{
				Level: &debug.SetLogLevelRequest_Pachyderm{
					Pachyderm: debug.SetLogLevelRequest_DEBUG,
				},
				Duration: durationpb.New(time.Minute),
				Recurse:  true,
			},
			wantResponse: &debug.SetLogLevelResponse{
				AffectedPods: []string{
					"the-tests",
				},
				ErroredPods: []string{
					"pachw@pod.invalid.:1653(dial: context deadline exceeded)",
				},
			},
			wantLevel: zapcore.DebugLevel,
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			ctx := log.Test(t)
			logLevel := new(level) // The zero value of zapcore.Level is InfoLevel.
			grpcLevel := new(level)

			s := &debugServer{
				name: "the-tests",
				env: &serviceenv.TestServiceEnv{
					KubeClient: fake.NewSimpleClientset(
						&v1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Name: "the-tests",
								Labels: map[string]string{
									"suite": "pachyderm",
									"app":   "pachd",
								},
							},
						},
						&v1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Name: "pachw",
								Labels: map[string]string{
									"suite": "pachyderm",
									"app":   "pachw",
								},
							},
							Status: v1.PodStatus{
								PodIP: "pod.invalid.",
							},
						},
					),
					Configuration: &pachconfig.Configuration{
						GlobalConfiguration: &pachconfig.GlobalConfiguration{
							Port:     1650,
							PeerPort: 1653,
						},
					},
				},
				logLevel:  logLevel,
				grpcLevel: grpcLevel,
			}

			res, err := s.SetLogLevel(ctx, test.req)
			if err == nil && test.wantErr {
				t.Fatal("expected error, but succeeded")
			} else if err != nil && test.wantErr {
				return
			} else if err != nil && !test.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(res, test.wantResponse, protocmp.Transform()); diff != "" {
				t.Errorf("response (-got +want):\n%s", diff)
			}

			if got, want := logLevel.Level(), test.wantLevel; got != want {
				t.Errorf("log level:\n  got: %v\n want: %v", got, want)
			}
			if got, want := grpcLevel.Level(), test.wantGRPCLevel; got != want {
				t.Errorf("grpc level:\n  got: %v\n want: %v", got, want)
			}
		})
	}
}
