[![CircleCI](https://circleci.com/gh/peter-edge/go-dockervolume/tree/master.png)](https://circleci.com/gh/peter-edge/go-dockervolume/tree/master)
[![GoDoc](http://img.shields.io/badge/GoDoc-Reference-blue.svg)](https://godoc.org/go.pedge.io/dockervolume)
[![MIT License](http://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/peter-edge/go-dockervolume/blob/master/LICENSE)

A small library taking care of the generic code for docker volume plugins written in Go.

* https://blog.docker.com/2015/06/extending-docker-with-plugins/
* https://github.com/docker/docker/blob/master/docs/extend/plugins.md
* https://github.com/docker/docker/blob/master/docs/extend/plugin_api.md
* https://github.com/docker/docker/blob/master/docs/extend/plugins_volume.md

This libary was originally inspired by https://github.com/calavera/dkvolume.

Note that most of my work now is going toward the [Open Storage Project](https://github.com/libopenstorage),
however this package is basically feature complete for the current docker volume plugin API.

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

All public functionality is exposed in the [dockervolume.go](dockervolume.go) file and
the generated [dockervolume.pb.go](dockervolume.pb.go) file.

Your volume plugin must implement the `VolumeDriver` interface.

The API in this package exposes additional functionality on top of the
docker volume plugin API. See [dockervolume.proto](dockervolume.proto) for more details.

To launch your plugin using Unix sockets, do:

```
func launch(volumeDriver dockervolume.VolumeDriver) error {
  return dockervolume.NewUnixServer(
    volumeDriver,
    "volume_driver_name",
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
    "address",
    dockervolume.ServerOptions{},
  ).Serve()
}
```

### Examples

* [example/cmd/dockervolume-example](example/cmd/dockervolume-example)
* [pfs-volume-driver](https://github.com/pachyderm/pachyderm/tree/master/src/cmd/pfs-volume-driver)
