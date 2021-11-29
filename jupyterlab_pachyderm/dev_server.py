# Quickly spin up a local service for testing purposes
import os

import tornado.ioloop
import tornado.web
import python_pachyderm

from . import setup_handlers
from .pachyderm import PachydermClient, PachydermMountClient
from .mock_pachyderm import MockPachydermClient

if __name__ == "__main__":
    app = tornado.web.Application(base_url="/")
    setup_handlers(app)
    app.listen(8888)
    tornado.ioloop.IOLoop.current().start()
