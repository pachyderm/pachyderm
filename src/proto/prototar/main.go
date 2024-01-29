package main

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/spf13/cobra"
	"github.com/zeebo/xxh3"
	"go.uber.org/zap"
)

// nameAndTarWriter records the filenames of files written to the archive.
type nameAndTarWriter struct {
	tar.Writer
	prefix string
	names  map[string]struct{}
}

func (w *nameAndTarWriter) WriteHeader(h *tar.Header) error {
	if err := w.Writer.WriteHeader(h); err != nil {
		return errors.Wrap(err, "write underlying header")
	}
	if w.names == nil {
		w.names = make(map[string]struct{})
	}
	w.names[path.Join(w.prefix, h.Name)] = struct{}{}
	return nil
}

func addToArchive(ctx context.Context, tw *nameAndTarWriter, f fs.FS) error {
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

type createRequest struct {
	tar, forgotten io.Writer
	root           fs.FS
	dirs           []string
}

func create(ctx context.Context, req *createRequest) (retErr error) {
	tw := &nameAndTarWriter{
		Writer: *tar.NewWriter(req.tar),
	}
	defer errors.Close(&retErr, tw, "close tar")
	for _, dir := range req.dirs {
		f, err := fs.Sub(req.root, dir)
		if err != nil {
			return errors.Wrapf(err, "root.Sub(%v)", dir)
		}
		tw.prefix = dir
		if err := addToArchive(ctx, tw, f); err != nil {
			return errors.Wrapf(err, "add dir %v", dir)
		}
	}
	tw.prefix = ""
	// Look in the out/ directory and report any files that didn't make it into the tar.
	if err := fs.WalkDir(req.root, "out", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return errors.Wrap(err, "WalkFn called with error")
		}
		if d.IsDir() {
			return nil
		}
		if _, archived := tw.names[path]; !archived {
			log.Debug(ctx, "forgotten file", zap.String("path", path))
			if _, err := fmt.Fprintf(req.forgotten, "%s\n", path); err != nil {
				return errors.Wrap(err, "write forgotten file")
			}
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "find forgotten files")
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

type applicationReport struct {
	Added, Modified, Deleted, Unchanged []string
}

// Tests that care will allocate this.  For real-world usage, log messages are sufficient.
var applyReport *applicationReport

func applyOne(ctx context.Context, h *tar.Header, r io.Reader) (retErr error) {
	currentHash, err := check(h.Name) // quick hash to see if the file is new
	if err != nil {
		return errors.Wrap(err, "check existing file")
	}
	if dir, _ := path.Split(h.Name); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return errors.Wrapf(err, "create output dir %v", dir)
		}
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
		if applyReport != nil {
			applyReport.Added = append(applyReport.Added, h.Name)
		}
		log.Info(ctx, "created new file", zap.String("name", h.Name))
	case newHash:
		if applyReport != nil {
			applyReport.Unchanged = append(applyReport.Unchanged, h.Name)
		}
		log.Debug(ctx, "unchanged file", zap.String("name", h.Name))
	default:
		if applyReport != nil {
			applyReport.Modified = append(applyReport.Modified, h.Name)
		}
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
		Use:   "create <output.tar> <forgotten-files.txt> <input...>",
		Short: "Archive the given directories.",
		Long:  "Archive the given directories.  Files are added to the archive relative to the provided inputs.  A 'forgotten files' report is written to forgotten-files.tar for manual inspection.",
		Args:  cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (retErr error) {
			tarPath := args[0]
			forgottenPath := args[1]

			pwd, err := os.Getwd()
			if err != nil {
				return errors.Wrap(err, "determine working directory")
			}
			req := &createRequest{
				root: os.DirFS(pwd),
				dirs: args[2:],
			}
			if req.tar, err = os.Create(tarPath); err != nil {
				return errors.Wrap(err, "create output tar")
			}
			defer errors.Close(&retErr, req.tar.(io.Closer), "close output tar")
			if req.forgotten, err = os.Create(forgottenPath); err != nil {
				return errors.Wrap(err, "create forgotten files file")
			}
			defer errors.Close(&retErr, req.forgotten.(io.Closer), "close forgotten files file")

			if err := create(cmd.Context(), req); err != nil {
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
