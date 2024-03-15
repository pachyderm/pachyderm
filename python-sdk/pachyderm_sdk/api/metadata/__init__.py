# Generated by the protocol buffer compiler.  DO NOT EDIT!
# sources: api/metadata/metadata.proto
# plugin: python-betterproto
# This file has been @generated
from dataclasses import dataclass
from typing import (
    TYPE_CHECKING,
    Dict,
    Optional,
)

import betterproto
import betterproto.lib.google.protobuf as betterproto_lib_google_protobuf
import grpc


if TYPE_CHECKING:
    import grpc


@dataclass(eq=False, repr=False)
class EditMetadataRequest(betterproto.Message):
    pass


@dataclass(eq=False, repr=False)
class EditMetadataResponse(betterproto.Message):
    pass


class ApiStub:

    def __init__(self, channel: "grpc.Channel"):
        self.__rpc_edit_metadata = channel.unary_unary(
            "/metadata.API/EditMetadata",
            request_serializer=EditMetadataRequest.SerializeToString,
            response_deserializer=EditMetadataResponse.FromString,
        )

    def edit_metadata(self) -> "EditMetadataResponse":

        request = EditMetadataRequest()

        return self.__rpc_edit_metadata(request)
