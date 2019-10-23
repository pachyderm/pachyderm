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
}

log = logging.getLogger(__name__)
handler = logging.StreamHandler()
handler.setFormatter(logging.Formatter("%(levelname)s:%(message)s"))
log.addHandler(handler)

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

class DefaultDriver:
    def available(self):
        return True

    def clear(self):
        if run("yes | pachctl delete all", shell=True, raise_on_error=False).rc != 0:
            log.error("could not call `pachctl delete all`; most likely this just means that a pachyderm cluster hasn't been setup, but may indicate a bad state")

    def start(self):
        pass

    def push_images(self, deploy_version, dash_image):
        pass

    def wait(self):
        while suppress("pachctl", "version") != 0:
            log.info("Waiting for pachyderm to come up...")
            time.sleep(1)

    def set_config(self):
        run("pachctl", "config", "update", "context", "--pachd-address=localhost:30650")

class MinikubeDriver(DefaultDriver):
    def available(self):
        return run("which", "minikube", raise_on_error=False).rc == 0

    def clear(self):
        run("minikube", "delete")

    def start(self):
        run("minikube", "start")

        while suppress("minikube", "status") != 0:
            log.info("Waiting for minikube to come up...")
            time.sleep(1)

    def push_images(self, deploy_version, dash_image):
        run("./etc/kube/push-to-minikube.sh", "pachyderm/pachd:{}".format(deploy_version))
        run("./etc/kube/push-to-minikube.sh", "pachyderm/worker:{}".format(deploy_version))
        run("./etc/kube/push-to-minikube.sh", ETCD_IMAGE)
        run("./etc/kube/push-to-minikube.sh", dash_image)

    def set_config(self):
        ip = capture("minikube", "ip")
        run("pachctl", "config", "update", "context", "--pachd-address={}".format(ip))

def parse_log_level(s):
    try:
        return LOG_LEVELS[s]
    except KeyError:
        raise Exception("Unknown log level: {}".format(s))

def run(cmd, *args, raise_on_error=True, shell=False, stdout_log_level="info", stderr_log_level="error"):
    log.debug("Running `%s %s`", cmd, " ".join(args))

    proc = subprocess.Popen([cmd, *args], shell=shell, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    stdout = Output(proc.stdout, stdout_log_level)
    stderr = Output(proc.stderr, stderr_log_level)
    timed_out_last = False

    while True:
        if (proc.poll() is not None and timed_out_last) or (stdout.pipe.closed and stderr.pipe.closed):
            break

        for io in select.select([stdout.pipe, stderr.pipe], [], [], 100)[0]:
            timed_out_last = False
            line = io.readline().decode().rstrip()

            if line == "":
                continue

            dest = stdout if io == stdout.pipe else stderr
            log.log(LOG_LEVELS[dest.level], "{}{}\x1b[0m".format(LOG_COLORS.get(dest.level, ""), line))
            dest.lines.append(line)
        else:
            timed_out_last = True

    rc = proc.wait()

    if raise_on_error and rc != 0:
        raise Exception("Unexpected return code for `{} {}`: {}".format(cmd, " ".join(args), rc))

    return ProcessResult(rc, "\n".join(stdout.lines), "\n".join(stderr.lines))

def capture(cmd, *args, shell=False):
    return run(cmd, *args, shell=shell, stdout_log_level="debug").stdout

def suppress(cmd, *args):
    return run(cmd, *args, stdout_log_level="debug", stderr_log_level="debug", raise_on_error=False).rc

def get_pachyderm(deploy_version):
    log.info("Deploying pachd:{}".format(deploy_version))

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
    parser = argparse.ArgumentParser(description="Recompiles pachyderm tooling and restarts the cluster with a clean slate.")
    parser.add_argument("--no-deploy", default=False, action="store_true", help="Disables deployment")
    parser.add_argument("--deploy-args", default="", help="Arguments to be passed into `pachctl deploy`")
    parser.add_argument("--deploy-version", default="local", help="Sets the deployment version")
    parser.add_argument("--log-level", default="info", type=parse_log_level, help="Sets the log level; defaults to 'info', other options include 'critical', 'error', 'warning', and 'debug'")
    args = parser.parse_args()

    log.setLevel(args.log_level)

    if "GOPATH" not in os.environ:
        log.critical("Must set GOPATH")
        sys.exit(1)
    if not args.no_deploy and "PACH_CA_CERTS" in os.environ:
        log.critical("Must unset PACH_CA_CERTS\nRun:\nunset PACH_CA_CERTS", file=sys.stderr)
        sys.exit(1)

    if MinikubeDriver().available():
        log.info("using the minikube driver")
        driver = MinikubeDriver()
    else:
        log.info("using the k8s for docker driver")
        log.warning("with this driver, it's not possible to fully reset the cluster")
        driver = DefaultDriver()

    driver.clear()

    gopath = os.environ["GOPATH"]

    if args.deploy_version == "local":
        try:
            os.remove(os.path.join(gopath, "bin", "pachctl"))
        except:
            pass

        join(
            driver.start,
            lambda: run("make", "install"),
            lambda: run("make", "docker-build"),
        )
    else:
        join(
            driver.start,
            lambda: get_pachyderm(args.deploy_version),
        )

    version = capture("pachctl", "version", "--client-only")
    log.info("Deploy pachyderm version v{}".format(version))

    while suppress("pachctl", "version", "--client-only") != 0:
        log.info("Waiting for pachctl to build...")
        time.sleep(1)

    run("which", "pachctl")

    dash_image = capture("pachctl deploy local -d --dry-run | jq -r '.. | select(.name? == \"dash\" and has(\"image\")).image'", shell=True)
    grpc_proxy_image = capture("pachctl deploy local -d --dry-run | jq -r '.. | select(.name? == \"grpc-proxy\").image'", shell=True)

    run("docker", "pull", dash_image)
    run("docker", "pull", grpc_proxy_image)
    run("docker", "pull", ETCD_IMAGE)
    driver.push_images(args.deploy_version, dash_image)

    if not args.no_deploy:
        if args.deploy_version == "local":
            run("pachctl deploy local -d {}".format(args.deploy_args), shell=True)
        else:
            run("pachctl deploy local -d {} --dry-run | sed \"s/:local/:{}/g\" | kubectl create -f -".format(args.deploy_args, args.deploy_version), shell=True)

        driver.wait()

    run("killall", "kubectl", raise_on_error=False)
    driver.set_config()

if __name__ == "__main__":
    main()
