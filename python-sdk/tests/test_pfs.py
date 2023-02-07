import io
import os
from pathlib import Path
from typing import NamedTuple

import betterproto
import grpc

from tests.test_client import TestClient
from tests.fixtures import *

from pachyderm_sdk.api import pfs
from pachyderm_sdk.api.pfs.file import PFSFile
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
        resp = client.pfs.inspect_project(project)
        assert resp.project.name == project.name

    @staticmethod
    def test_delete_project(client: TestClient):
        project = client.new_project()
        orig_project_count = len(list(client.pfs.list_project()))
        assert orig_project_count >= 1
        client.pfs.delete_project(project)
        assert len(list(client.pfs.list_project())) == orig_project_count - 1


class TestsUnitRepo:
    """Unit tests for the Repo management API."""

    @staticmethod
    def test_list_repo(client: TestClient, default_project: bool):
        client.new_repo(default_project)
        all_repos = list(client.pfs.list_repo())
        user_repos = list(client.pfs.list_repo(type="user"))
        assert len(all_repos) >= 1
        assert len(user_repos) >= 1
        assert len(all_repos) > len(user_repos)

    @staticmethod
    def test_inspect_repo(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project)
        response = client.pfs.inspect_repo(repo)
        assert response.repo.name == repo.name

    @staticmethod
    def test_delete_repo(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project)
        orig_repo_count = len(list(client.pfs.list_repo()))
        assert orig_repo_count >= 1
        client.pfs.delete_repo(repo)
        assert len(list(client.pfs.list_repo())) == orig_repo_count - 1

    @staticmethod
    def test_delete_non_existent_repo(client: TestClient):
        orig_repo_count = len(list(client.pfs.list_repo()))
        client.pfs.delete_repo(pfs.Repo.from_uri("BOGUS_NAME"))
        assert len(list(client.pfs.list_repo())) == orig_repo_count

    @staticmethod
    def test_delete_repos(client: TestClient):
        repo = client.new_repo(default_project=False)
        project = repo.project
        orig_repo_count = len(list(client.pfs.list_repo(projects=[project])))
        assert orig_repo_count >= 1
        client.pfs.delete_repos(projects=[project])
        assert len(list(client.pfs.list_repo())) == orig_repo_count - 1


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
        assert master_branch.name in {info.branch.name for info in branch_info}
        assert develop_branch.name in {info.branch.name for info in branch_info}

    @staticmethod
    def test_delete_branch(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project)
        branch = pfs.Branch(repo=repo, name="develop")

        client.pfs.create_branch(branch=branch)

        branch_info = list(client.pfs.list_branch(repo=repo))
        assert len(branch_info) == 1
        client.pfs.delete_branch(branch=branch)
        branch_info = list(client.pfs.list_branch(repo=repo))
        assert len(branch_info) == 0

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
        with pytest.raises(
            grpc.RpcError, match=r"parent commit .* has not been finished"
        ):
            client.pfs.start_commit(branch=branch)

        client.pfs.finish_commit(commit)

        with pytest.raises(grpc.RpcError, match=r"repo .* not found"):
            client.pfs.start_commit(
                branch=pfs.Branch.from_uri("fake-repo:fake-branch")
            )

    @staticmethod
    def test_start_with_parent(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project)
        branch = pfs.Branch(repo=repo, name="master")

        commit1 = client.pfs.start_commit(branch=branch)
        client.pfs.finish_commit(commit1)

        commit2 = client.pfs.start_commit(parent=commit1, branch=branch)
        assert commit2.branch.repo.name == repo.name

    @staticmethod
    def test_start_commit_fork(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project)
        master_branch = pfs.Branch(repo=repo, name="master")
        patch_branch = pfs.Branch(repo=repo, name="patch")

        commit1 = client.pfs.start_commit(branch=master_branch)
        client.pfs.finish_commit(commit1)

        commit2 = client.pfs.start_commit(parent=commit1, branch=patch_branch)
        assert commit2.branch.name == patch_branch.name
        assert commit2.branch.repo.name == repo.name

        branch_names = [
            branch_info.branch.name
            for branch_info in
            list(client.pfs.list_branch(repo=repo))
        ]
        assert master_branch.name in branch_names
        assert patch_branch.name in branch_names

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
            fake_branch = pfs.Branch.from_uri("fake-repo:fake-branch")
            with client.pfs.commit(branch=fake_branch):
                pass

        commit_infos = list(client.pfs.list_commit(repo=repo))
        assert len(commit_infos) == 2
        assert commit1.id in {info.commit.id for info in commit_infos}
        assert commit2.id in {info.commit.id for info in commit_infos}

    @staticmethod
    def test_wait_commit(client: TestClient, default_project: bool):
        # TODO: This is more of an integration test. May want to break out.
        repo1 = client.new_repo(default_project=True)
        branch1 = pfs.Branch(repo=repo1, name="master")
        repo2 = client.new_repo(default_project)
        branch2 = pfs.Branch(repo=repo2, name="master")

        # Create provenance between repos (which creates a new commit)
        client.pfs.create_branch(branch=branch2, provenance=[branch1])

        # Head commit is still open in repo2  # TODO: Verify this?
        client.pfs.finish_commit(commit=pfs.Commit(branch=branch2))

        with client.pfs.commit(branch=branch1) as commit1:
            client.pfs.put_file_from_bytes(commit1, "/input.json", b"hello world")
        client.pfs.finish_commit(commit=pfs.Commit(branch=branch2))  # TODO: WHY?

        # Just block until all of the commits are yielded
        # TODO: This will need to be updated because we no longer return commit set.
        info1 = client.pfs.wait_commit(commit=commit1)
        #assert len(commit) == 2
        assert info1.finished  # This should be a timestamp?

        with client.pfs.commit(branch=branch1) as commit2:
            client.pfs.put_file_from_bytes(commit2, "/input.json", b"bye world")
        client.pfs.finish_commit(commit=pfs.Commit(branch=branch2))  # TODO: WHY?

        # Just block until the commit in repo1 is finished
        info2 = client.pfs.wait_commit(commit=commit2)
        #assert len(commits) == 1
        assert info2.finished

        files = list(
            client.pfs.list_file(file=pfs.File(commit=commit2, path="/"))
        )
        assert len(files) == 1

    @staticmethod
    def test_inspect_commit(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project)
        branch = pfs.Branch(repo=repo, name="master")

        with client.pfs.commit(branch=branch) as commit:
            client.pfs.put_file_from_bytes(commit, "input.json", b"hello world")

        # Inspect commit at a specific repo
        commit_info = client.pfs.inspect_commit(
            commit=commit, wait=pfs.CommitState.FINISHED
        )
        assert commit.finished
        assert commit_info.commit.branch.name == branch.name
        assert commit.commit.branch.repo.name == repo.name
        assert commit.size_bytes == 11

    @staticmethod
    def test_squash_commit(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project)
        branch = pfs.Branch(repo=repo, name="master")

        with client.pfs.commit(branch=branch) as commit1:
            pass

        with client.pfs.commit(branch=branch) as commit2:
            pass

        client.pfs.wait_commit(commit2)

        commits = list(client.pfs.list_commit(repo=repo))
        assert len(commits) == 2

        client.pfs.squash_commit_set(commit_set=pfs.CommitSet(id=commit1.id))
        commits = list(client.pfs.list_commit(repo=repo))
        assert len(commits) == 1

    @staticmethod
    def test_drop_commit(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project)
        branch = pfs.Branch(repo=repo, name="master")

        with client.pfs.commit(branch=branch) as commit1:
            pass

        with client.pfs.commit(branch=branch) as commit2:
            pass

        client.pfs.wait_commit(commit2)

        commits = list(client.pfs.list_commit(repo=repo))
        assert len(commits) == 2

        client.pfs.drop_commit_set(commit_set=pfs.CommitSet(id=commit2.id))
        commits = list(client.pfs.list_commit(repo=repo))
        assert len(commits) == 1

    @staticmethod
    def test_subscribe_commit(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project)
        branch = pfs.Branch(repo=repo, name="master")
        commit_generator = client.pfs.subscribe_commit(branch=branch)

        with client.pfs.commit(branch=branch) as commit:
            pass

        generated_commit = next(commit_generator)
        assert generated_commit.commit.id == commit.id
        assert generated_commit.commit.branch.repo.name == repo.name
        assert generated_commit.commit.branch.name == "master"

    @staticmethod
    def test_list_commit(client: TestClient):
        repo = client.new_repo(default_project=False)
        branch = pfs.Branch(repo=repo, name="master")

        with client.pfs.commit(branch=branch) as commit1:
            pass

        with client.pfs.commit(branch=branch) as commit2:
            pass

        commit_info = list(client.pfs.list_commit(repo=repo))
        assert len(commit_info) == 2
        assert commit1.id in {commit.commit.id for commit in commit_info}
        assert commit2.id in {commit.commit.id for commit in commit_info}


