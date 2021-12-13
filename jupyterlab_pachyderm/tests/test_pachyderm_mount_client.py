import pytest

from jupyterlab_pachyderm.handlers import _parse_pfs_path
from jupyterlab_pachyderm.pachyderm import PachydermMountClient
from jupyterlab_pachyderm.mock_pachyderm import MockPachydermClient

MOUNT_DIR = "/pfs"


def test_parse_pfs_path():
    assert _parse_pfs_path("repo") == ("repo", "master", None)
    assert _parse_pfs_path("repo/branch") == ("repo", "branch", None)
    assert _parse_pfs_path("repo/branch/commit") == ("repo", "branch", "commit")


class TestPachydermMountClient:
    @pytest.fixture()
    def client(self) -> PachydermMountClient:
        c = PachydermMountClient(MockPachydermClient(), MOUNT_DIR)
        return c

    def test_list(self, client: PachydermMountClient):
        assert client.list() == {
            "edges": {
                "branches": {
                    "master": {
                        "mount": {
                            "mode": None,
                            "mountpoint": None,
                            "name": None,
                            "state": "unmounted",
                            "status": None,
                        }
                    }
                }
            },
            "images": {
                "branches": {
                    "master": {
                        "mount": {
                            "mode": None,
                            "mountpoint": None,
                            "name": None,
                            "state": "unmounted",
                            "status": None,
                        }
                    }
                }
            },
            "montage": {
                "branches": {
                    "master": {
                        "mount": {
                            "mode": None,
                            "mountpoint": None,
                            "name": None,
                            "state": "unmounted",
                            "status": None,
                        }
                    }
                }
            },
        }

    def test_list_updates_itself(self, client: PachydermMountClient):
        assert client.list().keys() == {"edges", "images", "montage"}

        client.client._create_repos([{"repo": "new_repo", "branches": ["master"]}])
        assert client.list().keys() == {
            "edges",
            "images",
            "montage",
            "new_repo",
        }

    def test_list_when_repo_is_removed(self, client: PachydermMountClient):
        assert client.list().keys() == {"edges", "images", "montage"}
        client.client._delete_repo("images")
        assert client.list().keys() == {"edges", "montage"}

    def test_current_mount_strings(self, client: PachydermMountClient):
        assert client._current_mount_strings() == []

        # make sure ro -> r for pachctl
        client.mount("images", "master", "ro")
        assert client._current_mount_strings() == ["images@master+r"]

        # idempotent
        client.mount("images", "master", "ro")
        assert client._current_mount_strings() == ["images@master+r"]

        # r also works
        client.unmount("images", "master")
        client.mount("images", "master", "r")
        assert client._current_mount_strings() == ["images@master+r"]

        # make sure rw -> w for pachctl
        client.unmount("images", "master")
        client.mount("images", "master", "rw")
        assert client._current_mount_strings() == ["images@master+w"]

        # add another repo
        client.mount("edges", "master", "ro")
        assert sorted(client._current_mount_strings()) == [
            "edges@master+r",
            "images@master+w",
        ]

    def test_mount(self, client: PachydermMountClient):
        assert client.mount("images", "master", "r") == {
            "repo": "images",
            "branch": "master",
            "mount": {
                "name": "images",
                "mode": "r",
                "state": "mounted",
                "status": None,
                "mountpoint": "/pfs/images",
            },
        }

    def test_mount_repo_noexist(self, client: PachydermMountClient):
        assert client.mount("noe_repo", "master", "r") == {}

    def test_unmount(self, client: PachydermMountClient):
        client.mount("images", "master", "r")
        assert client.unmount("images", "master", "r") == {
            "repo": "images",
            "branch": "master",
            "mount": {"state": "unmounted"},
        }

    def test_unmount_all(self, client: PachydermMountClient):
        client.mount("images", "master", "rw")
        client.mount("edges", "master", "ro")
        assert client.unmount_all() == [
            {"repo": "images", "branch": "master", "mount": {"state": "unmounted"}},
            {"repo": "edges", "branch": "master", "mount": {"state": "unmounted"}},
        ]
