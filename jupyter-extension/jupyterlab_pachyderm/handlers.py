from jupyter_server.base.handlers import APIHandler, path_regex
from jupyter_server.services.contents.handlers import ContentsHandler, validate_model
from jupyter_server.utils import url_path_join
import asyncio
import grpc
import json
from pachyderm_sdk import Client, errors
import tornado
import traceback

from .log import get_logger
from .pfs_manager import PFSManager, DatumManager
from .pps_client import PPSClient


# Frontend hard codes this in src/handler.ts
NAMESPACE = "pachyderm"
VERSION = "v2"


class BaseHandler(APIHandler):
    @property
    def client(self) -> Client:
        return self.settings["pachyderm_client"]

    @property
    def pfs_manager(self) -> PFSManager:
        return self.settings["pfs_contents_manager"]

    @property
    def datum_manager(self) -> DatumManager:
        return self.settings["datum_contents_manager"]

    @property
    def pps_client(self) -> PPSClient:
        return self.settings["pachyderm_pps_client"]


class MountsHandler(BaseHandler):
    @tornado.web.authenticated
    async def get(self):
        try:
            mounts = self.pfs_manager.list_mounts()
            response = json.dumps(mounts)
            get_logger().debug(f"Mounts: {response}")
            self.finish(response)
        except Exception as e:
            get_logger().error("Error listing mounts.", exc_info=True)
            raise tornado.web.HTTPError(
                status_code=getattr(e, "code", 500),
                reason=f"Error listing mounts: {e}.",
            )


class ProjectsHandler(BaseHandler):
    @tornado.web.authenticated
    async def get(self):
        try:
            projects = self.client.pfs.list_project()
            response = json.dumps([p.to_dict() for p in projects])
            get_logger().debug(f"Projects: {response}")
            self.finish(response)
        except Exception as e:
            get_logger().error("Error listing projects.", exc_info=True)
            raise tornado.web.HTTPError(
                status_code=getattr(e, "code", 500),
                reason=f"Error listing projects: {e}.",
            )


class MountHandler(BaseHandler):
    @tornado.web.authenticated
    async def put(self):
        try:
            body = self.get_json_body()
            for m in body["mounts"]:
                repo = m["repo"]
                branch = m["branch"]
                project = m["project"] if "project" in m else "default"
                name = m["name"] if "name" in m else None
                self.pfs_manager.mount_repo(
                    repo=repo, branch=branch, project=project, name=name
                )
            response = self.pfs_manager.list_mounts()
            get_logger().debug(f"Mount: {response}")
            self.finish(response)
        except ValueError as e:
            get_logger().debug(f"Bad mount request {body}: {e}")
            raise tornado.web.HTTPError(
                status_code=400, reason=f"Bad mount request: {e}"
            )
        except Exception as e:
            get_logger().error(f"Error mounting {body}.", exc_info=True)
            raise tornado.web.HTTPError(
                status_code=getattr(e, "code", 500),
                reason=f"Error mounting {body}: {e}.",
            )


class UnmountHandler(BaseHandler):
    @tornado.web.authenticated
    async def put(self):
        try:
            body = self.get_json_body()
            for name in body["mounts"]:
                self.pfs_manager.unmount_repo(name=name)
            response = self.pfs_manager.list_mounts()
            get_logger().debug(f"Unmount: {response}")
            self.finish(response)
        except ValueError as e:
            get_logger().debug(f"Bad unmount request {body}: {e}")
            raise tornado.web.HTTPError(
                status_code=400, reason=f"Bad unmount request: {e}"
            )
        except Exception as e:
            get_logger().error(f"Error unmounting {body}.", exc_info=True)
            raise tornado.web.HTTPError(
                status_code=getattr(e, "code", 500),
                reason=f"Error unmounting {body}: {e}.",
            )


# only used in tests now
class UnmountAllHandler(BaseHandler):
    """Unmounts all repos"""

    @tornado.web.authenticated
    async def put(self):
        try:
            self.pfs_manager.unmount_all()
            response = self.pfs_manager.list_mounts()
            get_logger().debug(f"Unmount all: {response}")
            self.finish(response)
        except Exception as e:
            get_logger().error("Error unmounting all.", exc_info=True)
            raise tornado.web.HTTPError(
                status_code=getattr(e, "code", 500),
                reason=f"Error unmounting all: {e}.",
            )


