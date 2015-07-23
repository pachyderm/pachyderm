package container

import "github.com/pachyderm/pachyderm/src/pps"

type BuildOptions struct {
	Dockerfile string
}

type PullOptions struct{}

type RunOptions struct {
	Input         *pps.Input
	Output        *pps.Output
	Commands      []string
	NumContainers int
}

type Client interface {
	Build(imageName string, contextDir string, buildOptions BuildOptions) error
	Pull(imageName string, pullOptions PullOptions) error
	Run(imageName string, runOptions RunOptions) error
}

type DockerClientOptions struct {
	Host             string
	DockerTLSOptions *DockerTLSOptions
}

type DockerTLSOptions struct {
	CertPEMBlock []byte
	KeyPEMBlock  []byte
	CaPEMCert    []byte
}

func NewDockerClient(dockerClientOptions DockerClientOptions) (Client, error) {
	return newDockerClient(dockerClientOptions)
}
