# dockervolume

A small library taking care of the generic code for docker volume plugins written in Go.

* https://blog.docker.com/2015/06/extending-docker-with-plugins/
* https://github.com/docker/docker/blob/master/docs/extend/plugins.md
* https://github.com/docker/docker/blob/master/docs/extend/plugin_api.md
* https://github.com/docker/docker/blob/master/docs/extend/plugins_volume.md

This libary was originally inspired by https://github.com/calavera/dkvolume, but exposes
similar functionality in a slightly different style, along with adding logging of all
calls by default using [go.pedge.io/protolog](http://go.pedge.io/protolog).

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
  return dockervolume.Serve(
    dockervolume.NewHandler(
      volumeDriver,
      dockervolume.HandlerOptions{},
    ),
    dockervolume.ProtocolUnix,
    "volume_driver_name",
    "root",
  )
}
```

To launch your plugin using TCP, do:

```
func launch(volumeDriver dockervolume.VolumeDriver) error {
  return dockervolume.Serve(
    dockervolume.NewHandler(
      volumeDriver,
      dockervolume.HandlerOptions{},
    ),
    dockervolume.ProtocolTCP,
    "volume_driver_name",
    "address",
  )
}
```

### Example

https://github.com/pachyderm/pachyderm/tree/master/src/cmd/pfs-volume-driver
