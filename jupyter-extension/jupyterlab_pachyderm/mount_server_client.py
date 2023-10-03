import subprocess
import time
import platform
import json
import asyncio
import os
import pycurl
from pathlib import Path

from tornado.httpclient import AsyncHTTPClient, HTTPClientError
from tornado import locks

from .pachyderm import MountInterface
from .log import get_logger
from .env import SIDECAR_MODE, MOUNT_SERVER_LOG_DIR, NONPRIV_CONTAINER, HTTP_UNIX_SOCKET_SCHEMA, HTTP_SCHEMA, DEFAULT_SCHEMA

lock = locks.Lock()
MOUNT_SERVER_PORT = 9002

class MountServerClient(MountInterface):
    """Client interface for the mount-server backend."""

    def __init__(
        self,
        mount_dir: str,
        sock_path: str,
        ):
        
        self.mount_dir = mount_dir

        if DEFAULT_SCHEMA == HTTP_UNIX_SOCKET_SCHEMA:
            self.address = 'http://localhost'
            AsyncHTTPClient.configure('tornado.curl_httpclient.CurlAsyncHTTPClient')
            self.sock_path = sock_path
        else:
            self.address = f"http://localhost:{MOUNT_SERVER_PORT}"
            self.sock_path = ""
        self.client = AsyncHTTPClient()
        # non-prived container flag (set via -e NONPRIV_CONTAINER=1)
        # TODO: Would be preferable to auto-detect this, but unclear how
        self.nopriv = NONPRIV_CONTAINER

    async def _is_mount_server_running(self):
        get_logger().debug("Checking if mount server running...")
        try:
            await self.health()
        except Exception as e:
            get_logger().debug(f"Unable to hit server at {self.address}")
            get_logger().debug(e)
            return False
        get_logger().debug(f"Able to hit server at {self.address}")
        return True

    def _unmount(self):
        if platform.system() == "Linux":
            subprocess.run(["bash", "-c", f"fusermount -uzq {self.mount_dir}"])
        else:
            subprocess.run(
                ["bash", "-c", f"umount {self.mount_dir}"],
                stdout=subprocess.DEVNULL,
                stderr=subprocess.DEVNULL,
            )

    async def _ensure_mount_server(self):
        """
        When we first start up, we might not have auth configured. So just try
        re-launching the mount-server on every command! If the port is already
        bound, it will exit straight away. If it's not, it might start up
        successfully with the updated config.
        """
        # TODO: add --socket and --log-file stdout args
        # TODO: add better error handling
        if await self._is_mount_server_running():
            return True

        if SIDECAR_MODE:
            get_logger().debug("Kubernetes is responsible for running mount server")
            return False

        get_logger().info("Starting mount server...")
        async with lock:
            if not await self._is_mount_server_running():
                self._unmount()
                mount_server_cmd = f"mount-server --mount-dir {self.mount_dir}"
                if self.nopriv:
                    # Cannot mount in non-privileged container, so use unshare for a private mount
                    get_logger().info("Non-privileged container...")
                    subprocess.run(['mkdir','-p', f'/mnt{self.mount_dir}'])
                    mount_server_cmd = f"unshare -Ufirm mount-server --mount-dir /mnt{self.mount_dir} --allow-other=false"

                if self.sock_path:
                    mount_server_cmd += f" --sock-path {self.sock_path}"

                if MOUNT_SERVER_LOG_DIR is not None and MOUNT_SERVER_LOG_DIR:
                    mount_server_cmd += f" >> {MOUNT_SERVER_LOG_DIR} 2>&1"

                get_logger().debug(f"Starting {mount_server_cmd} ")
                mount_process = subprocess.Popen(
                    [
                        "bash",
                        "-c",
                        "set -o pipefail; "
                        + mount_server_cmd
                    ]
                )

                tries = 0
                get_logger().debug("Waiting for mount server...")
                while not await self._is_mount_server_running():
                    time.sleep(1)
                    tries += 1

                    if tries == 10:
                        get_logger().debug("Unable to start mount server...")
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
        if DEFAULT_SCHEMA == HTTP_UNIX_SOCKET_SCHEMA:
            response = await self.client.fetch(f"{self.address}/{resource}", prepare_curl_callback=lambda curl: curl.setopt(pycurl.UNIX_SOCKET_PATH, self.sock_path))
        else:
            response = await self.client.fetch(f"{self.address}/{resource}")
        return response

    async def _put(self, resource, body):
        if DEFAULT_SCHEMA == HTTP_UNIX_SOCKET_SCHEMA:
            response = await self.client.fetch(f"{self.address}/{resource}", method="PUT", body=json.dumps(body), prepare_curl_callback=lambda curl: curl.setopt(pycurl.UNIX_SOCKET_PATH, self.sock_path))
        else:
            response = await self.client.fetch(f"{self.address}/{resource}", method="PUT", body=json.dumps(body),)
        return response

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
