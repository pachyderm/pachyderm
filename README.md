# jupyterlab_pachyderm

[![Github Actions Status](https://github.com/pachyderm/jupyterlab-pachyderm/workflows/Build/badge.svg)](https://github.com/pachyderm/jupyterlab-pachyderm/actions/workflows/build.yml)

A JupyterLab extension.

This extension is composed of a Python package named `jupyterlab_pachyderm`
for the server extension and a NPM package named `jupyterlab-pachyderm`
for the frontend extension.

# Requirements

- JupyterLab >= 3.0

- Python >=3.6,<4
    - [pyenv](https://github.com/pyenv/pyenv) is a great way to manage and install different versions of python. You can check which version you are using by running `pyenv versions`. Our Python extension is built to be compatible with Python versions 3.6 to 3.9. Therefore, it is best to run the lowest version (3.6.x) for highest compatibility.

- Node
    - If you are using [nvm](https://github.com/nvm-sh/nvm) first run `nvm install`. This will install and switch the version of node to the one defined in the `.nvmrc`. If you are upgrading the version of node used in the project, please check and make sure that the versions defined in the `.nvmrc`.


# Development Environment
There are two ways you can work on this project. One is setting up a local python virtual environment and the other is using a docker image developing on a virtual machine.

## Dev Container Workflow
You can also set up your local dev environment using a docker image ([Guide here](https://github.com/pachyderm/pachyderm-notebooks/blob/ide/doc/extensions_dev_container_workflow.md)). This way you do not need to worry about setting up the python virtual environment. Using this virtual development environment also allows you to use the pachyderm mount feature if you are developing on a mac.

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

The `npm` command is JupyterLab's pinned version of
[yarn](https://yarnpkg.com/) that is installed with JupyterLab. You may use
`yarn` or `npm` in lieu of `npm` below.

```bash
# Clone the repo to your local environment
# Change directory to the jupyterlab_pachyderm directory
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
jupyter lab
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
GET /repos/:repo_id # returns the state of a single repo
PUT /repos/:repo_id/_mount # mounts a single repo
PUT /repos/:repo_id/_unmount # unmounts a single repo
PUT /repos/:repo_id/_commit # commits any changes to the repo
```

Batch

```
GET /repos # returns a list of all repos and their mount_state
PUT /repos/_mount # mounts all repos by default unless request body contains a list of specific repos
PUT /repos/_unmount # unmounts all repos
```

The servers-side extension extends jupyter server, so it automatically starts as part of `jupyter lab`.

API endpoints can be accessed via `localhost:8888/pachyderm/v1`

The frontend can access the endpoints via `/v1`, for example:

```js
requestAPI<any>('/v1/repos')
      .then(data => {
        console.log(data);
      })
      .catch(reason => {
        console.error(reason);
      });
```

You can also access it via `localhost:8888/v1


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
    "jupyterlab-pachyderm:hub": true
  }
} 
```
Setting this config file is not part of the built extension and needs to be done by the user.

Adding the following to the package.json in the `jupyterlab` object will disable the hub plugin by default.
```
"disabledExtensions": ["jupyterlab-pachyderm:hub"]
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
