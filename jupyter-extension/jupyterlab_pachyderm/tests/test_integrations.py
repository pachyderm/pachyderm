import os
import sys
import subprocess
import time
import json
import shutil
from datetime import datetime
from pathlib import Path
from random import randint

import pytest
import requests

from jupyterlab_pachyderm.handlers import NAMESPACE, VERSION
from jupyterlab_pachyderm.env import (
    PFS_MOUNT_DIR,
    MOUNT_SERVER_LOG_FILE,
    PACH_CONFIG
)
from jupyterlab_pachyderm.pps_client import METADATA_KEY, PpsConfig
from pachyderm_sdk import Client
from pachyderm_sdk.api import pfs, pps

from . import TEST_NOTEBOOK, TEST_REQUIREMENTS

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

    client = Client()
    client.pfs.delete_all()

    for repo in repos:
        client.pfs.create_repo(repo=pfs.Repo(name=repo))
        for branch in branches:
            for file in files:
                with client.pfs.commit(
                    branch=pfs.Branch.from_uri(f"{repo}@{branch}")
                ) as c:
                    c.put_file_from_bytes(path=f"/{file}", data=b"some data")

    yield repos, branches, files


@pytest.fixture()
def dev_server():
    print("starting development server...")
    p = subprocess.Popen(
        [sys.executable, "-m", "jupyterlab_pachyderm.dev_server"],
        # preserve specifically:
        # PATH, PACH_CONFIG, PFS_MOUNT_DIR and MOUNT_SERVER_LOG_FILE
        # The args after os.environ should be no-ops, but they're here in case
        # env.py changes (mount-server should use jupyterlab-pach's defaults).
        env=dict(os.environ,
            PACH_CONFIG=PACH_CONFIG,
            PFS_MOUNT_DIR=PFS_MOUNT_DIR,
            MOUNT_SERVER_LOG_FILE=MOUNT_SERVER_LOG_FILE,
        ),
        stdout=subprocess.PIPE,
    )
    # Give time for python test server to start
    time.sleep(3)

    r = requests.put(
        f"{BASE_URL}/config", data=json.dumps({"pachd_address": "localhost:30650"})
    )

    # Give time for mount server to start
    running = False
    for _ in range(15):
        try:
            r = requests.get(f"{BASE_URL}/config", timeout=1)
            if r.status_code == 200 and r.json()["cluster_status"] != "INVALID":
                running = True
                break
        except Exception:
            pass
        time.sleep(1)

    if running:
        yield

    print("killing development server...")

    subprocess.run(["pkill", "-f", "mount-server"])
    subprocess.run([shutil.which("umount"), PFS_MOUNT_DIR])
    p.terminate()
    p.wait()
    time.sleep(1)

    if not running:
        raise RuntimeError("mount server is having issues starting up")


def test_list_repos(pachyderm_resources, dev_server):
    repos, branches, _ = pachyderm_resources

    r = requests.get(f"{BASE_URL}/repos")

    assert r.status_code == 200, r.text
    for repo_info in r.json():
        assert repo_info.keys() == {"authorization", "branches", "repo", "project"}
        assert repo_info["repo"] in repos
        for _branch in repo_info["branches"]:
            assert _branch in branches


def test_list_mounts(pachyderm_resources, dev_server):
    repos, branches, _ = pachyderm_resources

    r = requests.put(
        f"{BASE_URL}/_mount",
        data=json.dumps(
            {
                "mounts": [
                    {
                        "name": "mount1",
                        "repo": repos[0],
                    }
                ]
            }
        ),
    )
    assert r.status_code == 200, r.text

    r = requests.get(f"{BASE_URL}/mounts")
    assert r.status_code == 200, r.text

    resp = r.json()
    assert len(resp["mounted"]) == 1
    for mount_info in resp["mounted"]:
        assert mount_info.keys() == {
            "name",
            "repo",
            "branch",
            "project",
            "commit",
            "paths",
            "mode",
            "state",
            "status",
            "mountpoint",
            "latest_commit",
            "how_many_commits_behind",
        }

    for _repo_info in resp["unmounted"]:
        assert _repo_info["repo"] in repos
        assert _repo_info.keys() == {"authorization", "branches", "repo", "project"}
        for _branch in _repo_info["branches"]:
            assert _branch in branches
    assert len(resp["unmounted"]) == len(repos)


