package storage

import (
	"io"
	"mime/multipart"

	"github.com/pachyderm/pachyderm/src/btrfs"
)

type multipartPuller struct {
	r *multipart.Reader
}

func newMultipartPuller(reader *multipart.Reader) *multipartPuller {
	return &multipartPuller{reader}
}

func (m *multipartPuller) Pull(from string, cb btrfs.Pusher) error {
	for {
		part, err := m.r.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		err = cb.Push(part)
		if err != nil {
			return err
		}
	}
	return nil
}
