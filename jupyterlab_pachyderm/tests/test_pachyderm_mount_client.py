import pytest

from jupyterlab_pachyderm.handlers import _parse_pfs_path
from jupyterlab_pachyderm.pachyderm import PythonPachydermMountClient
from jupyterlab_pachyderm.mock_pachyderm import MockPachydermClient

MOUNT_DIR = "/pfs"


def test_parse_pfs_path():
    assert _parse_pfs_path("repo") == ("repo", "master", None)
    assert _parse_pfs_path("repo/branch") == ("repo", "branch", None)
    assert _parse_pfs_path("repo/branch/commit") == ("repo", "branch", "commit")


class TestPachydermMountClient:
    @pytest.fixture()
    def client(self) -> PythonPachydermMountClient:
        c = PythonPachydermMountClient(MockPachydermClient(), MOUNT_DIR)
        return c

    async def test_list(self, client: PythonPachydermMountClient):
        assert await client.list() == {
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

    async def test_list_updates_itself(self, client: PythonPachydermMountClient):
        repos = await client.list()
        assert repos.keys() == {"edges", "images", "montage"}

        client.client._create_repos([{"repo": "new_repo", "branches": ["master"]}])
        repos = await client.list()
        assert repos.keys() == {
            "edges",
            "images",
            "montage",
            "new_repo",
        }

    async def test_list_when_repo_is_removed(self, client: PythonPachydermMountClient):
        repos = await client.list()
        assert repos.keys() == {"edges", "images", "montage"}
        client.client._delete_repo("images")
        repos = await client.list()
        assert repos.keys() == {"edges", "montage"}

    async def test_current_mount_strings(self, client: PythonPachydermMountClient):
        assert client._current_mount_strings() == []

        # make sure ro -> r for pachctl
        await client.mount("images", "master", "ro")
        assert client._current_mount_strings() == ["images@master+r"]

        # idempotent
        await client.mount("images", "master", "ro")
        assert client._current_mount_strings() == ["images@master+r"]

        # r also works
        await client.unmount("images", "master")
        await client.mount("images", "master", "r")
        assert client._current_mount_strings() == ["images@master+r"]

        # make sure rw -> w for pachctl
        await client.unmount("images", "master")
        await client.mount("images", "master", "rw")
        assert client._current_mount_strings() == ["images@master+w"]

        # add another repo
        await client.mount("edges", "master", "ro")
        assert sorted(client._current_mount_strings()) == [
            "edges@master+r",
            "images@master+w",
        ]

    async def test_mount(self, client: PythonPachydermMountClient):
        assert await client.mount("images", "master", "r") == {
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

    async def test_mount_repo_noexist(self, client: PythonPachydermMountClient):
        assert await client.mount("noe_repo", "master", "r") == {}

    async def test_unmount(self, client: PythonPachydermMountClient):
        await client.mount("images", "master", "r")
        assert await client.unmount("images", "master", "r") == {
            "repo": "images",
            "branch": "master",
            "mount": {"state": "unmounted"},
        }

    async def test_unmount_all(self, client: PythonPachydermMountClient):
        await client.mount("images", "master", "rw")
        await client.mount("edges", "master", "ro")
        assert await client.unmount_all() == [
            {"repo": "images", "branch": "master", "mount": {"state": "unmounted"}},
            {"repo": "edges", "branch": "master", "mount": {"state": "unmounted"}},
        ]
