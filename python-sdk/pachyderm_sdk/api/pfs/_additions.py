""" This file patches methods onto the PFS Noun objects.

This is done, as opposed to subclassing them, for these methods
  to automatically be accessible on any objects included in the
  response from the API.

Note: These are internally patched and this file should
  not be imported directly by users.
"""
import re

from . import Branch, Commit, File, Project, Repo

branch_re = re.compile(r"^[a-zA-Z\d_-]+$")
uuid_re = re.compile(r"^[\da-f]{12}4[\da-f]{19}$")


def _Repo_from_uri(uri: str) -> Repo:
    """
    Parses the following format:
        [project/]repo

    If no project is specified it defaults to "default".
    """
    if "/" in uri:
        project, repo = uri.split("/", 1)
    else:
        project, repo = "default", uri
    return Repo(name=repo, type="user", project=Project(name=project))


def _Repo_as_uri(self: "Repo") -> str:
    """Returns the URI for the Repo object in the following format:
      project/repo

    If no project is specified it defaults to "default"
    """
    project = "default"
    if self.project and self.project.name:
        project = self.project.name
    return f"{project}/{self.name}"


def _Repo___post_init__(self: "Repo") -> None:
    if not self.project or not self.project.name:
        self.project = Project(name="default")
    super(self.__class__, self).__post_init__()


Repo.from_uri = _Repo_from_uri
Repo.as_uri = Repo.__str__ = _Repo_as_uri
Repo.__post_init__ = _Repo___post_init__


def _Branch_from_uri(uri: str) -> Branch:
    """
    Parses the following format:
        [project/]repo@branch

    If no project is specified it defaults to "default".

    Raises:
        ValueError: If no branch is specified.
    """
    if "@" not in uri:
        raise ValueError(
            "Could not parse branch/commit. URI must have the form: "
            "[project/]<repo>@branch"
        )
    project_repo, branch = uri.split("@", 1)
    if not branch_re.match(branch):
        raise ValueError(f"Invalid branch name: {branch}")
    return Branch(name=branch, repo=Repo.from_uri(project_repo))


def _Branch_as_uri(self: "Branch") -> str:
    """Returns the URI for the Branch object in the following format:
      project/repo@branch

    If no project is specified it defaults to "default"
    """
    return f"{self.repo.as_uri()}@{self.name}"


Branch.from_uri = _Branch_from_uri
Branch.as_uri = Branch.__str__ = _Branch_as_uri


def _Commit_from_uri(uri: str) -> Commit:
    """
    Parses the following format:
        [project/]repo@branch-or-commit
    where @branch-or-commit can take the form:
        @branch
        @branch=commit
        @commit
    Additionally @branch-or-commit can be augmented with caret notation:
        @branch^2

    All unspecified components will default to None, except for an unspecified
      project which defaults to "default".
    """
    if "@" not in uri:
        raise ValueError(
            "Could not parse branch/commit. URI must have the form: "
            "[project/]<repo>@(branch|branch=commit|commit)"
        )
    project_repo, branch_or_commit = uri.split("@", 1)
    repo = Repo.from_uri(project_repo)
    if "=" in branch_or_commit:
        branch, commit = branch_or_commit.split("=", 1)
    elif uuid_re.match(branch_or_commit) or not branch_re.match(branch_or_commit):
        branch, commit = None, branch_or_commit
    else:
        branch, commit = branch_or_commit, None
    # TODO: Commits are no longer pinned to branches.
    #       When officially deprecated, the `branch` field will be removed.
    return Commit(
        branch=Branch(name=branch, repo=repo),
        id=commit,
        repo=repo,
    )


def _Commit_as_uri(self: "Commit") -> str:
    """Returns the URI for the Commit object in one of the following formats:
        project/repo@branch
        project/repo@branch=commit
        project/repo@commit

    If no project is specified it defaults to "default"
    """
    project_repo = self.branch.repo.as_uri()
    if self.branch.name and self.id:
        return f"{project_repo}@{self.branch.name}={self.id}"
    elif self.branch.name:
        return f"{project_repo}@{self.branch.name}"
    else:
        return f"{project_repo}@{self.id}"


Commit.from_uri = _Commit_from_uri
Commit.as_uri = Commit.__str__ = _Commit_as_uri


def _File_from_uri(uri: str) -> File:
    """
    Parses the following format:
        [project/]repo@branch-or-commit[:path/in/pfs]
    where @branch-or-commit can take the form:
        @branch
        @branch=commit
        @commit
    Additionally @branch-or-commit can be augmented with caret notation:
        @branch^2

    All unspecified components will default to None, except for an unspecified
      project which defaults to "default".
    """
    if ":" in uri:
        project_repo_branch, path = uri.split(":", 1)
    else:
        project_repo_branch, path = uri, None

    return File(
        commit=Commit.from_uri(project_repo_branch),
        path=path,
    )


def _File_as_uri(self: "File") -> str:
    """Returns the URI for the File object in one of the following formats:
        project/repo@branch:/path
        project/repo@branch=commit:/path
        project/repo@commit:/path

    If no project is specified it defaults to "default"
    """
    return f"{self.commit.as_uri()}:{self.path}"


File.from_uri = _File_from_uri
File.as_uri = File.__str__ = _File_as_uri
