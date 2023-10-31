package fuse

import (
	"path/filepath"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/common"
)

const Separator = string(filepath.Separator)

// Adds a leading slash and a trailing slash to the path if they don't already exist
func standardizeSlashes(path string) string {
	addTrailingSlash := func(p string) string {
		if len(p) == 0 {
			return Separator
		}
		if p[len(p)-1] != filepath.Separator {
			return p + Separator
		}
		return p
	}
	addLeadingSlash := func(p string) string {
		if len(p) == 0 {
			return Separator
		}
		if p[0] != filepath.Separator {
			return Separator + p
		}
		return p
	}
	return addLeadingSlash(addTrailingSlash(path))
}

// Returns true if path1 is at least a grandparent of path2
// along with the intermediate path between path1 and path2
// (excludes last path component of path2)
func isGrandparentOf(path1, path2 string) (bool, string) {
	path1 = standardizeSlashes(path1)
	path2 = standardizeSlashes(path2)
	relative, err := filepath.Rel(path1, path2)
	if err != nil {
		return false, ""
	}
	parts := strings.Split(relative, string(filepath.Separator))
	if len(parts) < 2 || parts[0] == ".." {
		return false, ""
	}
	intermediate := filepath.Join(parts[:len(parts)-1]...)
	return true, standardizeSlashes(intermediate)
}

// Returns true if path1 is parent of path2
func isParentOf(path1, path2 string) bool {
	path1 = standardizeSlashes(path1)
	path2 = strings.TrimSuffix(standardizeSlashes(path2), Separator) // filepath.Dir doesn't work as expected if path2 ends with a slash
	parentDir := filepath.Dir(path2)
	return path1 == standardizeSlashes(parentDir)
}

// Returns true if path1 is a descendant of or equal to path2
func isDescendantOf(path1, path2 string) bool {
	path1 = standardizeSlashes(path1)
	path2 = standardizeSlashes(path2)
	return strings.HasPrefix(path1, path2)
}

// Does a DFS of the input tree, calling f on each node after visiting its children
func visitInput(input *pps.Input, level int, f func(*pps.Input, int) error) error {
	var source []*pps.Input
	switch {
	case input == nil:
		return errors.Errorf("spouts not supported") // Spouts may have nil input
	case input.Cross != nil:
		source = input.Cross
	case input.Join != nil:
		source = input.Join
	case input.Group != nil:
		source = input.Group
	case input.Union != nil:
		source = input.Union
	}
	for _, input := range source {
		if err := visitInput(input, level+1, f); err != nil {
			return err
		}
	}
	return f(input, level)
}

// Does the cartesian product among multiple datum slices
func crossDatums(datums [][]*pps.DatumInfo) []*pps.DatumInfo {
	if len(datums) == 0 {
		return nil
	}
	ret := datums[0]
	for _, slice := range datums[1:] {
		var temp []*pps.DatumInfo
		for _, d1 := range ret {
			for _, d2 := range slice {
				temp = append(temp, &pps.DatumInfo{
					Data: append(d1.Data, d2.Data...),
				})
			}
		}
		ret = temp
	}
	return ret
}

func copyMap(datumInputsToMounts map[string][]string) map[string][]string {
	datumInputsToMountsCopy := map[string][]string{}
	for k, v := range datumInputsToMounts {
		datumInputsToMountsCopy[k] = v
	}
	return datumInputsToMountsCopy
}

// Copied from PPS api_server.go
func validateNames(names map[string]bool, input *pps.Input) error {
	switch {
	case input == nil:
		return nil // spouts can have nil input
	case input.Pfs != nil:
		if err := validateName(input.Pfs.Name); err != nil {
			return err
		}
		if names[input.Pfs.Name] {
			return errors.Errorf(`name "%s" was used more than once`, input.Pfs.Name)
		}
		names[input.Pfs.Name] = true
	case input.Cron != nil:
		if err := validateName(input.Cron.Name); err != nil {
			return err
		}
		if names[input.Cron.Name] {
			return errors.Errorf(`name "%s" was used more than once`, input.Cron.Name)
		}
		names[input.Cron.Name] = true
	case input.Union != nil:
		for _, input := range input.Union {
			namesCopy := make(map[string]bool)
			merge(names, namesCopy)
			if err := validateNames(namesCopy, input); err != nil {
				return err
			}
			// we defer this because subinputs of a union input are allowed to
			// have conflicting names but other inputs that are, for example,
			// crossed with this union cannot conflict with any of the names it
			// might present
			defer merge(namesCopy, names)
		}
	case input.Cross != nil:
		for _, input := range input.Cross {
			if err := validateNames(names, input); err != nil {
				return err
			}
		}
	case input.Join != nil:
		for _, input := range input.Join {
			if err := validateNames(names, input); err != nil {
				return err
			}
		}
	case input.Group != nil:
		for _, input := range input.Group {
			if err := validateNames(names, input); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateName(name string) error {
	if name == "" {
		return errors.Errorf("input must specify a name")
	}
	switch name {
	case common.OutputPrefix, common.EnvFileName:
		return errors.Errorf("input cannot be named %v", name)
	}
	return nil
}

func merge(from, to map[string]bool) {
	for s := range from {
		to[s] = true
	}
}
