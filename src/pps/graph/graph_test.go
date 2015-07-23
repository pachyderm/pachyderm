package graph

import (
	"testing"

	"github.com/pachyderm/pachyderm/src/pps/parse"
	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	pipeline, err := parse.NewParser().ParsePipeline("../parse/testdata/basic", "")
	require.NoError(t, err)
	pipelineInfo, err := NewGrapher().GetPipelineInfo(pipeline)
	require.NoError(t, err)
	m := pipelineInfo.NameToNodeInfo
	require.Equal(t, []string{"baz-node-bar-in-bar-out-in"}, m["bar-node"].Children)
	require.Equal(t, []string{"bar-node"}, m["baz-node-bar-in-bar-out-in"].Parents)
}
