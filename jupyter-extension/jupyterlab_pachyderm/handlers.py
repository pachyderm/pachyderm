import json

from jupyter_server.base.handlers import APIHandler, path_regex
from jupyter_server.services.contents.handlers import ContentsHandler, validate_model
from jupyter_server.utils import url_path_join, ensure_async
import tornado

from .env import PFS_MOUNT_DIR
from .filemanager import PFSContentsManager
from .log import get_logger
from .pachyderm import READ_ONLY, READ_WRITE, MountInterface
from .mount_server_client import MountServerClient


# Frontend hard codes this in src/handler.ts
NAMESPACE = "pachyderm"
VERSION = "v2"


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


def _transform_response(resp):
    return [
        {
            "repo": repo_name,
            "authorization": repo_info["authorization"],
            "branches": [
                {"branch": branch_name, "mount": branch_info["mount"]}
                for branch_name, branch_info in repo_info["branches"].items()
            ],
        }
        for repo_name, repo_info in json.loads(resp).items()
    ]


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
        try:
            response = _transform_response(await self.mount_client.list())
            get_logger().debug(f"Repos: {response}")
            self.finish(json.dumps(response))
        except Exception as e:
            get_logger().error("Error listing repos.", exc_info=True)
            raise tornado.web.HTTPError(
                status_code=getattr(e, "code", 500), reason=f"Error listing repos: {e}."
            )


class ReposUnmountHandler(BaseHandler):
    """Unmounts all repos"""

    @tornado.web.authenticated
    async def put(self):
        try:
            response = _transform_response(await self.mount_client.unmount_all())
            get_logger().debug(f"RepoUnmount: {response}")
            self.finish(json.dumps(response))
        except Exception as e:
            get_logger().error("Error unmounting all repos.", exc_info=True)
            raise tornado.web.HTTPError(
                status_code=getattr(e, "code", 500),
                reason=f"Error unmounting all repos: {e}.",
            )


class RepoHandler(BaseHandler):
    @tornado.web.authenticated
    async def get(self, repo):
        try:
            repos = json.loads(await self.mount_client.list())
        except Exception as e:
            get_logger().error(f"Error listing repos.", exc_info=True)
            raise tornado.web.HTTPError(
                status_code=getattr(e, "code", 500), reason=f"Error listing repos: {e}."
            )

        if repo not in repos:
            raise tornado.web.HTTPError(
                status_code=400, reason=f"Error repo {repo} not found."
            )

        response = json.dumps(
            {
                "repo": repo,
                "authorization": repos[repo]["authorization"],
                "branches": [
                    {"branch": branch_name, "mount": mount_state["mount"]}
                    for branch_name, mount_state in repos[repo]["branches"].items()
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

        try:
            repo, branch, _ = _parse_pfs_path(path)
            response = _transform_response(
                await self.mount_client.mount(repo, branch, mode, name)
            )
            get_logger().debug(f"RepoMount: {response}")
            self.finish(json.dumps(response))
        except Exception as e:
            get_logger().error(f"Error mounting {(repo, branch, name)}.", exc_info=True)
            raise tornado.web.HTTPError(
                status_code=getattr(e, "code", 500),
                reason=f"Error mounting repo {repo}: {e}.",
            )


class RepoUnmountHandler(BaseHandler):
    @tornado.web.authenticated
    async def put(self, path):
        name = self.get_required_query_param_name()
        repo, branch, _ = _parse_pfs_path(path)
        try:
            response = _transform_response(
                await self.mount_client.unmount(repo, branch, name)
            )
            get_logger().debug(f"RepoUnmount: {response}")
            self.finish(json.dumps(response))
        except Exception as e:
            get_logger().error(f"Error unmounting repo {repo}.", exc_info=True)
            raise tornado.web.HTTPError(
                status_code=getattr(e, "code", 500),
                reason=f"Error unmounting repo {repo}: {e}.",
            )


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


class ConfigHandler(BaseHandler):
    @tornado.web.authenticated
    async def put(self):
        try:
            body = self.get_json_body()
            response = await self.mount_client.config(body)
            self.finish(response)
        except Exception as e:
            get_logger().error(
                f"Error updating config with endpoint {body['pachd_address']}.",
                exc_info=True,
            )
            raise tornado.web.HTTPError(
                status_code=500,
                reason=f"Error updating config with endpoint {body['pachd_address']}: {e}.",
            )

    @tornado.web.authenticated
    async def get(self):
        try:
            response = await self.mount_client.config()
            self.finish(response)
        except Exception as e:
            get_logger().error("Error getting config.", exc_info=True)
            raise tornado.web.HTTPError(
                status_code=500, reason=f"Error getting config: {e}."
            )


class AuthLoginHandler(BaseHandler):
    @tornado.web.authenticated
    async def put(self):
        try:
            response = await self.mount_client.auth_login()
            self.finish(response)
        except Exception as e:
            get_logger().error("Error logging in to auth.", exc_info=True)
            raise tornado.web.HTTPError(
                status_code=500, reason=f"Error logging in to auth: {e}."
            )


class AuthLogoutHandler(BaseHandler):
    @tornado.web.authenticated
    async def put(self):
        try:
            await self.mount_client.auth_logout()
            self.finish()
        except Exception as e:
            get_logger().error("Error logging out of auth.", exc_info=True)
            raise tornado.web.HTTPError(
                status_code=500, reason=f"Error logging out of auth: {e}."
            )


class HealthHandler(BaseHandler):
    @tornado.web.authenticated
    async def get(self):
        try:
            response = await self.mount_client.health()
            self.finish(response)
        except Exception as e:
            get_logger().error("Mount server not running.")
            raise tornado.web.HTTPError(
                status_code=500, reason=f"Mount server not running."
            )


def setup_handlers(web_app):
    get_logger().info(f"Using PFS_MOUNT_DIR={PFS_MOUNT_DIR}")
    web_app.settings["pfs_contents_manager"] = PFSContentsManager(PFS_MOUNT_DIR)
    web_app.settings["pachyderm_mount_client"] = MountServerClient(PFS_MOUNT_DIR)

    _handlers = [
        ("/repos", ReposHandler),
        ("/repos/_unmount", ReposUnmountHandler),
        (r"/repos/([^/]+)", RepoHandler),
        (r"/repos/(.+)/_mount", RepoMountHandler),
        (r"/repos/(.+)/_unmount", RepoUnmountHandler),
        (r"/repos/(.+)/_commit", RepoCommitHandler),
        (r"/pfs%s" % path_regex, PFSHandler),
        ("/config", ConfigHandler),
        ("/auth/_login", AuthLoginHandler),
        ("/auth/_logout", AuthLogoutHandler),
        ("/health", HealthHandler),
    ]

    base_url = web_app.settings["base_url"]
    handlers = [
        (url_path_join(base_url, NAMESPACE, VERSION, endpoint), handler)
        for endpoint, handler in _handlers
    ]
    host_pattern = ".*$"
    web_app.add_handlers(host_pattern, handlers)
