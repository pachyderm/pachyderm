""" This file patches methods onto the PFS Noun objects.

This is done, as opposed to subclassing them, for these methods
  to automatically be accessible on any objects included in the
  response from the API.

Note: These are internally patched and this file should
  not be imported directly by users.
"""

import re

from . import (
    Branch,
    BranchPicker,
    BranchPickerBranchName,
    Commit,
    CommitPicker,
    CommitPickerAncestorOf,
    CommitPickerBranchRoot,
    CommitPickerCommitByGlobalId,
    File,
    Project,
    ProjectPicker,
    Repo,
    RepoPicker,
    RepoPickerRepoName,
)

branch_re = re.compile(r"^[a-zA-Z\d_-]+$")
uuid_re = re.compile(r"^[\da-f]{12}4[\da-f]{19}$")


def _Project_as_picker(self: "Project") -> ProjectPicker:
    """Converts a Project to a ProjectPicker."""
    return ProjectPicker(name=self.name)


Project.as_picker = _Project_as_picker


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
    if not self.name:
        raise ValueError("Empty repo name")
    project = "default"
    if self.project and self.project.name:
        project = self.project.name
    return f"{project}/{self.name}"


def _Repo___post_init__(self: "Repo") -> None:
    if not self.project or not self.project.name:
        self.project = Project(name="default")
    if not self.type:
        self.type = "user"
    super(self.__class__, self).__post_init__()


def _Repo_as_picker(self: "Repo") -> RepoPicker:
    """Converts a Repo to a RepoPicker"""
    project = None if self.project is None else self.project.as_picker()
    return RepoPicker(
        name=RepoPickerRepoName(name=self.name, type=self.type, project=project)
    )


Repo.from_uri = _Repo_from_uri
Repo.as_uri = Repo.__str__ = _Repo_as_uri
Repo.__post_init__ = _Repo___post_init__
Repo.as_picker = _Repo_as_picker


def _RepoPicker_from_uri(uri: str) -> RepoPicker:
    """
    Parses the following format:
        [project/]repo

    If no project is specified it defaults to "default".
    """
    return Repo.from_uri(uri).as_picker()


def _RepoPicker_as_uri(self: RepoPicker) -> str:
    """
    Returns the URI for the RepoPicker object in the following format:
        project/repo

    If no project is specified it defaults to "default"
    """
    project = "default"
    if self.name.project and self.name.project.name:
        project = self.name.project.name
    return f"{project}/{self.name.name}"


def _RepoPickerRepoName___post_init__(self: "RepoPickerRepoName") -> None:
    if not self.project or not self.project.name:
        self.project = ProjectPicker(name="default")
    if not self.type:
        self.type = "user"
    super(self.__class__, self).__post_init__()


RepoPicker.from_uri = _RepoPicker_from_uri
RepoPicker.as_uri = RepoPicker.__str__ = _RepoPicker_as_uri
RepoPickerRepoName.__post_init__ = _RepoPickerRepoName___post_init__


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


def _Branch_as_picker(self: "Branch") -> BranchPicker:
    """Converts a Branch to a BranchPicker."""
    repo = self.repo.as_picker()
    return BranchPicker(name=BranchPickerBranchName(name=self.name, repo=repo))


Branch.from_uri = _Branch_from_uri
Branch.as_uri = Branch.__str__ = _Branch_as_uri
Branch.as_picker = _Branch_as_picker


def _BranchPicker_from_uri(uri: str) -> BranchPicker:
    """
    Parses the following format:
        [project/]repo@branch

    If no project is specified it defaults to "default".

    Raises:
        ValueError: If no branch is specified.
    """
    return Branch.from_uri(uri).as_picker()


def _BranchPicker_as_uri(self: BranchPicker) -> str:
    """Returns the URI for the BranchPicker object in the following format:
      project/repo@branch

    If no project is specified it defaults to "default"
    """
    return f"{self.name.repo.as_uri()}@{self.name.name}"


