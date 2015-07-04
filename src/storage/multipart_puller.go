package storage

import (
	"io"
	"mime/multipart"

	"github.com/pachyderm/pachyderm/src/btrfs"
)

type multipartPuller struct {
	reader *multipart.Reader
}

func newMultipartPuller(reader *multipart.Reader) *multipartPuller {
	return &multipartPuller{reader}
}

func (m *multipartPuller) Pull(from string, pusher btrfs.Pusher) error {
	for part, err := m.reader.NextPart(); err != io.EOF; part, err = m.reader.NextPart() {
		if err != nil {
			return err
		}
		if err = pusher.Push(part); err != nil {
			return err
		}
	}
	return nil
}
