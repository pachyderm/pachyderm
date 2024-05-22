import os
import tarfile
from io import RawIOBase
from typing import BinaryIO, Iterator, TYPE_CHECKING

if TYPE_CHECKING:
    from _typeshed import WriteableBuffer

import grpc
from betterproto import BytesValue


class PFSTarFile(tarfile.TarFile):
    """Wrapper to allow reading a TAR file from PFS.

    See the tarfile.TarFile class for more information on the supported methods.
    """

    def __iter__(self):
        for tarinfo in super().__iter__():
            if os.path.isabs(tarinfo.path):
                # Hack to prevent extraction to absolute paths.
                tarinfo.path = tarinfo.path[1:]
            if tarinfo.mode == 0:
                # Hack to prevent writing files with no permissions.
                tarinfo.mode = 0o700
            yield tarinfo


class PFSFile(RawIOBase, BinaryIO):
    """File-like objects containing content of a file stored in PFS.

    The data is read from an open gRPC stream, any connection error will close the file.

    Notes
    -----
    The following links wer instrumental to understanding how to implement a pythonic
    file object:
      * https://docs.python.org/3/library/io.html
      * https://github.com/python/cpython/blob/3.11/Lib/_pyio.py

    Examples
    --------
    >>> # client.pfs.pfs_file() returns a PFSFile
    >>> import shutil
    >>> from pachyderm_sdk import Client
    >>> from pachyderm_sdk.api import pfs
    >>>
    >>> client: Client
    >>> file = pfs.File.from_uri("montage@master:/montage.png")
    >>> source_file = client.pfs.pfs_file(file=file)
    >>> with open("montage.png", "wb") as dest_file:
    >>>     shutil.copyfileobj(source_file, dest_file)
    ...
    >>> with client.pfs.pfs_file(file=file) as pfs_file:
    >>>     contents = pfs_file.read()
    """

    def __init__(self, stream: Iterator[BytesValue]):
        self._stream = stream
        self._buffer = bytearray()

        try:
            self.peek()
        except grpc.RpcError as err:
            raise ValueError(
                "Error instantiating PFSFile -- see triggering exception"
            ) from err

    def __enter__(self) -> "PFSFile":
        return self

    def readinto(self, buffer: "WriteableBuffer") -> int:
        """Read bytes into a pre-allocated writeable buffer.

        Returns an int representing the number of bytes read.

        Note: This method is the only method that needs to be implemented in order
          to support the read(), readall(), readline(), etc. methods.
        """
        if self.closed:
            raise ValueError("I/O operation on closed file.")
        size = len(buffer)
        if len(self._buffer) < size:
            try:
                for message in self._stream:
                    self._buffer.extend(message.value)
                    if len(self._buffer) >= size:
                        break
            except grpc.RpcError as err:
                self.close()
                raise err

        size = min(size, len(self._buffer))
        buffer[:size], self._buffer[:size] = self._buffer[:size], b""
        return size

    def peek(self, size: int = 0) -> bytes:
        """Returns bytes from the stream without advancing the read position.
        At most one single read on the raw stream is done to satisfy the call.
        The number of bytes returned may be less than requested."""
        if self.closed:
            raise ValueError("I/O operation on closed file.")
        if len(self._buffer) == 0:
            try:
                message = next(self._stream)
            except grpc.RpcError as err:
                self.close()
                raise err
            except StopIteration:
                return b""
            self._buffer[:] = message.value
        return bytes(self._buffer[:size])

    def close(self) -> None:
        """Closes the PFSFile and cancels the gRPC stream."""
        if hasattr(self, "_stream"):
            del self._stream
        super().close()

    def readable(self) -> bool:
        return not self.closed
