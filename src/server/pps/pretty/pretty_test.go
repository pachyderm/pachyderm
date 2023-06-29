package pretty_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	pfsclient "github.com/pachyderm/pachyderm/v2/src/pfs"
	ppsclient "github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/server/pps/pretty"
)

func TestJobEgressTarget(t *testing.T) {
	var (
		buf   = new(bytes.Buffer)
		items = map[string]string{
			"URL":         "snowflake://user@ORG-ACCT",
			"secret name": "baz",
			"secret key":  "quux",
		}
		ji = &ppsclient.JobInfo{
			Job: &ppsclient.Job{
				Id: "foo",
				Pipeline: &ppsclient.Pipeline{
					Name:    "bar",
					Project: &pfsclient.Project{Name: pfsclient.DefaultProjectName},
				},
			},
			Stats: &ppsclient.ProcessStats{},
			Details: &ppsclient.JobInfo_Details{
				Egress: &ppsclient.Egress{
					Target: &ppsclient.Egress_SqlDatabase{
						SqlDatabase: &pfsclient.SQLDatabaseEgress{
							Url: items["URL"],
							Secret: &pfsclient.SQLDatabaseEgress_Secret{
								Name: items["secret name"],
								Key:  items["secret key"],
							},
						},
					},
				},
			},
		}
	)
	if err := pretty.PrintDetailedJobInfo(buf, pretty.NewPrintableJobInfo(ji, false)); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	for item, value := range items {
		if !strings.Contains(s, value) {
			t.Errorf("could not find egress %s in detailed job info; expected %q", item, value)
		}
	}
}

func TestJobEgressURL(t *testing.T) {
	var (
		buf   = new(bytes.Buffer)
		items = map[string]string{
			"URL": "example:baz",
		}
		ji = &ppsclient.JobInfo{
			Job: &ppsclient.Job{
				Id: "foo",
				Pipeline: &ppsclient.Pipeline{
					Name:    "bar",
					Project: &pfsclient.Project{Name: pfsclient.DefaultProjectName},
				},
			},
			Stats: &ppsclient.ProcessStats{},
			Details: &ppsclient.JobInfo_Details{
				Egress: &ppsclient.Egress{
					URL: items["URL"],
				},
			},
		}
	)
	if err := pretty.PrintDetailedJobInfo(buf, pretty.NewPrintableJobInfo(ji, false)); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	for item, value := range items {
		if !strings.Contains(s, value) {
			t.Errorf("could not find egress %s in detailed job info; expected %q", item, value)
		}
	}
}

func TestJobStarted(t *testing.T) {
	tests := map[string]struct {
		full     bool
		expected string
	}{
		"Full timestamp": {
			full:     true,
			expected: "Started: -",
		},
		"Pretty ago": {
			full:     false,
			expected: "Started: -",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			jobInfo := &ppsclient.JobInfo{
				Job: &ppsclient.Job{
					Id: "foo",
					Pipeline: &ppsclient.Pipeline{
						Project: &pfsclient.Project{Name: pfsclient.DefaultProjectName},
						Name:    "Bar",
					},
				},
				Started: nil,
				Stats:   &ppsclient.ProcessStats{},
				Details: &ppsclient.JobInfo_Details{},
			}
			printableJobInfo := pretty.NewPrintableJobInfo(jobInfo, test.full)
			buf := new(bytes.Buffer)
			require.NoError(t, pretty.PrintDetailedJobInfo(buf, printableJobInfo))
			require.True(t, strings.Contains(buf.String(), test.expected))
		})
	}
}

func TestPipelineEgressTarget(t *testing.T) {
	var (
		buf   = new(bytes.Buffer)
		items = map[string]string{
			"URL":         "snowflake://user@ORG-ACCT",
			"secret name": "baz",
			"secret key":  "quux",
		}
		pi = &ppsclient.PipelineInfo{
			Pipeline: &ppsclient.Pipeline{
				Name:    "foo",
				Project: &pfsclient.Project{Name: pfsclient.DefaultProjectName},
			},
			Details: &ppsclient.PipelineInfo_Details{
				Description: "bar",
				Egress: &ppsclient.Egress{
					Target: &ppsclient.Egress_SqlDatabase{
						SqlDatabase: &pfsclient.SQLDatabaseEgress{
							Url: items["URL"],
							Secret: &pfsclient.SQLDatabaseEgress_Secret{
								Name: items["secret name"],
								Key:  items["secret key"],
							},
						},
					},
				},
			},
		}
	)
	if err := pretty.PrintDetailedPipelineInfo(buf, pretty.NewPrintablePipelineInfo(pi)); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	for item, value := range items {
		if !strings.Contains(s, value) {
			t.Errorf("could not find egress %s in detailed pipeline info; expected %q", item, value)
		}
	}
}

