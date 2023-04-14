package archiveserver

import (
	"bytes"
	"encoding/base64"
	"sort"

	"github.com/klauspost/compress/zstd"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

// EncodeV1 generates a V1 archive URL string.
func EncodeV1(paths []string) (string, error) {
	buf := new(bytes.Buffer)
	b64 := base64.NewEncoder(base64.RawURLEncoding, buf)
	if _, err := b64.Write([]byte{0x01}); err != nil { // V1
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
