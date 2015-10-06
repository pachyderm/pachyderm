[![CircleCI](https://circleci.com/gh/peter-edge/go-dockervolume/tree/master.png)](https://circleci.com/gh/peter-edge/go-dockervolume/tree/master)
[![API Documentation](http://img.shields.io/badge/api-Godoc-blue.svg?style=flat-square)](https://godoc.org/go.pedge.io/dockervolume)

A small library taking care of the generic code for docker volume plugins written in Go.

* https://blog.docker.com/2015/06/extending-docker-with-plugins/
* https://github.com/docker/docker/blob/master/docs/extend/plugins.md
* https://github.com/docker/docker/blob/master/docs/extend/plugin_api.md
* https://github.com/docker/docker/blob/master/docs/extend/plugins_volume.md

This libary was originally inspired by https://github.com/calavera/dkvolume.

### Usage

Note the custom URL:

```
go get go.pedge.io/dockervolume/...
```

And for imports:

```
import (
  "go.pedge.io/dockervolume"
)
```

Your volume plugin must implement the `VolumeDriver` interface.

To launch your plugin using Unix sockets, do:

```
func launch(volumeDriver dockervolume.VolumeDriver) error {
  return dockervolume.NewUnixServer(
    volumeDriver,
    "volume_driver_name",
    1050,
    "root",
    dockervolume.ServerOptions{},
  ).Serve()
}
```

To launch your plugin using TCP, do:

```
func launch(volumeDriver dockervolume.VolumeDriver) error {
  return dockervolume.NewTCPServer(
    volumeDriver,
    "volume_driver_name",
    1050,
    "address",
    dockervolume.ServerOptions{},
  ).Serve()
}
```

### Example

https://github.com/pachyderm/pachyderm/tree/master/src/cmd/pfs-volume-driver
