import os
import json
import urllib.parse
from pathlib import Path

import pytest
from httpx import AsyncClient
from tornado.web import Application

from jupyterlab_pachyderm.env import PFS_MOUNT_DIR
from pachyderm_sdk import Client
from pachyderm_sdk.api import pfs

from jupyterlab_pachyderm.tests import DEFAULT_PROJECT


@pytest.fixture
def pachyderm_resources():
    repos = ["images", "edges", "montage"]
    branches = ["master", "dev"]
    files = ["file1", "file2"]

    client = Client.from_config()
    client.pfs.delete_all()

    for repo in repos:
        client.pfs.create_repo(repo=pfs.Repo(name=repo))
        for branch in branches:
            for file in files:
                with client.pfs.commit(
                    branch=pfs.Branch.from_uri(f"{repo}@{branch}")
                ) as c:
                    c.put_file_from_bytes(path=f"/{file}", data=b"some data")
                c.wait()

    yield repos, branches, files

    for repo in repos:
        client.pfs.delete_repo(repo=pfs.Repo(name=repo))


async def test_list_mounts(pachyderm_resources, http_client: AsyncClient):
    repos, branches, _ = pachyderm_resources

    payload = {"mounts": [{"name": "mount1", "repo": repos[0], "branch": "master"}]}
    r = await http_client.put("_mount", json=payload)
    assert r.status_code == 200, r.text

    r = await http_client.get("mounts")
    assert r.status_code == 200, r.text

    resp = r.json()
    assert len(resp["mounted"]) == 1
    for mount_info in resp["mounted"]:
        assert mount_info.keys() == {"name", "repo", "branch", "project"}

    for _repo_info in resp["unmounted"]:
        assert _repo_info["repo"] in repos
        assert _repo_info.keys() == {"authorization", "branches", "repo", "project"}
        for _branch in _repo_info["branches"]:
            assert _branch in branches
    assert len(resp["unmounted"]) == len(repos)


async def test_mount(pachyderm_resources, http_client: AsyncClient):
    repos, _, files = pachyderm_resources

    to_mount = {
        "mounts": [
            {
                "name": repos[0],
                "repo": repos[0],
                "branch": "master",
                "project": DEFAULT_PROJECT,
            },
            {
                "name": repos[0] + "_dev",
                "repo": repos[0],
                "branch": "dev",
                "project": DEFAULT_PROJECT,
            },
            {
                "name": repos[1],
                "repo": repos[1],
                "branch": "master",
                "project": DEFAULT_PROJECT,
            },
        ]
    }
    r = await http_client.put("/_mount", json=to_mount)
    assert r.status_code == 200, r.text

    resp = r.json()
    assert len(resp["mounted"]) == 3
    mounted_names = [mount["name"] for mount in resp["mounted"]]
    assert len(resp["unmounted"]) == 2
    for repo in r.json()["unmounted"]:
        if repo["repo"] == repos[1]:
            branches = repo["branches"]
            assert len(branches) == 1
    for repo in r.json()["unmounted"]:
        if repo["repo"] == repos[2]:
            branches = repo["branches"]
            assert len(branches) == 2

    r = await http_client.get("/pfs")
    assert r.status_code == 200, r.text
    resp = r.json()
    assert len(resp["content"]) == 3
    assert sorted([c["name"] for c in resp["content"]]) == sorted(mounted_names)

    r = await http_client.put("/_mount", json=to_mount)
    assert r.status_code == 400, r.text

    r = await http_client.put("/_unmount_all")
    assert r.status_code == 200, r.text
    assert r.json()["mounted"] == []
    assert len(r.json()["unmounted"]) == 3

    r = await http_client.get("/pfs")
    assert r.status_code == 200, r.text
    assert len(r.json()["content"]) == 0


