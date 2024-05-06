"""Functionality for parsing Pachyderm config files."""

import json
import os
import uuid
from base64 import b64decode
from dataclasses import dataclass, asdict
from pathlib import Path
from typing import Dict, Optional, Union

from .errors import ConfigError


class ConfigFile:
    """A parsed Pachyderm config file."""

    def __init__(self, data: dict):
        """
        Parameters
        ----------
        data : dict (JSON)
            The config file data. The only required fields are:
            {
              "v2": {
                "active_context": ...,
                "contexts": {...}
              }
            }
        """
        if (v2 := data.get("v2")) is None or not (
            "active_context" in v2 and "contexts" in v2
        ):
            raise ValueError(
                "Missing required fields v2.active_context and/or v2.contexts."
            )
        self._config_file_data = data

    @classmethod
    def from_path(cls, config_file: Union[Path, str]) -> "ConfigFile":
        """Parse a Pachyderm config file from a local file.

        Parameters
        ----------
        config_file : str
            The path to the config file.
        """
        config_file = Path(os.path.expanduser(config_file)).resolve()
        return cls(data=json.loads(config_file.read_bytes()))

    @classmethod
    def new_with_context(cls, name: str, context: "Context") -> "ConfigFile":
        """Create a new ConfigFile object from a single context.
        This will also set this specified context as the active context.

        Parameters
        ----------
        name : str
            The name of the context.
        context : Context
            The context object.
        """
        data = dict(
            user_id=str(uuid.uuid4()).replace("-", ""),
            v2=dict(
                active_context=name,
                contexts={name: asdict(context)},
                metrics=True,
            ),
        )
        return cls(data)

    @property
    def user_id(self) -> Optional[str]:
        """The user ID of the config file."""
        return self._config_file_data.get("user_id")

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

    def add_context(
        self, name: str, context: "Context", *, overwrite: bool = False
    ) -> None:
        """Add a context to the parsed config.
        The context name must be unique, unless overwrite is True.

        Parameters
        ----------
        name : str
            The name of the context.
        context : Context
            The context object.
        overwrite : bool (kwarg)
            Whether to overwrite an existing context.
        """
        contexts = self._config_file_data["v2"]["contexts"]
        if (name in contexts) and not overwrite:
            raise ValueError(f'context "{name}" already exists')
        contexts[name] = asdict(context)

    def set_active_context(self, name: str) -> None:
        """Set the active context to the specified name.
        The context must exist, else a ValueError is raise.

        Parameters
        ----------
        name : str
            The name of the context.
        """
        contexts = self._config_file_data["v2"]["contexts"]
        if name not in contexts:
            raise ValueError(f'context "{name}" does not exist')
        self._config_file_data["v2"]["active_context"] = name

    def write(self, config_file: Union[Path, str]) -> None:
        """Write the current config data to the specified file.

        Parameters
        ----------
        config_file : str
            The path to the config file.
        """
        config_file = Path(os.path.expanduser(config_file)).resolve()
        with config_file.open("w") as file_out:
            file_out.write(json.dumps(self._config_file_data, indent=2))
            file_out.write("\n")


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
