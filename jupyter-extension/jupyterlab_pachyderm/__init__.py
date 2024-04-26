import json
from pathlib import Path

from .env import PACH_CONFIG, PACHD_ADDRESS, DEX_TOKEN
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
    setup_handlers(server_app.web_app, PACH_CONFIG, PACHD_ADDRESS, DEX_TOKEN)
    server_app.log.info("Registered Pachyderm extension at URL path /pachyderm")


# For backward compatibility with notebook server - useful for Binder/JupyterHub
load_jupyter_server_extension = _load_jupyter_server_extension
