// Command dedupe-tar removes duplicate files from the tar provided on stdin, writing a
// zstd-compressed tar to stdout.  This is necessary for some container registries to accept the
// output of @rules_distroless#flatten.
//
// It starts with a list of files to not touch in common.txt.  This is specifically for dealing with
// Debian packages, which create their parent directories, but shouldn't.  It is not a big deal if
// it gets slightly out of sync, the important things never change (like /var and friends).
package main

import (
	"archive/tar"
	"bufio"
	"context"
	_ "embed"
	"io"
	"os"
	"strings"

	"github.com/klauspost/compress/zstd"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"go.uber.org/zap"
)

//go:embed "common.txt"
var common string

func fixName(x string) string {
	x = strings.TrimPrefix(x, "./")
	x = strings.TrimPrefix(x, "/")
	return x
}

func run(ctx context.Context, r io.Reader, w io.Writer) (retErr error) {
	tr := tar.NewReader(r)

	zw, err := zstd.NewWriter(w)
	if err != nil {
		return errors.Wrap(err, "new zstd writer")
	}
	defer errors.Close(&retErr, zw, "close zstd writer")
	tw := tar.NewWriter(zw)
	defer errors.Close(&retErr, tw, "close tar writer")

	// common.txt is a list of all directory names in the main distroless:static-nonroot image.
	// we don't want our layer to create any of those files.  (Debian typically installs
	// packages by extracting the tar with --keep-directory-symlink, but we can't "extract"
	// packages that way because we don't see the layer under ours until runtime.  So this
	// ensures that we skip those files.)
	got := map[string]struct{}{"": {}}
	s := bufio.NewScanner(strings.NewReader(common))
	for s.Scan() {
		got[fixName(s.Text())] = struct{}{}
	}
	if err := s.Err(); err != nil {
		return errors.Wrap(err, "scan common.txt")
	}

	for {
		h, err := tr.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return errors.Wrap(err, "read header")
		}
		orig, name := h.Name, fixName(h.Name)
		if _, ok := got[name]; ok {
			if _, err := io.Copy(io.Discard, tr); err != nil {
				return errors.Wrapf(err, "discard content of %v (%v)", name, orig)
			}
			continue
		}
		h.Name = name
		got[name] = struct{}{}
		h.Uname = "" // avoid contaminating output with host's uid->username mapping
		h.Gname = ""
		if err := tw.WriteHeader(h); err != nil {
			return errors.Wrapf(err, "write header for %v (%v)", name, orig)
		}
		if _, err := io.Copy(tw, tr); err != nil {
			return errors.Wrapf(err, "copy content of %v (%v)", name, orig)
		}
	}
	return nil
}

func main() {
	log.InitBatchLogger("")
	ctx, c := pctx.Interactive()
	defer c()

	if err := run(ctx, os.Stdin, os.Stdout); err != nil {
		log.Exit(ctx, "run failed", zap.Error(err))
	}
}
