"""Handwritten classes/methods that augment the existing Admin API."""
import pachyderm_sdk

from . import ApiStub as _GeneratedApiStub
from ..version import Version
from . import ClusterInfo


class ApiStub(_GeneratedApiStub):
    # noinspection PyMethodOverriding
    def inspect_cluster(self) -> "ClusterInfo":
        """Inspect the cluster and check version mismatch."""
        try:
            version = _extract_version(pachyderm_sdk.__version__)
        except ValueError:
            return super().inspect_cluster()
        return super().inspect_cluster(client_version=version)


def _extract_version(version: str) -> Version:
    """Attempt to create an api.Version object.

    Raises:
        ValueError: If the version is invalid.
    """
    major, minor, patch = version.split(".", 2)
    additional = None
    if "-" in patch:
        patch, additional = patch.split("-", 1)
    major, minor, patch = int(major), int(minor), int(patch)
    return Version(major=major, minor=minor, micro=patch, additional=additional)
