package parse

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	pipeline, err := ParsePipeline("testdata/basic")
	require.NoError(t, err)
	fmt.Printf("%+v\n", pipeline)
}

func TestGetAllFilePaths(t *testing.T) {
	files, err := getAllFilePaths("testdata/basic", []string{}, []string{"other", "root/ignore", "ignore-me.yml"})
	require.NoError(t, err)
	require.Equal(
		t,
		[]string{
			"root/foo-node.yml",
			"root/foo-service.yml",
			"root/include/bar-node.yml",
			"root/include/bar-service.yml",
			"root/include/bat-node.yml",
			"root/include/baz-node.yml",
			"root/include/baz-service.yml",
		},
		files,
	)
}
