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
				ID: "foo",
				Pipeline: &ppsclient.Pipeline{
					Name: "bar",
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
				ID: "foo",
				Pipeline: &ppsclient.Pipeline{
					Name: "bar",
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
					ID:       "foo",
					Pipeline: &ppsclient.Pipeline{Name: "Bar"},
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
				Name: "foo",
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
				Name: "foo",
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
