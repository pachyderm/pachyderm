# Quickly spin up a local service for testing purposes
import tornado.ioloop
import tornado.web
from jupyter_server.auth.identity import IdentityProvider, User
from jupyter_server.base.handlers import JupyterHandler

from . import setup_handlers
from .env import PACH_CONFIG, PACHD_ADDRESS, DEX_TOKEN


class TestIdentityProvider(IdentityProvider):
    """An identity provider for the tests, disable auth checks made."""

    def get_user(self, handler: JupyterHandler) -> User:
        """Get the user."""
        return User("testuser")


if __name__ == "__main__":
    app = tornado.web.Application(base_url="/")
    setup_handlers(app, PACH_CONFIG, PACHD_ADDRESS, DEX_TOKEN)

    # Disable authorization checks between test API and dev-server
    app.settings["identity_provider"] = TestIdentityProvider()
    app.settings["disable_check_xsrf"] = True

    app.listen(8888)
    tornado.ioloop.IOLoop.current().start()
