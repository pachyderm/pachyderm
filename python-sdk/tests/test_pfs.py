import io
from pathlib import Path
from typing import NamedTuple

import betterproto
import grpc

from tests.fixtures import *
from tests.utils import count

from pachyderm_sdk.api import pfs
from pachyderm_sdk.api.pfs.file import PFSFile
from pachyderm_sdk.api.pfs.extension import ClosedCommit, OpenCommit
from pachyderm_sdk.constants import MAX_RECEIVE_MESSAGE_SIZE


class TestUnitProject:
    """Unit tests for the project management API"""

    @staticmethod
    def test_list_project(client: TestClient):
        project = client.new_project()
        projects = list(client.pfs.list_project())
        assert len(projects) >= 1
        assert project.name in {resp.project.name for resp in projects}

    @staticmethod
    def test_inspect_project(client: TestClient):
        project = client.new_project()
        resp = client.pfs.inspect_project(project=project)
        assert resp.project.name == project.name

    @staticmethod
    def test_delete_project(client: TestClient):
        project = client.new_project()
        assert client.pfs.project_exists(project)
        client.pfs.delete_project(project=project)
        assert not client.pfs.project_exists(project)


class TestsUnitRepo:
    """Unit tests for the Repo management API."""

    @staticmethod
    def test_list_repo(client: TestClient, default_project: bool):
        client.new_repo(default_project)
        all_repos = list(client.pfs.list_repo())
        user_repos = list(client.pfs.list_repo(type="user"))
        assert len(all_repos) >= 1
        assert len(user_repos) >= 1

    @staticmethod
    def test_inspect_repo(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project)
        response = client.pfs.inspect_repo(repo=repo)
        assert response.repo.name == repo.name

    @staticmethod
    def test_delete_repo(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project)
        assert client.pfs.repo_exists(repo)
        client.pfs.delete_repo(repo=repo)
        assert not client.pfs.repo_exists(repo)

    @staticmethod
    def test_delete_repos(client: TestClient):
        repo = client.new_repo(default_project=False)
        project = repo.project
        assert count(client.pfs.list_repo(projects=[project])) >= 1
        client.pfs.delete_repos(projects=[project])
        assert count(client.pfs.list_repo(projects=[project])) == 0
        assert not client.pfs.repo_exists(repo)


class TestUnitBranch:
    """Unit tests for the branch management API."""

    @staticmethod
    def test_list_branch(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project)
        master_branch = pfs.Branch(repo=repo, name="master")
        develop_branch = pfs.Branch(repo=repo, name="develop")

        client.pfs.create_branch(branch=master_branch)
        client.pfs.create_branch(branch=develop_branch)

        branch_info = list(client.pfs.list_branch(repo=repo))
        assert len(branch_info) == 2
        assert client.pfs.branch_exists(master_branch)
        assert client.pfs.branch_exists(develop_branch)

    @staticmethod
    def test_delete_branch(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project)
        branch = pfs.Branch(repo=repo, name="develop")

        client.pfs.create_branch(branch=branch)

        assert client.pfs.branch_exists(branch)
        client.pfs.delete_branch(branch=branch)
        assert not client.pfs.branch_exists(branch)

    @staticmethod
    def test_inspect_branch(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project)
        branch = pfs.Branch(repo=repo, name="develop")

        client.pfs.create_branch(branch=branch)
        branch_info = client.pfs.inspect_branch(branch=branch)

        assert branch_info.branch.name == branch.name


