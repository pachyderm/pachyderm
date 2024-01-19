package kindenv

import (
	"bytes"
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/adrg/xdg"
	"github.com/bazelbuild/rules_go/go/runfiles"
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

func ensureRegistry(ctx context.Context, name string, expose bool) (string, error) {
	// Figure out where to store registry data.
	registryVolume, err := xdg.StateFile(name + "/registry")
	if err != nil {
		return "", errors.Wrap(err, "find suitable location for registry files")
	}
	if err := os.MkdirAll(registryVolume, 0o755); err != nil {
		return "", errors.Wrapf(err, "create registry storage directory %v", registryVolume)
	}

	// Check if there's an existing registry and that it's configured correctly.
	log.Debug(ctx, "checking for existing registry", zap.String("container", name))
	digestFile, err := runfiles.Rlocation("_main/src/internal/kindenv/zot.json.sha256")
	if err != nil {
		return "", errors.Wrap(err, "find zot image digest; build with bazel")
	}
	digest, err := os.ReadFile(digestFile)
	if err != nil {
		return "", errors.Wrap(err, "read zot image digest")
	}
	digest = bytes.TrimRight(digest, "\n")
	zotImage := "zot:" + string(digest[len("sha256:"):])

	dc, err := docker.NewClientWithOpts(docker.FromEnv)
	if err != nil {
		return "", errors.Wrap(err, "create docker client")
	}
	if status, err := dc.ContainerInspect(ctx, name); err == nil {
		if status.State.Running {
			if status.Config.Image == zotImage {
				// Container is created; don't attempt further work.
				log.Info(ctx, "reusing existing container registry", zap.String("container", name))
				return registryVolume, nil
			}
		}
		// This catches the case where we're using a different image, or where the container
		// is stopped or something.
		log.Info(ctx, "container registry exists, but isn't configured correctly; deleting", zap.String("container", name))
		if err := destroyRegistry(ctx, name); err != nil {
			return "", errors.Wrap(err, "destroy broken registry")
		}
	} else if strings.Contains(err.Error(), "No such container") {
		// That's what the rest of this function exists to handle.
	} else {
		// An actual error from docker, that's strange.
		return "", errors.Wrapf(err, "inspect container %q", name)
	}

	// Push the bazel-versioned Zot image to the docker daemon.
	log.Info(ctx, "making zot available to the local docker")
	imageDir, err := runfiles.Rlocation("_main/src/internal/kindenv/zot")
	if err != nil {
		return "", errors.Wrap(err, "find zot image; build with bazel")
	}
	if err := SkopeoCommand(ctx, "copy", "oci:"+imageDir, "docker-daemon:"+zotImage).Run(); err != nil {
		return "", errors.Wrap(err, "copy zot into the local docker daemon")
	}

	// Start zot.
	log.Info(ctx, "starting container registry", zap.String("container", name), zap.String("registry", registryVolume))
	cc := &container.Config{
		Image: zotImage,
		Cmd:   strslice.StrSlice{"serve", "/etc/zot.json"},
		ExposedPorts: nat.PortSet{
			"5001": struct{}{},
		},
		User: strconv.Itoa(os.Getuid()),
	}
	hc := &container.HostConfig{
		Binds: []string{
			registryVolume + ":/var/lib/registry",
		},
		RestartPolicy: container.RestartPolicy{
			Name: "always",
		},
	}
	nc := new(network.NetworkingConfig)
	pc := new(v1.Platform)
	if expose {
		hc.PortBindings = nat.PortMap{
			"5001": []nat.PortBinding{{
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
		if strings.Contains(err.Error(), "endpoint with name pach-registry already exists in network kind") {
			return nil
		}
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
