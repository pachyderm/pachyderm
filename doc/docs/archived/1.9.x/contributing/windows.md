# Developing Pachyderm in Windows with VSCode

## Install

* [go v1.13.x+](https://golang.org/dl/)
* [VSCode](https://code.visualstudio.com/download)
* [git](https://git-scm.com/download/win)
* [docker toolbox](https://github.com/docker/toolbox/releases)
* [VirtualBox](https://www.virtualbox.org/wiki/Downloads)
* [minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/#install-minikube-using-an-installer-executable)
* [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) - this is a raw exe, copy somewhere in PATH, like the minikube dir

## Configure

Add any VSCode extensions you may need for development, like `go` and `docker`.

Configure your terminal (opened via `` ctrl+` ``) to use `git bash` instead of `cmd`.  Add the following to your `settings.json` (depending on where your `git bash` actually exists):
```
  "terminal.integrated.shell.windows": "C:\\Program Files\\Git\\bin\\bash.exe",
```

## Getting started

* Clone the `pachyderm` repo into a local folder
  * In your terminal, run `git clone https://github.com/pachyderm/pachyderm`
* Launch the Docker Quickstart Terminal
  * This should show up in your start menu if you search 'docker'
  * You may need to do this any time your computer restarts
* Start minikube 
  * `minikube start --memory=8192mb`
  * The memory value can be tweaked, but I found the default (1 GB) to be too low for pachyderm compilation
* Pull relevant docker variables into your shell
  * `eval $(minikube docker-env --shell bash)`
  * This is only necessary if you want to run docker commands in a shell
* Build pachyderm `pachd` and `worker` images
  * Run the task `docker-build`
  * This can be done through the "Terminal > Run Task..." option, or by hitting `ctrl-p` and typing `task docker-build`
* Build and install pachctl and launch a pachyderm cluster
  * Run the task `launch-dev`
  * If the service does not come up promptly (the script never says all the pods are ready), you should check the 'Debugging' section below.

## Debugging

Determining the source of an error when launching a local pachyderm cluster can be difficult.  Some useful commands to get you started:

* `kubectl get all` - lists resources in the 'default' namespace (where we deploy locally)
* `kubectl logs -p <pod>` - gets the logs from the previous attempt at running a pod (usually this is where you will find errors)
* `minikube logs` - gets the logs from minikube itself, this may be useful if a pod ran into a `CreateContainerError`
* `docker container ls` - lists recently used or in-use docker containers, can be used to get logs more directly
* `docker logs <container>` - gets the logs from a specific docker container

### Past problems

For posterity's sake, here are some of the problems I've encountered when trying to get this working in Windows:

* Docker gets confused by command-line windows-style paths, it thinks the ':' is indicating a 'mode' and fails to parse.  In addition, Windows (or maybe bash) seems to automatically convert unix to windows-style paths, so you should export MSYS_NO_PATHCONV=1 to prevent this.
* Kubernetes resource specs (specifically `hostPath`) do not work if you use a windows-style path.  Instead, you must use a unix-style path where the drive letter is the first directory, e.g. '/C/path/to/file'.
* Etcd failed to mmap a file because it was in a directory shared with the host system.  Still not sure how I fixed this because it seems to be working now?

### Nuclear option

If your setup is completely fucked, it may be worthwhile to blow away your minikube and start over, this is pretty simple with:

```
minikube delete
minikube start --memory=8192mb
```
