package server

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/fsutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
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

func collectDebugFile(tw *tar.Writer, name, ext string, cb func(io.Writer) error, prefix ...string) (retErr error) {
	if len(prefix) > 0 {
		name = join(prefix[0], name)
	}
	defer func() {
		if retErr != nil {
			retErr = writeErrorFile(tw, retErr, name)
		}
	}()
	return fsutil.WithTmpFile("pachyderm_debug", func(f *os.File) error {
		if err := cb(f); err != nil {
			return err
		}
		fullName := name
		if ext != "" {
			fullName += "." + ext
		}
		return writeTarFile(tw, fullName, f)
	})
}

func writeErrorFile(tw *tar.Writer, err error, prefix ...string) error {
	file := "error.txt"
	if len(prefix) > 0 {
		file = join(prefix[0], file)
	}
	return fsutil.WithTmpFile("pachyderm_debug", func(f *os.File) error {
		if _, err := io.Copy(f, strings.NewReader(err.Error()+"\n")); err != nil {
			return errors.EnsureStack(err)
		}
		return writeTarFile(tw, file, f)
	})
}

func writeTarFile(tw *tar.Writer, name string, f *os.File) error {
	fi, err := os.Stat(f.Name())
	if err != nil {
		return errors.EnsureStack(err)
	}
	hdr := &tar.Header{
		Name: join(name),
		Size: fi.Size(),
		Mode: 0777,
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		return errors.EnsureStack(err)
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return errors.EnsureStack(err)
	}
	_, err = io.Copy(tw, f)
	return errors.EnsureStack(err)
}

func collectDebugStream(tw *tar.Writer, r io.Reader, prefix ...string) (retErr error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return errors.EnsureStack(err)
	}
	defer errors.Close(&retErr, gr, "close gzip reader")
	tr := tar.NewReader(gr)
	return copyTar(tw, tr, prefix...)
}

func copyTar(tw *tar.Writer, tr *tar.Reader, prefix ...string) error {
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return errors.EnsureStack(err)
		}
		if len(prefix) > 0 {
			hdr.Name = join(prefix[0], hdr.Name)
		}
		if err := fsutil.WithTmpFile("pachyderm_debug_copy_tar", func(f *os.File) error {
			_, err = io.Copy(f, tr)
			if err != nil {
				return errors.EnsureStack(err)
			}
			_, err = f.Seek(0, 0)
			if err != nil {
				return errors.EnsureStack(err)
			}
			if err := tw.WriteHeader(hdr); err != nil {
				return errors.EnsureStack(err)
			}
			_, err = io.Copy(tw, f)
			return errors.EnsureStack(err)
		}); err != nil {
			return err
		}
	}
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

func validateRepoInfo(pi *pfs.RepoInfo) error {
	if pi == nil {
		return errors.Errorf("nil repo info")
	}
	if err := validateRepo(pi.Repo); err != nil {
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
