package versionpb

import (
	fmt "fmt"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"golang.org/x/mod/semver"
)

// Canonical returns the version in canonical semver form, according to golang.org/x/mod/semver.
func (v *Version) Canonical() string {
	if v == nil {
		v = new(Version)
	}
	// vMAJOR[.MINOR[.PATCH[-PRERELEASE][+BUILD]]]
	b := new(strings.Builder)
	fmt.Fprintf(b, "v%d.%d.%d", v.Major, v.Minor, v.Micro)
	if additional := v.Additional; additional != "" {
		// We seem to build the - into the additional field, at least for nightlies.  Trim
		// it anyway.  We also seem to put + there in some build scripts, which isn't a
		// format semver can parse.  We map all these to - to fit their spec.
		additional = strings.TrimLeft(additional, "-.+")
		b.WriteByte('-')
		b.WriteString(additional)
	}
	return semver.Canonical(b.String())
}

// IsDevelopment returns true if the version is from a local unreleased build.
func (v *Version) IsDevelopment() bool {
	return v.Major == 0 && v.Minor == 0 && v.Micro == 0
}

var (
	ErrClientIsNil         = errors.New("client version is missing")
	ErrServerIsNil         = errors.New("server version is nil")
	ErrClientTooOld        = errors.New("client is outdated")
	ErrServerTooOld        = errors.New("server is outdated")
	ErrIncompatiblePreview = errors.New("client and server versions must match exactly")
)

// IsCompatible determines if two versions are compatible.  Versions are compatible if the versions
// are identical or one version is a development build, or if either the server or client are not
// prereleases, then the major and minor numbers are identical.
func IsCompatible(rawClient, rawServer *Version) error {
	if rawClient == nil {
		return ErrClientIsNil
	}
	if rawServer == nil {
		return ErrServerIsNil
	}
	if rawClient.IsDevelopment() || rawServer.IsDevelopment() {
		return nil
	}
	s, c := rawServer.Canonical(), rawClient.Canonical()
	if semver.Prerelease(s) != "" || semver.Prerelease(c) != "" {
		// If either the client or the server are a prerelease, then the versions must match
		// exactly.
		if semver.Compare(c, s) != 0 {
			return ErrIncompatiblePreview
		}
		return nil
	}
	switch semver.Compare(semver.MajorMinor(c), semver.MajorMinor(s)) {
	case -1:
		return ErrClientTooOld
	case 0:
		return nil
	case 1:
		return ErrServerTooOld
	}
	return errors.New("internal error: missing case in semver.Compare switch statement")
}
