# pfs-volume-driver

This is a docker volume plugin for PFS.

Some good reading:

* https://blog.docker.com/2015/06/extending-docker-with-plugins/
* https://github.com/docker/docker/blob/master/docs/extend/plugins.md
* https://github.com/docker/docker/blob/master/docs/extend/plugin_api.md
* https://github.com/docker/docker/blob/master/docs/extend/plugins_volume.md

### Status

Very early alpha. Have fun!

### How to use

Docker volume plugins are supposed to be started before the docker daemon starts, and
stopped after the docker daemon stops. In practice, when doing development, you can
start and stop the volume plugin IF you guarantee that you have cleaned up all state,
ie you have removed any volumes you created before stopping the volume plugin. With
that said, here's a strategy to test this out:

First, launch PFS:

```
make launch-pfsd
```

Then get the pfs-volume-driver binary:

```
make docker-build-pfs-volume-driver
docker create --name pfs-volume-driver pachyderm/pfs-volume-driver:latest
docker cp pfs-volume-driver:/pfs-volume-driver .
```

Copy this binary to *the docker host*. This binary needs to run where docker is running.
Then, go to that machine and run the binary:

```
sudo /path/to/pfs-volume-driver
```

This assumes that pfsd (launched with `make launch-pfsd`) is running on the same machine,
otherwise you must set `PFS_ADDRESS` to point to the host and port of pfsd.

A good example of how to use this is at [etc/examples/pfs-simple.bash](../../../etc/examples/pfs-simple.bash).
Note that you must have `PFS_ADDRESS` set if you are not running the example on the same machine
as docker is running. If using docker-machine, you'll need to open up port 650 and then set
`PFS_ADDRESS` accordingly, for example `192.168.10.10:650`. You also need to install the pfs CLI tool
before running this example:

```
make install
```

### Questions?

File an issue or get in contact with us, we are happy to help :)
