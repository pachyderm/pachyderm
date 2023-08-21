from datetime import timedelta

from tests.fixtures import *

from pachyderm_sdk.api import debug


def test_dump(client: TestClient):
    message = next(client.debug.dump())
    assert isinstance(message.content.content, bytes)
    assert message.progress.progress > 0


def test_profile_cpu(client: TestClient):
    profile = debug.Profile(name="cpu", duration=timedelta(seconds=1))
    message = next(client.debug.profile(profile=profile))
    assert isinstance(message.value, bytes)
    assert len(message.value) > 0


def test_binary(client: TestClient):
    message = next(client.debug.binary())
    assert isinstance(message.value, bytes)
    assert len(message.value) > 0
