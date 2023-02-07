import os
import tarfile
from typing import Iterator

import grpc
from betterproto import BytesValue


class PFSTarFile(tarfile.TarFile):
    def __iter__(self):
        for tarinfo in super().__iter__():
            if os.path.isabs(tarinfo.path):
                # Hack to prevent extraction to absolute paths.
                tarinfo.path = tarinfo.path[1:]
            if tarinfo.mode == 0:
                # Hack to prevent writing files with no permissions.
                tarinfo.mode = 0o700
            yield tarinfo


class PFSFile:
    """File-like objects containing content of a file stored in PFS.

    Examples
    --------
    >>> # client.get_file() returns a PFSFile
    >>> source_file = client.get_file(("montage", "master"), "/montage.png")
    >>> with open("montage.png", "wb") as dest_file:
    >>>     shutil.copyfileobj(source_file, dest_file)
    ...
    >>> with client.get_file(("montage", "master"), "/montage2.png") as f:
    >>>     content = f.read()
    """

    def __init__(self, stream: Iterator[BytesValue]):
        self._stream = stream
        self._buffer = bytearray()

        try:
            first_message = next(self._stream)
        except grpc.RpcError as err:
            raise ConnectionError("Error creating the PFSFile") from err
        self._buffer.extend(first_message.value)

    def __enter__(self):
        return self

    def __exit__(self, type, val, tb):
        self.close()

    def read(self, size: int = -1) -> bytes:
        """Reads from the :class:`.PFSFile` buffer.

        Parameters
        ----------
        size : int, optional
            If set, the number of bytes to read from the buffer.

        Returns
        -------
        bytes
            Content from the stream.
        """
        try:
            if size == -1:
                # Consume the entire stream.
                for message in self._stream:
                    self._buffer.extend(message.value)
                result, self._buffer[:] = self._buffer[:], b""
                return bytes(result)
            elif len(self._buffer) < size:
                for message in self._stream:
                    self._buffer.extend(message.value)
                    if len(self._buffer) >= size:
                        break
        except grpc.RpcError:
            pass

        size = min(size, len(self._buffer))
        result, self._buffer[:size] = self._buffer[:size], b""
        return bytes(result)

    def close(self) -> None:
        """Closes the :class:`.PFSFile`."""
        self._stream.cancel()
