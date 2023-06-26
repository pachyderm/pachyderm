from tests.fixtures import *

from pachyderm_sdk.api import identity
from pachyderm_sdk.constants import AUTH_TOKEN_ENV


@pytest.mark.skipif(
        not os.environ.get(AUTH_TOKEN_ENV),
        reason="auth code not available",
    )
class TestIdentity:

    @staticmethod
    def test_identity_server_config(auth_client: TestClient):
        isc = auth_client.identity.get_identity_server_config()
        assert isc.config.issuer == "http://pachd:30658/dex"

    @staticmethod
    def test_oidc_client(auth_client: TestClient):
        auth_client.identity.delete_all()

        oidc1 = auth_client.identity.create_oidc_client(
            client=identity.OidcClient(id="oidc1", name="pach1", secret="secret1")
        )
        oidc2 = auth_client.identity.create_oidc_client(
            client=identity.OidcClient(id="oidc2", name="pach2", secret="secret2")
        )

        assert len(auth_client.identity.list_oidc_clients().clients) == 2
        assert auth_client.identity.get_oidc_client(id=oidc1.client.id).client.name == "pach1"
        assert auth_client.identity.get_oidc_client(id=oidc2.client.id).client.name == "pach2"

        auth_client.identity.update_oidc_client(
            client=identity.OidcClient(id="oidc1", name="pach3", secret="secret1")
        )
        assert auth_client.identity.get_oidc_client(id=oidc1.client.id).client.name == "pach3"

        auth_client.identity.delete_oidc_client(id=oidc1.client.id)
        assert len(auth_client.identity.list_oidc_clients().clients) == 1
