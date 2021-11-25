import json
from pathlib import Path
import os

import python_pachyderm

from ._version import __version__
from .pachyderm import PachydermClient, PachydermMountClient
from .mock_pachyderm import MockPachydermClient
from .handlers import setup_handlers


HERE = Path(__file__).parent.resolve()

with (HERE / "labextension" / "package.json").open() as fid:
    data = json.load(fid)


def _jupyter_labextension_paths():
    return [{"src": "labextension", "dest": data["name"]}]


def _jupyter_server_extension_points():
    return [{"module": "jupyterlab_pachyderm"}]


def _load_jupyter_server_extension(server_app):
    """Registers the API handler to receive HTTP requests from the frontend extension.

    Parameters
    ----------
    server_app: jupyterlab.labapp.LabApp
        JupyterLab application instance
    """
    # swap real PachydermMountServide with mock given MOCK_PACHYDERM_SERVICE
    if "MOCK_PACHYDERM_SERVICE" in os.environ:
        server_app.log.info("Mock Pachyderm API selected")
        server_app.web_app.settings["PachydermMountClient"] = PachydermMountClient(
            MockPachydermClient(), "/pfs"
        )
    else:
        server_app.web_app.settings["PachydermMountClient"] = PachydermMountClient(
            PachydermClient(python_pachyderm.Client(), python_pachyderm.ExperimentalClient()), "/pfs"
        )

    setup_handlers(server_app.web_app)
    server_app.log.info("Registered Pachyderm extension at URL path /pachyderm")


# For backward compatibility with notebook server - useful for Binder/JupyterHub
load_jupyter_server_extension = _load_jupyter_server_extension
