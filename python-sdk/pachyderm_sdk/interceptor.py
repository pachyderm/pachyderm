"""
Implementation of a gRPC interceptor used to set request metadata
and catch connection errors.
"""

from os import environ
from typing import Callable, Sequence, Optional, Tuple, Union

import grpc
from betterproto import Message
from grpc_interceptor import ClientCallDetails, ClientInterceptor

from .errors import AuthServiceNotActivated

MetadataType = Sequence[Tuple[str, Union[str, bytes]]]


class MetadataClientInterceptor(ClientInterceptor):
    def __init__(self, metadata: MetadataType):
        self.metadata = metadata

    def intercept(
        self, method: Callable, request: Message, call_details: ClientCallDetails
    ):
        call_details_metadata = list(call_details.metadata or [])
        call_details_metadata.extend(self.metadata)
        new_details = call_details._replace(metadata=call_details_metadata)

        # Error handling happens different depending on whether the response
        #   is Unary or Stream. If the response is a stream, then the error
        #   will be raised before the "done_callback" is called and will
        #   therefore be caught by the try-except block. If the response is
        #   Unary then the error "done_callback" will be called before the
        #   error is raised.
        try:
            future = method(request, new_details)
        except grpc.RpcError as error:
            # gRPC error types are confusing - instantiated errors are also grpc.Call objects.
            # ref: github.com/grpc/grpc/issues/25334#issuecomment-772730080
            error: grpc.Call
            _check_errors(error, request)
        else:
            future.add_done_callback(lambda f: _check_errors(f.exception(), request))
            return future


def _check_errors(error: Optional[grpc.Call], request: Message):
    """Callback function that handles and sanitizes specific errors (if they occur)
    to better communicate these error states to the user.
    """
    if error is not None:
        code, details = error.code(), error.details()
        unable_to_connect = "failed to connect to all addresses" in details
        if code == grpc.StatusCode.UNAVAILABLE and unable_to_connect:
            error_message = "Could not connect to pachyderm instance\n"
            if "PACHD_PEER_SERVICE_HOST" in environ:
                error_message += (
                    "\tPACHD_PEER_SERVICE_HOST is detected. "
                    "Please use Client.new_in_cluster() when using"
                    " python_pachyderm within the pipeline. "
                )
            raise ConnectionError(error_message) from error

        unable_to_serialize = "Exception serializing request" in details
        if code == grpc.StatusCode.INTERNAL and unable_to_serialize:
            error_message = (
                "An error occurred while trying to serialize the following"
                f" {request.__class__.__qualname__} message.\n "
                " This is most likely due to one of the fields of"
                " this message having a value with an incorrect type.\n"
                f"\tMessage: {request}"
            )
            raise TypeError(error_message) from error

        auth_codes = (grpc.StatusCode.UNIMPLEMENTED, grpc.StatusCode.UNAUTHENTICATED)
        auth_not_activated = "the auth service is not activated" in details
        if code in auth_codes and auth_not_activated:
            raise AuthServiceNotActivated(details)

        raise error
