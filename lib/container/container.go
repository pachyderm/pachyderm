// package container contains convenience functions for manipulating containers
package container

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/samalba/dockerclient"
)

func RawStartContainer(containerConfig *dockerclient.ContainerConfig, hostConfig *dockerclient.HostConfig) (string, error) {
	docker, err := dockerclient.NewDockerClient("unix:///var/run/docker.sock", nil)
	containerId, err := docker.CreateContainer(containerConfig, "")
	if err != nil {
		log.Print(err)
		return "", nil
	}

	if err := docker.StartContainer(containerId, hostConfig); err != nil {
		log.Print(err)
		return "", err
	}

	return containerId, nil
}

func StartContainer(image string, command []string) (string, error) {
	containerConfig := &dockerclient.ContainerConfig{Image: image, Cmd: command}
	return RawStartContainer(containerConfig, &dockerclient.HostConfig{})
}

// spinupContainer pulls image and starts a container from it with command. It
// returns the container id or an error.
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

func StopContainer(containerId string) error {
	log.Print("stopContainer", " ", containerId)
	docker, err := dockerclient.NewDockerClient("unix:///var/run/docker.sock", nil)
	if err != nil {
		log.Print(err)
		return err
	}
	return docker.StopContainer(containerId, 5)
}

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
