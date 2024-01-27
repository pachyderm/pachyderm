package main

import (
	"archive/tar"
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/spf13/cobra"
	"github.com/zeebo/xxh3"
	"go.uber.org/zap"
)

func addToArchive(ctx context.Context, tw *tar.Writer, f fs.FS) error {
	// TODO(jrockway): For this to be hermetic, WalkDir needs to always visit files in the same
	// order.  This is a detail of the underlying filesystem, probably.
	if err := fs.WalkDir(f, ".", func(path string, d fs.DirEntry, err error) (retErr error) {
		if err != nil {
			return errors.Wrap(err, "WalkFn called with error")
		}
		if d.IsDir() {
			return nil
		}
		log.Debug(ctx, "adding file", zap.String("path", path), zap.Any("entry", d))
		fh, err := f.Open(path)
		if err != nil {
			return errors.Wrap(err, "open source")
		}
		defer errors.Close(&retErr, fh, "close source")
		info, err := fh.Stat()
		if err != nil {
			return errors.Wrap(err, "stat source")
		}
		if err := tw.WriteHeader(&tar.Header{
			Name:       path,
			Size:       info.Size(),
			Uid:        0,
			Uname:      "",
			Gid:        0,
			Gname:      "",
			ModTime:    time.Unix(0, 0),
			AccessTime: time.Unix(0, 0),
			ChangeTime: time.Unix(0, 0),
			Mode:       0o600,
		}); err != nil {
			return errors.Wrapf(err, "write header for %v", path)
		}
		if n, err := io.Copy(tw, fh); err != nil {
			return errors.Wrapf(err, "copy %v into tar (%v bytes)", path, n)
		}
		return nil
	}); err != nil {
		return errors.Wrapf(err, "walk %v", f)
	}
	return nil
}

func create(ctx context.Context, w io.Writer, srcs ...fs.FS) (retErr error) {
	tw := tar.NewWriter(w)
	defer errors.Close(&retErr, tw, "close tar")
	for i, src := range srcs {
		if err := addToArchive(ctx, tw, src); err != nil {
			return errors.Wrapf(err, "add fs %v (%v)", i, src)
		}
	}
	return nil
}

func check(name string) (result xxh3.Uint128, retErr error) {
	fh, err := os.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return result, errors.Wrapf(err, "open current version of %v", name)
	}
	defer errors.Close(&retErr, fh, "close current version of %v", name)
	h := xxh3.New()
	if _, err := io.Copy(h, fh); err != nil {
		return result, errors.Wrapf(err, "write current version of %v into hash", name)
	}
	return h.Sum128(), nil
}

func applyOne(ctx context.Context, h *tar.Header, r io.Reader) (retErr error) {
	currentHash, err := check(h.Name) // quick hash to see if the file is new
	if err != nil {
		return errors.Wrap(err, "check existing file")
	}
	dir, _ := filepath.Split(h.Name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return errors.Wrapf(err, "create output dir %v", dir)
	}
	hash := xxh3.New()
	dst, err := os.OpenFile(h.Name, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return errors.Wrap(err, "open destination")
	}
	defer errors.Close(&retErr, dst, "close destination")

	w := io.MultiWriter(hash, dst)
	if _, err := io.Copy(w, r); err != nil {
		return errors.Wrap(err, "copy file")
	}
	newHash := hash.Sum128()
	switch currentHash {
	case xxh3.Uint128{}:
		log.Info(ctx, "created new file", zap.String("name", h.Name))
	case newHash:
		log.Debug(ctx, "unchanged file", zap.String("name", h.Name))
	default:
		log.Info(ctx, "updated file", zap.String("name", h.Name))
	}
	return nil
}

func apply(ctx context.Context, r io.Reader) error {
	tr := tar.NewReader(r)
	for {
		h, err := tr.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
		}
		if err := applyOne(ctx, h, tr); err != nil {
			return errors.Wrapf(err, "extract one file %v", h.Name)
		}
	}
}

func test(r io.Reader) (bool, error) {
	return false, nil
}

var ErrExit1 = errors.New("exit status 1")

func main() {
	report := log.InitBatchLogger("")
	ctx, c := pctx.Interactive()
	defer c()
	root := &cobra.Command{
		Use:           "prototar",
		Short:         "Archive, un-archive, and test proto bundles.",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// It would be nice to be verbose by default, but Bazel fails genrules if
			// they produce too much stderr output.  So this is the escape hatch;
			// disable sandboxing and set this environment variable to see what's going
			// on when needed.
			if x := os.Getenv("PACHYDERM_TOOL_VERBOSE"); x != "" {
				log.SetLevel(log.DebugLevel)
			}
		},
	}
	root.AddCommand([]*cobra.Command{{
		Use:   "create <output.tar> <input...>",
		Short: "Archive the given directories.",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := args[0]
			var srcs []fs.FS
			for _, src := range args[1:] {
				srcs = append(srcs, os.DirFS(src))
			}
			ow, err := os.Create(out)
			if err != nil {
				return errors.Wrap(err, "create output")
			}
			if err := create(cmd.Context(), ow, srcs...); err != nil {
				return errors.Wrap(err, "create archive")
			}
			return nil
		},
	}, {
		Use:   "apply <input.tar>",
		Short: "Extract the archive into the current working directory.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			in := args[0]
			fh, err := os.Open(in)
			if err != nil {
				return errors.Wrap(err, "open input")
			}
			if err := apply(cmd.Context(), fh); err != nil {
				return errors.Wrap(err, "apply archive")
			}
			return nil
		},
	}, {
		Use:   "test <input.tar>",
		Short: "Test that input.tar matches the content of the current working directory.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			in := args[0]
			fh, err := os.Open(in)
			if err != nil {
				return errors.Wrap(err, "open input")
			}
			ok, err := test(fh)
			if err != nil {
				return errors.Wrap(err, "test archive")
			}
			if !ok {
				return ErrExit1
			}
			return nil
		},
	}}...)

	if err := root.ExecuteContext(ctx); err != nil {
		if errors.Is(err, ErrExit1) {
			report(nil)
			os.Exit(1)
		}
		log.Error(ctx, "job failed", zap.Error(err))
		report(err)
		os.Exit(1)
	}
	report(nil)
	os.Exit(0)
}
