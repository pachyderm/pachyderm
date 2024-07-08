import os
import shutil
from pathlib import Path
from typing import Dict, Tuple

import pytest

from tests.fixtures import *

from pachyderm_sdk.api import cdr, storage
from .utils import count


@pytest.fixture
def fileset(client: TestClient) -> Tuple[str, Dict[str, bytes]]:
    """Creates a fileset.

    Returns
    -------
        (fileset_id, Dict[path, data])
    """
    files = {f"/file{i}": os.urandom(1024) for i in range(10)}
    request_iterator = (
        storage.CreateFilesetRequest(append_file=storage.AppendFile(path=file, data=data))
        for file, data in files.items()
    )
    fileset = client.storage.create_fileset(request_iterator).fileset_id
    return fileset, files


class TestStorage:
    """
    Note:
      As our tests and local development uses in-cluster object storage
      (usually Minio), we need to patch the host of the presigned URL with
      the correct, external host.
    """

    @staticmethod
    def test_fetch_chunks(
        client: TestClient,
        fileset: Tuple[str, Dict[str, bytes]],
        tmp_path: Path,
    ):
        """Test fetching chunks are stored in a local cache."""
        # Arrange
        fileset_id, _ = fileset
        expected_prefix = cdr.HashAlgo.BLAKE2b_256.name.lower()

        # Act
        # This is hardcoded to replace the host of the presigned URL with
        # localhost:9000. This test needs to be run simultaneously with:
        #   - kubectl port-forward service/minio 9000:9000
        client.storage.fetch_chunks(
            fileset_id,
            path="/",
            cache_location=tmp_path,
            http_host_replacement="localhost:9000",
        )

        # Assert
        assert count(tmp_path.iterdir()) > 0
        for file in tmp_path.iterdir():
            assert file.name.startswith(expected_prefix)

    @staticmethod
    def test_fetch_chunks_prune(
        client: TestClient,
        fileset: Tuple[str, Dict[str, bytes]],
        tmp_path: Path,
    ):
        """Test fetching chunks with prune removes unused files."""
        # Arrange
        fileset_id, _ = fileset
        unused_chunk = tmp_path.joinpath("abcdefgh12345678")
        unused_chunk.touch()

        # Act
        # This is hardcoded to replace the host of the presigned URL with
        # localhost:9000. This test needs to be run simultaneously with:
        #   - kubectl port-forward service/minio 9000:9000
        client.storage.fetch_chunks(
            fileset_id,
            path="/",
            cache_location=tmp_path,
            http_host_replacement="localhost:9000",
            prune=True,
        )

        # Assert
        assert not unused_chunk.exists()

    @staticmethod
    def test_assemble_fileset(
        client: TestClient,
        fileset: Tuple[str, Dict[str, bytes]],
        tmp_path: Path,
    ):
        """Test assembling fileset.

        This test assembles the fileset once with fetch_missing_chunks=True,
        then wipes the fileset and runs again with fetch_missing_chunks=False.
        This is to test that assembling the fileset writes chunks to the cache
        when fetch_missing_chunks=True.
        """
        # Arrange
        fileset_id, files = fileset
        destination = tmp_path.joinpath("pfs")

        # Act
        # This is hardcoded to replace the host of the presigned URL with
        # localhost:9000. This test needs to be run simultaneously with:
        #   - kubectl port-forward service/minio 9000:9000
        client.storage.assemble_fileset(
            fileset_id,
            path="/",
            destination=destination,
            cache_location=tmp_path,
            fetch_missing_chunks=True,
            http_host_replacement="localhost:9000",
        )

        # Assert
        for file_name, data in files.items():
            file = destination.joinpath(file_name[1:])
            assert file.exists()
            assert file.read_bytes() == data

        # Arrange
        shutil.rmtree(destination)

        # Act
        # This is hardcoded to replace the host of the presigned URL with
        # localhost:9000. This test needs to be run simultaneously with:
        #   - kubectl port-forward service/minio 9000:9000
        client.storage.assemble_fileset(
            fileset_id,
            path="/",
            destination=destination,
            cache_location=tmp_path,
            fetch_missing_chunks=False,  # Use the cache for last assembly
            http_host_replacement="localhost:9000",
        )

        # Assert
        for file in destination.iterdir():
            key = "/" + file.name
            assert file.read_bytes() == files[key]

    @staticmethod
    def test_assemble_fileset_incomplete_cache(
        client: TestClient,
        fileset: Tuple[str, Dict[str, bytes]],
        tmp_path: Path,
    ):
        """Test that missing chunks with fetch_missing_chunks=False
        raises a FileNotFoundError."""
        # Arrange
        fileset_id, files = fileset
        destination = tmp_path.joinpath("pfs")

        # Act & Assert
        with pytest.raises(FileNotFoundError):
            client.storage.assemble_fileset(
                fileset_id,
                path="/",
                destination=destination,
                cache_location=tmp_path,
                fetch_missing_chunks=False,
                http_host_replacement="localhost:9000",
            )
