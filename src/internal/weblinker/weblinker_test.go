package weblinker

import (
	"regexp"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/version/versionpb"
)

func r(x string) *regexp.Regexp {
	return regexp.MustCompile(x)
}

func TestLinks(t *testing.T) {
	testData := []struct {
		name                                                                 string
		l                                                                    *Linker
		matchArchiveDownloadBaseURL, matchCreatePipelineRequestJSONSchemaURL *regexp.Regexp
	}{
		{
			name:                                    "empty",
			l:                                       new(Linker),
			matchArchiveDownloadBaseURL:             r("^/archive/$"),
			matchCreatePipelineRequestJSONSchemaURL: r("^https://raw.githubusercontent.*/master/.*/pps_v2/CreatePipelineRequest.schema.json$"),
		},
		{
			name: "release, http",
			l: &Linker{
				HTTPS:    false,
				HostPort: "pachyderm.example.com",
				Version: &versionpb.Version{
					Major:           2,
					Minor:           8,
					Micro:           0,
					GitCommit:       "abc123",
					GitTreeModified: "false",
				},
			},
			matchArchiveDownloadBaseURL:             r("^http://pachyderm.example.com/archive/$"),
			matchCreatePipelineRequestJSONSchemaURL: r("^http://pachyderm.example.com/jsonschema/pps_v2/CreatePipelineRequest.schema.json$"),
		},
		{
			name: "release, https",
			l: &Linker{
				HTTPS:    true,
				HostPort: "pachyderm.example.com",
				Version: &versionpb.Version{
					Major:           2,
					Minor:           8,
					Micro:           0,
					GitCommit:       "abc123",
					GitTreeModified: "false",
				},
			},
			matchArchiveDownloadBaseURL:             r("^https://pachyderm.example.com/archive/$"),
			matchCreatePipelineRequestJSONSchemaURL: r("^https://pachyderm.example.com/jsonschema/pps_v2/CreatePipelineRequest.schema.json$"),
		},
		{
			name: "release, http with custom port",
			l: &Linker{
				HTTPS:    false,
				HostPort: "pachyderm.example.com:1234",
				Version: &versionpb.Version{
					Major:           2,
					Minor:           8,
					Micro:           0,
					GitCommit:       "abc123",
					GitTreeModified: "false",
				},
			},
			matchArchiveDownloadBaseURL:             r("^http://pachyderm.example.com:1234/archive/$"),
			matchCreatePipelineRequestJSONSchemaURL: r("^http://pachyderm.example.com:1234/jsonschema/pps_v2/CreatePipelineRequest.schema.json$"),
		},
		{
			name: "release, no proxy",
			l: &Linker{
				HTTPS:    false,
				HostPort: "",
				Version: &versionpb.Version{
					Major:           2,
					Minor:           8,
					Micro:           0,
					GitCommit:       "abc123",
					GitTreeModified: "false",
				},
			},
			matchArchiveDownloadBaseURL:             r("^/archive/$"),
			matchCreatePipelineRequestJSONSchemaURL: r("^https://raw.githubusercontent.com/.*/v2.8.0/.*/pps_v2/CreatePipelineRequest.schema.json$"),
		},
		{
			name: "git checkout, not modified, no proxy",
			l: &Linker{
				HTTPS:    false,
				HostPort: "",
				Version: &versionpb.Version{
					Major:           0,
					Minor:           0,
					Micro:           0,
					GitCommit:       "abc123",
					GitTreeModified: "false",
				},
			},
			matchArchiveDownloadBaseURL:             r("^/archive/$"),
			matchCreatePipelineRequestJSONSchemaURL: r("^https://raw.githubusercontent.com/.*/abc123/.*/pps_v2/CreatePipelineRequest.schema.json$"),
		},
		{
			name: "git checkout, modified, no proxy",
			l: &Linker{
				HTTPS:    false,
				HostPort: "",
				Version: &versionpb.Version{
					Major:           0,
					Minor:           0,
					Micro:           0,
					GitCommit:       "abc123",
					GitTreeModified: "true",
				},
			},
			matchArchiveDownloadBaseURL:             r("^/archive/$"),
			matchCreatePipelineRequestJSONSchemaURL: r("^https://raw.githubusercontent.com/.*/master/.*/pps_v2/CreatePipelineRequest.schema.json$"),
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			got := test.l.ArchiveDownloadBaseURL()
			if !test.matchArchiveDownloadBaseURL.MatchString(got) {
				t.Errorf("AchiveDownloadBaseURL:\n  got:  %v\n want: /%v/", got, test.matchArchiveDownloadBaseURL.String())
			}
			got = test.l.CreatePipelineRequestJSONSchemaURL()
			if !test.matchCreatePipelineRequestJSONSchemaURL.MatchString(got) {
				t.Errorf("CreatePipelineRequestJSONSchemaURL:\n  got:  %v\n want: /%v/", got, test.matchCreatePipelineRequestJSONSchemaURL.String())
			}
		})
	}
}
