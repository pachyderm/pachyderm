#!/usr/bin/env python3

import os
import sys
import time
import select
import logging
import argparse
import threading
import subprocess

ETCD_IMAGE = "quay.io/coreos/etcd:v3.3.5"

LOG_LEVELS = {
    "critical": logging.CRITICAL,
    "error": logging.ERROR,
    "warning": logging.WARNING,
    "info": logging.INFO,
    "debug": logging.DEBUG,
}

LOG_COLORS = {
    "critical": "\x1b[31;1m",
    "error": "\x1b[31;1m",
    "warning": "\x1b[33;1m",
    #"info": "\x1b[32;1m",
    #"debug": "\x1b[35;1m",
}

log = logging.getLogger(__name__)
handler = logging.StreamHandler()
handler.setFormatter(logging.Formatter("%(levelname)s:%(message)s"))
log.addHandler(handler)

def parse_log_level(s):
    try:
        return LOG_LEVELS[s]
    except KeyError:
        raise Exception("Unknown log level: {}".format(s))

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

class Output:
    def __init__(self, pipe, level):
        self.pipe = pipe
        self.level = level
        self.lines = []

class ProcessResult:
    def __init__(self, rc, stdout, stderr):
        self.rc = rc
        self.stdout = stdout
        self.stderr = stderr

def redirect_to_logger(stdout, stderr):
    for io in select.select([stdout.pipe, stderr.pipe], [], [], 1000)[0]:
        line = io.readline().decode().rstrip()

        if line == "":
            continue

        dest = stdout if io == stdout.pipe else stderr
        log.log(LOG_LEVELS[dest.level], "{}{}\x1b[0m".format(LOG_COLORS.get(dest.level, ""), line))
        dest.lines.append(line)

def run(cmd, *args, raise_on_error=True, shell=False, stdout_log_level="info", stderr_log_level="error"):
    log.debug("Running `%s %s`", cmd, " ".join(args))

    proc = subprocess.Popen([cmd, *args], shell=shell, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    stdout = Output(proc.stdout, stdout_log_level)
    stderr = Output(proc.stderr, stderr_log_level)

    while proc.poll() is None:
        redirect_to_logger(stdout, stderr)
    redirect_to_logger(stdout, stderr)

    rc = proc.wait()

    if raise_on_error and rc != 0:
        raise Exception("Unexpected return code for `{} {}`: {}".format(cmd, " ".join(args), rc))

    return ProcessResult(rc, "\n".join(stdout.lines), "\n".join(stderr.lines))

def capture(cmd, *args, shell=False):
    return run(cmd, *args, shell=shell, stdout_log_level="debug").stdout

def suppress(cmd, *args):
    return run(cmd, *args, stdout_log_level="debug", raise_on_error=False).rc

def get_pachyderm(deploy_version):
    if deploy_version != "local":
        print("Deploying pachd:{}".format(deploy_version))

        should_download = suppress("which", "pachctl") != 0 \
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
    parser.add_argument("--deploy-args", default="", help="Arguments to be passed into `pachctl deploy`")
    parser.add_argument("--deploy-version", default="local", help="Sets the deployment version")
    parser.add_argument("--log-level", default="info", type=parse_log_level, help="Sets the log level; defaults to 'info', other options include 'critical', 'error', 'warning', and 'debug'")
    args = parser.parse_args()

    log.setLevel(args.log_level)

    if "GOPATH" not in os.environ:
        print("Must set GOPATH")
        sys.exit(1)
    if not args.no_deploy and "PACH_CA_CERTS" in os.environ:
        print("Must unset PACH_CA_CERTS\nRun:\nunset PACH_CA_CERTS", file=sys.stderr)
        sys.exit(1)
    if args.deploy_version == "local" and not os.getcwd().startswith(os.path.join(os.environ["GOPATH"], "src", "github.com", "pachyderm", "pachyderm")):
        print("Must be in a Pachyderm client", file=sys.stderr)
        sys.exit(1)

    reset_minikube = True
    if run("which", "minikube", raise_on_error=False).rc == 1:
        print("`minikube` not detected; running without minikube resets")
        reset_minikube = False

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
        if reset_minikube:
            procs.append(lambda: run("minikube", "start"))
        join(*procs)
    else:
        if reset_minikube:
            join(
                lambda: run("minikube", "start"),
                lambda: get_pachyderm(args.deploy_version),
            )
        else:
            get_pachyderm(args.deploy_version)

    version = capture("pachctl", "version", "--client-only")
    print("Deploy pachyderm version v{}".format(version))

    if reset_minikube:
        while suppress("minikube", "status") != 0:
            print("Waiting for minikube to come up...")
            time.sleep(1)

    while suppress("pachctl", "version", "--client-only") != 0:
        print("Waiting for pachctl to build...")
        time.sleep(1)

    # TODO: does this do anything when run in a python sub-process?
    # run("hash", "-r")

    run("which", "pachctl")

    dash_image = capture("pachctl deploy local -d --dry-run | jq -r '.. | select(.name? == \"dash\" and has(\"image\")).image'", shell=True)
    grpc_proxy_image = capture("pachctl deploy local -d --dry-run | jq -r '.. | select(.name? == \"grpc-proxy\").image'", shell=True)

    run("docker", "pull", dash_image)
    run("docker", "pull", grpc_proxy_image)
    run("docker", "pull", ETCD_IMAGE)

    if reset_minikube:
        run("./etc/kube/push-to-minikube.sh", "pachyderm/pachd:{}".format(args.deploy_version))
        run("./etc/kube/push-to-minikube.sh", "pachyderm/worker:{}".format(args.deploy_version))
        run("./etc/kube/push-to-minikube.sh", ETCD_IMAGE)
        run("./etc/kube/push-to-minikube.sh", dash_image)

    if not args.no_deploy:
        if args.deploy_version == "local":
            run("pachctl deploy local -d {}".format(args.deploy_args), shell=True)
        else:
            run("pachctl deploy local -d {} --dry-run | sed \"s/:local/:{}/g\" | kubectl create -f -".format(args.deploy_args, args.deploy_version), shell=True)

        while suppress("pachctl", "version") != 0:
            print("No pachyderm yet")
            time.sleep(1)

    run("killall", "kubectl", raise_on_error=False)

if __name__ == "__main__":
    main()
