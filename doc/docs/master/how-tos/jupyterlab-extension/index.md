
# Pachyderm JupyterLab Mount Extension

!!! Warning "Experimental"
    The JupyterLab Mount Extension is an [experimental feature](https://docs.pachyderm.com/latest/reference/supported-releases/#experimental){target=_blank}. We hope you'll try it out (and work with us to improve it! [Get in touch](https://www.pachyderm.com/slack/)), but it's not ready for self-service usage in production, as it may make sudden, breaking changes

![Mount extension in action](../images/mount-extension.gif)

Use the [JupyterLab extension](https://pypi.org/project/jupyterlab-pachyderm/) to:

- Connect your Notebook to a Pachyderm cluster.
- Browse, explore, analyze data stored in Pachyderm directly from your Notebook.
- Run and test out your pipeline code before creating a Docker image. 

## Before You Start 

- You must have a Pachyderm cluster running.
- You must install the Jupyterlab Helm repository.
    ```sh
    helm repo add jupyterhub https://jupyterhub.github.io/helm-chart/
    helm repo update
    ```
--- 

## Install The Extension 

You can choose between Pachyderm's pre-built image (a custom version of [`jupyter/scipy-notebook`](https://jupyter-docker-stacks.readthedocs.io/en/latest/using/selecting.html#jupyter-scipy-notebook)) or add the extension to your own image. Pachyderm's image includes:

- The extension jupyterlab-pachyderm.
- FUSE
- A pre-created `/pfs` directory that mounts to and grants ownership to the JupyterLab User
- A `mount-server` binary. 

### Local Installation (Using Provided Image)

1. Open your terminal.
2. Run the following:
    ```sh
    docker run -it -p 8888:8888 -e GRANT_SUDO=yes --user root --device /dev/fuse --privileged --entrypoint /opt/conda/bin/jupyter pachyderm/notebooks-user:{{ config.jupyterlab_extension_image_tag }}  lab --allow-root
    ```
3. Open the UI using the link provided in the terminal following:
    ```sh 
    Jupyter Server [...] is running at:
    ```
4. Navigate to the connection tab. You will need to provide a link formatted like the following:
    ```sh
    grpc://<cluster-ip>:<port>
    ```
5. Open another terminal and run the following to get the IP address and port number:
     ```sh 
     kubectl get services | grep -w "pachd "
     ```
6. Find the `servic/pachd` line item and copy the **IP address** and first **port number**.
    
    ```sh
       NAME                          TYPE           CLUSTER-IP       EXTERNAL-IP   PORT
       pachd                         ClusterIP      10.106.225.116   <none>        30650/TCP,30657/TCP,30658/TCP,30600/TCP,30656/TCP
    ```
7. Input the full connection URL (`grpc://10.106.225.116:30650`).
8. Navigate to the **Launcher** view in Jupyter and select **Terminal**.
9. Input the following command:
    ```sh
    pachctl version
    ```
10. If you see a `pachctl` and `pachd` version, you are good to go.

### Install to Existing Docker Image 

Replace the following `${PACHCTL_VERSION}` with the version of `pachctl` that matches your cluster's, and update `<version>` with the release number of the extension.

You can find the latest available version of our Pachyderm Mount Extension in [PyPi](https://pypi.org/project/jupyterlab-pachyderm/). 

```sh
# This runs the following section as root; if adding to an existing Dockerfile, set the user back to whatever you need. 
USER root

# This is the directory files will be mounted to, mirroring how pipelines are run. 
RUN mkdir -p /pfs 

# If you are not using "jovyan" as your notebook user, replace the user here. 
RUN chown $NB_USER /pfs

# Fuse is a requirement for the mount extension 
RUN apt-get clean && RUN apt-get update && apt-get -y install curl fuse 

# Install the mount-server binary
RUN curl -f -o mount-server.deb -L https://github.com/pachyderm/pachyderm/releases/download/v${PACHCTL_VERSION}/mount-server_${PACHCTL_VERSION}_amd64.deb
RUN dpkg -i mount-server.deb

# Optionally Install Pachctl - Set the version of Pachctl that matches your cluster deployment. 
RUN curl -f -o pachctl.deb -L https://github.com/pachyderm/pachyderm/releases/download/v${PACHCTL_VERSION}/pachctl_${PACHCTL_VERSION}_amd64.deb 
RUN dpkg -i pachctl.deb

# This sets the user back to the notebook user account (i.e., Jovyan) 
USER $NB_UID

# Replace the version here with the version of the extension you would like to install from https://pypi.org/project/jupyterlab-pachyderm/ 
RUN pip install jupyterlab-pachyderm==<version> 
```

Then, [build, tag, and push your image](../developer-workflow/working-with-pipelines/#step-2-build-your-docker-image).

### Install to JupyterHub With Helm

#### Option 1: Pachyderm Defaults

1. Open a terminal.
2. Run the following:
   ```sh
   helm upgrade --cleanup-on-fail \
   --install jupyter jupyterhub/jupyterhub \
   --values https://raw.githubusercontent.com/pachyderm/pachyderm/{{ config.pach_branch }}/etc/helm/examples/jupyterhub-ext-values.yaml
   ```
#### Option 2: Custom

1. Add the following to your Jupyterhub helm chart `values.YAML` file:
    ```yaml
    singleuser:
        defaultUrl: "/lab"
        cmd:   "start-singleuser.sh"
        image:
            name: pachyderm/notebooks-user
            tag: {{ config.jupyterlab_extension_image_tag }}
        uid:   0
        fsGid: 0
        extraEnv:
            "GRANT_SUDO": "yes"
            "NOTEBOOK_ARGS": "--allow-root"
            "JUPYTER_ENABLE_LAB": "yes"
            "CHOWN_HOME": "yes"
            "CHOWN_HOME_OPTS": "-R"
    hub:
        extraConfig:
            enableRoot: |
                from kubernetes import client
                def modify_pod_hook(spawner, pod):
                    pod.spec.containers[0].security_context = client.V1SecurityContext(
                        allow_privilege_escalation=True,
                        run_as_user=0,
                        privileged=True,
                        capabilities=client.V1Capabilities(
                            add=['SYS_ADMIN']
                        )
                    )
                    return pod
                c.KubeSpawner.modify_pod_hook = modify_pod_hook
    ```
2. Update the fields `singleuser.image.name` and `singleuser.image.tag` to match your user image.
3. Run the following commands to install JupyterHub:
    ```sh
    helm repo add jupyterhub https://jupyterhub.github.io/helm-chart/
    helm repo update

    helm upgrade --cleanup-on-fail \
    --install jupyter jupyterhub/jupyterhub \
    --values <your-jupyterhub-values.yaml>
    ```
4. Find the **IP address** you used to access the JupyterHub as described in these [Helm installation instructions](https://zero-to-jupyterhub.readthedocs.io/en/latest/jupyterhub#setup-jupyterhub) (Step 5 and 6) and open Jupyterlab.
5. Click on the link provided in the `stdout` of your terminal to run JupyterLab in a browser.
6. Connect to your cluster using the `grpc://<cluster-ip>:<port>` format.
7. Run the following command to refresh the mount server:

    ``` shell
    umount /pfs
    ```

##### Automate Cluster Details

You can specify your `pachd` cluster details in your Helm chart via `extraFiles` to avoid having to provide them every time Jupyter Hub starts. The `mountPath` input is required, however the location does not matter.

```yaml
singleuser:
    defaultUrl: "/lab"
    image:
        name: pachyderm/notebooks-user
        tag: <latest-release>
    extraEnv:
        "SIDECAR_MODE": "True"
    extraContainers:
        - name: mount-server-manager
          image: pachyderm/mount-server:<latest-release> 
          command: ["/bin/bash"]
          args: ["-c", "mkdir -p ~/.pachyderm && cp /config/config.json ~/.pachyderm && mount-server"]
          volumeMounts:
              - name: shared-pfs
                mountPath: /pfs
                mountPropagation: Bidirectional
              - name: files
                mountPath: /config
          securityContext:
              privileged: true
              runAsUser: 0
    storage:
        extraVolumeMounts:
            - name: shared-pfs
              mountPath: /pfs
              mountPropagation: HostToContainer
        extraVolumes:
            - name: shared-pfs
              emptyDir: {}
    extraFiles:
      config.json:
        mountPath: </any/path/file.json>
        data:
          v2:
            active_context: mount-server
            contexts:
              mount-server:
                source: 2
                pachd_address: <cluster_endpoint>
                server_cas: <b64e_cert_string>
                session_token: <token>
            metrics: true
```


<!--
// the following can be used if you do not know your cluster IP address (production)
```
docker run -it -v ~/.pachyderm/config.json:/home/jovyan/.pachyderm/config.json -p 8888:8888 -e GRANT_SUDO=yes --user root --device /dev/fuse --privileged --entrypoint /opt/conda/bin/jupyter pachyderm/notebooks-user:v0.6.0 lab --allow-root
``` 
-->

--- 

## How to Use 

### Mount a Branch

1. Open the Jupyterlab UI.
2. Open a **Terminal** from the launcher.
3. Navigate to the **Mounted Repositories** tab.
4. Input the following to see a demo repo appear:
    ```sh 
    pachctl create repo demo
    pachctl create branch demo@master
    ```
5. Scroll to the **Unmounted Repositories** section.
6. Select **Mount** next to the **Demo** repository. 
7. Input the following to create a simple text file:
    ```sh 
    echo "Version 1 of file!" | pachctl put file demo@master:/myfile.txt
    ```
8. You can read the file using the following:
    ```sh 
    cat /pfs/demo/myfile.txt
    ```

### Explore Directories & Files

At the bottom of the **Mounted Repositories** tab, you'll find the file browser. 

- Mounted repositories are nested within the root `/pfs` (Pachyderm's File System)
- These repositories are **read-only**
- Mounted repositories have a `/` glob pattern applied to their directories and files.
- Files only downloaded locally when you access them (saving you time).

Using the previous example, while the **Demo** repository is mounted, you can select the **demo** folder to reveal the example `myfile.txt`. 

### Examples 

Make sure to check our [data science notebook examples](https://github.com/pachyderm/examples){target=_blank} running on Pachyderm, from a market sentiment NLP implementation using a FinBERT model to pipelines training a regression model on the Boston Housing Dataset. You will also find integration examples with open-source products, such as labeling or model serving applications. 

--- 

## Troubleshooting 

 Restarting your server should resolve most issues. To restart your server, run the following command from the terminal window in Jupyterlab:

```shell
pkill -f "mount-server"
```
The server restarts by itself.

### M1 users with Docker Desktop < `4.6`

A [documented issue between qemu and Docker Desktop](https://gitlab.com/qemu-project/qemu/-/issues/340){target=_blank} prevents you from running our pre-built Mount Extension Image in Docker Desktop.

We recommend the following:

  - Use [Podman](https://podman.io){target=_blank} (See installation instructions)
  ```shell
  brew install podman
  podman machine init --disk-size 50
  podman machine start
  podman machine ssh
  sudo rpm-ostree install qemu-user-static && sudo systemctl reboot THEN
  ```
  then replace the keyword `docker` with `podman` in all the commands above. 
  - Or make sure that your qemu version is > `6.2`.