func TestPipelineEgressURL(t *testing.T) {
	var (
		buf   = new(bytes.Buffer)
		items = map[string]string{
			"URL": "example:baz",
		}
		pi = &ppsclient.PipelineInfo{
			Pipeline: &ppsclient.Pipeline{
				Name:    "foo",
				Project: &pfsclient.Project{Name: pfsclient.DefaultProjectName},
			},
			Details: &ppsclient.PipelineInfo_Details{
				Description: "bar",
				Egress: &ppsclient.Egress{
					URL: items["URL"],
				},
			},
		}
	)
	if err := pretty.PrintDetailedPipelineInfo(buf, pretty.NewPrintablePipelineInfo(pi)); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	for item, value := range items {
		if !strings.Contains(s, value) {
			t.Errorf("could not find egress %s in detailed pipeline info; expected %q", item, value)
		}
	}
}

func TestParseKubeEvent(t *testing.T) {
	testData := []struct {
		name        string
		line        string
		wantMessage string
		wantErr     bool
	}{
		{
			name:    "empty",
			line:    "",
			wantErr: true,
		},
		{
			name:    "invalid json",
			line:    "{this is not json}",
			wantErr: true,
		},
		{
			name:    "useless json",
			line:    "{}",
			wantErr: true,
		},
		{
			name:        "docker json",
			line:        `{"log":"{\"msg\":\"ok\"}"}`,
			wantMessage: "ok",
		},
		{
			name:    "docker with invalid json inside",
			line:    `{"log":"{this is not json}"}`,
			wantErr: true,
		},
		{
			name:        "native json",
			line:        `{"msg":"ok"}`,
			wantMessage: "ok",
		},
		{
			name:        "cri format with flags and valid message",
			line:        `2022-01-01T00:00:00.1234 stdout F {"msg":"ok"}`,
			wantMessage: "ok",
		},
		{
			name:        "cri format without flags and valid message",
			line:        `2022-01-01T00:00:00.1234 stdout {"msg":"ok"}`,
			wantMessage: "ok",
		},
		{
			name:        "cri format with flags and valid message and extra fields",
			line:        `2022-01-01T00:00:00.1234 stdout F {"msg":"ok","extraField":42}`,
			wantMessage: "ok",
		},
		{
			name:        "cri format without flags and valid message and extra fields",
			line:        `2022-01-01T00:00:00.1234 stdout {"msg":"ok","extraField":42}`,
			wantMessage: "ok",
		},
		{
			name:    "cri format with flags and EOF",
			line:    `2022-01-01T00:00:00.1234 stdout F`,
			wantErr: true,
		},
		{
			name:    "cri format without flags and EOF",
			line:    `2022-01-01T00:00:00.1234 stdout`,
			wantErr: true,
		},
		{
			name:    "cri format with flags and invalid json",
			line:    `2022-01-01T00:00:00.1234 stdout F this is not json`,
			wantErr: true,
		},
		{
			name:    "cri format without flags and invalid json",
			line:    `2022-01-01T00:00:00.1234 stdout this is not json`,
			wantErr: true,
		},
		{
			name:    "cri format with flags and EOF right after {",
			line:    `2022-01-01T00:00:00.1234 stdout F {`,
			wantErr: true,
		},
		{
			name:    "cri format without flags and EOF right after {",
			line:    `2022-01-01T00:00:00.1234 stdout {`,
			wantErr: true,
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			var event pretty.KubeEvent
			err := pretty.ParseKubeEvent(test.line, &event)
			t.Logf("err: %v", err)
			if test.wantErr && err == nil {
				t.Fatal("parse: got success, want error")
			} else if !test.wantErr && err != nil {
				t.Fatalf("parse: unexpected error: %v", err)
			}
			if got, want := event.Message, test.wantMessage; got != want {
				t.Fatalf("parse: message:\n  got: %v\n want: %v", got, want)
			}
		})
	}
}
