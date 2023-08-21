package server

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

func withDebugWriter(w io.Writer, cb func(*tar.Writer) error) (retErr error) {
	gw := gzip.NewWriter(w)
	defer errors.Close(&retErr, gw, "close gzip writer")
	tw := tar.NewWriter(gw)
	defer errors.Close(&retErr, tw, "close tar writer")
	return cb(tw)
}

type DumpFS interface {
	Write(string, func(io.Writer) error) error
	WithPrefix(string) DumpFS
}

type dumpFS struct {
	hiddenDir string
	prefix    string
}

func NewDumpFS(mountPath string) *dumpFS {
	return &dumpFS{hiddenDir: mountPath}
}

func (dfs *dumpFS) Write(path string, cb func(io.Writer) error) error {
	fullPath := filepath.Join(dfs.hiddenDir, dfs.prefix, path)
	realDir := filepath.Dir(fullPath)
	if err := os.MkdirAll(realDir, 0744); err != nil && !os.IsExist(err) {
		return errors.Wrapf(err, "mkdir %q", realDir)
	}
	f, err := os.Create(fullPath)
	if err != nil {
		return errors.Wrapf(err, "create file %q", fullPath)
	}
	return func() (retErr error) {
		defer func() {
			closeErr := errors.Wrapf(f.Close(), "close file %q", fullPath)
			errors.JoinInto(&retErr, closeErr)
			if retErr != nil {
				rmErr := errors.Wrapf(os.Remove(fullPath), "cleanup file %q", fullPath)
				errors.JoinInto(&retErr, rmErr)
			}
		}()
		return cb(f)
	}()
}

func (dfs *dumpFS) WithPrefix(prefix string) DumpFS {
	return &dumpFS{hiddenDir: dfs.hiddenDir, prefix: prefix}
}

func writeErrorFile(dfs DumpFS, err error, prefix string) error {
	path := filepath.Join(prefix, "error.txt")
	return errors.Wrapf(
		dfs.Write(path, func(w io.Writer) error {
			_, err := w.Write([]byte(err.Error() + "\n"))
			return errors.Wrapf(err, "write error file %q", path)
		}),
		"failed to upload error file %q with message %q",
		path,
		err.Error(),
	)
}

func writeTarFile(tw *tar.Writer, dest string, src string, fi os.FileInfo) error {
	f, err := os.Open(src)
	if err != nil {
		return errors.Wrapf(err, "open file %q", src)
	}
	if err := tw.WriteHeader(&tar.Header{
		Name: dest,
		Size: fi.Size(),
		Mode: 0777,
	}); err != nil {
		return errors.Wrapf(err, "write header for file %q", dest)
	}
	_, err = io.Copy(tw, f)
	return errors.Wrapf(err, "write tar file %q", dest)

}

func validateProject(p *pfs.Project) error {
	if p == nil {
		return errors.Errorf("nil project")
	}
	if p.Name == "" {
		return errors.Errorf("empty project name")
	}
	return nil
}

func validateRepo(r *pfs.Repo) error {
	if r == nil {
		return errors.Errorf("nil repo")
	}
	if r.Name == "" {
		return errors.Errorf("empty repo name")
	}
	if err := validateProject(r.Project); err != nil {
		return errors.Wrapf(err, "invalid project in repo %q", r.Name)
	}
	return nil
}

func validateRepoInfo(ri *pfs.RepoInfo) error {
	if ri == nil {
		return errors.Errorf("nil repo info")
	}
	if err := validateRepo(ri.Repo); err != nil {
		return errors.Wrap(err, "invalid repo:")
	}
	return nil
}
