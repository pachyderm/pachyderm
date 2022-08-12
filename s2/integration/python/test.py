#!/usr/bin/env python3

import os
from pathlib import Path

import boto3
import minio
import requests
import urllib3

ADDRESS = os.environ["S2_HOST_ADDRESS"]
NETLOC = os.environ["S2_HOST_NETLOC"]
SCHEME = os.environ["S2_HOST_SCHEME"]
ACCESS_KEY = os.environ["S2_ACCESS_KEY"]
SECRET_KEY = os.environ["S2_SECRET_KEY"]

SMALL_FILE = Path("../testdata/small.txt").resolve()
LARGE_FILE = Path("../testdata/large.txt").resolve()
EXPECTED_KEYS = set([
    ("small", SMALL_FILE.stat().st_size),
    ("large", LARGE_FILE.stat().st_size),
])

def test_boto_lib():
    client = boto3.client(
        "s3",
        endpoint_url=ADDRESS,
        aws_access_key_id=ACCESS_KEY,
        aws_secret_access_key=SECRET_KEY,
    )

    bucket_name = "test-boto-lib"
    client.create_bucket(Bucket=bucket_name)
    try:
        client.upload_file(str(SMALL_FILE), bucket_name, "small")
        client.upload_file(str(LARGE_FILE), bucket_name, "large")

        res = client.list_objects_v2(Bucket=bucket_name)
        assert not res["IsTruncated"]
        assert set((k["Key"], k["Size"]) for k in res["Contents"]) == EXPECTED_KEYS

        assert client.get_object(Bucket=bucket_name, Key="small")["Body"].read() == SMALL_FILE.read_bytes()
        assert client.get_object(Bucket=bucket_name, Key="large")["Body"].read() == LARGE_FILE.read_bytes()

    finally:
        client.delete_object(Bucket=bucket_name, Key="small")
        client.delete_object(Bucket=bucket_name, Key="large")
        client.delete_bucket(Bucket=bucket_name)

def test_minio_lib():
    pool_manager = urllib3.PoolManager(
        timeout=30.0,
        retries=urllib3.Retry(total=3),
    )

    client = minio.Minio(
        NETLOC,
        access_key=ACCESS_KEY,
        secret_key=SECRET_KEY,
        secure=SCHEME == "https",
        http_client=pool_manager,
    )

    bucket_name = "test-minio-python-lib"
    client.make_bucket(bucket_name)

    try:
        client.put_object(bucket_name, "small", SMALL_FILE.open("rb"), SMALL_FILE.stat().st_size)
        client.put_object(bucket_name, "large", LARGE_FILE.open("rb"), LARGE_FILE.stat().st_size)

        res = client.list_objects(bucket_name)
        assert set((o.object_name, o.size) for o in res) == EXPECTED_KEYS

        assert client.get_object(bucket_name, "small").data == SMALL_FILE.read_bytes()
        assert client.get_object(bucket_name, "large").data == LARGE_FILE.read_bytes()

    finally:
        if client.bucket_exists(bucket_name):
            for object in client.list_objects(bucket_name):
                    client.remove_object(bucket_name, object.object_name)
            if client.bucket_exists(bucket_name):
                client.remove_bucket(bucket_name)


def test_unquoted_etags():
    """Test that use of unquoted values in etag headers, such as "If-None-Match",
    is corrected and handled by the s2 library.
    """
    url = f"{SCHEME}://{NETLOC}"
    bucket_name = "test-etag-quoting"
    file_name = "greeting.txt"

    resp = requests.put(f"{url}/{bucket_name}")
    assert resp.status_code == 200

    response = requests.put(f"{url}/{bucket_name}/{file_name}", data="hello")
    assert resp.status_code == 200

    response = requests.head(f"{url}/{bucket_name}/{file_name}")
    assert resp.status_code == 200

    etag = response.headers["Etag"]
    if etag and etag.startswith('"'):
        etag = etag[1:]
    if etag and etag.endswith('"'):
        etag = etag[:-1]
    response = requests.get(f"{url}/{bucket_name}/{file_name}", headers={"If-None-Match": etag})
    assert response.status_code == 304, "expected 304 Not Modified"

    requests.delete(f"{url}/{bucket_name}/{file_name}")
    requests.delete(f"{url}/{bucket_name}")
