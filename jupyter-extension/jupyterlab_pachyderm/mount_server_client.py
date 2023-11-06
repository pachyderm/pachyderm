import subprocess
import time
import platform
import json
import asyncio
import os
import socket
import shutil
from pathlib import Path

from tornado.httpclient import AsyncHTTPClient, HTTPClientError
from tornado.netutil import Resolver
from tornado import locks

from .pachyderm import MountInterface
from .log import get_logger
from .env import (
    DET_RESOURCES_TYPE,
    MOUNT_SERVER_LOG_FILE,
    NONPRIV_CONTAINER,
    PACH_CONFIG,
    PFS_MOUNT_DIR,
    PFS_SOCK_PATH,
    SIDECAR_MODE,
    SLURM_JOB,
)

lock = locks.Lock()

# MOUNT_SERVER_PORT is the port over which MountServerClient and mount-server
# will communicate if a unix socket is not being used.
# TODO: dynamic port assignment, if needed
MOUNT_SERVER_PORT = 9002

# UNIX_SOCKET_ADDRESS is a fake address that UnixSocketResolver will resolve to
# PFS_SOCK_PATH (below). This allows MountServerClient to connect to an instance
# of mount-server that is serving over a unix socket transparently.
UNIX_SOCKET_ADDRESS = "unix"


# UnixSockerResolver is a custom Tornado resolver that allows a Tornado client
# to send/recieve traffic over a local unix socket by resolving 'unix://<path>'
# addresses to an AF_SOCK socket at the indicated path. This allows
# communication with mount-server over a unix socket, which is required in HPC
# environments.
#
# Implementation taken from
# https://github.com/tornadoweb/tornado/issues/2671#issuecomment-499190469
class UnixSocketResolver(Resolver):
    def initialize(self, old_resolver, *args, **kwargs):
        self.old_resolver = old_resolver

    def close(self):
        self.old_resolver.close()

    async def resolve(
        self,
        host: str,
        port: int,
        family: socket.AddressFamily = socket.AF_UNSPEC,
        *args,
        **kwargs,
    ):
        if host == UNIX_SOCKET_ADDRESS:
            return [(socket.AF_UNIX, PFS_SOCK_PATH)]
        return await self.old_resolver.resolve(host, port, family, *args, **kwargs)


