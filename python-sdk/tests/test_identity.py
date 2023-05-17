from tests.fixtures import *

from pachyderm_sdk.api import identity
from pachyderm_sdk.constants import ENTERPRISE_CODE_ENV


@pytest.mark.skipif(
        not os.environ.get(ENTERPRISE_CODE_ENV),
        reason="enterprise code not available",
    )
class TestIdentity:

    @staticmethod
    def test_identity_server_config(auth_client: TestClient):
        isc = auth_client.identity.get_identity_server_config()
        assert isc.issuer == "http://localhost:1658"

    @staticmethod
    def test_oidc_client(auth_client: TestClient):
        oidc1 = auth_client.identity.create_oidc_client(
            client=identity.OidcClient(id="oidc1", name="pach1", secret="secret1")
        )
        oidc2 = auth_client.identity.create_oidc_client(
            client=identity.OidcClient(id="oidc2", name="pach2", secret="secret2")
        )

        assert len(auth_client.identity.list_oidc_clients().clients) == 2
        assert auth_client.identity.get_oidc_client(oidc1.id).name == "pach1"
        assert auth_client.identity.get_oidc_client(oidc2.id).name == "pach2"

        auth_client.identity.update_oidc_client(
            client=identity.OidcClient(id="oidc1", name="pach3", secret="secret1")
        )
        assert auth_client.identity.get_oidc_client(oidc1.id).name == "pach3"

        auth_client.identity.delete_oidc_client(oidc1.id)
        assert len(auth_client.identity.list_oidc_clients()) == 1
