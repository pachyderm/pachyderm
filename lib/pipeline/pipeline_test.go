package pipeline

import (
	"path"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/pachyderm/pfs/lib/btrfs"
)

func check(err error, t *testing.T) {
	if err != nil {
		debug.PrintStack()
		t.Fatal(err)
	}
}

// TestOutput tests a simple job that outputs some data.
func TestOutput(t *testing.T) {
	outRepo := "TestOuputRepo"
	btrfs.Init(outRepo)
	pipeline := NewPipeline("", outRepo, "", "master")
	pachfile := `
image ubuntu

run touch /out/foo
run touch /out/bar
`
	err := pipeline.RunPachFile(strings.NewReader(pachfile))
	check(err, t)

	exists, err := btrfs.FileExists(path.Join(outRepo, "0", "foo"))
	check(err, t)
	if exists != true {
		t.Fatal("File `foo` doesn't exist when it should.")
	}

	exists, err = btrfs.FileExists(path.Join(outRepo, "1", "bar"))
	check(err, t)
	if exists != true {
		t.Fatal("File `bar` doesn't exist when it should.")
	}
}
