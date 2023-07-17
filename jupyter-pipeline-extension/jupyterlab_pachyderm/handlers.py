from jupyter_server.base.handlers import APIHandler, path_regex
from jupyter_server.services.contents.handlers import ContentsHandler, validate_model
from jupyter_server.utils import url_path_join, ensure_async
import tornado
import traceback

from .env import PFS_MOUNT_DIR
from .filemanager import PFSContentsManager
from .log import get_logger
from .pachyderm import MountInterface
from .mount_server_client import MountServerClient
from .pps_client import PPSClient


# Frontend hard codes this in src/handler.ts
NAMESPACE = "pachyderm"
VERSION = "v2"


class BaseHandler(APIHandler):
    @property
    def mount_client(self) -> MountInterface:
        return self.settings["pachyderm_mount_client"]

    @property
    def pps_client(self) -> PPSClient:
        return self.settings["pachyderm_pps_client"]


class ReposHandler(BaseHandler):
    # The following decorator should be present on all verb methods (head, get, post,
    # patch, put, delete, options) to ensure only authorized user can request the
    # Jupyter server
    @tornado.web.authenticated
    async def get(self):
        try:
            response = await self.mount_client.list_repos()
            get_logger().debug(f"Repos: {response}")
            self.finish(response)
        except Exception as e:
            get_logger().error("Error listing repos.", exc_info=True)
            raise tornado.web.HTTPError(
                status_code=getattr(e, "code", 500), reason=f"Error listing repos: {e}."
            )


class MountsHandler(BaseHandler):
    @tornado.web.authenticated
    async def get(self):
        try:
            response = await self.mount_client.list_mounts()
            get_logger().debug(f"Mounts: {response}")
            self.finish(response)
        except Exception as e:
            get_logger().error("Error listing mounts.", exc_info=True)
            raise tornado.web.HTTPError(
                status_code=getattr(e, "code", 500), reason=f"Error listing mounts: {e}."
            )


class ProjectsHandler(BaseHandler):
    @tornado.web.authenticated
    async def get(self):
        try:
            response = await self.mount_client.list_projects()
            get_logger().debug(f"Projects: {response}")
            self.finish(response)
        except Exception as e:
            get_logger().error("Error listing projects.", exc_info=True)
            raise tornado.web.HTTPError(
                status_code=getattr(e, "code", 500), reason=f"Error listing projects: {e}."
            )


class MountHandler(BaseHandler):
    @tornado.web.authenticated
    async def put(self):
        try:
            body = self.get_json_body()
            response = await self.mount_client.mount(body)
            get_logger().debug(f"Mount: {response}")
            self.finish(response)
        except Exception as e:
            get_logger().error(f"Error mounting {body}.", exc_info=True)
            raise tornado.web.HTTPError(
                status_code=getattr(e, "code", 500), reason=f"Error mounting {body}: {e}."
            )


class UnmountHandler(BaseHandler):
    @tornado.web.authenticated
    async def put(self):
        try:
            body = self.get_json_body()
            response = await self.mount_client.unmount(body)
            get_logger().debug(f"Unmount: {response}")
            self.finish(response)
        except Exception as e:
            get_logger().error(f"Error unmounting {body}.", exc_info=True)
            raise tornado.web.HTTPError(
                status_code=getattr(e, "code", 500), reason=f"Error unmounting {body}: {e}.",
            )


class CommitHandler(BaseHandler):
    @tornado.web.authenticated
    async def post(self):
        try:
            body = self.get_json_body()
            response = await self.mount_client.commit(body)
            get_logger().debug(f"Commit: {response}")
            self.finish(response)
        except Exception as e:
            get_logger().error(f"Error committing {body}.", exc_info=True)
            raise tornado.web.HTTPError(
                status_code=getattr(e, "code", 500), reason=f"Error committing {body}: {e}.",
            )


class UnmountAllHandler(BaseHandler):
    """Unmounts all repos"""

    @tornado.web.authenticated
    async def put(self):
        try:
            response = await self.mount_client.unmount_all()
            get_logger().debug(f"Unmount all: {response}")
            self.finish(response)
        except Exception as e:
            get_logger().error("Error unmounting all.", exc_info=True)
            raise tornado.web.HTTPError(
                status_code=getattr(e, "code", 500), reason=f"Error unmounting all: {e}."
            )


