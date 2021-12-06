import os

MOCK_PACHYDERM_SERVICE = os.environ.get("MOCK_PACHYDERM_SERVICE", False)
PFS_MOUNT_DIR = os.environ.get("PFS_MOUNT_DIR", "/pfs")
