package jobs

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

// Blob is something that can be pushed to an OCI registry.  For example, a "docker manifest" is a
// blob containing the image config, and then a blob for each layer's descriptor, which refers to a
// blob for the actual layer content (typically a tar.gz/tar.zst filesystem image).
type Blob struct {
	SHA256     [sha256.Size]byte
	Underlying *File
}

func (b Blob) String() string {
	return fmt.Sprintf("blob:sha256:%x", b.SHA256)
}

// NewBlobFromReader sets up a blob by reading an io.Reader.
func NewBlobFromReader(r io.Reader) (_ Blob, retErr error) {
	var keepFile bool
	f, err := os.CreateTemp("", "blob-*")
	if err != nil {
		return Blob{}, errors.Wrap(err, "create temp file for blob")
	}
	defer errors.Close(&retErr, f, "close temp file for blob")
	defer func() {
		if !keepFile {
			errors.JoinInto(&retErr, errors.Wrap(os.Remove(f.Name()), "unlink unwanted blob"))
		}
	}()
	h := sha256.New()
	mw := io.MultiWriter(f, h)
	if _, err := io.Copy(mw, r); err != nil {
		return Blob{}, errors.Wrap(err, "write out blob")
	}
	keepFile = true
	var sum [sha256.Size]byte
	copy(sum[:], h.Sum(nil))
	return Blob{
		SHA256: sum,
		Underlying: &File{
			Name: fmt.Sprintf("file:blob:%x", sum),
			Path: f.Name(),
			Digest: Digest{
				Algorithm: "sha256",
				Value:     sum[:],
			},
		},
	}, nil
}
