"""
Implementation of a gRPC interceptor used to set request metadata
and catch connection errors.
"""

from os import environ
from typing import Any, Callable, Sequence, Optional, Tuple, Union

import grpc
from grpc_interceptor import ClientCallDetails, ClientInterceptor

from .errors import AuthServiceNotActivated

MetadataType = Sequence[Tuple[str, Union[str, bytes]]]


class MetadataClientInterceptor(ClientInterceptor):
    def __init__(self, metadata: MetadataType):
        self.metadata = metadata

    def intercept(self, method: Callable, request: Any, call_details: ClientCallDetails):
        call_details_metadata = list(call_details.metadata or [])
        call_details_metadata.extend(self.metadata)
        new_details = call_details._replace(metadata=call_details_metadata)
        future = method(request, new_details)
        future.add_done_callback(_check_connection_error)
        return future


def _check_connection_error(grpc_future: grpc.Future):
    """Callback function that checks if a gRPC.Future experienced a
    ConnectionError and attempt to sanitize the error message for the user.
    """
    error: Optional[grpc.Call] = grpc_future.exception()
    if error is not None:
        unable_to_connect = "failed to connect to all addresses" in error.details()
        if error.code() == grpc.StatusCode.UNAVAILABLE and unable_to_connect:
            error_message = "Could not connect to pachyderm instance\n"
            if "PACHD_PEER_SERVICE_HOST" in environ:
                error_message += (
                    "\tPACHD_PEER_SERVICE_HOST is detected. "
                    "Please use Client.new_in_cluster() when using"
                    " python_pachyderm within the pipeline. "
                )
            raise ConnectionError(error_message) from error
        raise AuthServiceNotActivated.try_from(error)
