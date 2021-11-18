import json

from jupyter_server.base.handlers import APIHandler, path_regex
from jupyter_server.utils import url_path_join
import tornado

from .pachyderm import PachydermMountClient

# Frontend hard codes this in src/handler.ts
NAMESPACE = "pachyderm"
VERSION = "v1"


def _parse_pfs_path(path):
    """
    a path can be one of
        - repo/branch/commit
        - repo/branch
        - repo -> defaults to master branch
    returns a 3-tuple (repo, branch, commit)
    """
    parts = path.split("/")
    if len(parts) == 3:
        return tuple(parts)
    if len(parts) == 2:
        return parts[0], parts[1], None
    if len(parts) == 1:
        return parts[0], "master", None


class BaseHandler(APIHandler):
    @property
    def mount_client(self) -> PachydermMountClient:
        return self.settings["PachydermMountService"]
    
    def get_name_query_parameter(self) -> str:
        name = self.get_argument("name", None)
        if name is None:
            raise tornado.web.HTTPError(400, "Missing {name} query parameter")
        return name


class ReposHandler(BaseHandler):
    # The following decorator should be present on all verb methods (head, get, post,
    # patch, put, delete, options) to ensure only authorized user can request the
    # Jupyter server
    @tornado.web.authenticated
    def get(self):
        resp = self.mount_client.list()
        self.finish(json.dumps(resp))


class RepoHandler(BaseHandler):
    @tornado.web.authenticated
    def get(self, path):
        repo, _, _ = _parse_pfs_path(path)
        resp = self.mount_client.get(repo)
        self.finish(json.dumps(resp))


class RepoMountHandler(BaseHandler):
    """
    /repos/:repo/:branch/_mount?mode=rw&name=foo
    """
    @tornado.web.authenticated
    def put(self, path):
        name = self.get_name_query_parameter() 
        mode = self.get_query_argument("mode", None)
        repo, branch, _ = _parse_pfs_path(path)
        resp = self.mount_client.mount(repo, branch, mode, name)
        self.finish(json.dumps(resp))


class RepoUnMountHandler(BaseHandler):
    @tornado.web.authenticated
    def put(self, path):
        name = self.get_name_query_parameter()
        repo, branch, _ = _parse_pfs_path(path)
        resp = self.mount_client.unmount(repo, branch, name)
        self.finish(json.dumps(resp))


class RepoCommitHandler(BaseHandler):
    @tornado.web.authenticated
    def post(self, path):
        name = self.get_name_query_parameter()
        body = self.get_json_body()
        message = body.get("message", "")
        repo, branch, _ = _parse_pfs_path(path)
        resp = self.mount_client.commit(repo, branch, name, message)
        self.finish(json.dumps(resp))


def setup_handlers(web_app):

    _handlers = [
        ("/repos", ReposHandler),
        (r"/repos/([^/]+)", RepoHandler),
        (r"/repos/(.+)/_mount", RepoMountHandler),
        (r"/repos/(.+)/_unmount", RepoUnMountHandler),
        (r"/repos/(.+)/_commit", RepoCommitHandler),
    ]

    base_url = web_app.settings["base_url"]
    handlers = [
        (url_path_join(base_url, NAMESPACE, VERSION, endpoint), handler)
        for endpoint, handler in _handlers
    ]
    host_pattern = ".*$"
    web_app.add_handlers(host_pattern, handlers)
