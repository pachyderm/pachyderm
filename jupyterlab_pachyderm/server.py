# Quickly spin up a local service for testing purposes
import os

import tornado.ioloop
import tornado.web

from .handlers import setup_handlers
from .pachyderm import PachydermMountClient
from .mock_pachyderm import MockPachydermMountClient


if __name__ == "__main__":
    app = tornado.web.Application(base_url="/")
    app.settings["PachydermMountService"] = PachydermMountClient()
    # swap real PachydermMountServide with mock given MOCK_PACHYDERM_SERVICE
    if "MOCK_PACHYDERM_SERVICE" in os.environ:
        app.settings["PachydermMountService"] = MockPachydermMountClient()
    setup_handlers(app)
    app.listen(12345)
    tornado.ioloop.IOLoop.current().start()
