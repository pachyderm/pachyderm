import json
import os
import sys
import time
from pathlib import Path
from unittest.mock import patch, mock_open

import pytest
import tornado

from jupyterlab_pachyderm.handlers import NAMESPACE, VERSION
from jupyterlab_pachyderm.pachyderm import MountInterface
from jupyterlab_pachyderm.mount_server_client import write_token_to_config

from . import TEST_NOTEBOOK

pytest_plugins = ["jupyter_server.pytest_plugin"]


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
    mock_client.list_repos.return_value = json.dumps({
        "repo1": {
            "authorization": "read",
            "branches": ["dev", "master"],
            "repo": "repo1",
        },
        "repo2": {
            "authorization": "write",
            "branches": ["master"],
            "repo": "repo2",
        }
    })

    r = await jp_fetch(f"/{NAMESPACE}/{VERSION}/repos")
    assert r.code == 200
    assert json.loads(r.body) == {
        "repo1": {
            "authorization": "read",
            "branches": ["dev", "master"],
            "repo": "repo1",
        },
        "repo2": {
            "authorization": "write",
            "branches": ["master"],
            "repo": "repo2",
        }
    }


@pytest.mark.skipif(sys.version_info < (3, 7), reason="requires python3.7 or higher")
@patch("jupyterlab_pachyderm.handlers.ReposHandler.mount_client", spec=MountInterface)
async def test_list_repos_error(mock_client, jp_fetch):
    status_code = 500
    mock_client.list_repos.side_effect = ErrorWithCode(status_code)
    with pytest.raises(tornado.httpclient.HTTPClientError) as e:
        await jp_fetch(f"/{NAMESPACE}/{VERSION}/repos")
        # note must exit context to capture response

    assert e.value.code == status_code
    assert e.value.response.reason == f"Error listing repos: {status_code}."


@pytest.mark.skipif(sys.version_info < (3, 7), reason="requires python3.7 or higher")
@patch("jupyterlab_pachyderm.handlers.MountsHandler.mount_client", spec=MountInterface)
async def test_list_mounts(mock_client, jp_fetch):
    mock_client.list_mounts.return_value = json.dumps({
        "mounted":{
            "mount1":{
                "name":"mount1",
                "repo":"repo1",
                "branch":"master",
                "commit":"",
                "files":None,
                "glob":"",
                "mode":"ro",
                "state":"mounted",
                "status":"unable to load current commit",
                "mountpoint":"",
                "actual_mounted_commit":"",
                "latest_commit":"",
                "how_many_commits_behind":0
            }
        },
        "unmounted":{
            "repo1":{
                "repo":"repo1",
                "branches":[
                    "dev"
                ],
                "authorization":"off"
            },
            "repo2":{
                "repo":"repo2",
                "branches":[
                    "dev",
                    "master"
                ],
                "authorization":"off"
            }
        }
    })

    r = await jp_fetch(f"/{NAMESPACE}/{VERSION}/mounts")
    assert r.code == 200
    assert json.loads(r.body) == {
        "mounted":{
            "mount1":{
                "name":"mount1",
                "repo":"repo1",
                "branch":"master",
                "commit":"",
                "files":None,
                "glob":"",
                "mode":"ro",
                "state":"mounted",
                "status":"unable to load current commit",
                "mountpoint":"",
                "actual_mounted_commit":"",
                "latest_commit":"",
                "how_many_commits_behind":0
            }
        },
        "unmounted":{
            "repo1":{
                "repo":"repo1",
                "branches":[
                    "dev"
                ],
                "authorization":"off"
            },
            "repo2":{
                "repo":"repo2",
                "branches":[
                    "dev",
                    "master"
                ],
                "authorization":"off"
            }
        }
    }


@pytest.mark.skipif(sys.version_info < (3, 7), reason="requires python3.7 or higher")
@patch("jupyterlab_pachyderm.handlers.MountsHandler.mount_client", spec=MountInterface)
async def test_list_mounts_error(mock_client, jp_fetch):
    status_code = 500
    mock_client.list_mounts.side_effect = ErrorWithCode(status_code)
    with pytest.raises(tornado.httpclient.HTTPClientError) as e:
        await jp_fetch(f"/{NAMESPACE}/{VERSION}/mounts")
        # note must exit context to capture response

    assert e.value.code == status_code
    assert e.value.response.reason == f"Error listing mounts: {status_code}."


