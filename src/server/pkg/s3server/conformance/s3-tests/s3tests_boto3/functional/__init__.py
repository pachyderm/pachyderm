import boto3
from botocore import UNSIGNED
from botocore.client import Config
from botocore.handlers import disable_signing
import ConfigParser
import os
import bunch
import random
import string
import itertools

config = bunch.Bunch

# this will be assigned by setup()
prefix = None

def get_prefix():
    assert prefix is not None
    return prefix

def choose_bucket_prefix(template, max_len=30):
    """
    Choose a prefix for our test buckets, so they're easy to identify.

    Use template and feed it more and more random filler, until it's
    as long as possible but still below max_len.
    """
    rand = ''.join(
        random.choice(string.ascii_lowercase + string.digits)
        for c in range(255)
        )

    while rand:
        s = template.format(random=rand)
        if len(s) <= max_len:
            return s
        rand = rand[:-1]

    raise RuntimeError(
        'Bucket prefix template is impossible to fulfill: {template!r}'.format(
            template=template,
            ),
        )

def get_buckets_list(client=None, prefix=None):
    if client == None:
        client = get_client()
    if prefix == None:
        prefix = get_prefix()
    response = client.list_buckets()
    bucket_dicts = response['Buckets']
    buckets_list = []
    for bucket in bucket_dicts:
        if prefix in bucket['Name']:
            buckets_list.append(bucket['Name'])

    return buckets_list

def get_objects_list(bucket, client=None, prefix=None):
    if client == None:
        client = get_client()

    if prefix == None:
        response = client.list_objects(Bucket=bucket)
    else:
        response = client.list_objects(Bucket=bucket, Prefix=prefix)
    objects_list = []

    if 'Contents' in response:
        contents = response['Contents']
        for obj in contents:
            objects_list.append(obj['Key'])

    return objects_list

def get_versioned_objects_list(bucket, client=None):
    if client == None:
        client = get_client()
    response = client.list_object_versions(Bucket=bucket)
    versioned_objects_list = []

    if 'Versions' in response:
        contents = response['Versions']
        for obj in contents:
            key = obj['Key']
            version_id = obj['VersionId']
            versioned_obj = (key,version_id)
            versioned_objects_list.append(versioned_obj)

    return versioned_objects_list

def get_delete_markers_list(bucket, client=None):
    if client == None:
        client = get_client()
    response = client.list_object_versions(Bucket=bucket)
    delete_markers = []

    if 'DeleteMarkers' in response:
        contents = response['DeleteMarkers']
        for obj in contents:
            key = obj['Key']
            version_id = obj['VersionId']
            versioned_obj = (key,version_id)
            delete_markers.append(versioned_obj)

    return delete_markers


def nuke_prefixed_buckets(prefix, client=None):
    if client == None:
        client = get_client()

    buckets = get_buckets_list(client, prefix)

    if buckets != []:
        for bucket_name in buckets:
            objects_list = get_objects_list(bucket_name, client)
            for obj in objects_list:
                response = client.delete_object(Bucket=bucket_name,Key=obj)
            versioned_objects_list = get_versioned_objects_list(bucket_name, client)
            for obj in versioned_objects_list:
                response = client.delete_object(Bucket=bucket_name,Key=obj[0],VersionId=obj[1])
            delete_markers = get_delete_markers_list(bucket_name, client)
            for obj in delete_markers:
                response = client.delete_object(Bucket=bucket_name,Key=obj[0],VersionId=obj[1])
            client.delete_bucket(Bucket=bucket_name)

    print('Done with cleanup of buckets in tests.')