class MountDatumsHandler(BaseHandler):
    @tornado.web.authenticated
    async def put(self):
        try:
            body = self.get_json_body()
            response = await self.mount_client.mount_datums(body)
            get_logger().debug(f"Mount datums: {response}")
            self.finish(response)
        except Exception as e:
            get_logger().error(f"Error mounting datums with input {body}", exc_info=True)
            raise tornado.web.HTTPError(
                status_code=getattr(e, "code", 500), reason=f"Error mounting datums with input {body}: {e}."
            )


class ShowDatumHandler(BaseHandler):
    @tornado.web.authenticated
    async def put(self):
        try:
            slug = {
                "idx": self.get_argument("idx", None),
                "id": self.get_argument("id", None)
            }
            response = await self.mount_client.show_datum(slug)
            get_logger().debug(f"Show datum: {response}")
            self.finish(response)
        except Exception as e:
            get_logger().error(f"Error showing datum {slug}", exc_info=True)
            raise tornado.web.HTTPError(
                status_code=getattr(e, "code", 500), reason=f"Error showing datum {slug}: {e}."
            )


class DatumsHandler(BaseHandler):
    @tornado.web.authenticated
    async def get(self):
        try:
            response = await self.mount_client.get_datums()
            get_logger().debug(f"Datums info: {response}")
            self.finish(response)
        except Exception as e:
            get_logger().error("Error getting datum info.", exc_info=True)
            raise tornado.web.HTTPError(
                status_code=getattr(e, "code", 500), reason=f"Error getting datum info: {e}."
            )


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
            get_logger().debug(f"Health: {response}")
            self.finish(response)
        except Exception:
            get_logger().error("Mount server not running.")
            raise tornado.web.HTTPError(
                status_code=500, reason=f"Mount server not running."
            )


class PPSCreateHandler(BaseHandler):

    @tornado.web.authenticated
    async def get(self, path):
        """Get the pipeline spec for the specified notebook."""
        try:
            response = await self.pps_client.generate(path)
            get_logger().debug(f"GetPipelineSpec: {response}")
            await self.finish(response)
        except Exception as e:
            if isinstance(e, tornado.web.HTTPError):
                # Common case: only way to print the "reason" field of HTTPError
                get_logger().error(f"couldn't create pipeline: {e.reason}")
            get_logger().error("\n".join(traceback.format_exception(
              type(e), value=e, tb=None)))
            raise e

    @tornado.web.authenticated
    async def put(self, path):
        """Create the pipeline for the specified notebook."""
        try:
            body = self.get_json_body()
            response = await self.pps_client.create(path, body)
            get_logger().debug(f"CreatePipeline: {response}")
            await self.finish(response)
        except Exception as e:
            if isinstance(e, tornado.web.HTTPError):
                # Common case: only way to print the "reason" field of HTTPError
                get_logger().error(f"couldn't create pipeline: {e.reason}")
            get_logger().error("\n".join(traceback.format_exception(
              type(e), value=e, tb=None)))
            raise e

def setup_handlers(web_app):
    get_logger().info(f"Using PFS_MOUNT_DIR={PFS_MOUNT_DIR}")
    web_app.settings["pfs_contents_manager"] = PFSContentsManager(PFS_MOUNT_DIR)
    web_app.settings["pachyderm_mount_client"] = MountServerClient(PFS_MOUNT_DIR)
    web_app.settings["pachyderm_pps_client"] = PPSClient()

    _handlers = [
        ("/repos", ReposHandler),
        ("/mounts", MountsHandler),
        ("/projects", ProjectsHandler),
        ("/_mount", MountHandler),
        ("/_unmount", UnmountHandler),
        ("/_commit", CommitHandler),
        ("/_unmount_all", UnmountAllHandler),
        ("/_mount_datums", MountDatumsHandler),
        (r"/_show_datum", ShowDatumHandler),
        (r"/pfs%s" % path_regex, PFSHandler),
        ("/datums", DatumsHandler),
        ("/config", ConfigHandler),
        ("/auth/_login", AuthLoginHandler),
        ("/auth/_logout", AuthLogoutHandler),
        ("/health", HealthHandler),
        (r"/pps/_create%s" % path_regex, PPSCreateHandler),
    ]

    base_url = web_app.settings["base_url"]
    handlers = [
        (url_path_join(base_url, NAMESPACE, VERSION, endpoint), handler)
        for endpoint, handler in _handlers
    ]
    host_pattern = ".*$"
    web_app.add_handlers(host_pattern, handlers)
