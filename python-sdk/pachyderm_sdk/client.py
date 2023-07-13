"""The Client used to interact with a Pachyderm instance."""
import contextlib
import os
from pathlib import Path
from typing import Optional, Union
from urllib.parse import urlparse

import grpc

from .api.admin.extension import ApiStub as _AdminStub
from .api.auth import ApiStub as _AuthStub
from .api.debug.extension import ApiStub as _DebugStub
from .api.enterprise import ApiStub as _EnterpriseStub
from .api.identity import ApiStub as _IdentityStub
from .api.license import ApiStub as _LicenseStub
from .api.pfs.extension import ApiStub as _PfsStub
from .api.pps.extension import ApiStub as _PpsStub
from .api.transaction.extension import ApiStub as _TransactionStub
from .api.version import ApiStub as _VersionStub, Version
from .api.worker.extension import WorkerStub as _WorkerStub
from .config import ConfigFile
from .constants import (
    AUTH_TOKEN_ENV,
    CONFIG_PATH_LOCAL,
    CONFIG_PATH_SPOUT,
    DEFAULT_HOST,
    DEFAULT_PORT,
    GRPC_CHANNEL_OPTIONS,
    OIDC_TOKEN_ENV,
    PACHD_SERVICE_HOST_ENV,
    PACHD_SERVICE_PORT_ENV,
    WORKER_PORT_ENV,
)
from .errors import AuthServiceNotActivated, BadClusterDeploymentID
from .interceptor import MetadataClientInterceptor, MetadataType

__all__ = ("Client",)


