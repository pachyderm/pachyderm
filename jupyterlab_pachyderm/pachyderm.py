import os
from collections import defaultdict

import python_pachyderm

READ_ONLY = "ro"
READ_WRITE = "rw"


def _normalize_mode(mode):
    if mode in ["r", "ro"]:
        return READ_ONLY
    elif mode in ["w", "rw"]:
        return READ_WRITE
    else:
        raise Exception("Mode not valid")


def _parse_pfs_path(path):
    """
    a path can be one of
        - repo/branch/commit
        - repo/branch
        - repo -> defaults to master branch
    returns a 3-tuple (repo, branch, commit)
    """
    parts = path.split("/")
    if len(parts) == 3:
        return tuple(parts)
    if len(parts) == 2:
        return parts[0], parts[1], None
    if len(parts) == 1:
        return parts[0], "master", None


class PachydermClient:
    """Interface for interacting with real Pachyderm backend."""

    def __init__(
        self,
        client: python_pachyderm.Client,
        exp_client: python_pachyderm.ExperimentalClient,
    ):
        self.client = client
        self.exp_client = exp_client

    def list_repo(self):
        return self.client.list_repo()

    def mount(self, mount_dir, repos):
        if repos:
            return self.exp_client.mount(mount_dir, repos)

    def unmount(self, mount_dir):
        return self.exp_client.unmount(mount_dir=mount_dir)


class PachydermMountClient:
    """Interface for handlers to consume to call mount operations."""

    def __init__(
        self,
        client: PachydermClient,
        mount_dir: str,
    ):
        self.client = client
        self.mount_dir = mount_dir
        """
        repos schema
        {
            $repo_name: {
                "branches": {
                    $branch_name: {
                        "mount": {
                            "state": "unmounted",
                            "status": "",
                            "mode": "",
                            "mountpoint": ""
                        }
                    }
                },
                "commits": {
                    $commit_id: {},
                    ...
                }
            }
        }
        """
        self.repos = defaultdict(
            lambda: defaultdict(
                lambda: defaultdict(lambda: defaultdict(lambda: defaultdict(str)))
            )
        )

    def _prepare_repos_for_pachctl_mount(self):
        result = []
        for repo_name, repo in self.repos.items():
            for branch_name, branch in repo.get("branches").items():
                if branch["mount"]["state"] == "mounted":
                    mode = branch["mount"]["mode"]
                    if mode == READ_ONLY:
                        mode = "r"
                    elif mode == READ_WRITE:
                        mode = "w"
                    result.append(f"{repo_name}@{branch_name}+{mode}")
        return result

    def list(self):
        """Gets a list of repos and the branches per repo from Pachyderm.
        Update in memory db to reflect the most recent info.
        """
        # TODO cleanup repos that don't exist anymore
        repos = list(self.client.list_repo())
        for repo in repos:
            name = repo.repo.name
            branches = [branch.name for branch in repo.branches]
            for branch in branches:
                if branch not in self.repos[name]["branches"]:
                    self.repos[name]["branches"][branch]["mount"]["name"] = None
                    self.repos[name]["branches"][branch]["mount"]["state"] = "unmounted"
                    self.repos[name]["branches"][branch]["mount"]["status"] = None
                    self.repos[name]["branches"][branch]["mount"]["mode"] = None
                    self.repos[name]["branches"][branch]["mount"]["mountpoint"] = None
        return self.repos

    def get(self, repo):
        return self.repos.get(repo)

    def mount(self, repo, branch, mode, name=None):
        """Mounts a repo@branch+mode to /pfs/name
        Note that `name` is currently not used.
        """
        if name is None:
            name = repo
        repo_branch = self.repos.get(repo, {}).get("branches", {}).get(branch)
        if repo_branch:
            if (
                repo_branch["mount"]["state"] == "unmounted"
                or repo_branch["mount"]["mode"] != mode
            ):
                repo_branch["mount"]["name"] = name
                repo_branch["mount"]["state"] = "mounted"
                repo_branch["mount"]["mode"] = mode
                repo_branch["mount"]["mountpoint"] = os.path.join(self.mount_dir, name)

                self.client.mount(
                    self.mount_dir, self._prepare_repos_for_pachctl_mount()
                )
            return repo_branch

    def unmount(self, repo, branch, name=None):
        """Unmounts all, updates metadata, and remounts what's left"""
        repo_branch = self.repos.get(repo, {}).get("branches", {}).get(branch)
        if repo_branch:
            if repo_branch["mount"]["state"] == "mounted":
                repo_branch["mount"]["name"] = None
                repo_branch["mount"]["state"] = "unmounted"
                repo_branch["mount"]["mode"] = None
                repo_branch["mount"]["mountpoint"] = None

                self.client.unmount(self.mount_dir)
                self.client.mount(
                    self.mount_dir, self._prepare_repos_for_pachctl_mount()
                )
            return repo_branch

    def commit(self, repo, branch, name, message):
        # TODO
        pass
