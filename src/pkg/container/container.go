/*
Package container provides functionality to interact with containers.
*/
package container //import "go.pachyderm.com/pachyderm/src/pkg/container"

import (
	"io"

	"github.com/fsouza/go-dockerclient"
)

type BuildOptions struct {
	Dockerfile   string
	OutputStream io.Writer
}

type PullOptions struct {
	NoPullIfLocal bool
	OutputStream  io.Writer
}

type CreateOptions struct {
	Binds      []string
	HasCommand bool
	Shell      string
}

type StartOptions struct {
	Commands []string
}

type LogsOptions struct {
	Stdout io.Writer
	Stderr io.Writer
}

type WaitOptions struct{}

type KillOptions struct{}

type RemoveOptions struct{}

type Client interface {
	Build(imageName string, contextDir string, options BuildOptions) error
	Pull(imageName string, options PullOptions) error
	Create(imageName string, options CreateOptions) (string, error)
	Start(containerID string, options StartOptions) error
	Logs(containerID string, options LogsOptions) error
	Wait(containerID string, options WaitOptions) error
	Kill(containerID string, options KillOptions) error
	Remove(containerID string, options RemoveOptions) error
}

func NewDockerClient(dockerClient *docker.Client) Client {
	return newDockerClient(dockerClient)
}