class Client:
    """The Client used to interact with a Pachyderm instance.

    Examples
    --------
    Connect to a pachyderm instance using your local config file:
    >>> from pachyderm_sdk import Client
    >>> client = Client.from_config()

    Connect to a pachyderm instance using a URL/address:
    >>> from pachyderm_sdk import Client
    >>> client = Client.from_pachd_address("test.work.com:30080")
    """

    def __init__(
        self,
        host: str = DEFAULT_HOST,
        port: int = DEFAULT_PORT,
        auth_token: Optional[str] = None,
        root_certs: Optional[bytes] = None,
        transaction_id: str = None,
        tls: bool = False,
    ):
        """
        Creates a Pachyderm client.

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
        tls : bool
            Whether TLS should be used. If `root_certs` are specified, they are
            used. Otherwise, we use the certs provided by certifi.
        """
        host = host or DEFAULT_HOST
        port = port or DEFAULT_PORT
        if auth_token is None:
            auth_token = os.environ.get(AUTH_TOKEN_ENV)

        tls = tls or (root_certs is not None)
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

        self._auth_token = auth_token
        self._transaction_id = transaction_id
        self._metadata = self._build_metadata()
        self._channel = _apply_metadata_interceptor(channel, self._metadata)

        # See implementation for api layout.
        self._init_api()
        # Worker stub is loaded when accessed through the worker property.
        self._worker = None

        if not auth_token and (oidc_token := os.environ.get(OIDC_TOKEN_ENV)):
            self.auth_token = self.auth.authenticate(id_token=oidc_token)

    def _init_api(self):
        self.admin = _AdminStub(self._channel)
        self.auth = _AuthStub(self._channel)
        self.debug = _DebugStub(self._channel)
        self.enterprise = _EnterpriseStub(self._channel)
        self.identity = _IdentityStub(self._channel)
        self.license = _LicenseStub(self._channel)
        self.pfs = _PfsStub(
            self._channel,
            get_transaction_id=lambda: self.transaction_id,
        )
        self.pps = _PpsStub(self._channel)
        self.transaction = _TransactionStub(
            self._channel,
            get_transaction_id=lambda: self.transaction_id,
            set_transaction_id=lambda value: setattr(self, "transaction_id", value),
        )
        self._version_api = _VersionStub(self._channel)
        self._worker: Optional[_WorkerStub]

    @classmethod
    def new_in_cluster(
        cls, auth_token: Optional[str] = None, transaction_id: Optional[str] = None
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
        """
        if CONFIG_PATH_SPOUT.exists():
            # TODO: Should we notify the user that we are using spout config?
            return cls.from_config(CONFIG_PATH_SPOUT)

        host = os.environ.get(PACHD_SERVICE_HOST_ENV)
        if host is None:
            raise RuntimeError(
                f"Environment variable {PACHD_SERVICE_HOST_ENV} not set "
                f"-- cannot connect. Are you running in a cluster?"
            )
        port = os.environ.get(PACHD_SERVICE_PORT_ENV)
        if port is None:
            raise RuntimeError(
                f"Environment variable {PACHD_SERVICE_PORT_ENV} not set "
                f"-- cannot connect. Are you running in a cluster?"
            )

        return cls(
            host=host,
            port=int(port),
            auth_token=auth_token,
            transaction_id=transaction_id,
        )

    @classmethod
    def from_pachd_address(
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
        """
        if "://" not in pachd_address:
            pachd_address = "grpc://{}".format(pachd_address)

        u = urlparse(pachd_address)

        if u.scheme not in ("grpc", "http", "grpcs", "https"):
            raise ValueError("unrecognized pachd address scheme: {}".format(u.scheme))
        if u.path or u.params or u.query or u.fragment or u.username or u.password:
            raise ValueError("invalid pachd address")

        return cls(
            host=u.hostname,
            port=u.port,
            auth_token=auth_token,
            root_certs=root_certs,
            transaction_id=transaction_id,
            tls=u.scheme == "grpcs" or u.scheme == "https",
        )

    @classmethod
    def from_config(cls, config_file: Union[Path, str] = CONFIG_PATH_LOCAL) -> "Client":
        """Creates a Pachyderm client from a config file.

        Parameters
        ----------
        config_file : Union[Path, str]
            The path to a config json file.
            config_file defaults to the local config.

        Returns
        -------
        Client
            A properly configured Client.
        """
        config = ConfigFile(config_file)
        active_context = config.active_context
        client = cls.from_pachd_address(
            active_context.active_pachd_address,
            auth_token=active_context.session_token,
            root_certs=active_context.server_cas_decoded,
            transaction_id=active_context.active_transaction,
        )

        # Verify the deployment ID of the active context with the cluster.
        expected_deployment_id = active_context.cluster_deployment_id
        if expected_deployment_id:
            cluster_info = client.admin.inspect_cluster()
            if cluster_info.deployment_id != expected_deployment_id:
                raise BadClusterDeploymentID(
                    expected_deployment_id, cluster_info.deployment_id
                )

        return client

    @property
    def auth_token(self):
        """The authentication token. Used if authentication is enabled on the cluster."""
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
        self._init_api()

    @property
    def transaction_id(self):
        """The ID of the transaction to run operations on."""
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
        self._init_api()

    @property
    def worker(self) -> _WorkerStub:
        """Access the worker API stub.

        This is dynamically loaded in order to provide a helpful error message
        to the user if they try to interact the worker API from outside a worker.
        """
        if self._worker is None:
            port = os.environ.get(WORKER_PORT_ENV)
            if port is None:
                raise ConnectionError(
                    f"Cannot connect to the worker since {WORKER_PORT_ENV} is not set. "
                    "Are you running inside a pipeline?"
                )
            # Note: This channel does not go through the metadata interceptor.
            channel = _create_channel(
                address=f"localhost:{port}", root_certs=None, options=GRPC_CHANNEL_OPTIONS
            )
            self._worker = _WorkerStub(channel)
        return self._worker

    def _build_metadata(self):
        metadata = []
        if self._auth_token is not None:
            metadata.append(("authn-token", self._auth_token))
        if self._transaction_id is not None:
            metadata.append(("pach-transaction", self._transaction_id))
        return metadata

    def get_version(self) -> Version:
        """Requests version information from the pachd cluster."""
        return self._version_api.get_version()


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
