package migration

import (
	"fmt"
	"strings"
)

// migrationFunc is a function that migrates Pachyderm's internal state from
// one version to another.
type migrationFunc func(etcdAddress, pfsPrefix, ppsPrefix string) error

var migrationRoutines map[string]map[string]migrationFunc

func init() {
	migrationRoutines = make(map[string]map[string]migrationFunc)

	// Register 1.4.* -> 1.5.0
	for _, version := range allPatchVersions("1.4") {
		migrationRoutines[version] = make(map[string]migrationFunc)
		migrationRoutines[version]["1.5.0"] = oneFourToOneFive
	}
}

// Given a macro.micro version number like 1.4, returns a slice of version
// numbers that include all possible patch numbers such as 1.4.0, 1.4.1, etc.
func allPatchVersions(version string) []string {
	var res []string
	for _, patch := range strings.Split("0 1 2 3 4 5 6 7 8 9 *", " ") {
		res = append(res, fmt.Sprintf("%s.%s", version, patch))
	}
	return res
}

// Run executes a migration routine that migrates from one version to another.
func Run(etcdAddress, pfsPrefix, ppsPrefix, from, to string) error {
	routines := migrationRoutines[from]
	if routines == nil {
		return fmt.Errorf("unable to find a migration routine that migrates from version %v", from)
	}

	routine := routines[to]
	if routine == nil {
		return fmt.Errorf("unable to find a migration routine that migrates from version %v to version %v", from, to)
	}

	return routine(etcdAddress, pfsPrefix, ppsPrefix)
}
