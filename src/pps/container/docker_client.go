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

func (c *dockerClient) Build(imageName string, contextDir string, options BuildOptions) error {
	return c.client.BuildImage(
		docker.BuildImageOptions{
			Name:           imageName,
			Dockerfile:     options.Dockerfile,
			SuppressOutput: true,
			OutputStream:   log.Writer(),
			ContextDir:     contextDir,
		},
	)
}

func (c *dockerClient) Pull(imageName string, options PullOptions) error {
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

func (c *dockerClient) Create(imageName string, options CreateOptions) (retVal []string, retErr error) {
	createContainerOptions, err := getDockerCreateContainerOptions(imageName, options)
	if err != nil {
		return nil, err
	}
	numContainers := options.NumContainers
	if numContainers == 0 {
		numContainers = 1
	}
	var containerIDs []string
	defer func() {
		if retErr != nil {
			for _, containerID := range containerIDs {
				_ = c.Remove(containerID, RemoveOptions{})
			}
		}
	}()
	for i := 0; i < numContainers; i++ {
		container, err := c.client.CreateContainer(createContainerOptions)
		if err != nil {
			return nil, err
		}
		containerIDs = append(containerIDs, container.ID)
	}
	return containerIDs, nil
}

func (c *dockerClient) Start(containerID string, options StartOptions) error {
	container, err := c.client.InspectContainer(containerID)
	if err != nil {
		return err
	}
	return c.client.StartContainer(container.ID, container.HostConfig)
}

func (c *dockerClient) Wait(containerID string, options WaitOptions) error {
	errC := make(chan error)
	go func() {
		errC <- c.client.Logs(
			docker.LogsOptions{
				Container:    containerID,
				OutputStream: log.Writer(),
				ErrorStream:  log.Writer(),
				Stdout:       true,
				Stderr:       true,
				Timestamps:   true,
			},
		)
	}()
	exitCode, err := c.client.WaitContainer(containerID)
	logsErr := <-errC
	if err != nil {
		return err
	}
	if exitCode != 0 {
		return fmt.Errorf("container %s had exit code %d", containerID, exitCode)
	}
	return logsErr
}

func (c *dockerClient) Kill(containerID string, options KillOptions) error {
	return c.client.KillContainer(
		docker.KillContainerOptions{
			ID: containerID,
		},
	)
}

func (c *dockerClient) Remove(containerID string, options RemoveOptions) error {
	return c.client.RemoveContainer(
		docker.RemoveContainerOptions{
			ID:    containerID,
			Force: true,
		},
	)
}

func getDockerCreateContainerOptions(imageName string, options CreateOptions) (docker.CreateContainerOptions, error) {
	config, err := getDockerConfig(imageName, options)
	if err != nil {
		return docker.CreateContainerOptions{}, err
	}
	hostConfig, err := getDockerHostConfig()
	if err != nil {
		return docker.CreateContainerOptions{}, err
	}
	return docker.CreateContainerOptions{
		Config:     config,
		HostConfig: hostConfig,
	}, nil
}

func getDockerConfig(imageName string, options CreateOptions) (*docker.Config, error) {
	return &docker.Config{
		Image: imageName,
	}, nil
}

func getDockerHostConfig() (*docker.HostConfig, error) {
	return &docker.HostConfig{}, nil
}