def setup():
    cfg = ConfigParser.RawConfigParser()
    try:
        path = os.environ['S3TEST_CONF']
    except KeyError:
        raise RuntimeError(
            'To run tests, point environment '
            + 'variable S3TEST_CONF to a config file.',
            )
    with file(path) as f:
        cfg.readfp(f)

    if not cfg.defaults():
        raise RuntimeError('Your config file is missing the DEFAULT section!')
    if not cfg.has_section("s3 main"):
        raise RuntimeError('Your config file is missing the "s3 main" section!')
    if not cfg.has_section("s3 alt"):
        raise RuntimeError('Your config file is missing the "s3 alt" section!')
    if not cfg.has_section("s3 tenant"):
        raise RuntimeError('Your config file is missing the "s3 tenant" section!')

    global prefix

    defaults = cfg.defaults()

    # vars from the DEFAULT section
    config.default_host = defaults.get("host")
    config.default_port = int(defaults.get("port"))
    config.default_is_secure = cfg.getboolean('DEFAULT', "is_secure")

    proto = 'https' if config.default_is_secure else 'http'
    config.default_endpoint = "%s://%s:%d" % (proto, config.default_host, config.default_port)

    # vars from the main section
    config.main_access_key = cfg.get('s3 main',"access_key")
    config.main_secret_key = cfg.get('s3 main',"secret_key")
    config.main_display_name = cfg.get('s3 main',"display_name")
    config.main_user_id = cfg.get('s3 main',"user_id")
    config.main_email = cfg.get('s3 main',"email")
    try:
        config.main_kms_keyid = cfg.get('s3 main',"kms_keyid")
    except (ConfigParser.NoSectionError, ConfigParser.NoOptionError):
        config.main_kms_keyid = None
        pass

    try:
        config.main_api_name = cfg.get('s3 main',"api_name")
    except (ConfigParser.NoSectionError, ConfigParser.NoOptionError):
        config.main_api_name = ""
        pass

    config.alt_access_key = cfg.get('s3 alt',"access_key")
    config.alt_secret_key = cfg.get('s3 alt',"secret_key")
    config.alt_display_name = cfg.get('s3 alt',"display_name")
    config.alt_user_id = cfg.get('s3 alt',"user_id")
    config.alt_email = cfg.get('s3 alt',"email")

    config.tenant_access_key = cfg.get('s3 tenant',"access_key")
    config.tenant_secret_key = cfg.get('s3 tenant',"secret_key")
    config.tenant_display_name = cfg.get('s3 tenant',"display_name")
    config.tenant_user_id = cfg.get('s3 tenant',"user_id")
    config.tenant_email = cfg.get('s3 tenant',"email")

    # vars from the fixtures section
    try:
        template = cfg.get('fixtures', "bucket prefix")
    except (ConfigParser.NoOptionError):
        template = 'test-{random}-'
    prefix = choose_bucket_prefix(template=template)

    alt_client = get_alt_client()
    tenant_client = get_tenant_client()
    nuke_prefixed_buckets(prefix=prefix)
    nuke_prefixed_buckets(prefix=prefix, client=alt_client)
    nuke_prefixed_buckets(prefix=prefix, client=tenant_client)

def teardown():
    alt_client = get_alt_client()
    tenant_client = get_tenant_client()
    nuke_prefixed_buckets(prefix=prefix)
    nuke_prefixed_buckets(prefix=prefix, client=alt_client)
    nuke_prefixed_buckets(prefix=prefix, client=tenant_client)

def get_client(client_config=None):
    if client_config == None:
        client_config = Config(signature_version='s3v4')

    client = boto3.client(service_name='s3',
                        aws_access_key_id=config.main_access_key,
                        aws_secret_access_key=config.main_secret_key,
                        endpoint_url=config.default_endpoint,
                        use_ssl=config.default_is_secure,
                        config=client_config)
    return client

def get_v2_client():
    client = boto3.client(service_name='s3',
                        aws_access_key_id=config.main_access_key,
                        aws_secret_access_key=config.main_secret_key,
                        endpoint_url=config.default_endpoint,
                        use_ssl=config.default_is_secure,
                        config=Config(signature_version='s3'))
    return client

