package storage

import (
	"io"
	"mime/multipart"
	"net/textproto"
)

type multipartPusher struct {
	writer *multipart.Writer
}

func newMultipartPusher(writer *multipart.Writer) *multipartPusher {
	return &multipartPusher{writer}
}

func (m *multipartPusher) From() (string, error) {
	return "", nil
}

func (m *multipartPusher) Push(diff io.Reader) error {
	mimeHeader := make(textproto.MIMEHeader)
	writer, err := m.writer.CreatePart(mimeHeader)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, diff)
	return err
}
