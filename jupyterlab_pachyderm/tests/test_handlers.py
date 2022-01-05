import json
import sys
from unittest.mock import patch

import pytest
import tornado

from jupyterlab_pachyderm.handlers import NAMESPACE, VERSION
from jupyterlab_pachyderm.pachyderm import MountInterface


pytest_plugins = ["jupyter_server.pytest_plugin"]


@pytest.fixture
def jp_server_config():
    return {"ServerApp": {"jpserver_extensions": {"jupyterlab_pachyderm": True}}}


@pytest.mark.skipif(sys.version_info < (3, 7), reason="requires python3.7 or higher")
@patch("jupyterlab_pachyderm.handlers.ReposHandler.mount_client", spec=MountInterface)
async def test_list_repos(mock_client, jp_fetch):
    mock_client.list.return_value = {
        "images": {
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
        },
        "edges": {
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
        },
    }
    r = await jp_fetch(f"/{NAMESPACE}/{VERSION}/repos")
    assert r.code == 200
    assert json.loads(r.body) == [
        {
            "repo": "images",
            "branches": [
                {
                    "branch": "master",
                    "mount": {
                        "name": None,
                        "state": "unmounted",
                        "mode": None,
                        "status": None,
                        "mountpoint": None,
                    },
                },
            ],
        },
        {
            "repo": "edges",
            "branches": [
                {
                    "branch": "master",
                    "mount": {
                        "name": None,
                        "state": "unmounted",
                        "mode": None,
                        "status": None,
                        "mountpoint": None,
                    },
                },
            ],
        },
    ]


@pytest.mark.skipif(sys.version_info < (3, 7), reason="requires python3.7 or higher")
@patch("jupyterlab_pachyderm.handlers.RepoHandler.mount_client", spec=MountInterface)
async def test_get_repo(mock_client, jp_fetch):
    mock_client.list.return_value = {
        "images": {
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
        },
        "edges": {
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
        },
    }

    r = await jp_fetch(f"/{NAMESPACE}/{VERSION}/repos/images")

    assert r.code == 200
    assert json.loads(r.body) == {
        "repo": "images",
        "branches": [
            {
                "branch": "master",
                "mount": {
                    "name": None,
                    "state": "unmounted",
                    "mode": None,
                    "status": None,
                    "mountpoint": None,
                },
            }
        ],
    }


@pytest.mark.skipif(sys.version_info < (3, 7), reason="requires python3.7 or higher")
async def test_mount_without_name(jp_fetch):
    # checked by client side path parser, so no mock is needed
    with pytest.raises(tornado.httpclient.HTTPClientError) as e:
        await jp_fetch(
            f"/{NAMESPACE}/{VERSION}/repos/images/_mount", method="PUT", body="{}"
        )
        # note must exit context to capture response
    assert e.value.code == 400
    assert e.value.response.reason == "Missing `name` query parameter"


@pytest.mark.skipif(sys.version_info < (3, 7), reason="requires python3.7 or higher")
@patch(
    "jupyterlab_pachyderm.handlers.RepoMountHandler.mount_client",
    spec=MountInterface,
)
async def test_mount(mock_client, jp_fetch):
    repo, name, mode = "myrepo", "myrepo_mount_name", "ro"
    mock_client.mount.return_value = {
        "repo": repo,
        "branch": "master",
        "mount": {
            "name": name,
            "mode": mode,
            "state": "mounted",
            "status": None,
            "mountpoint": f"/pfs/{name}",
        },
    }

    r = await jp_fetch(
        f"/pachyderm/v1/repos/{repo}/_mount",
        method="PUT",
        params={"name": name, "mode": mode},
        body="{}",
    )

    mock_client.mount.assert_called_with(repo, "master", mode, name)

    assert r.code == 200
    assert json.loads(r.body) == {
        "repo": repo,
        "branch": "master",
        "mount": {
            "name": name,
            "mode": "ro",
            "state": "mounted",
            "status": None,
            "mountpoint": f"/pfs/{name}",
        },
    }


@pytest.mark.skipif(sys.version_info < (3, 7), reason="requires python3.7 or higher")
@patch(
    "jupyterlab_pachyderm.handlers.RepoMountHandler.mount_client",
    spec=MountInterface,
)
async def test_mount_with_branch_and_mode(mock_client, jp_fetch):
    repo, branch, mode, name = "myrepo", "mybranch", "rw", "myrepo_mount_name"
    mock_client.mount.return_value = {
        "repo": repo,
        "branch": branch,
        "mount": {
            "name": name,
            "mode": mode,
            "state": "mounted",
            "status": None,
            "mountpoint": f"/pfs/{name}",
        },
    }

    r = await jp_fetch(
        f"/{NAMESPACE}/{VERSION}/repos/{repo}/{branch}/_mount",
        method="PUT",
        params={"name": name, "mode": mode},
        body="{}",
    )

    mock_client.mount.assert_called_with(repo, branch, mode, name)

    assert r.code == 200
    assert json.loads(r.body) == {
        "repo": repo,
        "branch": branch,
        "mount": {
            "name": name,
            "mode": mode,
            "state": "mounted",
            "status": None,
            "mountpoint": f"/pfs/{name}",
        },
    }


@pytest.mark.skipif(sys.version_info < (3, 7), reason="requires python3.7 or higher")
@patch(
    "jupyterlab_pachyderm.handlers.RepoUnmountHandler.mount_client",
    spec=MountInterface,
)
async def test_unmount_with_branch(mock_client, jp_fetch):
    repo, branch, name = "myrepo", "mybranch", "mount_name"
    mock_client.unmount.return_value = {
        "mount": {
            "name": None,
            "mode": None,
            "state": "unmounted",
            "status": None,
            "mountpoint": None,
        },
    }

    r = await jp_fetch(
        f"/{NAMESPACE}/{VERSION}/repos/{repo}/{branch}/_unmount",
        method="PUT",
        params={"name": name},
        body="{}",
    )

    mock_client.unmount.assert_called_with(repo, branch, name)

    assert r.code == 200
    assert json.loads(r.body) == {
        "repo": repo,
        "branch": branch,
        "mount": {
            "mount": {
                "name": None,
                "mode": None,
                "state": "unmounted",
                "status": None,
                "mountpoint": None,
            }
        },
    }


@pytest.mark.skipif(sys.version_info < (3, 7), reason="requires python3.7 or higher")
@patch(
    "jupyterlab_pachyderm.handlers.ReposUnmountHandler.mount_client",
    spec=MountInterface,
)
async def test_unmount_all(mock_client, jp_fetch):
    mock_client.unmount_all.return_value = [("images", "master")]

    r = await jp_fetch(
        f"/{NAMESPACE}/{VERSION}/repos/_unmount", method="PUT", body="{}"
    )

    assert r.code == 200
    assert json.loads(r.body) == {"unmounted": [["images", "master"]]}


@pytest.mark.skipif(sys.version_info < (3, 7), reason="requires python3.7 or higher")
@patch(
    "jupyterlab_pachyderm.handlers.RepoCommitHandler.mount_client",
    spec=MountInterface,
)
async def test_commit(mock_client, jp_fetch):
    repo, branch, name, message = "myrepo", "mybranch", "mount_name", "First commit"
    mock_client.commit.return_value = True

    r = await jp_fetch(
        f"/{NAMESPACE}/{VERSION}/repos/{repo}/{branch}/_commit",
        method="POST",
        params={"name": name},
        body=json.dumps({"message": message}),
    )

    mock_client.commit.assert_called_with(repo, branch, name, message)
    assert r.code == 200
