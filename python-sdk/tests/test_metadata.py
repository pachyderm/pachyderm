from tests.fixtures import *

from pachyderm_sdk.api import metadata


class TestMetadata:
    @staticmethod
    def test_setting_metadata(client: TestClient):
        """Sanity check for the Metadata API"""
        # Arrange
        repo = client.new_repo()
        key, value = "TestKey", "TestValue"

        # Act
        client.metadata.edit_metadata(
            edits=[
                metadata.Edit(
                    add_key=metadata.EditAddKey(key=key, value=value),
                    repo=repo.as_picker(),
                )
            ]
        )
        repo_info = client.pfs.inspect_repo(repo=repo)

        # Assert
        assert repo_info.metadata == {key: value}