@pytest.mark.skipif(sys.version_info < (3, 7), reason="requires python3.7 or higher")
@patch("jupyterlab_pachyderm.handlers.MountHandler.mount_client", spec=MountInterface)
async def test_mount(mock_client, jp_fetch):
    body = {
        "mounts": [
            {
                "name": "mount1",
                "repo": "repo1",
                "branch": "dev",
                "mode": "rw",
            }
        ]
    }
    mock_client.mount.return_value = json.dumps({
        "mounted":{
            body["mounts"][0]["name"]:{
                "name":body["mounts"][0]["name"],
                "repo":body["mounts"][0]["repo"],
                "branch":body["mounts"][0]["branch"],
                "commit":"",
                "files":None,
                "glob":"",
                "mode":body["mounts"][0]["mode"],
                "state":"mounted",
                "status":"unable to load current commit",
                "mountpoint":"",
                "actual_mounted_commit":"",
                "latest_commit":"",
                "how_many_commits_behind":0
            }
        },
        "unmounted":{
            "repo1":{
                "repo":"repo1",
                "branches":[
                    "master"
                ],
                "authorization":"off"
            },
            "repo2":{
                "repo":"repo2",
                "branches":[
                    "dev",
                    "master"
                ],
                "authorization":"off"
            }
        }
    })

    r = await jp_fetch(
        f"/{NAMESPACE}/{VERSION}/_mount",
        method="PUT",
        body=json.dumps(body),
    )
    mock_client.mount.assert_called_with(body)

    assert r.code == 200
    assert json.loads(r.body) == {
        "mounted":{
            body["mounts"][0]["name"]:{
                "name":body["mounts"][0]["name"],
                "repo":body["mounts"][0]["repo"],
                "branch":body["mounts"][0]["branch"],
                "commit":"",
                "files":None,
                "glob":"",
                "mode":body["mounts"][0]["mode"],
                "state":"mounted",
                "status":"unable to load current commit",
                "mountpoint":"",
                "actual_mounted_commit":"",
                "latest_commit":"",
                "how_many_commits_behind":0
            }
        },
        "unmounted":{
            "repo1":{
                "repo":"repo1",
                "branches":[
                    "master"
                ],
                "authorization":"off"
            },
            "repo2":{
                "repo":"repo2",
                "branches":[
                    "dev",
                    "master"
                ],
                "authorization":"off"
            }
        }
    }


@pytest.mark.skipif(sys.version_info < (3, 7), reason="requires python3.7 or higher")
@patch("jupyterlab_pachyderm.handlers.UnmountHandler.mount_client", spec=MountInterface)
async def test_unmount(mock_client, jp_fetch):
    body = {
        "mounts": ["mount1"]
    }
    mock_client.unmount.return_value = json.dumps({
        "mounted":{},
        "unmounted":{
            "repo1":{
                "repo":"repo1",
                "branches":[
                    "master",
                    "dev"
                ],
                "authorization":"off"
            },
            "repo2":{
                "repo":"repo2",
                "branches":[
                    "dev",
                    "master"
                ],
                "authorization":"off"
            }
        }
    })

    r = await jp_fetch(
        f"/{NAMESPACE}/{VERSION}/_unmount",
        method="PUT",
        body=json.dumps(body),
    )
    mock_client.unmount.assert_called_with(body)

    assert r.code == 200
    assert json.loads(r.body) == {
        "mounted":{},
        "unmounted":{
            "repo1":{
                "repo":"repo1",
                "branches":[
                    "master",
                    "dev"
                ],
                "authorization":"off"
            },
            "repo2":{
                "repo":"repo2",
                "branches":[
                    "dev",
                    "master"
                ],
                "authorization":"off"
            }
        }
    }


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
@patch("jupyterlab_pachyderm.handlers.UnmountAllHandler.mount_client", spec=MountInterface)
async def test_unmount_all(mock_client, jp_fetch):
    mock_client.unmount_all.return_value = json.dumps({
        "mounted":{},
        "unmounted":{
            "repo1":{
                "repo":"repo1",
                "branches":[
                    "master",
                    "dev"
                ],
                "authorization":"off"
            },
            "repo2":{
                "repo":"repo2",
                "branches":[
                    "dev",
                    "master"
                ],
                "authorization":"off"
            }
        }
    })

    r = await jp_fetch(
        f"/{NAMESPACE}/{VERSION}/_unmount_all", method="PUT", body="{}"
    )

    assert r.code == 200
    assert json.loads(r.body) == {
        "mounted":{},
        "unmounted":{
            "repo1":{
                "repo":"repo1",
                "branches":[
                    "master",
                    "dev"
                ],
                "authorization":"off"
            },
            "repo2":{
                "repo":"repo2",
                "branches":[
                    "dev",
                    "master"
                ],
                "authorization":"off"
            }
        }
    }


