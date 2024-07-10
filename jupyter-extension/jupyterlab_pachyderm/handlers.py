from base64 import b64encode
from pathlib import Path
import json
import traceback
from typing import Optional
from urllib.parse import urlparse

import grpc
import grpc.aio
from jupyter_server.base.handlers import APIHandler, path_regex
from jupyter_server.services.contents.handlers import ContentsHandler, validate_model
from jupyter_server.utils import url_path_join
import os
from pachyderm_sdk import Client, errors
from pachyderm_sdk.api import pfs
from pachyderm_sdk.api.auth import AuthenticateRequest, AuthenticateResponse
from pachyderm_sdk.config import ConfigFile, Context
import tornado
import tornado.concurrent
import tornado.web

from .log import get_logger
from .pfs_manager import PFSManager, DatumManager
from .pps_client import PPSClient


# Frontend hard codes this in src/handler.ts
NAMESPACE = "pachyderm"
VERSION = "v2"


class BaseHandler(APIHandler):
    _no_client_error = tornado.web.HTTPError(
        status_code=401, reason="no instantiated pachyderm client"
    )

    HEALTHCHECK_UNHEALTHY = "UNHEALTHY"
    HEALTHCHECK_INVALID_CLUSTER = "HEALTHY_INVALID_CLUSTER"
    HEALTHCHECK_NO_AUTH = "HEALTHY_NO_AUTH"
    HEALTHCHECK_LOGGED_IN = "HEALTHY_LOGGED_IN"
    HEALTHCHECK_LOGGED_OUT = "HEALTHY_LOGGED_OUT"

    def status(self, client: Client) -> str:
        """Determines the status of the client's connection to the cluster."""
        try:
            client.auth.who_am_i()
        except grpc.RpcError as err:
            err: grpc.Call
            if err.code() == grpc.StatusCode.UNAUTHENTICATED:
                return self.HEALTHCHECK_LOGGED_OUT
            elif (
                err.code() == grpc.StatusCode.UNIMPLEMENTED
                and "the auth service is not activated" in err.details()
            ):
                return self.HEALTHCHECK_NO_AUTH
            else:
                return self.HEALTHCHECK_INVALID_CLUSTER
        except errors.AuthServiceNotActivated:
            return self.HEALTHCHECK_NO_AUTH
        except ConnectionError:
            # Cannot connect to Pachyderm
            return self.HEALTHCHECK_INVALID_CLUSTER
        else:
            return self.HEALTHCHECK_LOGGED_IN

    @property
    def client(self) -> Client:
        client = self.settings.get("pachyderm_client")
        if client is None:
            raise self._no_client_error
        return client

    @client.setter
    def client(self, new_client: Client) -> None:
        # server_root_dir is the root path from which all data and notebook files
        #   are relative. This may be different from the CWD.
        root_dir = Path(self.settings.get("server_root_dir", os.getcwd())).resolve()

        self.settings["pachyderm_client"] = new_client
        self.settings["pfs_contents_manager"] = PFSManager(client=new_client)
        self.settings["datum_contents_manager"] = DatumManager(client=new_client)
        self.settings["pachyderm_pps_client"] = PPSClient(client=new_client, root_dir=root_dir)

    @property
    def config_file(self) -> Path:
        return self.settings["pachyderm_config_file"]

    @property
    def pfs_manager(self) -> PFSManager:
        pfs_manager = self.settings.get("pfs_contents_manager")
        if pfs_manager is None:
            raise self._no_client_error
        return pfs_manager

    @property
    def datum_manager(self) -> DatumManager:
        datum_manager = self.settings.get("datum_contents_manager")
        if datum_manager is None:
            raise self._no_client_error
        return datum_manager

    @property
    def pps_client(self) -> PPSClient:
        pps_client = self.settings.get("pachyderm_pps_client")
        if pps_client is None:
            raise self._no_client_error
        return pps_client


