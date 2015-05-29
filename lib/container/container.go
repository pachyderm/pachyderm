// package container contains convenience functions for manipulating containers
package container

import (
	"io"
	"log"

	"github.com/fsouza/go-dockerclient"
)

func RawStartContainer(opts docker.CreateContainerOptions) (string, error) {
	client, err := docker.NewClient("unix:///var/run/docker.sock")
	if err != nil {
		log.Print(err)
		return "", nil
	}
	container, err := client.CreateContainer(opts)
	if err != nil {
		log.Print(err)
		return "", nil
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
		return nil
	}
	return client.StopContainer(id, 5)
}

func IpAddr(containerId string) (string, error) {
	client, err := docker.NewClient("unix:///var/run/docker.sock")
	if err != nil {
		log.Print(err)
		return "", nil
	}
	container, err := client.InspectContainer(containerId)
	if err != nil {
		log.Print(err)
		return "", err
	}

	return container.NetworkSettings.IPAddress, nil
}

func ContainerLogs(id string, out io.Writer) error {
	client, err := docker.NewClient("unix:///var/run/docker.sock")
	if err != nil {
		log.Print(err)
		return nil
	}
	return client.AttachToContainer(docker.AttachToContainerOptions{
		Container:    id,
		OutputStream: out,
		ErrorStream:  out,
		Stdout:       true,
		Stderr:       true,
	})
}
