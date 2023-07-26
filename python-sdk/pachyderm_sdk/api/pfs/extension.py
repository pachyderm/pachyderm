"""Handwritten classes/methods that augment the existing PFS API."""
import io
import os
from contextlib import contextmanager
from dataclasses import fields
from functools import wraps
from pathlib import Path
from typing import Callable, ContextManager, Iterable, List, Union, TYPE_CHECKING

from betterproto.lib.google.protobuf import Empty
import grpc

from ...errors import InvalidTransactionOperation
from . import ApiStub as _GeneratedApiStub
from . import (
    Branch,
    Commit,
    CommitInfo,
    CommitSet,
    CommitState,
    File,
    Project,
    Repo,
    ModifyFileRequest,
    AddFile,
    AddFileUrlSource,
    CopyFile,
    DeleteFile,
)
from .file import PFSFile, PFSTarFile

if TYPE_CHECKING:
    from _typeshed import SupportsRead

BUFFER_SIZE = 19 * 1024 * 1024  # 19MB

__all__ = ("ApiStub", "ClosedCommit", "OpenCommit")


def transaction_incompatible(pfs_method: Callable) -> Callable:
    """Decorator for marking methods of the PFS API which are
    not allowed to occur during a transaction."""

    @wraps(pfs_method)
    def wrapper(stub: "ApiStub", *args, **kwargs):
        if stub.within_transaction:
            raise InvalidTransactionOperation()
        return pfs_method(stub, *args, **kwargs)

    return wrapper


class ClosedCommit(Commit):
    """A ClosedCommit is an extension of the pfs.Commit message with some
    helpful methods. Cannot write to a closed commit.
    """

    def __init__(self, commit: "Commit", stub: "ApiStub"):
        """Internal Use: Do not create this object yourself.

        Parameters
        ----------
        commit : pfs.Commit
            The commit.
        stub : pfs.ApiStub
            The API class to route requests though.
        """
        self._commit = commit
        self._stub = stub

        # This is required to maintain serialization capabilities while being
        #   future compatible with any new fields to the pfs.Commit message.
        super().__init__(
            **{field.name: getattr(commit, field.name) for field in fields(commit)}
        )

    def wait(self) -> "CommitInfo":
        """Waits until the commit is finished being created.

        See OpenCommit docstring for an example.
        """
        return self._stub.wait_commit(self)

    def wait_all(self) -> List["CommitInfo"]:
        """Similar to Commit.wait but streams back the pfs.CommitInfo
        from all the downstream jobs that were initiated by this commit.
        """
        return self._stub.wait_commit_set(CommitSet(id=self._commit.id))


