package archiveserver

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"io"
	"sort"

	"github.com/klauspost/compress/zstd"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

const (
	EncodingVersion1 = uint8(0x01)
)

// EncodeV1 generates a V1 archive URL string.
func EncodeV1(paths []string) (string, error) {
	buf := new(bytes.Buffer)
	b64 := base64.NewEncoder(base64.RawURLEncoding, buf)
	if _, err := b64.Write([]byte{EncodingVersion1}); err != nil { // V1
		return "", errors.Wrap(err, "write version")
	}
	cmp, err := zstd.NewWriter(b64, zstd.WithEncoderLevel(zstd.SpeedBestCompression))
	if err != nil {
		return "", errors.Wrap(err, "zstd.NewWriter")
	}
	sort.Strings(paths)
	for _, path := range paths {
		if _, err := cmp.Write([]byte(path)); err != nil {
			return "", errors.Wrapf(err, "write path %q", path)
		}
		if _, err := cmp.Write([]byte{0x00}); err != nil {
			return "", errors.Wrapf(err, "write terminator for path %q", path)
		}
	}
	if err := cmp.Close(); err != nil {
		return "", errors.Wrap(err, "close zstd")
	}
	if err := b64.Close(); err != nil {
		return "", errors.Wrap(err, "close base64")
	}
	return buf.String(), nil
}

// Decode begins the decoding process; all versions are base64url-encoded bytes prefixed with a
// uint8 version number.  Decode returns a reader of the bytes right after the version, and the
// version number.
func Decode(r io.Reader) (io.Reader, uint8, error) {
	b64 := base64.NewDecoder(base64.RawURLEncoding, r)

	var version uint8
	if err := binary.Read(b64, binary.BigEndian, &version); err != nil {
		return nil, 0, errors.Wrap(err, "read version")
	}
	return b64, version, nil
}

// DecodeV1 returns a bufio.Scanner that returns each encoded path.  The reader should already be
// read past the version number; Decode() above does exactly this.
func DecodeV1(r io.Reader) (*bufio.Scanner, error) {
	dcmp, err := zstd.NewReader(r)
	if err != nil {
		return nil, errors.Wrap(err, "zstd.NewReader")
	}
	s := bufio.NewScanner(dcmp)
	s.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		// This is bufio.SplitLines, but with 0x00 instead of \n, and \r-chomping code
		// deleted.
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
	return s, nil
}

type decodeV1PathState int

const (
	projectState decodeV1PathState = iota
	repoState
	refState
	fileState
)

// String implements fmt.Stringer.
func (s decodeV1PathState) String() string {
	//exhaustive:enforce
	switch s {
	case projectState:
		return "project"
	case repoState:
		return "repo"
	case refState:
		return "ref"
	case fileState:
		return "file"
	}
	panic("unknown DecodeV1 state")
}

// DecodeV1Path decodes a V1 path.  pachctl's convention of branch=commit is not supported.
func DecodeV1Path(path string) (*pfs.File, error) {
	repo := &pfs.Repo{
		Project: &pfs.Project{},
		Type:    "user",
	}
	result := &pfs.File{
		Commit: &pfs.Commit{
			Repo: repo,
			Branch: &pfs.Branch{
				Repo: repo,
			},
		},
	}
	state := projectState
	buf := new(bytes.Buffer)
	for _, b := range []byte(path) {
		//exhaustive:enforce
		switch state {
		// This can be improved if we want better error messages; the special characters
		// mostly can't appear in other parts.  For example, if we see an @, we know we're
		// about to start parsing the ref, even if we're in the project state and not the
		// repo state.
		case projectState:
			if b == '/' {
				repo.Project.Name = buf.String()
				buf.Reset()
				state = repoState
				continue
			}
		case repoState:
			if b == '@' {
				repo.Name = buf.String()
				buf.Reset()
				state = refState
				continue
			}
		case refState:
			if b == ':' {
				if x := buf.String(); uuid.IsUUIDWithoutDashes(x) {
					result.Commit.Id = x
					result.Commit.Branch = nil
				} else {
					result.Commit.Branch.Name = x
				}
				buf.Reset()
				state = fileState
				continue
			}
		case fileState:
		}
		buf.WriteByte(b)
	}
	if state != fileState {
		return nil, errors.Errorf("incomplete file reference; in state %v (result: %v)", state.String(), result.String())
	}
	result.Path = buf.String()
	return result, nil
}
