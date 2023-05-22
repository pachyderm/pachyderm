import pytest

from pachyderm_sdk.api.pfs import Branch, Commit, File, Project, Repo


@pytest.mark.parametrize(
    "message, uri",
    [
        (Repo(name="images", project=None), "default/images"),
        (Repo(name="data", project=Project()), "default/data"),
        (Repo(name="results", project=Project(name="prod")), "prod/results")
    ]
)
def test_repo(message: Repo, uri: str):
    """Test the Repo.from_uri and Repo.as_uri methods."""
    assert message.as_uri() == uri
    assert Repo.from_uri(uri).as_uri() == uri


@pytest.mark.parametrize(
    "message, uri",
    [
        (
            Branch(name="master", repo=Repo(name="images", project=None)),
            "default/images@master"
        ),
        (
            Branch(name="develop", repo=Repo(name="data", project=Project())),
            "default/data@develop"
        ),
        (
            Branch(
                name="explore",
                repo=Repo(name="results", project=Project(name="prod"))
            ),
            "prod/results@explore"
        )
    ]
)
def test_branch(message: Branch, uri: str):
    assert message.as_uri() == uri
    assert Branch.from_uri(uri).as_uri() == uri


@pytest.mark.parametrize(
    "bad_uri",
    ["project/repo", "project/repo@bad=branch", "project/repo@bad.branch"]
)
def test_branch_error(bad_uri: str):
    with pytest.raises(ValueError):
        Branch.from_uri(bad_uri)


@pytest.mark.parametrize(
    "message, uri",
    [
        (
            Commit(branch=Branch(name="master", repo=Repo(name="images", project=None))),
            "default/images@master"
        ),
        (
            Commit(
                id="abcdef1234564abcdef123456789abcd",
                branch=Branch(repo=Repo(name="data", project=Project()))
            ),
            "default/data@abcdef1234564abcdef123456789abcd"
        ),
        (
            Commit(
                id="123456789abc4def1234567890abcdef",
                branch=Branch(
                    name="explore",
                    repo=Repo(name="results", project=Project(name="prod"))
                )
            ),
            "prod/results@explore=123456789abc4def1234567890abcdef"
        )
    ]
)
def test_commit(message: Commit, uri: str):
    assert message.as_uri() == uri
    assert Commit.from_uri(uri).as_uri() == uri


def test_commit_error():
    bad_uri = "project/repo"
    with pytest.raises(ValueError):
        Commit.from_uri(bad_uri)


@pytest.mark.parametrize(
    "message, uri",
    [
        (
            File(
                commit=Commit(
                    branch=Branch(name="master", repo=Repo(name="images", project=None))
                ),
                path="/"
            ),
            "default/images@master:/"
        ),
        (
            File(
                commit=Commit(
                    id="abcdef1234564abcdef123456789abcd",
                    branch=Branch(repo=Repo(name="data", project=Project()))
                ),
                path="/index.html"
            ),
            "default/data@abcdef1234564abcdef123456789abcd:/index.html"
        ),
        (
            File(
                commit=Commit(
                    id="123456789abc4def1234567890abcdef",
                    branch=Branch(
                        name="explore",
                        repo=Repo(name="results", project=Project(name="prod"))
                    )
                ),
                path="/dir/binary.dat"
            ),
            "prod/results@explore=123456789abc4def1234567890abcdef:/dir/binary.dat"
        )
    ]
)
def test_file(message: File, uri: str):
    assert message.as_uri() == uri
    assert File.from_uri(uri).as_uri() == uri
