"""Tests for the config screen and handler."""
from pathlib import Path

import pytest
from httpx import AsyncClient
from tornado.web import Application

from pachyderm_sdk.config import ConfigFile

from jupyterlab_pachyderm.env import PACH_CONFIG

from .conftest import ENV_VAR_TEST_ADDR


@pytest.mark.no_config
async def test_config_no_file(app: Application, http_client: AsyncClient):
    """Test that if there is no config file present, the extension does not
    use a default config and the cluster status is INVALID."""
    # Arrange
    config_file = app.settings.get("pachyderm_config_file")
    assert config_file is not None and not config_file.exists()

    # Act
    config_response = await http_client.get("/config")
    health_response = await http_client.get("/health")

    # Assert
    config_response.raise_for_status()
    health_response.raise_for_status()
    config_payload = config_response.json()
    health_payload = health_response.json()
    assert health_payload["status"] == "HEALTHY_INVALID_CLUSTER"
    assert config_payload["pachd_address"] == ""


@pytest.mark.auth_enabled
async def test_auth_config(app: Application, http_client: AsyncClient):
    """Test that the extension correctly loads an auth-enabled config file."""
    # Arrange
    config_file = app.settings.get("pachyderm_config_file")
    assert config_file is not None and config_file.exists()

    # Act
    config_response = await http_client.get("/config")
    health_response = await http_client.get("/health")

    # Assert
    config_response.raise_for_status()
    health_response.raise_for_status()
    config_payload = config_response.json()
    health_payload = health_response.json()
    assert health_payload["status"] == "HEALTHY_LOGGED_IN"
    assert config_payload["pachd_address"]


@pytest.mark.no_config
@pytest.mark.env_var_addr
async def test_config_env_var(app: Application, http_client: AsyncClient):
    """Test that when the PACHD_ADDRESS env var is set and there is no config
    file, that the client is configured using the env var"""
    config_file: Path = app.settings.get("pachyderm_config_file")
    assert config_file is not None and config_file.exists()
    config = ConfigFile.from_path(config_file)
    assert config.active_context.pachd_address == ENV_VAR_TEST_ADDR

    config_response = await http_client.get("/config")
    health_response = await http_client.get("/health")

    config_response.raise_for_status()
    health_response.raise_for_status()
    config_payload = config_response.json()
    health_payload = health_response.json()
    assert health_payload["status"] == "HEALTHY_INVALID_CLUSTER"
    assert config_payload["pachd_address"] == ENV_VAR_TEST_ADDR


@pytest.mark.env_var_addr
async def test_config_env_var_overridden(app: Application, http_client: AsyncClient):
    """Test that when the PACHD_ADDRESS env var is set and there is a config file,
    that the client uses the config file and not the env var"""
    config_file = app.settings.get("pachyderm_config_file")
    assert config_file is not None and config_file.exists()

    config_response = await http_client.get("/config")
    health_response = await http_client.get("/health")

    config_response.raise_for_status()
    health_response.raise_for_status()
    config_payload = config_response.json()
    health_payload = health_response.json()
    assert (
        health_payload["status"] == "HEALTHY_NO_AUTH"
        or health_payload["status"] == "HEALTHY_LOGGED_IN"
    )
    assert config_payload["pachd_address"] != ENV_VAR_TEST_ADDR


@pytest.mark.no_config
async def test_do_not_set_invalid_config(app: Application, http_client: AsyncClient):
    """Test that PUT /config does not store an invalid config nor write to a config file."""
    # Arrange
    assert app.settings.get("pachyderm_client") is None
    config_file = app.settings.get("pachyderm_config_file")
    assert config_file is not None and not config_file.exists()

    invalid_address = "localhost:33333"
    data = {"pachd_address": invalid_address}

    # Act
    config_response = await http_client.put("/config", json=data)
    health_response = await http_client.get("/health")
    config_get_response = await http_client.get("/config")

    # Assert
    assert config_response.status_code == 400
    config_get_response.raise_for_status()
    health_response.raise_for_status()
    health_payload = health_response.json()
    config_get_payload = config_get_response.json()
    assert health_payload["status"] == "HEALTHY_INVALID_CLUSTER"
    assert config_get_payload["pachd_address"] == ""
    assert app.settings.get("pachyderm_client") is None

    # Ensure that no config file was created.
    assert config_file is not None and not config_file.exists()


async def test_config(pach_config: Path, http_client: AsyncClient):
    """Test that PUT /config with a valid configuration is processed
    by the ConfigHandler as expected."""
    # PUT request
    local_active_context = ConfigFile.from_path(PACH_CONFIG).active_context

    payload = {"pachd_address": local_active_context.pachd_address}
    r = await http_client.put("/config", json=payload)
    health_response = await http_client.get("/health")

    config = ConfigFile.from_path(pach_config)

    assert r.status_code == 200, r.text
    assert r.json()["pachd_address"] in config.active_context.pachd_address
    assert health_response.json()["status"] != "HEALTHY_INVALID"
    assert health_response.json()["status"] != "UNHEALTHY"

    r = await http_client.get("/config")
    assert r.status_code == 200, r.text
