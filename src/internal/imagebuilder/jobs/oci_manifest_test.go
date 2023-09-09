package jobs

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmputil"
)

func TestManifestJobs(t *testing.T) {
	cfg := v1.Image{
		Author: "example@example.com",
	}
	testData := []struct {
		name    string
		input   [][]Reference
		want    []BuildManifest
		wantErr string
	}{
		{
			name: "empty",
		},
		{
			name: "basic",
			input: [][]Reference{
				{NameAndPlatform{"a", "linux/amd64"}, NameAndPlatform{"a", "linux/arm64"}},
				{NameAndPlatform{"b", "linux/amd64"}, NameAndPlatform{"b", "linux/arm64"}},
			},
			want: []BuildManifest{
				{
					NameAndPlatform: NameAndPlatform{
						Name:     "test",
						Platform: "linux/amd64",
					},
					Config: cfg,
					Layers: []Reference{NameAndPlatform{"a", "linux/amd64"}, NameAndPlatform{"b", "linux/amd64"}},
				},
				{
					NameAndPlatform: NameAndPlatform{
						Name:     "test",
						Platform: "linux/arm64",
					},
					Config: cfg,
					Layers: []Reference{NameAndPlatform{"a", "linux/arm64"}, NameAndPlatform{"b", "linux/arm64"}},
				},
			},
		},
		{
			name: "not enough refs to build every platform",
			input: [][]Reference{
				{NameAndPlatform{"a", "linux/amd64"}},
				{NameAndPlatform{"b", "linux/amd64"}, NameAndPlatform{"b", "linux/arm64"}},
			},
			want: []BuildManifest{
				{
					NameAndPlatform: NameAndPlatform{
						Name:     "test",
						Platform: "linux/amd64",
					},
					Config: cfg,
					Layers: []Reference{NameAndPlatform{"a", "linux/amd64"}, NameAndPlatform{"b", "linux/amd64"}},
				},
			},
		},
		{
			name: "single platform",
			input: [][]Reference{
				{Name("a")},
				{NameAndPlatform{"b", AllPlatforms}},
			},
			want: []BuildManifest{
				{
					NameAndPlatform: NameAndPlatform{
						Name:     "test",
						Platform: AllPlatforms,
					},
					Config: cfg,
					Layers: []Reference{Name("a"), NameAndPlatform{"b", AllPlatforms}},
				},
			},
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			got, err := BuildManifestJobs("test", cfg, test.input)
			cont := cmputil.WantErr(t, err, test.wantErr)
			if !cont {
				return
			}
			sortManifests := func(a, b Manifest) bool {
				return a.NameAndPlatform.String() < b.NameAndPlatform.String()
			}
			if diff := cmp.Diff(test.want, got, cmpopts.SortSlices(sortManifests), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("jobs: (-want +got)\n%s", diff)
			}
		})
	}
}
