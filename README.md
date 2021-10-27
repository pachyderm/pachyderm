# jupyterlab_pachyderm

[![Github Actions Status](https://github.com/pachyderm/jupyterlab-pachyderm/workflows/Build/badge.svg)](https://github.com/pachyderm/jupyterlab-pachyderm/actions/workflows/build.yml)

A JupyterLab extension.

This extension is composed of a Python package named `jupyterlab_pachyderm`
for the server extension and a NPM package named `jupyterlab-pachyderm`
for the frontend extension.

# Requirements

- JupyterLab >= 3.0

- Python 3.7
    - [pyenv](https://github.com/pyenv/pyenv) is a great way to manage and install different versions of python. You can check which version you are using by running `pyenv versions`. If you are upgrading the version of python used in the project, please check and make sure that the versions defined in the `.python-version`. 

- Node
    - If you are using [nvm](https://github.com/nvm-sh/nvm) first run `nvm install`. This will install and switch the version of node to the one defined in the `.nvmrc`. If you are upgrading the version of node used in the project, please check and make sure that the versions defined in the `.nvmrc`.


## Virtual Environment Setup

When developing in python, it is good practice to set up a virtual environment. A simple guid to set up a virtual environment is as follows:

create a virtual environment using venv

`python -m venv venv`

Activate the environment

`source venv/bin/activate`

When you are done using the environment you can close your shell or deactivate the environment: 
`deactivate`

### Development install

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
jlpm run build
```

You can watch the source directory and run JupyterLab at the same time in different terminals to watch for changes in the extension's source and automatically rebuild the extension.

```bash
# Watch the source directory in one terminal, automatically rebuilding when needed
jlpm run build
# Run JupyterLab in another terminal
jupyter lab
```

With the watch command running, every saved change will immediately be built locally and available in your running JupyterLab. Refresh JupyterLab to load the change in your browser (you may need to wait several seconds for the extension to be rebuilt).

By default, the `npm run build` command generates the source maps for this extension to make it easier to debug using the browser dev tools. To also generate source maps for the JupyterLab core extensions, you can run the following command:

```bash
jupyter lab build --minimize=False
```

### Development uninstall

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