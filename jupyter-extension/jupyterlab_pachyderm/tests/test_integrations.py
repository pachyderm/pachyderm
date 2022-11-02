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
    # TODO: uncomment when python pachyderm has Projects support
    # print("creating pachyderm resources")
    # import python_pachyderm
    # from python_pachyderm.pfs import Commit

    repos = ["images", "edges", "montage"]
    branches = ["master", "dev"]
    files = ["file1", "file2"]

    # client = python_pachyderm.Client()
    # client.delete_all()

    # for repo in repos:
    #     client.create_repo(repo)
    #     for branch in branches:
    #         for file in files:
    #             client.put_file_bytes(
    #                 Commit(repo=repo, branch=branch), file, value=b"some data"
    #             )

    yield repos, branches, files


@pytest.fixture()
def dev_server():
    print("starting development server...")
    p = subprocess.Popen(
        [sys.executable, "-m", "jupyterlab_pachyderm.dev_server"],
        env={"PFS_MOUNT_DIR": PFS_MOUNT_DIR},
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
    subprocess.run(
        ["bash", "-c", f"umount {PFS_MOUNT_DIR}"],
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
    )
    p.terminate()
    p.wait()
    time.sleep(1)

    if not running:
        raise RuntimeError("mount server is having issues starting up")


def test_list_repos(pachyderm_resources, dev_server):
    repos, branches, _ = pachyderm_resources

    r = requests.get(f"{BASE_URL}/repos")

    assert r.status_code == 200
    for _, repo_info in r.json().items():
        assert repo_info.keys() == {"authorization", "branches", "repo"}
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
    assert r.status_code == 200

    r = requests.get(f"{BASE_URL}/mounts")
    assert r.status_code == 200

    assert len(r.json()["mounted"]) == 1
    for _, mount_info in r.json()["mounted"].items():
        assert mount_info.keys() == {
            "name",
            "repo",
            "branch",
            "commit",
            "files",
            "glob",
            "mode",
            "state",
            "status",
            "mountpoint",
            "actual_mounted_commit",
            "latest_commit",
            "how_many_commits_behind",
        }

    for _, _repo_info in r.json()["unmounted"].items():
        assert _repo_info["repo"] in repos
        assert _repo_info.keys() == {"authorization", "branches", "repo"}
        for _branch in _repo_info["branches"]:
            assert _branch in branches
    assert len(r.json()["unmounted"]) == len(repos)


def test_mount(pachyderm_resources, dev_server):
    repos, _, files = pachyderm_resources

    to_mount = {
        "mounts": [
            {
                "name": repos[0],
                "repo": repos[0],
                "branch": "master",
                "mode": "ro",
            },
            {
                "name": repos[0] + "_dev",
                "repo": repos[0],
                "branch": "dev",
                "mode": "ro",
            },
            {
                "name": repos[1],
                "repo": repos[1],
                "branch": "master",
                "mode": "ro",
            },
        ]
    }
    r = requests.put(f"{BASE_URL}/_mount", data=json.dumps(to_mount))
    assert r.status_code == 200

    resp = r.json()
    assert len(resp["mounted"]) == 3
    assert len(list(os.walk(PFS_MOUNT_DIR))[0][1]) == 3
    for _, mount_info in resp["mounted"].items():
        assert sorted(
            list(os.walk(os.path.join(PFS_MOUNT_DIR, mount_info["name"])))[0][2]
        ) == sorted(files)
    assert len(resp["unmounted"]) == 3
    assert len(resp["unmounted"][repos[1]]["branches"]) == 2
    assert len(resp["unmounted"][repos[2]]["branches"]) == 2

    r = requests.put(
        f"{BASE_URL}/_unmount_all",
    )
    assert r.status_code == 200
    assert r.json()["mounted"] == {}
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
                "mode": "ro",
            },
            {
                "name": repos[0] + "_dev",
                "repo": repos[0],
                "branch": "dev",
                "mode": "ro",
            },
        ]
    }
    r = requests.put(f"{BASE_URL}/_mount", data=json.dumps(to_mount))
    assert r.status_code == 200
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
    assert r.status_code == 200
    assert len(r.json()["mounted"]) == 1
    assert len(r.json()["unmounted"]) == 3
    assert len(r.json()["unmounted"][repos[0]]["branches"]) == 2

    r = requests.put(
        f"{BASE_URL}/_unmount",
        data=json.dumps({"mounts": [repos[0]]}),
    )
    assert r.status_code == 200
    assert len(r.json()["mounted"]) == 0
    assert len(r.json()["unmounted"]) == 3
    assert len(r.json()["unmounted"][repos[0]]["branches"]) == 2
    assert list(os.walk(PFS_MOUNT_DIR)) == [(PFS_MOUNT_DIR, [], [])]


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

    r = requests.put(f"{BASE_URL}/_mount_datums", data=json.dumps(input_spec))
    assert r.status_code == 200
    assert r.json()["idx"] == 0
    assert r.json()["num_datums"] == 4
    list(os.walk(os.path.join(PFS_MOUNT_DIR, "out")))  # makes "out" dir appear
    assert len(list(os.walk(PFS_MOUNT_DIR))[0][1]) == 4
    datum0_id = r.json()["id"]

    assert sorted(list(os.walk(os.path.join(PFS_MOUNT_DIR, repos[0])))[0][2]) == sorted(
        files
    )
    assert repos[1] + "_dev" in list(os.walk(PFS_MOUNT_DIR))[0][1]
    assert len(list(os.walk(os.path.join(PFS_MOUNT_DIR, repos[2])))[0][2]) == 1

    r = requests.put(f"{BASE_URL}/_show_datum", params={"idx": "2"})
    assert r.status_code == 200
    assert r.json()["idx"] == 2
    assert r.json()["num_datums"] == 4
    assert r.json()["id"] != datum0_id

    r = requests.put(
        f"{BASE_URL}/_show_datum",
        params={
            "idx": "2",
            "id": datum0_id,
        },
    )
    assert r.status_code == 200
    assert r.json()["idx"] == 0
    assert r.json()["num_datums"] == 4
    assert r.json()["id"] == datum0_id

    r = requests.get(f"{BASE_URL}/datums")
    assert r.status_code == 200
    assert r.json()["input"] == input_spec["input"]
    assert r.json()["num_datums"] == 4
    assert r.json()["curr_idx"] == 0

    r = requests.put(f"{BASE_URL}/_unmount_all")
    assert r.status_code == 200


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
