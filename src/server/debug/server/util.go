package server

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
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

func collectDebugFile(tw *tar.Writer, name string, cb func(io.Writer) error, prefix ...string) (retErr error) {
	if len(prefix) > 0 {
		name = join(prefix[0], name)
	}
	defer func() {
		if retErr != nil {
			retErr = writeErrorFile(tw, retErr, name)
		}
	}()
	return withTmpFile(func(f *os.File) error {
		if err := cb(f); err != nil {
			return err
		}
		return writeTarFile(tw, name, f)
	})
}

func writeErrorFile(tw *tar.Writer, err error, prefix ...string) error {
	file := "error"
	if len(prefix) > 0 {
		file = join(prefix[0], file)
	}
	return withTmpFile(func(f *os.File) error {
		if _, err := io.Copy(f, strings.NewReader(err.Error()+"\n")); err != nil {
			return err
		}
		return writeTarFile(tw, file, f)
	})
}

func withTmpFile(cb func(*os.File) error) (retErr error) {
	if err := os.MkdirAll(os.TempDir(), 0700); err != nil {
		return err
	}
	f, err := ioutil.TempFile(os.TempDir(), "pachyderm_debug")
	if err != nil {
		return err
	}
	defer func() {
		if err := os.Remove(f.Name()); retErr == nil {
			retErr = err
		}
		if err := f.Close(); retErr == nil {
			retErr = err
		}
	}()
	return cb(f)
}

func writeTarFile(tw *tar.Writer, name string, f *os.File) error {
	fi, err := os.Stat(f.Name())
	if err != nil {
		return err
	}
	hdr := &tar.Header{
		Name: join(name),
		Size: fi.Size(),
		Mode: 0777,
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		return err
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	_, err = io.Copy(tw, f)
	return err
}

// TODO: May want to add some buffering in case we error midstream.
func collectDebugStream(tw *tar.Writer, r io.Reader, prefix ...string) (retErr error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return err
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
			return err
		}
		if len(prefix) > 0 {
			hdr.Name = join(prefix[0], hdr.Name)
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		_, err = io.Copy(tw, tr)
		if err != nil {
			return err
		}
	}
}
