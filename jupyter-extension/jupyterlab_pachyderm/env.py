import os
from pathlib import Path
from distutils.util import strtobool

PACH_CONFIG = os.environ.get("PACH_CONFIG", Path.home() / ".pachyderm/config.json")
PFS_MOUNT_DIR = os.environ.get("PFS_MOUNT_DIR", "/pfs")
PACHD_ADDRESS = os.environ.get("PACHD_ADDRESS", None)
DEX_TOKEN = os.environ.get("DEX_TOKEN", None)

PACHYDERM_EXT_DEBUG = strtobool(os.environ.get("PACHYDERM_EXT_DEBUG", "False").lower())
if PACHYDERM_EXT_DEBUG:
    from jupyterlab_pachyderm.log import get_logger

    logger = get_logger()
    logger.setLevel("DEBUG")
    logger.debug("DEBUG mode activated for pachyderm extension")
