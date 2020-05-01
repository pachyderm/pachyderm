#!/usr/bin/env python3

import os
import re
import json
import asyncio
import secrets
import argparse
import collections

ETCD_IMAGE = "quay.io/coreos/etcd:v3.3.5"
IDE_USER_IMAGE = "pachyderm/ide-user:local"
IDE_HUB_IMAGE = "pachyderm/ide-hub:local"

DELETABLE_RESOURCES = [
    "roles.rbac.authorization.k8s.io",
    "rolebindings.rbac.authorization.k8s.io"
]

NEWLINE_SEPARATE_OBJECTS_PATTERN = re.compile(r"\}\n+\{", re.MULTILINE)

GCP_KUBE_CONTEXT_NAME_PATTERN = re.compile(r"gke_([^_]+)_(.+)")

RunResult = collections.namedtuple("RunResult", ["rc", "stdout", "stderr"])

class RedactedString(str):
    pass

class BaseDriver:
    async def start(self):
        pass
    
    async def clear(self):
        # Check for the presence of the pachyderm IDE to see whether it should
        # be undeployed too. Using kubectl rather than helm here because
        # this'll work even if the helm CLI is not installed.
        undeploy_args = []
        jupyterhub_apps = json.loads(await capture("kubectl", "get", "pod", "-lapp=jupyterhub", "-o", "json"))
        if len(jupyterhub_apps["items"]) > 0:
            undeploy_args.append("--ide")

         # ignore errors here because most likely no cluster is just deployed
         # yet
        await run("pachctl", "undeploy", "--metadata", *undeploy_args, stdin="y\n", raise_on_error=False)
        # clear out resources not removed from the undeploy process
        await run("kubectl", "delete", ",".join(DELETABLE_RESOURCES), "-l", "suite=pachyderm")

    async def init_image_registry(self):
        pass

    async def push_image(self, images):
        pass

    async def update_config(self):
        pass

    def extra_deploy_args(self):
        # We use hostpaths for storage. On docker for mac and minikube,
        # hostpaths aren't cleared until the VM is restarted. Because of this
        # behavior, re-deploying on the same hostpath without a restart will
        # cause us to bring up a new pachyderm cluster with access to the old
        # cluster volume, causing a bad state. This works around the issue by
        # just using a different hostpath on every deployment.
        return ["--host-path", "/var/pachyderm-{}".format(secrets.token_hex(5))]

    def ide_user_image(self):
        return IDE_USER_IMAGE

    def ide_hub_image(self):
        return IDE_HUB_IMAGE

class DockerDesktopDriver(BaseDriver):
    pass

class MinikubeDriver(BaseDriver):
    async def start(self):
        await run("minikube", "start")

    async def push_image(self, image):
        await run("./etc/kube/push-to-minikube.sh", image)

    async def update_config(self):
        ip = (await capture("minikube", "ip")).strip()
        await run("pachctl", "config", "update", "context", f"--pachd-address={ip}:30650")

class GCPDriver(BaseDriver):
    def __init__(self, project_id):
        self.project_id = project_id

    def _image(self, name):
        return f"gcr.io/{self.project_id}/{name}"

    async def clear(self):
        await super().clear()
        await run("kubectl", "delete", "secret", "regcred", raise_on_error=False)

    async def init_image_registry(self):
        docker_config_path = os.path.expanduser("~/.docker/config.json")
        await run("kubectl", "create", "secret", "generic", "regcred",
            f"--from-file=.dockerconfigjson={docker_config_path}",
            "--type=kubernetes.io/dockerconfigjson")

    async def push_image(self, image):
        image_url = self._image(image[8:] if image.startswith("quay.io/") else image)
        await run("docker", "tag", image, image_url)
        await run("docker", "push", image_url)

    def extra_deploy_args(self):
        return ["--image-pull-secret", "regcred", "--registry", f"gcr.io/{self.project_id}"]

    def ide_user_image(self):
        return self._image(IDE_USER_IMAGE)

    def ide_hub_image(self):
        return self._image(IDE_HUB_IMAGE)

async def run(cmd, *args, raise_on_error=True, stdin=None, capture_output=False, timeout=None):
    print_args = [cmd, *[a if not isinstance(a, RedactedString) else "[redacted]" for a in args]]
    print_status("running: `{}`".format(" ".join(print_args)))

    proc = await asyncio.create_subprocess_exec(
        cmd, *args,
        stdin=asyncio.subprocess.PIPE if stdin is not None else None,
        stdout=asyncio.subprocess.PIPE if capture_output else None,
        stderr=asyncio.subprocess.PIPE if capture_output else None,
    )
    
    if timeout is None:
        stdin = stdin.encode("utf8") if stdin is not None else None
        result = await proc.communicate(input=stdin)
    else:
        result = await asyncio.wait_for(proc.communicate(input=stdin), timeout=timeout)

    if capture_output:
        stdout, stderr = result
        stdout = stdout.decode("utf8")
        stderr = stderr.decode("utf8")
    else:
        stdout, stderr = None, None

    if raise_on_error and proc.returncode:
        raise Exception("unexpected return code from `{}`: {}".format(cmd, proc.returncode))

    return RunResult(rc=proc.returncode, stdout=stdout, stderr=stderr)

