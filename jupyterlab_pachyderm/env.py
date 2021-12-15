import os

MOCK_PACHYDERM_SERVICE = os.environ.get("MOCK_PACHYDERM_SERVICE", False)
MOUNT_SERVER_ENABLED = os.environ.get("MOUNT_SERVER_ENABLED", False)
PFS_MOUNT_DIR = os.environ.get("PFS_MOUNT_DIR", "/pfs")