class MountServerClient(MountInterface):
    """Client interface for the mount-server backend."""

    def __init__(self, mount_dir: str):
        self.mount_dir = mount_dir

        self.client = AsyncHTTPClient()
        if PFS_SOCK_PATH and not SIDECAR_MODE:
            self.address = f"http://" + UNIX_SOCKET_ADDRESS
            self.address_desc = (
                f'http://"{UNIX_SOCKET_ADDRESS}' + f'({PFS_SOCK_PATH})"'
                if PFS_SOCK_PATH
                else '"'
            )
        else:
            self.address = f"http://localhost:{MOUNT_SERVER_PORT}"
            self.address_desc = self.address
        self.nopriv = NONPRIV_CONTAINER or DET_RESOURCES_TYPE == SLURM_JOB

    async def _is_mount_server_running(self):
        get_logger().debug("Checking if mount-server running...")
        try:
            await self.health()
        except Exception as e:
            get_logger().error(
                f"Unable to reach mount-server (at {self.address_desc}): {e}"
            )
            return False
        get_logger().debug(f"Successfully reached mount-server at {self.address_desc}")
        return True

    def _unmount(self):
        # Our extension may be run locally, on peoples' macbooks, so also try
        # to handle unmounting on Darwin
        if platform.system() == "Linux":
            subprocess.run([shutil.which("fusermount"), "-uzq", self.mount_dir])
        else:
            subprocess.run([shutil.which("umount"), self.mount_dir])

    async def _ensure_mount_server(self):
        """
        When we first start up, we might not have auth configured. So just try
        re-launching the mount-server on every command! If the port is already
        bound, it will exit straight away. If it's not, it might start up
        successfully with the updated config.
        """
        if await self._is_mount_server_running():
            return True

        if SIDECAR_MODE:
            get_logger().debug("Kubernetes is responsible for running mount server")
            return False

        get_logger().info("Starting mount server...")
        async with lock:
            if not await self._is_mount_server_running():
                self._unmount()

                mount_server_path = shutil.which("mount-server")
                if not mount_server_path:
                    get_logger().error("Cannot locate mount-server binary")
                    return False

                mount_server_cmd = [
                    mount_server_path,
                    "--mount-dir",
                    self.mount_dir,
                ]
                if self.nopriv:
                    # Cannot mount in non-privileged container, so create a new
                    # Linux namespace (via 'unshare') in which we _are_ allowed
                    # to mount. Caveats:
                    # - This is forbidden in Kubernetes
                    # - This will simply crash on Darwin.
                    # - this._unmount does not unmount anything inside the
                    #   container. If this fails, I believe the mount-server
                    #   will not recover (todo: confirm with @jerryharrow)
                    #
                    # This is a solution for our HPC container runtimes (podman,
                    # singularity, and enroot), where we must run
                    # jupyterlab-pachyderm in an unprivileged container, but
                    # where unshare is allowed, giving us a path to making FUSE
                    # work.
                    get_logger().info(
                        "Preparing to run mount-server in new namespace, per NONPRIV_CONTAINER ({NONPRIV_CONTAINER}) or DET_RESOURCES_TYPE ({DET_RESOURCES_TYPE})."
                    )
                    relative_mount_dir = Path("/mnt") / Path(
                        self.mount_dir
                    ).relative_to("/")
                    subprocess.run(["mkdir", "-p", relative_mount_dir])
                    mount_server_cmd = [
                        # What we're unsharing:
                        # -U: unshare the (U)ser table - new namespace will have
                        #     its own user table (possibly removeable?)
                        # -f: (f)ork the command; it will run as a child of
                        #     unshare, rather than taking over the unshare proc
                        #     (possibly removeable?)
                        # -i: unshare the (i)pc namespace (possibly removeable?)
                        # -r: map the current user and group to the (r)oot user
                        #     and group in the new namespace; this allows the
                        #     mount-server to create a FUSE mount without issue
                        # -m: unshare the (m)ount namespace: this is nessary for
                        #     mount-server command to succeed in the new
                        #     namespace, as this process doesn't have the
                        #     ability to mount in the current namespace. The
                        #     mount in the new namespace will accessible from
                        #     the current namespace via symlink.
                        "unshare",
                        "-Ufirm",
                        mount_server_path,
                        "--mount-dir",
                        relative_mount_dir,
                        "--allow-other=false",
                    ]

                if PFS_SOCK_PATH and not SIDECAR_MODE:
                    # As PFS_SOCK_PATH is set by default, simply ignore it if
                    # SIDECAR_MODE is also set (they are incompatible).
                    mount_server_cmd += ["--sock-path", PFS_SOCK_PATH]

                if MOUNT_SERVER_LOG_FILE:
                    mount_server_cmd += ["--log-file", MOUNT_SERVER_LOG_FILE]

                get_logger().info(
                    'Starting mount-server: "'
                    + " ".join(map(str, mount_server_cmd))
                    + '"'
                )
                mount_process = subprocess.Popen(mount_server_cmd)

                tries = 0
                get_logger().debug("Waiting for mount server...")
                while not await self._is_mount_server_running():
                    time.sleep(1)
                    tries += 1

                    if tries == 10:
                        get_logger().error("Unable to start mount server...")
                        return False

                if self.nopriv:
                    # Using un-shared mount, replace /pfs with a softlink to the mount point
                    mount_server_proc = subprocess.run(
                        [
                            "pgrep",
                            "-s",
                            str(os.getsid(mount_process.pid)),
                            "mount-server",
                        ],
                        capture_output=True,
                    )
                    mount_server_pid = mount_server_proc.stdout.decode("utf-8")
                    if not mount_server_pid or not int(mount_server_pid) > 0:
                        get_logger().debug(
                            f"Unable to find mount-server process: {mount_server_pid}"
                        )
                        return False
                    mount_server_pid = int(mount_server_pid)
                    get_logger().info(
                        f"Link non-privileged /pfs to /proc/{mount_server_pid}/root/mnt{self.mount_dir}"
                    )
                    if os.path.exists(self.mount_dir) or os.path.islink(self.mount_dir):
                        if os.path.isdir(self.mount_dir):
                            get_logger().debug(f"Removing dir {self.mount_dir}")
                            try:
                                os.rmdir(self.mount_dir)
                            except PermissionError as ex:
                                get_logger().debug(
                                    f"Removing dir {self.mount_dir} failed with {str(ex)}",
                                    exc_info=1,
                                )
                                # Make / writable so we can remove /pfs and replace with a link
                                subprocess.run(["sudo", "/usr/bin/chmod", "777", "/"])
                                # Retry the removal
                                os.rmdir(self.mount_dir)
                        else:
                            get_logger().debug(f"Removing file {self.mount_dir}")
                            os.remove(self.mount_dir)
                    os.symlink(
                        f"/proc/{mount_server_pid}/root/mnt{self.mount_dir}",
                        "/pfs",
                        target_is_directory=True,
                    )

        return True

    async def _get(self, resource):
        await self._ensure_mount_server()
        response = await self.client.fetch(f"{self.address}/{resource}")
        return response.body

    async def _put(self, resource, body, request_timeout=None):
        await self._ensure_mount_server()
        response = await self.client.fetch(
            f"{self.address}/{resource}",
            method="PUT",
            body=json.dumps(body),
            request_timeout=request_timeout,
        )
        if resource == "auth/_login_token":
            response.rethrow()  # slight hack
        return response.body

    async def list_repos(self):
        return await self._get("repos")

    async def list_mounts(self):
        return await self._get("mounts")

    async def list_projects(self):
        return await self._get("projects")

    async def mount(self, body):
        return await self._put("_mount", body)

    async def unmount(self, body):
        return await self._put("_unmount", body)

    async def commit(self, body):
        await self._ensure_mount_server()
        pass

    async def unmount_all(self):
        return await self._put("_unmount_all", {})

    async def mount_datums(self, body):
        # TODO: request_timeout isn't passed through correctly
        return await self._put("datums/_mount", body, request_timeout=0)

    async def next_datum(self):
        return await self._put("datums/_next", {})

    async def prev_datum(self):
        return await self._put("datums/_prev", {})

    async def get_datums(self):
        return await self._get("datums")

    async def config(self, body=None):
        await self._ensure_mount_server()
        if body is None:
            try:
                return await self._get("config")
            except HTTPClientError as e:
                if e.code == 404:
                    return json.dumps({"cluster_status": "INVALID"})
                raise e
        return await self._put("config", body)

    async def auth_login(self):
        response = await self._put("auth/_login", {})
        resp_json = json.loads(response.decode())

        # we explicitly send the login_token request and do not await here.
        # the reason for this is that we want the user to be redirected to
        # the login page--if this awaits the result of logging in without ever
        # giving the user a chance to log in, this will block forever.
        #
        # (this bubbles the exception up if oidc_state is not in the response)
        oidc = resp_json["oidc_state"]
        asyncio.create_task(self.auth_login_token(oidc))
        return response

    async def auth_login_token(self, oidc):
        response = await self._put("auth/_login_token", {"oidc": oidc})
        if Path(PACH_CONFIG).is_file():
            # if config already exists, need to add new context into it and
            # switch active context over
            try:
                write_token_to_config(PACH_CONFIG, response.body.decode())
            except Exception as e:
                get_logger().warn("Failed writing session token: ", e.args)
                raise e
        else:
            # otherwise, write the entire config to file
            os.makedirs(os.path.dirname(PACH_CONFIG), exist_ok=True)
            with open(PACH_CONFIG, "w") as f:
                f.write(response.decode())
        return response

    async def auth_logout(self):
        return await self._put("auth/_logout", {})

    async def health(self):
        return await self._get("health")


def write_token_to_config(pach_config_path, mount_server_config_str):
    """
    updates the pachyderm config with the session token of the mount server
    config. this will try to insert the token into the mount_server context
    if it already exists, or copies the mount server config if it doesn't.
    then, it switches the active context over to the mount_server context.

    parameters:
        pach_config_path: the path to the pachyderm config file
        mount_server_config_str: json containing the mount server pachyderm
                                    configuration
    """
    with open(pach_config_path) as config_file:
        config = json.load(config_file)
    mount_server_config = json.loads(mount_server_config_str)

    active_context = mount_server_config["v2"]["active_context"]
    config["v2"]["active_context"] = active_context
    if active_context in config["v2"]["contexts"]:
        # if config contains this context already, write token to it
        token = mount_server_config["v2"]["contexts"][active_context]["session_token"]
        config["v2"]["contexts"][active_context]["session_token"] = token
    else:
        # otherwise, write the entire context to it
        config["v2"]["contexts"][active_context] = mount_server_config["v2"][
            "contexts"
        ][active_context]
    with open(pach_config_path, "w") as config_file:
        json.dump(config, config_file, indent=2)
