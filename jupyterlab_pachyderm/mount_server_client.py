import json
import subprocess
import time
import psutil

from tornado.httpclient import AsyncHTTPClient

from .pachyderm import MountInterface

MOUNT_SERVER_PORT = 9002


class MountServerClient(MountInterface):
    """Client interface for the pachctl mount-server backend."""

    def __init__(
        self,
        mount_dir: str,
    ):
        self.client = AsyncHTTPClient()
        self.mount_dir = mount_dir
        self.address = f"http://localhost:{MOUNT_SERVER_PORT}"
        self._ensure_mount_server()


    def _ensure_mount_server(self):
        """
        When we first start up, we might not have auth configured. So just try
        re-launching the mount-server on every command! If the port is already
        bound, it will exit straight away. If it's not, it might start up
        successfully with the updated config.
        """
        # TODO: add --socket and --log-file stdout args
        # TODO: add better error handling
        found = False
        for proc in psutil.process_iter(['pid', 'name', 'cmdline']):
            if proc.info['name'] == 'pachctl' and 'mount-server' in proc.info['cmdline']:
                found = True
        if not found:
            subprocess.Popen(
                [
                    "bash", "-c",
                    "set -o pipefail; "
                    +f"pachctl mount-server --mount-dir {self.mount_dir}"
                    +" &>> /tmp/pachctl-mount-server.log",
                ]
            )
            time.sleep(5)

    async def list(self):
        self._ensure_mount_server()
        response = await self.client.fetch(f"{self.address}/repos")
        # TODO: in the future we could unify the response formats from the go
        # mount-server and python and then this could just be:
        # return json.loads(response.body)?
        return {
            repo_name: {
                "branches": {
                    branch_name: {"mount": branch_info["mount"]}
                    for branch_name, branch_info in repo_info["branches"].items()
                }
            }
            for repo_name, repo_info in json.loads(response.body).items()
        }

    async def mount(self, repo, branch, mode, name):
        self._ensure_mount_server()
        await self.client.fetch(
            f"{self.address}/repos/{repo}/{branch}/_mount?name={name}&mode={mode}",
            method="PUT",
            body="{}",
        )
        return {"repo": repo, "branch": branch}

    async def unmount(self, repo, branch, name):
        self._ensure_mount_server()
        await self.client.fetch(
            f"{self.address}/repos/{repo}/{branch}/_unmount?name={name}",
            method="PUT",
            body="{}",
        )
        return {"repo": repo, "branch": branch}

    async def unmount_all(self):
        self._ensure_mount_server()
        all = await self.list()
        accum = []
        for name, repo_info in all.items():
            for branch_name, branch_info in repo_info["branches"].items():
                if branch_info["mount"]["state"] == "mounted":
                    await self.unmount(branch_info["mount"]["mount_key"]["Repo"], branch_info["mount"]["mount_key"]["Branch"], name)
                    accum.append([branch_info["mount"]["mount_key"]["Repo"], branch_info["mount"]["mount_key"]["Branch"]])
        return accum

    async def commit(self, repo, branch, name, message):
        self._ensure_mount_server()
        pass