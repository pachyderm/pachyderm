import re

from . import Branch, Commit, File, Repo

branch_re = re.compile(r"^[a-zA-Z0-9_-]+$")
uuid_re = re.compile(r"[0-9a-f]{12}4[0-9a-f]{19}")


def _File_parse(path: str) -> File:
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
    project_repo_branch, path = path.split(":")
    project_repo, branch_or_commit = project_repo_branch.split("@")
    if "=" in branch_or_commit:
        branch, commit = branch_or_commit.split("=")
    elif uuid_re.match(branch_or_commit) or not branch_re.match(branch_or_commit):
        branch, commit = None, branch_or_commit
    else:
        branch, commit = branch_or_commit, None
    if "/" in project_repo:
        project, repo = project_repo.split("/")
    else:
        project, repo = "default", project_repo

    return File(
        commit=Commit(
            branch=Branch(
                name=branch,
                repo=Repo(
                    name=repo,
                    type="user"
                )
            ),
            id=commit
        ),
        path=path or None,
    )


File.parse = _File_parse
