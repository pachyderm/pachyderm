import re

from . import Branch, Commit, File, Project, Repo

branch_re = re.compile(r"^[a-zA-Z0-9_-]+$")
uuid_re = re.compile(r"[0-9a-f]{12}4[0-9a-f]{19}")


def _Repo_from_uri(uri: str) -> Repo:
    """
    Parses the following format:
        [project/]<repo>

    If no project is specified it defaults to "default".
    """
    if "/" in uri:
        project, repo = uri.split("/")
    else:
        project, repo = "default", uri
    return Repo(name=repo, type="user", project=Project(name=project))


def _Repo_as_uri(self: "Repo") -> str:
    project = self.project.name or "default"
    return f"{project}/{self.name}"


Repo.from_uri = _Repo_from_uri
Repo.as_uri = _Repo_as_uri


def _Commit_from_uri(uri: str) -> Commit:
    """
    Parses the following format:
        [project/]<repo>@<branch-or-commit>
    where @<branch-or-commit> can take the form:
        @branch
        @branch=commit
        @commit
    Additionally @<branch-or-commit> can be augmented with caret notation:
        @branch^2

    All unspecified components will default to None, except for an unspecified
      project which defaults to "default".
    """
    project_repo, branch_or_commit = uri.split("@")
    if "=" in branch_or_commit:
        branch, commit = branch_or_commit.split("=")
    elif uuid_re.match(branch_or_commit) or not branch_re.match(branch_or_commit):
        branch, commit = None, branch_or_commit
    else:
        branch, commit = branch_or_commit, None
    return Commit(
            branch=Branch(
                name=branch,
                repo=Repo.from_uri(project_repo)
            ),
            id=commit
        )


def _Commit_as_uri(self: "Commit") -> str:
    project_repo = self.branch.repo.as_uri()
    if self.branch.name and self.id:
        return f"{project_repo}@{self.branch.name}={self.id}"
    elif self.branch:
        return f"{project_repo}@{self.branch.name}"
    else:
        return f"{project_repo}@{self.id}"


Commit.from_uri = _Commit_from_uri
Commit.as_uri = _Commit_as_uri


def _File_from_uri(uri: str) -> File:
    """
    Parses the following format:
        [project/]<repo>@<branch-or-commit>[:<path/in/pfs>]
    where @<branch-or-commit> can take the form:
        @branch
        @branch=commit
        @commit
    Additionally @<branch-or-commit> can be augmented with caret notation:
        @branch^2

    All unspecified components will default to None, except for an unspecified
      project which defaults to "default".
    """
    project_repo_branch, path = uri.split(":")

    return File(
        commit=Commit.from_uri(project_repo_branch),
        path=path or None,
    )


def _File_as_uri(self: "File") -> str:
    return f"{self.commit.as_uri()}:{self.path}"


File.from_uri = _File_from_uri
File.as_uri = _File_as_uri
