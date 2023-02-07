import io
import subprocess
from contextlib import contextmanager
from dataclasses import asdict
from pathlib import Path
from typing import Iterator, Union

from betterproto.lib.google.protobuf import Empty
import grpc

from . import ApiStub as _GeneratedApiStub
from . import (
    Branch,
    Commit,
    CommitInfo,
    CommitState,
    File,
    ModifyFileRequest,
    AddFile,
    AddFileUrlSource,
    CopyFile,
    DeleteFile,
)
from .file import PFSFile, PFSTarFile
from .utils import check_pachctl

BUFFER_SIZE = 19 * 1024 * 1024  # 19MB


class OpenCommit(Commit):

    def __init__(self, commit: "Commit", stub: "ApiStub"):
        self._commit = commit
        self._stub = stub
        super().__init__(**asdict(commit))

    def wait(self) -> CommitInfo:
        return self._stub.wait_commit(self)

    def put_file_from_bytes(
        self,
        path: str,
        data: bytes,
        append: bool = False
    ) -> "File":
        """Uploads a PFS file from a bytestring.

        Parameters
        ----------
        path : str
            The path in the repo the data will be written to.
        data : bytes
            The file contents as bytes.
        append : bool, optional
            If true, appends the data to the file specified at `path`, if
            they already exist. Otherwise, overwrites them.

        Examples
        --------
        Commit needs to be open still, either from the result of
        ``start_commit()`` or within scope of ``commit()``

        >>> with client.pfs.commit(branch=pfs.Branch.from_uri("images@master")) as commit:
        >>>     commit.put_file_from_bytes(path="/file.txt", data=b"SOME BYTES")
        """
        self._stub.put_file_from_bytes(
            commit=self, path=path, data=data, append=append
        )
        return File(commit=self, path=path)

    def put_file_from_url(
        self,
        *,
        path: str,
        url: str,
        recursive: bool = False,
    ) -> "File":
        """Uploads a PFS file from an url.

        Parameters
        ----------
        path : str
            The path in the repo the data will be written to.
        url : str
            The URL of the file to put.
        recursive : bool
            If true, allows for recursive scraping on some types URLs, for
            example on s3:// URLs
        """
        self._stub.put_file_from_url(
            commit=self, path=path, url=url, recursive=recursive
        )
        return File(commit=self, path=path)

    def put_file_from_file(
        self,
        *,
        path: str,
        file: io.BytesIO,  # TODO: Get correct type.
        append: bool = False
    ) -> "File":
        """Uploads a PFS file from an open file object.

        Parameters
        ----------
        path : str
            The path in the repo the data will be written to.
        file : io.BytesIO  # TODO: Get correct type.
            An open file object to read the data from.
        append : bool, optional
            If true, appends the data to the file specified at `path`, if
            they already exist. Otherwise, overwrites them.
        """
        self._stub.put_file_from_file(
            commit=self, path=path, file=file, append=append
        )
        return File(commit=self, path=path)

    def copy_file(
        self,
        *,
        src: "File",
        dst: str,
        append: bool = True
    ) -> "File":
        """Copies a file within PFS

        Parameters
        ----------
        src : pfs.File
            This file to be copied.
        dst : str
            The destination of the file, as a string path.
        append : bool
            If true, appends the contents of src to dst if it exists.
            Otherwise, overwrites the file.
        """
        self._stub.copy_file(commit=self, src=src, dst=dst, append=append)
        return File(commit=self, path=dst)

    def delete_file(self, *, path: str) -> "File":
        """Copies a file within PFS

        Parameters
        ----------
        path : str
            The path of the file to be deleted.
        """
        self._stub.delete_file(commit=self, path=path)
        return File(commit=self, path=path)