class MountDatumsHandler(BaseHandler):
    @tornado.web.authenticated
    async def put(self):
        try:
            body = self.get_json_body()
            self.datum_manager.mount_datums(input_dict=body)
            response = self.datum_manager.datum_state()
            get_logger().debug(f"Mount datums: {response}")
            self.finish(response)
        except Exception as e:
            get_logger().error(
                f"Error mounting datums with input {body}", exc_info=True
            )
            raise tornado.web.HTTPError(
                status_code=getattr(e, "code", 500),
                reason=f"Error mounting datums with input {body}: {e}.",
            )


class DatumNextHandler(BaseHandler):
    @tornado.web.authenticated
    async def put(self):
        try:
            self.datum_manager.next_datum()
            response = self.datum_manager.datum_state()
            get_logger().debug(f"Next datum: {response}")
            self.finish(response)
        except Exception as e:
            get_logger().error(f"Error mounting next datum", exc_info=True)
            raise tornado.web.HTTPError(
                status_code=getattr(e, "code", 500),
                reason=f"Error mounting next datum: {e}.",
            )


class DatumPrevHandler(BaseHandler):
    @tornado.web.authenticated
    async def put(self):
        try:
            self.datum_manager.prev_datum()
            response = self.datum_manager.datum_state()
            get_logger().debug(f"Prev datum: {response}")
            self.finish(response)
        except Exception as e:
            get_logger().error(f"Error mounting prev datum", exc_info=True)
            raise tornado.web.HTTPError(
                status_code=getattr(e, "code", 500),
                reason=f"Error mounting prev datum: {e}.",
            )


class DatumDownloadHandler(BaseHandler):
    @tornado.web.authenticated
    async def put(self):
        try:
            self.datum_manager.download()
            self.finish()
        except ValueError as e:
            get_logger().error(f"Invalid datum download: {e}")
            raise tornado.web.HTTPError(
                status_code=400, reason=f"Bad download request: {e}"
            )
        except Exception as e:
            get_logger().error(f"Error downloading datum", exc_info=True)
            raise tornado.web.HTTPError(
                status_code=getattr(e, "code", 500),
                reason=f"Error downloading datum: {e}.",
            )


class DatumsHandler(BaseHandler):
    @tornado.web.authenticated
    async def get(self):
        try:
            response = self.datum_manager.current_datum()
            get_logger().debug(f"Datums info: {response}")
            self.finish(response)
        except Exception as e:
            get_logger().error("Error getting datum info.", exc_info=True)
            raise tornado.web.HTTPError(
                status_code=getattr(e, "code", 500),
                reason=f"Error getting datum info: {e}.",
            )


class PFSHandler(ContentsHandler):
    @property
    def pfs_manager(self) -> PFSManager:
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
        content = bool(int(content))

        model = self.pfs_manager.get(
            path=path,
            type=type,
            format=format,
            content=content,
        )
        validate_model(model, expect_content=content)
        self._finish_model(model, location=False)


class ViewDatumHandler(ContentsHandler):
    @property
    def datum_manager(self) -> DatumManager:
        return self.settings["datum_contents_manager"]

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

        model = self.datum_manager.get(
            path=path,
            type=type,
            format=format,
            content=content,
        )
        validate_model(model, expect_content=content)
        self._finish_model(model, location=False)


