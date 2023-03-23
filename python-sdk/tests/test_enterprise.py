from tests.fixtures import *

from pachyderm_sdk.api import enterprise
from pachyderm_sdk.constants import ENTERPRISE_CODE_ENV


@pytest.mark.skipif(
        not os.environ.get(ENTERPRISE_CODE_ENV),
        reason="enterprise code not available",
    )
class TestEnterprise:

    @staticmethod
    def test_enterprise(auth_client: TestClient):
        clusters = auth_client.license.list_clusters().clusters
        user_clusters = auth_client.license.list_user_clusters().clusters
        assert len(clusters) == len(user_clusters)
        assert auth_client.enterprise.get_state().state == enterprise.State.ACTIVE
        assert (
            auth_client.enterprise.get_activation_code().activation_code
            == os.environ.get(ENTERPRISE_CODE_ENV)
        )
        assert (
            auth_client.enterprise.get_pause_status().status
            == enterprise.PauseStatusResponse.PauseStatus.UNPAUSED
    )