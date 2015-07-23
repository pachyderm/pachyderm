package container

import "github.com/fsouza/go-dockerclient"

type dockerClient struct {
	// confusing
	client *docker.Client
}

func newDockerClient(dockerClientOptions DockerClientOptions) (*dockerClient, error) {
	var client *docker.Client
	var err error
	if dockerClientOptions.DockerTLSOptions != nil {
		client, err = docker.NewTLSClientFromBytes(
			dockerClientOptions.Host,
			dockerClientOptions.DockerTLSOptions.CertPEMBlock,
			dockerClientOptions.DockerTLSOptions.KeyPEMBlock,
			dockerClientOptions.DockerTLSOptions.CaPEMCert,
		)
	} else {
		client, err = docker.NewClient(
			dockerClientOptions.Host,
		)
	}
	if err != nil {
		return nil, err
	}
	return &dockerClient{
		client,
	}, nil
}

func (c *dockerClient) Build(imageName string, contextDir string, buildOptions BuildOptions) error {
	return nil
}

func (c *dockerClient) Pull(imageName string, pullOptions PullOptions) error {
	return nil
}

func (c *dockerClient) Run(imageName string, runOptions RunOptions) error {
	return nil
}