@pytest.mark.skipif(sys.version_info < (3, 7), reason="requires python3.7 or higher")
@patch("jupyterlab_pachyderm.handlers.MountDatumsHandler.mount_client", spec=MountInterface)
async def test_mount_datums(mock_client, jp_fetch):
    body = {
        "input": {
            "pfs": {
                "repo": "images",
                "branch": "dev", 
                "glob": "/*"
            }
        }
    }
    mock_client.mount_datums.return_value = json.dumps({
        "id": "ad9329d",
        "idx": 0,
        "num_datums": 3,
    })

    r = await jp_fetch(
        f"/{NAMESPACE}/{VERSION}/_mount_datums",
        method="PUT",
        body=json.dumps(body)
    )
    mock_client.mount_datums.assert_called_with(body)

    assert json.loads(r.body) == {
        "id": "ad9329d",
        "idx": 0,
        "num_datums": 3,
    }


@pytest.mark.skipif(sys.version_info < (3, 7), reason="requires python3.7 or higher")
@patch("jupyterlab_pachyderm.handlers.ShowDatumHandler.mount_client", spec=MountInterface)
async def test_show_datum(mock_client, jp_fetch):
    mock_client.show_datum.return_value = json.dumps({
        "id": "jdkw9j23",
        "idx": 2,
        "num_datums": 3,
    })

    r = await jp_fetch(
        f"/{NAMESPACE}/{VERSION}/_show_datum",
        method="PUT",
        body=json.dumps({"idx": "2"})
    )

    assert json.loads(r.body) == {
        "id": "jdkw9j23",
        "idx": 2,
        "num_datums": 3,
    }


@pytest.mark.skipif(sys.version_info < (3, 7), reason="requires python3.7 or higher")
@patch("jupyterlab_pachyderm.handlers.DatumsHandler.mount_client", spec=MountInterface)
async def test_get_datums(mock_client, jp_fetch):
    mock_client.get_datums.return_value = json.dumps({
        "input": {"pfs": {"repo": "repo", "branch": "dev", "glob": "/*"}},
        "num_datums": 3,
        "curr_idx": 2,
    })

    r = await jp_fetch(f"/{NAMESPACE}/{VERSION}/datums")

    assert json.loads(r.body) == {
        "input": {"pfs": {"repo": "repo", "branch": "dev", "glob": "/*"}},
        "num_datums": 3,
        "curr_idx": 2,
    }


@pytest.mark.skipif(sys.version_info < (3, 7), reason="requires python3.7 or higher")
@patch(
    "jupyterlab_pachyderm.handlers.ConfigHandler.mount_client",
    spec=MountInterface,
)
async def test_config(mock_client, jp_fetch):
    mock_client.config.return_value = json.dumps({
        "cluster_status": "AUTH_ENABLED",
        "pachd_address": "123.45.1.12:99999"
    })

    # PUT request
    r = await jp_fetch(
        f"/{NAMESPACE}/{VERSION}/config",
        method="PUT",
        body=json.dumps({"pachd_address": "123.45.1.12:99999"})
    )
    
    assert json.loads(r.body) == {
        "cluster_status": "AUTH_ENABLED",
        "pachd_address": "123.45.1.12:99999"
    }
    
    # GET request
    r = await jp_fetch(f"/{NAMESPACE}/{VERSION}/config")
    
    assert json.loads(r.body) == {
        "cluster_status": "AUTH_ENABLED",
        "pachd_address": "123.45.1.12:99999"
    }


@pytest.mark.skipif(sys.version_info < (3, 7), reason="requires python3.7 or higher")
@patch(
    "jupyterlab_pachyderm.handlers.AuthLoginHandler.mount_client",
    spec=MountInterface,
)
async def test_auth_login(mock_client, jp_fetch):
    mock_client.auth_login.return_value = json.dumps({
        "auth_url": "http://some-dex-url"
    })

    r = await jp_fetch(
        f"/{NAMESPACE}/{VERSION}/auth/_login",
        method="PUT",
        body="{}"
    )

    assert json.loads(r.body) == {
        "auth_url": "http://some-dex-url"
    }


@pytest.mark.skipif(sys.version_info < (3, 7), reason="requires python3.7 or higher")
@patch(
    "jupyterlab_pachyderm.handlers.AuthLogoutHandler.mount_client",
    spec=MountInterface,
)
async def test_auth_logout(mock_client, jp_fetch):
    await jp_fetch(
        f"/{NAMESPACE}/{VERSION}/auth/_logout",
        method="PUT",
        body="{}"
    )

    mock_client.auth_logout.assert_called()


