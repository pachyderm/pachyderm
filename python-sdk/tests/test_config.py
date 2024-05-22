import json
from filecmp import cmp
from pathlib import Path

import pytest

from pachyderm_sdk.config import ConfigFile, Context
from pachyderm_sdk.errors import ConfigError

TEST_CONFIG = """{
  "user_id": "some_user",
  "v2": {
    "active_context": "default",
    "contexts": {
      "default": {
        "cluster_deployment_id": "id",
        "project": "default",
        "session_token": "token"
      },
      "other_context": {
        "cluster_deployment_id": "other_id",
        "pachd_address": "grpc://localhost:12345",
        "project": "other_project",
        "session_token": "other_token"
      }
    },
    "metrics": true
  }
}
"""

TEST_CONFIG_INVALID_CONTEXT = """
{
  "user_id": "some_user",
  "v2": {
    "active_context": "invalid_context",
    "contexts": {
      "default": {
        "session_token": "token",
        "cluster_deployment_id": "id",
        "project": "default"
      },
      "other_context": {
        "pachd_address": "grpc://localhost:12345",
        "session_token": "other_token",
        "cluster_deployment_id": "other_id",
        "project": "other_project"
      }
    },
    "metrics": true
  }
}
"""


class TestConfig:
    @staticmethod
    def test_from_file(tmp_path):
        config_path = tmp_path / "config.json"
        with open(config_path, "w") as config_file:
            config_file.write(TEST_CONFIG)

        config = ConfigFile.from_path(config_path)
        context = config.active_context
        assert config.user_id == "some_user"
        assert context.session_token == "token"
        assert context.cluster_deployment_id == "id"
        assert context.project == "default"

    @staticmethod
    def test_invalid_context(tmp_path):
        config = ConfigFile(json.loads(TEST_CONFIG_INVALID_CONTEXT))
        with pytest.raises(ConfigError) as error:
            _ = config.active_context
        assert "active context not found: invalid_context" in str(error.value)

    @staticmethod
    def test_to_from_file(tmp_path: Path):
        """Test that reading from and writing to a file compatibly implemented"""
        # Arrange
        original_path = tmp_path / "config.json"
        with open(original_path, "w") as config_file:
            config_file.write(TEST_CONFIG)
        original_config = ConfigFile.from_path(original_path)

        # Act
        new_path = tmp_path / "new_config.json"
        original_config.write(new_path)

        assert original_path.read_text() == new_path.read_text()
        assert cmp(original_path, new_path, shallow=False)

    @staticmethod
    def test_new_with_context():
        """Test that a new config file object can be created from a single context."""
        # Arrange
        context = Context(
            cluster_deployment_id="abc123",
            pachd_address="grpc://localhost:12345",
            project="default",
        )

        # Act
        config = ConfigFile.new_with_context(name="my-new-context", context=context)

        # Assert
        assert config.active_context == context
        assert config.user_id is not None
        assert len(config.user_id) == 32

    @staticmethod
    def test_set_active_context():
        """Test that set_active_context updates the active context as expected."""
        # Arrange
        config = ConfigFile(data=json.loads(TEST_CONFIG))

        # Act & Assert
        assert config.active_context.cluster_deployment_id == "id"
        config.set_active_context("other_context")
        assert config.active_context.cluster_deployment_id == "other_id"

        with pytest.raises(ValueError):
            config.set_active_context("fake-context")
