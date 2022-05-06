#!/usr/bin/env python3

import os

import boto3
import minio
import urllib3

ADDRESS = os.environ["S2_HOST_ADDRESS"]
NETLOC = os.environ["S2_HOST_NETLOC"]
SCHEME = os.environ["S2_HOST_SCHEME"]
ACCESS_KEY = os.environ["S2_ACCESS_KEY"]
SECRET_KEY = os.environ["S2_SECRET_KEY"]

EXPECTED_KEYS = set([
    ("small", 1),
    ("large", 65*1024*1024),
])

def test_boto_lib():
    client = boto3.client(
        "s3",
        endpoint_url=ADDRESS,
        aws_access_key_id=ACCESS_KEY,
        aws_secret_access_key=SECRET_KEY,
    )

    client.create_bucket(Bucket="test-boto-lib")

    client.upload_file("../testdata/small.txt", "test-boto-lib", "small")
    client.upload_file("../testdata/large.txt", "test-boto-lib", "large")

    res = client.list_objects_v2(Bucket="test-boto-lib")
    assert not res["IsTruncated"]
    assert set((k["Key"], k["Size"]) for k in res["Contents"]) == EXPECTED_KEYS

    with open("../testdata/small.txt", "rb") as f:
        assert client.get_object(Bucket="test-boto-lib", Key="small")["Body"].read() == f.read()
    with open("../testdata/large.txt", "rb") as f:
        assert client.get_object(Bucket="test-boto-lib", Key="large")["Body"].read() == f.read()

    client.delete_object(Bucket="test-boto-lib", Key="small")
    client.delete_object(Bucket="test-boto-lib", Key="large")
    client.delete_bucket(Bucket="test-boto-lib")

def test_minio_lib():
    pool_manager = urllib3.PoolManager(
        timeout=30.0,
        retries=urllib3.Retry(total=1),
    )

    client = minio.Minio(
        NETLOC,
        access_key=ACCESS_KEY,
        secret_key=SECRET_KEY,
        secure=SCHEME == "https",
        http_client=pool_manager,
    )

    client.make_bucket("test-minio-python-lib")

    with open("../testdata/small.txt", "rb") as f:
        client.put_object("test-minio-python-lib", "small", f, 1)
    with open("../testdata/large.txt", "rb") as f:
        client.put_object("test-minio-python-lib", "large", f, 65*1024*1024)

    res = client.list_objects_v2("test-minio-python-lib")
    assert set((o.object_name, o.size) for o in res) == EXPECTED_KEYS

    with open("../testdata/small.txt", "rb") as f:
        assert client.get_object("test-minio-python-lib", "small").read() == f.read()
    with open("../testdata/large.txt", "rb") as f:
        assert client.get_object("test-minio-python-lib", "large").read() == f.read()

    client.remove_object("test-minio-python-lib", "small")
    client.remove_object("test-minio-python-lib", "large")
    client.remove_bucket("test-minio-python-lib")