def test_mount(pachyderm_resources, dev_server):
    repos, _, files = pachyderm_resources

    to_mount = {
        "mounts": [
            {
                "name": repos[0],
                "repo": repos[0],
                "branch": "master",
                "project": DEFAULT_PROJECT,
                "mode": "ro",
            },
            {
                "name": repos[0] + "_dev",
                "repo": repos[0],
                "branch": "dev",
                "project": DEFAULT_PROJECT,
                "mode": "ro",
            },
            {
                "name": repos[1],
                "repo": repos[1],
                "branch": "master",
                "project": DEFAULT_PROJECT,
                "mode": "ro",
            },
        ]
    }
    r = requests.put(f"{BASE_URL}/_mount", data=json.dumps(to_mount))
    assert r.status_code == 200, r.text

    resp = r.json()
    assert len(resp["mounted"]) == 3
    assert len(list(os.walk(PFS_MOUNT_DIR))[0][1]) == 3
    for mount_info in resp["mounted"]:
        assert sorted(
            list(os.walk(os.path.join(PFS_MOUNT_DIR, mount_info["name"])))[0][2]
        ) == sorted(files)
    assert len(resp["unmounted"]) == 3
    assert len(resp["unmounted"][0]["branches"]) == 2
    assert len(resp["unmounted"][1]["branches"]) == 2

    r = requests.put(
        f"{BASE_URL}/_unmount_all",
    )
    assert r.status_code == 200, r.text
    assert r.json()["mounted"] == []
    assert len(r.json()["unmounted"]) == 3
    assert list(os.walk(PFS_MOUNT_DIR)) == [(PFS_MOUNT_DIR, [], [])]


def test_unmount(pachyderm_resources, dev_server):
    repos, branches, files = pachyderm_resources

    to_mount = {
        "mounts": [
            {
                "name": repos[0],
                "repo": repos[0],
                "branch": "master",
                "project": DEFAULT_PROJECT,
                "mode": "ro",
            },
            {
                "name": repos[0] + "_dev",
                "repo": repos[0],
                "branch": "dev",
                "project": DEFAULT_PROJECT,
                "mode": "ro",
            },
        ]
    }
    r = requests.put(f"{BASE_URL}/_mount", data=json.dumps(to_mount))
    assert r.status_code == 200, r.text
    assert sorted(list(os.walk(os.path.join(PFS_MOUNT_DIR, repos[0])))[0][2]) == sorted(
        files
    )
    assert sorted(
        list(os.walk(os.path.join(PFS_MOUNT_DIR, repos[0] + "_dev")))[0][2]
    ) == sorted(files)
    assert len(r.json()["mounted"]) == 2
    assert len(r.json()["unmounted"]) == 3

    r = requests.put(
        f"{BASE_URL}/_unmount",
        data=json.dumps({"mounts": [repos[0] + "_dev"]}),
    )
    assert r.status_code == 200, r.text
    assert len(r.json()["mounted"]) == 1
    assert len(r.json()["unmounted"]) == 3
    assert len(r.json()["unmounted"][0]["branches"]) == 2

    r = requests.put(
        f"{BASE_URL}/_unmount",
        data=json.dumps({"mounts": [repos[0]]}),
    )
    assert r.status_code == 200, r.text
    assert len(r.json()["mounted"]) == 0
    assert len(r.json()["unmounted"]) == 3
    assert len(r.json()["unmounted"][0]["branches"]) == 2
    assert list(os.walk(PFS_MOUNT_DIR)) == [(PFS_MOUNT_DIR, [], [])]


