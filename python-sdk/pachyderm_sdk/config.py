"""Functionality for parsing Pachyderm config files."""
import json
import os
from base64 import b64decode
from dataclasses import dataclass
from pathlib import Path
from typing import Dict, Optional, Union

from .errors import ConfigError


class ConfigFile:
    """A parsed Pachyderm config file."""
    def __init__(self, config_file: Union[Path, str]):
        """
        Parameters
        ----------
        config_file : str
            The path to the config file.
        """
        config_file = Path(os.path.expanduser(config_file)).resolve()
        self._config_file_data = json.loads(config_file.read_bytes())

    @property
    def user_id(self) -> str:
        """The user ID of the config file."""
        return self._config_file_data["user_id"]

    @property
    def active_context(self) -> "Context":
        """The currently-active context."""
        active_context_name = self._config_file_data["v2"]["active_context"]
        contexts = self._config_file_data["v2"]["contexts"]
        if active_context_name not in contexts:
            raise ConfigError(f"active context not found: {active_context_name}")
        return Context(**contexts[active_context_name])

    @property
    def active_enterprise_context(self) -> "Context":
        """The currently-active context that is enterprise-enabled."""
        context_name = self._config_file_data["v2"].get("active_enterprise_context")
        if context_name is None:
            raise ConfigError("active enterprise context is not specified")
        contexts = self._config_file_data["v2"]["contexts"]
        if context_name not in contexts:
            raise ConfigError(f"active enterprise context not found: {context_name}")
        return Context(**contexts[context_name])


@dataclass
class Context:
    """A context contains all the information needed to connect to a pachyderm
    instance. Not all fields need to be present for the context to be valid."""

    source: Optional[int] = None
    """An integer that specifies where the config came from. 
    This parameter is for internal use only and should not be modified."""

    pachd_address: Optional[str] = None
    """A host:port specification for connecting to pachd."""

    server_cas: Optional[str] = None
    """Trusted root certificates for the cluster, formatted as a 
    base64-encoded PEM. This is only set when TLS is enabled."""

    session_token: Optional[str] = None
    """A secret token identifying the current user within their pachyderm
    cluster. This is included in all RPCs and used to determine if a user's
    actions are authorized. This is only set when auth is enabled."""

    active_transaction: Optional[str] = None
    """The currently active transaction for batching together commands."""

    cluster_name: Optional[str] = None
    """The name of the underlying Kubernetes cluster."""

    auth_info: Optional[str] = None
    """The name of the underlying Kubernetes cluster’s auth credentials"""

    namespace: Optional[str] = None
    """The underlying Kubernetes cluster’s namespace"""

    cluster_deployment_id: Optional[str] = None
    """The pachyderm cluster deployment ID that is used to ensure the
    operations run on the expected cluster."""

    project: Optional[str] = None

    enterprise_server: bool = False
    """Whether the context represents an enterprise server."""

    port_forwarders: Dict[str, int] = None
    """A mapping of service name -> local port."""

    @property
    def active_pachd_address(self) -> str:
        """This pachd address factors in port-forwarding."""
        if self.pachd_address is None:
            port = 30650
            if self.port_forwarders:
                port = self.port_forwarders.get("pachd", 30650)
            return f"grpc://localhost:{port}"
        return self.pachd_address

    @property
    def server_cas_decoded(self) -> Optional[bytes]:
        """The base64 decoded root certificates in PEM format, if they exist."""
        if self.server_cas:
            return b64decode(bytes(self.server_cas, "utf-8"))
