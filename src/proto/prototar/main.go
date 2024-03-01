package main

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"sort"
	"strings"
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
	var paths []string
	if err := fs.WalkDir(f, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return errors.Wrap(err, "WalkFn called with error")
		}
		if d.IsDir() {
			return nil
		}
		paths = append(paths, path)
		return nil
	}); err != nil {
		return errors.Wrapf(err, "walk %v", f)
	}
	sort.Strings(paths)
	add := func(path string) (retErr error) {
		log.Debug(ctx, "adding file", zap.String("path", path))
		fh, err := f.Open(path)
		if err != nil {
			return errors.Wrap(err, "open")
		}
		defer errors.Close(&retErr, fh, "close")
		info, err := fh.Stat()
		if err != nil {
			return errors.Wrap(err, "stat")
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
			return errors.Wrap(err, "write header")
		}
		if n, err := io.Copy(tw, fh); err != nil {
			return errors.Wrapf(err, "copy into tar (%v bytes)", n)
		}
		return nil
	}
	for _, path := range paths {
		if err := add(path); err != nil {
			return errors.Wrapf(err, "add %v", path)
		}
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

type ApplicationReport struct {
	Added, Modified, Deleted, Unchanged []string
}

func applyOne(ctx context.Context, h *tar.Header, r io.Reader, dryRun bool, applyReport *ApplicationReport) (retErr error) {
	currentHash, err := check(h.Name) // quick hash to see if the file is new
	if err != nil {
		return errors.Wrap(err, "check existing file")
	}
	if dir, _ := path.Split(h.Name); dir != "" && !dryRun {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return errors.Wrapf(err, "create output dir %v", dir)
		}
	}
	hasher := xxh3.New()
	var w io.Writer = hasher
	if !dryRun {
		dst, err := os.OpenFile(h.Name, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
		if err != nil {
			return errors.Wrap(err, "open destination")
		}
		defer errors.Close(&retErr, dst, "close destination")
		w = io.MultiWriter(hasher, dst)
	}
	if _, err := io.Copy(w, r); err != nil {
		return errors.Wrap(err, "copy file")
	}
	newHash := hasher.Sum128()

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

func apply(ctx context.Context, r io.Reader, dryRun bool) (*ApplicationReport, error) {
	report := new(ApplicationReport)
	jsonSchemas := map[string]struct{}{}

	// Find the current set of JSON schemas, so we can delete any files that aren't in the proto
	// bundle.
	if err := fs.WalkDir(os.DirFS("src/internal/jsonschema"), ".", func(file string, d fs.DirEntry, err error) error {
		if err != nil {
			return errors.Wrap(err, "WalkFn called with error")
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(file, ".schema.json") {
			jsonSchemas[path.Join("src/internal/jsonschema", file)] = struct{}{}
		}
		return nil
	}); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, errors.Wrap(err, "find json schemas")
		}
	}

	// Extract each file in the archive, updating the report with
	// added/modified/etc. information.
	if err := func() error {
		tr := tar.NewReader(r)
		for {
			h, err := tr.Next()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				return errors.Wrap(err, "read tar")
			}
			if err := applyOne(ctx, h, tr, dryRun, report); err != nil {
				return errors.Wrapf(err, "extract one file %v", h.Name)
			}
			delete(jsonSchemas, h.Name)
		}
	}(); err != nil {
		return nil, errors.Wrap(err, "extract archive")
	}

	// Finally, get rid of all JSON schemas that weren't extracted.
	for schema := range jsonSchemas {
		if !dryRun {
			if err := os.Remove(schema); err != nil {
				return nil, errors.Wrap(err, "remove stale json schema")
			}
		}
		log.Info(ctx, "removed stale json schema", zap.String("name", schema))
		report.Deleted = append(report.Deleted, schema)
	}

	return report, nil
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
			if _, err := apply(cmd.Context(), fh, false); err != nil {
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
			report, err := apply(cmd.Context(), fh, true)
			if err != nil {
				return errors.Wrap(err, "apply archive")
			}
			if len(report.Added) == 0 && len(report.Deleted) == 0 && len(report.Modified) == 0 {
				fmt.Printf("%d file(s) ok\n", len(report.Unchanged))
				return nil
			}
			fmt.Fprintf(os.Stderr, "\n*** It looks like you might need to regenerate the protos in your working copy.\n")
			return ErrExit1
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
