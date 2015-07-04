package storage

import (
	"io"
	"mime/multipart"
	"net/textproto"
)

type multipartPusher struct {
	w *multipart.Writer
}

func newMultipartPusher(writer *multipart.Writer) *multipartPusher {
	return &multipartPusher{writer}
}

func (m *multipartPusher) From() (string, error) {
	return "", nil
}

func (m *multipartPusher) Push(diff io.Reader) error {
	h := make(textproto.MIMEHeader)
	w, err := m.w.CreatePart(h)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, diff)
	if err != nil {
		return err
	}
	return nil
}
