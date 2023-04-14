package archiveserver

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"net/url"
	"strings"

	"github.com/go-git/go-git/v5/utils/binary"
	"github.com/klauspost/compress/zstd"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

// ArchiveFormat is the file format to produce.
type ArchiveFormat string

const (
	ArchiveFormatZip ArchiveFormat = "zip" // A ZIP file.
)

func (f ArchiveFormat) ContentType() string {
	//exhaustive:enforce
	switch f {
	case ArchiveFormatZip:
		return "application/zip"
	}
	panic("unknown archive format")
}

// An ArchiveRequest is a valid archive download URL.
type ArchiveRequest struct {
	rawFiles string        // This is the compressed+base64'd list of files to get.
	Format   ArchiveFormat // The desired format of the archive.
}

// ArchiveFromURL parses a URL into an ArchiveRequest.
func ArchiveFromURL(u *url.URL) (*ArchiveRequest, error) {
	if u == nil {
		return nil, errors.New("nil URL")
	}
	parts := strings.Split(u.Path, "/")
	if len(parts) != 3 {
		return nil, errors.Errorf("invalid download path; expected /download/<spec>, but got %v path parts", len(parts))
	}
	if got, want := parts[0], ""; got != want {
		return nil, errors.New("expected leading /")
	}
	if got, want := parts[1], "download"; got != want {
		return nil, errors.New("expected download/ as the first part of the path")
	}
	rawFilename := parts[2]
	fileParts := strings.SplitN(rawFilename, ".", 2)
	if len(fileParts) != 2 {
		return nil, errors.New("no extension on provided archive filename")
	}
	rawFormat := fileParts[1]
	switch ArchiveFormat(rawFormat) {
	case ArchiveFormatZip:
		return &ArchiveRequest{
			rawFiles: fileParts[0],
			Format:   ArchiveFormatZip,
		}, nil
	}
	return nil, errors.Errorf("unknown archive format %v", rawFormat)
}

// ForEachPath calls the callback with each requested file.
func (req *ArchiveRequest) ForEachPath(cb func(path string) error) error {
	base := strings.NewReader(req.rawFiles)
	b64 := base64.NewDecoder(base64.RawURLEncoding, base)

	var version uint8
	if err := binary.Read(b64, &version); err != nil {
		return errors.Wrap(err, "read version")
	}
	if got, want := version, uint8(0x01); got != want {
		return errors.Errorf("unknown version; got 0x%x want 0x%x", got, want)
	}
	// This is a version 1 format URL.
	dcmp, err := zstd.NewReader(b64)
	if err != nil {
		return errors.Wrap(err, "zstd.NewReader")
	}
	s := bufio.NewScanner(dcmp)
	s.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := bytes.IndexByte(data, 0x00); i >= 0 {
			return i + 1, data[0:i], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		// Request more data.
		return 0, nil, nil
	})
	for s.Scan() {
		path := s.Text()
		if err := cb(path); err != nil {
			return errors.Wrap(err, "path callback")
		}
	}
	if err := s.Err(); err != nil {
		return errors.Wrap(err, "scan")
	}
	return nil
}
