package jobs

import (
	"context"
	"fmt"

	"github.com/zeebo/xxh3"
)

type PlatformManifest struct {
	Name     Name
	Config   Reference
	Platform Platform
	Layers   []Reference
}

func (m PlatformManifest) ID() uint64 {
	return xxh3.HashString(fmt.Sprint(m.Name, m.Config, m.Platform, m.Layers))
}

func (m PlatformManifest) Inputs() []Reference {
	var result []Reference
	result = append(result, m.Config)
	result = append(result, m.Layers...)
	return result
}

func (m PlatformManifest) Outputs() []Reference {
	return []Reference{
		NameAndPlatform{
			Name:     "manifest:" + string(m.Name),
			Platform: m.Platform,
		},
	}
}

func (m PlatformManifest) Run(ctx context.Context, jc *JobContext, inputs []Artifact) ([]Artifact, error) {
	return nil, nil
}

func PlatformManifests(name Name, template Name, layers []NameAndPlatform) []PlatformManifest {
	platforms := make(map[Platform][]Reference)
	for _, l := range layers {
		platforms[l.Platform] = append(platforms[l.Platform], l)
	}
	var result []PlatformManifest
	for platform, layers := range platforms {
		result = append(result, PlatformManifest{
			Name:     name,
			Config:   template,
			Layers:   layers,
			Platform: platform,
		})
	}
	return result
}
