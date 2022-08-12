import os
import json
from base64 import b64decode
from pathlib import Path
from typing import Optional, TextIO
from urllib.parse import urlparse

import grpc

from .api.admin_v2 import ApiStub as _AdminStub
from .api.auth_v2 import ApiStub as _AuthStub
from .api.debug_v2 import DebugStub as _DebugStub
from .api.enterprise_v2 import ApiStub as _EnterpriseStub
from .api.identity_v2 import ApiStub as _IdentityStub
from .api.license_v2 import ApiStub as _LicenseStub
from .api.pfs_v2 import ApiStub as _PfsStub
from .api.pps_v2 import ApiStub as _PpsStub
from .api.transaction_v2 import ApiStub as _TransactionStub
from .api.versionpb_v2 import ApiStub as _VersionStub

from .errors import AuthServiceNotActivated, BadClusterDeploymentID, ConfigError
from .interceptor import MetadataClientInterceptor, MetadataType


MB = 1024**2
MAX_RECEIVE_MESSAGE_SIZE = 20 * MB
GRPC_CHANNEL_OPTIONS = [
    ("grpc.max_receive_message_length", MAX_RECEIVE_MESSAGE_SIZE),
]


class Client:
    """The :class:`.Client` class that users will primarily interact with.
    Initialize an instance with ``python_pachyderm.Client()``.

    To see documentation on the methods :class:`.Client` can call, refer to the
    `mixins` module.
    """

    # Class variables for checking config
    env_config = "PACH_CONFIG"
    spout_config = "/pachctl/config.json"
    local_config = f"{Path.home()}/.pachyderm/config.json"

    def __init__(
        self,
        host: str = None,
        port: int = None,
        auth_token: str = None,
        root_certs: bytes = None,
        transaction_id: str = None,
        tls: bool = None,
        use_default_host: bool = True,
    ):
        """
        Creates a Pachyderm client. If host and port are unset, checks the
        ``PACH_CONFIG`` env var for a path. If that's unset, it checks two
        file paths for a config file. If both files don't exist, a client
        with default settings is created.

        Parameters
        ----------
        host : str, optional
            The pachd host. Default is 'localhost', which is used with
            ``pachctl port-forward``.
        port : int, optional
            The port to connect to. Default is 30650.
        auth_token : str, optional
            The authentication token. Used if authentication is enabled on the
            cluster.
        root_certs : bytes, optional
            The PEM-encoded root certificates as byte string.
        transaction_id : str, optional
            The ID of the transaction to run operations on.
        tls : bool, optional
            Whether TLS should be used. If `root_certs` are specified, they are
            used. Otherwise, we use the certs provided by certifi.
        use_default_host : bool, optional
            Whether to replicate `pachctl` behavior of searching for config.

        Examples
        --------
        >>> client = python_pachyderm.Client()
        ...
        >>> # Manually set host and port
        >>> client = python_pachyderm.Client("pachd.example.com", 12345)
        """

        # replicate pachctl behavior to searching for config
        # if host and port are unset
        if host is None and port is None and use_default_host:
            config = Client._check_for_config()

            if config is not None:
                (
                    host,
                    port,
                    _,
                    auth_token,
                    root_certs,
                    transaction_id,
                    tls,
                ) = Client._parse_config(config)

        host = host or "localhost"
        port = port or 30650

        if auth_token is None:
            auth_token = os.environ.get("PACH_PYTHON_AUTH_TOKEN")

        if tls is None:
            tls = root_certs is not None
        if tls and root_certs is None:
            # load default certs if none are specified
            import certifi

            with open(certifi.where(), "rb") as f:
                root_certs = f.read()

        self.address = "{}:{}".format(host, port)
        self.root_certs = root_certs
        channel = _create_channel(
            self.address, self.root_certs, options=GRPC_CHANNEL_OPTIONS
        )

        self._stubs = {}
        self._auth_token = auth_token
        self._transaction_id = transaction_id
        self._metadata = self._build_metadata()
        self._channel = _apply_metadata_interceptor(channel, self._metadata)
        if not auth_token and os.environ.get("PACH_PYTHON_OIDC_TOKEN"):
            resp = self.authenticate_id_token(os.environ.get("PACH_PYTHON_OIDC_TOKEN"))
            self._auth_token = resp
            self._metadata = self._build_metadata()
            self._channel = _apply_metadata_interceptor(channel, self._metadata)

        self.admin = _AdminStub(channel)
        self.auth = _AuthStub(channel)
        self.debug = _DebugStub(channel)
        self.enterprise = _EnterpriseStub(channel)
        self.identity = _IdentityStub(channel)
        self.license = _LicenseStub(channel)
        self.pfs = _PfsStub(channel)
        self.pps = _PpsStub(channel)
        self.transaction = _TransactionStub(channel)
        self.version = _VersionStub(channel)

    @classmethod
    def new_in_cluster(
        cls, auth_token: str = None, transaction_id: str = None
    ) -> "Client":
        """Creates a Pachyderm client that operates within a Pachyderm cluster.

        Parameters
        ----------
        auth_token : str, optional
            The authentication token. Used if authentication is enabled on the
            cluster.
        transaction_id : str, optional
            The ID of the transaction to run operations on.

        Returns
        -------
        Client
            A python_pachyderm client instance.

        Examples
        --------
        >>> from python_pachyderm import Client
        >>> client = Client.new_in_cluster()
        """

        if (
            "PACHD_PEER_SERVICE_HOST" in os.environ
            and "PACHD_PEER_SERVICE_PORT" in os.environ
        ):
            # Try to use the pachd peer service if it's available. This is
            # only supported in pachyderm>=1.10, but is more reliable because
            # it'll work when TLS is enabled on the cluster.
            host = os.environ["PACHD_PEER_SERVICE_HOST"]
            port = int(os.environ["PACHD_PEER_SERVICE_PORT"])
        else:
            # Otherwise use the normal service host/port, which will not work
            # when TLS is enabled on the cluster.
            host = os.environ["PACHD_SERVICE_HOST"]
            port = int(os.environ["PACHD_SERVICE_PORT"])

        return cls(
            host=host,
            port=port,
            auth_token=auth_token,
            transaction_id=transaction_id,
            use_default_host=False,
        )

    @classmethod
    def new_from_pachd_address(
        cls,
        pachd_address: str,
        auth_token: str = None,
        root_certs: bytes = None,
        transaction_id: str = None,
    ) -> "Client":
        """Creates a Pachyderm client from a given pachd address.

        Parameters
        ----------
        pachd_address : str
            The address of pachd server
        auth_token : str, optional
            The authentication token. Used if authentication is enabled on the
            cluster.
        root_certs : bytes, optional
            The PEM-encoded root certificates as byte string. If unspecified,
            this will load default certs from certifi.
        transaction_id : str, optional
            The ID of the transaction to run operations on.

        Returns
        -------
        Client
            A python_pachyderm client instance.

        Examples
        --------
        >>> from python_pachyderm import Client
        >>> client = Client.new_from_pachd_address("grpc://pachyderm.com:80/")
        ...
        >>> client = Client.new_from_pachd_address("https://pachyderm.com:80", root_certs=b"foo")

        .. # noqa: W505
        """

        u = Client._parse_address(pachd_address)

        return cls(
            host=u.hostname,
            port=u.port,
            auth_token=auth_token,
            root_certs=root_certs,
            transaction_id=transaction_id,
            tls=u.scheme == "grpcs" or u.scheme == "https",
            use_default_host=False,
        )

    @classmethod
    def new_from_config(cls, config_file: TextIO) -> "Client":
        """Creates a Pachyderm client from a config file-like object.

        Parameters
        ----------
        config_file : TextIO
            A file-like object containing the config json file.

        Returns
        -------
        Client
            A python_pachyderm client instance.

        Examples
        --------
        >>> from python_pachyderm import Client
        >>> config = '''{
        ...   "v2": {
        ...     "active_context": "local",
        ...     "contexts": {
        ...       "local": {
        ...         "pachd_address": "grpcs://172.17.0.6:30650",
        ...         "server_cas": "foo",
        ...         "session_token": "bar",
        ...         "active_transaction": "baz"
        ...       }
        ...     }
        ...   }
        ... }'''
        >>> client = Client.new_from_config(io.StringIO(config))
        """

        if config_file is None:
            raise ConfigError("no config object provided")

        config = json.load(config_file)
        (
            _,
            _,
            pachd_address,
            auth_token,
            root_certs,
            transaction_id,
            _,
        ) = cls._parse_config(config)

        client = cls.new_from_pachd_address(
            pachd_address,
            auth_token=auth_token,
            root_certs=root_certs,
            transaction_id=transaction_id,
        )

        context = cls._get_active_context(config)
        expected_deployment_id = context.get("cluster_deployment_id")
        if expected_deployment_id:
            cluster_info = client.inspect_cluster()
            if cluster_info.deployment_id != expected_deployment_id:
                raise BadClusterDeploymentID(
                    expected_deployment_id, cluster_info.deployment_id
                )

        return client

    @staticmethod
    def _check_for_config():
        """Checks for Pachyderm config file locally."""

        j = Client._check_pach_config_env_var()
        if j is not None:
            return j

        j = Client._check_pach_config_spout()
        if j is not None:
            return j

        j = Client._check_pach_config_local()
        if j is not None:
            return j

        print("no config found, proceeding with default behavior")

        return j

    @staticmethod
    def _check_pach_config_env_var():
        j = None
        if Client.env_config in os.environ:
            with open(os.environ.get(Client.env_config), "r") as config_file:
                j = json.load(config_file)

        return j

    @staticmethod
    def _check_pach_config_spout():
        j = None
        if os.path.isfile(Client.spout_config):
            with open(Client.spout_config, "r") as config_file:
                j = json.load(config_file)

        return j

    @staticmethod
    def _check_pach_config_local():
        j = None
        if os.path.isfile(Client.local_config):
            with open(Client.local_config, "r") as config_file:
                j = json.load(config_file)

        return j

    @staticmethod
    def _parse_address(pachd_address):
        if "://" not in pachd_address:
            pachd_address = "grpc://{}".format(pachd_address)

        u = urlparse(pachd_address)

        if u.scheme not in ("grpc", "http", "grpcs", "https"):
            raise ValueError("unrecognized pachd address scheme: {}".format(u.scheme))
        if u.path != "" or u.params != "" or u.query != "" or u.fragment != "":
            raise ValueError("invalid pachd address")
        if u.username is not None or u.password is not None:
            raise ValueError("invalid pachd address")

        return u

    @staticmethod
    def _get_active_context(config):
        try:
            active_context = config["v2"]["active_context"]
        except KeyError:
            raise ConfigError("no active context")

        try:
            context = config["v2"]["contexts"][active_context]
        except KeyError:
            raise ConfigError("missing active context '{}'".format(active_context))

        return context

    @staticmethod
    def _parse_config(config):
        context = Client._get_active_context(config)

        auth_token = context.get("session_token")
        root_certs = context.get("server_cas")
        transaction_id = context.get("active_transaction")

        pachd_address = context.get("pachd_address")
        if not pachd_address:
            port_forwarders = context.get("port_forwarders", {})
            pachd_port = port_forwarders.get("pachd", 30650)
            pachd_address = "grpc://localhost:{}".format(pachd_port)

            root_certs = None

        u = Client._parse_address(pachd_address)

        host = u.hostname
        port = u.port
        tls = u.scheme == "grpcs" or u.scheme == "https"
        root_certs = (
            b64decode(bytes(root_certs, "utf-8")) if root_certs is not None else None
        )

        return host, port, pachd_address, auth_token, root_certs, transaction_id, tls

    @property
    def auth_token(self):
        return self._auth_token

    @auth_token.setter
    def auth_token(self, value):
        self._auth_token = value
        self._metadata = self._build_metadata()
        self._channel = _apply_metadata_interceptor(
            channel=_create_channel(
                self.address, self.root_certs, options=GRPC_CHANNEL_OPTIONS
            ),
            metadata=self._metadata,
        )
        super().__init__()

    @property
    def transaction_id(self):
        return self._transaction_id

    @transaction_id.setter
    def transaction_id(self, value):
        self._transaction_id = value
        self._metadata = self._build_metadata()
        self._channel = _apply_metadata_interceptor(
            channel=_create_channel(
                self.address, self.root_certs, options=GRPC_CHANNEL_OPTIONS
            ),
            metadata=self._metadata,
        )
        super().__init__()

    def _build_metadata(self):
        metadata = []
        if self._auth_token is not None:
            metadata.append(("authn-token", self._auth_token))
        if self._transaction_id is not None:
            metadata.append(("pach-transaction", self._transaction_id))
        return metadata

    def delete_all(self) -> None:
        """Delete all repos, commits, files, pipelines, and jobs. This resets
        the cluster to its initial state.
        """
        # Try removing all identities if auth is activated.
        try:
            self.delete_all_identity()
        except AuthServiceNotActivated:
            pass

        # Try deactivating auth if activated.
        try:
            self.deactivate_auth()
        except AuthServiceNotActivated:
            pass

        # Try removing all licenses if auth is activated.
        try:
            self.delete_all_license()
        except AuthServiceNotActivated:
            pass

        self.delete_all_pipelines()
        self.delete_all_repos()
        self.delete_all_transactions()


def _apply_metadata_interceptor(
    channel: grpc.Channel, metadata: MetadataType
) -> grpc.Channel:
    metadata_interceptor = MetadataClientInterceptor(metadata)
    return grpc.intercept_channel(channel, metadata_interceptor)


def _create_channel(
    address: str,
    root_certs: Optional[bytes],
    options: MetadataType,
) -> grpc.Channel:
    if root_certs is not None:
        ssl = grpc.ssl_channel_credentials(root_certificates=root_certs)
        return grpc.secure_channel(address, ssl, options=options)
    return grpc.insecure_channel(address, options=options)
