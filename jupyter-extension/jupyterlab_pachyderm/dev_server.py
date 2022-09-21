# Quickly spin up a local service for testing purposes
import tornado.ioloop
import tornado.web

from . import setup_handlers

if __name__ == "__main__":
    app = tornado.web.Application(base_url="/")
    setup_handlers(app)
    app.listen(8888)
    tornado.ioloop.IOLoop.current().start()
