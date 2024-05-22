import betterproto
import pytest

from pachyderm_sdk.api.pfs import (
    Branch,
    BranchPicker,
    Commit,
    CommitPicker,
    File,
    Project,
    Repo,
    RepoPicker,
)


# fmt: off
@pytest.mark.parametrize(
    "message, uri",
    [
        (Repo(name="images", project=None), "default/images"),
        (Repo(name="data", project=Project()), "default/data"),
        (Repo(name="results", project=Project(name="prod")), "prod/results")
    ]
)
# fmt: on
def test_repo(message: Repo, uri: str):
    """Test the Repo.from_uri and Repo.as_uri methods."""
    picker = message.as_picker()
    assert message.as_uri() == str(message) == uri
    assert picker.as_uri() == str(picker) == uri
    assert Repo.from_uri(uri).as_uri() == uri
    assert RepoPicker.from_uri(uri).as_uri() == uri


# fmt: off
@pytest.mark.parametrize(
    "message, uri",
    [
        (Branch(name="master", repo=Repo(name="images", project=None)), "default/images@master"),
        (Branch(name="develop", repo=Repo(name="data", project=Project())), "default/data@develop"),
        (Branch(name="explore", repo=Repo(name="results", project=Project(name="prod"))), "prod/results@explore")
    ]
)
# fmt: on
def test_branch(message: Branch, uri: str):
    picker = message.as_picker()
    assert message.as_uri() == str(message) == uri
    assert picker.as_uri() == str(picker) == uri
    assert Branch.from_uri(uri).as_uri() == uri
    assert BranchPicker.from_uri(uri).as_uri() == uri


@pytest.mark.parametrize(
    "bad_uri", ["project/repo", "project/repo@bad=branch", "project/repo@bad.branch"]
)
def test_branch_error(bad_uri: str):
    with pytest.raises(ValueError):
        Branch.from_uri(bad_uri)


# fmt: off
@pytest.mark.parametrize(
    "message, uri, picker_field",
    [(Commit(branch=Branch(name="master", repo=Repo(name="images", project=None))),
      "default/images@master",
      "branch_head"
      ),
     (Commit(id="abcdef1234564abcdef123456789abcd", branch=Branch(repo=Repo(name="data", project=Project()))),
      "default/data@abcdef1234564abcdef123456789abcd",
      "id"
      ),
     (Commit(id="dev.6", branch=Branch(name=None, repo=Repo(name="images", project=None))),
      "default/images@dev.6",
      "branch_root"
      ),
     (Commit(id="123456789abc4def1234567890abcdef^2", branch=Branch(name=None, repo=Repo(name="results", project=Project(name="prod")))),
      "prod/results@123456789abc4def1234567890abcdef^2",
      "ancestor"
      ),
     ]
)
# fmt: on
def test_commit(message: Commit, uri: str, picker_field: str):
    picker = message.as_picker()
    assert message.as_uri() == str(message) == uri
    assert picker.as_uri() == str(picker) == uri
    assert betterproto.which_one_of(picker, "picker")[0] == picker_field
    assert Commit.from_uri(uri).as_uri() == uri
    assert CommitPicker.from_uri(uri).as_uri() == uri


def test_commit_error():
    bad_uri = "project/repo"
    with pytest.raises(ValueError):
        Commit.from_uri(bad_uri)


def test_commit_picker_error():
    with pytest.raises(ValueError):
        CommitPicker.from_uri("project/repo")
    with pytest.raises(ValueError):
        CommitPicker.from_uri("project/repo@branch.abcd")
    with pytest.raises(ValueError):
        CommitPicker.from_uri("project/repo@branch^abcd")


def test_commit_picker_uri_error():
    with pytest.raises(ValueError):
        CommitPicker().as_uri()


# fmt: off
@pytest.mark.parametrize(
    "message, uri",
    [
        (File(
            commit=Commit(branch=Branch(name="master", repo=Repo(name="images", project=None))),
            path="/"
        ), "default/images@master:/"
        ),
        (File(
            commit=Commit(
                id="abcdef1234564abcdef123456789abcd",
                branch=Branch(repo=Repo(name="data", project=Project()))
            ), path="/index.html"
        ), "default/data@abcdef1234564abcdef123456789abcd:/index.html"),
        (File(
            commit=Commit(
                id="123456789abc4def1234567890abcdef",
                branch=Branch(name="explore", repo=Repo(name="results", project=Project(name="prod")))
            ), path="/dir/binary.dat"
            ), "prod/results@explore=123456789abc4def1234567890abcdef:/dir/binary.dat")
    ]
)
# fmt: on
def test_file(message: File, uri: str):
    assert message.as_uri() == str(message) == uri
    assert File.from_uri(uri).as_uri() == uri
