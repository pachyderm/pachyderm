import os
from typing import Callable

import python_pachyderm
from .log import get_logger

READ_ONLY = "ro"
READ_WRITE = "rw"


class MountInterface:
    async def list(self):
        pass

    async def mount(self, repo, branch, mode, name):
        pass

    async def unmount(self, repo, branch, name):
        pass

    async def unmount_all(self):
        pass

    async def commit(self, repo, branch, name, message):
        pass

    async def config(self, body):
        pass

    async def auth_login(self):
        pass

    async def auth_logout(self):
        pass

    async def health(self):
        pass


class PythonPachydermClient:
    """Interface for interacting with real Pachyderm backend."""

    def __init__(
        self,
        client: python_pachyderm.Client,
        exp_client: python_pachyderm.experimental.Client,
    ):
        self.client = client
        self.exp_client = exp_client

    def list_repo(self):
        return self.client.list_repo()

    def mount(self, mount_dir, repos):
        if repos:
            get_logger().info(f"calling pachctl mount on {mount_dir} {repos}")
            return self.exp_client.mount(mount_dir, repos)

    def unmount(self, mount_dir=None, all_mounts=None):
        return self.exp_client.unmount(mount_dir=mount_dir, all_mounts=all_mounts)


class PythonPachydermMountClient(MountInterface):
    """Interface for handlers to consume to call mount operations."""

    def __init__(
        self,
        create_client: Callable[[], PythonPachydermClient],
        mount_dir: str,
    ):
        self.create_client = create_client
        self.mount_dir = mount_dir
        """[mount_states schema]
        {
            (repo, branch/commit): {
                "state": "unmounted",
                "status": "",
                "mode": "",
                "mountpoint": ""
            }
        }
        """
        self.mount_states = {}

    @property
    def client(self):
        return self.create_client()

    def _mount_string(self, repo, branch, mode):
        if mode == READ_ONLY:
            mode = "r"
        elif mode == READ_WRITE:
            mode = "w"
        return f"{repo}@{branch}+{mode}"

    def _default_mount_state(self):
        return {
            "name": None,
            "state": "unmounted",
            "status": None,
            "mode": None,
            "mountpoint": None,
        }

    def _current_mount_strings(self):
        result = []
        for (repo, branch), mount_state in self.mount_states.items():
            if mount_state["state"] == "mounted":
                result.append(self._mount_string(repo, branch, mount_state["mode"]))
        return result

    async def list(self):
        """Get a list of available repos and their branches,
        and overlay with their mount states.
        Returns
        {
            "repo": {
                "branches": {
                    "branch": {
                        "mount": {
                            "name": name,
                            "state": state,
                            "status": status,
                            "mode": mode,
                            "mountpoint": mountpoint
                        }
                    }
                }
            }
        }
        """
        result = {}
        repos = list(self.client.list_repo())
        for repo in repos:
            repo_name = repo.repo.name
            result[repo_name] = {"branches": {}}
            for branch_name in [branch.name for branch in repo.branches]:
                result[repo_name]["branches"][branch_name] = {
                    "mount": self.mount_states.get(
                        (repo_name, branch_name), self._default_mount_state()
                    ),
                }
        return result

    async def mount(self, repo, branch, mode, name=None):
        """Mounts a repo@branch+mode to /pfs/name
        Note that `name` is currently not used.
        Does nothing if repo is already mounted.
        """
        if name is None:
            name = repo
        mount_key = (repo, branch)
        mount_state = self.mount_states.get(mount_key, self._default_mount_state())
        if mount_state.get("state") == "unmounted":
            # unmounts all and remounts existing mounted repos + new repo
            mount_strings = self._current_mount_strings()
            mount_strings.append(self._mount_string(repo, branch, mode))

            self.client.mount(self.mount_dir, mount_strings)
            # update mount_state after mount is successful
            mount_state["state"] = "mounted"
            mount_state["name"] = name
            mount_state["mode"] = mode
            mount_state["mountpoint"] = os.path.join(self.mount_dir, name)
            self.mount_states[mount_key] = mount_state

            return {
                "repo": repo,
                "branch": branch,
                "mount": self.mount_states.get(mount_key),
            }

    async def unmount(self, repo, branch, name=None):
        """Unmounts all, update mount_state, and remounts what's left"""
        mount_key = (repo, branch)
        if self.mount_states.get(mount_key, {}).get("state") == "mounted":
            # need to explicitly unmount for the case when there is only one repo left to unmount
            # _prepare_repos_for_pachctl_mount would return empty, and no mount call would be made
            self.client.unmount(self.mount_dir)
            del self.mount_states[mount_key]
            self.client.mount(self.mount_dir, self._current_mount_strings())
            return {"repo": repo, "branch": branch, "mount": {"state": "unmounted"}}

    async def unmount_all(self):
        result = []
        for (repo, branch) in self.mount_states:
            result.append([repo, branch])

        self.client.unmount(all_mounts=True)
        self.mount_states = {}
        return result

    def commit(self, repo, branch, name, message):
        # TODO
        pass

    async def config(self, body=None):
        pass

    async def auth_login(self):
        pass

    async def auth_logout(self):
        pass

    async def health(self):
        pass
