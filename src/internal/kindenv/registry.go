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

	// It is unlikely that localhost:5001 will work for all setups, like pushing from Mac ->
	// registry on Linux VM.  So we'll pretend this is configurable and then come up with a plan
	// for making it configurable.  The port has to be 5001 for compatability with "docker
	// push"; it's the magic number that tells it "hey don't check the TLS cert".
	registryHostname = "localhost"
)

//go:embed zot.json
var zotJSON []byte

func ensureRegistry(ctx context.Context) (string, error) {
	registryVolume, err := xdg.StateFile("pach-registry/registry")
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
	if _, err := dc.ContainerInspect(ctx, "pach-registry"); err == nil {
		// Container is created; don't attempt further work.
		return registryVolume, nil
	} else if strings.Contains(err.Error(), "No such container") {
		// That's what the rest of this function exists to handle.
	} else {
		return "", errors.Wrap(err, "inspect pach-registry")
	}

	configVolume, err := xdg.StateFile("pach-registry/config/zot.json")
	if err != nil {
		return "", errors.Wrap(err, "find suitable location for zot.json")
	}
	if err := os.WriteFile(configVolume, zotJSON, 0o644); err != nil {
		return "", errors.Wrap(err, "write zot.json")
	}

	log.Info(ctx, "starting container registry localhost:5001", zap.String("registry", registryVolume), zap.String("config", configVolume))
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
		PortBindings: nat.PortMap{
			"5000": []nat.PortBinding{{
				HostIP:   "0.0.0.0",
				HostPort: "5001",
			}},
		},
	}
	nc := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			"kind": &network.EndpointSettings{},
		},
	}
	pc := &v1.Platform{}
	container, err := dc.ContainerCreate(ctx, cc, hc, nc, pc, "pach-registry")
	if err != nil {
		return "", errors.Wrap(err, "create zot container")
	}
	if err := dc.ContainerStart(ctx, container.ID, types.ContainerStartOptions{}); err != nil {
		return "", errors.Wrap(err, "start zot container")
	}
	log.Info(ctx, "registry started ok")
	return registryVolume, nil
}
