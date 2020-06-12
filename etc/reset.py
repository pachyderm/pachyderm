#!/usr/bin/env python3

import os
import re
import json
import asyncio
import secrets
import argparse
import collections

ETCD_IMAGE = "pachyderm/etcd:v3.3.5"
IDE_USER_IMAGE = "pachyderm/ide-user:local"
IDE_HUB_IMAGE = "pachyderm/ide-hub:local"

DELETABLE_RESOURCES = [
    "roles.rbac.authorization.k8s.io",
    "rolebindings.rbac.authorization.k8s.io"
]

NEWLINE_SEPARATE_OBJECTS_PATTERN = re.compile(r"\}\n+\{", re.MULTILINE)

RunResult = collections.namedtuple("RunResult", ["rc", "stdout", "stderr"])

class RedactedString(str):
    pass

class BaseDriver:
    async def reset(self):
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

    async def push_image(self, images):
        pass

    async def update_config(self):
        pass

    def deploy_args(self):
        # We use hostpaths for storage. On docker for mac and minikube,
        # hostpaths aren't cleared until the VM is restarted. Because of this
        # behavior, re-deploying on the same hostpath without a restart will
        # cause us to bring up a new pachyderm cluster with access to the old
        # cluster volume, causing a bad state. This works around the issue by
        # just using a different hostpath on every deployment.
        return ["local", "-d", "--no-guaranteed", f"--host-path=/var/pachyderm-{secrets.token_hex(5)}"]

    def ide_user_image(self):
        return IDE_USER_IMAGE

    def ide_hub_image(self):
        return IDE_HUB_IMAGE

class DockerDesktopDriver(BaseDriver):
    pass

class MinikubeDriver(BaseDriver):
    async def reset(self):
        async def is_minikube_running():
            proc = await run("minikube", "status", raise_on_error=False, capture_output=True)
            return proc.rc == 0

        if not (await is_minikube_running()):
            await run("minikube", "start")
            while not (await is_minikube_running()):
                print("Waiting for minikube to come up...")
                await asyncio.sleep(1)

        await super().reset()

    async def push_image(self, image):
        await run("./etc/kube/push-to-minikube.sh", image)

    async def update_config(self):
        ip = (await capture("minikube", "ip")).strip()
        await run("pachctl", "config", "update", "context", f"--pachd-address={ip}:30650")

class GCPDriver(BaseDriver):
    def __init__(self, project_id, cluster_name=None):
        if cluster_name is None:
            cluster_name = f"pach-{secrets.token_hex(5)}"
        self.cluster_name = cluster_name
        self.object_storage_name = f"{cluster_name}-storage"
        self.project_id = project_id

    def image(self, name):
        return f"gcr.io/{self.project_id}/{name}"

    async def reset(self):
        cluster_exists = (await run("gcloud", "container", "clusters", "describe", self.cluster_name,
            raise_on_error=False, capture_output=True)).rc == 0

        if cluster_exists:
            await super().reset()
        else:
            await run("gcloud", "config", "set", "container/cluster", self.cluster_name)
            await run("gcloud", "container", "clusters", "create", self.cluster_name, "--scopes=storage-rw",
                "--machine-type=n1-standard-8", "--num-nodes=2")

            account = (await run("gcloud", "config", "get-value", "account")).strip()
            await run("kubectl", "create", "clusterrolebinding", "cluster-admin-binding",
                "--clusterrole=cluster-admin", f"--user={account}")

            await run("gsutil", "mb", f"gs://{self.object_storage_name}")

            docker_config_path = os.path.expanduser("~/.docker/config.json")
            await run("kubectl", "create", "secret", "generic", "regcred",
                f"--from-file=.dockerconfigjson={docker_config_path}",
                "--type=kubernetes.io/dockerconfigjson")

    async def push_image(self, image):
        image_url = self.image(image)
        if ":local" in image_url:
            version_tag = (await capture("pachctl", "version", "--client-only")).strip()
            image_url = image_url.replace(":local", f":{version_tag}")
        await run("docker", "tag", image, image_url)
        await run("docker", "push", image_url)

    def deploy_args(self):
        return ["google", self.object_storage_name, "32", "--dynamic-etcd-nodes=1", "--image-pull-secret=regcred",
            f"--registry=gcr.io/{self.project_id}"]

    def ide_user_image(self):
        return self.image(IDE_USER_IMAGE)

    def ide_hub_image(self):
        return self.image(IDE_HUB_IMAGE)

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
        raise Exception(f"unexpected return code from `{cmd}`: {proc.returncode}")

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
    parser.add_argument("--target", default="local", help="Where to deploy ('gcp[:<cluster name>]' or 'local')")
    parser.add_argument("--dash", action="store_true", help="Deploy dash")
    parser.add_argument("--ide", action="store_true", help="Deploy IDE")
    args = parser.parse_args()

    if "GOPATH" not in os.environ:
        raise Exception("Must set GOPATH")
    if "PACH_CA_CERTS" in os.environ:
        raise Exception("Must unset PACH_CA_CERTS\nRun:\nunset PACH_CA_CERTS")

    driver = None

    # derive which driver to use from the k8s context name
    if args.target == "local":
        kube_context = await capture("kubectl", "config", "current-context", raise_on_error=False)
        kube_context = kube_context.strip() if kube_context else ""
        if kube_context == "minikube":
            print_status("using the minikube driver")
            driver = MinikubeDriver()
        elif kube_context == "docker-desktop":
            print_status("using the docker desktop driver")
            driver = DockerDesktopDriver()
        # minikube won't set the k8s context if the VM isn't running. This checks
        # for the presence of the minikube executable as an alternate means.
        if driver is None and (await run("minikube", "version", raise_on_error=False, capture_output=True)).rc == 0:
            print_status("using the minikube driver")
            driver = MinikubeDriver()
    elif args.target.startswith("gcp"):
        project_id = (await capture("gcloud", "config", "get-value", "project")).strip()
        target_parts = args.target.split(":", maxsplit=1)
        cluster_name = target_parts[1] if len(target_parts) == 2 else None
        driver = GCPDriver(project_id, cluster_name=cluster_name)
    else:
        raise Exception(f"unknown target: {args.target}")

    if driver is None:
        raise Exception(f"could not derive driver from context name: {kube_context}")

    await asyncio.gather(
        run("make", "install"),
        run("make", "docker-build"),
        driver.reset(),
    )
    
    version = (await capture("pachctl", "version", "--client-only")).strip()
    print_status(f"deploy pachyderm version v{version}")

    deploy_args = ["pachctl", "deploy", *driver.deploy_args(), "--dry-run", "--create-context", "--log-level=debug"]
    if not args.dash:
        deploy_args.append("--no-dashboard")

    deployments_str = await capture(*deploy_args)
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
