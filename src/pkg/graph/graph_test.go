package graph

import (
	"errors"
	"sync/atomic"
	"testing"

	"go.pedge.io/protolog"
	"go.pedge.io/protolog/logrus"

	stdlogrus "github.com/Sirupsen/logrus"
	"github.com/pachyderm/pachyderm/src/pkg/require"
)

func init() {
	logrus.SetPusherOptions(
		logrus.PusherOptions{
			Formatter: &stdlogrus.TextFormatter{
				ForceColors: true,
			},
		},
	)
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
	counter := int32(0)
	nameToNodeFunc := map[string]func() error{
		"1":   testNodeFunc(&counter, intC, "1", 1, ""),
		"2":   testNodeFunc(&counter, intC, "2", 2, ""),
		"3-1": testNodeFunc(&counter, intC, "3-1", 3, ""),
		"3-2": testNodeFunc(&counter, intC, "3-2", 4, ""),
		"3-3": testNodeFunc(&counter, intC, "3-3", 5, ""),
		"4-1": testNodeFunc(&counter, intC, "4-1", 6, ""),
		"4-2": testNodeFunc(&counter, intC, "4-2", 7, ""),
		"5":   testNodeFunc(&counter, intC, "5", 8, ""),
	}

	run, err := build(nameToNodeInfo, nameToNodeFunc)
	require.NoError(t, err)
	err = run.Do()
	require.NoError(t, err)

	require.Equal(t, int32(8), counter)
	i := <-intC
	require.True(t, i == 1 || i == 2)
	i = <-intC
	require.True(t, i == 1 || i == 2)
	i = <-intC
	require.True(t, i == 3 || i == 4 || i == 5)
	i = <-intC
	require.True(t, i == 3 || i == 4 || i == 5)
	i = <-intC
	require.True(t, i == 3 || i == 4 || i == 5)
	i = <-intC
	require.True(t, i == 6 || i == 7)
	i = <-intC
	require.True(t, i == 6 || i == 7)
	i = <-intC
	require.True(t, i == 8)
}

func TestBuildWithError(t *testing.T) {
	intC := make(chan int, 5)
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
	counter := int32(0)
	nameToNodeFunc := map[string]func() error{
		"1":   testNodeFunc(&counter, intC, "1", 1, ""),
		"2":   testNodeFunc(&counter, intC, "2", 2, ""),
		"3-1": testNodeFunc(&counter, intC, "3-1", 3, "3-1:error"),
		"3-2": testNodeFunc(&counter, intC, "3-2", 4, ""),
		"3-3": testNodeFunc(&counter, intC, "3-3", 5, ""),
		"4-1": testNodeFunc(&counter, intC, "4-1", 6, ""),
		"4-2": testNodeFunc(&counter, intC, "4-2", 7, ""),
		"5":   testNodeFunc(&counter, intC, "5", 8, ""),
	}

	run, err := build(nameToNodeInfo, nameToNodeFunc)
	require.NoError(t, err)
	err = run.Do()
	require.NotNil(t, err)
	require.Equal(t, "3-1:error", err.Error())

	require.Equal(t, int32(5), counter)
	i := <-intC
	require.True(t, i == 1 || i == 2)
	i = <-intC
	require.True(t, i == 1 || i == 2)
	i = <-intC
	require.True(t, i == 3 || i == 4 || i == 5)
	i = <-intC
	require.True(t, i == 3 || i == 4 || i == 5)
	i = <-intC
	require.True(t, i == 3 || i == 4 || i == 5)
}

func testNodeFunc(counter *int32, intC chan int, nodeName string, i int, errString string) func() error {
	var err error
	if errString != "" {
		err = errors.New(errString)
	}
	return func() error {
		atomic.AddInt32(counter, 1)
		intC <- i
		protolog.Infof("ran %s, sent %d, returning %v\n", nodeName, i, err)
		return err
	}
}
