package server

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"go.uber.org/multierr"
)

func join(names ...string) string {
	return strings.TrimPrefix(path.Join(names...), "/")
}

func withDebugWriter(w io.Writer, cb func(*tar.Writer) error) (retErr error) {
	gw := gzip.NewWriter(w)
	defer errors.Close(&retErr, gw, "close gzip writer")
	tw := tar.NewWriter(gw)
	defer errors.Close(&retErr, tw, "close tar writer")
	return cb(tw)
}

func collectDebugFile(dir string, name string, cb func(io.Writer) error) (retErr error) {
	if err := os.MkdirAll(dir, os.ModeDir); err != nil && !os.IsExist(err) {
		return errors.Wrapf(err, "mkdir %q", dir)
	}
	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		return errors.Wrapf(err, "create file %q", path)
	}
	defer func() {
		closeErr := errors.Wrapf(f.Close(), "close file %q", path)
		retErr = multierr.Append(retErr, closeErr)
	}()
	w := bufio.NewWriter(f)
	if err := cb(w); err != nil {
		return errors.Wrapf(err, "write to file %q", path)
	}
	return errors.Wrapf(w.Flush(), "flush file %q", path)
}

func writeErrorFile(dir string, err error) error {
	return errors.Wrapf(
		os.WriteFile(filepath.Join(dir, "error.txt"), []byte(err.Error()+"\n"), 0666),
		"failed to upload error file %q with message %q",
		filepath.Join(dir, "error.txt"),
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

func validatePipeline(p *pps.Pipeline) error {
	if p == nil {
		return errors.Errorf("nil pipeline")
	}
	if p.Name == "" {
		return errors.Errorf("empty pipeline name")
	}
	if err := validateProject(p.Project); err != nil {
		return errors.Wrapf(err, "invalid project in pipeline %q", p.Name)
	}
	return nil
}

func validatePipelineInfo(pi *pps.PipelineInfo) error {
	if pi == nil {
		return errors.Errorf("nil pipeline info")
	}
	if err := validatePipeline(pi.Pipeline); err != nil {
		return errors.Wrap(err, "invalid pipeline:")
	}
	return nil
}
