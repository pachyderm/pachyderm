package pipeline

import (
	"io"

	"github.com/fsouza/go-dockerclient"
)

var DefaultConfig = docker.Config{
	AttachStdin:  true,
	AttachStdout: true,
	AttachStderr: true,
	OpenStdin:    true,
	StdinOnce:    true,
}

func startContainer(opts docker.CreateContainerOptions) (string, error) {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return "", err
	}
	container, err := client.CreateContainer(opts)
	if err != nil {
		return "", err
	}
	err = client.StartContainer(container.ID, opts.HostConfig)
	if err != nil {
		return "", err
	}

	return container.ID, nil
}

func stopContainer(id string) error {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return err
	}
	return client.StopContainer(id, 5)
}

func pullImage(image string) error {
	repository, tag := docker.ParseRepositoryTag(image)
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return err
	}
	opts := docker.PullImageOptions{Repository: repository, Tag: "latest"}
	if tag != "" {
		opts.Tag = tag
	}
	return client.PullImage(opts, docker.AuthConfiguration{})
}

func pipeToStdin(id string, in io.Reader) error {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return err
	}
	return client.AttachToContainer(docker.AttachToContainerOptions{
		Container:   id,
		InputStream: in,
		Stdin:       true,
		Stream:      true,
	})
}

func containerLogs(id string, out io.Writer) error {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return err
	}
	return client.AttachToContainer(docker.AttachToContainerOptions{
		Container:    id,
		OutputStream: out,
		ErrorStream:  out,
		Stdout:       true,
		Stderr:       true,
		Logs:         true,
	})
}

func waitContainer(id string) (int, error) {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return 0, err
	}

	return client.WaitContainer(id)
}

func isImageLocal(image string) bool {
	repository, _ := docker.ParseRepositoryTag(image)
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return false
	}
	images, err := client.ListImages(docker.ListImagesOptions{All: true, Digests: false})
	if err != nil {
		return false
	}

	for _, image := range images {
		if image.ID == repository {
			return true
		}
	}

	return false
}
