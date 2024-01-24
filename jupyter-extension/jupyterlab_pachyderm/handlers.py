from base64 import b64encode
from pathlib import Path
import json
import traceback
from typing import Optional

import grpc
import grpc.aio
from jupyter_server.base.handlers import APIHandler, path_regex
from jupyter_server.services.contents.handlers import ContentsHandler, validate_model
from jupyter_server.utils import url_path_join
import os
from pachyderm_sdk import Client, errors
from pachyderm_sdk.api.auth import AuthenticateRequest, AuthenticateResponse
from pachyderm_sdk.config import ConfigFile, Context
import tornado
import tornado.concurrent
import tornado.web

from .env import PACH_CONFIG
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

    @client.setter
    def client(self, new_client: Client) -> None:
        self.settings["pachyderm_client"] = new_client
        self.settings["pfs_contents_manager"] = PFSManager(client=new_client)
        self.settings["datum_contents_manager"] = DatumManager(client=new_client)
        self.settings["pachyderm_pps_client"] = PPSClient(client=new_client)

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


class ConfigHandler(BaseHandler):
    CLUSTER_INVALID = "INVALID"
    CLUSTER_VALID_NO_AUTH = "VALID_NO_AUTH"
    CLUSTER_VALID_LOGGED_IN = "VALID_LOGGED_IN"
    CLUSTER_VALID_LOGGED_OUT = "VALID_LOGGED_OUT"

    @property
    def cluster_status(self) -> str:
        if not self.client:
            return self.CLUSTER_INVALID
        try:
            self.client.auth.who_am_i()
        except grpc.RpcError as err:
            err: grpc.Call
            if err.code() == grpc.StatusCode.UNAUTHENTICATED:
                return self.CLUSTER_VALID_LOGGED_OUT
            elif (
                err.code() == grpc.StatusCode.UNIMPLEMENTED
                and "the auth service is not activated" in err.details()
            ):
                return self.CLUSTER_VALID_NO_AUTH
            else:
                return self.CLUSTER_INVALID
        except errors.AuthServiceNotActivated:
            return self.CLUSTER_VALID_NO_AUTH
        except ConnectionError:
            return self.CLUSTER_INVALID
        else:
            return self.CLUSTER_VALID_LOGGED_IN

    @tornado.web.authenticated
    async def put(self):
        # Validate input.
        body = self.get_json_body()
        address = body.get("pachd_address")
        if not address:
            get_logger().error("_config/put: no pachd address provided")
            raise tornado.web.HTTPError(500, "no pachd address provided")
        cas = bytes(body["server_cas"], "utf-8") if "server_cas" in body else None

        # Attempt to instantiate client and test connection
        try:
            self.client = Client.from_pachd_address(address, root_certs=cas)
            cluster_status = self.cluster_status
            get_logger().info(f"({address}) cluster status: {cluster_status}")
        except Exception as e:
            get_logger().error(
                f"Error updating config with endpoint {body['pachd_address']}.",
                exc_info=True,
            )
            raise tornado.web.HTTPError(
                status_code=500,
                reason=f"Error updating config with endpoint {body['pachd_address']}: {e}.",
            )

        if cluster_status != self.CLUSTER_INVALID:
            # Attempt to write new pachyderm context to config.
            try:
                write_config(self.client.address, self.client.root_certs, None)
            except RuntimeError as e:
                get_logger().error(f"Error writing local config: {e}.", exc_info=True)

        payload = {"cluster_status": cluster_status, "pachd_address": self.client.address}
        await self.finish(json.dumps(payload))

    @tornado.web.authenticated
    async def get(self):
        try:
            payload = {
                "cluster_status": self.cluster_status,
                "pachd_address": self.client.address,
            }
            await self.finish(json.dumps(payload))
        except Exception as e:
            get_logger().error("Error getting config.", exc_info=True)
            raise tornado.web.HTTPError(
                status_code=500, reason=f"Error getting config: {e}."
            )


