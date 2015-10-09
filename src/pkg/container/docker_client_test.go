package container

import (
	"bytes"
	"testing"

	"github.com/fsouza/go-dockerclient"

	"go.pachyderm.com/pachyderm/src/pkg/require"
)

func TestCommandsSimple(t *testing.T) {
	testRun(
		t,
		"ubuntu:14.04",
		[]string{
			"echo hello",
		},
		"hello\n",
		"",
	)
}

func TestCommandsForLoops(t *testing.T) {
	testRun(
		t,
		"ubuntu:14.04",
		[]string{
			"for i in 1 2 3 4 5; do echo $i; done",
			"for i in 1 2 3 4 5; do echo X$i >&2; done",
		},
		"1\n2\n3\n4\n5\n",
		"X1\nX2\nX3\nX4\nX5\n",
	)
}

func testRun(t *testing.T, imageName string, commands []string, expectedStdout string, expectedStderr string) {
	client, err := newTestDockerClient()
	require.NoError(t, err)
	err = client.Pull(imageName, PullOptions{})
	require.NoError(t, err)
	container, err := client.Create(
		imageName,
		CreateOptions{
			HasCommand: commands != nil,
		},
	)
	require.NoError(t, err)
	err = client.Start(
		container,
		StartOptions{
			Commands: commands,
		},
	)
	require.NoError(t, err)
	err = client.Wait(container, WaitOptions{})
	require.NoError(t, err)
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)
	err = client.Logs(container, LogsOptions{Stdout: stdout, Stderr: stderr})
	require.NoError(t, err)
	require.Equal(t, expectedStdout, stdout.String())
	require.Equal(t, expectedStderr, stderr.String())
	err = client.Remove(container, RemoveOptions{})
	require.NoError(t, err)
}

func newTestDockerClient() (*dockerClient, error) {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, err
	}
	return newDockerClient(client), nil
}
