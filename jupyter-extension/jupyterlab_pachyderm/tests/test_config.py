"""Tests for the config screen and handler."""
from pathlib import Path

import pytest
from httpx import AsyncClient
from tornado.web import Application

from pachyderm_sdk.config import ConfigFile

from jupyterlab_pachyderm.env import PACH_CONFIG


@pytest.mark.no_config
async def test_config_no_file(app: Application, http_client: AsyncClient):
    """Test that if there is no config file present, the extension does not
    use a default config and the cluster status is INVALID."""
    # Arrange
    config_file = app.settings.get("pachyderm_config_file")
    assert config_file is not None and not config_file.exists()

    # Act
    response = await http_client.get("/config")

    # Assert
    response.raise_for_status()
    payload = response.json()
    assert payload["cluster_status"] == "UNKNOWN"
    assert payload["pachd_address"] == ""


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
    response = await http_client.put("/config", json=data)

    # Assert
    response.raise_for_status()
    payload = response.json()
    assert payload["cluster_status"] == "INVALID"
    assert payload["pachd_address"] == invalid_address
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

    config = ConfigFile.from_path(pach_config)

    assert r.status_code == 200, r.text
    assert r.json()["cluster_status"] != "INVALID"
    assert r.json()["pachd_address"] in config.active_context.pachd_address

    # GET request
    r = await http_client.get("/config")

    assert r.status_code == 200, r.text
    assert r.json()["cluster_status"] != "INVALID"
