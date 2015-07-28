package container

import (
	"bytes"
	"fmt"

	"go.pedge.io/protolog"

	"github.com/fsouza/go-dockerclient"
)

const (
	defaultShell = "sh"
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
			OutputStream:   protolog.Writer(),
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
			OutputStream: protolog.Writer(),
		},
		docker.AuthConfiguration{},
	)
}

func (c *dockerClient) Create(imageName string, options CreateOptions) (_ []string, retErr error) {
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
	if err := c.client.StartContainer(container.ID, container.HostConfig); err != nil {
		return err
	}
	if options.Commands != nil && len(options.Commands) > 0 {
		buffer := bytes.NewBuffer(nil)
		for _, command := range options.Commands {
			if _, err := buffer.WriteString(command + "\n"); err != nil {
				return err
			}
		}
		if err := c.client.AttachToContainer(
			docker.AttachToContainerOptions{
				Container:   container.ID,
				InputStream: buffer,
				Stdin:       true,
				Stream:      true,
			},
		); err != nil {
			return err
		}
	}
	return nil
}

func (c *dockerClient) Logs(containerID string, options LogsOptions) error {
	return c.client.Logs(
		docker.LogsOptions{
			Container:    containerID,
			OutputStream: options.Stdout,
			ErrorStream:  options.Stderr,
			Stdout:       options.Stdout != nil,
			Stderr:       options.Stderr != nil,
		},
	)
}

func (c *dockerClient) Wait(containerID string, options WaitOptions) error {
	errC := make(chan error)
	go func() {
		errC <- c.client.Logs(
			docker.LogsOptions{
				Container:    containerID,
				OutputStream: protolog.Writer(),
				ErrorStream:  protolog.Writer(),
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
	hostConfig, err := getDockerHostConfig(options)
	if err != nil {
		return docker.CreateContainerOptions{}, err
	}
	return docker.CreateContainerOptions{
		Config:     config,
		HostConfig: hostConfig,
	}, nil
}

func getDockerConfig(imageName string, options CreateOptions) (*docker.Config, error) {
	config := &docker.Config{
		Image: imageName,
	}
	if options.HasCommand {
		config.AttachStdin = true
		config.OpenStdin = true
		config.StdinOnce = true
		if options.Shell != "" {
			config.Entrypoint = []string{options.Shell}
		} else {
			config.Entrypoint = []string{defaultShell}
		}
	}
	return config, nil
}

func getDockerHostConfig(options CreateOptions) (*docker.HostConfig, error) {
	hostConfig := &docker.HostConfig{}
	if options.Input != nil {
		if hostConfig.Binds == nil {
			hostConfig.Binds = make([]string, 0)
		}
		for key, value := range options.Input.Host {
			hostConfig.Binds = append(hostConfig.Binds, fmt.Sprintf("%s:%s:ro", key, value))
		}
	}
	if options.Output != nil {
		if hostConfig.Binds == nil {
			hostConfig.Binds = make([]string, 0)
		}
		for key, value := range options.Output.Host {
			hostConfig.Binds = append(hostConfig.Binds, fmt.Sprintf("%s:%s:rw", key, value))
		}
	}
	return hostConfig, nil
}