class TestModifyFile:
    """Tests for the modify file API."""

    @staticmethod
    def test_put_file_from_bytes(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project)
        branch = pfs.Branch(repo=repo, name="master")

        with client.pfs.commit(branch=branch) as commit:
            client.pfs.put_file_from_bytes(commit=commit, path="/file.dat", data=b"DATA")

        commit_infos = list(client.pfs.list_commit(repo=repo))
        assert len(commit_infos) == 1
        assert commit_infos[0].commit.id == commit.id
        files = list(client.pfs.list_file(file=pfs.File(commit=commit, path="/")))
        assert len(files) == 1

    @staticmethod
    def test_put_file_bytes_filelike(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project)
        branch = pfs.Branch(repo=repo, name="master")

        with client.pfs.commit(branch=branch) as commit:
            client.pfs.put_file_from_file(commit=commit, path="/file.dat", file=io.BytesIO(b"DATA"))

        files = list(client.pfs.list_file(file=pfs.File(commit=commit, path="/")))
        assert len(files) == 1

    @staticmethod
    def test_put_empty_file(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project)
        branch = pfs.Branch(repo=repo, name="master")

        with client.pfs.commit(branch=branch) as commit:
            client.pfs.put_file_from_bytes(commit=commit, path="/file.dat", data=b"")

        files = list(client.pfs.list_file(file=pfs.File(commit=commit, path="/")))
        assert len(files) == 1
        file = client.pfs.inspect_file(
            file=pfs.File(commit=commit, path="/file.dat")
        )
        assert file.size_bytes == 0

    @staticmethod
    def test_append_file(client: TestClient, default_project: bool):
        """Test appending to a file."""
        repo = client.new_repo(default_project)
        branch = pfs.Branch(repo=repo, name="master")

        with client.pfs.commit(branch=branch) as commit1:
            client.pfs.put_file_from_bytes(
                commit=commit1, path="/file.dat", data=b"TWELVE BYTES"
            )
        with client.pfs.commit(branch=branch) as commit2:
            client.pfs.put_file_from_bytes(
                commit=commit2, path="/file.dat", data=b"TWELVE BYTES", append=True
            )

        file_info = client.pfs.inspect_file(pfs.File(commit=commit2, path="/file.dat"))
        assert file_info.size_bytes == 24

    @staticmethod
    def test_put_large_file(client: TestClient, default_project: bool):
        """Put a file larger than the maximum message size."""
        repo = client.new_repo(default_project)
        branch = pfs.Branch(repo=repo, name="master")

        with client.pfs.commit(branch=branch) as commit:
            data = b"#" * MAX_RECEIVE_MESSAGE_SIZE * 2
            client.pfs.put_file_from_bytes(commit=commit, path="/file.dat", data=data)

        files = list(client.pfs.list_file(file=pfs.File(commit=commit, path="/")))
        assert len(files) == 1

    @staticmethod
    def test_put_file_url(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project)
        branch = pfs.Branch(repo=repo, name="master")

        with client.pfs.commit(branch=branch) as commit:
            url = "https://gist.githubusercontent.com/ysimonson/1986773831f6c4c292a7290c5a5d4405/raw/fb2b4d03d317816e36697a6864a9c27645baa6c0/wheel.html"  # TODO: Replace with different url.
            client.pfs.put_file_from_url(commit=commit, path="/index.html", url=url)

        file_info = list(client.pfs.list_file(file=pfs.File(commit=commit, path="/")))
        assert len(file_info) == 1
        assert file_info[0].file.path == "/index.html"

    @staticmethod
    def test_copy_file(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project)
        branch = pfs.Branch(repo=repo, name="master")

        with client.pfs.commit(branch=branch) as src_commit:
            client.pfs.put_file_bytes(commit=src_commit, path="/file1.dat", data=io.BytesIO(b"DATA1"))
            client.pfs.put_file_bytes(commit=src_commit, path="/file2.dat", data=io.BytesIO(b"DATA2"))

        with client.pfs.commit(branch=branch) as dest_commit:
            src = pfs.File(commit=src_commit, path="/file1.dat")
            client.pfs.copy_file(commit=dest_commit, src=src, dst="/copy.dat")

        file_info = list(
            client.pfs.list_file(file=pfs.File(commit=dest_commit, path="/"))
        )
        assert len(file_info) == 3
        file_paths = {info.file.path for info in file_info}
        for test_file in ("/copy.dat", "/file1.dat", "/file2.dat"):
            assert test_file in file_paths

    @staticmethod
    def test_delete_file(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project)
        branch = pfs.Branch(repo=repo, name="master")

        with client.pfs.commit(branch=branch) as commit1:
            client.pfs.put_file_from_bytes(commit=commit1, path="/file1.dat", data=b"DATA")

        file_info = list(
            client.pfs.list_file(file=pfs.File(commit=commit1, path="/file1.dat"))
        )
        assert len(file_info) == 1

        with client.pfs.commit(branch=branch) as commit2:
            client.pfs.delete_file(commit=commit2, path="/file1.dat")

        file_info = list(
            client.pfs.list_file(file=pfs.File(commit=commit2, path="/file1.dat"))
        )
        assert len(file_info) == 0