async def test_unmount(pachyderm_resources, http_client: AsyncClient):
    repos, branches, files = pachyderm_resources

    to_mount = {
        "mounts": [
            {
                "name": repos[0],
                "repo": repos[0],
                "branch": "master",
                "project": DEFAULT_PROJECT,
            },
            {
                "name": repos[0] + "_dev",
                "repo": repos[0],
                "branch": "dev",
                "project": DEFAULT_PROJECT,
            },
        ]
    }
    r = await http_client.put("/_mount", json=to_mount)
    assert r.status_code == 200, r.text
    assert len(r.json()["mounted"]) == 2
    assert len(r.json()["unmounted"]) == 2

    r = await http_client.get(f"/pfs/{repos[0]}")
    assert r.status_code == 200, r.text
    assert sorted([c["name"] for c in r.json()["content"]]) == sorted(files)

    r = await http_client.get(f"/pfs/{repos[0]}_dev")
    assert r.status_code == 200, r.text
    assert sorted([c["name"] for c in r.json()["content"]]) == sorted(files)

    r = await http_client.put("/_unmount", json={"mounts": [repos[0] + "_dev"]})
    assert r.status_code == 200, r.text
    assert len(r.json()["mounted"]) == 1
    assert len(r.json()["unmounted"]) == 3
    for repo in r.json()["unmounted"]:
        if repo["repo"] == repos[0]:
            branches = repo["branches"]
            assert len(branches) == 1

    r = await http_client.put("/_unmount", json={"mounts": [repos[0]]})
    assert r.status_code == 200, r.text
    assert len(r.json()["mounted"]) == 0
    assert len(r.json()["unmounted"]) == 3
    for repo in r.json()["unmounted"]:
        if repo["repo"] == repos[0]:
            branches = repo["branches"]
            assert len(branches) == 2

    r = await http_client.get("/pfs")
    assert r.status_code == 200, r.text
    assert len(r.json()["content"]) == 0

    r = await http_client.put(f"/_unmount", json={"mounts": [repos[0]]})
    assert r.status_code == 400, r.text


async def test_pfs_pagination(pachyderm_resources, http_client: AsyncClient):
    repos, _, files = pachyderm_resources
    to_mount = {
        "mounts": [
            {
                "name": repos[0],
                "repo": repos[0],
                "branch": "master",
                "project": DEFAULT_PROJECT,
            },
        ]
    }

    # Mount images repo on master branch for pfs calls
    r = await http_client.put("/_mount", json=to_mount)
    assert r.status_code == 200, r.text

    # Assert default parameters return all
    r = await http_client.get("/pfs/images")
    assert r.status_code == 200, r.text
    r = r.json()
    assert len(r["content"]) == 2
    assert sorted([c["name"] for c in r["content"]]) == sorted(files)

    # Assert pagination_marker=None and number=1 returns file1
    url_params = {'number': 1}
    r = await http_client.get(f"/pfs/images?{urllib.parse.urlencode(url_params)}")
    assert r.status_code == 200, r.text
    r = r.json()
    assert len(r["content"]) == 1
    assert r["content"][0]["name"] == 'file1'

    # Assert pagination_marker=file1 and number=1 returns file2
    url_params = {
        'number': 1,
        'pagination_marker': 'default/images@master:/file1.py'
    }
    r = await http_client.get(f"/pfs/images?{urllib.parse.urlencode(url_params)}")
    assert r.status_code == 200, r.text
    r = r.json()
    assert len(r["content"]) == 1
    assert r["content"][0]["name"] == 'file2'


async def test_view_datum_pagination(pachyderm_resources, http_client: AsyncClient):
    repos, _, files = pachyderm_resources
    input_spec = {
        "input": {
            "pfs": {
                "name": repos[0],
                "repo": repos[0],
                "branch": "master",
                "project": DEFAULT_PROJECT,
            }
        },
    }

    # Mount images repo on master branch for view_datum calls
    r = await http_client.put("/datums/_mount", json=input_spec)
    assert r.status_code == 200, r.text
    r = r.json()
    assert r["idx"] == 0
    assert r["num_datums"] == 1
    assert r["all_datums_received"] == 1

    # Assert default parameters return all
    r = await http_client.get(f"/view_datum/{repos[0]}")
    assert r.status_code == 200, r.text
    r = r.json()
    assert len(r["content"]) == 2
    assert sorted([c["name"] for c in r["content"]]) == sorted(files)

    # Assert pagination_marker=None and number=1 returns file1
    url_params = {'number': 1}
    r = await http_client.get(f"/view_datum/{repos[0]}?{urllib.parse.urlencode(url_params)}")
    assert r.status_code == 200, r.text
    r = r.json()
    assert len(r["content"]) == 1
    assert r["content"][0]["name"] == 'file1'

    # Assert pagination_marker=file1 and number=1 returns file2
    url_params = {
        'number': 1,
        'pagination_marker': 'default/images@master:/file1.py'
    }
    r = await http_client.get(f"/view_datum/{repos[0]}?{urllib.parse.urlencode(url_params)}")
    assert r.status_code == 200, r.text
    r = r.json()
    assert len(r["content"]) == 1
    assert r["content"][0]["name"] == 'file2'


