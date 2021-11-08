import json

from jupyter_server.base.handlers import APIHandler
from jupyter_server.utils import url_path_join

import tornado

# Frontend hard codes this in src/handler.ts
NAMESPACE = "pachyderm"
VERSION = "v1"

# Fake data to simulate a mock API
repos = [
    {"id": "images", "mount_state": False},
    {"id": "edges", "mount_state": False},
]


def _find_repo(repo_id):
    for repo in repos:
        if repo_id == repo["id"]:
            return repo


class ReposHandler(APIHandler):
    # The following decorator should be present on all verb methods (head, get, post,
    # patch, put, delete, options) to ensure only authorized user can request the
    # Jupyter server
    @tornado.web.authenticated
    def get(self):
        self.finish(json.dumps(repos))


class RepoHandler(APIHandler):
    # The following decorator should be present on all verb methods (head, get, post,
    # patch, put, delete, options) to ensure only authorized user can request the
    # Jupyter server
    @tornado.web.authenticated
    def get(self, repo_id):
        repo = _find_repo(repo_id)
        if repo:
            self.finish(json.dumps(repo))
        else:
            self.write_error(404)


class RepoMountHandler(APIHandler):
    @tornado.web.authenticated
    def put(self, repo_id=None):
        repo = _find_repo(repo_id)
        repo["mount_state"] = True
        self.finish(json.dumps(repo))


class RepoUnMountHandler(APIHandler):
    @tornado.web.authenticated
    def put(self, repo_id=None):
        repo = _find_repo(repo_id)
        repo["mount_state"] = False
        self.finish(json.dumps(repo))


class RepoCommitHandler(APIHandler):
    @tornado.web.authenticated
    def post(self, repo_id=None):
        print(RepoCommitHandler.put)


def setup_handlers(web_app):

    _handlers = [
        ("/repos", ReposHandler),
        ("/repos/([^/]+)", RepoHandler),
        ("/repos/(.+)/_mount", RepoMountHandler),
        ("/repos/(.+)/_unmount", RepoUnMountHandler),
        ("/repos/(.+)/_commit", RepoCommitHandler)
    ]

    base_url = web_app.settings["base_url"]
    handlers = [
        (url_path_join(base_url, NAMESPACE, VERSION, endpoint), handler)
        for endpoint, handler in _handlers
    ]
    host_pattern = ".*$"
    web_app.add_handlers(host_pattern, handlers)
