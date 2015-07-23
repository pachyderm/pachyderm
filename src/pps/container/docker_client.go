package container

import (
	"fmt"

	"github.com/fsouza/go-dockerclient"
	"github.com/pachyderm/pachyderm/src/log"
)

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
	return c.client.BuildImage(
		docker.BuildImageOptions{
			Name:           imageName,
			Dockerfile:     buildOptions.Dockerfile,
			SuppressOutput: true,
			OutputStream:   log.Writer(),
			ContextDir:     contextDir,
		},
	)
}

func (c *dockerClient) Pull(imageName string, pullOptions PullOptions) error {
	repository, tag := docker.ParseRepositoryTag(imageName)
	if tag == "" {
		tag = "latest"
	}
	return c.client.PullImage(
		docker.PullImageOptions{
			Repository:   repository,
			Tag:          tag,
			OutputStream: log.Writer(),
		},
		docker.AuthConfiguration{},
	)
}

func (c *dockerClient) Run(imageName string, runOptions RunOptions) (retErr error) {
	createContainerOptions, err := getDockerCreateContainerOptions(imageName, runOptions)
	if err != nil {
		return err
	}
	container, err := c.client.CreateContainer(createContainerOptions)
	if err != nil {
		return err
	}
	defer func() {
		if err := c.client.RemoveContainer(
			docker.RemoveContainerOptions{
				ID:    container.ID,
				Force: true,
			},
		); err != nil && retErr == nil {
			retErr = err
		}
	}()
	if err := c.client.StartContainer(container.ID, createContainerOptions.HostConfig); err != nil {
		return err
	}
	errC := make(chan error)
	go func() {
		errC <- c.client.Logs(
			docker.LogsOptions{
				Container:    container.ID,
				OutputStream: log.Writer(),
				ErrorStream:  log.Writer(),
				Stdout:       true,
				Stderr:       true,
				Timestamps:   true,
			},
		)
	}()
	exitCode, err := c.client.WaitContainer(container.ID)
	logsErr := <-errC
	if err != nil {
		return err
	}
	if exitCode != 0 {
		return fmt.Errorf("container %s for image %s had exit code %d", container.ID, imageName, exitCode)
	}
	return logsErr
}

func getDockerCreateContainerOptions(imageName string, runOptions RunOptions) (docker.CreateContainerOptions, error) {
	config, err := getDockerConfig(imageName, runOptions)
	if err != nil {
		return docker.CreateContainerOptions{}, err
	}
	hostConfig, err := getDockerHostConfig(imageName, runOptions)
	if err != nil {
		return docker.CreateContainerOptions{}, err
	}
	return docker.CreateContainerOptions{
		Config:     config,
		HostConfig: hostConfig,
	}, nil
}

func getDockerConfig(imageName string, runOptions RunOptions) (*docker.Config, error) {
	return &docker.Config{
		Image: imageName,
	}, nil
}

func getDockerHostConfig(imageName string, runOptions RunOptions) (*docker.HostConfig, error) {
	return &docker.HostConfig{}, nil
}
