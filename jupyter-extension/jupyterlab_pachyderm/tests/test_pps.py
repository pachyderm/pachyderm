"""Tests for the "Publish" view of the extension."""
import json
import os
import time
from datetime import datetime
from pathlib import Path
from random import randint
from typing import Tuple

import pytest
from httpx import AsyncClient

from pachyderm_sdk import Client
from pachyderm_sdk.api import pfs, pps

from jupyterlab_pachyderm.pps_client import METADATA_KEY, PpsConfig
from jupyterlab_pachyderm.tests import DEFAULT_PROJECT, TEST_NOTEBOOK


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
