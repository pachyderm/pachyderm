package jobs

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/opencontainers/go-digest"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

// Blob is something that can be pushed to an OCI registry.  For example, a "docker manifest" is a
// blob containing the image config, and then a blob for each layer's descriptor, which refers to a
// blob for the actual layer content (typically a tar.gz/tar.zst filesystem image).
type Blob struct {
	SHA256     [sha256.Size]byte
	Size       int64
	Underlying *File
}

func (b Blob) String() string {
	return fmt.Sprintf("blob:sha256:%x", b.SHA256)
}

func (b Blob) Digest() digest.Digest {
	return digest.NewDigestFromBytes(digest.SHA256, b.SHA256[:])
}

func (b Blob) Open() (fs.File, error) {
	return os.Open(b.Underlying.Path)
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
	n, err := io.Copy(mw, r)
	if err != nil {
		return Blob{}, errors.Wrap(err, "write out blob")
	}
	keepFile = true
	var sum [sha256.Size]byte
	copy(sum[:], h.Sum(nil))
	return Blob{
		SHA256: sum,
		Size:   n,
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

func NewJSONBlob(x any) (Blob, error) {
	// This should use "canonical JSON", but all of the libraries for that seem to be
	// abandonware.  The idea is that every time you marshal the same Go object, you get the
	// same bytes out.  That doesn't really matter to us.  The things that would be the same
	// across releases (like config.json) aren't very big, so if we send a new blob with the
	// same content as an existing blob, oh well.
	js, err := json.Marshal(x)
	if err != nil {
		return Blob{}, errors.Wrap(err, "marshal")
	}
	b, err := NewBlobFromReader(bytes.NewReader(js))
	if err != nil {
		return Blob{}, errors.Wrap(err, "blobify")
	}
	return b, nil
}
