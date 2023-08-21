package server

import (
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"google.golang.org/protobuf/types/known/wrapperspb"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func mergeDefaultOptions(project, name string, add *workerOptions) *workerOptions {
	base := &workerOptions{
		rcName: ppsutil.PipelineRcName(&pps.PipelineInfo{Pipeline: &pps.Pipeline{Project: &pfs.Project{Name: project}, Name: name}}),
		labels: map[string]string{
			"app":             "pipeline",
			"component":       "worker",
			"pipelineName":    name,
			"pipelineProject": project,
			"pipelineVersion": "0",
			"suite":           "pachyderm",
		},
		annotations: map[string]string{
			"authTokenHash":   hashAuthToken(""),
			"pachVersion":     "0.0.0",
			"pipelineName":    name,
			"pipelineProject": project,
			"pipelineVersion": "0",
			"specCommit":      "",
		},
		userImage: "ubuntu:20.04",
		resourceRequests: &v1.ResourceList{
			v1.ResourceCPU:              resource.MustParse("1"),
			v1.ResourceEphemeralStorage: resource.MustParse("100M"),
			v1.ResourceMemory:           resource.MustParse("100M"),
		},
		workerEnv: []v1.EnvVar{},
		volumes: []v1.Volume{
			{
				Name:         "pach-bin",
				VolumeSource: v1.VolumeSource{EmptyDir: &v1.EmptyDirVolumeSource{}},
			},
			{
				Name:         "pachyderm-worker",
				VolumeSource: v1.VolumeSource{EmptyDir: &v1.EmptyDirVolumeSource{}},
			},
		},
		volumeMounts: []v1.VolumeMount{
			{Name: "pach-bin", MountPath: "/pach-bin"},
			{Name: "pachyderm-worker", MountPath: "/pfs"},
		},
		postgresSecret: &v1.SecretKeySelector{
			LocalObjectReference: v1.LocalObjectReference{
				Name: "secret",
			},
			Key: "key",
		},
	}
	if add == nil {
		return base
	}
	// This is not ideal, but copy your thing over when you notice the test isn't working.  Be
	// sure to add a condition to not do this when your value is the zero value.
	if r := add.resourceRequests; r != nil {
		base.resourceRequests = r
	}
	base.tolerations = append(base.tolerations, add.tolerations...)
	return base
}

func TestGetWorkerOptions(t *testing.T) {
	testData := []struct {
		name        string
		pipeline    *pps.PipelineInfo
		wantOptions *workerOptions
		wantError   *regexp.Regexp
	}{
		{
			name: "empty",
			pipeline: &pps.PipelineInfo{
				Pipeline: &pps.Pipeline{
					Project: &pfs.Project{
						Name: "default",
					},
					Name: "empty",
				},
				Details: &pps.PipelineInfo_Details{
					Transform: &pps.Transform{},
				},
				SpecCommit: &pfs.Commit{},
			},
			wantOptions: mergeDefaultOptions("default", "empty", nil),
		},
		{
			name: "no default resources",
			pipeline: &pps.PipelineInfo{
				Pipeline: &pps.Pipeline{
					Project: &pfs.Project{
						Name: "default",
					},
					Name: "empty",
				},
				Details: &pps.PipelineInfo_Details{
					Transform:        &pps.Transform{},
					ResourceRequests: &pps.ResourceSpec{},
				},
				SpecCommit: &pfs.Commit{},
			},
			wantOptions: mergeDefaultOptions("default", "empty", &workerOptions{
				resourceRequests: &v1.ResourceList{},
			}),
		},
		{
			name: "one toleration",
			pipeline: &pps.PipelineInfo{
				Pipeline: &pps.Pipeline{
					Project: &pfs.Project{
						Name: "default",
					},
					Name: "toleration",
				},
				Details: &pps.PipelineInfo_Details{
					Transform: &pps.Transform{},
					Tolerations: []*pps.Toleration{
						{
							Key:      "special-machine",
							Operator: pps.TolerationOperator_EXISTS,
							Effect:   pps.TaintEffect_NO_SCHEDULE,
						},
					},
				},
				SpecCommit: &pfs.Commit{},
			},
			wantOptions: mergeDefaultOptions("default", "toleration", &workerOptions{
				tolerations: []v1.Toleration{
					{
						Key:      "special-machine",
						Operator: v1.TolerationOpExists,
						Effect:   v1.TaintEffectNoSchedule,
					},
				},
			}),
		},
		{
			name: "many tolerations",
			pipeline: &pps.PipelineInfo{
				Pipeline: &pps.Pipeline{
					Project: &pfs.Project{
						Name: "default",
					},
					Name: "tolerations",
				},
				Details: &pps.PipelineInfo_Details{
					Transform: &pps.Transform{},
					Tolerations: []*pps.Toleration{
						{
							Key:      "special-machine",
							Operator: pps.TolerationOperator_EXISTS,
							Effect:   pps.TaintEffect_NO_SCHEDULE,
						},
						{
							Key:               "foo",
							Operator:          pps.TolerationOperator_EQUAL,
							Value:             "bar",
							Effect:            pps.TaintEffect_NO_EXECUTE,
							TolerationSeconds: wrapperspb.Int64(100),
						},
					},
				},
				SpecCommit: &pfs.Commit{},
			},
			wantOptions: mergeDefaultOptions("default", "tolerations", &workerOptions{
				tolerations: []v1.Toleration{
					{
						Key:      "special-machine",
						Operator: v1.TolerationOpExists,
						Effect:   v1.TaintEffectNoSchedule,
					},
					{
						Key:               "foo",
						Operator:          v1.TolerationOpEqual,
						Value:             "bar",
						Effect:            v1.TaintEffectNoExecute,
						TolerationSeconds: int64Ptr(100),
					},
				},
			}),
		},
		{
			name: "invalid toleration",
			pipeline: &pps.PipelineInfo{
				Pipeline: &pps.Pipeline{
					Project: &pfs.Project{
						Name: "default",
					},
					Name: "tolerations",
				},
				Details: &pps.PipelineInfo_Details{
					Transform: &pps.Transform{},
					Tolerations: []*pps.Toleration{
						{
							Key:      "special-machine",
							Operator: pps.TolerationOperator_EXISTS,
							Effect:   pps.TaintEffect_NO_SCHEDULE,
						},
						{
							Key:               "foo",
							Operator:          pps.TolerationOperator_EMPTY,
							Value:             "bar",
							Effect:            pps.TaintEffect_NO_EXECUTE,
							TolerationSeconds: wrapperspb.Int64(100),
						},
					},
				},
				SpecCommit: &pfs.Commit{},
			},
			wantError: regexp.MustCompile("toleration 2/2: cannot omit"),
		},
	}

	// Build a fake kubedriver.
	kd := &kubeDriver{
		namespace: "test",
		config: pachconfig.Configuration{
			// Some of these values end up in mergeDefaultOptions above.
			GlobalConfiguration: &pachconfig.GlobalConfiguration{
				PipelineDefaultCPURequest:     resource.MustParse("1"),
				PipelineDefaultMemoryRequest:  resource.MustParse("100M"),
				PipelineDefaultStorageRequest: resource.MustParse("100M"),
			},
			PachdSpecificConfiguration: &pachconfig.PachdSpecificConfiguration{
				PachdPodName: "pachd",
			},
			WorkerSpecificConfiguration:     &pachconfig.WorkerSpecificConfiguration{},
			EnterpriseSpecificConfiguration: &pachconfig.EnterpriseSpecificConfiguration{},
		},
		kubeClient: fake.NewSimpleClientset(&v1.Pod{
			// Minimal pachd pod for some introspection that happens.
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pachd",
				Namespace: "test",
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name: "pachd",
						Env: []v1.EnvVar{
							{
								Name: "POSTGRES_PASSWORD",
								ValueFrom: &v1.EnvVarSource{
									SecretKeyRef: &v1.SecretKeySelector{
										Key: "key",
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
									},
								},
							},
						},
					},
				},
			},
		}),
	}

	// Test starts here.
	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			ctx := pctx.TestContext(t)
			got, err := kd.getWorkerOptions(ctx, test.pipeline)
			if err != nil && test.wantError == nil {
				t.Fatalf("unexpected error: %v", err)
			} else if err == nil && test.wantError != nil {
				t.Error("expected error, but got nil")
				// wantOptions is probably nil, but fall through so cmp.Diff can
				// print "got"
			} else if err != nil && test.wantError != nil {
				if ok := test.wantError.MatchString(err.Error()); !ok {
					t.Fatalf("expected error:\n  got: %v\n want: /%s/", err.Error(), test.wantError.String())
				}
				return
			}
			if diff := cmp.Diff(got, test.wantOptions, cmp.AllowUnexported(workerOptions{})); diff != "" {
				t.Errorf("options (+got -want):\n%s", diff)
			}
		})
	}
}
