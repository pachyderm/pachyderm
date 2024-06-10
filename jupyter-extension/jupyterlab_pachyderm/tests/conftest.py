import socket
from pathlib import Path
from shutil import copyfile
from typing import Tuple

import asyncio
import httpx
import pytest
import tornado.testing
import tornado.web

from jupyterlab_pachyderm.handlers import NAMESPACE, VERSION, setup_handlers
from jupyterlab_pachyderm.env import PACH_CONFIG

PortType = Tuple[socket.socket, int]
ENV_VAR_TEST_ADDR = "foo.pachyderm.bar:1234"
PACH_AUTH_CONFIG = Path.home().joinpath(".pachyderm", "auth_config.json")


@pytest.fixture
def pach_config(request, tmp_path) -> Path:
    """Temporary path used to write the pach config for tests.

    If the test is marked with @pytest.mark.no_config then the config
      file is not written.

    If the test is marked with @pytest.mark.auth_enabled then the auth
      enabled config file is used.
    """
    config_path = tmp_path / "config.json"
    if not request.node.get_closest_marker("no_config"):
        if request.node.get_closest_marker("auth_enabled"):
            copyfile(PACH_AUTH_CONFIG, config_path)
        else:
            copyfile(PACH_CONFIG, config_path)
    yield Path(config_path)


@pytest.fixture
def pach_address_env_var(request) -> str:
    """Set the env var to configure the pachd address to a test value.

    Only works if the test is marked with @pytest.mark.env_var.
    """
    if request.node.get_closest_marker("env_var_addr"):
        return ENV_VAR_TEST_ADDR
    return None


@pytest.fixture
def app(pach_config, pach_address_env_var) -> tornado.web.Application:
    """Create a instance of our application.
    This fixture is used by the http_server fixture.
    """
    from jupyter_server.auth.identity import IdentityProvider, User
    from jupyter_server.base.handlers import JupyterHandler

    class TestIdentityProvider(IdentityProvider):
        """An identity provider for the tests, disable auth checks made."""

        def get_user(self, handler: JupyterHandler) -> User:
            """Get the user."""
            return User("test-user")

    app = tornado.web.Application(base_url="/")
    setup_handlers(app, pach_config, pach_address_env_var)
    app.settings["identity_provider"] = TestIdentityProvider()
    app.settings["disable_check_xsrf"] = True
    return app


@pytest.fixture
def http_server_port() -> PortType:
    """Port used by `http_server`"""
    return tornado.testing.bind_unused_port()


@pytest.fixture(name="http_server")
def http_server_fixture(
    app: tornado.web.Application,
    event_loop: asyncio.BaseEventLoop,
    http_server_port: PortType,
) -> tornado.httpserver.HTTPServer:
    """Start a tornado HTTP server that listens on all available handlers.
    ref: github.com/eukaryote/pytest-tornasync/blob/0.6.0.post2/src/pytest_tornasync/plugin.py

    The event_loop fixture is from the pytest-asyncio package.
    """
    server = tornado.httpserver.HTTPServer(app)
    server.add_socket(http_server_port[0])

    yield server

    server.stop()

    if hasattr(server, "close_all_connections"):
        event_loop.run_until_complete(server.close_all_connections())


@pytest.fixture
async def http_client(
    http_server: tornado.httpserver.HTTPServer,
    http_server_port: PortType,
) -> httpx.AsyncClient:
    """Creates a httpx.AsyncClient set with the correct base_url to hit our handlers."""
    assert http_server  # Ensure server is created for tests.
    base_url = f"http://127.0.0.1:{http_server_port[1]}/{NAMESPACE}/{VERSION}"
    async with httpx.AsyncClient(base_url=base_url) as client:
        await client.put("/explore/unmount")
        yield client
