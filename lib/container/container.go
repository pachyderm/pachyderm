// package container contains convenience functions for manipulating containers
package container

import (
	"log"

	"github.com/samalba/dockerclient"
)

// TODO(rw): pull this out into a separate library
func StartContainer(image string, command []string) (string, error) {
	docker, err := dockerclient.NewDockerClient("unix:///var/run/docker.sock", nil)
	containerConfig := &dockerclient.ContainerConfig{Image: image, Cmd: command}

	containerId, err := docker.CreateContainer(containerConfig, "")
	if err != nil {
		log.Print(err)
		return "", nil
	}

	if err := docker.StartContainer(containerId, &dockerclient.HostConfig{}); err != nil {
		log.Print(err)
		return "", err
	}

	return containerId, nil
}

// spinupContainer pulls image and starts a container from it with command. It
// returns the container id or an error.
// TODO(rw): pull this out into a separate library
func SpinupContainer(image string, command []string) (string, error) {
	log.Print("spinupContainer", " ", image, " ", command)
	docker, err := dockerclient.NewDockerClient("unix:///var/run/docker.sock", nil)
	if err != nil {
		log.Print(err)
		return "", err
	}
	if err := docker.PullImage(image, nil); err != nil {
		log.Print("Failed to pull ", image, " with error: ", err)
		// We keep going here because it might be a local image.
	}

	return StartContainer(image, command)
}

// TODO(rw): pull this out into a separate library
func StopContainer(containerId string) error {
	log.Print("stopContainer", " ", containerId)
	docker, err := dockerclient.NewDockerClient("unix:///var/run/docker.sock", nil)
	if err != nil {
		log.Print(err)
		return err
	}
	return docker.StopContainer(containerId, 5)
}

// TODO(rw): pull this out into a separate library
func IpAddr(containerId string) (string, error) {
	docker, err := dockerclient.NewDockerClient("unix:///var/run/docker.sock", nil)
	if err != nil {
		return "", err
	}
	containerInfo, err := docker.InspectContainer(containerId)
	if err != nil {
		return "", err
	}

	return containerInfo.NetworkSettings.IPAddress, nil
}