class TestUnitCommit:
    """Unit tests for the commit management API."""

    @staticmethod
    def test_start_commit(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project)
        branch = pfs.Branch(repo=repo, name="master")
        commit = client.pfs.start_commit(branch=branch)
        assert commit.branch.repo.name == repo.name

        # cannot start new commit before the previous one is finished
        with pytest.raises(grpc.RpcError, match=r"parent commit .* has not been finished"):
            client.pfs.start_commit(branch=branch)

        client.pfs.finish_commit(commit=commit)

        with pytest.raises(grpc.RpcError, match=r"repo .* not found"):
            client.pfs.start_commit(branch=pfs.Branch.from_uri("fake-repo@fake-branch"))

    @staticmethod
    def test_start_with_parent(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project)
        branch = pfs.Branch(repo=repo, name="master")

        commit1 = client.pfs.start_commit(branch=branch)
        client.pfs.finish_commit(commit=commit1)

        commit2 = client.pfs.start_commit(parent=commit1, branch=branch)
        assert commit2.branch.repo.name == repo.name

    @staticmethod
    def test_start_commit_fork(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project)
        master_branch = pfs.Branch(repo=repo, name="master")
        patch_branch = pfs.Branch(repo=repo, name="patch")

        commit1 = client.pfs.start_commit(branch=master_branch)
        client.pfs.finish_commit(commit=commit1)

        commit2 = client.pfs.start_commit(parent=commit1, branch=patch_branch)
        assert commit2.branch.name == patch_branch.name
        assert commit2.branch.repo.name == repo.name

        assert client.pfs.branch_exists(master_branch)
        assert client.pfs.branch_exists(patch_branch)

    @staticmethod
    def test_commit_context_mgr(client: TestClient, default_project: bool):
        """Start and finish a commit using a context manager."""
        repo = client.new_repo(default_project)
        branch = pfs.Branch(repo=repo, name="master")

        with client.pfs.commit(branch=branch) as commit1:
            pass
        with client.pfs.commit(branch=branch) as commit2:
            pass

        with pytest.raises(grpc.RpcError):
            with client.pfs.commit(branch=pfs.Branch.from_uri("fake-repo@fake-branch")):
                pass

        assert client.pfs.commit_exists(commit1)
        assert client.pfs.commit_exists(commit2)

    @staticmethod
    def test_wait_commit(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project=True)
        branch = pfs.Branch(repo=repo, name="master")

        with client.pfs.commit(branch=branch) as commit:
            commit.put_file_from_bytes(path="/input.json", data=b"hello world")

        assert client.pfs.wait_commit(commit=commit).finished

    @staticmethod
    def test_inspect_commit(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project)
        branch = pfs.Branch(repo=repo, name="master")

        with client.pfs.commit(branch=branch) as commit:
            commit.put_file_from_bytes(path="/input.json", data=b"hello world")

        # Inspect commit at a specific repo
        commit_info = client.pfs.inspect_commit(
            commit=commit, wait=pfs.CommitState.FINISHED
        )
        assert commit_info.finished
        assert commit_info.commit.branch.name == branch.name
        assert commit_info.commit.branch.repo.name == repo.name

    @staticmethod
    def test_squash_commit(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project)
        branch = pfs.Branch(repo=repo, name="master")

        with client.pfs.commit(branch=branch) as commit1:
            pass

        with client.pfs.commit(branch=branch) as commit2:
            pass

        client.pfs.wait_commit(commit2)

        commit_info = client.pfs.list_commit(repo=repo)
        commits = [info.commit.id for info in commit_info]
        assert commit1.id in commits
        assert commit2.id in commits

        client.pfs.squash_commit_set(commit_set=pfs.CommitSet(id=commit1.id))
        commit_info = client.pfs.list_commit(repo=repo)
        commits = [info.commit.id for info in commit_info]
        assert commit1.id not in commits
        assert commit2.id in commits

    @staticmethod
    def test_drop_commit(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project)
        branch = pfs.Branch(repo=repo, name="master")

        with client.pfs.commit(branch=branch) as commit1:
            pass

        with client.pfs.commit(branch=branch) as commit2:
            pass

        client.pfs.wait_commit(commit2)

        assert client.pfs.commit_exists(commit1)
        assert client.pfs.commit_exists(commit2)

        client.pfs.drop_commit_set(commit_set=pfs.CommitSet(id=commit2.id))

        assert client.pfs.commit_exists(commit1)
        assert not client.pfs.commit_exists(commit2)

    @staticmethod
    def test_subscribe_commit(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project)
        branch = pfs.Branch(repo=repo, name="master")
        commit_generator = client.pfs.subscribe_commit(repo=repo, branch=branch.name)

        with client.pfs.commit(branch=branch) as commit:
            pass

        _generated_commit = next(commit_generator)
        generated_commit = next(commit_generator)
        assert generated_commit.commit.id == commit.id
        assert generated_commit.commit.branch.repo.name == repo.name
        assert generated_commit.commit.branch.name == "master"

    @staticmethod
    def test_list_commit(client: TestClient):
        repo = client.new_repo(default_project=False)
        branch = pfs.Branch(repo=repo, name="master")

        with client.pfs.commit(branch=branch) as _commit1:
            pass

        with client.pfs.commit(branch=branch) as _commit2:
            pass

        commit_info = client.pfs.list_commit(repo=repo)
        assert count(commit_info) >= 2

    @staticmethod
    def test_closed_commit(client: TestClient):
        """Test that an OpenCommit becomes a ClosedCommit when exiting
        the context manager."""
        repo = client.new_repo(default_project=False)
        branch = pfs.Branch(repo=repo, name="master")

        with client.pfs.commit(branch=branch) as commit1:
            assert isinstance(commit1, OpenCommit)

        assert not isinstance(commit1, OpenCommit)
        assert isinstance(commit1, ClosedCommit)