@pytest.mark.skip(reason="test flakes due to 'missing chunk' error that hasn't been diagnosed")
def test_mount_datums(pachyderm_resources, dev_server):
    repos, branches, files = pachyderm_resources
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

    r = requests.put(f"{BASE_URL}/datums/_mount", data=json.dumps(input_spec))
    assert r.status_code == 200, r.text
    assert r.json()["idx"] == 0
    assert r.json()["num_datums"] == 4
    assert r.json()["all_datums_received"] == True
    list(os.walk(os.path.join(PFS_MOUNT_DIR, "out")))  # makes "out" dir appear
    assert len(list(os.walk(PFS_MOUNT_DIR))[0][1]) == 4
    datum0_id = r.json()["id"]

    assert sorted(
        list(
            os.walk(
                os.path.join(PFS_MOUNT_DIR, "".join([DEFAULT_PROJECT, "_", repos[0]]))
            )
        )[0][2]
    ) == sorted(files)
    assert (
        "".join([DEFAULT_PROJECT, "_", repos[1], "_dev"])
        in list(os.walk(PFS_MOUNT_DIR))[0][1]
    )
    assert (
        len(
            list(
                os.walk(
                    os.path.join(
                        PFS_MOUNT_DIR, "".join([DEFAULT_PROJECT, "_", repos[2]])
                    )
                )
            )[0][2]
        )
        == 1
    )

    r = requests.put(f"{BASE_URL}/datums/_next")
    assert r.status_code == 200, r.text
    assert r.json()["idx"] == 1
    assert r.json()["num_datums"] == 4
    # TODO: uncomment this and below line when we transition fully to using ListDatum when getting datums. #ListDatumPagination
    # assert r.json()["id"] != datum0_id
    assert r.json()["all_datums_received"] == True

    r = requests.put(f"{BASE_URL}/datums/_prev")
    assert r.status_code == 200, r.text
    assert r.json()["idx"] == 0
    assert r.json()["num_datums"] == 4
    # assert r.json()["id"] == datum0_id
    assert r.json()["all_datums_received"] == True

    r = requests.get(f"{BASE_URL}/datums")
    assert r.status_code == 200, r.text
    assert r.json()["input"] == {
        "cross": [
            {
                "pfs": {
                    "repo": repos[0],
                    "name": "".join([DEFAULT_PROJECT, "_", repos[0]]),
                    "glob": "/",
                    "project": DEFAULT_PROJECT,
                    "branch": "master",
                }
            },
            {
                "pfs": {
                    "repo": repos[1],
                    "name": "".join([DEFAULT_PROJECT, "_", repos[1], "_dev"]),
                    "branch": "dev",
                    "glob": "/*",
                    "project": DEFAULT_PROJECT,
                }
            },
            {
                "pfs": {
                    "repo": repos[2],
                    "name": "".join([DEFAULT_PROJECT, "_", repos[2]]),
                    "glob": "/*",
                    "project": DEFAULT_PROJECT,
                    "branch": "master",
                }
            },
        ]
    }
    assert r.json()["num_datums"] == 4
    assert r.json()["idx"] == 0
    assert r.json()["all_datums_received"] == True

    r = requests.put(f"{BASE_URL}/_unmount_all")
    assert r.status_code == 200, r.text


def test_config(dev_server):
    # PUT request
    test_endpoint = "localhost:30650"
    r = requests.put(
        f"{BASE_URL}/config", data=json.dumps({"pachd_address": test_endpoint})
    )

    config = json.load(open(os.path.expanduser(PACH_CONFIG)))
    active_context = config["v2"]["active_context"]
    try:
        endpoint_in_config = config["v2"]["contexts"][active_context]["pachd_address"]
    except:
        endpoint_in_config = str(
            config["v2"]["contexts"][active_context]["port_forwarders"]["pachd"]
        )

    assert r.status_code == 200, r.text
    assert r.json()["cluster_status"] != "INVALID"
    assert "30650" in endpoint_in_config

    # GET request
    r = requests.get(f"{BASE_URL}/config")

    assert r.status_code == 200, r.text
    assert r.json()["cluster_status"] != "INVALID"


