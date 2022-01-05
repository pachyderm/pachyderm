import json
import subprocess

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
        # TODO: add --socket and --log-file stdout args
        subprocess.Popen(
            [
                "pachctl",
                "mount-server",
                "--daemonize",
                "--mount-dir",
                self.mount_dir,
            ]
        )

    async def list(self):
        response = await self.client.fetch(f"{self.address}/repos")
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
        await self.client.fetch(
            f"{self.address}/repos/{repo}/{branch}/_mount?name={name}&mode={mode}",
            method="PUT",
            body="{}",
        )
        return {"repo": repo, "branch": branch}

    async def unmount(self, repo, branch, name):
        await self.client.fetch(
            f"{self.address}/repos/{repo}/{branch}/_unmount?name={name}",
            method="PUT",
            body="{}",
        )
        return {"repo": repo, "branch": branch}

    async def unmount_all(self):
        pass

    async def commit(self, repo, branch, name, message):
        pass
