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
)

func join(names ...string) string {
	return strings.TrimPrefix(path.Join(names...), "/")
}

func withDebugWriter(w io.Writer, cb func(*tar.Writer) error) (retErr error) {
	gw := gzip.NewWriter(w)
	defer func() {
		if retErr == nil {
			retErr = gw.Close()
		}
	}()
	tw := tar.NewWriter(gw)
	defer func() {
		if retErr == nil {
			retErr = tw.Close()
		}
	}()
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

// TODO: May want to add some buffering in case we error midstream.
func collectDebugStream(tw *tar.Writer, r io.Reader, prefix ...string) (retErr error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return errors.EnsureStack(err)
	}
	defer func() {
		if err := gr.Close(); retErr == nil {
			retErr = err
		}
	}()
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
		if err := tw.WriteHeader(hdr); err != nil {
			return errors.EnsureStack(err)
		}
		_, err = io.Copy(tw, tr)
		if err != nil {
			return errors.EnsureStack(err)
		}
	}
}
