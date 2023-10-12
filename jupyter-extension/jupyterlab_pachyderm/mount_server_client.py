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
    SIDECAR_MODE,
    MOUNT_SERVER_LOG_FILE,
    NONPRIV_CONTAINER,
    PFS_SOCK_PATH,
    DET_RESOURCES_TYPE,
    SLURM_JOB
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
    def initialize(self, old_resolver, socket_path, *args, **kwargs):
        self.old_resolver = old_resolver
        self.socket_path = socket_path

    def close(self):
      self.old_resolver.close()

    async def resolve(self,
        host: str, port: int, family: socket.AddressFamily = socket.AF_UNSPEC,
        *args, **kwargs
    ):
        if host == "unix":
            return [(socket.AF_UNIX, self.socket_path)]
        return await self.old_resolver.resolve(host, port, family, *args, **kwargs)


class MountServerClient(MountInterface):
    """Client interface for the mount-server backend."""

    def __init__(
        self,
        mount_dir: str,
        sock_path: str,
    ):
        self.mount_dir = mount_dir

        self.client = AsyncHTTPClient()
        if PFS_SOCK_PATH and not SIDECAR_MODE:
            self.address = f"http://unix"
            self.address_desc = (f"http://\"unix"+
            f"({PFS_SOCK_PATH})\"" if PFS_SOCK_PATH else "\"")
        else:
            self.address = f"http://localhost:{MOUNT_SERVER_PORT}"
            self.address_desc = self.address
        # non-prived container flag (set via -e NONPRIV_CONTAINER=1)
        # or use DET_RESOURCES_TYPE environment variable to auto-detect this.
        self.nopriv = NONPRIV_CONTAINER
        if DET_RESOURCES_TYPE == SLURM_JOB:
            get_logger().debug("Inferring non privileged container for launcher/MLDE...")
            self.nopriv = 1

    async def _is_mount_server_running(self):
        get_logger().debug("Checking if mount-server running...")
        try:
            await self.health()
        except Exception as e:
            get_logger().error(f"Unable to reach mount-server at {self.address_desc}: {e}")
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
                mount_server_cmd = [
                   shutil.which("mount-server"),
                   "--mount-dir", self.mount_dir,
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
                    get_logger().info("Starting mount-server in new namespace, per NONPRIV_CONTAINER ({NONPRIV_CONTAINER}) or DET_RESOURCES_TYPE ({DET_RESOURCES_TYPE}).")
                    relative_mount_dir = Path("/mnt") / Path(self.mount_dir).relative_to("/")
                    subprocess.run(['mkdir','-p', relative_mount_dir])
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
                        "unshare", "-Ufirm",
                        shutil.which("mount-server"),
                        "--mount-dir", relative_mount_dir,
                        "--allow-other=false",
                    ]

                if PFS_SOCK_PATH and not SIDECAR_MODE:
                    mount_server_cmd += ["--sock-path", PFS_SOCK_PATH]

                if MOUNT_SERVER_LOG_FILE:
                    mount_server_cmd += ["--log-file", MOUNT_SERVER_LOG_FILE]

                get_logger().info(f"Starting mount-server: \"{' '.join(mount_server_cmd)}\"")
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
                    mount_server_proc = subprocess.run(['pgrep', '-s', str(os.getsid(mount_process.pid)), 'mount-server'], capture_output=True)
                    mount_server_pid = mount_server_proc.stdout.decode("utf-8")
                    if not mount_server_pid or not int(mount_server_pid) > 0:
                        get_logger().debug(f"Unable to find mount-server process: {mount_server_pid}")
                        return False
                    mount_server_pid = int(mount_server_pid)
                    get_logger().info(f"Link non-privileged /pfs to /proc/{mount_server_pid}/root/mnt{self.mount_dir}")
                    if os.path.exists(self.mount_dir) or os.path.islink(self.mount_dir):
                        if os.path.isdir(self.mount_dir):
                            get_logger().debug(f"Removing dir {self.mount_dir}")
                            try:
                                os.rmdir(self.mount_dir)
                            except PermissionError as ex:
                                get_logger().debug(f"Removing dir {self.mount_dir} failed with {str(ex)}", exc_info=1)
                                # Make / writable so we can remove /pfs and replace with a link
                                subprocess.run(["sudo", "/usr/bin/chmod", "777","/"])
                                # Retry the removal
                                os.rmdir(self.mount_dir)
                        else:
                            get_logger().debug(f"Removing file {self.mount_dir}")
                            os.remove(self.mount_dir)
                    os.symlink( f'/proc/{mount_server_pid}/root/mnt{self.mount_dir}', '/pfs', target_is_directory=True)

        return True

    async def _get(self, resource):
        return await self.client.fetch(f"{self.address}/{resource}")

    async def _put(self, resource, body):
        return await self.client.fetch(f"{self.address}/{resource}", method="PUT", body=json.dumps(body))

    async def list_repos(self):
        await self._ensure_mount_server()
        resource = "repos"
        response = await self._get(resource)
        return response.body

    async def list_mounts(self):
        await self._ensure_mount_server()
        resource = "mounts"
        response = await self._get(resource)
        return response.body

    async def list_projects(self):
        await self._ensure_mount_server()
        resource = "projects"
        response = await self._get(resource)
        return response.body

    async def mount(self, body):
        await self._ensure_mount_server()
        resource = "_mount"
        response = await self._put(resource, body)
        return response.body

    async def unmount(self, body):
        await self._ensure_mount_server()
        resource = "_unmount"
        response = await self._put(resource, body)
        return response.body

    async def commit(self, body):
        await self._ensure_mount_server()
        pass

    async def unmount_all(self):
        await self._ensure_mount_server()
        resource = "_unmount_all"
        response = await self._put(resource, {})
        return response.body

    async def mount_datums(self, body):
        await self._ensure_mount_server()
        resource = "_mount_datums"
        response = await self._put(resource, body)
        return response.body

    async def show_datum(self, slug):
        await self._ensure_mount_server()
        slug = '&'.join(f"{k}={v}" for k,v in slug.items() if v is not None)
        resource = "_show_datum?" + slug
        response = await self._put(resource, {})
        return response.body

    async def get_datums(self):
        await self._ensure_mount_server()
        resource = "datums"
        response = await self._get(resource)
        return response.body

    async def config(self, body=None):
        await self._ensure_mount_server()
        if body is None:
            try:
                resource = "config"
                response = await self._get(resource)
            except HTTPClientError as e:
                if e.code == 404:
                    return json.dumps({"cluster_status": "INVALID"})
                raise e
        else:
            resource = "config"
            response = await self._put(resource, body)
        return response.body

    async def auth_login(self):
        await self._ensure_mount_server()
        resource = "_login"
        response = await self._put(resource, {})
        resp_json = json.loads(response.body.decode())
        # may bubble exception up to handler if oidc_state not in response
        oidc = resp_json['oidc_state']

        # we explicitly send the login_token request and do not await here.
        # the reason for this is that we want the user to be redirected to
        # the login page without awaiting the result of the login before
        # doing so.
        asyncio.create_task(self.auth_login_token(oidc))
        return response.body

    async def auth_login_token(self, oidc):
        resource = "auth/_login_token"
        response = await self._put(resource, {})
        response.rethrow()
        pach_config_path = Path.home().joinpath('.pachyderm', 'config.json')
        if pach_config_path.is_file():
            # if config already exists, need to add new context into it and
            # switch active context over
            try:
                write_token_to_config(pach_config_path, response.body.decode())
            except Exception as e:
                get_logger().warn("Failed writing session token: ", e.args)
                raise e
        else:
            # otherwise, write the entire config to file
            os.makedirs(os.path.dirname(pach_config_path), exist_ok=True)
            with open(pach_config_path, 'w') as f:
                f.write(response.body.decode())
        return response.body

    async def auth_logout(self):
        await self._ensure_mount_server()
        resource = "auth/_logout"
        response = await self._put(resource, {})
        return response

    async def health(self):
        resource = "health"
        response = await self._get(resource)
        return response.body


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

    active_context = mount_server_config['v2']['active_context']
    config['v2']['active_context'] = active_context
    if active_context in config['v2']['contexts']:
        # if config contains this context already, write token to it
        token = mount_server_config['v2']['contexts'][active_context]['session_token']
        config['v2']['contexts'][active_context]['session_token'] = token
    else:
        # otherwise, write the entire context to it
        config['v2']['contexts'][active_context] = mount_server_config['v2']['contexts'][active_context]
    with open(pach_config_path, 'w') as config_file:
        json.dump(config, config_file, indent=2)