@pytest.fixture(params=[True, False])
def simple_pachyderm_env(request):
    client = Client()
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
    yield client, repo, pipeline
    client.pps.delete_pipeline(pipeline=pipeline, force=True)
    client.pfs.delete_repo(repo=companion_repo, force=True)
    client.pfs.delete_repo(repo=repo, force=True)


def _update_metadata(notebook: Path, repo: pfs.Repo, pipeline: pps.Pipeline) -> str:
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
    config.requirements = str(
        notebook.with_name(config.requirements).relative_to(os.getcwd())
    )
    notebook_data["metadata"][METADATA_KEY]["config"] = config.to_dict()
    return json.dumps(notebook_data)


@pytest.fixture
def notebook_path(simple_pachyderm_env) -> Path:
    """Yields a path to a notebook file suitable for testing.

    This writes a temporary notebook file with its metadata populated
      with the expected pipeline and repo names provided by the
      simple_pachyderm_env fixture.
    """
    _client, repo, pipeline = simple_pachyderm_env

    # Do a considerable amount of data munging.
    notebook_data = _update_metadata(TEST_NOTEBOOK, repo, pipeline)
    notebook_path = TEST_NOTEBOOK.with_stem(f"{TEST_NOTEBOOK.stem}_generated")
    notebook_path.write_text(notebook_data)

    yield notebook_path.relative_to(os.getcwd())
    if notebook_path.exists():
        notebook_path.unlink()


def test_pps(dev_server, simple_pachyderm_env, notebook_path):
    client, repo, pipeline = simple_pachyderm_env
    with client.pfs.commit(branch=pfs.Branch(repo=repo, name="master")) as commit:
        client.pfs.put_file_from_bytes(commit=commit, path="/data", data=b"data")
    last_modified = datetime.utcfromtimestamp(os.path.getmtime(notebook_path))
    data = dict(last_modified_time=f"{datetime.isoformat(last_modified)}Z")
    r = requests.put(f"{BASE_URL}/pps/_create/{notebook_path}", data=json.dumps(data))
    assert r.status_code == 200, r.text
    job_info = next(client.pps.list_job(pipeline=pipeline))
    job_info = client.pps.inspect_job(job=job_info.job, wait=True)
    assert job_info.state == pps.JobState.JOB_SUCCESS
    assert r.json()["message"] == (
        "Create pipeline request sent. You may monitor its "
        'status by running "pachctl list pipelines" in a terminal.'
    )


def test_pps_validation_errors(dev_server, notebook_path):
    r = requests.put(f"{BASE_URL}/pps/_create/{notebook_path}", data=json.dumps({}))
    assert r.status_code == 400, r.text
    assert r.json()["reason"] == f"Bad Request: last_modified_time not specified"


@pytest.mark.parametrize("simple_pachyderm_env", [True], indirect=True)
def test_pps_reuse_pipeline_name_different_project(
    dev_server, simple_pachyderm_env, notebook_path
):
    """This tests creating a pipeline from a notebook within a project, and then creating a new
    pipeline with the same name inside the default project. A bug existed where reusing the pipeline
    name caused an error."""
    client, repo, pipeline = simple_pachyderm_env
    test_pps(dev_server, simple_pachyderm_env, notebook_path)

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
        r = requests.put(
            f"{BASE_URL}/pps/_create/{new_notebook}", data=json.dumps(data)
        )
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
    dev_server, simple_pachyderm_env, notebook_path
):
    """This tests creating and then updating a pipeline within the default project,
    but doing so using an empty string. A bug existed where we would incorrectly try to
    recreate the existing context repo."""
    _client, repo, pipeline = simple_pachyderm_env
    repo: pfs.Repo
    pipeline: pps.Pipeline
    empty_project = pfs.Project(name="")
    empty_repo = pfs.Repo(name=repo.name, project=empty_project)
    empty_pipeline = pps.Pipeline(project=empty_project, name=pipeline.name)

    new_notebook_data = _update_metadata(TEST_NOTEBOOK, empty_repo, empty_pipeline)
    notebook_path.write_text(new_notebook_data)

    test_pps(dev_server, simple_pachyderm_env, notebook_path)
    test_pps(dev_server, simple_pachyderm_env, notebook_path)
