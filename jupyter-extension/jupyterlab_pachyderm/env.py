import os
from pathlib import Path
from distutils.util import strtobool

from pachyderm_sdk.constants import CONFIG_PATH_LOCAL

PACH_CONFIG = Path(
    os.path.expanduser(os.environ.get("PACH_CONFIG", CONFIG_PATH_LOCAL))
).resolve()
PFS_MOUNT_DIR = os.environ.get("PFS_MOUNT_DIR", "/pfs")

PACHYDERM_EXT_DEBUG = strtobool(os.environ.get("PACHYDERM_EXT_DEBUG", "False").lower())
if PACHYDERM_EXT_DEBUG:
    from jupyterlab_pachyderm.log import get_logger

    logger = get_logger()
    logger.setLevel("DEBUG")
    logger.debug("DEBUG mode activated for pachyderm extension")
