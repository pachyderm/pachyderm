package parse

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	_, err := ParsePipeline("testdata/basic")
	require.NoError(t, err)
}
