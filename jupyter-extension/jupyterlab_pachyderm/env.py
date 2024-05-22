import os
from pathlib import Path
from distutils.util import strtobool

from pachyderm_sdk.constants import CONFIG_PATH_LOCAL

# JUPYTER_SERVER_ROOT is the root path from which all data and notebook files
#   are relative to. This may be different from the CWD which, if different,
#   is specified under JUPYTER_RUNTIME_DIR.
JUPYTER_SERVER_ROOT = Path(os.environ["JUPYTER_SERVER_ROOT"])

PACH_CONFIG = Path(
    os.path.expanduser(os.environ.get("PACH_CONFIG", CONFIG_PATH_LOCAL))
).resolve()
PFS_MOUNT_DIR = os.environ.get("PFS_MOUNT_DIR", "/pfs")
PACHD_ADDRESS = os.environ.get("PACHD_ADDRESS", None)
# If specified, this is the ID_TOKEN used in auth
DEX_TOKEN = os.environ.get("DEX_TOKEN", None)

PACHYDERM_EXT_DEBUG = strtobool(os.environ.get("PACHYDERM_EXT_DEBUG", "False").lower())
if PACHYDERM_EXT_DEBUG:
    from jupyterlab_pachyderm.log import get_logger

    logger = get_logger()
    logger.setLevel("DEBUG")
    logger.debug("DEBUG mode activated for pachyderm extension")
