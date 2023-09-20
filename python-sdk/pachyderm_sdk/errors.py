""" Errors that can be raised by this library. """
from typing import Union

from grpc import RpcError
from grpc._channel import _InactiveRpcError


class AuthServiceNotActivated(ConnectionError):
    """If the auth service is not activated but is required."""

    @classmethod
    def try_from(cls, error: RpcError) -> Union["AuthServiceNotActivated", RpcError]:
        if isinstance(error, _InactiveRpcError):
            details = error.details()
            if "the auth service is not activated" in details:
                return cls(details)
        return error


class ConfigError(Exception):
    """Error for issues related to the pachyderm config file."""

    def __init__(self, message):
        super().__init__(message)


class InvalidTransactionOperation(RuntimeError):
    """Error triggered when an invalid operation (i.e. file write)
    is called when inside a transaction.
    """

    def __init__(self):
        super().__init__(
            "File operations are not permitted within a pachyderm transaction."
        )
