package parse

import (
	"fmt"
	"testing"

	"go.pachyderm.com/pachyderm/src/pkg/require"
)

func TestBasic(t *testing.T) {
	pipeline, err := NewParser().ParsePipeline("testdata/basic")
	require.NoError(t, err)
	fmt.Println(pipeline)
}
