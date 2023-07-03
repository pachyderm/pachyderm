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
        client1 = identity.OidcClient(id="oidc1", name="pach1", secret="secret1")
        client2 = identity.OidcClient(id="oidc2", name="pach2", secret="secret2")
        try:
            oidc1 = auth_client.identity.create_oidc_client(client=client1)
            oidc2 = auth_client.identity.create_oidc_client(client=client2)

            oidc_clients = auth_client.identity.list_oidc_clients().clients
            assert oidc1.client in oidc_clients
            assert oidc2.client in oidc_clients

            assert (
                auth_client.identity.get_oidc_client(id=client1.id).client.name
                == client1.name
            )
            assert (
                auth_client.identity.get_oidc_client(id=client2.id).client.name
                == client2.name
            )

            client1.name = "pach3"
            auth_client.identity.update_oidc_client(client=client1)
            assert (
                auth_client.identity.get_oidc_client(id=client1.id).client.name
                == client1.name
            )

            auth_client.identity.delete_oidc_client(id=oidc1.client.id)
            oidc_clients = auth_client.identity.list_oidc_clients().clients
            assert oidc1 not in oidc_clients

        finally:
            oidc_clients = auth_client.identity.list_oidc_clients().clients
            if client1 in oidc_clients:
                auth_client.identity.delete_oidc_client(id=client1.id)
            if client2 in oidc_clients:
                auth_client.identity.delete_oidc_client(id=client2.id)
