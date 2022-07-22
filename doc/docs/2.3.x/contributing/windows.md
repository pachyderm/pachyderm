# Developing Pachyderm in Windows with VSCode

## Before You Start 

### Installation Requirements

You must have all of the following installed before you can start development:

- [Docker Desktop](https://www.docker.com/products/docker-desktop/)
- [Go v1.15.x+](https://go.dev/dl/)
- [GoReleaser](https://github.com/goreleaser/goreleaser/releases)
- [Git](https://git-scm.com/download/win)
- [HyperV](https://docs.microsoft.com/en-us/virtualization/hyper-v-on-windows/quick-start/enable-hyper-v)
- [jq](https://stedolan.github.io/jq/download/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Make](http://gnuwin32.sourceforge.net/packages/make.htm)
- [minikube](https://minikube.sigs.k8s.io/docs/start/)
- [ShellCheck](https://github.com/koalaman/shellcheck#user-content-installing)
- [VSCode](https://code.visualstudio.com/download)

### Terminal Settings

1. Open VS Code.
2. Open your terminal (`` ctrl+` ``).
3. Add the following to your `settings.json`: 
     ```
       "terminal.integrated.shell.windows": "C:\\Program Files\\Git\\bin\\bash.exe",
     ``` 
This path may vary depending on where your `git bash` actually exists. 

---

## Getting started

1. Open a terminal and navigate to a directory you'd like to store Pachyderm.
2. Clone the pachyderm repo using `git clone https://github.com/pachyderm/pachyderm`.
3. Launch Docker Desktop (with Kubernetes enabled) or start minikube.  
4. Provision  ~10 GB of memory and ~4CPUs.
      - **Via minikube**:  `minikube start --memory=10000mb --cpus=4 --disk-size=40000mb --driver=hyperv`
      - **Via Docker Desktop**: Open Docker Desktop and navigate to **Preferences** > **Resources** > **Advanced**. 
5. Build your pachyderm `pachd` and `worker` images via the task `docker-build`.
      - **Option 1**: Navigate to **Terminal** > **Run Task...**
      - **Option 2**: Press `ctrl+p` and input `task docker-build`
6. Build and install pachctl.
7. Launch a Pachyderm cluster by running the task `launch-dev`.  

 If the service does not come up promptly (the script never says all the pods are ready), see the Debugging section.

 ---

## Debugging

### Common Commands

The following commands are used frequently when working with Pachyderm: 

- `kubectl get all`: lists resources in the 'default' namespace, where we deploy locally. 
- `kubectl logs -p <pod>`: gets the logs from the previous attempt at running a pod; a good place to find errors.
- `minikube logs`: gets the logs from minikube itself, useful when a pod runs into a `CreateContainerError`.
- `docker container ls`: lists recently used or in-use docker containers; used to get logs more directly.
- `docker logs <container>`: gets the logs from a specific docker container.

### Gotchas 

- Docker can get confused by command-line windows-style paths; it reads `:` as a **mode** and fails to parse.
- You may want to export `MSYS_NO_PATHCONV=1` to prevent the automated conversion of unix-to-windows paths.
- Kubernetes resource specs (specifically `hostPath`) do not work if you use a windows-style path.  Instead, you must use a unix-style path where the drive letter is the first directory, e.g. `/C/path/to/file`.
- Etcd may fail to mmap files when in a directory shared with the host system. 


### Full Restart

#### Minikube 
If you'd like to completely restart, use the following terminal commands:

```
minikube delete
kubectl delete pvc -l suite=pachyderm 
minikube start --memory=10000mb --cpus=4 --disk-size=40000mb
```
