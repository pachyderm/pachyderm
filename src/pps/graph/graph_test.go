package graph

import (
	"testing"

	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/parse"
	"github.com/stretchr/testify/require"
)

func TestGetNameToNodeInfo(t *testing.T) {
	pipeline, err := parse.NewParser().ParsePipeline("../parse/testdata/basic", "")
	require.NoError(t, err)
	nodes := pps.GetNameToNode(pipeline)
	nodeInfos, err := GetNameToNodeInfo(nodes)
	require.NoError(t, err)
	require.Equal(t, []string{"baz-node-bar-in-bar-out-in"}, nodeInfos["bar-node"].Children)
	require.Equal(t, []string{"bar-node"}, nodeInfos["baz-node-bar-in-bar-out-in"].Parents)
}
