import os
import sys
import subprocess
import time
import json

import pytest
import requests

from jupyterlab_pachyderm.handlers import NAMESPACE, VERSION
from jupyterlab_pachyderm.env import PFS_MOUNT_DIR


ADDRESS = "http://localhost:8888"
BASE_URL = f"{ADDRESS}/{NAMESPACE}/{VERSION}"
CONFIG_PATH = "~/.pachyderm/config.json"
ROOT_TOKEN = "iamroot"


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


@pytest.fixture(scope="module")
def dev_server():
    print("starting development server...")
    p = subprocess.Popen(
        [sys.executable, "-m", "jupyterlab_pachyderm.dev_server"],
        env={"PFS_MOUNT_DIR": PFS_MOUNT_DIR},
        stdout=subprocess.PIPE,
    )
    # Give time for python test server to start
    time.sleep(3)

    payload = json.dumps({"pachd_address": "localhost:30650"})
    r = requests.put(f"{BASE_URL}/config", data=payload)

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

    subprocess.run(["bash", "-c", f"umount {PFS_MOUNT_DIR}"])
    p.terminate()
    p.wait()
    time.sleep(1)

    if not running:
        raise RuntimeError("mount server is having issues starting up")


def test_list_repos(pachyderm_resources, dev_server):
    repos, files = pachyderm_resources

    r = requests.get(f"{BASE_URL}/repos")

    assert r.status_code == 200
    for _repo in r.json():
        assert _repo.keys() == {"authorization", "branches", "repo"}
        assert _repo["repo"] in repos
        for _branch in _repo["branches"]:
            assert _branch.keys() == {"branch", "mount"}
            assert _branch["mount"][0].keys() == {
                "name",
                "state",
                "status",
                "mode",
                "mountpoint",
                "mount_key",
                "actual_mounted_commit",
                "latest_commit",
                "how_many_commits_behind",
            }


def test_mount(pachyderm_resources, dev_server):
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
        assert len(r.json()) == 3

        for repo_info in r.json():
            if repo_info["repo"] == repo:
                assert repo_info["branches"][0]["mount"][0]["state"] == "mounted"

        assert sorted(list(os.walk(os.path.join(PFS_MOUNT_DIR, repo)))[0][2]) == sorted(
            files
        )

    r = requests.put(
        f"{BASE_URL}/repos/_unmount",
    )
    assert r.status_code == 200
    assert len(r.json()) == 3
    assert r.json()[0]["branches"][0]["mount"][0]["state"] == "unmounted"

    assert list(os.walk(PFS_MOUNT_DIR)) == [(PFS_MOUNT_DIR, [], [])]


def test_unmount(pachyderm_resources, dev_server):
    repos, files = pachyderm_resources

    # populate in-memory map of repos and mount states
    # TODO automatically call list repos for mount
    requests.get(f"{BASE_URL}/repos")

    r = requests.put(
        f"{BASE_URL}/repos/images/_mount",
        params={"name": "images"},
    )
    assert r.status_code == 200
    assert sorted(list(os.walk(os.path.join(PFS_MOUNT_DIR, "images")))[0][2]) == sorted(
        files
    )

    # unmount
    r = requests.put(
        f"{BASE_URL}/repos/images/_unmount",
        params={"name": "images"},
    )
    assert r.status_code == 200

    for repo_info in r.json():
        if repo_info["repo"] == "images":
            assert repo_info["branches"][0]["mount"][0]["state"] == "unmounted"

    assert list(os.walk(PFS_MOUNT_DIR)) == [(PFS_MOUNT_DIR, [], [])]


def test_config(dev_server):
    # PUT request
    test_endpoint = "localhost:30650"
    r = requests.put(
        f"{BASE_URL}/config", data=json.dumps({"pachd_address": test_endpoint})
    )

    config = json.load(open(os.path.expanduser(CONFIG_PATH)))
    active_context = config["v2"]["active_context"]
    try:
        endpoint_in_config = config["v2"]["contexts"][active_context]["pachd_address"]
    except:
        endpoint_in_config = str(
            config["v2"]["contexts"][active_context]["port_forwarders"]["pachd"]
        )

    assert r.status_code == 200
    assert r.json()["cluster_status"] != "INVALID"
    assert "30650" in endpoint_in_config

    # GET request
    r = requests.get(f"{BASE_URL}/config")

    assert r.status_code == 200
    assert r.json()["cluster_status"] != "INVALID"
