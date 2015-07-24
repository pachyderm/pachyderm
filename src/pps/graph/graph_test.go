package graph

import (
	"errors"
	"fmt"
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
	require.Equal(t, []string{"bar-node"}, nodeInfos["baz-node-bar-in-bar-out-in"].Parents)
}

func TestBuild(t *testing.T) {
	intC := make(chan int, 8)
	nameToNodeInfo := map[string]*NodeInfo{
		"1": &NodeInfo{
			Parents: []string{},
		},
		"2": &NodeInfo{
			Parents: []string{},
		},
		"3-1": &NodeInfo{
			Parents: []string{
				"1",
				"2",
			},
		},
		"3-2": &NodeInfo{
			Parents: []string{
				"1",
				"2",
			},
		},
		"3-3": &NodeInfo{
			Parents: []string{
				"1",
				"2",
			},
		},
		"4-1": &NodeInfo{
			Parents: []string{
				"3-1",
				"3-2",
				"3-3",
			},
		},
		"4-2": &NodeInfo{
			Parents: []string{
				"3-1",
				"3-2",
				"3-3",
			},
		},
		"5": &NodeInfo{
			Parents: []string{
				"4-1",
				"4-2",
			},
		},
	}
	nameToNodeFunc := map[string]func() error{
		"1":   testNodeFunc(intC, "1", 1, ""),
		"2":   testNodeFunc(intC, "2", 2, ""),
		"3-1": testNodeFunc(intC, "3-1", 3, ""),
		"3-2": testNodeFunc(intC, "3-2", 4, ""),
		"3-3": testNodeFunc(intC, "3-3", 5, ""),
		"4-1": testNodeFunc(intC, "4-1", 6, ""),
		"4-2": testNodeFunc(intC, "4-2", 7, ""),
		"5":   testNodeFunc(intC, "5", 8, ""),
	}

	run, err := build(newTestNodeErrorRecorder(), nameToNodeInfo, nameToNodeFunc)
	require.NoError(t, err)
	run.Do()

	i := <-intC
	require.True(t, i == 1 || i == 2)
	i = <-intC
	require.True(t, i == 1 || i == 2)
	i = <-intC
	require.True(t, i == 3 || i == 4 || i == 5)
}

func testNodeFunc(intC chan int, nodeName string, i int, errString string) func() error {
	var err error
	if errString != "" {
		err = errors.New(errString)
	}
	return func() error {
		intC <- i
		fmt.Printf("ran %s, sent %d, returning %v\n", nodeName, i, err)
		return err
	}
}

type testNodeErrorRecorder struct {
	stringC chan string
}

func newTestNodeErrorRecorder() *testNodeErrorRecorder {
	return &testNodeErrorRecorder{make(chan string)}
}

func (t *testNodeErrorRecorder) Record(nodeName string, err error) {
	t.stringC <- fmt.Sprintf("%s:%s", nodeName, err.Error())
}

func (t *testNodeErrorRecorder) strings() []string {
	close(t.stringC)
	var slice []string
	for s := range t.stringC {
		slice = append(slice, s)
	}
	return slice
}
