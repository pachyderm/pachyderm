from pathlib import Path

AUTH_TOKEN_ENV = "PACH_PYTHON_AUTH_TOKEN"
OIDC_TOKEN_ENV = "PACH_PYTHON_OIDC_TOKEN"
ENTERPRISE_CODE_ENV = "PACH_PYTHON_ENTERPRISE_CODE"
PACH_CONFIG_ENV = "PACH_CONFIG"
PACHD_SERVICE_HOST_ENV = "PACHD_PEER_SERVICE_HOST"
PACHD_SERVICE_PORT_ENV = "PACHD_PEER_SERVICE_PORT"
WORKER_PORT_ENV = "PPS_WORKER_GRPC_PORT"

CONFIG_PATH_SPOUT = Path("/").joinpath("pachctl", "config.json")
CONFIG_PATH_LOCAL = Path.home().joinpath(".pachyderm", "config.json")

MAX_RECEIVE_MESSAGE_SIZE = 20 * 1024**2  # 20MB
GRPC_CHANNEL_OPTIONS = [
    ("grpc.max_receive_message_length", MAX_RECEIVE_MESSAGE_SIZE),
]
