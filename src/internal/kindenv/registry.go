package kindenv

import (
	"context"
	_ "embed"
	"os"
	"strconv"
	"strings"

	"github.com/adrg/xdg"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	docker "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"go.uber.org/zap"
)

const (
	// Zot is a container registry that stores containers in the same format that "bazel build"
	// stores them.  To push, we just have to copy some files around.
	zotImage = "ghcr.io/project-zot/zot-linux-amd64:v1.4.3@sha256:e5a5be113155d1e0032e5d669888064209da95c107497524f8d4eac8ed50b378"
)

//go:embed zot.json
var zotJSON []byte

func ensureRegistry(ctx context.Context, name string, expose bool) (string, error) {
	registryVolume, err := xdg.StateFile(name + "/registry")
	if err != nil {
		return "", errors.Wrap(err, "find suitable location for registry files")
	}
	if err := os.MkdirAll(registryVolume, 0o755); err != nil {
		return "", errors.Wrapf(err, "create registry storage directory %v", registryVolume)
	}

	dc, err := docker.NewClientWithOpts(docker.FromEnv)
	if err != nil {
		return "", errors.Wrap(err, "create docker client")
	}
	if status, err := dc.ContainerInspect(ctx, name); err == nil {
		if status.State.Running {
			// Container is created; don't attempt further work.
			log.Info(ctx, "reusing existing container registry", zap.String("container", name))
			return registryVolume, nil
		}
		log.Info(ctx, "container registry exists, but isn't running; deleting", zap.String("container", name))
		if err := destroyRegistry(ctx, name); err != nil {
			return "", errors.Wrap(err, "destroy broken registry")
		}
	} else if strings.Contains(err.Error(), "No such container") {
		// That's what the rest of this function exists to handle.
	} else {
		return "", errors.Wrapf(err, "inspect container %q", name)
	}

	configVolume, err := xdg.StateFile(name + "/config/zot.json")
	if err != nil {
		return "", errors.Wrap(err, "find suitable location for zot.json")
	}
	if err := os.WriteFile(configVolume, zotJSON, 0o644); err != nil {
		return "", errors.Wrap(err, "write zot.json")
	}

	log.Info(ctx, "starting container registry", zap.String("container", name), zap.String("registry", registryVolume), zap.String("config", configVolume))
	cc := &container.Config{
		Image: zotImage,
		Cmd:   strslice.StrSlice{"serve", "/etc/zot.json"},
		ExposedPorts: nat.PortSet{
			"5000": struct{}{},
		},
		User: strconv.Itoa(os.Getuid()),
	}
	hc := &container.HostConfig{
		Binds: []string{
			registryVolume + ":/var/lib/registry",
			configVolume + ":/etc/zot.json",
		},
		RestartPolicy: container.RestartPolicy{
			Name: "always",
		},
	}
	nc := new(network.NetworkingConfig)
	pc := new(v1.Platform)
	if expose {
		hc.PortBindings = nat.PortMap{
			"5000": []nat.PortBinding{{
				HostIP:   "0.0.0.0",
				HostPort: "5001",
			}},
		}
	}
	container, err := dc.ContainerCreate(ctx, cc, hc, nc, pc, name)
	if err != nil {
		return "", errors.Wrap(err, "create zot container")
	}
	if err := dc.ContainerStart(ctx, container.ID, types.ContainerStartOptions{}); err != nil {
		return "", errors.Wrap(err, "start zot container")
	}
	log.Info(ctx, "registry started ok")
	return registryVolume, nil
}

func connectRegistry(ctx context.Context, name string) error {
	dc, err := docker.NewClientWithOpts(docker.FromEnv)
	if err != nil {
		return errors.Wrap(err, "create docker client")
	}
	if err := dc.NetworkConnect(ctx, "kind", name, &network.EndpointSettings{}); err != nil {
		return errors.Wrapf(err, "docker network connect kind %v", name)
	}
	return nil
}

func destroyRegistry(ctx context.Context, name string) error {
	dc, err := docker.NewClientWithOpts(docker.FromEnv)
	if err != nil {
		return errors.Wrap(err, "create docker client")
	}
	if err := dc.ContainerRemove(ctx, name, types.ContainerRemoveOptions{Force: true}); err != nil {
		return errors.Wrapf(err, "docker rm -f %v", name)
	}
	return nil
}
