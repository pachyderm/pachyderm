from contextlib import contextmanager
from typing import Iterator, Callable

import grpc

from . import ApiStub as _GeneratedApiStub
from . import (
    Transaction,
    TransactionInfo,
)


class ApiStub(_GeneratedApiStub):

    def __init__(
        self,
        channel: grpc.Channel,
        *,
        get_transaction_id: Callable[[], str],
        set_transaction_id: Callable[[str], None],
    ):
        self._get_transaction_id = get_transaction_id
        self._set_transaction_id = set_transaction_id
        super().__init__(channel=channel)

    def start_transaction(self) -> "Transaction":
        # TODO: Should we do this?
        response = super().start_transaction()
        self._set_transaction_id(response.id)
        return response

    def finish_transaction(
        self, *, transaction: "Transaction" = None
    ) -> "TransactionInfo":
        # TODO: Should we do this?
        response = super().finish_transaction(transaction=transaction)
        self._set_transaction_id("")
        return response

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
        old_transaction_id = self._get_transaction_id()
        transaction = super().start_transaction()
        self._set_transaction_id(transaction.id)

        try:
            yield transaction
        except Exception:
            super().delete_transaction(transaction=transaction)
            raise
        else:
            super().finish_transaction(transaction=transaction)
        finally:
            self._set_transaction_id(old_transaction_id)