async def test_download_file(
    pachyderm_resources,
    http_client: AsyncClient,
    app: Application,
    tmp_path: Path,
):
    repos, _, files = pachyderm_resources

    # Set root dir to a temporary path to ensure test is repeatable.
    pfs_manager = app.settings.get("pfs_contents_manager")
    assert pfs_manager is not None
    pfs_manager.root_dir = str(tmp_path)

    to_mount = {
        "mounts": [
            {
                "name": repos[0],
                "repo": repos[0],
                "branch": "master",
                "project": DEFAULT_PROJECT,
            },
            {
                "name": repos[0] + "_dev",
                "repo": repos[0],
                "branch": "dev",
                "project": DEFAULT_PROJECT,
            },
            {
                "name": repos[1],
                "repo": repos[1],
                "branch": "master",
                "project": DEFAULT_PROJECT,
            },
        ]
    }
    r = await http_client.put("/_mount", json=to_mount)
    assert r.status_code == 200, r.text

    r = await http_client.put(f"/download/explore/{repos[0]}/{files[0]}")
    assert r.status_code == 200, r.text
    local_file = tmp_path / files[0]
    assert local_file.exists()
    assert local_file.read_text() == "some data"

    r = await http_client.put(f"/download/explore/{repos[0]}/{files[0]}")
    assert r.status_code == 400, r.text

    r = await http_client.put(f"/download/explore/{repos[1]}")
    assert r.status_code == 200, r.text
    local_path = tmp_path / repos[1]
    assert local_path.exists()
    assert local_path.is_dir()
    assert len(list(local_path.iterdir())) == 2

    r = await http_client.put(f"/download/explore/{repos[1]}")
    assert r.status_code == 400, r.text


