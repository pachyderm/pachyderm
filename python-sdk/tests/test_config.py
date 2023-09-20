import pytest

from pachyderm_sdk.config import ConfigFile
from pachyderm_sdk.errors import ConfigError

TEST_CONFIG = """
{
  "user_id": "some_user",
  "v2": {
    "active_context": "default",
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
            config_file.flush()
            c = ConfigFile(config_file.name)
            context = c.active_context
            assert c.user_id == "some_user"
            assert context.session_token == "token"
            assert context.cluster_deployment_id == "id"
            assert context.project == "default"

    @staticmethod
    def test_invalid_context(tmp_path):
        config_path = tmp_path / "config.json"
        with open(config_path, "w") as config_file:
            config_file.write(TEST_CONFIG_INVALID_CONTEXT)
            config_file.flush()
            with pytest.raises(ConfigError) as e:
                c = ConfigFile(config_file=config_file.name)
                context = c.active_context
            assert "active context not found: invalid_context" in str(e.value)
