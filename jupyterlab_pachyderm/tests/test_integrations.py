import os
import sys
import subprocess
import time
import json

import pytest
import requests
import python_pachyderm
from python_pachyderm.service import identity_proto, auth_proto

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
    # Give time for python server and Go server to start
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

    subprocess.run(["fusermount", "-u", PFS_MOUNT_DIR])
    p.terminate()
    p.wait()
    time.sleep(1)

    if not running:
        raise RuntimeError("mount server is having issues starting up")


@pytest.fixture()
def activate_auth():
    print("activating auth...")
    client = python_pachyderm.Client()
    client.delete_all()

    client.activate_license(os.environ.get("ENT_ACT_CODE"))
    client.add_cluster("localhost", "localhost:1650", secret="secret")
    client.activate_enterprise("localhost:1650", "localhost", "secret")

    client.auth_token = ROOT_TOKEN
    client.activate_auth(ROOT_TOKEN)
    client.set_identity_server_config(
        config=identity_proto.IdentityServerConfig(issuer="http://localhost:1658")
    )
    client.set_auth_configuration(
        auth_proto.OIDCConfig(
            issuer="http://localhost:1658",
            client_id="client",
            client_secret="secret",
            redirect_uri="http://test.example.com",
        )
    )

    config = json.load(open(os.path.expanduser(CONFIG_PATH)))
    active_context = config["v2"]["active_context"]
    config["v2"]["contexts"][active_context]["session_token"] = ROOT_TOKEN
    json.dump(config, open(os.path.expanduser(CONFIG_PATH), "w"))

    # Reload mount server so it sees auth token
    subprocess.run(["fusermount", "-u", PFS_MOUNT_DIR])

    yield

    client.auth_token = ROOT_TOKEN
    client.delete_all_identity()
    client.deactivate_auth()
    client.deactivate_enterprise()
    client.delete_all_license()


def test_list_repos(pachyderm_resources, dev_server):
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
                "mount_key",
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
        assert sorted(list(os.walk(os.path.join(PFS_MOUNT_DIR, repo)))[0][2]) == sorted(
            files
        )

    r = requests.put(
        f"{BASE_URL}/repos/_unmount",
    )
    assert r.status_code == 200

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
    assert list(os.walk(PFS_MOUNT_DIR)) == [(PFS_MOUNT_DIR, [], [])]


def test_config(dev_server):
    # PUT request
    test_endpoint = "localhost:30650"
    r = requests.put(f"{BASE_URL}/config", data=json.dumps({"pachd_address": test_endpoint}))

    config = json.load(open(os.path.expanduser(CONFIG_PATH)))
    active_context = config["v2"]["active_context"]
    try:
        endpoint_in_config = config["v2"]["contexts"][active_context]["pachd_address"]
    except:
        endpoint_in_config = str(config["v2"]["contexts"][active_context]["port_forwarders"]["pachd"])

    assert r.status_code == 200
    assert r.json()["cluster_status"] != "INVALID"
    assert "30650" in endpoint_in_config

    # GET request
    r = requests.get(f"{BASE_URL}/config")

    assert r.status_code == 200
    assert r.json()["cluster_status"] != "INVALID"


@pytest.mark.skipif(os.environ.get("ENT_ACT_CODE") is None, reason="No enterprise token at env var ENT_ACT_CODE")
def test_auth(pachyderm_resources, activate_auth, dev_server):
    r = requests.put(f"{BASE_URL}/auth/_logout")
    assert r.status_code == 200

    config = json.load(open(os.path.expanduser(CONFIG_PATH)))
    active_context = config["v2"]["active_context"]
    assert config["v2"]["contexts"][active_context].get("session_token") is None

    # make login request and verify URL returned
    r = requests.put(f"{BASE_URL}/auth/_login")
    assert r.status_code == 200
    assert r.json()["auth_url"] is not None