async def test_mount_datums(pachyderm_resources, http_client: AsyncClient):
    repos, branches, files = pachyderm_resources
    input_spec = {
        "input": {
            "cross": [
                {
                    "pfs": {
                        "repo": repos[0],
                        "glob": "/",
                        "name": "test_name"
                    }
                },
                {
                    "pfs": {
                        "repo": repos[1],
                        "branch": "dev",
                        "glob": "/*",
                    }
                },
                {
                    "pfs": {
                        "repo": repos[2],
                        "glob": "/*",
                    }
                },
            ]
        }
    }

    r = await http_client.put("/datums/_mount", json=input_spec)
    assert r.status_code == 200, r.text
    assert r.json()["idx"] == 0
    assert r.json()["num_datums"] == 4
    assert r.json()["all_datums_received"] is True
    datum0_id = r.json()["id"]

    r = await http_client.get("/view_datum")
    assert r.status_code == 200, r.text
    assert len(r.json()["content"]) == 3

    r = await http_client.get("/view_datum/test_name")
    assert r.status_code == 200, r.text
    assert sorted([c["name"] for c in r.json()["content"]]) == sorted(files)

    r = await http_client.get(f"/view_datum/{repos[1]}_dev")
    assert r.status_code == 200, r.text
    assert len(r.json()["content"]) == 1

    r = await http_client.get(f"/view_datum/{repos[2]}")
    assert r.status_code == 200, r.text
    assert len(r.json()["content"]) == 1

    r = await http_client.put("/datums/_next")
    assert r.status_code == 200, r.text
    assert r.json()["idx"] == 1
    assert r.json()["num_datums"] == 4
    assert r.json()["id"] != datum0_id
    assert r.json()["all_datums_received"] is True

    r = await http_client.get("/view_datum/test_name")
    assert r.status_code == 200, r.text
    assert sorted([c["name"] for c in r.json()["content"]]) == sorted(files)

    r = await http_client.get(f"/view_datum/{repos[1]}_dev")
    assert r.status_code == 200, r.text
    assert len(r.json()["content"]) == 1

    r = await http_client.get(f"/view_datum/{repos[2]}")
    assert r.status_code == 200, r.text
    assert len(r.json()["content"]) == 1

    r = await http_client.put("/datums/_prev")
    assert r.status_code == 200, r.text
    assert r.json()["idx"] == 0
    assert r.json()["num_datums"] == 4
    assert r.json()["id"] == datum0_id
    assert r.json()["all_datums_received"] is True

    r = await http_client.get("/view_datum/test_name")
    assert r.status_code == 200, r.text
    assert sorted([c["name"] for c in r.json()["content"]]) == sorted(files)

    r = await http_client.get(f"/view_datum/{repos[1]}_dev")
    assert r.status_code == 200, r.text
    assert len(r.json()["content"]) == 1

    r = await http_client.get(f"/view_datum/{repos[2]}")
    assert r.status_code == 200, r.text
    assert len(r.json()["content"]) == 1

    r = await http_client.get("/datums")
    assert r.status_code == 200, r.text
    assert json.loads(r.json()["input"]) == input_spec["input"]
    assert r.json()["num_datums"] == 4
    assert r.json()["idx"] == 0
    assert r.json()["all_datums_received"] is True

    # Test mounting a new datum while one is already mounted
    input_spec = {
        "input": {
            "pfs": {
                "repo": repos[0],
                "glob": "/",
                "branch": "dev",
                "name": "test_name_2",
            }
        }
    }

    r = await http_client.put("/datums/_mount", json=input_spec)
    assert r.status_code == 200, r.text
    assert r.json()["idx"] == 0
    assert r.json()["num_datums"] == 1
    assert r.json()["all_datums_received"] is True
    datum0_id = r.json()["id"]

    r = await http_client.get("/view_datum")
    assert r.status_code == 200, r.text
    assert len(r.json()["content"]) == 1

    r = await http_client.get("/view_datum/test_name_2")
    assert r.status_code == 200, r.text
    assert sorted([c["name"] for c in r.json()["content"]]) == sorted(files)

    r = await http_client.put("/_unmount_all")
    assert r.status_code == 200, r.text


async def test_download_datum(pachyderm_resources, http_client: AsyncClient):
    repos, _, files = pachyderm_resources
    input_spec = {
        "input": {
            "cross": [
                {
                    "pfs": {
                        "repo": repos[0],
                        "glob": "/",
                    }
                },
                {
                    "pfs": {
                        "repo": repos[1],
                        "branch": "dev",
                        "glob": "/*",
                    }
                },
                {
                    "pfs": {
                        "repo": repos[2],
                        "glob": "/*",
                    }
                },
            ]
        }
    }

    r = await http_client.put("/datums/_mount", json=input_spec)
    assert r.status_code == 200, r.text
    assert r.json()["idx"] == 0
    assert r.json()["num_datums"] == 4
    assert r.json()["all_datums_received"] is True

    r = await http_client.put("/datums/_download")
    assert r.status_code == 200, r.text
    assert len(list(os.walk(PFS_MOUNT_DIR))[0][1]) == 3
    assert sorted(
        list(os.walk(os.path.join(PFS_MOUNT_DIR, repos[0])))[0][2]
    ) == sorted(files)
    assert f"{repos[1]}_dev" in list(os.walk(PFS_MOUNT_DIR))[0][1]
    assert len(list(os.walk(os.path.join(PFS_MOUNT_DIR, repos[2])))[0][2]) == 1

    r = await http_client.put("/datums/_next")
    assert r.status_code == 200, r.text
    assert r.json()["idx"] == 1
    assert r.json()["num_datums"] == 4
    assert r.json()["all_datums_received"] is True

    r = await http_client.put("/datums/_download")
    assert r.status_code == 200, r.text
    assert len(list(os.walk(PFS_MOUNT_DIR))[0][1]) == 3
    assert sorted(
        list(os.walk(os.path.join(PFS_MOUNT_DIR, repos[0])))[0][2]
    ) == sorted(files)
    assert f"{repos[1]}_dev" in list(os.walk(PFS_MOUNT_DIR))[0][1]
    assert len(list(os.walk(os.path.join(PFS_MOUNT_DIR, repos[2])))[0][2]) == 1