# TODO: see about writing to/from config file
class ConfigHandler(BaseHandler):
    CLUSTER_AUTH_ENABLED = "AUTH_ENABLED"
    CLUSTER_AUTH_DISABLED = "AUTH_DISABLED"
    CLUSTER_INVALID = "INVALID"

    def config_response(self) -> bytes:
        if not self.client:
            return json.dumps({"cluster_status": self.CLUSTER_INVALID})

        try:
            self.client.auth.who_am_i()
            cluster_status = self.CLUSTER_AUTH_ENABLED
        except grpc.RpcError as err:
            if err.code() == grpc.StatusCode.UNAUTHENTICATED:
                cluster_status = self.CLUSTER_AUTH_ENABLED
            elif (
                err.code() == grpc.StatusCode.UNIMPLEMENTED
                and "the auth service is not activated" in err.details()
            ):
                cluster_status = self.CLUSTER_AUTH_DISABLED
            else:
                cluster_status = self.CLUSTER_INVALID
        except errors.AuthServiceNotActivated:
            cluster_status = self.CLUSTER_AUTH_DISABLED
        except ConnectionError:
            cluster_status = self.CLUSTER_INVALID

        return json.dumps(
            {"cluster_status": cluster_status, "pachd_address": self.client.address}
        )

    @tornado.web.authenticated
    async def put(self):
        try:
            body = self.get_json_body()
            address = body["pachd_address"]
            cas = bytes(body["server_cas"], "utf-8") if "server_cas" in body else None

            client = Client().from_pachd_address(
                pachd_address=address, root_certs=cas
            )
            self.settings["pachyderm_client"] = client
            self.settings["pfs_contents_manager"] = PFSManager(client=client)
            self.settings["datum_contents_manager"] = DatumManager(client=client)
            self.settings["pachyderm_pps_client"] = PPSClient(client=client)

            response = self.config_response()
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
            response = self.config_response()
            self.finish(response)
        except Exception as e:
            get_logger().error("Error getting config.", exc_info=True)
            raise tornado.web.HTTPError(
                status_code=500, reason=f"Error getting config: {e}."
            )


class AuthLoginHandler(BaseHandler):
    async def get_token(self, oidc_state: str):
        token = self.client.auth.authenticate(oidc_state=oidc_state).pach_token
        self.settings["pachyderm_client"].auth_token = token
        self.settings["pfs_contents_manager"] = PFSManager(client=self.client)
        self.settings["datum_contents_manager"] = DatumManager(client=self.client)
        self.settings["pachyderm_pps_client"] = PPSClient(client=self.client)

    @tornado.web.authenticated
    async def put(self):
        try:
            oidc_response = self.client.auth.get_oidc_login()
            asyncio.create_task(self.get_token(oidc_response.state))
            response = oidc_response.to_json()
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
            self.client.auth_token = None
            self.settings["pfs_contents_manager"] = PFSManager(client=self.client)
            self.settings["datum_contents_manager"] = DatumManager(client=self.client)
            self.settings["pachyderm_pps_client"] = PPSClient(client=self.client)
            self.finish()
        except Exception as e:
            get_logger().error("Error logging out of auth.", exc_info=True)
            raise tornado.web.HTTPError(
                status_code=500, reason=f"Error logging out of auth: {e}."
            )


class HealthHandler(BaseHandler):
    @tornado.web.authenticated
    async def get(self):
        response = {"status": "running"}
        self.finish(response)


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
            get_logger().error(
                "\n".join(traceback.format_exception(type(e), value=e, tb=None))
            )
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
            get_logger().error(
                "\n".join(traceback.format_exception(type(e), value=e, tb=None))
            )
            raise e


def setup_handlers(web_app):
    try:
        client = Client().from_config()
        get_logger().debug(
            f"Created Pachyderm client for {client.address} from local config"
        )
    except FileNotFoundError:
        client = Client()
        get_logger().debug(
            "Could not find config file, creating localhost Pachyderm client"
        )

    web_app.settings["pachyderm_client"] = client
    web_app.settings["pachyderm_pps_client"] = PPSClient(client=client)
    web_app.settings["pfs_contents_manager"] = PFSManager(client=client)
    web_app.settings["datum_contents_manager"] = DatumManager(client=client)

    _handlers = [
        ("/mounts", MountsHandler),
        ("/projects", ProjectsHandler),
        ("/_mount", MountHandler),
        ("/_unmount", UnmountHandler),
        ("/_unmount_all", UnmountAllHandler),
        ("/datums/_mount", MountDatumsHandler),
        ("/datums/_next", DatumNextHandler),
        ("/datums/_prev", DatumPrevHandler),
        ("/datums/_download", DatumDownloadHandler),
        ("/datums", DatumsHandler),
        (r"/pfs%s" % path_regex, PFSHandler),
        (r"/view_datum%s" % path_regex, ViewDatumHandler),
        ("/config", ConfigHandler),
        ("/auth/_login", AuthLoginHandler),
        ("/auth/_logout", AuthLogoutHandler),
        (r"/pps/_create%s" % path_regex, PPSCreateHandler),
    ]

    base_url = web_app.settings["base_url"]
    handlers = [
        (url_path_join(base_url, NAMESPACE, VERSION, endpoint), handler)
        for endpoint, handler in _handlers
    ]
    host_pattern = ".*$"
    web_app.add_handlers(host_pattern, handlers)
