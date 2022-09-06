import json
import sys
from unittest.mock import patch

import pytest
import tornado

from jupyterlab_pachyderm.handlers import NAMESPACE, VERSION, _parse_pfs_path
from jupyterlab_pachyderm.pachyderm import MountInterface


pytest_plugins = ["jupyter_server.pytest_plugin"]


def test_parse_pfs_path():
    assert _parse_pfs_path("repo") == ("repo", "master", None)
    assert _parse_pfs_path("repo/branch") == ("repo", "branch", None)
    assert _parse_pfs_path("repo/branch/commit") == ("repo", "branch", "commit")


class ErrorWithCode(Exception):
    def __init__(self, code):
        self.code = code

    def __str__(self):
        return repr(self.code)


@pytest.fixture
def jp_server_config():
    return {"ServerApp": {"jpserver_extensions": {"jupyterlab_pachyderm": True}}}


@pytest.mark.skipif(sys.version_info < (3, 7), reason="requires python3.7 or higher")
@patch("jupyterlab_pachyderm.handlers.ReposHandler.mount_client", spec=MountInterface)
async def test_list_repos(mock_client, jp_fetch):
    mock_client.list.return_value = json.dumps(
        {
            "images": {
                "authorization": "read",
                "branches": {
                    "master": {
                        "mount": [
                            {
                                "name": "",
                                "state": "unmounted",
                                "status": "",
                                "mode": "",
                                "mountpoint": "",
                                "mount_key": {
                                    "repo": "images",
                                    "branch": "master",
                                    "commit": "",
                                },
                            }
                        ],
                        "name": "master",
                    }
                },
                "name": "images",
            },
            "edges": {
                "authorization": "write",
                "branches": {
                    "master": {
                        "mount": [
                            {
                                "name": "",
                                "state": "unmounted",
                                "status": "",
                                "mode": "",
                                "mountpoint": "",
                                "mount_key": {
                                    "repo": "edges",
                                    "branch": "master",
                                    "commit": "",
                                },
                            }
                        ],
                        "name": "master",
                    }
                },
                "name": "edges",
            },
        }
    )

    r = await jp_fetch(f"/{NAMESPACE}/{VERSION}/repos")
    assert r.code == 200
    assert json.loads(r.body) == [
        {
            "repo": "images",
            "authorization": "read",
            "branches": [
                {
                    "branch": "master",
                    "mount": [
                        {
                            "name": "",
                            "state": "unmounted",
                            "status": "",
                            "mode": "",
                            "mountpoint": "",
                            "mount_key": {
                                "repo": "images",
                                "branch": "master",
                                "commit": "",
                            },
                        }
                    ],
                },
            ],
        },
        {
            "repo": "edges",
            "authorization": "write",
            "branches": [
                {
                    "branch": "master",
                    "mount": [
                        {
                            "name": "",
                            "state": "unmounted",
                            "status": "",
                            "mode": "",
                            "mountpoint": "",
                            "mount_key": {
                                "repo": "edges",
                                "branch": "master",
                                "commit": "",
                            },
                        }
                    ],
                },
            ],
        },
    ]


@pytest.mark.skipif(sys.version_info < (3, 7), reason="requires python3.7 or higher")
@patch("jupyterlab_pachyderm.handlers.ReposHandler.mount_client", spec=MountInterface)
async def test_list_repos_error(mock_client, jp_fetch):
    status_code = 500
    mock_client.list.side_effect = ErrorWithCode(status_code)
    with pytest.raises(tornado.httpclient.HTTPClientError) as e:
        await jp_fetch(f"/{NAMESPACE}/{VERSION}/repos")
        # note must exit context to capture response

    assert e.value.code == status_code
    assert e.value.response.reason == f"Error listing repos: {status_code}."


