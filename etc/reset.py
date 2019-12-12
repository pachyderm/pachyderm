#!/usr/bin/env python3

import os
import sys
import json
import time
import select
import shutil
import argparse
import threading
import subprocess
import urllib.request

ETCD_IMAGE = "quay.io/coreos/etcd:v3.3.5"

DELETABLE_RESOURCES = [
    "daemonsets",
    "replicasets",
    "services",
    "deployments",
    "pods",
    "rc",
    "pvc",
    "serviceaccounts",
    "secrets",
]

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

class DefaultDriver:
    def available(self):
        return True

    def clear(self):
        run("kubectl", "delete", ",".join(DELETABLE_RESOURCES), "--all")
        run("kubectl", "delete", "clusterrole.rbac.authorization.k8s.io/pachyderm")
        run("kubectl", "delete", "clusterrolebinding.rbac.authorization.k8s.io/pachyderm")

    def start(self):
        pass

    def push_images(self, deploy_version, dash_image):
        pass

    def set_config(self):
        kube_context = capture("kubectl", "config", "current-context").strip()
        run("pachctl", "config", "set", "context", kube_context, "--kubernetes", kube_context, "--overwrite")
        run("pachctl", "config", "set", "active-context", kube_context)

class DockerDriver(DefaultDriver):
    def set_config(self):
        super().set_config()
        run("pachctl", "config", "update", "context", "--pachd-address=localhost:30650")

class MinikubeDriver(DefaultDriver):
    def available(self):
        return run("which", "minikube", raise_on_error=False).returncode == 0

    def clear(self):
        run("minikube", "delete")

    def start(self):
        run("minikube", "start")

        while suppress("minikube", "status") != 0:
            print("Waiting for minikube to come up...")
            time.sleep(1)

    def push_images(self, deploy_version, dash_image):
        run("./etc/kube/push-to-minikube.sh", "pachyderm/pachd:{}".format(deploy_version))
        run("./etc/kube/push-to-minikube.sh", "pachyderm/worker:{}".format(deploy_version))
        run("./etc/kube/push-to-minikube.sh", ETCD_IMAGE)
        run("./etc/kube/push-to-minikube.sh", dash_image)

    def set_config(self):
        super().set_config()
        ip = capture("minikube", "ip")
        run("pachctl", "config", "update", "context", "--pachd-address={}".format(ip))

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

def run(cmd, *args, raise_on_error=True, stdin=None, capture_output=False):
    return subprocess.run([cmd, *args], check=raise_on_error, capture_output=capture_output, input=stdin, encoding="utf8")

def capture(cmd, *args):
    return run(cmd, *args, capture_output=True).stdout

def suppress(cmd, *args):
    return run(cmd, *args, capture_output=True, raise_on_error=False).returncode

def main():
    parser = argparse.ArgumentParser(description="Recompiles pachyderm tooling and restarts the cluster with a clean slate.")
    parser.add_argument("--no-config-rewrite", default=False, action="store_true", help="Disables config rewriting")
    parser.add_argument("--deploy-args", default="", help="Arguments to be passed into `pachctl deploy`")
    parser.add_argument("--deploy-to", default="local", help="Set where to deploy")
    parser.add_argument("--deploy-version", default="head", help="Sets the deployment version")
    args = parser.parse_args()

    if "GOPATH" not in os.environ:
        raise Exception("Must set GOPATH")
    if "PACH_CA_CERTS" in os.environ:
        raise Exception("Must unset PACH_CA_CERTS\nRun:\nunset PACH_CA_CERTS")

    if args.deploy_to == "local":
        if MinikubeDriver().available():
            print("using the minikube driver")
            driver = MinikubeDriver()
        else:
            print("using the k8s for docker driver")
            driver = DockerDriver()
    else:
        print("using the default driver")
        driver = DefaultDriver()

    driver.clear()

    gopath = os.environ["GOPATH"]

    if args.deploy_version == "head":
        try:
            os.remove(os.path.join(gopath, "bin", "pachctl"))
        except:
            pass

        procs = [
            driver.start,
            lambda: run("make", "install"),
            lambda: run("make", "docker-build"),
        ]

        join(*procs)
    else:
        should_download = suppress("which", "pachctl") != 0 \
            or capture("pachctl", "version", "--client-only") != args.deploy_version

        if should_download:
            release_url = "https://github.com/pachyderm/pachyderm/releases/download/v{}/pachctl_{}_{}_amd64.tar.gz".format(args.deploy_version, args.deploy_version, sys.platform)
            bin_path = os.path.join(os.environ["GOPATH"], "bin")

            with urllib.request.urlopen(release_url) as response:
                with gzip.GzipFile(fileobj=response) as uncompressed:
                    with open(bin_path, "wb") as f:
                        shutil.copyfileobj(uncompressed, f)

        run("docker", "pull", "pachyderm/pachd:{}".format(args.deploy_version))
        run("docker", "pull", "pachyderm/worker:{}".format(args.deploy_version))

    version = capture("pachctl", "version", "--client-only")
    print("Deploy pachyderm version v{}".format(version))

    run("which", "pachctl")

    deployments_str = capture("pachctl", "deploy", "local", "-d", "--dry-run")
    if args.deploy_to != "local":
        deployments_str = deployments_str.replace("local", args.deploy_to)
    deployments_json = json.loads("[{}]".format(deployments_str.replace("}\n{", "},{")))

    dash_image = find_in_json(deployments_json, lambda j: isinstance(j, dict) and j.get("name") == "dash" and j.get("image") is not None)["image"]
    grpc_proxy_image = find_in_json(deployments_json, lambda j: isinstance(j, dict) and j.get("name") == "grpc-proxy")["image"]

    run("docker", "pull", dash_image)
    run("docker", "pull", grpc_proxy_image)
    run("docker", "pull", ETCD_IMAGE)
    driver.push_images(args.deploy_version, dash_image)

    run("kubectl", "create", "-f", "-", stdin=deployments_str)

    while suppress("pachctl", "version") != 0:
        print("Waiting for pachyderm to come up...")
        time.sleep(1)

    driver.set_config()

if __name__ == "__main__":
    main()
