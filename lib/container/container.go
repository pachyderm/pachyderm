// package container contains convenience functions for manipulating containers
package container

import (
	"io"
	"log"

	"github.com/fsouza/go-dockerclient"
)

var DefaultConfig = docker.Config{
	AttachStdin:  true,
	AttachStdout: true,
	AttachStderr: true,
	OpenStdin:    true,
	StdinOnce:    true,
}

func RawStartContainer(opts docker.CreateContainerOptions) (string, error) {
	client, err := docker.NewClient("unix:///var/run/docker.sock")
	if err != nil {
		log.Print(err)
		return "", err
	}
	container, err := client.CreateContainer(opts)
	if err != nil {
		log.Print(err)
		return "", err
	}
	err = client.StartContainer(container.ID, opts.HostConfig)
	if err != nil {
		log.Print(err)
		return "", err
	}

	return container.ID, nil
}

func StartContainer(image string, command []string) (string, error) {
	config := docker.Config{Image: image, Cmd: command}
	opts := docker.CreateContainerOptions{Config: &config}
	return RawStartContainer(opts)
}

func StopContainer(id string) error {
	client, err := docker.NewClient("unix:///var/run/docker.sock")
	if err != nil {
		log.Print(err)
		return err
	}
	return client.StopContainer(id, 5)
}

func KillContainer(id string) error {
	client, err := docker.NewClient("unix:///var/run/docker.sock")
	if err != nil {
		log.Print(err)
		return err
	}
	return client.KillContainer(docker.KillContainerOptions{ID: id})
}

func IpAddr(containerId string) (string, error) {
	client, err := docker.NewClient("unix:///var/run/docker.sock")
	if err != nil {
		log.Print(err)
		return "", err
	}
	container, err := client.InspectContainer(containerId)
	if err != nil {
		log.Print(err)
		return "", err
	}

	return container.NetworkSettings.IPAddress, nil
}

func PullImage(image string) error {
	client, err := docker.NewClient("unix:///var/run/docker.sock")
	if err != nil {
		log.Print(err)
		return err
	}
	return client.PullImage(docker.PullImageOptions{Repository: image}, docker.AuthConfiguration{})
}

func PipeToStdin(id string, in io.Reader) error {
	client, err := docker.NewClient("unix:///var/run/docker.sock")
	if err != nil {
		log.Print(err)
		return err
	}
	return client.AttachToContainer(docker.AttachToContainerOptions{
		Container:   id,
		InputStream: in,
		Stdin:       true,
		Stream:      true,
	})
}

func ContainerLogs(id string, out io.Writer) error {
	client, err := docker.NewClient("unix:///var/run/docker.sock")
	if err != nil {
		log.Print(err)
		return err
	}
	return client.AttachToContainer(docker.AttachToContainerOptions{
		Container:    id,
		OutputStream: out,
		ErrorStream:  out,
		Stdout:       true,
		Stderr:       true,
		Stream:       true,
	})
}

func WaitContainer(id string) (int, error) {
	client, err := docker.NewClient("unix:///var/run/docker.sock")
	if err != nil {
		log.Print(err)
		return 0, err
	}

	return client.WaitContainer(id)
}
