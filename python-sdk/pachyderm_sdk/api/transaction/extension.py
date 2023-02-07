from contextlib import contextmanager
from typing import Iterator

from . import ApiStub as _GeneratedApiStub
from . import (
    Transaction
)

class ApiStub(_GeneratedApiStub):

    @contextmanager
    def transaction(self) -> Iterator[Transaction]:
        """A context manager for running operations within a transaction. When
        the context manager completes, the transaction will be deleted if an
        error occurred, or otherwise finished.

        Yields
        -------
        transaction_pb2.Transaction
            A protobuf object that represents a transaction.

        Examples
        --------
        If a pipeline has two input repos, `foo` and `bar`, a transaction is
        useful for adding data to both atomically before the pipeline runs
        even once.

        >>> with client.transaction() as t:
        >>>     c1 = client.start_commit("foo", "master")
        >>>     c2 = client.start_commit("bar", "master")
        >>>
        >>>     client.put_file_bytes(c1, "/joint_data.txt", b"DATA1")
        >>>     client.put_file_bytes(c2, "/joint_data.txt", b"DATA2")
        >>>
        >>>     client.finish_commit(c1)
        >>>     client.finish_commit(c2)
        """