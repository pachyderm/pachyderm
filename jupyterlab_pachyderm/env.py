import os

MOCK_PACHYDERM_SERVICE = os.environ.get("MOCK_PACHYDERM_SERVICE", 'False').lower() in ('true', '1', 't')
MOUNT_SERVER_ENABLED = os.environ.get("MOUNT_SERVER_ENABLED", 'True').lower() in ('true', '1', 't')
PFS_MOUNT_DIR = os.environ.get("PFS_MOUNT_DIR", "/pfs")
SIDECAR_MODE = os.environ.get("SIDECAR_MODE", 'False').lower() in ('true', '1', 't')
JUPYTERLAB_CI_TESTS = os.environ.get("JUPYTERLAB_CI_TESTS", 'False').lower() in ('true', '1', 't')

PACHYDERM_EXT_DEBUG = os.environ.get("PACHYDERM_EXT_DEBUG", 'False').lower() in ('true', '1', 't')
if PACHYDERM_EXT_DEBUG:
    from jupyterlab_pachyderm.log import get_logger

    logger = get_logger()
    logger.setLevel("DEBUG")
    logger.debug("DEBUG mode activated for pachyderm extension")
