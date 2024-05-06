// Binary compile-with-coverage takes a GOPATH created by the Bazel go_path rule, resolves all
// symlinks inside it by copying, and invokes the Go compiler.  This is because when the GOPATH is
// used as a runfile, it contains symlinks back into the original repository, but //go:embed can't
// resolve symlinks.
package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

var (
	gopath            = flag.String("gopath", "", "the gopath to build in")
	target            = flag.String("target", "", "the target to build")
	out               = flag.String("out", "", "where to put the output")
	gobin             = flag.String("go", "", "the path to the go (compiler) binary")
	appVersion        = flag.String("app_version", "", "the app version of the build")
	additionalVersion = flag.String("additional_version", "", "the additional version of the build")
)

func resolve(dir string, out string) error {
	if err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) (retErr error) {
		if err != nil {
			return errors.Wrapf(err, "%v called with error", path)
		}
		newpath := filepath.Join(out, path)
		if d.IsDir() {
			if err := os.MkdirAll(newpath, 0o755); err != nil {
				return errors.Wrap(err, "create directory in output")
			}
			return nil
		}
		src, err := os.OpenFile(path, os.O_RDONLY, 0)
		if err != nil {
			return errors.Wrap(err, "open source")
		}
		defer errors.Close(&retErr, src, "close src")
		dst, err := os.OpenFile(newpath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			return errors.Wrap(err, "open destination")
		}
		defer errors.Close(&retErr, dst, "close dst")
		if _, err := io.Copy(dst, src); err != nil {
			return errors.Wrapf(err, "copy %v to %v", target, path)
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "walk")
	}
	return nil
}

func main() {
	flag.Parse()
	if *gopath == "" {
		log.Fatal("gopath must be set")
	}
	if *target == "" {
		log.Fatal("target must be set")
	}
	if *out == "" {
		log.Fatal("out must be set")
	}
	if *gobin == "" {
		log.Fatal("go must be set")
	}

	resolvedGopath, err := os.MkdirTemp("", "pachyderm-resolved-gopath-")
	if err != nil {
		log.Fatalf("MkdirTemp for resolved $GOPATH: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(resolvedGopath); err != nil {
			log.Printf("failed to clean up resolved $GOPATH: %v", err)
		}
	}()
	if err := resolve(*gopath, resolvedGopath); err != nil {
		log.Fatalf("resolve symlinks in %v: %v", *gopath, err)
	}

	gocache, err := os.MkdirTemp("", "pachyderm-compile-with-coverage-")
	if err != nil {
		log.Fatalf("MkdirTemp for $GOCACHE: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(gocache); err != nil {
			log.Printf("failed to clean up $GOCACHE: %v", err)
		}
	}()
	ldflags := fmt.Sprintf("-X %s=%s -X %s=%s", "github.com/pachyderm/pachyderm/v2/src/version.AppVersion", *appVersion, "github.com/pachyderm/pachyderm/v2/src/version.AdditionalVersion", *additionalVersion)
	cmd := exec.Command(*gobin, "build", "-o", *out, "-cover", "-ldflags", ldflags, "-coverpkg=github.com/pachyderm/pachyderm/v2/...", "-covermode=atomic", *target)
	cmd.Env = []string{"GOPATH=" + filepath.Join(resolvedGopath, *gopath), "GO111MODULE=off", "GOCACHE=" + gocache}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("running go: %v", err)
	}
}
