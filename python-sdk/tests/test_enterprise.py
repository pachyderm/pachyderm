from tests.fixtures import *

from pachyderm_sdk.api import enterprise
from pachyderm_sdk.constants import AUTH_TOKEN_ENV


@pytest.mark.skipif(
    not os.environ.get(AUTH_TOKEN_ENV),
    reason="auth code not available",
)
class TestEnterprise:
    @staticmethod
    def test_enterprise(auth_client: TestClient):
        """Testing the enterprise specific features."""
        assert auth_client.enterprise.get_state().state == enterprise.State.ACTIVE
        assert isinstance(auth_client.enterprise.heartbeat(), enterprise.HeartbeatResponse)
        assert (
            auth_client.enterprise.pause_status().status
            == enterprise.PauseStatusResponsePauseStatus.UNPAUSED
        )