class ReposHandler(BaseHandler):
    @tornado.web.authenticated
    async def get(self):
        try:
            mounts = self.pfs_manager.list_repos()
            response = json.dumps(mounts)
            get_logger().debug(f"Repos: {response}")
            self.finish(response)
        except Exception as e:
            get_logger().error("Error listing repos.", exc_info=True)
            raise tornado.web.HTTPError(
                status_code=getattr(e, "code", 500),
                reason=f"Error listing mounts: {e}.",
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
        except ValueError as e:
            get_logger().error(
                f"Error mounting datums with invalid input {body}", exc_info=True
            )
            raise tornado.web.HTTPError(
                status_code=getattr(e, "code", 400),
                reason=f"Error mounting datums with invalid input {body}: {e}.",
            )
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
            path = self.datum_manager.download()
            self.finish(json.dumps({"path": path}))
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
        try:
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
            pagination_marker = None
            pagination_marker_uri = self.get_query_argument(
                "pagination_marker", default=None
            )
            if pagination_marker_uri:
                pagination_marker = pfs.File.from_uri(pagination_marker_uri)
            number = int(self.get_query_argument("number", default="100"))
            model = self.pfs_manager.get(
                path=path,
                type=type,
                format=format,
                content=content,
                pagination_marker=pagination_marker,
                number=number,
            )
            validate_model(model, expect_content=content)
            self._finish_model(model, location=False)
        except ValueError as e:
            raise tornado.web.HTTPError(status_code=400, reason=repr(e))
        except Exception as e:
            raise tornado.web.HTTPError(status_code=500, reason=repr(e))

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
        pagination_marker = None
        pagination_marker_uri = self.get_query_argument(
            "pagination_marker", default=None
        )
        if pagination_marker_uri:
            pagination_marker = pfs.File.from_uri(pagination_marker_uri)
        number = int(self.get_query_argument("number", default="100"))

        model = self.datum_manager.get(
            path=path,
            type=type,
            format=format,
            content=content,
            pagination_marker=pagination_marker,
            number=number,
        )
        validate_model(model, expect_content=content)
        self._finish_model(model, location=False)


class ConfigHandler(BaseHandler):
    # TODO: should we return server CAs?
    @tornado.web.authenticated
    async def put(self):
        # Validate input.
        body = self.get_json_body()
        address = body.get("pachd_address")
        if not address:
            get_logger().error("config/put: no pachd address provided")
            raise tornado.web.HTTPError(400, "no pachd address provided")
        cas = bytes(body["server_cas"], "utf-8") if "server_cas" in body else None

        # Attempt to instantiate client and test connection
        client = Client.from_pachd_address(address, root_certs=cas)
        cluster_status = self.status(client)
        get_logger().info(f"({address}) cluster status: {cluster_status}")

        if cluster_status == self.HEALTHCHECK_UNHEALTHY:
            raise tornado.web.HTTPError(500, "An error occurred.")
        elif cluster_status == self.HEALTHCHECK_INVALID_CLUSTER:
            raise tornado.web.HTTPError(400, f"Could not connect to {address}")
        else:
            self.client = client  # Set client only if valid.
            # Attempt to write new pachyderm context to config.
            try:
                write_config(self.config_file, client.address, client.root_certs, None)
            except RuntimeError as e:
                get_logger().error(f"Error writing local config: {e}.", exc_info=True)

        payload = {"pachd_address": client.address}
        await self.finish(json.dumps(payload))

    @tornado.web.authenticated
    async def get(self):
        # Try to get a pachyderm client.
        try:
            client = self.client
        except tornado.web.HTTPError:
            payload = {"pachd_address": ""}
        else:
            payload = {"pachd_address": client.address}
        await self.finish(json.dumps(payload))


class HealthHandler(BaseHandler):
    @tornado.web.authenticated
    async def get(self):
        try:
            client = self.client
        except tornado.web.HTTPError:
            # client not instantiated, user needs to configure the Pachyderm cluster
            status = self.HEALTHCHECK_INVALID_CLUSTER
        else:
            try:
                status = self.status(client)
            except Exception as e:
                await self.finish(
                    json.dumps(
                        {
                            "status": self.HEALTHCHECK_UNHEALTHY,
                            "message": e,
                        }
                    )
                )
                return
        await self.finish(json.dumps({"status": status}))


class AuthLoginHandler(BaseHandler):
    @tornado.web.authenticated
    async def put(self):
        client = self.client
        try:
            # Note: The auth workflow is for the backend to initiate the process by
            # calling auth.get_oidc_login() which returns a url for the user to login
            # with, and a state token that the backend can use to complete the workflow.
            # Therefore, we send the url to the frontend for the user to login with and
            # wait until the server has created a new session token, which we then retrieve.
            oidc_response = client.auth.get_oidc_login()

            # The login url will always specify the http protocol. If the user is connecting
            # over a secure https/grpcs connection, we should update the url to reflect this.
            # Checking if root certificates exists on the client is the best proxy currently
            # available to check if the client is communicating over a secure grpc channel.
            if client.root_certs is not None:
                # Usage of _replace method comes from urlparse documentation:
                #   https://docs.python.org/3/library/urllib.parse.html#urllib.parse.urlparse
                # noinspection PyProtectedMember
                oidc_response.login_url = (
                    urlparse(oidc_response.login_url)._replace(scheme="https").geturl()
                )

            response = oidc_response.to_json()
            await self.finish(response)

            # Unfortunately, there doesn't seem to be a way for us to use the synchronous
            # version of client.auth.authenticate because it is blocking -- using it blocks
            # the entire server until either the user successfully logs in or the OIDC login
            # attempt times out. So, we need to manually do this Authenticate asynchronously.
            async with grpc.aio.insecure_channel(client.address) as channel:
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
            client.auth_token = token

            # Attempt to write new pachyderm context to config.
            try:
                write_config(self.config_file, client.address, client.root_certs, token)
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
            await self.finish()
        except Exception as e:
            get_logger().error("Error logging out of auth.", exc_info=True)
            raise tornado.web.HTTPError(
                status_code=500, reason=f"Error logging out of auth: {e}."
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
        except ValueError as e:
            get_logger().error(f"bad pipeline spec: {e}")
            raise tornado.web.HTTPError(
                status_code=400, reason=f"Bad pipeline spec: {e}"
            )
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
            raise tornado.web.HTTPError(status_code=400, reason=repr(e))
        except Exception as e:
            raise tornado.web.HTTPError(status_code=500, reason=repr(e))
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
            raise tornado.web.HTTPError(status_code=400, reason=repr(e))
        except Exception as e:
            raise tornado.web.HTTPError(status_code=500, reason=repr(e))
        await self.finish()

class ExploreMountHandler(BaseHandler):
    @tornado.web.authenticated
    async def put(self):
        try:
            body = self.get_json_body()
            if not body:
                raise ValueError("commit_uri does not exist in body of request")
            commit_uri = body.get("commit_uri")
            if commit_uri is None:
                raise ValueError("commit_uri does not exist in body of request")
            commit = pfs.Commit.from_uri(commit_uri)
            self.pfs_manager.mount_commit(commit=commit)
        except ValueError as e:
            raise tornado.web.HTTPError(status_code=400, reason=repr(e))
        except Exception as e:
            raise tornado.web.HTTPError(status_code=500, reason=repr(e))
        await self.finish()

class ExploreUnmountHandler(BaseHandler):
    @tornado.web.authenticated
    async def put(self):
        try:
            self.pfs_manager.unmount_commit()
        except ValueError as e:
            raise tornado.web.HTTPError(status_code=400, reason=repr(e))
        except Exception as e:
            raise tornado.web.HTTPError(status_code=500, reason=repr(e))
        await self.finish()


def write_config(
    config_file: Path,
    pachd_address: str,
    server_cas: Optional[bytes],
    session_token: Optional[str],
) -> None:
    """Writes the pachd_address/server_cas context to a local config file.
    This will create a new config file if file does not already exist.

    Parameters
    ----------
    config_file : pathlib.Path
        The location to write the config file.
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
    if config_file.exists():
        try:
            config = ConfigFile.from_path(config_file)
        except Exception:
            raise RuntimeError(f"failed to load config file: {config_file}")
        config.add_context(name, context, overwrite=True)
    else:
        config = ConfigFile.new_with_context(name, context)
        os.makedirs(config_file.parent, exist_ok=True)
    config.write(config_file)


def setup_handlers(
    web_app, config_file: Path, pachd_address: str = None, dex_token: str = None
):
    """
    Sets up the Pachyderm client and the HTTP request handler.

    Config for the Pachyderm client will first be attempted by reading
    the local config file. This falls back to the PACHD_ADDRESS and
    DEX_TOKEN env vars, and finally defaulting to a localhost client
    on the default port 30650 failing that.
    """
    client = None
    try:
        client = Client().from_config(config_file)
        get_logger().debug(
            f"Created Pachyderm client for {client.address} from local config"
        )
    except FileNotFoundError:
        if pachd_address:
            client = Client().from_pachd_address(pachd_address=pachd_address)
            if dex_token:
                client.auth_token = client.auth.authenticate(
                    id_token=dex_token
                ).pach_token
            get_logger().debug(
                f"Created Pachyderm client for {client.address} from env var"
            )
            # Attempt to write new pachyderm context to config.
            try:
                write_config(config_file, client.address, client.root_certs, None)
            except RuntimeError as e:
                get_logger().error(f"Error writing local config: {e}.", exc_info=True)
        else:
            get_logger().debug(
                "Could not find config file -- no pachyderm client instantiated"
            )

    # server_root_dir is the root path from which all data and notebook files
    #   are relative. This may be different from the CWD.
    root_dir = Path(web_app.settings.get("server_root_dir", os.getcwd())).resolve()
    if client:
        web_app.settings["pachyderm_client"] = client
        web_app.settings["pachyderm_pps_client"] = PPSClient(client=client, root_dir=root_dir)
        web_app.settings["pfs_contents_manager"] = PFSManager(client=client)
        web_app.settings["datum_contents_manager"] = DatumManager(client=client)

    # This value is used by the AuthHandler and ConfigHandler.
    web_app.settings["pachyderm_config_file"] = config_file

    _handlers = [
        ("/repos", ReposHandler),
        ("/explore/mount", ExploreMountHandler),
        ("/explore/unmount", ExploreUnmountHandler),
        ("/datums/_mount", MountDatumsHandler),
        ("/datums/_next", DatumNextHandler),
        ("/datums/_prev", DatumPrevHandler),
        ("/datums/_download", DatumDownloadHandler),
        ("/datums", DatumsHandler),
        (r"/pfs%s" % path_regex, PFSHandler),
        (r"/view_datum%s" % path_regex, ViewDatumHandler),
        ("/config", ConfigHandler),
        ("/health", HealthHandler),
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
