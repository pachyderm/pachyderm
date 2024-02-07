import os
import sys
import subprocess
import time
import json
from datetime import datetime
from pathlib import Path
from random import randint
from typing import Tuple
import urllib.parse

import pytest
import requests
from httpx import AsyncClient
from tornado.web import Application

from jupyterlab_pachyderm.handlers import NAMESPACE, VERSION
from jupyterlab_pachyderm.env import PACH_CONFIG, PFS_MOUNT_DIR
from jupyterlab_pachyderm.pps_client import METADATA_KEY, PpsConfig
from pachyderm_sdk import Client
from pachyderm_sdk.api import pfs, pps
from pachyderm_sdk.config import ConfigFile

from jupyterlab_pachyderm.tests import TEST_NOTEBOOK

ADDRESS = "http://localhost:8888"
BASE_URL = f"{ADDRESS}/{NAMESPACE}/{VERSION}"
ROOT_TOKEN = "iamroot"
DEFAULT_PROJECT = "default"


@pytest.fixture()
def pachyderm_resources():
    print("creating pachyderm resources")

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


@pytest.fixture
def dev_server(pach_config: Path):
    print("starting development server...")
    p = subprocess.Popen(
        [sys.executable, "-m", "jupyterlab_pachyderm.dev_server"],
        env=dict(
            os.environ,
            PACH_CONFIG=str(pach_config),
        ),
        stdout=subprocess.PIPE,
    )
    try:
        # Give time for python test server to start
        for _ in range(15):
            try:
                if requests.get(f"{BASE_URL}/health", timeout=1).ok:
                    yield
                    break
            except Exception:
                pass
            time.sleep(1)
        else:
            raise RuntimeError("could not start development server")
    finally:
        print("killing development server...")
        p.terminate()
        p.wait()
        time.sleep(1)


@pytest.fixture
def dev_server_with_unmount(dev_server):
    yield
    requests.put(f"{BASE_URL}/_unmount_all")


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
    assert len(resp["unmounted"][0]["branches"]) == 1
    assert len(resp["unmounted"][1]["branches"]) == 2

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
    assert len(r.json()["unmounted"][0]["branches"]) == 1

    r = await http_client.put("/_unmount", json={"mounts": [repos[0]]})
    assert r.status_code == 200, r.text
    assert len(r.json()["mounted"]) == 0
    assert len(r.json()["unmounted"]) == 3
    assert len(r.json()["unmounted"][0]["branches"]) == 2

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
    assert r["content"][0]["name"] == 'file1'
    assert r["content"][1]["name"] == 'file2'

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
    assert r["content"][0]["name"] == 'file1'
    assert r["content"][1]["name"] == 'file2'

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
    local_file = Path.cwd() / files[0]
    assert local_file.exists()
    assert local_file.read_text() == "some data"

    r = await http_client.put(f"/download/explore/{repos[0]}/{files[0]}")
    assert r.status_code == 400, r.text

    r = await http_client.put(f"/download/explore/{repos[1]}")
    assert r.status_code == 200, r.text
    local_path = Path.cwd() / repos[1]
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