class TestModifyFile:
    """Tests for the modify file API."""

    @staticmethod
    def test_put_file_from_bytes(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project)
        branch = pfs.Branch(repo=repo, name="master")

        with client.pfs.commit(branch=branch) as commit:
            file = commit.put_file_from_bytes(path="/file.dat", data=b"DATA")

        assert client.pfs.commit_exists(commit)
        assert client.pfs.path_exists(file)

    @staticmethod
    def test_put_file_bytes_filelike(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project)
        branch = pfs.Branch(repo=repo, name="master")

        with client.pfs.commit(branch=branch) as commit:
            file = commit.put_file_from_file(path="/file.dat", file=io.BytesIO(b"DATA"))
        assert client.pfs.path_exists(file)

    @staticmethod
    def test_put_empty_file(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project)
        branch = pfs.Branch(repo=repo, name="master")

        with client.pfs.commit(branch=branch) as commit:
            file = commit.put_file_from_bytes(path="/file.dat", data=b"")

        file_info = client.pfs.inspect_file(file=file)
        assert file_info.size_bytes == 0

    @staticmethod
    def test_append_file(client: TestClient, default_project: bool):
        """Test appending to a file."""
        repo = client.new_repo(default_project)
        branch = pfs.Branch(repo=repo, name="master")

        with client.pfs.commit(branch=branch) as commit1:
            commit1.put_file_from_bytes(path="/file.dat", data=b"TWELVE BYTES")
        with client.pfs.commit(branch=branch) as commit2:
            file = commit2.put_file_from_bytes(
                path="/file.dat", data=b"TWELVE BYTES", append=True
            )

        file_info = client.pfs.inspect_file(file=file)
        assert file_info.size_bytes == 24

    @staticmethod
    def test_put_large_file(client: TestClient):
        """Put a file larger than the maximum message size."""
        # TODO: This functionality is tested in TestPFSFile::test_get_large_file.
        repo = client.new_repo()
        branch = pfs.Branch(repo=repo, name="master")

        with client.pfs.commit(branch=branch) as commit:
            data = b"#" * MAX_RECEIVE_MESSAGE_SIZE * 2
            file = commit.put_file_from_bytes(path="/file.dat", data=data)
        assert client.pfs.path_exists(file)

    @staticmethod
    def test_put_file_url(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project)
        branch = pfs.Branch(repo=repo, name="master")

        with client.pfs.commit(branch=branch) as commit:
            url = "https://www.pachyderm.com/index.html"
            file = commit.put_file_from_url(path="/index.html", url=url)
        assert client.pfs.path_exists(file)

    @staticmethod
    def test_copy_file(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project)
        branch = pfs.Branch(repo=repo, name="master")

        with client.pfs.commit(branch=branch) as src_commit:
            file1 = src_commit.put_file_from_bytes(path="/file1.dat", data=b"DATA1")
            file2 = src_commit.put_file_from_bytes(path="/file2.dat", data=b"DATA2")

        with client.pfs.commit(branch=branch) as dest_commit:
            file_copy = dest_commit.copy_file(src=file1, dst="/copy.dat")

        for file in (file1, file2, file_copy):
            assert client.pfs.path_exists(file)

    @staticmethod
    def test_delete_file(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project)
        branch = pfs.Branch(repo=repo, name="master")

        with client.pfs.commit(branch=branch) as commit1:
            file = commit1.put_file_from_bytes(path="/file1.dat", data=b"DATA")
        assert client.pfs.path_exists(file)

        with client.pfs.commit(branch=branch) as commit2:
            commit2.delete_file(path="/file1.dat")
        assert not client.pfs.path_exists(file=pfs.File.from_uri(f"{commit2}:/file1.dat"))

    @staticmethod
    def test_walk_file(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project)
        branch = pfs.Branch(repo=repo, name="master")

        with client.pfs.commit(branch=branch) as commit:
            commit.put_file_from_bytes(path="/file1.dat", data=b"DATA")
            commit.put_file_from_bytes(path="/a/file2.dat", data=b"DATA")
            commit.put_file_from_bytes(path="/a/b/file3.dat", data=b"DATA")

        files = list(client.pfs.walk_file(file=pfs.File(commit=commit, path="/a")))
        assert len(files) == 4
        assert files[0].file.path == "/a/"
        assert files[1].file.path == "/a/b/"
        assert files[2].file.path == "/a/b/file3.dat"
        assert files[3].file.path == "/a/file2.dat"

    @staticmethod
    def test_glob_file(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project)
        branch = pfs.Branch(repo=repo, name="master")

        with client.pfs.commit(branch=branch) as commit:
            commit.put_file_from_bytes(path="file1.dat", data=b"DATA")
            commit.put_file_from_bytes(path="file2.dat", data=b"DATA")

        file_info = list(client.pfs.glob_file(commit=commit, pattern="/*.dat"))
        assert len(file_info) == 2

        file_info = list(client.pfs.glob_file(commit=commit, pattern="/*1.dat"))
        assert len(file_info) == 1
        assert file_info[0].file.path == "/file1.dat"

    @staticmethod
    def test_diff_file(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project)
        branch = pfs.Branch(repo=repo, name="master")

        with client.pfs.commit(branch=branch) as commit:
            file1_old = commit.put_file_from_bytes(path="/file1.dat", data=b"old data 1")
            file2_old = commit.put_file_from_bytes(path="/file2.dat", data=b"old data 2")

        with client.pfs.commit(branch=branch) as commit:
            file1_new = commit.put_file_from_bytes(path="/file1.dat", data=b"new data 1")

        diff = list(client.pfs.diff_file(new_file=file1_new, old_file=file2_old))
        assert diff[0].new_file.file.path == file1_old.path
        assert diff[1].old_file.file.path == file2_old.path

    @staticmethod
    def test_path_exists(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project)
        branch = pfs.Branch(repo=repo, name="master")

        with client.pfs.commit(branch=branch) as commit:
            commit.put_file_from_bytes(path="/dir/file1", data=b"I'm a file in a dir.")
            commit.put_file_from_bytes(path="/file2", data=b"I'm a file.")

        assert client.pfs.path_exists(file=pfs.File(commit=commit, path="/"))
        assert client.pfs.path_exists(file=pfs.File(commit=commit, path="/dir/"))
        assert client.pfs.path_exists(file=pfs.File(commit=commit, path="/dir"))
        assert client.pfs.path_exists(file=pfs.File(commit=commit, path="/dir/file1"))
        assert client.pfs.path_exists(file=pfs.File(commit=commit, path="/dir/file1/"))
        assert client.pfs.path_exists(file=pfs.File(commit=commit, path="/file2"))
        assert not client.pfs.path_exists(file=pfs.File(commit=commit, path="/file1"))

        with pytest.raises(ValueError, match=r"commit does not exist"):
            assert not client.pfs.path_exists(
                file=pfs.File.from_uri("fake_repo@master:/dir")
            )


