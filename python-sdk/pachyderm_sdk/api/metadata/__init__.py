# Generated by the protocol buffer compiler.  DO NOT EDIT!
# sources: api/metadata/metadata.proto
# plugin: python-betterproto
# This file has been @generated
from dataclasses import dataclass
from typing import (
    TYPE_CHECKING,
    Dict,
    List,
    Optional,
)

import betterproto
import betterproto.lib.google.protobuf as betterproto_lib_google_protobuf
import grpc

from .. import pfs as _pfs__


if TYPE_CHECKING:
    import grpc


@dataclass(eq=False, repr=False)
class Edit(betterproto.Message):
    """Edit represents editing one piece of metadata."""

    project: "_pfs__.ProjectPicker" = betterproto.message_field(1, group="target")
    """project targets a named project's metadata."""

    commit: "_pfs__.CommitPicker" = betterproto.message_field(2, group="target")
    """commit targets a commit's metadata."""

    branch: "_pfs__.BranchPicker" = betterproto.message_field(3, group="target")
    """branch targets a branch's metadata."""

    repo: "_pfs__.RepoPicker" = betterproto.message_field(4, group="target")
    """repo targets a repo's metadata."""

    replace: "EditReplace" = betterproto.message_field(10, group="op")
    """replace replaces a target's metadata with a new metadata mapping."""

    add_key: "EditAddKey" = betterproto.message_field(11, group="op")
    """add_key adds a new key to the target object's metadata."""

    edit_key: "EditEditKey" = betterproto.message_field(12, group="op")
    """edit_key adds or changes a key in the target object's metadata."""

    delete_key: "EditDeleteKey" = betterproto.message_field(13, group="op")
    """delete_key removes a key from the target object's metadata."""


@dataclass(eq=False, repr=False)
class EditReplace(betterproto.Message):
    """Replace is an operation that replaces metadata."""

    replacement: Dict[str, str] = betterproto.map_field(
        1, betterproto.TYPE_STRING, betterproto.TYPE_STRING
    )
    """replacement is the map to replace the object's metadata with."""


@dataclass(eq=False, repr=False)
class EditAddKey(betterproto.Message):
    """AddKey is an operation that adds a new metadata key."""

    key: str = betterproto.string_field(1)
    """key is the metadata key to add.  It may not be the empty string."""

    value: str = betterproto.string_field(2)
    """value is the value to assign to the metadata key."""


@dataclass(eq=False, repr=False)
class EditEditKey(betterproto.Message):
    """EditKey is an operation that changes or adds a metadata key."""

    key: str = betterproto.string_field(1)
    """
    key is the metadata key to change or add.  It may not be the empty string.
    """

    value: str = betterproto.string_field(2)
    """value is the value to assign to the metadata key."""


@dataclass(eq=False, repr=False)
class EditDeleteKey(betterproto.Message):
    """DeleteKey is an operation that removes a metadata key."""

    key: str = betterproto.string_field(1)
    """key is the metadata key to remove.  It may not be the empty string."""


@dataclass(eq=False, repr=False)
class EditMetadataRequest(betterproto.Message):
    """EditMetadataRequest is a sequence of edits to apply to metadata."""

    edits: List["Edit"] = betterproto.message_field(1)
    """edits is the ordered list of metadata edits to perform."""


@dataclass(eq=False, repr=False)
class EditMetadataResponse(betterproto.Message):
    """EditMetadataResponse is the result of editing metadata."""

    pass


class ApiStub:

    def __init__(self, channel: "grpc.Channel"):
        self.__rpc_edit_metadata = channel.unary_unary(
            "/metadata.API/EditMetadata",
            request_serializer=EditMetadataRequest.SerializeToString,
            response_deserializer=EditMetadataResponse.FromString,
        )

    def edit_metadata(
        self, *, edits: Optional[List["Edit"]] = None
    ) -> "EditMetadataResponse":
        edits = edits or []

        request = EditMetadataRequest()
        if edits is not None:
            request.edits = edits

        return self.__rpc_edit_metadata(request)