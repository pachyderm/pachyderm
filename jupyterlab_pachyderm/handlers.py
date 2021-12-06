import os
import json

from jupyter_server.base.handlers import APIHandler, path_regex
from jupyter_server.services.contents.handlers import ContentsHandler, validate_model
from jupyter_server.utils import url_path_join
import python_pachyderm
import tornado

from .env import MOCK_PACHYDERM_SERVICE, PFS_MOUNT_DIR
from .filemanager import PFSContentsManager
from .mock_pachyderm import MockPachydermClient
from .pachyderm import (
    READ_ONLY,
    PachydermClient,
    PachydermMountClient,
    _parse_pfs_path,
    _normalize_mode,
)


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


class ReposUnmountHandler(BaseHandler):
    """Unmounts all repos"""

    @tornado.web.authenticated
    def put(self):
        self.finish(json.dumps({"unmounted": self.mount_client.unmount_all()}))


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
        if result:
            self.finish(
                json.dumps(
                    {"repo": repo, "branch": branch, "mount": result.get("mount")}
                )
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


class PFSHandler(ContentsHandler):
    @property
    def pfs_contents_manager(self) -> PFSContentsManager:
        return self.settings["pfs_contents_manager"]

    @tornado.web.authenticated
    def get(self, path):
        """Copied from https://github.com/jupyter-server/jupyter_server/blob/29be9c6658d7ef04f9b124c54102f7334b610253/jupyter_server/services/contents/handlers.py#L86

        The main purpose for this method is to override contents_manager.root_dir at request time,
        then reset root_dir to its original value so that our implementation does not conflict with the default content manager.
        """
        path = path or ""
        type = self.get_query_argument("type", default=None)
        if type not in {None, "directory", "file", "notebook"}:
            raise tornado.web.HTTPError(400, "Type %r is invalid" % type)
        format = self.get_query_argument("format", default=None)
        if format not in {None, "text", "base64"}:
            raise tornado.web.HTTPError(400, "Format %r is invalid" % format)
        content = self.get_query_argument("content", default="1")
        if content not in {"0", "1"}:
            raise tornado.web.HTTPError(400, "Content %r is invalid" % content)
        content = int(content)

        model = self.pfs_contents_manager.get(
            path=path,
            type=type,
            format=format,
            content=content,
        )
        validate_model(model, expect_content=content)
        self._finish_model(model, location=False)


def setup_handlers(web_app):
    if MOCK_PACHYDERM_SERVICE:
        web_app.settings["PachydermMountClient"] = PachydermMountClient(
            MockPachydermClient(), PFS_MOUNT_DIR
        )
    else:
        web_app.settings["pfs_contents_manager"] = PFSContentsManager(PFS_MOUNT_DIR)
        web_app.settings["PachydermMountClient"] = PachydermMountClient(
            PachydermClient(
                python_pachyderm.Client(), python_pachyderm.ExperimentalClient()
            ),
            PFS_MOUNT_DIR,
        )

    _handlers = [
        ("/repos", ReposHandler),
        ("/repos/_unmount", ReposUnmountHandler),
        (r"/repos/([^/]+)", RepoHandler),
        (r"/repos/(.+)/_mount", RepoMountHandler),
        (r"/repos/(.+)/_unmount", RepoUnmountHandler),
        (r"/repos/(.+)/_commit", RepoCommitHandler),
        (r"/pfs%s" % path_regex, PFSHandler),
    ]

    base_url = web_app.settings["base_url"]
    handlers = [
        (url_path_join(base_url, NAMESPACE, VERSION, endpoint), handler)
        for endpoint, handler in _handlers
    ]
    host_pattern = ".*$"
    web_app.add_handlers(host_pattern, handlers)
