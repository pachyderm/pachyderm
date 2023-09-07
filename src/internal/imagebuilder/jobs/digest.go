package jobs

import (
	"bytes"
	"crypto/sha256"
	"encoding"
	"encoding/hex"
	"hash"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/zeebo/blake3"
)

// Digest is an arbitrary message digest.
type Digest struct {
	Algorithm string
	Value     []byte
}

var _ encoding.TextMarshaler = (*Digest)(nil)
var _ encoding.TextUnmarshaler = (*Digest)(nil)

func (d Digest) MarshalText() ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.WriteString(d.Algorithm)
	buf.WriteRune(':')
	hw := hex.NewEncoder(buf)
	hw.Write(d.Value) // Can only error if buf.Write errors, which can't happen.
	return buf.Bytes(), nil
}

func (d Digest) String() string {
	b, _ := d.MarshalText()
	return string(b)
}

// UnmarshalText implements encoding.TextUnmarshaler.  The format is "algorithm:hex-encoded bytes of
// hash". An empty string unmarshals to an empty blake3 hash.  A string without a : unmarhsals to a
// hash of the type text (i.e. "sha256" unmarshals into an empty sha256.)
func (d *Digest) UnmarshalText(b []byte) error {
	sep := bytes.IndexByte(b, ':')
	if sep == -1 {
		sep = len(b) - 1
	}
	d.Algorithm = string(b[:sep])
	if d.Algorithm == "" {
		d.Algorithm = "blake3"
	}
	d.Value = make([]byte, hex.DecodedLen(len(b[sep+1:])))
	if _, err := hex.Decode(d.Value, b[sep+1:]); err != nil {
		return err
	}
	return nil
}

func (d Digest) Hash() (hash.Hash, error) {
	switch d.Algorithm {
	case "sha256":
		return sha256.New(), nil
	case "blake3":
		return blake3.New(), nil
	default:
		return nil, errors.Errorf("unrecognized hash algorithm %q", d.Algorithm)
	}
}