class TestConfigHandler:

    @staticmethod
    @pytest.mark.no_config
    async def test_config_no_file(app: Application, http_client: AsyncClient):
        """Test that if there is no config file present, the extension does not
        use a default config."""
        # Arrange
        config_file = app.settings.get("pachyderm_config_file")
        assert config_file is not None and not config_file.exists()

        # Act
        response = await http_client.get("/config")

        # Assert
        response.raise_for_status()
        payload = response.json()
        assert payload["cluster_status"] == "INVALID"
        assert payload["pachd_address"] == ""

    @staticmethod
    @pytest.mark.no_config
    async def test_do_not_set_invalid_config(app: Application, http_client: AsyncClient):
        """Test that PUT /config does not store an invalid config."""
        # Arrange
        assert app.settings.get("pachyderm_client") is None
        invalid_address = "localhost:33333"
        data = {"pachd_address": invalid_address}

        # Act
        response = await http_client.put("/config", json=data)

        # Assert
        response.raise_for_status()
        payload = response.json()
        assert payload["cluster_status"] == "INVALID"
        assert payload["pachd_address"] == invalid_address
        assert app.settings.get("pachyderm_client") is None

    @staticmethod
    async def test_config(pach_config: Path, http_client: AsyncClient):
        """Test that PUT /config with a valid configuration is processed
        by the ConfigHandler as expected."""
        # PUT request
        local_active_context = ConfigFile.from_path(PACH_CONFIG).active_context

        payload = {"pachd_address": local_active_context.pachd_address}
        r = await http_client.put("/config", json=payload)

        config = ConfigFile.from_path(pach_config)
        new_active_context = config.active_context

        assert r.status_code == 200, r.text
        assert r.json()["cluster_status"] != "INVALID"
        assert r.json()["pachd_address"] == new_active_context.pachd_address

        # GET request
        r = await http_client.get("/config")

        assert r.status_code == 200, r.text
        assert r.json()["cluster_status"] != "INVALID"


@pytest.fixture(params=[True, False])
def simple_pachyderm_env(request) -> Tuple[Client, pfs.Repo, pfs.Repo, pps.Pipeline]:
    client = Client.from_config()
    suffix = str(randint(100000, 999999))

    if request.param:
        # Use non-default project
        project = pfs.Project(name=f"test_{suffix}")
        client.pfs.create_project(project=project)
    else:
        # Use default project
        project = pfs.Project(name=DEFAULT_PROJECT)

    repo = pfs.Repo(name=f"images_{suffix}", project=project)
    pipeline = pps.Pipeline(project=project, name=f"test_pipeline_{suffix}")
    companion_repo = pfs.Repo(name=f"{pipeline.name}__context", project=project)
    client.pfs.create_repo(repo=repo)
    yield client, repo, companion_repo, pipeline
    client.pps.delete_pipeline(pipeline=pipeline, force=True)
    client.pfs.delete_repo(repo=companion_repo, force=True)
    client.pfs.delete_repo(repo=repo, force=True)


def _update_metadata(notebook: Path, repo: pfs.Repo, pipeline: pps.Pipeline, external_files: str = '') -> str:
    """Updates the metadata of the specified notebook file with the specified
    project/repo/pipeline information.

    Returns a serialized JSON object that can be written to a file.
    """
    notebook_data = json.loads(notebook.read_bytes())
    config = PpsConfig.from_notebook(notebook)
    config.pipeline = pipeline
    config.input_spec = f'pfs:\n  repo: {repo.name}\n  glob: "/*"'
    # this is currently not being tested so it is set to the empty string
    config.resource_spec = ""
    config.external_files = external_files
    notebook_data["metadata"][METADATA_KEY]["config"] = config.to_dict()
    return json.dumps(notebook_data)


@pytest.fixture
def notebook_path(simple_pachyderm_env) -> Path:
    """Yields a path to a notebook file suitable for testing.

    This writes a temporary notebook file with its metadata populated
      with the expected pipeline and repo names provided by the
      simple_pachyderm_env fixture.
    """
    _client, repo, _companion_repo, pipeline = simple_pachyderm_env

    # Do a considerable amount of data munging.
    notebook_data = _update_metadata(TEST_NOTEBOOK, repo, pipeline)
    notebook_path = TEST_NOTEBOOK.with_stem(f"{TEST_NOTEBOOK.stem}_generated")
    notebook_path.write_text(notebook_data)

    yield notebook_path.relative_to(os.getcwd())
    if notebook_path.exists():
        notebook_path.unlink()


