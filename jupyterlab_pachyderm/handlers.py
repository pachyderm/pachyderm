import json

from jupyter_server.base.handlers import APIHandler, path_regex
from jupyter_server.services.contents.handlers import ContentsHandler, validate_model
from jupyter_server.utils import url_path_join, ensure_async
import python_pachyderm
import tornado

from .env import MOCK_PACHYDERM_SERVICE, MOUNT_SERVER_ENABLED, PFS_MOUNT_DIR
from .filemanager import PFSContentsManager
from .log import get_logger
from .mock_pachyderm import MockPachydermClient
from .pachyderm import (
    READ_ONLY,
    READ_WRITE,
    MountInterface,
    PythonPachydermClient,
    PythonPachydermMountClient,
)
from .mount_server_client import MountServerClient


# Frontend hard codes this in src/handler.ts
NAMESPACE = "pachyderm"
VERSION = "v1"


def _normalize_mode(mode):
    if mode in ["r", "ro"]:
        return READ_ONLY
    elif mode in ["w", "rw"]:
        return READ_WRITE
    else:
        raise Exception("Mode not valid")


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
    def mount_client(self) -> MountInterface:
        return self.settings["pachyderm_mount_client"]

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
    async def get(self):
        repos = await self.mount_client.list()
        # convert nested structure to list
        response = [
            {
                "repo": repo_name,
                "branches": [
                    {"branch": branch_name, "mount": branch_info["mount"]}
                    for branch_name, branch_info in repo_info["branches"].items()
                ],
            }
            for repo_name, repo_info in repos.items()
        ]
        get_logger().debug(f"Repos: {response}")
        self.finish(json.dumps(response))


class ReposUnmountHandler(BaseHandler):
    """Unmounts all repos"""

    @tornado.web.authenticated
    async def put(self):
        response = json.dumps({"unmounted": await self.mount_client.unmount_all()})
        get_logger().debug(f"RepoUnmount: {response}")
        self.finish(response)


class RepoHandler(BaseHandler):
    @tornado.web.authenticated
    async def get(self, repo):
        repos = await self.mount_client.list()
        if repo in repos:
            response = json.dumps(
                {
                    "repo": repo,
                    "branches": [
                        {"branch": branch_name, "mount": mount_state["mount"]}
                        for branch_name, mount_state in repos[repo][
                            "branches"
                        ].items()
                    ],
                }
            )
            get_logger().debug(f"Repo: {response}")
            self.finish(response)


class RepoMountHandler(BaseHandler):
    """
    /repos/:repo/:branch/_mount?mode=rw&name=foo
    """

    @tornado.web.authenticated
    async def put(self, path):
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
        response = await self.mount_client.mount(repo, branch, mode, name)
        get_logger().debug(f"RepoMount: {response}")
        self.finish(json.dumps(response))


class RepoUnmountHandler(BaseHandler):
    @tornado.web.authenticated
    async def put(self, path):
        name = self.get_required_query_param_name()
        repo, branch, _ = _parse_pfs_path(path)
        result = await self.mount_client.unmount(repo, branch, name)
        response = json.dumps(
            {"repo": repo, "branch": branch, "mount": result}
        )
        get_logger().debug(f"RepoUnmount: {response}")
        self.finish(response)


class RepoCommitHandler(BaseHandler):
    @tornado.web.authenticated
    async def post(self, path):
        name = self.get_required_query_param_name()
        body = self.get_json_body()
        message = body.get("message", "")
        repo, branch, _ = _parse_pfs_path(path)
        response = await self.mount_client.commit(repo, branch, name, message)
        self.finish(json.dumps(response))


class PFSHandler(ContentsHandler):
    @property
    def pfs_contents_manager(self) -> PFSContentsManager:
        return self.settings["pfs_contents_manager"]

    @tornado.web.authenticated
    async def get(self, path):
        """Copied from https://github.com/jupyter-server/jupyter_server/blob/29be9c6658d7ef04f9b124c54102f7334b610253/jupyter_server/services/contents/handlers.py#L86

        Serves files rooted at PFS_MOUNT_DIR instead of the default content manager's root_dir
        The reason for this is that we want the ability to serve the browser files rooted outside of the default root_dir without overriding it.
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

        model = await ensure_async(
            self.pfs_contents_manager.get(
                path=path,
                type=type,
                format=format,
                content=content,
            )
        )
        validate_model(model, expect_content=content)
        self._finish_model(model, location=False)


def setup_handlers(web_app):
    get_logger().info(f"Using PFS_MOUNT_DIR={PFS_MOUNT_DIR}")
    web_app.settings["pfs_contents_manager"] = PFSContentsManager(PFS_MOUNT_DIR)
    if MOCK_PACHYDERM_SERVICE:
        get_logger().info(
            f"MOCK_PACHYDERM_SERVICE=true -- using the MockPachydermClient"
        )
        web_app.settings["pachyderm_mount_client"] = PythonPachydermMountClient(
            MockPachydermClient(), PFS_MOUNT_DIR
        )
    elif MOUNT_SERVER_ENABLED:
        web_app.settings["pachyderm_mount_client"] = MountServerClient(
            PFS_MOUNT_DIR
        )
    else:
        web_app.settings["pachyderm_mount_client"] = PythonPachydermMountClient(
            PythonPachydermClient(
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
