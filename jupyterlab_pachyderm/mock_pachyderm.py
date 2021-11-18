from .pachyderm import PachydermMountClient


class MockPachydermMountClient(PachydermMountClient):
    def __init__(self, data=None):
        self.repos = (
            data
            if data
            else {
                "images": {
                    "master": {"state": "unmounted", "name": None, "mode": None},
                    "dev": {"state": "unmounted", "name": None, "mode": None},
                },
                "edges": {
                    "master": {"state": "unmounted", "name": None, "mode": None},
                },
            }
        )

    def _find_repo(self, repo, branch):
        return self.repos.get(repo, {}).get(branch, None)

    def list(self):
        return self.repos

    def get(self, repo):
        return self.repos.get(repo, {})

    def mount(self, repo, branch, mode, name):
        mount_info = self._find_repo(repo, branch)
        if mount_info and mount_info["state"] == "unmounted":
            mount_info["state"] = "mounted"
            mount_info["name"] = name
            mount_info["mode"] = mode
            return {repo: {branch: mount_info}}

    def unmount(self, repo, branch, name):
        mount_info = self._find_repo(repo, branch)
        if (
            mount_info
            and mount_info["state"] == "mounted"
            and mount_info["name"] == name
        ):
            mount_info["state"] = "unmounted"
            mount_info["mode"] = None
            mount_info["name"] = None
            return {repo: {branch: mount_info}}
