//go:build linux

package proc

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/prometheus/procfs"
)

func TestGetProcessStats(t *testing.T) {
	fs, err := procfs.NewDefaultFS()
	if err != nil {
		t.Fatal(err)
	}
	got, err := getProcessStats(fs, 0)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(got, &ProcessStats{}); diff == "" {
		t.Error("got no stats")
	}
}

func TestGetSystemStats(t *testing.T) {
	fs, err := procfs.NewDefaultFS()
	if err != nil {
		t.Fatal(err)
	}
	got, err := getSystemStats(fs)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(got, &SystemStats{}); diff == "" {
		t.Error("got no stats")
	}
}
