package tarutil

import (
	"bytes"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/tar"
)

type File interface {
	Header() (*tar.Header, error)
	Content(io.Writer) error
}

type file struct {
	hdr  *tar.Header
	data []byte
}

func NewFile(name string, data []byte) File {
	return newFile(NewHeader(name, int64(len(data))), data)
}

func newFile(hdr *tar.Header, data []byte) File {
	return &file{
		hdr:  hdr,
		data: data,
	}
}

func NewHeader(name string, size int64) *tar.Header {
	return &tar.Header{
		Name: name,
		Size: size,
	}
}

func (f *file) Header() (*tar.Header, error) {
	return f.hdr, nil
}

func (f *file) Content(w io.Writer) error {
	_, err := w.Write(f.data)
	return err
}

type streamFile struct {
	hdr *tar.Header
	r   io.Reader
}

func NewStreamFile(name string, size int64, r io.Reader) File {
	return newStreamFile(NewHeader(name, size), r)
}

func newStreamFile(hdr *tar.Header, r io.Reader) File {
	return &streamFile{
		hdr: hdr,
		r:   r,
	}
}

func (sf *streamFile) Header() (*tar.Header, error) {
	return sf.hdr, nil
}

func (sf *streamFile) Content(w io.Writer) error {
	_, err := io.Copy(w, sf.r)
	return err
}

func WriteFile(tw *tar.Writer, file File) error {
	hdr, err := file.Header()
	if err != nil {
		return err
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	if err := file.Content(tw); err != nil {
		return err
	}
	return tw.Flush()
}

func WithWriter(w io.Writer, cb func(*tar.Writer) error) (retErr error) {
	tw := tar.NewWriter(w)
	defer func() {
		if retErr == nil {
			retErr = tw.Close()
		}
	}()
	return cb(tw)
}

func Iterate(r io.Reader, cb func(File) error, stream ...bool) error {
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		if len(stream) > 0 && stream[0] {
			if err := cb(newStreamFile(hdr, tr)); err != nil {
				return err
			}
			continue
		}
		buf := &bytes.Buffer{}
		_, err = io.Copy(buf, tr)
		if err != nil {
			return err
		}
		if err := cb(newFile(hdr, buf.Bytes())); err != nil {
			return err
		}
	}
}

func Equal(file1, file2 File, full ...bool) (bool, error) {
	buf1, buf2 := &bytes.Buffer{}, &bytes.Buffer{}
	if len(full) > 0 && full[0] {
		// Check the serialized tar entries.
		tw1, tw2 := tar.NewWriter(buf1), tar.NewWriter(buf2)
		if err := WriteFile(tw1, file1); err != nil {
			return false, err
		}
		if err := WriteFile(tw2, file2); err != nil {
			return false, err
		}
		return bytes.Equal(buf1.Bytes(), buf2.Bytes()), nil
	}
	// Check the header name and content.
	hdr1, err := file1.Header()
	if err != nil {
		return false, err
	}
	hdr2, err := file2.Header()
	if err != nil {
		return false, err
	}
	if hdr1.Name != hdr2.Name {
		return false, nil
	}
	if err := file1.Content(buf1); err != nil {
		return false, err
	}
	if err := file2.Content(buf2); err != nil {
		return false, err
	}
	return bytes.Equal(buf1.Bytes(), buf2.Bytes()), nil
}

func TarToLocal(storageRoot string, r io.Reader, cb ...func(*tar.Header) error) error {
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		if len(cb) > 0 {
			if err := cb[0](hdr); err != nil {
				return err
			}
		}
		// TODO: Use the tar header metadata.
		fullPath := path.Join(storageRoot, hdr.Name)
		if hdr.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(fullPath, 0700); err != nil {
				return err
			}
			continue
		}
		if err := writeFile(fullPath, tr); err != nil {
			return err
		}
	}
}

func writeFile(filePath string, r io.Reader) (retErr error) {
	if err := os.MkdirAll(path.Dir(filePath), 0700); err != nil {
		return err
	}
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); retErr == nil {
			retErr = err
		}
	}()
	_, err = io.Copy(f, r)
	return err
}

func LocalToTar(storageRoot string, w io.Writer, cb ...func(*tar.Header) error) error {
	return WithWriter(w, func(tw *tar.Writer) error {
		return filepath.Walk(storageRoot, func(file string, fi os.FileInfo, err error) (retErr error) {
			if err != nil {
				return err
			}
			if file == storageRoot {
				return nil
			}
			// TODO: link name?
			hdr, err := tar.FileInfoHeader(fi, "")
			if err != nil {
				return err
			}
			hdr.Name, err = filepath.Rel(storageRoot, file)
			if err != nil {
				return err
			}
			// TODO: Remove when path cleaning is in.
			hdr.Name = filepath.Join("/", hdr.Name)
			if len(cb) > 0 {
				if err := cb[0](hdr); err != nil {
					return err
				}
			}
			if err := tw.WriteHeader(hdr); err != nil {
				return err
			}
			if hdr.Typeflag == tar.TypeDir {
				return nil
			}
			f, err := os.Open(file)
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); retErr == nil {
					retErr = err
				}
			}()
			_, err = io.Copy(tw, f)
			return err
		})
	})
}
