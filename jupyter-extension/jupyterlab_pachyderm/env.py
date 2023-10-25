import os
from pathlib import Path
from distutils.util import strtobool

HTTP_UNIX_SOCKET_SCHEMA="http_unix"
HTTP_SCHEMA="http"
PFS_MOUNT_DIR = os.environ.get("PFS_MOUNT_DIR", "/pfs")
PFS_SOCK_PATH = os.environ.get("PFS_SOCK_PATH", "/tmp/pfs.sock")
DEFAULT_SCHEMA = os.environ.get("DEFAULT_SCHEMA", HTTP_UNIX_SOCKET_SCHEMA)
SIDECAR_MODE = strtobool(os.environ.get("SIDECAR_MODE", "False").lower())
MOUNT_SERVER_LOG_FILE = os.environ.get("MOUNT_SERVER_LOG_FILE", "") # defaults to stdout/stderr
# TODO(msteffen) Maybe depend on Pachyderm SDK for this
PACH_CONFIG = os.environ.get("PACH_CONFIG", Path.home() / ".pachyderm/config.json")
NONPRIV_CONTAINER = strtobool(os.environ.get("NONPRIV_CONTAINER", "False").lower()) # Unset assume --privileged container
DET_RESOURCES_TYPE = os.environ.get("DET_RESOURCES_TYPE") # Assume MLDE if set
SLURM_JOB="slurm-job" # For launcher this is set in DispatcherRM

PACHYDERM_EXT_DEBUG = strtobool(os.environ.get("PACHYDERM_EXT_DEBUG", "False").lower())
if PACHYDERM_EXT_DEBUG:
    from jupyterlab_pachyderm.log import get_logger

    logger = get_logger()
    logger.setLevel("DEBUG")
    logger.debug("DEBUG mode activated for pachyderm extension")
