package pfs_test

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

func TestProjectNameValidation(t *testing.T) {
	var cases = map[string]bool{
		"prOject_0":          true,
		"project 1":          false,
		"projectsß":          false,
		"project/subproject": false, // this may change in the future
	}
	for name, expected := range cases {
		if (pfs.ValidateProjectName(name) == nil) != expected {
			t.Errorf("incorrect validation of project name %q (expected %v)", name, expected)
		}
	}
}
