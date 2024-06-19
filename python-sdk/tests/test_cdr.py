import os

from tests.fixtures import *

from pachyderm_sdk.api import storage
from pachyderm_sdk.api.cdr.resolver import CdrResolver


class TestCDR:
    @staticmethod
    def test_resolve_cdr(client: TestClient):
        """Test that the CdrResolver correctly resolves the CDRs returned by
        the storage.read_fileset_cdr RPC.

        Notes:
          * This tests most CDR resolution capabilities, but not all. This test
          is at the whim of whichever CDR functionality PFS is choosing to use.
          * As our tests and local development uses in-cluster object storage
          (usually Minio), we need to patch the host of the presigned URL with
          the correct, external host.
        """
        # Arrange
        files = {f"/file{i}": os.urandom(1024) for i in range(10)}

        # Act
        request_iterator = (
            storage.CreateFilesetRequest(
                append_file=storage.AppendFile(path=path, data=data)
            )
            for path, data in files.items()
        )
        fileset_id = client.storage.create_fileset(request_iterator).fileset_id
        response = client.storage.read_fileset_cdr(fileset_id=fileset_id)

        # This is hardcoded to replace the host of the presigned URL with
        # localhost:9000. This test needs to be run simultaneously with:
        #   - kubectl port-forward service/minio 9000:9000
        resolver = CdrResolver(http_host_replacement="localhost:9000")

        # Assert
        for msg in response:
            assert resolver.resolve(msg.ref) == files[msg.path]
