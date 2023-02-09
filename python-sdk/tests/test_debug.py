from datetime import timedelta

from tests.fixtures import *

from pachyderm_sdk.api import debug


def test_dump(client: TestClient):
    for b in client.debug.dump():
        assert isinstance(b.value, bytes)
        assert len(b.value) > 0


def test_profile_cpu(client: TestClient):
    profile = debug.Profile(name="cpu", duration=timedelta(seconds=1))
    for b in client.debug.profile(profile=profile):
        assert isinstance(b.value, bytes)
        assert len(b.value) > 0


def test_binary(client: TestClient):
    for b in client.debug.binary():
        assert isinstance(b.value, bytes)
        assert len(b.value) > 0