async def test_pps(http_client: AsyncClient, simple_pachyderm_env, notebook_path):
    client, repo, _companion_repo, pipeline = simple_pachyderm_env
    with client.pfs.commit(branch=pfs.Branch(repo=repo, name="master")) as commit:
        client.pfs.put_file_from_bytes(commit=commit, path="/data", data=b"data")
    last_modified = datetime.utcfromtimestamp(os.path.getmtime(notebook_path))
    data = dict(last_modified_time=f"{datetime.isoformat(last_modified)}Z")
    r = await http_client.put(f"/pps/_create/{notebook_path}", json=data)

    assert r.status_code == 200, r.text
    job_info = next(client.pps.list_job(pipeline=pipeline))
    job_info = client.pps.inspect_job(job=job_info.job, wait=True)
    assert job_info.state == pps.JobState.JOB_SUCCESS
    assert r.json()["message"] == (
        "Create pipeline request sent. You may monitor its "
        'status by running "pachctl list pipelines" in a terminal.'
    )


async def test_pps_last_modified_time_not_specified_validation(http_client: AsyncClient, notebook_path):
    r = await http_client.put(f"/pps/_create/{notebook_path}", json={})
    assert r.status_code == 400, r.text
    assert r.json()["reason"] == f"Bad Request: last_modified_time not specified"


@pytest.mark.parametrize("simple_pachyderm_env", [True], indirect=True)
async def test_pps_external_files_do_not_exist_validation(
    http_client: AsyncClient, simple_pachyderm_env, notebook_path
):
    _client, repo, _companion_repo, pipeline = simple_pachyderm_env

    new_notebook_data = _update_metadata(TEST_NOTEBOOK, repo, pipeline, 'does_not_exist.py')
    notebook_path.write_text(new_notebook_data)
    last_modified = datetime.utcfromtimestamp(os.path.getmtime(notebook_path))
    data = dict(last_modified_time=f"{datetime.isoformat(last_modified)}Z")
    r = await http_client.put(f"/pps/_create/{notebook_path}", json=data)

    assert r.status_code == 400
    assert r.json()["message"] == 'Bad Request'
    assert r.json()["reason"] == 'external file does_not_exist.py could not be found in the directory of the Jupyter notebook'


@pytest.mark.parametrize("simple_pachyderm_env", [True], indirect=True)
async def test_pps_external_files_exists_in_another_directory_validation(
    http_client: AsyncClient, simple_pachyderm_env, notebook_path
):
    _client, repo, _companion_repo, pipeline = simple_pachyderm_env

    notebook_directory = TEST_NOTEBOOK.parent
    new_notebook_data = _update_metadata(TEST_NOTEBOOK, repo, pipeline, 'world/hello.py')
    notebook_path.write_text(new_notebook_data)
    last_modified = datetime.utcfromtimestamp(os.path.getmtime(notebook_path))
    data = dict(last_modified_time=f"{datetime.isoformat(last_modified)}Z")

    try:
        notebook_directory.joinpath("world").mkdir()
        notebook_directory.joinpath("world/hello.py").write_text("print('hello')")
        time.sleep(5)

        r = await http_client.put(f"/pps/_create/{notebook_path}", json=data)

        assert r.status_code == 400
        assert r.json()["message"] == 'Bad Request'
        assert r.json()["reason"] == 'external file hello.py could not be found in the directory of the Jupyter notebook'
    finally:
        notebook_directory.joinpath("world/hello.py").unlink()
        notebook_directory.joinpath("world").rmdir()


