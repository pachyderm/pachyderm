import json

from jupyter_server.base.handlers import APIHandler, path_regex
from jupyter_server.utils import url_path_join
import tornado

from .pachyderm import READ_ONLY, PachydermMountClient, _parse_pfs_path, _normalize_mode

# Frontend hard codes this in src/handler.ts
NAMESPACE = "pachyderm"
VERSION = "v1"


class BaseHandler(APIHandler):
    @property
    def mount_client(self) -> PachydermMountClient:
        return self.settings["PachydermMountClient"]

    def get_required_query_param_name(self) -> str:
        name = self.get_argument("name", None)
        if name is None:
            raise tornado.web.HTTPError(
                status_code=400, reason="Missing `name` query parameter"
            )
        return name


class ReposHandler(BaseHandler):
    # The following decorator should be present on all verb methods (head, get, post,
    # patch, put, delete, options) to ensure only authorized user can request the
    # Jupyter server
    @tornado.web.authenticated
    def get(self):
        repos = self.mount_client.list()
        # convert nested structure to list
        resp = [
            {
                "repo": repo_name,
                "branches": [
                    {"branch": branch_name, "mount": branch_info["mount"]}
                    for branch_name, branch_info in repo_info["branches"].items()
                ],
            }
            for repo_name, repo_info in repos.items()
        ]
        self.finish(json.dumps(resp))


class RepoHandler(BaseHandler):
    @tornado.web.authenticated
    def get(self, path):
        repo, _, _ = _parse_pfs_path(path)
        result = self.mount_client.get(repo)
        self.finish(
            json.dumps(
                {
                    "repo": repo,
                    "branches": [
                        {"branch": branch_name, "mount": info["mount"]}
                        for branch_name, info in result["branches"].items()
                    ],
                }
            )
        )


class RepoMountHandler(BaseHandler):
    """
    /repos/:repo/:branch/_mount?mode=rw&name=foo
    """

    @tornado.web.authenticated
    def put(self, path):
        name = self.get_required_query_param_name()
        mode = self.get_query_argument("mode", READ_ONLY)  # default ro
        try:
            mode = _normalize_mode(mode)
        except:
            raise tornado.web.HTTPError(
                status_code=400,
                reason=f"{mode} is not valid; valid modes are in {{ro, rw}}",
            )
        repo, branch, _ = _parse_pfs_path(path)
        result = self.mount_client.mount(repo, branch, mode, name)
        self.finish(
            json.dumps({"repo": repo, "branch": branch, "mount": result["mount"]})
        )


class RepoUnmountHandler(BaseHandler):
    @tornado.web.authenticated
    def put(self, path):
        name = self.get_required_query_param_name()
        repo, branch, _ = _parse_pfs_path(path)
        result = self.mount_client.unmount(repo, branch, name)
        self.finish(
            json.dumps({"repo": repo, "branch": branch, "mount": result["mount"]})
        )


class RepoCommitHandler(BaseHandler):
    @tornado.web.authenticated
    def post(self, path):
        name = self.get_required_query_param_name()
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
        (r"/repos/(.+)/_unmount", RepoUnmountHandler),
        (r"/repos/(.+)/_commit", RepoCommitHandler),
    ]

    base_url = web_app.settings["base_url"]
    handlers = [
        (url_path_join(base_url, NAMESPACE, VERSION, endpoint), handler)
        for endpoint, handler in _handlers
    ]
    host_pattern = ".*$"
    web_app.add_handlers(host_pattern, handlers)
