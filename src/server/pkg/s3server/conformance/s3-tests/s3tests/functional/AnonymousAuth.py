from boto.auth_handler import AuthHandler

class AnonymousAuthHandler(AuthHandler):
    def add_auth(self, http_request, **kwargs):
        return # Nothing to do for anonymous access!