def get_alt_client(client_config=None):
    if client_config == None:
        client_config = Config(signature_version='s3v4')

    client = boto3.client(service_name='s3',
                        aws_access_key_id=config.alt_access_key,
                        aws_secret_access_key=config.alt_secret_key,
                        endpoint_url=config.default_endpoint,
                        use_ssl=config.default_is_secure,
                        config=client_config)
    return client

def get_tenant_client(client_config=None):
    if client_config == None:
        client_config = Config(signature_version='s3v4')

    client = boto3.client(service_name='s3',
                        aws_access_key_id=config.tenant_access_key,
                        aws_secret_access_key=config.tenant_secret_key,
                        endpoint_url=config.default_endpoint,
                        use_ssl=config.default_is_secure,
                        config=client_config)
    return client

def get_unauthenticated_client():
    client = boto3.client(service_name='s3',
                        aws_access_key_id='',
                        aws_secret_access_key='',
                        endpoint_url=config.default_endpoint,
                        use_ssl=config.default_is_secure,
                        config=Config(signature_version=UNSIGNED))
    return client

def get_bad_auth_client(aws_access_key_id='badauth'):
    client = boto3.client(service_name='s3',
                        aws_access_key_id=aws_access_key_id,
                        aws_secret_access_key='roflmao',
                        endpoint_url=config.default_endpoint,
                        use_ssl=config.default_is_secure,
                        config=Config(signature_version='s3v4'))
    return client

bucket_counter = itertools.count(1)

def get_new_bucket_name():
    """
    Get a bucket name that probably does not exist.

    We make every attempt to use a unique random prefix, so if a
    bucket by this name happens to exist, it's ok if tests give
    false negatives.
    """
    name = '{prefix}{num}'.format(
        prefix=prefix,
        num=next(bucket_counter),
        )
    return name

def get_new_bucket_resource(name=None):
    """
    Get a bucket that exists and is empty.

    Always recreates a bucket from scratch. This is useful to also
    reset ACLs and such.
    """
    s3 = boto3.resource('s3', 
                        aws_access_key_id=config.main_access_key,
                        aws_secret_access_key=config.main_secret_key,
                        endpoint_url=config.default_endpoint,
                        use_ssl=config.default_is_secure)
    if name is None:
        name = get_new_bucket_name()
    bucket = s3.Bucket(name)
    bucket_location = bucket.create()
    return bucket

def get_new_bucket(client=None, name=None):
    """
    Get a bucket that exists and is empty.

    Always recreates a bucket from scratch. This is useful to also
    reset ACLs and such.
    """
    if client is None:
        client = get_client()
    if name is None:
        name = get_new_bucket_name()

    client.create_bucket(Bucket=name)
    return name


def get_config_is_secure():
    return config.default_is_secure

def get_config_host():
    return config.default_host

def get_config_port():
    return config.default_port

def get_config_endpoint():
    return config.default_endpoint

def get_main_aws_access_key():
    return config.main_access_key

def get_main_aws_secret_key():
    return config.main_secret_key

def get_main_display_name():
    return config.main_display_name

def get_main_user_id():
    return config.main_user_id

def get_main_email():
    return config.main_email

def get_main_api_name():
    return config.main_api_name

def get_main_kms_keyid():
    return config.main_kms_keyid

def get_alt_aws_access_key():
    return config.alt_access_key

def get_alt_aws_secret_key():
    return config.alt_secret_key

def get_alt_display_name():
    return config.alt_display_name

def get_alt_user_id():
    return config.alt_user_id

def get_alt_email():
    return config.alt_email

def get_tenant_aws_access_key():
    return config.tenant_access_key

def get_tenant_aws_secret_key():
    return config.tenant_secret_key

def get_tenant_display_name():
    return config.tenant_display_name

def get_tenant_user_id():
    return config.tenant_user_id

def get_tenant_email():
    return config.tenant_email
