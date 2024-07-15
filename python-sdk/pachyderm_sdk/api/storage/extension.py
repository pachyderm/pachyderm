"""Handwritten classes/methods that augment the existing Storage API."""

import os
from pathlib import Path
from typing import Union

from betterproto import which_one_of

from ..cdr.resolver import (
    CdrResolver,
    Ref,
    Cipher,
    Compress,
    Concat,
    ContentHash,
    Http,
    SizeLimits,
    Slice,
)
from . import FilesetStub as _GeneratedApiStub
from . import FileFilter, PathRange
from pachyderm_sdk.constants import CDR_CACHE_LOCATION


class ApiStub(_GeneratedApiStub):

    def assemble_fileset(
        self,
        fileset: str,
        path: Union[str, PathRange],
        destination: os.PathLike,
        *,
        cache_location: os.PathLike = CDR_CACHE_LOCATION,
        fetch_missing_chunks: bool = True,
        http_host_replacement: str = "",
    ) -> None:
        """Assemble the data contained at the specified path within the fileset
        and write that data to the specified destination directory. This method
        uses a local cache and will automatically fetch missing data from PFS,
        unless disabled.

        Parameters
        ----------
        fileset : str
            The UUID of the fileset.
        path : typing.Union[str, PathRange]
            The path, regex, or PathRange within the fileset to assemble.
            A value of "/" assembles the entire fileset.
        destination : os.PathLike
            The local directory at which to assemble the fileset.
        cache_location : os.PathLike (kwarg only)
            The location of the chunk cache. This is also configurable thru ENV VAR.
        fetch_missing_chunks : bool (kwarg only)
            If true, fetch any missing chunks needed to assemble the fileset.
            If false, raise an error if any chunks are missing.
        http_host_replacement : str (kwarg only)
            The value of this parameter replaces the host (including port) within
            the presigned URLs when resolving CDRs. This configuration is useful
            if, for some reason, the URL that pachd uses to interact with object
            storage is different from the one that you use. For example, if your
            Pachyderm object storage is running within your kubernetes cluster and
            pachd is configured to use a URL which is only valid in cluster.
            This may fix any issues, depending on the object storage and how they
            generate presigned URLs.
            Example: localhost:9000

        Raises
        ------
            FileNotFoundError:
                if chunk is missing from cache and fetch_missing_chunks=False.
        """
        destination = Path(os.path.expanduser(destination)).resolve()
        destination.mkdir(parents=True, exist_ok=True)
        resolver = CdrResolver(
            cache_location=cache_location,
            fetch_missing_chunks=fetch_missing_chunks,
            http_host_replacement=http_host_replacement,
        )

        for msg in self.read_fileset_cdr(fileset_id=fileset, filters=[as_filter(path)]):
            content = resolver.resolve(msg.ref)

            # File paths returned from PFS are absolute.
            # By removing the leading "/", we convert the path from absolute
            #   to relative from the repo root.
            file_path = msg.path.removeprefix("/")
            destination_file = destination.joinpath(file_path)
            destination_file.parent.mkdir(parents=True, exist_ok=True)
            destination_file.write_bytes(content)

    def fetch_chunks(
        self,
        fileset: str,
        path: Union[str, PathRange],
        *,
        cache_location: os.PathLike = CDR_CACHE_LOCATION,
        http_host_replacement: str = "",
        prune: bool = False,
    ) -> None:
        """Fetch all of the PFS chunks required to assemble the data contained
        at the specified path within the fileset and write them to the cache
        location.

        Parameters
        ----------
        fileset : str
            The UUID of the fileset.
        path : typing.Union[str, PathRange]
            A path, regex, or PathRange within the fileset.
            A value of "/" fetches the entire fileset.
        cache_location : os.PathLike (kwarg only)
            The location of the chunk cache. This is also configurable thru ENV VAR.
        http_host_replacement : str (kwarg only)
            The value of this parameter replaces the host (including port) within
            the presigned URLs when resolving CDRs. This configuration is useful
            if, for some reason, the URL that pachd uses to interact with object
            storage is different from the one that you use. For example, if your
            Pachyderm object storage is running within your kubernetes cluster and
            pachd is configured to use a URL which is only valid in cluster.
            This may fix any issues, depending on the object storage and how they
            generate presigned URLs.
            Example: localhost:9000
        prune : bool (kwarg only)
            If true, prune (delete) any chunks that are not required to assemble the fileset.
        """
        resolver = CdrResolver(
            cache_location=cache_location, http_host_replacement=http_host_replacement
        )

        unused_chunks = set()
        if prune:
            unused_chunks = {file.name for file in resolver.cache.iterdir()}

        def cache_chunks(ref: Ref) -> None:
            """This is an optimization. We could simply call resolver.resolve() and
            discard the output, but that would preform additional computation for
            decompressing, decrypting, assembling, etc."""
            field, body = which_one_of(ref, "body")
            if isinstance(body, Http):
                raise ValueError(
                    "malformed CDR: no ContentHash message contained within the CDR"
                )
            elif isinstance(body, ContentHash):
                unused_chunks.discard(resolver._chunk_name(body))
                resolver._dereference_content_hash(body)
                return
            elif isinstance(body, (Cipher, Compress, SizeLimits, Slice)):
                return cache_chunks(body.inner)
            elif isinstance(body, Concat):
                for inner_ref in body.refs:
                    cache_chunks(inner_ref)
            else:
                raise ValueError(f"unsupported Ref variant: {body}")

        for msg in self.read_fileset_cdr(fileset_id=fileset, filters=[as_filter(path)]):
            cache_chunks(msg.ref)

        for file_name in unused_chunks:
            file = resolver.cache.joinpath(file_name)
            if file.is_file():
                file.unlink(missing_ok=True)


def as_filter(path: Union[str, PathRange]) -> FileFilter:
    file_filter = FileFilter()
    if isinstance(path, PathRange):
        file_filter.path_range = path
    else:
        file_filter.path_regex = path
    return file_filter
