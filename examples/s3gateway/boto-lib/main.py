#!/usr/bin/env python3

import argparse

import boto3

def main(address, auth_token):
    client = boto3.client(
        "s3",
        endpoint_url=address,
        aws_access_key_id=auth_token,
        aws_secret_access_key=auth_token,
    )

    print(client.list_buckets())

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--address", default="http://localhost:30600", help="the s3 gateway hostname:port")
    parser.add_argument("--auth-token", default="", help="the Pachyderm auth token, which should only be set if Pachyderm auth is enabled")
    args = parser.parse_args()
    main(args.address, args.auth_token)
