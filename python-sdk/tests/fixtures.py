import pytest

from tests.client import TestClient as _TestClient


@pytest.fixture(params=[True, False])
def default_project(request) -> bool:
    """Parametrized fixture with values True|False.

    Use this fixture to easily test resources against
      default and non-default projects.
    """
    return request.param


@pytest.fixture
def client(request) -> _TestClient:
    client = _TestClient(nodeid=request.node.nodeid)
    yield client
    client.tear_down()
