package jobs

import (
	"fmt"
	"testing"
)

func TestPlatformMatch(t *testing.T) {
	testData := []struct {
		input     Platform
		target    Reference
		wantMatch bool
	}{
		{
			input:     "linux/amd64",
			target:    Platform("linux/amd64"),
			wantMatch: true,
		},
		{
			input:     "linux/arm64",
			target:    Platform("linux/amd64"),
			wantMatch: false,
		},
		{
			input:     "linux/amd64",
			target:    Platform("linux/arm64"),
			wantMatch: false,
		},
		{
			input:     "linux/amd64",
			target:    AllPlatforms,
			wantMatch: true,
		},
		{
			input:     AllPlatforms,
			target:    Platform("linux/amd64"),
			wantMatch: true,
		},
		{
			input:     AllPlatforms,
			target:    Platform("linux/amd64"),
			wantMatch: true,
		},
		{
			input:     AllPlatforms,
			target:    AllPlatforms,
			wantMatch: true,
		},
		{
			input:     "linux/amd64",
			target:    Name("linux/amd64"),
			wantMatch: false,
		},
		{
			input: "linux/amd64",
			target: NameAndPlatform{
				Name:     "test",
				Platform: "linux/amd64",
			},
			wantMatch: true,
		},
	}

	for _, test := range testData {
		t.Run(fmt.Sprintf("%v match %v", test.input, test.target), func(t *testing.T) {
			match := test.input.Match(test.target)
			if match != test.wantMatch {
				if test.wantMatch {
					t.Errorf("want %v to match %v, but it doesn't", test.input, test.target)
				} else {
					t.Errorf("do not want %v to match %v, but it does", test.input, test.target)
				}
			}
		})
	}
}