# TODO: Started using OpenCommit from here.

    @staticmethod
    def test_walk_file(client: TestClient, default_project: bool):
        repo = client.new_repo(default_project)
        branch = pfs.Branch(repo=repo, name="master")

        with client.pfs.commit(branch=branch) as commit:
            commit.put_file_from_bytes(path="/file1.dat", data=b"DATA")
            commit.put_file_from_bytes(path="/a/file2.dat", data=b"DATA")
            commit.put_file_from_bytes(path="/a/b/file3.dat", data=b"DATA")

        files = list(client.pfs.walk_file(commit, "/a"))
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
                commit=pfs.Commit.from_uri("fake_repo@master"), path="/dir"
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
            streamed_data = file.read()

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
        invalid_file = pfs.File.from_uri(f"{repo.to_uri()}@master:/NO_FILE.HERE")

        # Act & Assert
        with pytest.raises(ConnectionError):
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
            assert pfs_file._stream.is_active()
        assert not pfs_file._stream.is_active()

    @staticmethod
    def test_cancelled_stream(client: TestClient):
        """Test that a cancelled stream maintains the integrity of the
        already-streamed data.
        """
        # Arrange
        repo = client.new_repo()
        branch = pfs.Branch(repo=repo, name="master")
        data = os.urandom(200)

        with client.pfs.commit(branch=branch) as commit:
            file = commit.put_file_from_bytes(path="/test_file.dat", data=data)

        # Act & Assert
        with client.pfs.pfs_file(file=file) as pfs_file:
            assert pfs_file.read(100) == data[:100]
        assert pfs_file.read(100) == data[100:]

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
        with client.pfs.pfs_tar_file(commit, "/") as pfs_tar_file:
            pfs_tar_file.extractall(tmp_path)

        # Assert
        for test_file in test_files:
            local_file = tmp_path.joinpath(test_file.path[1:])
            assert local_file.exists()
            assert local_file.read_bytes() == test_file.data