@pytest.mark.parametrize("simple_pachyderm_env", [True], indirect=True)
async def test_pps_uploads_external_files(
    http_client: AsyncClient, simple_pachyderm_env, notebook_path
):
    client, repo, companion_repo, pipeline = simple_pachyderm_env
    new_notebook_data = _update_metadata(TEST_NOTEBOOK, repo, pipeline, 'hello.py,world.py')
    notebook_path.write_text(new_notebook_data)
    last_modified = datetime.utcfromtimestamp(os.path.getmtime(notebook_path))
    data = dict(last_modified_time=f"{datetime.isoformat(last_modified)}Z")
    try:
        TEST_NOTEBOOK.with_name("hello.py").write_text("print('hello')")
        TEST_NOTEBOOK.with_name("world.py").write_text("print('world')")

        r = await http_client.put(f"/pps/_create/{notebook_path}", json=data)

        assert r.status_code == 200, r.text
        job_info = next(client.pps.list_job(pipeline=pipeline))
        job_info = client.pps.inspect_job(job=job_info.job, wait=True)
        assert job_info.state == pps.JobState.JOB_SUCCESS
        assert r.json()["message"] == (
            "Create pipeline request sent. You may monitor its "
            'status by running "pachctl list pipelines" in a terminal.'
        )
        commits = [info.commit.id for info in client.pfs.list_commit(repo=companion_repo)]
        assert len(commits) == 1
        file_uri = f'{companion_repo}@{commits[0]}:'
        with client.pfs.pfs_file(file=pfs.File.from_uri(f'{file_uri}/hello.py')) as pfs_file:
            assert pfs_file.read().decode() == "print('hello')"
        with client.pfs.pfs_file(file=pfs.File.from_uri(f'{file_uri}/world.py')) as pfs_file:
            assert pfs_file.read().decode() == "print('world')"
    finally:
        TEST_NOTEBOOK.with_name("hello.py").unlink()
        TEST_NOTEBOOK.with_name("world.py").unlink()


@pytest.mark.parametrize("simple_pachyderm_env", [True], indirect=True)
async def test_pps_reuse_pipeline_name_different_project(
    http_client: AsyncClient, simple_pachyderm_env, notebook_path
):
    """This tests creating a pipeline from a notebook within a project, and then creating a new
    pipeline with the same name inside the default project. A bug existed where reusing the pipeline
    name caused an error."""
    client, repo, _companion_repo, pipeline = simple_pachyderm_env
    await test_pps(http_client, simple_pachyderm_env, notebook_path)

    default_project = pfs.Project(name="default")
    default_repo = pfs.Repo(name=repo.name, project=default_project)
    default_pipeline = pps.Pipeline(project=default_project, name=pipeline.name)
    new_notebook_data = _update_metadata(TEST_NOTEBOOK, default_repo, default_pipeline)
    new_notebook = TEST_NOTEBOOK.with_stem(f"{TEST_NOTEBOOK.stem}_generated_2")
    try:
        client.pfs.create_repo(repo=default_repo)
        new_notebook.write_text(new_notebook_data)
        new_notebook = new_notebook.relative_to(os.getcwd())
        last_modified = datetime.utcfromtimestamp(os.path.getmtime(new_notebook))
        data = dict(last_modified_time=f"{datetime.isoformat(last_modified)}Z")
        r = await http_client.put(f"/pps/_create/{new_notebook}", json=data)
        assert r.status_code == 200, r.text
        assert client.pps.inspect_pipeline(pipeline=default_pipeline)
    finally:
        client.pps.delete_pipeline(pipeline=default_pipeline, force=True)
        client.pfs.delete_repo(
            repo=pfs.Repo(name=f"{pipeline.name}__context", project=default_project),
            force=True,
        )
        client.pfs.delete_repo(repo=default_repo, force=True)
        if new_notebook.exists():
            new_notebook.unlink()


@pytest.mark.parametrize("simple_pachyderm_env", [False], indirect=True)
def test_pps_update_default_project_pipeline(
    http_client: AsyncClient, simple_pachyderm_env, notebook_path
):
    """This tests creating and then updating a pipeline within the default project,
    but doing so using an empty string. A bug existed where we would incorrectly try to
    recreate the existing context repo."""
    _client, repo, _companion_repo, pipeline = simple_pachyderm_env
    repo: pfs.Repo
    pipeline: pps.Pipeline
    empty_project = pfs.Project(name="")
    empty_repo = pfs.Repo(name=repo.name, project=empty_project)
    empty_pipeline = pps.Pipeline(project=empty_project, name=pipeline.name)

    new_notebook_data = _update_metadata(TEST_NOTEBOOK, empty_repo, empty_pipeline)
    notebook_path.write_text(new_notebook_data)

    test_pps(http_client, simple_pachyderm_env, notebook_path)
    test_pps(http_client, simple_pachyderm_env, notebook_path)