@pytest.mark.skipif(sys.version_info < (3, 7), reason="requires python3.7 or higher")
@patch("jupyterlab_pachyderm.handlers.RepoHandler.mount_client", spec=MountInterface)
async def test_get_repo(mock_client, jp_fetch):
    mock_client.list.return_value = json.dumps(
        {
            "images": {
                "authorization": "off",
                "branches": {
                    "master": {
                        "mount": [
                            {
                                "name": "",
                                "state": "unmounted",
                                "status": "",
                                "mode": "",
                                "mountpoint": "",
                                "mount_key": {
                                    "repo": "images",
                                    "branch": "master",
                                    "commit": "",
                                },
                            }
                        ],
                        "name": "master",
                    }
                },
                "name": "images",
            },
            "edges": {
                "authorization": "off",
                "branches": {
                    "master": {
                        "mount": [
                            {
                                "name": "",
                                "state": "unmounted",
                                "status": "",
                                "mode": "",
                                "mountpoint": "",
                                "mount_key": {
                                    "repo": "edges",
                                    "branch": "master",
                                    "commit": "",
                                },
                            }
                        ],
                        "name": "master",
                    }
                },
                "name": "edges",
            },
        }
    )

    r = await jp_fetch(f"/{NAMESPACE}/{VERSION}/repos/images")

    assert r.code == 200
    assert json.loads(r.body) == {
        "repo": "images",
        "authorization": "off",
        "branches": [
            {
                "branch": "master",
                "mount": [
                    {
                        "name": "",
                        "state": "unmounted",
                        "status": "",
                        "mode": "",
                        "mountpoint": "",
                        "mount_key": {
                            "repo": "images",
                            "branch": "master",
                            "commit": "",
                        },
                    }
                ],
            }
        ],
    }


@pytest.mark.skipif(sys.version_info < (3, 7), reason="requires python3.7 or higher")
@patch("jupyterlab_pachyderm.handlers.RepoHandler.mount_client", spec=MountInterface)
async def test_get_repo_not_exist_error(mock_client, jp_fetch):
    mock_client.list.return_value = json.dumps({})

    repo = "somerepo"
    with pytest.raises(tornado.httpclient.HTTPClientError) as e:
        await jp_fetch(
            f"/{NAMESPACE}/{VERSION}/repos/{repo}",
        )

    assert e.value.code == 400
    assert e.value.response.reason == f"Error repo {repo} not found."


@pytest.mark.skipif(sys.version_info < (3, 7), reason="requires python3.7 or higher")
async def test_mount_without_name(jp_fetch):
    # checked by client side path parser, so no mock is needed
    with pytest.raises(tornado.httpclient.HTTPClientError) as e:
        await jp_fetch(
            f"/{NAMESPACE}/{VERSION}/repos/images/_mount", method="PUT", body="{}"
        )
        # note must exit context to capture response
    assert e.value.code >= 400


@pytest.mark.skipif(sys.version_info < (3, 7), reason="requires python3.7 or higher")
@patch(
    "jupyterlab_pachyderm.handlers.RepoMountHandler.mount_client",
    spec=MountInterface,
)
async def test_mount(mock_client, jp_fetch):
    repo, name, mode = "myrepo", "myrepo_mount_name", "ro"
    mock_client.mount.return_value = json.dumps(
        {
            repo: {
                "authorization": "read",
                "branches": {
                    "master": {
                        "mount": [
                            {
                                "name": name,
                                "mode": mode,
                                "state": "mounted",
                                "status": "",
                                "mountpoint": f"/pfs/{name}",
                                "mount_key": {
                                    "repo": repo,
                                    "branch": "master",
                                    "commit": "",
                                },
                            }
                        ],
                        "name": "master",
                    }
                },
                "name": repo,
            }
        }
    )

    r = await jp_fetch(
        f"/{NAMESPACE}/{VERSION}/repos/{repo}/_mount",
        method="PUT",
        params={"name": name, "mode": mode},
        body="{}",
    )

    mock_client.mount.assert_called_with(repo, "master", mode, name)

    assert r.code == 200
    assert json.loads(r.body) == [
        {
            "authorization": "read",
            "repo": repo,
            "branches": [
                {
                    "branch": "master",
                    "mount": [
                        {
                            "name": name,
                            "mode": mode,
                            "state": "mounted",
                            "status": "",
                            "mountpoint": f"/pfs/{name}",
                            "mount_key": {
                                "repo": repo,
                                "branch": "master",
                                "commit": "",
                            },
                        }
                    ],
                },
            ],
        }
    ]


