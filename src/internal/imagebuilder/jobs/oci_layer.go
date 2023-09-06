package jobs

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/imagebuilder/oci"
	"github.com/zeebo/xxh3"
)

type Blob struct {
	SHA256     [sha256.Size]byte
	Underlying *File
}

func (b Blob) String() string {
	return fmt.Sprintf("blob:sha256:%x", b.SHA256)
}

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

type Layer struct {
	NameAndPlatform
	Content, Descriptor Blob
}

type DirectoryLayer struct {
	Input NameAndPlatform
}

var _ Job = (*DirectoryLayer)(nil)

func (l DirectoryLayer) ID() uint64 {
	return xxh3.HashString(string(l.Input.Name) + string(l.Input.Platform))
}

func (l DirectoryLayer) String() string {
	return fmt.Sprintf("<oci layer from %v>", l.Input)
}

func (l DirectoryLayer) Inputs() []Reference {
	return []Reference{l.Input}
}

func (l DirectoryLayer) Outputs() []Reference {
	return []Reference{
		Layer{
			NameAndPlatform: NameAndPlatform{
				Name:     "layer:" + l.Input.Name,
				Platform: l.Input.Platform,
			},
		},
	}
}

func (l DirectoryLayer) Run(ctx context.Context, jc *JobContext, in []Artifact) ([]Artifact, error) {
	if len(in) != 1 {
		return nil, errors.New("wrong number of inputs")
	}
	file, ok := in[0].(WithFS)
	if !ok {
		return nil, errors.Errorf("input is type %T, want WithFS", in[0])
	}
	layer, err := oci.NewLayerFromFS(file.FS())
	if err != nil {
		return nil, errors.Wrap(err, "build layer blob")
	}
	layer.Descriptor.Platform = l.Input.Platform.OCIPlatform()

	js, err := json.Marshal(layer.Descriptor)
	if err != nil {
		return nil, errors.Wrap(err, "marshal layer descriptor")
	}
	db, err := NewBlobFromReader(bytes.NewReader(js))
	if err != nil {
		return nil, errors.Wrap(err, "blobify layer descriptor")
	}
	return []Artifact{
		Layer{
			NameAndPlatform: NameAndPlatform{
				Name:     "layer:" + l.Input.Name,
				Platform: l.Input.Platform,
			},
			Content: Blob{
				Underlying: &File{
					Name: fmt.Sprintf("layer-content:blob:%s", layer.Descriptor.Digest.Hex()),
					Path: layer.Underlying,
				},
			},
			Descriptor: db,
		},
	}, nil
}

func PlatformLayers(inputs []Reference) (result []DirectoryLayer) {
	for _, ref := range inputs {
		var (
			name     string
			platform Platform
		)
		if x, ok := ref.(WithName); ok {
			name = x.GetName()
		}
		if x, ok := ref.(WithPlatform); ok {
			platform = x.GetPlatform()
		}
		result = append(result, DirectoryLayer{
			Input: NameAndPlatform{
				Name:     name,
				Platform: platform,
			},
		})
	}
	return
}
