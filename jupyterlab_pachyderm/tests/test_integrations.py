import os
import sys
import subprocess
import time

import pytest
import requests

from jupyterlab_pachyderm.handlers import NAMESPACE, VERSION


ADDRESS = "http://localhost:8888"
MOUNT_DIR = os.environ.get("PFS_MOUNT_DIR", "/pfs")
BASE_URL = f"{ADDRESS}/{NAMESPACE}/{VERSION}"


@pytest.fixture(scope="module")
def pachyderm_resources():
    print("creating pachyderm resources")
    import python_pachyderm
    from python_pachyderm.pfs import Commit

    repos = ["images", "edges", "montage"]
    files = ["file1", "file2"]

    client = python_pachyderm.Client()
    client.delete_all()

    for repo in repos:
        client.create_repo(repo)
        for file in files:
            client.put_file_bytes(
                Commit(repo=repo, branch="master"), file, value=b"some data"
            )
    yield repos, files


@pytest.fixture
def dev_server():
    print("starting development server")
    p = subprocess.Popen(
        [sys.executable, "-m", "jupyterlab_pachyderm.dev_server"],
        env={"PFS_MOUNT_DIR": MOUNT_DIR},
        stdout=subprocess.PIPE,
    )

    stdout, _ = p.communicate()
    print(stdout)
    yield p
    print("killing development server")
    p.kill()
    time.sleep(1)


def test_list_repos(pachyderm_resources):
    repos, files = pachyderm_resources

    r = requests.get(f"{BASE_URL}/repos")

    assert r.status_code == 200
    for _repo in r.json():
        assert _repo["repo"] in repos
        for _branch in _repo["branches"]:
            assert _branch.keys() == {"branch", "mount"}
            assert _branch["mount"].keys() == {
                "name",
                "state",
                "status",
                "mode",
                "mountpoint",
            }


def test_mount(pachyderm_resources):
    repos, files = pachyderm_resources

    # populate in-memory map of repos and mount states
    # TODO automatically call list repos for mount
    requests.get(f"{BASE_URL}/repos")

    for repo in repos:
        r = requests.put(
            f"{BASE_URL}/repos/{repo}/_mount",
            params={"name": repo},
        )
        assert r.status_code == 200
        assert sorted(list(os.walk(os.path.join(MOUNT_DIR, repo)))[0][2]) == sorted(
            files
        )

    # unmount
    r = requests.put(
        f"{BASE_URL}/repos/_unmount",
    )
    assert r.status_code == 200
    assert list(os.walk(MOUNT_DIR)) == [(MOUNT_DIR, [], [])]
