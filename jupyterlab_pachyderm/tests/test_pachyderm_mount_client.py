import pytest

from jupyterlab_pachyderm.pachyderm import PachydermMountClient, _parse_pfs_path
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
        c.list()  # need this to get the repo data
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

    def test_get(self, client: PachydermMountClient):
        assert client.get("images") == {
            "branches": {
                "master": {
                    "mount": {
                        "name": None,
                        "state": "unmounted",
                        "status": None,
                        "mode": None,
                        "mountpoint": None,
                    }
                }
            }
        }

    def test_get_noe(self, client: PachydermMountClient):
        assert client.get("repo_noe") == None

    def test_prepare_repos_for_pachctl_mount_repos(self, client: PachydermMountClient):
        assert client._prepare_repos_for_pachctl_mount() == []

        client.mount("images", "master", "r")
        assert client._prepare_repos_for_pachctl_mount() == ["images@master+r"]

        # make sure ro -> r for pachctl
        client.mount("images", "master", "ro")
        assert client._prepare_repos_for_pachctl_mount() == ["images@master+r"]

        # idempotent
        client.mount("images", "master", "r")
        assert client._prepare_repos_for_pachctl_mount() == ["images@master+r"]

        # change the mode
        client.mount("images", "master", "w")
        assert client._prepare_repos_for_pachctl_mount() == ["images@master+w"]

        # make sure rw -> w for pachctl
        client.mount("images", "master", "rw")
        assert client._prepare_repos_for_pachctl_mount() == ["images@master+w"]

        # add another repo
        client.mount("edges", "master", "ro")
        assert sorted(client._prepare_repos_for_pachctl_mount()) == [
            "edges@master+r",
            "images@master+w",
        ]

    def test_mount(self, client: PachydermMountClient):
        assert client.mount("images", "master", "r") == {
            "mount": {
                "name": "images",
                "mode": "r",
                "state": "mounted",
                "status": None,
                "mountpoint": "/pfs/images",
            },
        }
        assert client.get("images") == {
            "branches": {
                "master": {
                    "mount": {
                        "name": "images",
                        "state": "mounted",
                        "status": None,
                        "mode": "r",
                        "mountpoint": "/pfs/images",
                    }
                }
            }
        }

    def test_mount_repo_noexist(self, client: PachydermMountClient):
        assert client.mount("noe_repo", "master", "r") == {}

    def test_unmount(self, client: PachydermMountClient):
        client.mount("images", "master", "r")
        assert client.unmount("images", "master", "r") == {
            "mount": {
                "name": None,
                "mode": None,
                "state": "unmounted",
                "status": None,
                "mountpoint": None,
            },
        }

    def test_unmount_all(self, client: PachydermMountClient):
        client.mount("images", "master", "rw")
        client.mount("edges", "master", "ro")
        assert client.unmount_all() == [("images", "master"), ("edges", "master")]
