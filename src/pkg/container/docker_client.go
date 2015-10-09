package container

import (
	"bytes"
	"fmt"

	"github.com/fsouza/go-dockerclient"
)

const (
	defaultShell = "sh"
)

type dockerClient struct {
	// confusing
	client *docker.Client
}

func newDockerClient(client *docker.Client) *dockerClient {
	return &dockerClient{client}
}

func (c *dockerClient) Build(imageName string, contextDir string, options BuildOptions) error {
	return c.client.BuildImage(
		docker.BuildImageOptions{
			Name:           imageName,
			Dockerfile:     options.Dockerfile,
			SuppressOutput: true,
			OutputStream:   options.OutputStream,
			ContextDir:     contextDir,
		},
	)
}

func (c *dockerClient) Pull(imageName string, options PullOptions) error {
	repository, tag := docker.ParseRepositoryTag(imageName)
	if tag == "" {
		tag = "latest"
	}
	if options.NoPullIfLocal {
		images, err := c.client.ListImages(
			docker.ListImagesOptions{
				All:     true,
				Digests: false,
			},
		)
		if err != nil {
			return err
		}
		repositoryTag := fmt.Sprintf("%s:%s", repository, tag)
		for _, image := range images {
			for _, foundRepositoryTag := range image.RepoTags {
				if repositoryTag == foundRepositoryTag {
					return nil
				}
			}
		}
	}
	return c.client.PullImage(
		docker.PullImageOptions{
			Repository:   repository,
			Tag:          tag,
			OutputStream: options.OutputStream,
		},
		docker.AuthConfiguration{},
	)
}

func (c *dockerClient) Create(imageName string, options CreateOptions) (_ string, retErr error) {
	createContainerOptions, err := getDockerCreateContainerOptions(imageName, options)
	if err != nil {
		return "", err
	}
	var containerID string
	defer func() {
		if retErr != nil {
			_ = c.Remove(containerID, RemoveOptions{})
		}
	}()
	container, err := c.client.CreateContainer(createContainerOptions)
	if err != nil {
		return "", err
	}
	containerID = container.ID
	return containerID, nil
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
	exitCode, err := c.client.WaitContainer(containerID)
	if err != nil {
		return err
	}
	if exitCode != 0 {
		return fmt.Errorf("container %s had exit code %d", containerID, exitCode)
	}
	return nil
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
	return &docker.HostConfig{
		Binds: options.Binds,
	}, nil
}
