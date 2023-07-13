"""Handwritten classes/methods that augment the existing Transaction API."""
from contextlib import contextmanager
from typing import Callable, ContextManager

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
        """Starts a transaction and sets the transaction ID within the client.
        This will make all subsequent resource operations performed by the client
        occur within the returned transaction, until the transaction is finished.
        """
        response = super().start_transaction()
        self._set_transaction_id(response.id)
        return response

    def finish_transaction(self, *, transaction: "Transaction" = None) -> "TransactionInfo":
        """Finishes a transaction and removes the transaction ID within the client."""
        response = super().finish_transaction(transaction=transaction)
        self._set_transaction_id("")
        return response

    @contextmanager
    def transaction(self) -> ContextManager[Transaction]:
        """A context manager for running operations within a transaction. When
        the context manager completes, the transaction will be deleted if an
        error occurred, or otherwise finished.

        Yields
        -------
        transaction.Transaction
            A protobuf object that represents a transaction.

        Examples
        --------
        If a pipeline has two input repos, `foo` and `bar`, a transaction is
        useful for adding data to both atomically before the pipeline runs
        even once.

        >>> from pachyderm_sdk import Client
        >>> from pachyderm_sdk.api import pfs
        >>> client: Client
        >>> with client.transaction.transaction() as txn:
        >>>     with client.pfs.commit(branch=pfs.Branch.from_uri("foo@master")) as c1:
        >>>         c1.put_file_from_bytes("/joint_data.txt", b"DATA1")
        >>>     with client.pfs.commit(branch=pfs.Branch.from_uri("bar@master")) as c2:
        >>>         c2.put_file_from_bytes("/joint_data.txt", b"DATA2")
        >>> c1.wait()
        >>> c2.wait()
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

    def transaction_exists(self, transaction: "Transaction") -> bool:
        """Checks whether a transaction exists.

        Parameters
        ----------
        transaction: transaction.Transaction
            The transaction to check.

        Returns
        -------
        bool
            Whether the transaction exists.
        """
        try:
            super().inspect_transaction(transaction=transaction)
            return True
        except grpc.RpcError as err:
            err: grpc.Call
            if err.code() == grpc.StatusCode.NOT_FOUND:
                return False
            raise err
