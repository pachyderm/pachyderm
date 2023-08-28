// Package weblinker generates links to Pachyderm resources served over HTTP.
package weblinker

import (
	"fmt"
	"net/url"

	"github.com/pachyderm/pachyderm/v2/src/version"
	"github.com/pachyderm/pachyderm/v2/src/version/versionpb"
)

// Linker can generate links to HTTP resources.
type Linker struct {
	// HTTPS is true if resources should be fetched with HTTPS.
	HTTPS bool
	// HostPort is the host:port (port optional if 80/443) of the Pachyderm proxy.
	HostPort string
	// Version is the version of Pachyderm that will handle the links.
	Version *versionpb.Version
}

func (l *Linker) baseURL() url.URL {
	scheme := "http"
	if l.HTTPS {
		scheme = "https"
	}
	return url.URL{
		Scheme: scheme,
		Host:   l.HostPort,
	}
}

// ArchiveDownloadBaseURL returns the prefix of archive download URLs.
func (l *Linker) ArchiveDownloadBaseURL() string {
	if l.HostPort == "" {
		return "/archive/"
	}
	base := l.baseURL()
	base.Path = "/archive/"
	return base.String()
}

// CreatePipelineRequestJSONSchemaURL returns the URL of the JSON Schema for a CreatePipelineRequest.
func (l *Linker) CreatePipelineRequestJSONSchemaURL() string {
	if l.Version == nil {
		l.Version = version.Version
	}
	if l.HostPort == "" {
		v := l.Version
		tag := "master"
		major, minor, micro := v.GetMajor(), v.GetMinor(), v.GetMicro()
		switch {
		case major > 0 && v.GetGitTreeModified() == "false": // 3.0.0 will have this feature.
			// This looks like a release, so use the release tag.
			tag = fmt.Sprintf("v%v.%v.%v", major, minor, micro)
		case major == 0 && v.GetGitTreeModified() == "false" && v.GetGitCommit() != "":
			// This looks like a local build, so use the git commit if the tree is
			// unmodified.  If git is modified, this is probably an in-progress PR and
			// we should just use master, since the commit might not be on Github yet.
			tag = v.GetGitCommit()
		}

		return "https://raw.githubusercontent.com/pachyderm/pachyderm/" + tag + "/src/internal/jsonschema/CreatePipelineRequest.schema.json"
	}
	base := l.baseURL()
	base.Path = "/jsonschema/CreatePipelineRequest.schema.json"
	return base.String()
}