@pytest.mark.skipif(sys.version_info < (3, 7), reason="requires python3.7 or higher")
@patch(
    "jupyterlab_pachyderm.handlers.RepoMountHandler.mount_client",
    spec=MountInterface,
)
async def test_mount_with_branch_and_mode(mock_client, jp_fetch):
    repo, branch, mode, name = "myrepo", "mybranch", "rw", "myrepo_mount_name"
    mock_client.mount.return_value = json.dumps(
        {
            repo: {
                "authorization": "write",
                "branches": {
                    branch: {
                        "mount": [
                            {
                                "name": name,
                                "mode": mode,
                                "state": "mounted",
                                "status": "",
                                "mountpoint": f"/pfs/{name}",
                                "mount_key": {
                                    "repo": repo,
                                    "branch": branch,
                                    "commit": "",
                                },
                            }
                        ],
                        "name": branch,
                    }
                },
                "name": repo,
            }
        }
    )

    r = await jp_fetch(
        f"/{NAMESPACE}/{VERSION}/repos/{repo}/{branch}/_mount",
        method="PUT",
        params={"name": name, "mode": mode},
        body="{}",
    )

    mock_client.mount.assert_called_with(repo, branch, mode, name)

    assert r.code == 200
    assert json.loads(r.body) == [
        {
            "repo": repo,
            "authorization": "write",
            "branches": [
                {
                    "branch": branch,
                    "mount": [
                        {
                            "name": name,
                            "mode": mode,
                            "state": "mounted",
                            "status": "",
                            "mountpoint": f"/pfs/{name}",
                            "mount_key": {"repo": repo, "branch": branch, "commit": ""},
                        }
                    ],
                },
            ],
        }
    ]


@pytest.mark.skipif(sys.version_info < (3, 7), reason="requires python3.7 or higher")
@patch(
    "jupyterlab_pachyderm.handlers.RepoMountHandler.mount_client", spec=MountInterface
)
async def test_mount_with_error(mock_client, jp_fetch):
    status_code = 500
    mock_client.mount.side_effect = ErrorWithCode(status_code)

    repo, branch, name = "somerepo", "master", "somename"
    with pytest.raises(tornado.httpclient.HTTPClientError) as e:
        await jp_fetch(
            f"/{NAMESPACE}/{VERSION}/repos/{repo}/{branch}/_mount",
            method="PUT",
            params={"name": "{name}", "mode": "ro"},
            body="{}",
        )
        # note must exit context to capture response

    assert e.value.code == status_code
    assert e.value.response.reason == f"Error mounting repo {repo}: {status_code}."


@pytest.mark.skipif(sys.version_info < (3, 7), reason="requires python3.7 or higher")
@patch(
    "jupyterlab_pachyderm.handlers.RepoUnmountHandler.mount_client",
    spec=MountInterface,
)
async def test_unmount_with_branch(mock_client, jp_fetch):
    repo, branch, name = "myrepo", "mybranch", "mount_name"
    mock_client.unmount.return_value = json.dumps(
        {
            repo: {
                "authorization": "write",
                "branches": {
                    branch: {
                        "mount": [
                            {
                                "name": "",
                                "mode": "",
                                "state": "unmounted",
                                "status": "",
                                "mountpoint": "",
                                "mount_key": {
                                    "repo": repo,
                                    "branch": branch,
                                    "commit": "",
                                },
                            }
                        ],
                        "name": branch,
                    }
                },
                "name": repo,
            }
        }
    )

    r = await jp_fetch(
        f"/{NAMESPACE}/{VERSION}/repos/{repo}/{branch}/_unmount",
        method="PUT",
        params={"name": name},
        body="{}",
    )

    mock_client.unmount.assert_called_with(repo, branch, name)

    assert r.code == 200
    assert json.loads(r.body) == [
        {
            "repo": repo,
            "authorization": "write",
            "branches": [
                {
                    "branch": branch,
                    "mount": [
                        {
                            "name": "",
                            "mode": "",
                            "state": "unmounted",
                            "status": "",
                            "mountpoint": "",
                            "mount_key": {"repo": repo, "branch": branch, "commit": ""},
                        }
                    ],
                }
            ],
        }
    ]


@pytest.mark.skipif(sys.version_info < (3, 7), reason="requires python3.7 or higher")
@patch(
    "jupyterlab_pachyderm.handlers.RepoUnmountHandler.mount_client", spec=MountInterface
)
async def test_unmount_with_error(mock_client, jp_fetch):
    status_code = 500
    mock_client.unmount.side_effect = ErrorWithCode(status_code)

    repo, branch, name = "somerepo", "master", "somename"
    with pytest.raises(tornado.httpclient.HTTPClientError) as e:
        await jp_fetch(
            f"/{NAMESPACE}/{VERSION}/repos/{repo}/{branch}/_unmount",
            method="PUT",
            params={"name": "{name}"},
            body="{}",
        )
        # note must exit context to capture response

    assert e.value.code == status_code
    assert e.value.response.reason == f"Error unmounting repo {repo}: {status_code}."


