package v2_8_0

import (
	"reflect"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/pps"
	"google.golang.org/protobuf/proto"
)

func TestSynthesizeClusterDefaults(t *testing.T) {
	var cases = map[string]struct {
		env       map[string]string
		expected  *pps.CreatePipelineRequest
		expectErr bool
	}{
		"defaults": {
			env: nil,
			expected: &pps.CreatePipelineRequest{
				ResourceRequests: &pps.ResourceSpec{
					Memory: "256Mi",
					Cpu:    1.0,
					Disk:   "1Gi",
				},
				SidecarResourceRequests: &pps.ResourceSpec{
					Memory: "256Mi",
					Cpu:    1.0,
					Disk:   "1Gi",
				},
			},
		},
		"badMemory": {
			env: map[string]string{
				"PIPELINE_DEFAULT_MEMORY_REQUEST": "kjhsdfkjhsdflkjj",
			},
			expectErr: true,
		},
		"badCPU": {
			env: map[string]string{
				"PIPELINE_DEFAULT_CPU_REQUEST": "-3.a",
			},
			expectErr: true,
		},
		"badDisk": {
			env: map[string]string{
				"PIPELINE_DEFAULT_STORAGE_REQUEST": "-400",
			},
			expectErr: true,
		},
		"good/sidecar defaults": {
			env: map[string]string{
				"PIPELINE_DEFAULT_MEMORY_REQUEST":  "128Mi",
				"PIPELINE_DEFAULT_CPU_REQUEST":     "4m",
				"PIPELINE_DEFAULT_STORAGE_REQUEST": "512Mi",
			},
			expected: &pps.CreatePipelineRequest{
				ResourceRequests: &pps.ResourceSpec{
					Memory: "128Mi",
					Cpu:    0.004,
					Disk:   "512Mi",
				},
				SidecarResourceRequests: &pps.ResourceSpec{
					Memory: "256Mi",
					Cpu:    1.0,
					Disk:   "1Gi",
				},
			},
		},
		"good/pipeline defaults": {
			env: map[string]string{
				"SIDECAR_DEFAULT_MEMORY_REQUEST":  "128Mi",
				"SIDECAR_DEFAULT_CPU_REQUEST":     "4m",
				"SIDECAR_DEFAULT_STORAGE_REQUEST": "512Mi",
			},
			expected: &pps.CreatePipelineRequest{
				ResourceRequests: &pps.ResourceSpec{
					Memory: "256Mi",
					Cpu:    1.0,
					Disk:   "1Gi",
				},
				SidecarResourceRequests: &pps.ResourceSpec{
					Memory: "128Mi",
					Cpu:    0.004,
					Disk:   "512Mi",
				},
			},
		},
		"good": {
			env: map[string]string{
				"PIPELINE_DEFAULT_MEMORY_REQUEST":  "196Mi",
				"PIPELINE_DEFAULT_CPU_REQUEST":     "12m",
				"PIPELINE_DEFAULT_STORAGE_REQUEST": "256Mi",
				"SIDECAR_DEFAULT_MEMORY_REQUEST":   "128Mi",
				"SIDECAR_DEFAULT_CPU_REQUEST":      "4m",
				"SIDECAR_DEFAULT_STORAGE_REQUEST":  "512Mi",
			},
			expected: &pps.CreatePipelineRequest{
				ResourceRequests: &pps.ResourceSpec{
					Memory: "196Mi",
					Cpu:    0.012,
					Disk:   "256Mi",
				},
				SidecarResourceRequests: &pps.ResourceSpec{
					Memory: "128Mi",
					Cpu:    0.004,
					Disk:   "512Mi",
				},
			},
		},
	}

	for n, c := range cases {
		got, err := defaultsFromEnv(c.env)
		if (err != nil) != c.expectErr {
			if err != nil {
				t.Errorf("%s: unexpected error: %v", n, err)
			} else {
				t.Errorf("%s: expected error", n)
			}
			continue
		} else if c.expectErr {
			continue
		}
		if got := got.CreatePipelineRequest; !proto.Equal(c.expected, got) {
			t.Errorf("%s: expected %v; got %v", n, c.expected, got)
		}
	}
}

func TestEnvMap(t *testing.T) {
	var cases = map[string]struct {
		environ   []string
		expected  map[string]string
		expectErr bool
	}{
		"empty environment is not wrong": {expected: make(map[string]string)},
		"good": {
			environ: []string{"FOO=BAR", "BAZ=QUUX=QUUUX"},
			expected: map[string]string{
				"BAZ": "QUUX=QUUUX",
				"FOO": "BAR",
			},
		},
		"bad": {
			environ:   []string{"FOO=BAR", "BAZ"},
			expectErr: true,
		},
	}
	for n, c := range cases {
		got, err := envMap(c.environ)
		if (err != nil) != c.expectErr {
			if err != nil {
				t.Errorf("%s: unexpected error: %v", n, err)
			} else {
				t.Errorf("%s: expected error", n)
			}
			continue
		} else if c.expectErr {
			continue
		}
		if !reflect.DeepEqual(c.expected, got) {
			t.Errorf("%s: expected %v; got %v", n, c.expected, got)
		}
	}
}
