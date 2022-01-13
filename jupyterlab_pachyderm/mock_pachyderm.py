from collections import namedtuple

# Need to conform to Repo object in the RPC response
Repo = namedtuple("Repo", ["repo", "branches"])
RepoName = namedtuple("RepoName", ["name"])
BranchName = namedtuple("BranchName", ["name"])

repos = [
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


def reset_repos():
    global repos
    repos = [
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


def add_repos(new_repos):
    for repo in new_repos:
        repos.append(
            Repo(
                repo=RepoName(repo["repo"]),
                branches=[BranchName(branch) for branch in repo["branches"]],
            )
        )


def delete_repo(name):
    found = False
    for repo in repos:
        if repo.repo.name == name:
            found = True
            break
    if found:
        repos.remove(repo)


class MockPachydermClient:
    """Mocks python_pachyderm.Client"""

    def _repo_exists(self, name):
        for repo in repos:
            if repo.repo.name == name:
                return True
        return False

    def list_repo(self):
        return repos

    def mount(self, _, repo_mount_strings):
        for repo in repo_mount_strings:
            repo, _ = repo.split("@")
            if not self._repo_exists(repo):
                raise ValueError(f"{repo} does not exist")

    def unmount(self, mount_dir=None, all_mounts=None):
        pass
