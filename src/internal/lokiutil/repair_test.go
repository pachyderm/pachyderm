package lokiutil

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestTryDocker(t *testing.T) {
	testData := []struct {
		name  string
		input map[string]any
		want  string
	}{
		{
			name: "empty",
			want: "",
		},
		{
			name: "valid",
			input: map[string]any{
				"log":    "log is here",
				"time":   "2020-01-01T12:34:56.789Z",
				"stream": "stdout",
			},
			want: "log is here",
		},
		{
			name: "minimal valid",
			input: map[string]any{
				"log": "log is here",
			},
			want: "log is here",
		},
		{
			name: "missing log",
			input: map[string]any{
				"time":   "2020-01-01T12:34:56.789Z",
				"stream": "stdout",
			},
			want: "",
		},
		{
			name: "log of wrong type",
			input: map[string]any{
				"log":    42,
				"time":   "2020-01-01T12:34:56.789Z",
				"stream": "stdout",
			},
			want: "",
		},
		{
			name: "extra fields",
			input: map[string]any{
				"log":              "log is here",
				"time":             "2020-01-01T12:34:56.789Z",
				"stream":           "stdout",
				"isThisDockerJSON": "no it is not",
			},
			want: "",
		},
		{
			name: "minimal extra fields",
			input: map[string]any{
				"log":              "log is here",
				"isThisDockerJSON": "no it is not",
			},
			want: "",
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			got := tryDocker(test.input)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("extract log message (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTryCRI(t *testing.T) {
	testData := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty",
			input: "",
			want:  "",
		},
		{
			name:  "simple",
			input: "2020-01-01T12:34:56.789Z stdout F this is a log",
			want:  "this is a log",
		},
		{
			name:  "invalid",
			input: "<not_the_time> <not_the_stream> <not_F_or_P> this is a log",
			want:  "<not_the_time> <not_the_stream> <not_F_or_P> this is a log",
		},
		{
			name:  "invalid time",
			input: "<not_the_time> stdout F this is a log",
			want:  "<not_the_time> stdout F this is a log",
		},
		{
			name:  "invalid stream",
			input: "2020-01-01T12:34:56.789Z not_the_stream F this is a log",
			want:  "2020-01-01T12:34:56.789Z not_the_stream F this is a log",
		},
		{
			name:  "invalid flag",
			input: "2020-01-01T12:34:56.789Z stdout X this is a log",
			want:  "2020-01-01T12:34:56.789Z stdout X this is a log",
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			got := tryCRI(test.input)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("extract log message (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRepairLine(t *testing.T) {
	testData := []struct {
		name     string
		input    string
		wantLine string
		wantJSON map[string]any
	}{
		{
			name: "empty",
		},
		{
			name:     "plain text",
			input:    "plain text",
			wantLine: "plain text",
		},
		{
			name:     "docker plain text",
			input:    `{"log":"plain text"}`,
			wantLine: "plain text",
		},
		{
			name:     "cri plain text",
			input:    "2020-01-01T12:34:56.789Z stdout F plain text",
			wantLine: "plain text",
		},
		{
			name:     "invalid json",
			input:    "{not json}",
			wantLine: "{not json}",
		},
		{
			name:     "invalid json inside docker",
			input:    `{"log":"{not json}"}`,
			wantLine: "{not json}",
		},
		{
			name:     "invalid json inside cri",
			input:    `2020-01-01T12:34:56.789Z stderr F {not json}`,
			wantLine: "{not json}",
		},
		{
			name:     "valid json",
			input:    `{"this":"is json"}`,
			wantLine: `{"this":"is json"}`,
			wantJSON: map[string]any{"this": "is json"},
		},
		{
			name:     "valid json inside docker",
			input:    `{"log":"{\"this\":\"is json\"}"}`,
			wantLine: `{"this":"is json"}`,
			wantJSON: map[string]any{"this": "is json"},
		},
		{
			name:     "valid json inside cri",
			input:    `2020-01-01T12:34:56.789Z stderr F {"this":"is json"}`,
			wantLine: `{"this":"is json"}`,
			wantJSON: map[string]any{"this": "is json"},
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			gotLine, gotJSON := RepairLine(test.input)
			if diff := cmp.Diff(test.wantLine, gotLine); diff != "" {
				t.Errorf("log message (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(test.wantJSON, gotJSON); diff != "" {
				t.Errorf("log object (-want +got):\n%s", diff)
			}
		})
	}
}

func BenchmarkRepairCRI(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RepairLine("2020-01-02T12:34:56.789Z stdout F this is a line")
		RepairLine("foobar stdout F this is a line")
	}
}
