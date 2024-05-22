import importlib.metadata as metadata

__version__ = ""
try:
    __version__ = metadata.version(__name__)  # type: ignore
except (FileNotFoundError, ModuleNotFoundError):
    pass

from .api.pfs import _additions as __pfs_additions
from .client import Client
from .datum_batching import batch_all_datums

__all__ = [
    "Client",
    "batch_all_datums",
]