class AuthLoginHandler(BaseHandler):

    @tornado.web.authenticated
    async def put(self):
        try:
            # Note: The auth workflow is for the backend to initiate the process by
            # calling auth.get_oidc_login() which returns a url for the user to login
            # with, and a state token that the backend can use to complete the workflow.
            # Therefore, we send the url to the frontend for the user to login with and
            # wait until the server has created a new session token, which we then retrieve.
            oidc_response = self.client.auth.get_oidc_login()
            response = oidc_response.to_json()
            await self.finish(response)

            # Unfortunately, there doesn't seem to be a way for us to use the synchronous
            # version of client.auth.authenticate because it is blocking -- using it blocks
            # the entire server until either the user successfully logs in or the OIDC login
            # attempt times out. So, we need to manually do this Authenticate asynchronously.
            async with grpc.aio.insecure_channel(self.client.address) as channel:
                async_authenticate = channel.unary_unary(
                    "/auth_v2.API/Authenticate",
                    request_serializer=AuthenticateRequest.SerializeToString,
                    response_deserializer=AuthenticateResponse.FromString,
                )
                try:
                    response = await async_authenticate(
                        AuthenticateRequest(oidc_state=oidc_response.state),
                    )
                except grpc.RpcError as err:
                    # If the user opens several login attempts, we don't have a clear
                    # mechanism to cull them after they have successfully logged in.
                    # When the call inevitably times out, we don't raise and log a warning.
                    details = err.details()
                    if "OIDC state token expired" in details:
                        get_logger().warning(details)
                        return
                    raise err
                token = response.pach_token
            self.client.auth_token = token

            # Attempt to write new pachyderm context to config.
            try:
                write_config(self.client.address, self.client.root_certs, token)
            except RuntimeError as e:
                get_logger().error(f"Error updating local config: {e}.", exc_info=True)
                raise tornado.web.HTTPError(500, f"Error updating local config: {e}.")

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


class ExploreDownloadHandler(BaseHandler):
    @tornado.web.authenticated
    async def put(self, path):
        try:
            self.pfs_manager.download_file(path=path)
        except FileExistsError:
            raise tornado.web.HTTPError(
                status_code=400,
                reason=f"Downloading {Path(path).name} which already exists in the current working directory",
            )
        except ValueError as e:
            raise tornado.web.HTTPError(status_code=400, reason=e)
        except Exception as e:
            raise tornado.web.HTTPError(status_code=500, reason=e)
        await self.finish()


class TestDownloadHandler(BaseHandler):
    @tornado.web.authenticated
    async def put(self, path):
        try:
            self.datum_manager.download_file(path=path)
        except FileExistsError:
            raise tornado.web.HTTPError(
                status_code=400,
                reason=f"Downloading {Path(path).name} which already exists in the current working directory",
            )
        except ValueError as e:
            raise tornado.web.HTTPError(status_code=400, reason=e)
        except Exception as e:
            raise tornado.web.HTTPError(status_code=500, reason=e)
        await self.finish()


def write_config(
    pachd_address: str, server_cas: Optional[bytes], session_token: Optional[str],
) -> None:
    """Writes the pachd_address/server_cas context to the local config file.
    This will create a new config file if one does not exist.

    Parameters
    ----------
    pachd_address : str
        The address to the pachd instance.
    server_cas : bytes, optional
        The certs used to establish TLS connections to the pachd instance.
    session_token : str, optional
        The pachyderm token used to authenticate your user session.

    Raises
    ------
    RuntimeError: If unable to parse an already existing config file.
    """
    if server_cas is not None:
        # server_cas are base64 encoded strings in the config file.
        server_cas = b64encode(server_cas).decode("utf-8")
    context = Context(
        pachd_address=pachd_address,
        session_token=session_token,
        server_cas=server_cas,
    )
    name = f"jupyter-{pachd_address}"
    if PACH_CONFIG.exists():
        try:
            config = ConfigFile.from_path(PACH_CONFIG)
        except Exception:
            raise RuntimeError(f"failed to load config file: {PACH_CONFIG}")
        config.add_context(name, context, overwrite=True)
    else:
        config = ConfigFile.new_with_context(name, context)
        os.makedirs(PACH_CONFIG.parent, exist_ok=True)
    config.write(PACH_CONFIG)


def setup_handlers(web_app):
    try:
        client = Client().from_config(PACH_CONFIG)
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
        (r"/download/explore%s" % path_regex, ExploreDownloadHandler),
        (r"/download/test%s" % path_regex, TestDownloadHandler),
    ]

    base_url = web_app.settings["base_url"]
    handlers = [
        (url_path_join(base_url, NAMESPACE, VERSION, endpoint), handler)
        for endpoint, handler in _handlers
    ]
    host_pattern = ".*$"
    web_app.add_handlers(host_pattern, handlers)
