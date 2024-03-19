package pachd_test

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/pachd"
)

func TestNewTestPachd(t *testing.T) {
	pachd.NewTestPachd(t)
}

func TestNewTestPachd_underscore(t *testing.T) {
	pachd.NewTestPachd(t)
}
