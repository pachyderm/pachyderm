import pytest

pytest_plugins = ["jupyter_server.pytest_plugin"]

@pytest.fixture
def jp_server_config():
    return {
        "ServerApp": {"jpserver_extensions": {"jupyterlab_pachyderm": True}}
    }


async def test_handler_default(jp_fetch):
    r = await jp_fetch("/pachyderm/v1/repos")
    assert r.code == 200
    assert r.body == b'[{"id": "images", "mount_state": false}, {"id": "edges", "mount_state": false}]'
