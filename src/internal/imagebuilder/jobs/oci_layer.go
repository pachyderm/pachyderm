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
	"go.starlark.net/starlark"
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

// Layer represents an image layer.
type Layer struct {
	NameAndPlatform
	Content, Descriptor Blob
}

var _ Reference = (*Layer)(nil)

// FSLayer is a Job that builds a layer from an input with an FS() method.
type FSLayer struct {
	Input NameAndPlatformAndFS
}

var _ Job = (*FSLayer)(nil)

func (l FSLayer) ID() uint64 {
	return xxh3.HashString(string(l.Input.Name) + string(l.Input.Platform))
}

func (l FSLayer) String() string {
	return fmt.Sprintf("<oci layer from %v>", l.Input)
}

func (l FSLayer) Inputs() []Reference {
	return []Reference{l.Input}
}

func (l FSLayer) Outputs() []Reference {
	return []Reference{
		Layer{
			NameAndPlatform: NameAndPlatform{
				Name:     "layer:" + l.Input.Name,
				Platform: l.Input.Platform,
			},
		},
	}
}

func (l FSLayer) Run(ctx context.Context, jc *JobContext, in []Artifact) ([]Artifact, error) {
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

// NewFromStarlark builds FSLayer jobs from input references.  Each input reference becomes its own
// layer, and retains any platform specificity of the input.  Each argument can be a list of
// references or an individual reference.
func (FSLayer) NewFromStarlark(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) ([]Job, error) {
	if len(kwargs) != 0 {
		return nil, errors.New("unexpected kwargs")
	}
	var inputs []Reference
	for i, arg := range args {
		refs, err := UnpackReferences(arg)
		if err != nil {
			return nil, errors.Wrapf(err, "args[%v]", i)
		}
		inputs = append(inputs, refs...)
	}
	var result []Job
	for _, layer := range PlatformLayers(inputs) {
		result = append(result, layer)
	}
	return result, nil
}

// PlatformLayers creates a layer for each input.
func PlatformLayers(inputs []Reference) (result []FSLayer) {
	for _, ref := range inputs {
		var (
			name     string
			platform Platform
		)
		if x, ok := ref.(WithName); ok {
			name = x.GetName()
		}
		platform = AllPlatforms
		if x, ok := ref.(WithPlatform); ok {
			platform = x.GetPlatform()
		}
		result = append(result, FSLayer{
			Input: NameAndPlatformAndFS{
				Name:     name,
				Platform: platform,
			},
		})
	}
	return
}
