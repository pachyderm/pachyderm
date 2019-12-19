#!/usr/bin/env python3

import os
import re
import json
import time
import argparse
import threading
import subprocess

ETCD_IMAGE = "quay.io/coreos/etcd:v3.3.5"

DELETABLE_RESOURCES = [
    "replicasets",
    "services",
    "deployments",
    "pods",
    "rc",
    "serviceaccounts",
    "secrets",
    "clusterrole",
    "clusterrolebinding",
]

NEWLINE_SEPARATE_OBJECTS_PATTERN = re.compile(r"\}\n+\{", re.MULTILINE)

class ExcThread(threading.Thread):
    def __init__(self, target):
        super().__init__(target=target)
        self.error = None

    def run(self):
        try:
            self._target()
        except Exception as e:
            self.error = e

def join(*targets):
    threads = []

    for target in targets:
        t = ExcThread(target)
        t.start()
        threads.append(t)

    for t in threads:
        t.join()
    for t in threads:
        if t.error is not None:
            raise Exception("Thread error") from t.error

class DockerDesktopDriver:
    def clear(self):
        run("kubectl", "delete", ",".join(DELETABLE_RESOURCES), "-l", "suite=pachyderm")

    def start(self):
        pass

    def push_images(self, dash_image):
        pass

    def update_config(self):
        run("pachctl", "config", "update", "context", "--pachd-address=localhost:30650")

class MinikubeDriver:
    def clear(self):
        run("minikube", "delete")

    def start(self):
        run("minikube", "start")

        while run("minikube", "status", raise_on_error=False).returncode != 0:
            print("Waiting for minikube to come up...")
            time.sleep(1)

    def push_images(self, dash_image):
        run("./etc/kube/push-to-minikube.sh", "pachyderm/pachd:local")
        run("./etc/kube/push-to-minikube.sh", "pachyderm/worker:local")
        run("./etc/kube/push-to-minikube.sh", ETCD_IMAGE)
        run("./etc/kube/push-to-minikube.sh", dash_image)

    def update_config(self):
        ip = capture("minikube", "ip").strip()
        run("pachctl", "config", "update", "context", "--pachd-address={}:30650".format(ip))

def find_in_json(j, f):
    if f(j):
        return j

    iter = None
    if isinstance(j, dict):
        iter = j.values()
    elif isinstance(j, list):
        iter = j

    if iter is not None:
        for sub_j in iter:
            v = find_in_json(sub_j, f)
            if v is not None:
                return v

def print_status(status):
    print("===> {}".format(status))

def run(cmd, *args, raise_on_error=True, stdin=None, capture_output=False, timeout=None):
    all_args = [cmd, *args]
    print_status("running: `{}`".format(" ".join(all_args)))
    return subprocess.run(all_args, check=raise_on_error, capture_output=capture_output, input=stdin, encoding="utf8", timeout=timeout)

def capture(cmd, *args):
    return run(cmd, *args, capture_output=True).stdout

def main():
    parser = argparse.ArgumentParser(description="Recompiles pachyderm tooling and restarts the cluster with a clean slate.")
    parser.add_argument("--args", default="", help="Arguments to be passed into `pachctl deploy`")
    args = parser.parse_args()

    if "GOPATH" not in os.environ:
        raise Exception("Must set GOPATH")
    if "PACH_CA_CERTS" in os.environ:
        raise Exception("Must unset PACH_CA_CERTS\nRun:\nunset PACH_CA_CERTS")

    kube_context = capture("kubectl", "config", "current-context").strip()

    if kube_context == "minikube":
        print_status("using the minikube driver")
        driver = MinikubeDriver()
    elif kube_context == "docker-desktop":
        print_status("using the docker desktop driver")
        driver = DockerDesktopDriver()

    # ignore errors on `pachctl delete all` because there's a variety of ways
    # it could fail that we can recover from:
    # 1) pachyderm isn't installed on the cluster yet
    # 2) pachctl isn't installed
    # 3) pachd is unresponsive
    # ... etc. Wrap in try/except rather than using `raise_on_error`, because
    # the latter doesn't catch when `pachctl` doesn't exist -- an error case
    # which we also want to ignore.
    try:
        run("pachctl", "delete", "all", stdin="yes\n", timeout=5)
    except:
        pass

    driver.clear()

    bin_path = os.path.join(os.environ["GOPATH"], "bin", "pachctl")

    if os.path.exists(bin_path):
        os.remove(bin_path)

    join(
        driver.start,
        lambda: run("make", "install"),
        lambda: run("make", "docker-build"),
    )
    
    version = capture("pachctl", "version", "--client-only")
    print_status("deploy pachyderm version v{}".format(version))

    deployments_str = capture("pachctl", "deploy", "local", "-d", "--dry-run")
    deployments_json = json.loads("[{}]".format(NEWLINE_SEPARATE_OBJECTS_PATTERN.sub("},{", deployments_str)))
    dash_image = find_in_json(deployments_json, lambda j: isinstance(j, dict) and j.get("name") == "dash" and j.get("image") is not None)["image"]
    grpc_proxy_image = find_in_json(deployments_json, lambda j: isinstance(j, dict) and j.get("name") == "grpc-proxy")["image"]

    run("docker", "pull", dash_image)
    run("docker", "pull", grpc_proxy_image)
    run("docker", "pull", ETCD_IMAGE)
    driver.push_images(dash_image)

    run("kubectl", "create", "-f", "-", stdin=deployments_str)
    driver.update_config()

    while run("pachctl", "version", raise_on_error=False).returncode:
        print_status("waiting for pachyderm to come up...")
        time.sleep(1)

if __name__ == "__main__":
    main()
