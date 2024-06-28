import os
import json
import urllib.parse
from pathlib import Path

import pytest
from httpx import AsyncClient
from tornado.web import Application

from jupyterlab_pachyderm.env import PFS_MOUNT_DIR
from jupyterlab_pachyderm.pfs_manager import DatumManager
from pachyderm_sdk import Client
from pachyderm_sdk.api import pfs

from jupyterlab_pachyderm.tests import DEFAULT_PROJECT


@pytest.fixture
async def pachyderm_resources(http_client: AsyncClient):
    repos = ["images", "edges", "montage"]
    branches = ["master", "dev"]
    commits = {} # {branch_uri: [commit_id, ...]}
    files = ["file1", "file2"]

    client = Client.from_config()
    for repo in repos:
        client.pfs.create_repo(repo=pfs.Repo(name=repo))
        for branch in branches:
            branch_uri = f"{repo}@{branch}"
            commits[branch_uri] = []
            for file in files:
                with client.pfs.commit(
                    branch=pfs.Branch.from_uri(branch_uri)
                ) as c:
                    c.put_file_from_bytes(path=f"/{file}", data=b"some data")
                c.wait()
                commits[branch_uri].append(c.id)

    yield repos, branches, files, commits

    for repo in repos:
        client.pfs.delete_repo(repo=pfs.Repo(name=repo))

async def test_list_repos(pachyderm_resources, http_client: AsyncClient):
    repos, branches, _, _ = pachyderm_resources

    r = await http_client.get("repos")
    assert r.status_code == 200, r.text

    resp = r.json()
    assert len(resp) == 3
    for repo_uri in resp:
        repo = resp[repo_uri]
        assert repo['name'] in repos
        for branch in repo['branches']:
            assert branch['name'] in branches

async def test_mount_missing_body(pachyderm_resources, http_client: AsyncClient):
    _, _, _, _ = pachyderm_resources

    r = await http_client.put("/explore/mount")
    assert r.status_code == 400, r.text

async def test_mount_missing_commit_uri(pachyderm_resources, http_client: AsyncClient):
    _, _, _, _ = pachyderm_resources

    r = await http_client.put("/explore/mount", json={})
    assert r.status_code == 400, r.text

async def test_mount_invalid_commit_uri(pachyderm_resources, http_client: AsyncClient):
    _, _, _, _ = pachyderm_resources

    r = await http_client.put("/explore/mount", json={'commit_uri': 'fake@repo/fakebranch'})
    assert r.status_code == 400, r.text

async def test_pfs_mount_commit_with_branch(pachyderm_resources, http_client: AsyncClient):
    repos, _, files, commits = pachyderm_resources

    branch_uri = f'{repos[0]}@master'
    commit_id = commits[branch_uri][1]
    r = await http_client.put("/explore/mount", json={"commit_uri": f'{repos[0]}@master={commit_id}'})
    assert r.status_code == 200, r.text
    r = await http_client.get("/pfs/images")
    assert r.status_code == 200, r.text
    r = r.json()
    assert len(r["content"]) == 2
    assert sorted([c["name"] for c in r["content"]]) == sorted(files)

async def test_pfs_mount_commit_without_branch(pachyderm_resources, http_client: AsyncClient):
    repos, _, files, commits = pachyderm_resources

    branch_uri = f'{repos[0]}@master'
    commit_id = commits[branch_uri][1]
    r = await http_client.put("/explore/mount", json={"commit_uri": f'{repos[0]}@{commit_id}'})
    assert r.status_code == 200, r.text
    r = await http_client.get("/pfs/images")
    assert r.status_code == 200, r.text
    r = r.json()
    assert len(r["content"]) == 2
    assert sorted([c["name"] for c in r["content"]]) == sorted(files)

async def test_pfs_pagination(pachyderm_resources, http_client: AsyncClient):
    repos, _, files, _ = pachyderm_resources

    # Assert default parameters return all
    r = await http_client.put("/explore/mount", json={"commit_uri": f'{repos[0]}@master'})
    assert r.status_code == 200, r.text
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
    url_params['pagination_marker'] = 'default/images@master:/file1.py'
    r = await http_client.get(f"/pfs/images?{urllib.parse.urlencode(url_params)}")
    assert r.status_code == 200, r.text
    r = r.json()
    assert len(r["content"]) == 1
    assert r["content"][0]["name"] == 'file2'


