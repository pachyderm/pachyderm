package jobs

import (
	"context"
	"encoding/json"
	"fmt"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/zeebo/xxh3"
	"go.starlark.net/starlark"
)

// Manifest is a job that assembles an OCI manifest:
// https://github.com/opencontainers/image-spec/blob/main/manifest.md
type Manifest struct {
	NameAndPlatform
	Config v1.ImageConfig // https://github.com/opencontainers/image-spec/blob/main/config.md
	Layers []Reference
}

func (m Manifest) String() string {
	return fmt.Sprintf("<oci-manifest %v=%v%v#%v>", m.Name, imageConfigWrapper(m.Config).String(), m.Layers, m.Platform)
}

func (m Manifest) ID() uint64 {
	return xxh3.HashString(fmt.Sprint(m.Name, m.Config, m.Platform, m.Layers))
}

func (m Manifest) Inputs() []Reference {
	var result []Reference
	result = append(result, m.Layers...)
	return result
}

func (m Manifest) Outputs() []Reference {
	return []Reference{
		NameAndPlatform{
			Name:     fmt.Sprintf("manifest:%s", m.Name),
			Platform: m.Platform,
		},
	}
}

func (m Manifest) Run(ctx context.Context, jc *JobContext, inputs []Artifact) ([]Artifact, error) {
	return nil, nil
}

type imageConfigWrapper v1.ImageConfig

var _ starlark.Value = (*imageConfigWrapper)(nil)

func (imageConfigWrapper) Freeze() {} // Always frozen.
func (imageConfigWrapper) Hash() (uint32, error) {
	return 0, errors.New("v1.ImageConfig is unhashable")
}
func (w imageConfigWrapper) Truth() starlark.Bool { return true }
func (w imageConfigWrapper) Type() string         { return "v1.ImageConfig" }
func (w imageConfigWrapper) String() string {
	js, err := json.Marshal(v1.ImageConfig(w))
	if err != nil {
		return fmt.Sprintf("%#v", v1.ImageConfig(w))
	}
	return string(js)
}

// NewImageConfigFromStarlark builds a v1.ImageConfig from Starlark.
func NewImageConfigFromStarlark(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var workingDir string
	var user, stopSignal starlark.Value
	var exposedPorts, volumes *starlark.Set
	var env, entrypoint, cmd *starlark.List
	var labels map[string]string
	if len(args) != 0 {
		return nil, errors.New("unexpected positional args")
	}
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "user?", &user, "exposed_ports?", &exposedPorts, "env?", &env, "entrypoint?", &entrypoint, "cmd?", &cmd, "volumes", &volumes, "working_dir", &workingDir, "labels", &labels, "stop_signal", &stopSignal); err != nil {
		return nil, errors.Wrap(err, "unpack args")
	}
	return &imageConfigWrapper{
		WorkingDir: workingDir,
		Labels:     labels,
	}, nil
}

// NewFromStarlark builds manifest jobs.
// Starlark: (name: string, layers: Iterable[ReferenceList], config: v1.ImageConfig)
//
// The iterable should yield sets of references that become each layer.  They are shareded by
// platform into a separate manifest job for each platform.  If there are a different number of
// layers between platforms, an error is returned.
func (Manifest) NewFromStarlark(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) ([]Job, error) {
	var name string
	var layerIter starlark.Iterable
	var config *imageConfigWrapper
	if err := starlark.UnpackArgs("oci_manifest", args, kwargs, "name", &name, "layers", &layerIter, "config", &config); err != nil {
		return nil, errors.Wrap(err, "unpack args")
	}

	// Get all layers, so we know what list of platforms we're targeting.
	var layers [][]Reference
	iter := layerIter.Iterate()
	defer iter.Done()
	var v starlark.Value
	for i := 0; iter.Next(&v); i++ {
		if v == nil {
			continue
		}
		refs, err := UnpackReferences(v)
		if err != nil {
			return nil, errors.Wrap(err, "layers[%d]: unpack reference")
		}
		layers = append(layers, refs)
	}
	manifests, err := ManifestJobs(name, v1.ImageConfig(*config), layers)
	if err != nil {
		return nil, errors.Wrap(err, "create manifest jobs")
	}
	jobs := make([]Job, len(manifests))
	for i := range manifests {
		jobs[i] = manifests[i]
	}
	return jobs, nil
}

func ManifestJobs(name string, config v1.ImageConfig, layers [][]Reference) ([]Manifest, error) {
	var allPlatformLayers []Reference
	platformLayers := make(map[Platform][]Reference)

	// Inspect each layer's references, and build a list of platforms that are desired.
	for li, refs := range layers {
		for _, ref := range refs {
			all := true
			if r, ok := ref.(WithPlatform); ok {
				if p := r.GetPlatform(); p != AllPlatforms {
					platformLayers[p] = []Reference{}
					all = false
				}
			} else {
				all = true
			}
			if all && len(refs) != 1 {
				return nil, errors.Errorf("layers[%d]: layers that have references to AllPlatforms or no platform must only contain a single reference, got %d", li, len(layers))
			} else if all {
				allPlatformLayers = append(allPlatformLayers, ref)
			}
		}
	}

	// If no layers, specified platforms, then just return a single manifest job now.
	if len(platformLayers) == 0 {
		if len(layers) == 0 {
			return nil, nil
		}
		return []Manifest{
			{
				NameAndPlatform: NameAndPlatform{
					Name:     name,
					Platform: AllPlatforms,
				},
				Config: config,
				Layers: allPlatformLayers,
			},
		}, nil
	}

	// Otherwise, put each layer under the right platform key.
	for _, refs := range layers {
		for _, ref := range refs {
			var allPlatforms bool
			if r, ok := ref.(WithPlatform); ok {
				p := r.GetPlatform()
				if p == AllPlatforms {
					allPlatforms = true
				} else {
					platformLayers[p] = append(platformLayers[p], ref)
				}
			} else {
				allPlatforms = true
			}
			if allPlatforms {
				// We checked above that there is only one reference in this slice.
				for p := range platformLayers {
					platformLayers[p] = append(platformLayers[p], ref)
				}
			}
		}
	}
	// Check that the manifest we're about to build has the same number of layers as the
	// previous manifests.  The API seems cleanest if you can insert all the refs you have into
	// each layer, and end up supporting the platforms that are common between all layers.
	var n int
	for _, layers := range platformLayers {
		n = max(n, len(layers))
	}
	// Build the manifests.
	var result []Manifest
	for p, layers := range platformLayers {
		if len(layers) != n {
			continue
		}
		result = append(result, Manifest{
			NameAndPlatform: NameAndPlatform{
				Name:     name,
				Platform: p,
			},
			Config: config,
			Layers: layers,
		})
	}
	return result, nil
}
