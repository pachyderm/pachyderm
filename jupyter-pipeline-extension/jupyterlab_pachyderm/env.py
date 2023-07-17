import os
from distutils.util import strtobool

PFS_MOUNT_DIR = os.environ.get("PFS_MOUNT_DIR", "/pfs")
SIDECAR_MODE = strtobool(os.environ.get("SIDECAR_MODE", "False").lower())
MOUNT_SERVER_LOG_DIR = os.environ.get("MOUNT_SERVER_LOG_DIR") # defaults to stdout/stderr
NONPRIV_CONTAINER = os.environ.get("NONPRIV_CONTAINER") # Unset assume --privileged container


PACHYDERM_EXT_DEBUG = strtobool(os.environ.get("PACHYDERM_EXT_DEBUG", "False").lower())
if PACHYDERM_EXT_DEBUG:
    from jupyterlab_pachyderm.log import get_logger

    logger = get_logger()
    logger.setLevel("DEBUG")
    logger.debug("DEBUG mode activated for pachyderm extension")
