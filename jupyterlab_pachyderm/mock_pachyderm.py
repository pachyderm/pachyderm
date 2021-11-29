from collections import namedtuple


# Need to conform to Repo object in the RPC response
Repo = namedtuple("Repo", ["repo", "branches"])
RepoName = namedtuple("RepoName", ["name"])
BranchName = namedtuple("BranchName", ["name"])


class MockPachydermClient:
    """Mocks python_pachyderm.Client"""

    def __init__(self, data=None):
        self.repos = (
            data
            if data
            else [
                Repo(
                    repo=RepoName("images"),
                    branches=[
                        BranchName("master"),
                    ],
                ),
                Repo(
                    repo=RepoName("edges"),
                    branches=[
                        BranchName("master"),
                    ],
                ),
                Repo(
                    repo=RepoName("montage"),
                    branches=[
                        BranchName("master"),
                    ],
                ),
            ]
        )

    def _create_repos(self, repos):
        for repo in repos:
            self.repos.append(
                Repo(
                    repo=RepoName(repo["repo"]),
                    branches=[BranchName(branch) for branch in repo["branches"]],
                )
            )

    def list_repo(self):
        return self.repos

    def mount(self, mount_dir, repos):
        pass

    def unmount(self, mount_dir=None, all_mounts=None):
        pass
