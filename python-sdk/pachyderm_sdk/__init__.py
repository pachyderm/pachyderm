from .client import Client

# Python version compatibility.
try:
    # >= 3.8
    import importlib.metadata as metadata
except ImportError:
    #  < 3.8
    import importlib_metadata as metadata  # type: ignore


__pdoc__ = {"proto": False}

__all__ = [
    "Client",
    "RpcError",
    "put_files",
    "PFSFile",
    "ModifyFileClient",
    "parse_json_pipeline_spec",
    "parse_dict_pipeline_spec",
    "ConfigError",
    "BadClusterDeploymentID",
]

__version__ = ""
try:
    __version__ = metadata.version(__name__)  # type: ignore
except (FileNotFoundError, ModuleNotFoundError):
    pass


from .api.pfs import _additions as __pfs_additions
