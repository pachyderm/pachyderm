package run

import (
	"testing"

	"github.com/pachyderm/pachyderm/src/pkg/require"
	"github.com/pachyderm/pachyderm/src/pps/parse"
)

func TestGetNameToNodeInfo(t *testing.T) {
	pipeline, err := parse.NewParser().ParsePipeline("../parse/testdata/basic")
	require.NoError(t, err)
	nodeInfos, err := getNameToNodeInfo(pipeline.NameToNode)
	require.NoError(t, err)
	require.Equal(t, []string{"bar-node"}, nodeInfos["baz-node-bar-in-bar-out-in"].Parents)
}
