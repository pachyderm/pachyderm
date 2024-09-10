import os
import unittest.mock

import requests

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

    @staticmethod
    @unittest.mock.patch("requests.get")
    def test_log_network_error(mock, client: TestClient):
        """Test that the CdrResolver logs the HTTP response when a network error occurs."""
        # Arrange
        mocked_response = requests.Response()
        mocked_response.reason = "Bad Request"
        mocked_response.status_code = 400
        mocked_response.url = "http://localhost:9000"
        mocked_response._content = (
            b"<Error>"
            b"<Code>ExpiredToken</Code>"
            b"<Message>The provided token has expired.</Message>"
            b"</Error>"
        )
        mock.return_value = mocked_response

        append_file = storage.AppendFile(path="/file", data=os.urandom(1024))
        request_iterator = iter([storage.CreateFilesetRequest(append_file=append_file)])

        # Act & Assert
        fileset_id = client.storage.create_fileset(request_iterator).fileset_id
        response = client.storage.read_fileset_cdr(fileset_id=fileset_id)

        resolver = CdrResolver()
        with pytest.raises(requests.HTTPError) as err:
            resolver.resolve(next(response).ref)

        assert str(err.value) == (
            "Error 400 - HTTP response: "
            "<Error><Code>ExpiredToken</Code><Message>The provided token has expired.</Message></Error>"
        )
        assert str(err.value.__cause__) == (
            "400 Client Error: Bad Request for url: http://localhost:9000"
        )
