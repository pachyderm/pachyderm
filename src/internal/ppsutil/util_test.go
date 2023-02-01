package ppsutil

import (
	"fmt"
	"testing"
)

func Example_truncatePair() {
	fmt.Println(truncatePair("", "abc", 2))
	fmt.Println(truncatePair("abc", "", 2))
	fmt.Println(truncatePair("a", "bc", 2))
	fmt.Println(truncatePair("ab", "c", 2))
	fmt.Println(truncatePair("abc", "def", 8))
	fmt.Println(truncatePair("ab", "defghijklmnop", 6))
	fmt.Println(truncatePair("abcdefghijkl", "12", 6))
	// Output: ab
	// ab
	// ab
	// ac
	// abcdef
	// abdefg
	// abcd12
}

func TestPipelineRcName(t *testing.T) {
	for _, c := range []struct {
		nameVersion, projectName, pipelineName string
		version                                uint64
		name                                   string
		isError                                bool
	}{
		{"", "default", "foo", 1, "", true},
		{"v2.4", "default", "foo", 1, "pipeline-foo-v1", false},
		{"v2.5", "default", "foo", 1, "pipeline-default-foo-3ambdookm6ghudshdorhi6p2bvuphuvl-v1", false},
		{"v2.5", "a0123456789b0123456789c0123456789d0123456789e0123456789", "foo", 1, "pipeline-a0123456789b01-foo-12r7pie5r21cm9uf5kfl011grk4njvti-v1", false},
		{"v2.5", "a0123456789b0123456789c0123456789d0123456789e0123456789", "foo", 10, "pipeline-a0123456789b0-foo-12r7pie5r21cm9uf5kfl011grk4njvti-v10", false},
		{"v2.5", "this-is-a-longish-project-name", "this-pipeline-name-really-is-longer-than-it-should-be-you-know", 1, "pipeline-this-is-a-this-pip-f7o9tpu5vsgi9hiribbbk3ioj225u30m-v1", false},
		{"v2.5", "this_is-a-longish-project-name", "this-pipeline-name-really-is-longer-than-it-should-be-you-know", 1, "pipeline-this-is-a-this-pip-7c2f89plo3gcrnjrg5dp350o8ss9uvtc-v1", false},
		{"v2.5", "myProject", "myPipeline", 1, "pipeline-myproject-mypipeli-2pi5b3ok7ocu5r77i12klq1jfu3stsei-v1", false},
		{"v2.5", "myProject", "my_Pipeline-has_a_truly-terribleName-but-I-like-it-that-way", 1, "pipeline-myproject-my-pipel-d097gttbe3n2auumg7cvv8reh59ltj17-v1", false},
		{"v2.5", "myProject", "my-Pipeline-has_a_truly-terribleName-but-I-like-it-that-way", 1, "pipeline-myproject-my-pipel-p8r2mj30egjtmi2tnegh0jo3ejrmki1g-v1", false},
		{"v2.5", "myProject", "my_Pipeline-has_a_truly-terribleName-but-I-like-it-that-way", 123, "pipeline-myprojec-my-pipe-d097gttbe3n2auumg7cvv8reh59ltj17-v123", false},
	} {
		if name, err := pipelineRcName(c.nameVersion, c.projectName, c.pipelineName, c.version); err != nil {
			if !c.isError {
				t.Errorf("case %v failed unexpectedly: %v", c, err)
			}
		} else if name != c.name {
			t.Errorf("case %v: expected %q; got %q", c, c.name, name)
		}
	}
}
