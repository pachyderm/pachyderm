package main

import (
	"context"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"go.uber.org/zap"
)

var (
	src  = flag.String("src", "", "bazel target of binary to install")
	dst  = flag.String("dst", "", "if set, install to this directory instead of a good candidate from $PATH")
	name = flag.String("name", "", "what the binary will be called after installation")
)

func find(home string, path []string, binary string) (paths []string) {
	// Look for an existing binary first.
	for i, p := range path {
		if _, err := os.Stat(filepath.Join(p, binary)); err == nil {
			paths = append(paths, p)
			path[i] = ""
			break
		}
	}

	// Look for the best possible locations next, before we choose locations in $PATH order.
	// When running with Bazelisk, ~/.cache/bazelisk appears in $PATH first.  When using things
	// like NVM, common on the console team, node appears first.  When using Homebrew, it
	// usually appears before user-set paths.  We would like to avoid those because while they
	// will work, it's very ugly and confusing.
	//
	// Additionally, go has no way to get $GOBIN the same way that "go install" does, so we
	// can't just ask for that and prefer it.  $GOPATH is similar, but the default is easy to
	// calculate so we try it.
	for i, p := range path {
		if p == "" {
			continue
		}
		if x := filepath.Join(home, "bin"); p == x || p == x+"/" {
			paths = append(paths, x)
			path[i] = ""
			continue
		}
		if x := filepath.Join(os.Getenv("GOPATH"), "bin"); p == x || p == x+"/" {
			paths = append(paths, x)
			path[i] = ""
			continue
		}
		if x := filepath.Join(home, "go", "bin"); p == x || p == x+"/" {
			paths = append(paths, p)
			path[i] = ""
			continue
		}
	}
	// Now add things that start with $HOME first.
	for i, p := range path {
		if p == "" {
			continue
		}
		if strings.HasPrefix(p, home) {
			paths = append(paths, p)
			path[i] = ""
		}
	}
	// Then finish up with the rest.
	for _, p := range path {
		if p == "" {
			continue
		}
		paths = append(paths, p)
	}
	return
}

func install(ctx context.Context, dst string, src string) (retErr error) {
	log.Debug(ctx, "attempting to install", zap.String("path", dst))
	w, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return errors.Wrap(err, "open destination")
	}
	defer errors.Close(&retErr, w, "close destination")

	r, err := os.Open(src)
	if err != nil {
		return errors.Wrap(err, "open source")
	}
	defer errors.Close(&retErr, r, "close source")

	if _, err := io.Copy(w, r); err != nil {
		defer func() {
			if err := os.Remove(dst); err != nil {
				errors.JoinInto(&retErr, errors.Wrapf(err, "cleanup failed copy %v", dst))
			}
		}()
		return errors.Wrap(err, "copy binary")
	}
	log.Info(ctx, "tool installed", zap.String("path", dst))
	return nil
}

func main() {
	flag.Parse()

	log.InitBatchLogger("")
	ctx, c := pctx.Interactive()
	defer c()

	home, err := os.UserHomeDir()
	if err != nil {
		log.Info(ctx, "cannot determine homedir; not treating it specially", zap.Error(err))
	}

	paths := []string{*dst}
	if paths[0] == "" {
		paths = find(home, strings.Split(os.Getenv("PATH"), ":"), *name)
	}

	for i, p := range paths {
		dst := filepath.Join(p, *name)
		if err := install(ctx, dst, *src); err != nil {
			log.Info(ctx, "failed install; trying next candidate", zap.String("path", dst), zap.Error(err))
			paths[i] = ""
			continue
		}
		return
	}
	log.Exit(ctx, "no suitable installation destinations found; add a writable directory to your $PATH or supply -dst")
}