@pytest.mark.skipif(sys.version_info < (3, 7), reason="requires python3.7 or higher")
@patch("jupyterlab_pachyderm.handlers.HealthHandler.mount_client", spec=MountInterface)
async def test_health(mock_client, jp_fetch):
    mock_client.health.return_value = json.dumps({"status": "running"})
    r = await jp_fetch(f"/{NAMESPACE}/{VERSION}/health")
    assert json.loads(r.body) == {"status": "running"}


async def test_pps_get(jp_fetch):
    notebook_path = TEST_NOTEBOOK.relative_to(os.getcwd())
    response = await jp_fetch(f"/{NAMESPACE}/{VERSION}/pps/_create/{notebook_path}")
    assert response.code == 200
    body = json.loads(response.body)
    for expected_key in ("pipeline", "description", "transform", "input"):
        assert expected_key in body


@pytest.mark.skipif(sys.version_info < (3, 7), reason="requires python3.7 or higher")
@patch(
    "jupyterlab_pachyderm.handlers.ProjectsHandler.mount_client",
    spec=MountInterface,
)
async def test_get_projects(mock_client, jp_fetch):
    test_projects = [
        {
            "project": {"name": "default"},
            "auth_info":{"permissions": [1, 2, 3], "roles": ["clusterAdmin", "projectOwner"]}
        },
        {
            "project": {"name": "p1"},
            "auth_info":{"permissions": [4, 5, 6], "roles": ["test"]}
        }
    ]

    mock_client.list_projects.return_value = json.dumps(test_projects)
    resp = await jp_fetch(f"/{NAMESPACE}/{VERSION}/projects")
    assert json.loads(resp.body) == test_projects


async def test_write_token_to_config_no_context():
    timestamp = time.time_ns()
    test_config_path = f"/tmp/pach_test_config_{timestamp}.json"

    # test non-existent context
    # we expect the entire context to be copied
    test_mount_server_config_str = """{
  "user_id": "test_user",
  "v2": {
    "active_context": "mount-server",
    "contexts": {
      "mount-server": {
        "session_token": "foo",
        "cluster_deployment_id": "bar",
        "project": "default"
      }
    },
    "metrics": true
  }
}
"""
    test_config_str = """{
  "user_id": "test_user",
  "v2": {
    "active_context": "default",
    "contexts": {
      "default": {
      }
    },
    "metrics": true
  }
}
"""
    with open(test_config_path, 'w') as f:
        f.write(test_config_str)
    write_token_to_config(test_config_path, test_mount_server_config_str)
    test_json = json.loads(test_mount_server_config_str)
    with open(test_config_path) as f:
        test_file_json = json.load(f)
    assert(test_json['v2']['contexts']['mount-server'] ==
           test_file_json['v2']['contexts']['mount-server'])
    assert(test_json['v2']['active_context'] ==
           test_file_json['v2']['active_context'])

async def test_write_token_to_config_existing_context():
    timestamp = time.time_ns()
    test_config_path = f"/tmp/pach_test_config_{timestamp}.json"

    # test pre-existing context
    # we expect the token to get copied and nothing else
    # additionally, we expect the active context to change
    test_mount_server_config_str = """{
  "user_id": "test_user",
  "v2": {
    "active_context": "mount-server",
    "contexts": {
      "mount-server": {
        "session_token": "foo",
        "cluster_deployment_id": "bar",
        "project": "default"
      }
    },
    "metrics": true
  }
}
"""
    test_config_str = """{
  "user_id": "test_user",
  "v2": {
    "active_context": "default",
    "contexts": {
      "mount-server": {
        "cluster_deployment_id": "foo"
      }
    },
    "metrics": true
  }
}
"""
    with open(test_config_path, 'w') as f:
        f.write(test_config_str)
    write_token_to_config(test_config_path, test_mount_server_config_str)
    test_json = json.loads(test_mount_server_config_str)
    with open(test_config_path) as f:
        test_file_json = json.load(f)
    assert(test_json['v2']['contexts']['mount-server']['session_token'] ==
           test_file_json['v2']['contexts']['mount-server']['session_token'])
    assert(test_json['v2']['contexts']['mount-server']['cluster_deployment_id'] !=
           test_file_json['v2']['contexts']['mount-server']['cluster_deployment_id'])
    assert(test_json['v2']['active_context'] ==
           test_file_json['v2']['active_context'])