@pytest.mark.skipif(sys.version_info < (3, 7), reason="requires python3.7 or higher")
@patch(
    "jupyterlab_pachyderm.handlers.ReposUnmountHandler.mount_client",
    spec=MountInterface,
)
async def test_unmount_all(mock_client, jp_fetch):
    mock_client.unmount_all.return_value = json.dumps(
        {
            "repo": {
                "authorization": "off",
                "branches": {
                    "branch": {
                        "mount": [
                            {
                                "name": "",
                                "mode": "",
                                "state": "unmounted",
                                "status": "",
                                "mountpoint": "",
                                "mount_key": {
                                    "repo": "repo",
                                    "branch": "branch",
                                    "commit": "",
                                },
                            }
                        ],
                        "name": "branch",
                    }
                },
                "name": "repo",
            }
        }
    )

    r = await jp_fetch(
        f"/{NAMESPACE}/{VERSION}/repos/_unmount", method="PUT", body="{}"
    )

    assert r.code == 200
    assert json.loads(r.body) == [
        {
            "repo": "repo",
            "authorization": "off",
            "branches": [
                {
                    "branch": "branch",
                    "mount": [
                        {
                            "name": "",
                            "mode": "",
                            "state": "unmounted",
                            "status": "",
                            "mountpoint": "",
                            "mount_key": {
                                "repo": "repo",
                                "branch": "branch",
                                "commit": "",
                            },
                        }
                    ],
                }
            ],
        }
    ]


# @pytest.mark.skipif(sys.version_info < (3, 7), reason="requires python3.7 or higher")
# @patch(
#     "jupyterlab_pachyderm.handlers.RepoCommitHandler.mount_client",
#     spec=MountInterface,
# )
# async def test_commit(mock_client, jp_fetch):
#     repo, branch, name, message = "myrepo", "mybranch", "mount_name", "First commit"
#     mock_client.commit.return_value = True

#     r = await jp_fetch(
#         f"/{NAMESPACE}/{VERSION}/repos/{repo}/{branch}/_commit",
#         method="POST",
#         params={"name": name},
#         body=json.dumps({"message": message}),
#     )

#     mock_client.commit.assert_called_with(repo, branch, name, message)
#     assert r.code == 200


@pytest.mark.skipif(sys.version_info < (3, 7), reason="requires python3.7 or higher")
@patch(
    "jupyterlab_pachyderm.handlers.ConfigHandler.mount_client",
    spec=MountInterface,
)
async def test_config(mock_client, jp_fetch):
    mock_client.config.return_value = json.dumps(
        {"cluster_status": "AUTH_ENABLED", "pachd_address": "123.45.1.12:99999"}
    )

    # PUT request
    r = await jp_fetch(
        f"/{NAMESPACE}/{VERSION}/config",
        method="PUT",
        body=json.dumps({"pachd_address": "123.45.1.12:99999"}),
    )

    assert json.loads(r.body) == {
        "cluster_status": "AUTH_ENABLED",
        "pachd_address": "123.45.1.12:99999",
    }

    # GET request
    r = await jp_fetch(f"/{NAMESPACE}/{VERSION}/config")

    assert json.loads(r.body) == {
        "cluster_status": "AUTH_ENABLED",
        "pachd_address": "123.45.1.12:99999",
    }


@pytest.mark.skipif(sys.version_info < (3, 7), reason="requires python3.7 or higher")
@patch(
    "jupyterlab_pachyderm.handlers.AuthLoginHandler.mount_client",
    spec=MountInterface,
)
async def test_auth_login(mock_client, jp_fetch):
    mock_client.auth_login.return_value = json.dumps(
        {"auth_url": "http://some-dex-url"}
    )

    r = await jp_fetch(f"/{NAMESPACE}/{VERSION}/auth/_login", method="PUT", body="{}")

    assert json.loads(r.body) == {"auth_url": "http://some-dex-url"}


@pytest.mark.skipif(sys.version_info < (3, 7), reason="requires python3.7 or higher")
@patch(
    "jupyterlab_pachyderm.handlers.AuthLogoutHandler.mount_client",
    spec=MountInterface,
)
async def test_auth_logout(mock_client, jp_fetch):
    await jp_fetch(f"/{NAMESPACE}/{VERSION}/auth/_logout", method="PUT", body="{}")

    mock_client.auth_logout.assert_called()


@pytest.mark.skipif(sys.version_info < (3, 7), reason="requires python3.7 or higher")
@patch("jupyterlab_pachyderm.handlers.HealthHandler.mount_client", spec=MountInterface)
async def test_health(mock_client, jp_fetch):
    mock_client.health.return_value = json.dumps({"status": "running"})
    r = await jp_fetch(f"/{NAMESPACE}/{VERSION}/health")
    assert json.loads(r.body) == {"status": "running"}
