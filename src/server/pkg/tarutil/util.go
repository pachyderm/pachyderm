package tarutil

import (
	"bytes"
	"io"

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

func Iterate(r io.Reader, cb func(File) error) error {
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
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

func Equal(file1, file2 File) (bool, error) {
	buf1, buf2 := &bytes.Buffer{}, &bytes.Buffer{}
	tw1, tw2 := tar.NewWriter(buf1), tar.NewWriter(buf2)
	if err := WriteFile(tw1, file1); err != nil {
		return false, err
	}
	if err := WriteFile(tw2, file2); err != nil {
		return false, err
	}
	if !bytes.Equal(buf1.Bytes(), buf2.Bytes()) {
		return false, nil
	}
	return true, nil
}