async def test_view_datum_pagination(pachyderm_resources, http_client: AsyncClient):
    repos, _, files, _ = pachyderm_resources
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
    assert r["num_datums_received"] == 1
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
    repos, _, files, _ = pachyderm_resources

    # Set root dir to a temporary path to ensure test is repeatable.
    pfs_manager = app.settings.get("pfs_contents_manager")
    assert pfs_manager is not None
    pfs_manager.root_dir = str(tmp_path)

    r = await http_client.put("/explore/mount", json={"commit_uri": f'{repos[0]}@master'})
    assert r.status_code == 200, r.text
    r = await http_client.put(f"/download/explore/{repos[0]}/{files[0]}")
    assert r.status_code == 200, r.text
    local_file = tmp_path / files[0]
    assert local_file.exists()
    assert local_file.read_text() == "some data"

    r = await http_client.put(f"/download/explore/{repos[0]}/{files[0]}?")
    assert r.status_code == 400, r.text

    r = await http_client.put("/explore/mount", json={"commit_uri": f'{repos[1]}@master'})
    assert r.status_code == 200, r.text
    r = await http_client.put(f"/download/explore/{repos[1]}")
    assert r.status_code == 200, r.text
    local_path = tmp_path / repos[1]
    assert local_path.exists()
    assert local_path.is_dir()
    assert len(list(local_path.iterdir())) == 2

    r = await http_client.put(f"/download/explore/{repos[1]}")
    assert r.status_code == 400, r.text

async def test_mount_datums(pachyderm_resources, http_client: AsyncClient):
    repos, _, files, _ = pachyderm_resources
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
    assert r.json()["num_datums_received"] == 4
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
    assert r.json()["num_datums_received"] == 4
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
    assert r.json()["num_datums_received"] == 4
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
    assert r.json()["num_datums_received"] == 4
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
    assert r.json()["num_datums_received"] == 1
    assert r.json()["all_datums_received"] is True
    datum0_id = r.json()["id"]

    r = await http_client.get("/view_datum")
    assert r.status_code == 200, r.text
    assert len(r.json()["content"]) == 1

    r = await http_client.get("/view_datum/test_name_2")
    assert r.status_code == 200, r.text
    assert sorted([c["name"] for c in r.json()["content"]]) == sorted(files)


async def test_mount_datums_multiple_batches(http_client: AsyncClient):
    client = Client.from_config()
    repo = pfs.Repo(name="multiple_batches")
    client.pfs.delete_repo(repo=repo, force=True)
    client.pfs.create_repo(repo=repo)

    batch_size = DatumManager.DATUM_BATCH_SIZE
    total_datums = int(batch_size * 1.5)
    with client.pfs.commit(branch=pfs.Branch.from_uri(f"{repo.name}@master")) as c:
        for i in range(total_datums):
            # Number of files will require a request for a second batch of datums
            c.put_file_from_bytes(path=f"/{i}", data=b"some data")
    c.wait()

    input_spec = {
        "input": {
            "pfs": {
                "repo": repo.name,
                "glob": "/*",
            }
        }
    }
    r = await http_client.put("/datums/_mount", json=input_spec)
    assert r.status_code == 200, r.text
    assert r.json()["idx"] == 0
    assert r.json()["num_datums_received"] == batch_size
    assert r.json()["all_datums_received"] is False

    # Cycle to the last datum in the current batch
    for i in range(1, batch_size):
        r = await http_client.put("/datums/_next")
        assert r.status_code == 200, r.text
        assert r.json()["idx"] == i
        assert r.json()["num_datums_received"] == batch_size
        assert r.json()["all_datums_received"] is False

    # Grab the next (final) batch of datums
    for i in range(batch_size, total_datums):
        r = await http_client.put("/datums/_next")
        assert r.status_code == 200, r.text
        assert r.json()["idx"] == i
        assert r.json()["num_datums_received"] == total_datums
        assert r.json()["all_datums_received"] is True

    # Verify no more datums and cycler wraps to beginning
    r = await http_client.put("/datums/_next")
    assert r.status_code == 200, r.text
    assert r.json()["idx"] == 0

    client.pfs.delete_repo(repo=repo)


async def test_download_datum(pachyderm_resources, http_client: AsyncClient):
    repos, _, files, _ = pachyderm_resources
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
    assert r.json()["num_datums_received"] == 4
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
    assert r.json()["num_datums_received"] == 4
    assert r.json()["all_datums_received"] is True

    r = await http_client.put("/datums/_download")
    assert r.status_code == 200, r.text
    assert len(list(os.walk(PFS_MOUNT_DIR))[0][1]) == 3
    assert sorted(
        list(os.walk(os.path.join(PFS_MOUNT_DIR, repos[0])))[0][2]
    ) == sorted(files)
    assert f"{repos[1]}_dev" in list(os.walk(PFS_MOUNT_DIR))[0][1]
    assert len(list(os.walk(os.path.join(PFS_MOUNT_DIR, repos[2])))[0][2]) == 1