class OpenCommit(ClosedCommit):
    """An OpenCommit is an extension of the pfs.Commit message with some
    helpful methods that provide a more intuitive UX when writing to a commit.

    Examples
    --------
    >>> from pachyderm_sdk import Client
    >>> from pachyderm_sdk.api import pfs
    >>> client: Client
    >>> with client.pfs.commit(branch=pfs.Branch.from_uri("data@master")) as commit:
    >>>     commit.put_file_from_bytes("/greeting.txt", b"Hello!")
    >>>     commit.delete_file("/rude/insult.txt")
    >>> commit.wait()
    """

    def __init__(self, commit: "Commit", stub: "ApiStub"):
        """Internal Use: Do not create this object yourself.

        Parameters
        ----------
        commit : pfs.Commit
            The "open" commit to write to.
        stub : pfs.ApiStub
            The API class to route requests though.
        """
        self._commit = commit
        self._stub = stub
        super().__init__(commit, stub)

    def _close(self):
        """Transform an OpenCommit into a ClosedCommit."""
        self.__class__ = ClosedCommit

    def put_files(self, *, source: Union[Path, str], path: str) -> None:
        """Recursively insert the contents of source into the open commit under path,
        matching the directory structure of source.

        This is roughly equivalent to ``pachctl put file -r``

        Parameters
        ----------
        source : Union[Path, str]
            The directory to recursively insert content from.
        path : str
            The destination path in PFS.

        Examples
        --------
        >>> from pachyderm_sdk import Client
        >>> from pachyderm_sdk.api import pfs
        >>> client: Client
        >>> branch = pfs.Branch.from_uri("images@master")
        >>> with client.pfs.commit(branch=branch) as commit:
        >>>     commit.put_files(source="path/to/local/files", path="/")
        """
        self._stub.put_files(commit=self._commit, source=source, path=path)

    def put_file_from_bytes(self, path: str, data: bytes, append: bool = False) -> "File":
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

        Raises
        ------
        ValueError: If the commit is closed.

        Returns
        -------
        A pfs.File object that that points to the uploaded file.
        NOTE: The commit must be closed before you can read this file.

        Examples
        --------
        >>> from pachyderm_sdk import Client
        >>> from pachyderm_sdk.api import pfs
        >>> client: Client
        >>> with client.pfs.commit(branch=pfs.Branch.from_uri("images@master")) as commit:
        >>>     commit.put_file_from_bytes(path="/file.txt", data=b"SOME BYTES")
        """
        self._stub.put_file_from_bytes(commit=self, path=path, data=data, append=append)
        return File(commit=self._commit, path=path)

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

        Raises
        ------
        ValueError: If the commit is closed.

        Returns
        -------
        A pfs.File object that that points to the uploaded file.
        NOTE: The commit must be closed before you can read this file.

        Examples
        --------
        >>> from pachyderm_sdk import Client
        >>> from pachyderm_sdk.api import pfs
        >>> client: Client
        >>> with client.pfs.commit(branch=pfs.Branch.from_uri("images@master")) as commit:
        >>>     commit.put_file_from_url(
        >>>         path="/index.html", url="https://www.pachyderm.com/index.html"
        >>>     )
        """
        self._stub.put_file_from_url(commit=self, path=path, url=url, recursive=recursive)
        return File(commit=self._commit, path=path)

    def put_file_from_file(
        self, *, path: str, file: "SupportsRead[bytes]", append: bool = False
    ) -> "File":
        """Uploads a PFS file from an open file object.

        Parameters
        ----------
        path : str
            The path in the repo the data will be written to.
        file : SupportsRead[bytes]
            An open file object to read the data from.
        append : bool, optional
            If true, appends the data to the file specified at `path`, if
            they already exist. Otherwise, overwrites them.

        Raises
        ------
        ValueError: If the commit is closed.

        Returns
        -------
        A pfs.File object that that points to the uploaded file.
        NOTE: The commit must be closed before you can read this file.

        Examples
        --------
        >>> from pachyderm_sdk import Client
        >>> from pachyderm_sdk.api import pfs
        >>> client: Client
        >>> with client.pfs.commit(branch=pfs.Branch.from_uri("images@master")) as commit:
        >>>     with open("local_file.dat", "rb") as source:
        >>>         commit.put_file_from_file(path="/index.html", file=source)
        """
        self._stub.put_file_from_file(commit=self, path=path, file=file, append=append)
        return File(commit=self._commit, path=path)

    def copy_file(self, *, src: "File", dst: str, append: bool = False) -> "File":
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

        Raises
        ------
        ValueError: If the commit is closed.

        Returns
        -------
        A pfs.File object that that points to the new file.
        NOTE: The commit must be closed before you can read this file.

        Examples
        --------
        >>> from pachyderm_sdk import Client
        >>> from pachyderm_sdk.api import pfs
        >>> client: Client
        >>> source = pfs.File.from_uri("images@master:/file.dat")
        >>> with client.pfs.commit(branch=pfs.Branch.from_uri("images@master")) as commit:
        >>>     commit.copy_file(src=source, dst="/copy.dat")
        """
        self._stub.copy_file(commit=self, src=src, dst=dst, append=append)
        return File(commit=self._commit, path=dst)

    def delete_file(self, *, path: str) -> "File":
        """Copies a file within PFS

        Parameters
        ----------
        path : str
            The path of the file to be deleted.

        Raises
        ------
        ValueError: If the commit is closed.

        Returns
        -------
        A pfs.File object that that points to the deleted file.

        Examples
        --------
        >>> from pachyderm_sdk import Client
        >>> from pachyderm_sdk.api import pfs
        >>> client: Client
        >>> with client.pfs.commit(branch=pfs.Branch.from_uri("images@master")) as commit:
        >>>     commit.delete_file(path="/file.dat")
        """
        self._stub.delete_file(commit=self, path=path)
        return File(commit=self._commit, path=path)


class ApiStub(_GeneratedApiStub):
    """An extension to the API stub generated from the PFS protobufs."""

    def __init__(
        self,
        channel: grpc.Channel,
        *,
        get_transaction_id: Callable[[], str],
    ):
        self._get_transaction_id = get_transaction_id
        super().__init__(channel=channel)

    @property
    def within_transaction(self) -> bool:
        """For internal use.

        Whether the client is currently within a transaction.
        """
        return bool(self._get_transaction_id())

    @contextmanager
    def commit(
        self, *, parent: "Commit" = None, description: str = "", branch: "Branch" = None
    ) -> ContextManager["OpenCommit"]:
        """A context manager for running operations within a commit.

        When inside this context, the returned object is an OpenCommit which accepts
          write-operations. Upon exiting the context, the commit is closed and the
          OpenCommit becomes a ClosedCommit and no longer allowing write-operations.

        Parameters
        ----------
        parent : pfs.Commit, optional
            The parent commit of the new commit. parent may be empty in which case
            the commit that Branch points to will be used as the parent.
            If the branch does not exist, the commit will have no parent.
        description : str, optional
            A description of the commit.
        branch : pfs.Branch, optional
            The branch where the commit is created.

        Yields
        -------
        pfs.Commit
            A protobuf object that represents a commit.

        Examples
        --------
        >>> from pachyderm_sdk import Client
        >>> from pachyderm_sdk.api import pfs
        >>> client: Client
        >>> with client.pfs.commit(branch=pfs.Branch.from_uri("images@master")) as c:
        >>>     c.delete_file(c, "/dir/delete_me.txt")
        >>>     c.put_file_from_bytes(c, "/new_file.txt", b"DATA")
        """
        commit = self.start_commit(parent=parent, description=description, branch=branch)
        commit_obj = OpenCommit(commit=commit, stub=self)
        try:
            yield commit_obj
        finally:
            commit_obj._close()
            self.finish_commit(commit=commit)

    def wait_commit(self, commit: "Commit") -> "CommitInfo":
        """Waits until the commit is finished being created."""
        return self.inspect_commit(commit=commit, wait=CommitState.FINISHED)

    def wait_commit_set(self, commit_set: "CommitSet") -> List["CommitInfo"]:
        """Similar to client.pfs.wait_commit but streams back the pfs.CommitInfo
        from all the downstream jobs that were initiated by this commit.
        """
        return list(self.inspect_commit_set(commit_set=commit_set, wait=True))

    @transaction_incompatible
    def put_files(self, *, commit: "Commit", source: Union[Path, str], path: str) -> None:
        """Recursively insert the contents of source into the open commit under path,
        matching the directory structure of source.

        This is roughly equivalent to ``pachctl put file -r``

        Note: This method opens multiple gRPC streams and this appears to break in
          some REPL environment (such as those based in IDEs). If you encounter this
          problem, please try a different REPL environment or run as a script.

        Parameters
        ----------
        commit : pfs.Commit
            The open commit to add files to.
        source : Union[Path, str]
            The directory to recursively insert content from.
        path : str
            The destination path in PFS.

        Examples
        --------
        >>> from pachyderm_sdk import Client
        >>> from pachyderm_sdk.api import pfs
        >>> client: Client
        >>> with client.pfs.commit(branch=pfs.Branch.from_uri("images@master")) as c:
        >>>     client.pfs.put_files(commit=c, source="path/to/local/files", path="/")
        """
        source = Path(source)
        if not source.exists():
            raise FileNotFoundError(f"source does not exist: {source}")
        if not source.is_dir():
            raise NotADirectoryError(f"source is not a directory: {source}")
        for root, _, filenames in os.walk(source):
            for filename in filenames:
                src = os.path.join(root, filename)
                dst = os.path.join(path, os.path.relpath(src, start=source))
                with open(src, "rb") as file:
                    self.put_file_from_file(commit=commit, path=dst, file=file)

    @transaction_incompatible
    def put_file_from_bytes(
        self, *, commit: "Commit", path: str, data: bytes, append: bool = False
    ) -> Empty:
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
        >>> from pachyderm_sdk import Client
        >>> from pachyderm_sdk.api import pfs
        >>> client: Client
        >>> with client.pfs.commit(branch=pfs.Branch.from_uri("images@master")) as c:
        >>>     client.pfs.put_file_from_bytes(
        >>>         commit=c, path="/file.txt", data=b"SOME BYTES"
        >>>     )
        """
        return self.put_file_from_file(
            commit=commit, path=path, file=io.BytesIO(data), append=append
        )

    @transaction_incompatible
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

        Examples
        --------
        >>> from pachyderm_sdk import Client
        >>> from pachyderm_sdk.api import pfs
        >>> client: Client
        >>> with client.pfs.commit(branch=pfs.Branch.from_uri("images@master")) as c:
        >>>     client.pfs.put_file_from_url(
        >>>         commit=c, path="/index.html", url="www.pachyderm.com/index.html"
        >>>     )
        """
        operations = [
            ModifyFileRequest(set_commit=commit),
            ModifyFileRequest(delete_file=DeleteFile(path=path)),
            ModifyFileRequest(
                add_file=AddFile(
                    path=path, url=AddFileUrlSource(url=url, recursive=recursive)
                )
            ),
        ]
        return self.modify_file(iter(operations))

    @transaction_incompatible
    def put_file_from_file(
        self,
        *,
        commit: "Commit",
        path: str,
        file: "SupportsRead[bytes]",
        append: bool = False,
    ) -> Empty:
        """Uploads a PFS file from an open file object.

        Parameters
        ----------
        commit : pfs.Commit
            An open commit to modify.
        path : str
            The path in the repo the data will be written to.
        file : SupportsRead[bytes]
            An open file object to read the data from.
        append : bool, optional
            If true, appends the data to the file specified at `path`, if
            they already exist. Otherwise, overwrites them.

        Examples
        --------
        >>> from pachyderm_sdk import Client
        >>> from pachyderm_sdk.api import pfs
        >>> client: Client
        >>> with client.pfs.commit(branch=pfs.Branch.from_uri("images@master")) as c:
        >>>     with open("local_file.dat", "rb") as source:
        >>>         client.pfs.put_file_from_file(
        >>>             commit=c, path="/index.html", file=source
        >>>         )
        """
        check = file.read(BUFFER_SIZE)
        if len(check) > 0:
            if not isinstance(check, bytes):
                raise TypeError("File must output bytes")

        def file_iterator():
            if not check:
                return
            yield check
            while True:
                data = file.read(BUFFER_SIZE)
                if len(data) == 0:
                    return
                yield data

        def operations() -> Iterable[ModifyFileRequest]:
            yield ModifyFileRequest(set_commit=commit)
            if not append:
                yield ModifyFileRequest(delete_file=DeleteFile(path=path))
            yield ModifyFileRequest(add_file=AddFile(path=path, raw=b""))
            for data in file_iterator():
                yield ModifyFileRequest(add_file=AddFile(path=path, raw=data))

        return self.modify_file(operations())

    @transaction_incompatible
    def copy_file(
        self, *, commit: "Commit", src: "File", dst: str, append: bool = False
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

        Examples
        --------
        >>> from pachyderm_sdk import Client
        >>> from pachyderm_sdk.api import pfs
        >>> client: Client
        >>> source = pfs.File.from_uri("images@master:/file.dat")
        >>> with client.pfs.commit(branch=pfs.Branch.from_uri("images@master")) as c:
        >>>     commit.pfs.copy_file(commit=c, src=source, dst="/copy.dat")
        """
        operations = [
            ModifyFileRequest(set_commit=commit),
            ModifyFileRequest(copy_file=CopyFile(dst=dst, src=src, append=append)),
        ]
        return self.modify_file(iter(operations))

    @transaction_incompatible
    def delete_file(self, *, commit: "Commit", path: str) -> Empty:
        """Copies a file within PFS

        Parameters
        ----------
        commit : pfs.Commit
            An open commit to modify.
        path : str
            The path of the file to be deleted.

        Examples
        --------
        >>> from pachyderm_sdk import Client
        >>> from pachyderm_sdk.api import pfs
        >>> client: Client
        >>> with client.pfs.commit(branch=pfs.Branch.from_uri("images@master")) as c:
        >>>     commit.pfs.delete_file(commit=c, path="/file.dat")
        """
        operations = [
            ModifyFileRequest(set_commit=commit),
            ModifyFileRequest(delete_file=DeleteFile(path=path)),
        ]
        return self.modify_file(iter(operations))

    def project_exists(self, project: "Project") -> bool:
        """Checks whether a project exists.

        Parameters
        ----------
        project: pfs.Project
            The project to check.

        Returns
        -------
        bool
            Whether the project exists.
        """
        try:
            self.inspect_project(project=project)
            return True
        except grpc.RpcError as err:
            err: grpc.Call
            if err.code() == grpc.StatusCode.NOT_FOUND:
                return False
            raise err

    def repo_exists(self, repo: "Repo") -> bool:
        """Checks whether a repo exists.

        Parameters
        ----------
        repo: pfs.Repo
            The repo to check.

        Returns
        -------
        bool
            Whether the repo exists.
        """
        try:
            self.inspect_repo(repo=repo)
            return True
        except grpc.RpcError as err:
            err: grpc.Call
            if err.code() == grpc.StatusCode.NOT_FOUND:
                return False
            raise err

    def branch_exists(self, branch: "Branch") -> bool:
        """Checks whether a branch exists.

        Parameters
        ----------
        branch: pfs.Branch
            The branch to check.

        Returns
        -------
        bool
            Whether the branch exists.
        """
        try:
            self.inspect_branch(branch=branch)
            return True
        except grpc.RpcError as err:
            err: grpc.Call
            if err.code() == grpc.StatusCode.NOT_FOUND:
                return False
            raise err

    def commit_exists(self, commit: "Commit") -> bool:
        """Checks whether a commit exists.

        Parameters
        ----------
        commit: pfs.Commit
            The commit to check.

        Returns
        -------
        bool
            Whether the commit exists.
        """
        try:
            self.inspect_commit(commit=commit)
            return True
        except grpc.RpcError as err:
            err: grpc.Call
            if err.code() == grpc.StatusCode.NOT_FOUND:
                return False
            raise err

    def path_exists(self, file: "File") -> bool:
        """Checks whether the path exists in the specified commit, agnostic to
        whether `path` is a file or a directory.

        Parameters
        ----------
        file : pfs.File
            The file (or directory) to check.

        Raises
        ------
        ValueError: If commit does not exist.

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
            return True
        except grpc.RpcError as err:
            err: grpc.Call
            if err.code() == grpc.StatusCode.NOT_FOUND:
                return False
            raise err

    def pfs_file(self, file: "File") -> "PFSFile":
        """Wraps the response stream of a client.pfs.get_file() call with a
        PFSFile object. This wrapper class allows you to interact with the
        file stream as a normal file object.

        Parameters
        ----------
        file : pfs.File
            The file to retrieve.

        Examples
        --------
        >>> from pachyderm_sdk import Client
        >>> from pachyderm_sdk.api import pfs
        >>> client: Client
        >>> source = pfs.File.from_uri("images@master:/example.csv")
        >>> with client.pfs.pfs_file(file=source) as pfs_file:
        >>>     for line in pfs_file:
        >>>         print(line)
        """
        stream = self.get_file(file=file)
        return PFSFile(stream)

    def pfs_tar_file(self, file: "File") -> "PFSTarFile":
        """Wraps the response stream of a client.pfs.get_tar_file() call with a
        PFSTarFile object. This wrapper class allows you to interact with the
        file stream as a standard tarfile.TarFile object.

        Parameters
        ----------
        file : pfs.File
            The file (or directory) to retrieve.
        """
        stream = self.get_file_tar(file=file)
        return PFSTarFile.open(fileobj=PFSFile(stream), mode="r|*")
