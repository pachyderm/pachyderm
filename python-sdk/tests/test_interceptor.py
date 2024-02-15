"""Tests for the gRPC interceptor callback."""
from tests.fixtures import *


def test_bad_serialization(client: TestClient):
    """Test that errors that occur during message serialization are
    caught and explained to the user."""
    # Our interceptor raises TypeError from grpc.RpcError.
    with pytest.raises(TypeError) as err:
        client.pfs.inspect_repo(repo="banana")  # Field `repo` should be a Repo.
    assert err.value.args[0].startswith(
        "An error occurred while trying to serialize the following InspectRepoRequest message."
    )

    with pytest.raises(TypeError) as err:
        list(client.pfs.list_repo(type=True))  # Field `type` should be a string.
    assert err.value.args[0].startswith(
        "An error occurred while trying to serialize the following ListRepoRequest message."
    )


def test_bad_connection():
    """Test that errors which occur due to a failure to connect to
    the server are explained to the user."""
    client = TestClient(nodeid="test_bad_connection", port=9999)
    with pytest.raises(ConnectionError):
        client.get_version()

    # Catching a Connection Error from a stream is flaky - sometimes a
    #   grpc.RpcError is raised as expected and other times a
    #   grpc._channel._MultiThreadedRendezvous is raised in a code path
    #   that the interceptor cannot catch.
    from grpc._channel import _MultiThreadedRendezvous

    with pytest.raises((ConnectionError, _MultiThreadedRendezvous)):
        list(client.pfs.list_repo())
