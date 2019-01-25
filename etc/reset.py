#!/usr/bin/env python3

import os
import sys
import time
import argparse
import threading
import subprocess

ETCD_IMAGE = "quay.io/coreos/etcd:v3.3.5"

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

def run(cmd, *args, raise_on_error=True, shell=False):
    proc = subprocess.run([cmd, *args], shell=shell)

    if raise_on_error:
        proc.check_returncode()

    return proc

def suppress(cmd, *args, raise_on_error=True, shell=False):
    with open(os.devnull, "w") as f:
        return subprocess.run([cmd, *args], shell=shell, stdout=f)

def capture(cmd, *args, shell=False):
    return subprocess.check_output([cmd, *args], shell=shell).decode("utf-8").rstrip()

def get_pachyderm(deploy_version):
    if deploy_version != "local":
        print("Deploying pachd:{}".format(deploy_version))

        should_download = run("which", "pachctl", raise_on_error=False).returncode != 0 \
            or capture("pachctl", "version", "--client-only") != deploy_version

        if should_download:
            release_url = "https://github.com/pachyderm/pachyderm/releases/download/v{}/pachctl_{}_linux_amd64.tar.gz".format(deploy_version, deploy_version)
            outpath = os.path.join(os.environ["GOPATH"], "bin")
            filepath = "pachctl_{}_linux_amd64/pachctl".format(deploy_version)
            run("curl -L {} | tar -C \"{}\" --strip-components=1 -xzf - {}".format(release_url, outpath, filepath), shell=True)

        run("docker", "pull", "pachyderm/pachd:{}".format(deploy_version))
        run("docker", "pull", "pachyderm/worker:{}".format(deploy_version))

def main():
    parser = argparse.ArgumentParser(description="Resets the pachyderm cluster and tools.")
    parser.add_argument("--no-deploy", default=False, action="store_true", help="Disables deployment")
    parser.add_argument("--no-minikube", default=False, action="store_true", help="Do not use minikube; used if you're on a non-minikube k8s cluster")
    parser.add_argument("--deploy-args", default="", help="Arguments to be passed into `pachctl deploy`")
    parser.add_argument("--deploy-version", default="local", help="Sets the deployment version")
    args = parser.parse_args()

    if "GOPATH" not in os.environ:
        print("Must set GOPATH")
        sys.exit(1)
    if "ADDRESS" not in os.environ:
        print("Must set ADDRESS\nRun:\nexport ADDRESS=192.168.99.100:30650", file=sys.stderr)
        sys.exit(1)
    if not args.no_deploy and "PACH_CA_CERTS" in os.environ:
        print("Must unset PACH_CA_CERTS\nRun:\nunset PACH_CA_CERTS", file=sys.stderr)
        sys.exit(1)
    if args.deploy_version == "local" and not os.getcwd().startswith(os.path.join(os.environ["GOPATH"], "src", "github.com", "pachyderm", "pachyderm")):
        print("Must be in a Pachyderm client", file=sys.stderr)
        sys.exit(1)

    gopath = os.environ["GOPATH"]

    if args.deploy_version == "local":
        os.chdir(os.path.join(gopath, "src", "github.com", "pachyderm", "pachyderm"))
        
        try:
            os.remove(os.path.join(gopath, "bin", "pachctl"))
        except:
            pass

        procs = [
            lambda: run("make", "install"),
            lambda: run("make", "docker-build"),
        ]
        if not args.no_minikube:
            procs.append(lambda: run("minikube", "start"))
        join(*procs)
    else:
        if not args.no_minikube:
            join(
                lambda: run("minikube", "start"),
                lambda: get_pachyderm(args.deploy_version),
            )
        else:
            get_pachyderm(args.deploy_version)

    version = capture("pachctl", "version", "--client-only")
    print("Deploy pachyderm version v{}".format(version))

    if not args.no_minikube:
        while suppress("minikube", "status").returncode != 0:
            print("Waiting for minikube to come up...")
            time.sleep(1)

    while suppress("pachctl", "version", "--client-only").returncode != 0:
        print("Waiting for pachctl to build...")
        time.sleep(1)

    # TODO: does this do anything when run in a python sub-process?
    run("hash", "-r")

    run("which", "pachctl")

    dash_image = capture("pachctl deploy local -d --dry-run | jq -r '.. | select(.name? == \"dash\" and has(\"image\")).image'", shell=True)
    grpc_proxy_image = capture("pachctl deploy local -d --dry-run | jq -r '.. | select(.name? == \"grpc-proxy\").image'", shell=True)

    run("docker", "pull", dash_image)
    run("docker", "pull", grpc_proxy_image)
    run("docker", "pull", ETCD_IMAGE)

    if not args.no_minikube:
        run("./etc/kube/push-to-minikube.sh", "pachyderm/pachd:{}".format(args.deploy_version))
        run("./etc/kube/push-to-minikube.sh", "pachyderm/worker:{}".format(args.deploy_version))
        run("./etc/kube/push-to-minikube.sh", ETCD_IMAGE)
        run("./etc/kube/push-to-minikube.sh", dash_image)

    if not args.no_deploy:
        if args.deploy_version == "local":
            run("pachctl deploy local -d {}".format(args.deploy_args), shell=True)
        else:
            run("pachctl deploy local -d {} --dry-run | sed \"s/:local/:{}/g\" | kubectl create -f -".format(args.deploy_args, args.deploy_version), shell=True)

        while suppress("pachctl", "version").returncode != 0:
            print("No pachyderm yet")
            time.sleep(1)

    run("killall", "kubectl", raise_on_error=False)

if __name__ == "__main__":
    main()
