import grpc

from tests.fixtures import *

from pachyderm_sdk.api import auth
from pachyderm_sdk.constants import AUTH_TOKEN_ENV


@pytest.mark.skipif(
    not os.environ.get(AUTH_TOKEN_ENV),
    reason="auth code not available",
)
class TestUnitAuth:
    @staticmethod
    def test_auth_configuration(auth_client: TestClient):
        response = auth_client.auth.get_configuration()
        assert isinstance(response.configuration, auth.OidcConfig)

    @staticmethod
    def test_cluster_role_bindings(auth_client: TestClient):
        cluster_resource = auth.Resource(type=auth.ResourceType.CLUSTER)
        response = auth_client.auth.get_role_binding(resource=cluster_resource)
        assert response.binding.entries["pach:root"].roles["clusterAdmin"]

        auth_client.auth.modify_role_binding(
            resource=cluster_resource, principal="robot:someuser", roles=["clusterAdmin"]
        )
        response = auth_client.auth.get_role_binding(resource=cluster_resource)
        assert response.binding.entries["robot:someuser"].roles["clusterAdmin"]

    @staticmethod
    def test_authorize(auth_client: TestClient):
        auth_client.auth.authorize(
            resource=auth.Resource(type=auth.ResourceType.REPO, name="foobar"),
            permissions=[auth.Permission.REPO_READ],
        )

    @staticmethod
    def test_who_am_i(auth_client: TestClient):
        assert auth_client.auth.who_am_i().username

    @staticmethod
    def test_get_roles_for_permission(auth_client: TestClient):
        # Checks built-in roles
        response = auth_client.auth.get_roles_for_permission(
            permission=auth.Permission.REPO_READ
        )
        for role in response.roles:
            assert auth.Permission.REPO_READ in role.permissions

        response = auth_client.auth.get_roles_for_permission(
            permission=auth.Permission.CLUSTER_GET_PACHD_LOGS
        )
        for role in response.roles:
            assert auth.Permission.CLUSTER_GET_PACHD_LOGS in role.permissions

    @staticmethod
    def test_robot_token(auth_client: TestClient):
        username = "robot:root"
        auth_token = auth_client.auth.get_robot_token(robot="robot:root", ttl=30).token
        auth_client.auth_token = auth_token
        assert auth_client.auth.who_am_i().username == username
        auth_client.auth.revoke_auth_token(token=auth_token)
        with pytest.raises(grpc.RpcError):
            auth_client.auth.who_am_i()

    @staticmethod
    def test_groups(auth_client: TestClient):
        username = auth_client.auth.who_am_i().username
        group = "testgroup"
        assert auth_client.auth.get_groups().groups == []
        auth_client.auth.set_groups_for_user(username=username, groups=[group])
        assert auth_client.auth.get_groups().groups == [group]
        assert auth_client.auth.get_users(group=group).usernames == [username]
        auth_client.auth.modify_members(group=group, remove=[username])
        assert auth_client.auth.get_groups().groups == []
        assert auth_client.auth.get_users(group=group).usernames == []
