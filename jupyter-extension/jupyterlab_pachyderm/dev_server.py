# Quickly spin up a local service for testing purposes
import tornado.ioloop
import tornado.web
from jupyter_server.auth.identity import IdentityProvider, User
from jupyter_server.base.handlers import JupyterHandler

from . import setup_handlers


class TestIdentityProvider(IdentityProvider):
    """An identity provider for the tests, disable auth checks made."""

    def get_user(self, handler: JupyterHandler) -> User:
        """Get the user."""
        return User("testuser")


if __name__ == "__main__":
    app = tornado.web.Application(base_url="/")
    setup_handlers(app)

    # Disable authorization checks between test API and dev-server
    app.settings["identity_provider"] = TestIdentityProvider()
    app.settings["disable_check_xsrf"] = True

    # Increase the timeout on the MountServerClient
    app.settings["pachyderm_mount_client"].client.configure(
        None,defaults=dict(connect_timeout=20, request_timeout=60))

    app.listen(8888)
    tornado.ioloop.IOLoop.current().start()
