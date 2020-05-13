#!/usr/bin/env python3

import argparse

from minio import Minio

def main(address, auth_token, secure):
    client = Minio(address, access_key=auth_token, secret_key=auth_token, secure=secure)
    print(client.list_buckets())

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--address", default="localhost:30600", help="the s3 gateway hostname:port")
    parser.add_argument("--auth-token", default="", help="the Pachyderm auth token, which should only be set if Pachyderm auth is enabled")
    parser.add_argument("--secure", action="store_true", help="connect to the s3 gateway via TLS")
    args = parser.parse_args()
    main(args.address, args.auth_token, args.secure)
