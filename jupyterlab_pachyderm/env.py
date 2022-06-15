import os

MOCK_PACHYDERM_SERVICE = os.environ.get("MOCK_PACHYDERM_SERVICE", False)
MOUNT_SERVER_ENABLED = os.environ.get("MOUNT_SERVER_ENABLED", True)
PFS_MOUNT_DIR = os.environ.get("PFS_MOUNT_DIR", "/pfs")
ENSURE_MOUNT_SERVER = bool(os.environ.get("ENSURE_MOUNT_SERVER", True))
JUPYTERLAB_CI_TESTS = bool(os.environ.get("JUPYTERLAB_CI_TESTS", False))

PACHYDERM_EXT_DEBUG = bool(os.environ.get("PACHYDERM_EXT_DEBUG", False))
if PACHYDERM_EXT_DEBUG:
    from jupyterlab_pachyderm.log import get_logger

    logger = get_logger()
    logger.setLevel("DEBUG")
    logger.debug("DEBUG mode activated for pachyderm extension")
