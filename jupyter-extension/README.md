# jupyterlab-pachyderm

[![CircleCI](https://circleci.com/gh/pachyderm/jupyterlab-pachyderm/tree/main.svg?style=shield&circle-token=23e1645bde6312d903e50be2dec7073bf0bcfbd0)](https://circleci.com/gh/pachyderm/jupyterlab-pachyderm/tree/main)
[![PyPI version](https://badge.fury.io/py/jupyterlab-pachyderm.svg)](https://pypi.org/project/jupyterlab-pachyderm)

A JupyterLab extension for integrations with Pachyderm.

This extension is composed of a Python package named `jupyterlab_pachyderm`
for the server extension and a NPM package named `jupyterlab-pachyderm`
for the frontend extension.

# Requirements

- JupyterLab >= 3.0

- Python >=3.7,<4
    - [pyenv](https://github.com/pyenv/pyenv) is a great way to manage and install different versions of python. You can check which version you are using by running `pyenv versions`. Our Python extension is built to be compatible with Python versions 3.7 to 3.10. Therefore, it is best to run the lowest version (3.7.x) for highest compatibility.

- Node
    - If you are using [nvm](https://github.com/nvm-sh/nvm) first run `nvm install`. This will install and switch the version of node to the one defined in the `.nvmrc`. If you are upgrading the version of node used in the project, please check and make sure that the versions defined in the `.nvmrc`.


# Development Environment

There are two ways you can work on this project. One is setting up a local python virtual environment and the other is using a docker image developing on a virtual machine.

## Dev Container Workflow

Building and running the extension in a Docker container allows us to use mount, which is not possible in recent versions of macOS.

Assuming that you are running a Pachyderm instance somewhere, and that you can port-forward `pachd`

```
kubectl port-forward service/pachd 30650:30650
```

Start a bash session in the `pachyderm/notebooks-user` container from the top-level `jupyterlab-pachyderm` repo.

```
docker run --name jupyterlab_pachyderm_frontend_dev \
  --net=host \
  --rm \
  -it -e GRANT_SUDO=yes --user root \
  --device /dev/fuse --privileged \
  -v $(pwd):/home/jovyan/extension-wd \
  -w /home/jovyan/extension-wd \
  pachyderm/notebooks-user:77ce3a1ef2c73bf34d064cd0bdb5a64262bf3280 \
  bash
```

If you are running the frontend container on Linux, and want to be able to talk to pachd in minikube on `grpc://localhost:30650`, use `--net=host` at the start of the the `docker run` command as shown above. If you are on macOS, you will want to remove `--net=host` as it is not supported. On macOS you can specify `grpc://host.docker.internal:30650` to get the same effect.

Install the project in editable mode, and start JupyterLab

```
pip install -e .
jupyter labextension develop --overwrite

# Server extension must be manually installed in develop mode, for example
jupyter server extension enable jupyterlab_pachyderm

jupyter lab --allow-root
```

Open another bash inside the same container:

```
docker exec -it <container-id> bash
```

Within container run:

```
npm run watch
```

Iterating on the mount server, from inside a `pachyderm` checkout:
```
CGO_ENABLED=0 make install
docker cp /home/luke/gocode/bin/pachctl jupyterlab_pachyderm_frontend_dev:/usr/local/bin/pachctl
docker exec -ti jupyterlab_pachyderm_frontend_dev pkill -f pachctl
```

## Local Virtual Environment Setup 
When developing in python, it is good practice to set up a virtual environment. A simple guid to set up a virtual environment is as follows:
create a virtual environment using venv

`python -m venv venv`

Activate the environment

`source venv/bin/activate`

When you are done using the environment you can close your shell or deactivate the environment: 
`deactivate`

## Development install

Note: You will need NodeJS to build the extension package.

```bash
# Clone the repo to your local environment
# Change directory to the jupyterlab-pachyderm directory (top level of repo)
# Make sure you are using a virtual environment
# Install package in development mode
pip install -e .
# Link your development version of the extension with JupyterLab
jupyter labextension develop . --overwrite
# Server extension must be manually installed in develop mode
jupyter server extension enable jupyterlab_pachyderm
# Rebuild extension Typescript source after making changes
npm run build
```

You can watch the source directory and run JupyterLab at the same time in different terminals to watch for changes in the extension's source and automatically rebuild the extension.

```bash
# Watch the source directory in one terminal, automatically rebuilding when needed
npm run watch
# Run JupyterLab in another terminal
jupyter lab --allow-root
```

With the watch command running, every saved change will immediately be built locally and available in your running JupyterLab. Refresh JupyterLab to load the change in your browser (you may need to wait several seconds for the extension to be rebuilt).

By default, the `npm run build` command generates the source maps for this extension to make it easier to debug using the browser dev tools. To also generate source maps for the JupyterLab core extensions, you can run the following command:

```bash
jupyter lab build --minimize=False
```

## Development uninstall

```bash
# Server extension must be manually disabled in develop mode
jupyter server extension disable jupyterlab_pachyderm
pip uninstall jupyterlab_pachyderm
```

In development mode, you will also need to remove the symlink created by `jupyter labextension develop`
command. To find its location, you can run `jupyter labextension list` to figure out where the `labextensions`
folder is located. Then you can remove the symlink named `jupyterlab-pachyderm` within that folder.

## Locally building the docker image

Useful if iterating on the Dockerfile locally or iterating on changes to a version of mount-server.

Create & activate venv:
```
python3 -m venv venv
source venv/bin/activate
```

Build `dist` directory:
```
python -m pip install --upgrade pip
python -m pip install -r ci-requirements.txt
python -m build
```

Build docker image:
```
export PACHCTL_VERSION=aaa7434c714fab6130c3982ebdaa8f279bd850c2 # or whichever version you want
docker build --build-arg PACHCTL_VERSION=$PACHCTL_VERSION -t pachyderm/notebooks-user:dev .
docker run -ti -p 8888:8888 pachyderm/notebooks-user:dev
```

Navigate to the URL that's printed out by the docker container, then change `tree` to `lab` in your browser's address bar.

## Install

To install the extension, execute:

```bash
pip install jupyterlab_pachyderm
```

## Uninstall

To remove the extension, execute:

```bash
pip uninstall jupyterlab_pachyderm
```


## Server extension

First make sure the server extension is enabled:

```
jupyter server extension list 2>&1 | grep -ie "jupyterlab_pachyderm.*OK"
```

### API endpoints

Single

```
GET /repos/:repo # returns the state of a single repo
PUT /repos/:repo/_mount?name=foo&mode=w # mounts a single repo
PUT /repos/:repo/_unmount?name=foo # unmounts a single repo
PUT /repos/:repo/_commit # commits any changes to the repo
```

Batch

```
GET /repos # returns a list of all repos and their mount_state
```

The servers-side extension extends jupyter server, so it automatically starts as part of `jupyter lab`.

API endpoints can be accessed via `localhost:8888/pachyderm/v2`

The frontend can access the endpoints via `/v2`, for example:

```js
requestAPI<any>('/v2/repos')
      .then(data => {
        console.log(data);
      })
      .catch(reason => {
        console.error(reason);
      });
```

You can also access it via `localhost:8888/v2`

## SVG Images

We are leveraging [svgr](https://react-svgr.com/) to simplify the use of non icon svgs in in the project. Svg images that are to be converted live in the `svg-images` folder and get output to the `src/utils/components/Svgs` folder. If you want to add a new image to the project simply add the svg to the `svg-images` folder and run `npm run build:svg`. We have spent some time trying to get svgr to work through the `@svgr/webpack` plugin but have not been successful as of yet.

# Project Structure
Jupyter extensions are composed of several plugins. These plugins can be selectively enabled or disabled. Because of this we have decided separate the functionality in the extension using plugins. Plugins exported in this extension are as follows.

### Hub
This plugin contains custom styling and other features only used by hub. By default this extension is disabled.

### Mount
This plugin contains the mount feature currently under development.


## Plugin settings
You can disable certain plugins by specifying the following config data in the `<jupyter_config_path>/labconfig/page_config.json`:

```
{
  "disabledExtensions": {
    "jupyterlab-pachyderm:examples": true
  }
} 
```
You can check your config paths by running `jupyter --paths`. Setting this config file is not part of the built extension and needs to be done by the user.

Adding the following to the package.json in the `jupyterlab` object will disable the examples plugin by default.
```
"disabledExtensions": ["jupyterlab-pachyderm:examples"]
```
So we can build the extension with hub features turned off and override the setting for hub.


## Troubleshoot

If you are seeing the frontend extension, but it is not working, check
that the server extension is enabled:

```bash
jupyter server extension list
```

If the server extension is installed and enabled, but you are not seeing
the frontend extension, check the frontend extension is installed:

```bash
jupyter labextension list
```

## Contributing


### Packaging the extension

See [RELEASE](RELEASE.md)
