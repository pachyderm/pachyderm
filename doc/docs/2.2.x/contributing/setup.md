# Developer Setup Guide 

## 1. Set Up Go Workspace

Already have a Go workspace set up? Skip to the download section. 

1. [Install Go](https://go.dev/doc/install).
2. Create a workspace directory (for example, `/Documents/GitHub/go-workspace`).
3. Open your `/go-workspace` directory and create the following sub-directories:
      - `/src`
      - `/pkg`
      - `/bin` 
4. Open a terminal. 
5. Navigate to your home folder's `~/.zshrc` file (`nano ~/.zshrc`).
6. Define your Go settings:
    ```
    export GOROOT="/usr/local/go"
    export GOPATH="$HOME/Documents/GitHub`/go-workspace"

    source $GOPATH/src/pachyderm/etc/contributing/bash_helpers
    ```
7. Save the file (`ctr-x` > `y`).

### Download Pachyderm

1. Open a terminal.
2. Navigate to your workspace's `/src` director, such as `~/Documents/GitHub/go-workspace/src`.
3. Clone the repo using `https://github.com/pachyderm/pachyderm.git`.


### Increase File Descriptor Limit (MacOs)

Running local tests requires an update to your file descriptor limit.

1.  Open a terminal.
2.  Run the following to set up a LaunchDaemon:
   ```shell
   sudo cp ~/go-workspace/src/pachyderm/etc/contributing/com.apple.launchd.limit.plist /Library/LaunchDaemons/
   ```
3.  Check your limits:
    ```shell
    launchctl limit maxfiles
    ```
4.  Open your `/.zshrc` file an increase the per-process limit by adding:
    ```shell
     ulimit -n 12288
    ```
5.  Navigate to your `/pachyderm` directory.
6.  Test your setup:
    ```shell
    make test-pfs-server
    ```

If this fails with a timeout, you'll probably also see 'too many files' type of errors. If that test passes, you're all good!

### Install Coreutils

To run the makefile `make launch` task, you will need the `timeout` utility found in [Coreutils](https://www.gnu.org/software/coreutils/).

1. Open a terminal.
2. Run the following:
    ```shell
    brew install coreutils
    ```
3. Open your `/.zshrc` file and add:
   ```shell
   PATH="/usr/local/opt/coreutils/libexec/gnubin:$PATH"
   ```
### Install Pipe Viewer (PV)

1. Open a terminal.
2. Input the following command:
   ```shell
   brew install pv 
   ```
See the [Pipe Viewer documentation](https://www.ivarch.com/programs/pv.shtml) for details on this monitoring tool.

--- 

## 2. Launch Development Cluster

### Update Binary Access & Install Latest pachctl 

1. Open a terminal.
2. Navigate to your `/pachyderm` directory.
3. Run `make install-clean` 

### Launch Cluster

1. Open a terminal. 
2. Navigate into your `/pachyderm` directory.
3. Run the following Makefile command:
   ```shell
   make launch-dev-vm 
   ```
4. Check the status:
    ```shell 
    kubectl get all
    ```
You can also run `make clean-launch-kube`.
---

## 3. Load Images to Minikube Cluster

Pachyderm's tests rely on a few images you'll need to load the into Minikube cluster.

1. Open a terminal. 
2. Navigate into your `/pachyderm` directory.
3. Install for the `pachyderm_entrypoint` container:
    ```shell
    make docker-build-test-entrypoint
    ./etc/kube/push-to-minikube.sh pachyderm_entrypoint
    ```
4. Install for the `pachyderm/python-build` container: 
    ```shell
    (cd etc/pipeline-build; make push-to-minikube)
    ```
---

## How to Run Tests

You can run specific tests by using `go test` directly. For example: 

   ```shell
   go test -v ./src/server/cmd/pachctl/cmd
   ```

Running *all* of your tests locally can take a while; instead, use [CircleCI](https://circleci.com/). 

---

## How to Fully Reset Pachyderm Environment 

Instead of having to run makefile targets to recompile `pachctl`and redeploy a development cluster, you can use the following script to:

- Delete all existing cluster data
- Wipe the VM the cluster is running on 
- Re-compile `pachctl` 
- Re-deploy the development cluster 

This reset is a bit more time consuming than running one-off Makefile targets,
but comprehensively ensures that the cluster is in its expected state.

To run it, simply call `./etc/reset.py` from the pachyderm repo root.