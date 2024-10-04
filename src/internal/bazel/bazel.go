// Package bazel provides some utilities for working with Bazel.
package bazel

import (
	"maps"
	"path/filepath"
	"slices"
	"strings"

	"github.com/bazelbuild/rules_go/go/runfiles"
)

// LibraryPath returns a value for $LD_LIBRARY_PATH that contains the provided libraries.  libraries
// should be specified as rlocationpaths; in an x_defs clause, $(rlocationpath
// //bazel/package:libfoo.so).  If a library cannot be resolved, the unchanged value is used.
// environ is the existing environment, potentially with LD_LIBRARY_PATH already set.  And resolved
// libraries are prepended to the existing value.  If an empty string is returned, LD_LIBRARY_PATH
// can remain unset.
func LibraryPath(environ []string, libraries ...string) string {
	// Find the existing value of $LD_LIBRARY_PATH.
	var libraryPath string
	for _, e := range environ {
		if strings.HasPrefix(e, "LD_LIBRARY_PATH=") {
			libraryPath = e[len("LD_LIBRARY_PATH="):]
		}
		// It's expected that we use the LAST value if it's set multiple times, so we keep
		// going even when we've found LD_LIBRARY_PATH.
	}

	// Resolve to a unique set of paths.
	paths := make(map[string]struct{})
	for _, l := range libraries {
		if l == "" {
			continue
		}
		x, err := runfiles.Rlocation(l)
		if err != nil {
			x = l
		}
		if x != "" {
			paths[filepath.Dir(x)] = struct{}{}
		}
	}
	prefix := strings.Join(slices.Sorted(maps.Keys(paths)), ":")

	parts := make([]string, 0, 2)
	if prefix != "" {
		parts = append(parts, prefix)
	}
	if libraryPath != "" {
		parts = append(parts, libraryPath)
	}
	return strings.Join(parts, ":")
}
