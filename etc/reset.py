#!/usr/bin/env python3

import os
import re
import json
import asyncio
import secrets
import argparse
import http.client

# a way to import util.py without adding it to the PYTHONPATH :o
import pathlib
import importlib.util
util_spec = importlib.util.spec_from_file_location("util", pathlib.Path(__file__).parent / "util.py")
util = importlib.util.module_from_spec(util_spec)
util_spec.loader.exec_module(util)

ETCD_IMAGE = "pachyderm/etcd:v3.3.5"
IDE_USER_IMAGE = "pachyderm/ide-user:local"
IDE_HUB_IMAGE = "pachyderm/ide-hub:local"

# Kubernetes resources to delete that aren't removed via `pachctl undeploy`
DELETABLE_RESOURCES = [
    "roles.rbac.authorization.k8s.io",
    "rolebindings.rbac.authorization.k8s.io"
]

# Matches the end of one JSON object and the beginning of another
NEWLINE_SEPARATE_OBJECTS_PATTERN = re.compile(r"\}\n+\{", re.MULTILINE)

class BaseDriver:
    """
    The base driver. Others drivers build off this. Additionally, it can be
    used directly for less prescriptive local kubernetes environments, e.g.
    docker for mac.
    """

    def image(self, name):
        return name

    async def reset(self):
        # Check for the presence of the pachyderm IDE to see whether it should
        # be undeployed too. Using kubectl rather than helm here because
        # this'll work even if the helm CLI is not installed.
        undeploy_args = []
        jupyterhub_apps = json.loads(await util.capture("kubectl", "get", "pod", "-lapp=jupyterhub", "-o", "json"))
        if len(jupyterhub_apps["items"]) > 0:
            undeploy_args.append("--ide")

         # ignore errors here because most likely no cluster is just deployed
         # yet
        await util.run("pachctl", "undeploy", "--metadata", *undeploy_args, stdin="y\n", raise_on_error=False)
        # clear out resources not removed from the undeploy process
        await util.run("kubectl", "delete", ",".join(DELETABLE_RESOURCES), "-l", "suite=pachyderm")

    async def push_image(self, images):
        pass

    def deploy_args(self):
        # We use hostpaths for storage. On docker for mac and minikube,
        # hostpaths aren't cleared until the VM is restarted. Because of this
        # behavior, re-deploying on the same hostpath without a restart will
        # cause us to bring up a new pachyderm cluster with access to the old
        # cluster volume, causing a bad state. This works around the issue by
        # just using a different hostpath on every deployment.
        return ["local", "-d", "--no-guaranteed", f"--host-path=/var/pachyderm-{secrets.token_hex(5)}"]

    async def update_config(self):
        pass

class MinikubeDriver(BaseDriver):
    """Driver for deploying to minikube"""

    async def reset(self):
        async def minikube_status():
            await util.run("minikube", "status", capture_output=True)

        is_minikube_running = True
        try:
            await minikube_status()
        except:
            is_minikube_running = False

        if not is_minikube_running:
            await util.run("minikube", "start")
            await util.retry(minikube_status)

        await super().reset()

    async def push_image(self, image):
        await util.run("./etc/kube/push-to-minikube.sh", image)

    async def update_config(self):
        ip = (await util.capture("minikube", "ip")).strip()
        await util.run("pachctl", "config", "update", "context", f"--pachd-address={ip}:30650")

class GCPDriver(BaseDriver):
    """Driver for deploying to GCP"""

    def __init__(self, project_id, cluster_name=None):
        if cluster_name is None:
            cluster_name = f"pach-{secrets.token_hex(5)}"
        self.cluster_name = cluster_name
        self.object_storage_name = f"{cluster_name}-storage"
        self.project_id = project_id

    def image(self, name):
        return f"gcr.io/{self.project_id}/{name}"

    async def reset(self):
        cluster_exists = (await util.run("gcloud", "container", "clusters", "describe", self.cluster_name,
            raise_on_error=False, capture_output=True)).rc == 0

        if cluster_exists:
            await super().reset()
            return
        
        # create the cluster from scratch if it doesn't exist already
        await util.run("gcloud", "config", "set", "container/cluster", self.cluster_name)

        await asyncio.gather(
            util.run("gcloud", "container", "clusters", "create", self.cluster_name, "--scopes=storage-rw",
                "--machine-type=n1-standard-8", "--num-nodes=2"),
            util.run("gsutil", "mb", f"gs://{self.object_storage_name}"),
        )

        account = (await util.capture("gcloud", "config", "get-value", "account")).strip()
        docker_config_path = pathlib.Path.home() / ".docker" / "config.json"

        await asyncio.gather(
            util.run("kubectl", "create", "clusterrolebinding", "cluster-admin-binding",
                "--clusterrole=cluster-admin", f"--user={account}"),
            util.run("kubectl", "create", "secret", "generic", "regcred",
                f"--from-file=.dockerconfigjson={docker_config_path}",
                "--type=kubernetes.io/dockerconfigjson"),
        )

    async def push_image(self, image):
        image_url = self.image(image)
        if ":local" in image_url:
            image_url = image_url.replace(":local", ":" + (await util.get_client_version()))
        await util.run("docker", "tag", image, image_url)
        await util.run("docker", "push", image_url)

    def deploy_args(self):
        return ["google", self.object_storage_name, "32", "--dynamic-etcd-nodes=1", "--image-pull-secret=regcred",
            f"--registry=gcr.io/{self.project_id}"]

