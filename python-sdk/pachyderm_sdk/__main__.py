from argparse import ArgumentParser

from .client import Client
from .errors import RpcError


def check_connection():
    parser = ArgumentParser()
    parser.add_argument(
        "address", nargs="?", type=str, help="Address of the pachd instance."
    )
    args = parser.parse_args()

    if args.address:
        client = Client.from_pachd_address(args.address)
    else:
        client = Client.from_config()
    address = client.address

    try:
        client.get_version()
    except RpcError:
        print(f"Failed Connection: {address}")
        exit(1)
    else:
        print(f"Successful Connection: {address}")
        exit(0)
