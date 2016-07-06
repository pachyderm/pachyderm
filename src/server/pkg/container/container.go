/*
Package container provides functionality to interact with containers.
*/
package container

import (
	"io"

	"github.com/fsouza/go-dockerclient"
)

// BuildOptions represents container build options.
type BuildOptions struct {
	Dockerfile   string
	OutputStream io.Writer
}

// PullOptions represents container pull options.
type PullOptions struct {
	NoPullIfLocal bool
	OutputStream  io.Writer
}

// CreateOptions represents container create options.
type CreateOptions struct {
	Binds      []string
	HasCommand bool
	Shell      string
}

// StartOptions represents container start options.
type StartOptions struct {
	Commands []string
}

// LogsOptions represents container log options.
type LogsOptions struct {
	Stdout io.Writer
	Stderr io.Writer
}

// WaitOptions represents container wait options.
type WaitOptions struct{}

// KillOptions represents container kill options.
type KillOptions struct{}

// RemoveOptions represents container remove options.
type RemoveOptions struct{}

// Client defines Pachyderm's interface to container-engines such as docker.
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

// NewDockerClient create a Client from given docker.Client.
func NewDockerClient(dockerClient *docker.Client) Client {
	return newDockerClient(dockerClient)
}