async def capture(cmd, *args, **kwargs):
    _, stdout, _ = await run(cmd, *args, capture_output=True, **kwargs)
    return stdout

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
    print(f"===> {status}")

async def main():
    parser = argparse.ArgumentParser(description="Resets a pachyderm cluster.")
    parser.add_argument("--dash", action="store_true", help="Deploy dash")
    parser.add_argument("--ide", action="store_true", help="Deploy IDE")
    parser.add_argument("--skip-deploy", action="store_true", help="Just delete, don't re-deploy")
    args = parser.parse_args()

    if "GOPATH" not in os.environ:
        raise Exception("Must set GOPATH")
    if "PACH_CA_CERTS" in os.environ:
        raise Exception("Must unset PACH_CA_CERTS\nRun:\nunset PACH_CA_CERTS")

    driver = None

    # derive which driver to use from the k8s context name
    kube_context = await capture("kubectl", "config", "current-context", raise_on_error=False)
    kube_context = kube_context.strip() if kube_context else ""
    if kube_context == "minikube":
        print_status("using the minikube driver")
        driver = MinikubeDriver()
    elif kube_context == "docker-desktop":
        print_status("using the docker desktop driver")
        driver = DockerDesktopDriver()
    else:
        match = GCP_KUBE_CONTEXT_NAME_PATTERN.match(kube_context)
        if match is not None:
            print_status("using the GKE driver")
            driver = GCPDriver(match.groups()[0])

    # minikube won't set the k8s context if the VM isn't running. This checks
    # for the presence of the minikube executable as an alternate means.
    if driver is None and (await run("minikube", "version", raise_on_error=False, capture_output=True)).rc == 0:
        print_status("using the minikube driver")
        driver = MinikubeDriver()

    if driver is None:
        raise Exception(f"could not derive driver from context name: {kube_context}")

    await driver.start()

    if args.skip_deploy:
        await driver.clear()
        return

    await asyncio.gather(
        run("make", "install"),
        run("make", "docker-build"),
        driver.init_image_registry(),
        driver.clear(),
    )
    
    version = (await capture("pachctl", "version", "--client-only")).strip()
    print_status(f"deploy pachyderm version v{version}")

    deployment_args = [
        "pachctl", "deploy", "local", "-d", "--dry-run", "--create-context", "--no-guaranteed",
        *driver.extra_deploy_args()
    ]
    if not args.dash:
        deployment_args.append("--no-dashboard")

    deployments_str = await capture(*deployment_args)
    deployments_json = json.loads("[{}]".format(NEWLINE_SEPARATE_OBJECTS_PATTERN.sub("},{", deployments_str)))

    dash_spec = find_in_json(deployments_json, lambda j: \
        isinstance(j, dict) and j.get("name") == "dash" and j.get("image") is not None)
    grpc_proxy_spec = find_in_json(deployments_json, lambda j: \
        isinstance(j, dict) and j.get("name") == "grpc-proxy")
    
    pull_images = [run("docker", "pull", ETCD_IMAGE)]
    if dash_spec is not None:
        pull_images.append(run("docker", "pull", dash_spec["image"]))
    if grpc_proxy_spec is not None:
        pull_images.append(run("docker", "pull", grpc_proxy_spec["image"]))
    await asyncio.gather(*pull_images)

    push_images = [ETCD_IMAGE, "pachyderm/pachd:local", "pachyderm/worker:local"]
    if dash_spec is not None:
        push_images.append(dash_spec["image"])
    if grpc_proxy_spec is not None:
        push_images.append(grpc_proxy_spec["image"])

    await asyncio.gather(*[driver.push_image(i) for i in push_images])
    await run("kubectl", "create", "-f", "-", stdin=deployments_str)
    await driver.update_config()

    while (await run("pachctl", "version", raise_on_error=False, capture_output=True)).rc:
        print_status("waiting for pachyderm to come up...")
        await asyncio.sleep(1)

    if args.ide:
        await asyncio.gather(*[driver.push_image(i) for i in [IDE_USER_IMAGE, IDE_HUB_IMAGE]])

        enterprise_token = await capture(
            "aws", "s3", "cp",
            "s3://pachyderm-engineering/test_enterprise_activation_code.txt",
            "-"
        )
        await run("pachctl", "enterprise", "activate", RedactedString(enterprise_token))
        await run("pachctl", "auth", "activate", stdin="admin\n")
        await run("pachctl", "deploy", "ide", 
            "--user-image", driver.ide_user_image(),
            "--hub-image", driver.ide_hub_image(),
        )

if __name__ == "__main__":
    asyncio.run(main(), debug=True)
