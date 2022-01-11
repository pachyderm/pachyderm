package dockertestenv

import (
	"context"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/sirupsen/logrus"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

func newDockerClient() docker.APIClient {
	dclient, err := docker.NewClientWithOpts(docker.FromEnv)
	if err != nil {
		panic(err)
	}
	return dclient
}

func getDockerHost() string {
	client := newDockerClient()
	defer client.Close()
	host := client.DaemonHost()
	u, err := url.Parse(host)
	if err != nil {
		panic(err)
	}
	if u.Scheme == "unix" {
		return "127.0.0.1"
	}
	return u.Hostname()
}

type containerSpec struct {
	Image   string
	Cmd     []string
	PortMap map[uint16]uint16
	Env     map[string]string
}

// ensureContainer checks that a container is running with containerName, if it is not, one will be created using image and portMap.
// image should be the name of an image e.g. minio/minio:latest
// portMap is a mapping from host ports to container ports.  The image/values of this mapping are taken implicitly to be the exposed container ports.
func ensureContainer(ctx context.Context, dclient docker.APIClient, containerName string, spec containerSpec) error {
	imageName := spec.Image
	portMap := spec.PortMap
	if cjson, err := dclient.ContainerInspect(ctx, containerName); err != nil {
		if !isErrNoSuchContainer(err) {
			return errors.EnsureStack(err)
		}
	} else {
		if cjson.State.Running {
			logrus.Infof("container %s is running. skip creation.", containerName)
			return nil
		}
		logrus.Infof("container %s exists, but is not running. deleting...", containerName)
		if err := dclient.ContainerRemove(ctx, containerName, types.ContainerRemoveOptions{}); err != nil {
			return errors.EnsureStack(err)
		}
	}
	logrus.Infof("container %s does not exist. creating...", containerName)
	if err := ensureImage(ctx, dclient, imageName); err != nil {
		return err
	}

	// container config
	exposedPorts := nat.PortSet{}
	for _, containerPort := range portMap {
		exposedPorts[nat.Port(strconv.Itoa(int(containerPort)))] = struct{}{}
	}
	var envSlice []string
	for k, v := range spec.Env {
		envSlice = append(envSlice, k+"="+v)
	}
	containerConfig := &container.Config{
		Image:        imageName,
		Cmd:          spec.Cmd,
		Env:          envSlice,
		ExposedPorts: exposedPorts,
	}
	// host config
	portBindings := nat.PortMap{}
	for hostPort, containerPort := range portMap {
		portBindings[nat.Port(strconv.Itoa(int(containerPort)))] = []nat.PortBinding{
			{HostIP: "0.0.0.0", HostPort: strconv.Itoa(int(hostPort))},
		}
	}
	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
	}
	resp, err := dclient.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, containerName)
	if err != nil {
		return errors.EnsureStack(err)
	}
	if len(resp.Warnings) > 0 {
		logrus.Warn(resp.Warnings)
	}
	logrus.Info("created container ", containerName)
	if err := dclient.ContainerStart(ctx, containerName, types.ContainerStartOptions{}); err != nil {
		return errors.EnsureStack(err)
	}
	logrus.Info("started container ", containerName)
	return nil
}

func ensureImage(ctx context.Context, dclient docker.APIClient, imageName string) error {
	rc, err := dclient.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		return errors.EnsureStack(err)
	}
	defer rc.Close()
	if err := readResponseBody(rc); err != nil {
		return err
	}
	return nil
}

func readResponseBody(rc io.ReadCloser) error {
	_, err := io.Copy(os.Stderr, rc)
	return errors.EnsureStack(err)
}

func isErrNoSuchContainer(err error) bool {
	return strings.Contains(err.Error(), "No such container:")
}