class TestPFSFile:
    @staticmethod
    def test_get_large_file(client: TestClient):
        """Test that a large file (requires >1 gRPC message to stream)
        is successfully streamed in it's entirety.
        """
        # Arrange
        repo = client.new_repo()
        branch = pfs.Branch(repo=repo, name="master")
        data = os.urandom(int(MAX_RECEIVE_MESSAGE_SIZE * 1.1))

        with client.pfs.commit(branch=branch) as commit:
            file = commit.put_file_from_bytes(path="/large_file.dat", data=data)

        # Act
        with client.pfs.pfs_file(file) as pfs_file:
            assert len(pfs_file._buffer) < len(data), (
                "PFSFile initialization streams the first message. "
                "This asserts that the test file must be streamed over multiple messages. "
            )
            streamed_data = pfs_file.read()

        # Assert
        assert streamed_data[:10] == data[:10]
        assert streamed_data[-10:] == data[-10:]
        assert len(streamed_data) == len(data)

    @staticmethod
    def test_buffered_data():
        """Test that data gets buffered as expected."""
        # Arrange
        stream_items = [b"a", b"bc", b"def", b"ghij", b"klmno"]
        stream = (betterproto.BytesValue(value=item) for item in stream_items)

        # Act
        file = PFSFile(stream)

        # Assert
        assert file._buffer == b"a"
        assert file.read(0) == b""
        assert file._buffer == b"a"
        assert file.read(2) == b"ab"
        assert file._buffer == b"c"
        assert file.read(5) == b"cdefg"
        assert file._buffer == b"hij"

    @staticmethod
    def test_fail_early(client: TestClient):
        """Test that gRPC errors are caught and thrown early."""
        # Arrange
        repo = client.new_repo()
        invalid_file = pfs.File.from_uri(f"{repo}@master:/NO_FILE.HERE")

        # Act & Assert
        with pytest.raises(ValueError):
            with client.pfs.pfs_file(file=invalid_file):
                pass

    @staticmethod
    def test_context_manager(client: TestClient):
        """Test that the PFSFile context manager cleans up as expected.

        Note: The test file needs to be large in order to span multiple
          gRPC messages. If the file only requires a single message to stream
          then the stream will be terminated and the assertion within the
          `client.get_file` context will fail.
          Maybe this should be rethought or mocked in the future.
        """
        # Arrange
        repo = client.new_repo()
        branch = pfs.Branch(repo=repo, name="master")
        data = os.urandom(int(MAX_RECEIVE_MESSAGE_SIZE * 1.1))

        with client.pfs.commit(branch=branch) as commit:
            file = commit.put_file_from_bytes(path="/large_file.dat", data=data)

        # Act & Assert
        with client.pfs.pfs_file(file) as pfs_file:
            assert pfs_file.read(1)

    @staticmethod
    def test_file_obj(client: TestClient):
        """Test normal file object functionality of PFSFile"""
        # Arrange
        repo = client.new_repo()
        branch = pfs.Branch(repo=repo, name="master")
        data = b"0,1,2,3\n4,5,6,7\na,b,c,d"
        expected_lines = data.splitlines(keepends=True)

        with client.pfs.commit(branch=branch) as commit:
            file = commit.put_file_from_bytes(path="/test.csv", data=data)

        # Act & Assert
        with client.pfs.pfs_file(file) as pfs_file:
            assert pfs_file.readlines() == expected_lines

        # Test that we can feed it into stdlib functionality (csv reader).
        import csv

        with client.pfs.pfs_file(file) as pfs_file:
            with io.TextIOWrapper(pfs_file, encoding="utf-8") as text_file:
                assert len(list(csv.reader(text_file))) > 0

        # Test that we can feed it into a pandas dataframe.
        import pandas as pd

        with client.pfs.pfs_file(file) as pfs_file:
            assert len(pd.read_csv(pfs_file)) > 0

    @staticmethod
    def test_get_file_tar(client: TestClient, tmp_path: Path):
        """Test that retrieving a TAR of a PFS directory works as expected."""
        # Arrange
        TestPfsFile = NamedTuple("TestPfsFile", (("data", bytes), ("path", str)))
        test_files = [
            TestPfsFile(os.urandom(1024), "/a.dat"),
            TestPfsFile(os.urandom(2048), "/child/b.dat"),
            TestPfsFile(os.urandom(5000), "/child/grandchild/c.dat"),
        ]

        repo = client.new_repo()
        branch = pfs.Branch(repo=repo, name="master")

        with client.pfs.commit(branch=branch) as commit:
            for test_file in test_files:
                commit.put_file_from_bytes(path=test_file.path, data=test_file.data)

        # Act
        root_path = pfs.File(commit=commit, path="/")
        with client.pfs.pfs_tar_file(file=root_path) as pfs_tar_file:
            pfs_tar_file.extractall(tmp_path)

        # Assert
        for test_file in test_files:
            local_file = tmp_path.joinpath(test_file.path[1:])
            assert local_file.exists()
            assert local_file.read_bytes() == test_file.data