BranchPicker.from_uri = _BranchPicker_from_uri
BranchPicker.as_uri = BranchPicker.__str__ = _BranchPicker_as_uri


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
    if self.branch:
        project_repo = self.branch.repo.as_uri()
        if self.branch.name and self.id:
            return f"{project_repo}@{self.branch.name}={self.id}"
        elif self.branch.name:
            return f"{project_repo}@{self.branch.name}"
        else:
            return f"{project_repo}@{self.id}"

    project_repo = self.repo.as_uri()
    return f"{project_repo}@{self.id}"


def _Commit_as_picker(self: "Commit") -> CommitPicker:
    """Converts a Commit to a CommitPicker."""
    # Note: Due to the fact the ancestor/branch offset notation was
    #   encoded within the Commit.id field, a more direct conversion
    #   from Commit -> CommitPicker would already require a lot of
    #   str parsing. Therefore, the choice was made to implement all
    #   of this within the CommitPicker.from_uri method.
    return CommitPicker.from_uri(self.as_uri())


Commit.from_uri = _Commit_from_uri
Commit.as_uri = Commit.__str__ = _Commit_as_uri
Commit.as_picker = _Commit_as_picker


def _CommitPicker_from_uri(uri: str) -> CommitPicker:
    """
    Parses the following format:
        [project/]repo@branch-or-commit
    where @branch-or-commit can take the form:
        @branch
        @branch=commit
        @commit
    Additionally @branch-or-commit can be augmented in the following ways:
        @branch.2
        @commit^3

    All unspecified components will default to None, except for an unspecified
      project which defaults to "default".
    """

    def parse_offset(value: str) -> int:
        try:
            return int(value)
        except ValueError:
            raise ValueError(
                "Could not parse commit uri. "
                "Branch name is invalid - expected a integer offset. "
                f"uri: {uri}"
            )

    if "@" not in uri:
        raise ValueError(
            "Could not parse branch/commit. URI must have the form: "
            "[project/]<repo>@(branch|branch=commit|commit)"
        )
    project_repo, branch_or_commit = uri.split("@", 1)
    repo = RepoPicker.from_uri(project_repo)
    if "=" in branch_or_commit:
        _, commit = branch_or_commit.split("=", 1)
        return CommitPicker(id=CommitPickerCommitByGlobalId(repo=repo, id=commit))
    elif uuid_re.match(branch_or_commit) or not branch_re.match(branch_or_commit):
        if "." in branch_or_commit:
            branch_name, offset_str = branch_or_commit.split(".")
            return CommitPicker(
                branch_root=CommitPickerBranchRoot(
                    branch=BranchPicker(
                        name=BranchPickerBranchName(name=branch_name, repo=repo)
                    ),
                    offset=parse_offset(offset_str),
                )
            )
        if "^" in branch_or_commit:
            start_uri, offset_str = uri.split("^")
            return CommitPicker(
                ancestor=CommitPickerAncestorOf(
                    offset=parse_offset(offset_str),
                    start=_CommitPicker_from_uri(start_uri),
                )
            )
        return CommitPicker(id=CommitPickerCommitByGlobalId(repo=repo, id=branch_or_commit))
    else:
        return CommitPicker(
            branch_head=BranchPicker(
                name=BranchPickerBranchName(name=branch_or_commit, repo=repo)
            ),
        )


def _CommitPicker_as_uri(self: CommitPicker) -> str:
    """Returns the URI for the CommitPicker object.
    If no project is specified it defaults to "default"
    """
    if self.branch_head:
        return self.branch_head.as_uri()
    if self.id:
        return f"{self.id.repo}@{self.id.id}"
    if self.ancestor:
        return f"{self.ancestor.start}^{self.ancestor.offset}"
    if self.branch_root:
        return f"{self.branch_root.branch}.{self.branch_root.offset}"
    raise ValueError(f"invalid CommitPicker: {repr(self)}")


CommitPicker.from_uri = _CommitPicker_from_uri
CommitPicker.as_uri = CommitPicker.__str__ = _CommitPicker_as_uri


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