async def main():
    parser = argparse.ArgumentParser(description="Resets a pachyderm cluster.")
    parser.add_argument("--target", default="", help="Where to deploy")
    parser.add_argument("--dash", action="store_true", help="Deploy dash")
    parser.add_argument("--ide", action="store_true", help="Deploy IDE")
    parser.add_argument("--skip-build", action="store_true", help="Skip re-build of images")
    parser.add_argument("--skip-deploy", action="store_true", help="Skip deploy")
    args = parser.parse_args()

    # sanity checks
    if "GOPATH" not in os.environ:
        raise Exception("Must set GOPATH")
    if "PACH_CA_CERTS" in os.environ:
        raise Exception("Must unset PACH_CA_CERTS\nRun:\nunset PACH_CA_CERTS")
    if args.skip_deploy and (args.dash or args.ide):
        raise Exception("Cannot set `--dash` or `--ide` with `--skip-deploy`")

    # derive which driver to use from the target
    driver = None
    if args.target == "":
        # if no target is specified, derive which driver to use from the k8s
        # context name
        kube_context = await util.capture("kubectl", "config", "current-context", raise_on_error=False)
        kube_context = kube_context.strip() if kube_context else ""
        if kube_context == "minikube":
            util.print_status("using the minikube driver")
            driver = MinikubeDriver()
        elif kube_context == "docker-desktop":
            util.print_status("using the base driver")
            driver = BaseDriver()
        if driver is None:
            # minikube won't set the k8s context if the VM isn't running. This
            # checks for the presence of the minikube executable as an
            # alternate means.
            try:
                await util.run("minikube", "version", capture_output=True)
            except:
                pass
            else:
                driver = MinikubeDriver()
        if driver is None:
            raise Exception(f"could not derive driver from context name: {kube_context}")
    elif args.target == "minikube":
        util.print_status("using the minikube driver")
        driver = MinikubeDriver()
    elif args.target == "base":
        util.print_status("using the base driver")
        driver = BaseDriver()
    elif args.target.startswith("gcp"):
        util.print_status("using the gcp driver")
        project_id = (await util.capture("gcloud", "config", "get-value", "project")).strip()
        target_parts = args.target.split(":", maxsplit=1)
        cluster_name = target_parts[1] if len(target_parts) == 2 else None
        driver = GCPDriver(project_id, cluster_name)
    else:
        raise Exception(f"unknown target: {args.target}")

    # clear environment & build everything in parallel
    procs = [util.run("make", "install"), driver.reset()]
    if not args.skip_build:
        procs.append(util.run("make", "docker-build"))
    await asyncio.gather(*procs)

    if args.skip_deploy:
        return

    # generate a manifest
    deploy_args = ["pachctl", "deploy", *driver.deploy_args(), "--dry-run", "--create-context", "--log-level=debug"]
    if not args.dash:
        deploy_args.append("--no-dashboard")
    if os.environ.get("STORAGE_V2") == "true":
        deployment_args.append("--new-storage-layer")
    deployments_str = await util.capture(*deploy_args)
    deployments_json = json.loads("[{}]".format(NEWLINE_SEPARATE_OBJECTS_PATTERN.sub("},{", deployments_str)))

    # figure out what images we want to pull
    dash_spec = util.find_in_json(deployments_json, lambda j: \
        isinstance(j, dict) and j.get("name") == "dash" and j.get("image") is not None)
    grpc_proxy_spec = util.find_in_json(deployments_json, lambda j: \
        isinstance(j, dict) and j.get("name") == "grpc-proxy")
    
    # pull the images
    pull_images = [util.run("docker", "pull", ETCD_IMAGE)]
    if dash_spec is not None:
        pull_images.append(util.run("docker", "pull", dash_spec["image"]))
    if grpc_proxy_spec is not None:
        pull_images.append(util.run("docker", "pull", grpc_proxy_spec["image"]))
    await asyncio.gather(*pull_images)

    # push the images
    push_images = [ETCD_IMAGE, "pachyderm/pachd:local", "pachyderm/worker:local"]
    if dash_spec is not None:
        push_images.append(dash_spec["image"])
    if grpc_proxy_spec is not None:
        push_images.append(grpc_proxy_spec["image"])
    await asyncio.gather(*[driver.push_image(i) for i in push_images])

    # apply the manifest
    await util.run("kubectl", "create", "-f", "-", stdin=deployments_str)

    # patch the config
    await driver.update_config()

    # wait for the cluster to stand up
    await util.retry(util.ping, attempts=60)

    # deploy the IDE
    if args.ide:
        # push the IDE images
        await asyncio.gather(*[driver.push_image(i) for i in [IDE_USER_IMAGE, IDE_HUB_IMAGE]])

        # enable enterprise
        enterprise_token = await util.capture(
            "aws", "s3", "cp",
            "s3://pachyderm-engineering/test_enterprise_activation_code.txt",
            "-"
        )
        await util.run("pachctl", "enterprise", "activate", util.RedactedString(enterprise_token))

        # activate auth
        await util.run("pachctl", "auth", "activate", stdin="admin\n")

        # run the deploy IDE command
        await util.run("pachctl", "deploy", "ide", 
            "--user-image", driver.image(IDE_USER_IMAGE),
            "--hub-image", driver.image(IDE_HUB_IMAGE),
        )

if __name__ == "__main__":
    asyncio.run(main(), debug=True)