class ApiStub(_GeneratedApiStub):

    @contextmanager
    def commit(
        self, *, parent: "Commit" = None, description: str = "", branch: "Branch" = None
    ) -> Iterator["OpenCommit"]:
        """A context manager for running operations within a commit.

        Parameters
        ----------
        parent : pfs.Commit
            The parent commit of the new commit. parent may be empty in which case
            the commit that Branch points to will be used as the parent.
            If the branch does not exist, the commit will have no parent.
        description : str, optional
            A description of the commit.
        branch : pfs.Branch
            The branch where the commit is created.

        Yields
        -------
        pfs.Commit
            A protobuf object that represents a commit.

        Examples
        --------
        >>> with client.commit(branch=pfs.Branch.from_uri("images@master")) as c:
        >>>     client.delete_file(c, "/dir/delete_me.txt")
        >>>     client.put_file_bytes(c, "/new_file.txt", b"DATA")
        """
        commit = self.start_commit(parent=parent, description=description, branch=branch)
        try:
            yield OpenCommit(commit=commit, stub=self)
        finally:
            self.finish_commit(commit=commit)

    def wait_commit(self, commit: "Commit") -> "CommitInfo":
        return self.inspect_commit(commit=commit, wait=CommitState.FINISHED)

    def put_file_from_bytes(
        self,
        *,
        commit: "Commit",
        path: str,
        data: bytes,
        append: bool = False
    ):
        """Uploads a PFS file from a bytestring.

        Parameters
        ----------
        commit : pfs.Commit
            An open commit to modify.
        path : str
            The path in the repo the data will be written to.
        data : bytes
            The file contents as bytes.
        append : bool, optional
            If true, appends the data to the file specified at `path`, if
            they already exist. Otherwise, overwrites them.

        Examples
        --------
        Commit needs to be open still, either from the result of
        ``start_commit()`` or within scope of ``commit()``

        >>> with client.commit(branch=pfs.Branch.from_uri("images@master")) as c:
        >>>     client.put_file_bytes(c, "/file.txt", b"SOME BYTES")
        """
        # TODO: Should this just route through put_file_from_file?

        operations = [ModifyFileRequest(set_commit=commit)]
        if not append:
            operations.append(ModifyFileRequest(delete_file=DeleteFile(path=path)))
        operations.append(ModifyFileRequest(add_file=AddFile(path=path, raw=data)))
        return self.modify_file(iter(operations))

    def put_file_from_url(
        self,
        *,
        commit: "Commit",
        path: str,
        url: str,
        recursive: bool = False,
    ) -> Empty:
        """Uploads a PFS file from an url.

        Parameters
        ----------
        commit : pfs.Commit
            An open commit to modify.
        path : str
            The path in the repo the data will be written to.
        url : str
            The URL of the file to put.
        recursive : bool
            If true, allows for recursive scraping on some types URLs, for
            example on s3:// URLs
        """
        operations = [
            ModifyFileRequest(set_commit=commit),
            ModifyFileRequest(delete_file=DeleteFile(path=path)),
            ModifyFileRequest(
                add_file=AddFile(
                    path=path,
                    url=AddFileUrlSource(url=url, recursive=recursive)
                )
            )
        ]
        return self.modify_file(iter(operations))

    def put_file_from_file(
        self,
        *,
        commit: "Commit",
        path: str,
        file: io.BytesIO,  # TODO: Get correct type.
        append: bool = False
    ) -> Empty:
        """Uploads a PFS file from an open file object.

        Parameters
        ----------
        commit : pfs.Commit
            An open commit to modify.
        path : str
            The path in the repo the data will be written to.
        file : io.BytesIO  # TODO: Get correct type.
            An open file object to read the data from.
        append : bool, optional
            If true, appends the data to the file specified at `path`, if
            they already exist. Otherwise, overwrites them.

        Examples
        --------
        Commit needs to be open still, either from the result of
        ``start_commit()`` or within scope of ``commit()``

        >>> with client.commit(branch=pfs.Branch.from_uri("images@master")) as c:
        >>>     client.put_file_bytes(c, "/file.txt", b"SOME BYTES")
        """
        # TODO: Can we verify that the file is outputting bytes?
        def operations() -> Iterator[ModifyFileRequest]:
            yield ModifyFileRequest(set_commit=commit)
            if not append:
                yield ModifyFileRequest(delete_file=DeleteFile(path=path))
            while True:
                data = file.read(BUFFER_SIZE)
                if len(data) == 0:
                    return
                yield ModifyFileRequest(add_file=AddFile(path=path, raw=data))
        return self.modify_file(operations())

    def copy_file(
        self,
        *,
        commit: "Commit",
        src: "File",
        dst: str,
        append: bool = True
    ) -> Empty:
        """Copies a file within PFS

        Parameters
        ----------
        commit : pfs.Commit
            An open commit to modify.
        src : pfs.File
            This file to be copied.
        dst : str
            The destination of the file, as a string path.
        append : bool
            If true, appends the contents of src to dst if it exists.
            Otherwise, overwrites the file.
        """
        operations = [
            ModifyFileRequest(set_commit=commit),
            ModifyFileRequest(
                copy_file=CopyFile(dst=dst, src=src, append=append)
            )
        ]
        return self.modify_file(iter(operations))

    def delete_file(self, *, commit: "Commit", path: str) -> Empty:
        """Copies a file within PFS

        Parameters
        ----------
        commit : pfs.Commit
            An open commit to modify.
        path : str
            The path of the file to be deleted.
        """
        operations = [
            ModifyFileRequest(set_commit=commit),
            ModifyFileRequest(delete_file=DeleteFile(path=path)),
        ]
        return self.modify_file(iter(operations))

    def path_exists(self, *, file: "File") -> bool:
        """Checks whether the path exists in the specified commit, agnostic to
        whether `path` is a file or a directory.

        Parameters
        ----------
        file : pfs.File
            The file (or directory) to check.

        Returns
        -------
        bool
            True if the path exists.
        """
        try:
            self.inspect_commit(commit=file.commit)
        except grpc.RpcError as e:
            raise ValueError("commit does not exist") from e

        try:
            self.inspect_file(file=file)
        except grpc.RpcError:
            return False
        return True

    @contextmanager
    def pfs_file(self, file: "File") -> Iterator[PFSFile]:
        stream = self.get_file(file=file)
        yield PFSFile(stream)

    @contextmanager
    def pfs_tar_file(self, file: "File") -> Iterator[PFSTarFile]:
        stream = self.get_file_tar(file=file)
        yield PFSTarFile.open(fileobj=PFSFile(stream), mode="r|*")

    def _mount(self, mount_dir: Union[str, Path], commit: "Commit") -> subprocess.Popen:
        # TODO: Check SUDO (used in unmount).
        check_pachctl(ensure_mount=True)
        mount_dir = Path(mount_dir)
        if mount_dir.is_file():
            raise NotADirectoryError(mount_dir)

        mount_dir.mkdir(parents=True, exist_ok=True)
        if any(mount_dir.iterdir()):
            raise RuntimeError(
                f"{mount_dir} must be empty to mount (including hidden files)"
            )

        process = subprocess.Popen(
            ["pachctl", "mount", str(mount_dir), "-r", commit.as_uri()],
            stdout=subprocess.DEVNULL,
            stderr=subprocess.STDOUT
        )

        # Ensure mount has finished
        import time
        for _ in range(4):
            time.sleep(0.25)
            if any(mount_dir.iterdir()):
                return process
        else:
            self._unmount(commit)
            raise RuntimeError(
                "mount failed to expose data after four read attempts (1.0s)"
            )

    def _unmount(self, mount_dir: Union[str, Path]) -> None:
        check_pachctl(ensure_mount=True)
        subprocess.run(["sudo", "pachctl", "unmount", mount_dir])

    @contextmanager
    def mounted(self, commit: "Commit", mount_dir: Union[str, Path]) -> Iterator[Path]:
        """Mounts Pachyderm commits locally.

        Parameters
        ----------
        commit : pfs.Commit
            The commit to be mounted.
        mount_dir : str
            The directory to commit within.
            This directory must be empty (including hidden files).

        Notes
        -----
        Mounting uses FUSE, which causes some known issues on macOS. For the
        best experience, we recommend using mount on Linux. We do not fully
        support mounting on macOS 1.11 and later.

        Yields
        ------
        The subdirectory where the commit was mounted.

        Examples
        --------
        >>> with client.pfs.mounted(pfs.Commit.from_uri("images@mount^2"), "/pfs") as mount:
        >>>     print(list(mount.iterdir()))
        """
        _process = self._mount(mount_dir, commit)
        mounted_commit = Path(mount_dir, commit.branch.repo.name)
        assert mounted_commit.exists()
        yield mounted_commit
        self._unmount(mount_dir)
