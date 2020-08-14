## What

This directory, and its subdirectories, contain a compatibility table between
`pachctl` releases and other pieces of software:

* This directory stores version compatibilities with dash. It is here for
historic reasons.
* The `ide` subdirectory stores version compatibilities with our IDE images.
* The `jupterhub` subdirectory stores version compatibilities with the
[JupyterHub helm chart](https://jupyterhub.github.io/helm-chart/) used by IDE. 

In these directories, each file (e.g. `1.11.0`) contains a list of versions of
software that it is compatible with.

Generally these files should not be manually modified, though they can be if
necessary.

## Why

This allows us to retrieve the latest dash image a user can use with their version of pachctl, enabling:

```
# Easy upgrades
pachctl update-dash
# Implicit upgrades each time deploy is used
pachctl deploy ...
```

Likewise with the IDE.
