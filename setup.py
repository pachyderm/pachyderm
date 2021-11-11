"""
jupyterlab_pachyderm setup
"""
import json
import sys
from pathlib import Path

import setuptools

HERE = Path(__file__).parent.resolve()

# The name of the project
name = "jupyterlab_pachyderm"

lab_path = (HERE / name.replace("-", "_") / "labextension")

# Representative files that should exist after a successful build
ensured_targets = [
    str(lab_path / "package.json"),
    str(lab_path / "static/style.js")
]

labext_name = "jupyterlab-pachyderm"

data_files_spec = [
    ("share/jupyter/labextensions/%s" % labext_name, str(lab_path.relative_to(HERE)), "**"),
    ("share/jupyter/labextensions/%s" % labext_name, str("."), "install.json"),
    ("etc/jupyter/jupyter_server_config.d",
     "jupyter-config/server-config", "jupyterlab_pachyderm.json"),
    # For backward compatibility with notebook server
    ("etc/jupyter/jupyter_notebook_config.d",
     "jupyter-config/nb-config", "jupyterlab_pachyderm.json"),
]

long_description = (HERE / "README.md").read_text()

# Get the package info from package.json
pkg_json = json.loads((HERE / "package.json").read_bytes())
version = (
    pkg_json["version"]
    .replace("-alpha.", "a")
    .replace("-beta.", "b")
    .replace("-rc.", "rc")
) 

setup_args = dict(
    name=name,
    version=version,
    url=pkg_json["homepage"],
    author=pkg_json["author"]["name"],
    author_email=pkg_json["author"]["email"],
    description=pkg_json["description"],
    license=pkg_json["license"],
    license_file="LICENSE",
    keywords=pkg_json["keywords"],
)

try:
    from jupyter_packaging import (
        wrap_installers,
        npm_builder,
        get_data_files
    )
    post_develop = npm_builder(
        build_cmd="install:extension", source_dir="src", build_dir=lab_path
    )
    setup_args["cmdclass"] = wrap_installers(post_develop=post_develop, ensured_targets=ensured_targets)
    setup_args["data_files"] = get_data_files(data_files_spec)
except ImportError as e:
    import logging
    logging.basicConfig(format="%(levelname)s: %(message)s")
    logging.warning("Build tool `jupyter-packaging` is missing. Install it with pip or conda.")
    if not ("--name" in sys.argv or "--version" in sys.argv):
        raise e

if __name__ == "__main__":
    setuptools.setup(**setup_args)
