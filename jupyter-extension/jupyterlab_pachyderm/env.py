import os
from pathlib import Path
from distutils.util import strtobool

HTTP_UNIX_SOCKET_SCHEMA = "http_unix"
HTTP_SCHEMA = "http"
DEFAULT_SCHEMA = os.environ.get("DEFAULT_SCHEMA", HTTP_UNIX_SOCKET_SCHEMA)
PACH_CONFIG = os.environ.get("PACH_CONFIG", Path.home() / ".pachyderm/config.json")

PACHYDERM_EXT_DEBUG = strtobool(os.environ.get("PACHYDERM_EXT_DEBUG", "False").lower())
if PACHYDERM_EXT_DEBUG:
    from jupyterlab_pachyderm.log import get_logger

    logger = get_logger()
    logger.setLevel("DEBUG")
    logger.debug("DEBUG mode activated for pachyderm extension")
