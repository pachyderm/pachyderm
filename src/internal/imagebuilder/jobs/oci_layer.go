package jobs

import (
	"context"
	"fmt"

	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/imagebuilder/oci"
	"github.com/zeebo/xxh3"
	"go.starlark.net/starlark"
)

// Layer represents an image layer.
type Layer struct {
	NameAndPlatform
	Descriptor  v1.Descriptor
	ContentBlob Blob
	DiffID      digest.Digest
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
		NameAndPlatform{
			Name:     "layer:" + l.Input.Name,
			Platform: l.Input.Platform,
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
	return []Artifact{
		Layer{
			NameAndPlatform: NameAndPlatform{
				Name:     "layer:" + l.Input.Name,
				Platform: l.Input.Platform,
			},
			ContentBlob: Blob{
				Size:   layer.Descriptor.Size,
				SHA256: layer.SHA256,
				Underlying: &File{
					Name: fmt.Sprintf("layer-content:blob:%s", layer.Descriptor.Digest.Hex()),
					Path: layer.Underlying,
					Digest: Digest{
						Algorithm: "sha256",
						Value:     layer.SHA256[:],
					},
				},
			},
			Descriptor: layer.Descriptor,
			DiffID:     layer.DiffIDAsDigest(),
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
