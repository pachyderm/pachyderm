from cStringIO import StringIO
import boto.exception
import boto.s3.connection
import boto.s3.acl
import boto.s3.lifecycle
import bunch
import datetime
import time
import email.utils
import isodate
import nose
import operator
import socket
import ssl
import os
import requests
import base64
import hmac
import sha
import pytz
import json
import httplib2
import threading
import itertools
import string
import random
import re

from collections import defaultdict
from urlparse import urlparse

from nose.tools import eq_ as eq
from nose.plugins.attrib import attr
from nose.plugins.skip import SkipTest

import utils
from .utils import assert_raises

from .policy import Policy, Statement, make_json_policy

from . import (
    nuke_prefixed_buckets,
    get_new_bucket,
    get_new_bucket_name,
    s3,
    targets,
    config,
    get_prefix,
    is_slow_backend,
    _make_request,
    _make_bucket_request,
    _make_raw_request,
    )


def check_access_denied(fn, *args, **kwargs):
    e = assert_raises(boto.exception.S3ResponseError, fn, *args, **kwargs)
    eq(e.status, 403)
    eq(e.reason, 'Forbidden')
    eq(e.error_code, 'AccessDenied')

def check_bad_bucket_name(name):
    """
    Attempt to create a bucket with a specified name, and confirm
    that the request fails because of an invalid bucket name.
    """
    e = assert_raises(boto.exception.S3ResponseError, get_new_bucket, targets.main.default, name)
    eq(e.status, 400)
    eq(e.reason.lower(), 'bad request') # some proxies vary the case
    eq(e.error_code, 'InvalidBucketName')

def _create_keys(bucket=None, keys=[]):
    """
    Populate a (specified or new) bucket with objects with
    specified names (and contents identical to their names).
    """
    if bucket is None:
        bucket = get_new_bucket()

    for s in keys:
        key = bucket.new_key(s)
        key.set_contents_from_string(s)

    return bucket


def _get_alt_connection():
    return boto.s3.connection.S3Connection(
        aws_access_key_id=s3['alt'].aws_access_key_id,
        aws_secret_access_key=s3['alt'].aws_secret_access_key,
        is_secure=s3['alt'].is_secure,
        port=s3['alt'].port,
        host=s3['alt'].host,
        calling_format=s3['alt'].calling_format,
    )


# Breaks DNS with SubdomainCallingFormat
@attr('fails_with_subdomain')
@attr(resource='bucket')
@attr(method='put')
@attr(operation='create w/! in name')
@attr(assertion='fails with subdomain')
def test_bucket_create_naming_bad_punctuation():
    # characters other than [a-zA-Z0-9._-]
    check_bad_bucket_name('alpha!soup')

def check_versioning(bucket, status):
    try:
        eq(bucket.get_versioning_status()['Versioning'], status)
    except KeyError:
        eq(status, None)

# amazon is eventual consistent, retry a bit if failed
def check_configure_versioning_retry(bucket, status, expected_string):
    bucket.configure_versioning(status)

    read_status = None

    for i in xrange(5):
        try:
            read_status = bucket.get_versioning_status()['Versioning']
        except KeyError:
            read_status = None

        if (expected_string == read_status):
            break

        time.sleep(1)

    eq(expected_string, read_status)

@attr(resource='object')
@attr(method='create')
@attr(operation='create versioned object, read not exist null version')
@attr(assertion='read null version behaves correctly')
@attr('versioning')
def test_versioning_obj_read_not_exist_null():
    bucket = get_new_bucket()
    check_versioning(bucket, None)

    check_configure_versioning_retry(bucket, True, "Enabled")

    content = 'fooz'
    objname = 'testobj'

    key = bucket.new_key(objname)
    key.set_contents_from_string(content)

    key = bucket.get_key(objname, version_id='null')
    eq(key, None)

@attr(resource='object')
@attr(method='put')
@attr(operation='append object')
@attr(assertion='success')
@attr('fails_on_aws')
@attr('appendobject')
def test_append_object():
    bucket = get_new_bucket()
    key = bucket.new_key('foo')
    expires_in = 100000
    url = key.generate_url(expires_in, method='PUT')
    o = urlparse(url)
    path = o.path + '?' + o.query
    path1 = path + '&append&position=0'
    res = _make_raw_request(host=s3.main.host, port=s3.main.port, method='PUT', path=path1, body='abc', secure=s3.main.is_secure)
    path2 = path + '&append&position=3'
    res = _make_raw_request(host=s3.main.host, port=s3.main.port, method='PUT', path=path2, body='abc', secure=s3.main.is_secure)
    eq(res.status, 200)
    eq(res.reason, 'OK')

    key = bucket.get_key('foo')
    eq(key.size, 6) 

@attr(resource='object')
@attr(method='put')
@attr(operation='append to normal object')
@attr(assertion='fails 409')
@attr('fails_on_aws')
@attr('appendobject')
def test_append_normal_object():
    bucket = get_new_bucket()
    key = bucket.new_key('foo')
    key.set_contents_from_string('abc')
    expires_in = 100000
    url = key.generate_url(expires_in, method='PUT')
    o = urlparse(url)
    path = o.path + '?' + o.query
    path = path + '&append&position=3'
    res = _make_raw_request(host=s3.main.host, port=s3.main.port, method='PUT', path=path, body='abc', secure=s3.main.is_secure)
    eq(res.status, 409)


@attr(resource='object')
@attr(method='put')
@attr(operation='append position not right')
@attr(assertion='fails 409')
@attr('fails_on_aws')
@attr('appendobject')
def test_append_object_position_wrong():
    bucket = get_new_bucket()
    key = bucket.new_key('foo')
    expires_in = 100000
    url = key.generate_url(expires_in, method='PUT')
    o = urlparse(url)
    path = o.path + '?' + o.query
    path1 = path + '&append&position=0'
    res = _make_raw_request(host=s3.main.host, port=s3.main.port, method='PUT', path=path1, body='abc', secure=s3.main.is_secure)
    path2 = path + '&append&position=9'
    res = _make_raw_request(host=s3.main.host, port=s3.main.port, method='PUT', path=path2, body='abc', secure=s3.main.is_secure)
    eq(res.status, 409)
    eq(int(res.getheader('x-rgw-next-append-position')), 3)


# TODO rgw log_bucket.set_as_logging_target() gives 403 Forbidden
# http://tracker.newdream.net/issues/984
@attr(resource='bucket.log')
@attr(method='put')
@attr(operation='set/enable/disable logging target')
@attr(assertion='operations succeed')
@attr('fails_on_rgw')
def test_logging_toggle():
    bucket = get_new_bucket()
    log_bucket = get_new_bucket(targets.main.default, bucket.name + '-log')
    log_bucket.set_as_logging_target()
    bucket.enable_logging(target_bucket=log_bucket, target_prefix=bucket.name)
    bucket.disable_logging()
    # NOTE: this does not actually test whether or not logging works

def list_bucket_storage_class(bucket):
    result = defaultdict(list)
    for k in bucket.get_all_versions():
        result[k.storage_class].append(k)

    return result

# The test harness for lifecycle is configured to treat days as 10 second intervals.
@attr(resource='bucket')
@attr(method='put')
@attr(operation='test lifecycle expiration')
@attr('lifecycle')
@attr('lifecycle_transition')
@attr('fails_on_aws')
def test_lifecycle_transition():
    sc = configured_storage_classes()
    if len(sc) < 3:
        raise SkipTest

    bucket = set_lifecycle(rules=[{'id': 'rule1', 'transition': lc_transition(days=1, storage_class=sc[1]), 'prefix': 'expire1/', 'status': 'Enabled'},
                                  {'id':'rule2', 'transition': lc_transition(days=4, storage_class=sc[2]), 'prefix': 'expire3/', 'status': 'Enabled'}])
    _create_keys(bucket=bucket, keys=['expire1/foo', 'expire1/bar', 'keep2/foo',
                                      'keep2/bar', 'expire3/foo', 'expire3/bar'])
    # Get list of all keys
    init_keys = bucket.get_all_keys()
    eq(len(init_keys), 6)

    # Wait for first expiration (plus fudge to handle the timer window)
    time.sleep(25)
    expire1_keys = list_bucket_storage_class(bucket)
    eq(len(expire1_keys['STANDARD']), 4)
    eq(len(expire1_keys[sc[1]]), 2)
    eq(len(expire1_keys[sc[2]]), 0)

    # Wait for next expiration cycle
    time.sleep(10)
    keep2_keys = list_bucket_storage_class(bucket)
    eq(len(keep2_keys['STANDARD']), 4)
    eq(len(keep2_keys[sc[1]]), 2)
    eq(len(keep2_keys[sc[2]]), 0)

    # Wait for final expiration cycle
    time.sleep(20)
    expire3_keys = list_bucket_storage_class(bucket)
    eq(len(expire3_keys['STANDARD']), 2)
    eq(len(expire3_keys[sc[1]]), 2)
    eq(len(expire3_keys[sc[2]]), 2)

# The test harness for lifecycle is configured to treat days as 10 second intervals.
@attr(resource='bucket')
@attr(method='put')
@attr(operation='test lifecycle expiration')
@attr('lifecycle')
@attr('lifecycle_transition')
@attr('fails_on_aws')
def test_lifecycle_transition_single_rule_multi_trans():
    sc = configured_storage_classes()
    if len(sc) < 3:
        raise SkipTest

    bucket = set_lifecycle(rules=[
        {'id': 'rule1',
         'transition': lc_transitions([
                lc_transition(days=1, storage_class=sc[1]),
                lc_transition(days=4, storage_class=sc[2])]),
        'prefix': 'expire1/',
        'status': 'Enabled'}])

    _create_keys(bucket=bucket, keys=['expire1/foo', 'expire1/bar', 'keep2/foo',
                                      'keep2/bar', 'expire3/foo', 'expire3/bar'])
    # Get list of all keys
    init_keys = bucket.get_all_keys()
    eq(len(init_keys), 6)

    # Wait for first expiration (plus fudge to handle the timer window)
    time.sleep(25)
    expire1_keys = list_bucket_storage_class(bucket)
    eq(len(expire1_keys['STANDARD']), 4)
    eq(len(expire1_keys[sc[1]]), 2)
    eq(len(expire1_keys[sc[2]]), 0)

    # Wait for next expiration cycle
    time.sleep(10)
    keep2_keys = list_bucket_storage_class(bucket)
    eq(len(keep2_keys['STANDARD']), 4)
    eq(len(keep2_keys[sc[1]]), 2)
    eq(len(keep2_keys[sc[2]]), 0)

    # Wait for final expiration cycle
    time.sleep(20)
    expire3_keys = list_bucket_storage_class(bucket)
    eq(len(expire3_keys['STANDARD']), 4)
    eq(len(expire3_keys[sc[1]]), 0)
    eq(len(expire3_keys[sc[2]]), 2)

def generate_lifecycle_body(rules):
    body = '<?xml version="1.0" encoding="UTF-8"?><LifecycleConfiguration>'
    for rule in rules:
        body += '<Rule><ID>%s</ID><Status>%s</Status>' % (rule['ID'], rule['Status'])
        if 'Prefix' in rule.keys():
            body += '<Prefix>%s</Prefix>' % rule['Prefix']
        if 'Filter' in rule.keys():
            prefix_str= '' # AWS supports empty filters
            if 'Prefix' in rule['Filter'].keys():
                prefix_str = '<Prefix>%s</Prefix>' % rule['Filter']['Prefix']
            body += '<Filter>%s</Filter>' % prefix_str

        if 'Expiration' in rule.keys():
            if 'ExpiredObjectDeleteMarker' in rule['Expiration'].keys():
                body += '<Expiration><ExpiredObjectDeleteMarker>%s</ExpiredObjectDeleteMarker></Expiration>' \
                        % rule['Expiration']['ExpiredObjectDeleteMarker']
            elif 'Date' in rule['Expiration'].keys():
                body += '<Expiration><Date>%s</Date></Expiration>' % rule['Expiration']['Date']
            else:
                body += '<Expiration><Days>%d</Days></Expiration>' % rule['Expiration']['Days']
        if 'NoncurrentVersionExpiration' in rule.keys():
            body += '<NoncurrentVersionExpiration><NoncurrentDays>%d</NoncurrentDays></NoncurrentVersionExpiration>' % \
                    rule['NoncurrentVersionExpiration']['NoncurrentDays']
        if 'NoncurrentVersionTransition' in rule.keys():
            for t in rule['NoncurrentVersionTransition']:
                body += '<NoncurrentVersionTransition>'
                body += '<NoncurrentDays>%d</NoncurrentDays>' % \
                    t['NoncurrentDays']
                body += '<StorageClass>%s</StorageClass>' % \
                    t['StorageClass']
                body += '</NoncurrentVersionTransition>'
        if 'AbortIncompleteMultipartUpload' in rule.keys():
            body += '<AbortIncompleteMultipartUpload><DaysAfterInitiation>%d</DaysAfterInitiation>' \
                    '</AbortIncompleteMultipartUpload>' % rule['AbortIncompleteMultipartUpload']['DaysAfterInitiation']
        body += '</Rule>'
    body += '</LifecycleConfiguration>'
    return body


@attr(resource='bucket')
@attr(method='put')
@attr(operation='set lifecycle config with noncurrent version expiration')
@attr('lifecycle')
@attr('lifecycle_transition')
def test_lifecycle_set_noncurrent_transition():
    sc = configured_storage_classes()
    if len(sc) < 3:
        raise SkipTest

    bucket = get_new_bucket()
    rules = [
        {
            'ID': 'rule1',
            'Prefix': 'test1/',
            'Status': 'Enabled',
            'NoncurrentVersionTransition': [
                {
                    'NoncurrentDays': 2,
                    'StorageClass': sc[1]
                },
                {
                    'NoncurrentDays': 4,
                    'StorageClass': sc[2]
                }
            ],
            'NoncurrentVersionExpiration': {
                'NoncurrentDays': 6
            }
        },
        {'ID': 'rule2', 'Prefix': 'test2/', 'Status': 'Disabled', 'NoncurrentVersionExpiration': {'NoncurrentDays': 3}}
    ]
    body = generate_lifecycle_body(rules)
    fp = StringIO(body)
    md5 = boto.utils.compute_md5(fp)
    headers = {'Content-MD5': md5[1], 'Content-Type': 'text/xml'}
    res = bucket.connection.make_request('PUT', bucket.name, data=fp.getvalue(), query_args='lifecycle',
                                         headers=headers)
    eq(res.status, 200)
    eq(res.reason, 'OK')


@attr(resource='bucket')
@attr(method='put')
@attr(operation='test lifecycle non-current version expiration')
@attr('lifecycle')
@attr('lifecycle_expiration')
@attr('lifecycle_transition')
@attr('fails_on_aws')
def test_lifecycle_noncur_transition():
    sc = configured_storage_classes()
    if len(sc) < 3:
        raise SkipTest

    bucket = get_new_bucket()
    check_configure_versioning_retry(bucket, True, "Enabled")

    rules = [
        {
            'ID': 'rule1',
            'Prefix': 'test1/',
            'Status': 'Enabled',
            'NoncurrentVersionTransition': [
                {
                    'NoncurrentDays': 1,
                    'StorageClass': sc[1]
                },
                {
                    'NoncurrentDays': 3,
                    'StorageClass': sc[2]
                }
            ],
            'NoncurrentVersionExpiration': {
                'NoncurrentDays': 5
            }
        }
    ]
    body = generate_lifecycle_body(rules)
    fp = StringIO(body)
    md5 = boto.utils.compute_md5(fp)
    headers = {'Content-MD5': md5[1], 'Content-Type': 'text/xml'}
    bucket.connection.make_request('PUT', bucket.name, data=fp.getvalue(), query_args='lifecycle',
                                         headers=headers)

    create_multiple_versions(bucket, "test1/a", 3)
    create_multiple_versions(bucket, "test1/b", 3)
    init_keys = bucket.get_all_versions()
    eq(len(init_keys), 6)

    time.sleep(25)
    expire1_keys = list_bucket_storage_class(bucket)
    eq(len(expire1_keys['STANDARD']), 2)
    eq(len(expire1_keys[sc[1]]), 4)
    eq(len(expire1_keys[sc[2]]), 0)

    time.sleep(20)
    expire1_keys = list_bucket_storage_class(bucket)
    eq(len(expire1_keys['STANDARD']), 2)
    eq(len(expire1_keys[sc[1]]), 0)
    eq(len(expire1_keys[sc[2]]), 4)

    time.sleep(20)
    expire_keys = bucket.get_all_versions()
    expire1_keys = list_bucket_storage_class(bucket)
    eq(len(expire1_keys['STANDARD']), 2)
    eq(len(expire1_keys[sc[1]]), 0)
    eq(len(expire1_keys[sc[2]]), 0)


def transfer_part(bucket, mp_id, mp_keyname, i, part, headers=None):
    """Transfer a part of a multipart upload. Designed to be run in parallel.
    """
    mp = boto.s3.multipart.MultiPartUpload(bucket)
    mp.key_name = mp_keyname
    mp.id = mp_id
    part_out = StringIO(part)
    mp.upload_part_from_file(part_out, i+1, headers=headers)

def generate_random(size, part_size=5*1024*1024):
    """
    Generate the specified number random data.
    (actually each MB is a repetition of the first KB)
    """
    chunk = 1024
    allowed = string.ascii_letters
    for x in range(0, size, part_size):
        strpart = ''.join([allowed[random.randint(0, len(allowed) - 1)] for _ in xrange(chunk)])
        s = ''
        left = size - x
        this_part_size = min(left, part_size)
        for y in range(this_part_size / chunk):
            s = s + strpart
        if this_part_size > len(s):
            s = s + strpart[0:this_part_size - len(s)]
        yield s
        if (x == size):
            return

def _multipart_upload(bucket, s3_key_name, size, part_size=5*1024*1024, do_list=None, headers=None, metadata=None, storage_class=None, resend_parts=[]):
    """
    generate a multi-part upload for a random file of specifed size,
    if requested, generate a list of the parts
    return the upload descriptor
    """

    if storage_class is not None:
        if not headers:
            headers = {}
        headers['X-Amz-Storage-Class'] = storage_class

    upload = bucket.initiate_multipart_upload(s3_key_name, headers=headers, metadata=metadata)
    s = ''
    for i, part in enumerate(generate_random(size, part_size)):
        s += part
        transfer_part(bucket, upload.id, upload.key_name, i, part, headers)
        if i in resend_parts:
            transfer_part(bucket, upload.id, upload.key_name, i, part, headers)

    if do_list is not None:
        l = bucket.list_multipart_uploads()
        l = list(l)

    return (upload, s)

def _populate_key(bucket, keyname, size=7*1024*1024, storage_class=None):
    if bucket is None:
        bucket = get_new_bucket()
    key = bucket.new_key(keyname)
    if storage_class:
        key.storage_class = storage_class
    data_str = str(generate_random(size, size).next())
    data = StringIO(data_str)
    key.set_contents_from_file(fp=data)
    return (key, data_str)

def gen_rand_string(size, chars=string.ascii_uppercase + string.digits):
    return ''.join(random.choice(chars) for _ in range(size))

def verify_object(bucket, k, data=None, storage_class=None):
    if storage_class:
        eq(k.storage_class, storage_class)

    if data:
        read_data = k.get_contents_as_string()

        equal = data == read_data # avoid spamming log if data not equal
        eq(equal, True)

def copy_object_storage_class(src_bucket, src_key, dest_bucket, dest_key, storage_class):
            query_args=None

            if dest_key.version_id:
                query_arg='versionId={v}'.format(v=dest_key.version_id)

            headers = {}
            headers['X-Amz-Copy-Source'] = '/{bucket}/{object}'.format(bucket=src_bucket.name, object=src_key.name)
            if src_key.version_id:
                headers['X-Amz-Copy-Source-Version-Id'] = src_key.version_id
            headers['X-Amz-Storage-Class'] = storage_class

            res = dest_bucket.connection.make_request('PUT', dest_bucket.name, dest_key.name,
                    query_args=query_args, headers=headers)
            eq(res.status, 200)

def _populate_multipart_key(bucket, kname, size, storage_class=None):
    (upload, data) = _multipart_upload(bucket, kname, size, storage_class=storage_class)
    upload.complete_upload()

    k = bucket.get_key(kname)

    return (k, data)

# Create a lifecycle config.  Either days (int) and prefix (string) is given, or rules.
# Rules is an array of dictionaries, each dict has a 'days' and a 'prefix' key
def create_lifecycle(days = None, prefix = 'test/', rules = None):
    lifecycle = boto.s3.lifecycle.Lifecycle()
    if rules == None:
        expiration = boto.s3.lifecycle.Expiration(days=days)
        rule = boto.s3.lifecycle.Rule(id=prefix, prefix=prefix, status='Enabled',
                                      expiration=expiration)
        lifecycle.append(rule)
    else:
        for rule in rules:
            expiration = None
            transition = None
            try:
                expiration = boto.s3.lifecycle.Expiration(days=rule['days'])
            except:
                pass

            try:
                transition = rule['transition']
            except:
                pass

            _id = rule.get('id',None)
            rule = boto.s3.lifecycle.Rule(id=_id, prefix=rule['prefix'],
                                          status=rule['status'], expiration=expiration, transition=transition)
            lifecycle.append(rule)
    return lifecycle

def set_lifecycle(rules = None):
    bucket = get_new_bucket()
    lifecycle = create_lifecycle(rules=rules)
    bucket.configure_lifecycle(lifecycle)
    return bucket

def configured_storage_classes():
    sc = [ 'STANDARD' ]

    if 'storage_classes' in config['main']:
        extra_sc = re.split('\W+', config['main']['storage_classes'])

        for item in extra_sc:
            if item != 'STANDARD':
                sc.append(item)

    return sc

def lc_transition(days=None, date=None, storage_class=None):
    return boto.s3.lifecycle.Transition(days=days, date=date, storage_class=storage_class)

def lc_transitions(transitions=None):
    result = boto.s3.lifecycle.Transitions()
    for t in transitions:
        result.add_transition(days=t.days, date=t.date, storage_class=t.storage_class)

    return result


@attr(resource='object')
@attr(method='put')
@attr(operation='test create object with storage class')
@attr('storage_class')
@attr('fails_on_aws')
def test_object_storage_class():
    sc = configured_storage_classes()
    if len(sc) < 2:
        raise SkipTest

    bucket = get_new_bucket()

    for storage_class in sc:
        kname = 'foo-' + storage_class
        k, data = _populate_key(bucket, kname, size=9*1024*1024, storage_class=storage_class)

        verify_object(bucket, k, data, storage_class)

@attr(resource='object')
@attr(method='put')
@attr(operation='test create multipart object with storage class')
@attr('storage_class')
@attr('fails_on_aws')
def test_object_storage_class_multipart():
    sc = configured_storage_classes()
    if len(sc) < 2:
        raise SkipTest

    bucket = get_new_bucket()
    size = 11 * 1024 * 1024

    for storage_class in sc:
        key = "mymultipart-" + storage_class
        (upload, data) = _multipart_upload(bucket, key, size, storage_class=storage_class)
        upload.complete_upload()
        key2 = bucket.get_key(key)
        eq(key2.size, size)
        eq(key2.storage_class, storage_class)

def _do_test_object_modify_storage_class(obj_write_func, size):
    sc = configured_storage_classes()
    if len(sc) < 2:
        raise SkipTest

    bucket = get_new_bucket()

    for storage_class in sc:
        kname = 'foo-' + storage_class
        k, data = obj_write_func(bucket, kname, size, storage_class=storage_class)

        verify_object(bucket, k, data, storage_class)

        for new_storage_class in sc:
            if new_storage_class == storage_class:
                continue

            copy_object_storage_class(bucket, k, bucket, k, new_storage_class)
            verify_object(bucket, k, data, storage_class)

@attr(resource='object')
@attr(method='put')
@attr(operation='test changing objects storage class')
@attr('storage_class')
@attr('fails_on_aws')
def test_object_modify_storage_class():
    _do_test_object_modify_storage_class(_populate_key, size=9*1024*1024)


@attr(resource='object')
@attr(method='put')
@attr(operation='test changing objects storage class')
@attr('storage_class')
@attr('fails_on_aws')
def test_object_modify_storage_class_multipart():
    _do_test_object_modify_storage_class(_populate_multipart_key, size=11*1024*1024)

def _do_test_object_storage_class_copy(obj_write_func, size):
    sc = configured_storage_classes()
    if len(sc) < 2:
        raise SkipTest

    src_bucket = get_new_bucket()
    dest_bucket = get_new_bucket()
    kname = 'foo'

    src_key, data = obj_write_func(src_bucket, kname, size)
    verify_object(src_bucket, src_key, data)

    for new_storage_class in sc:
        if new_storage_class == src_key.storage_class:
            continue

        dest_key = dest_bucket.get_key('foo-' + new_storage_class, validate=False)

        copy_object_storage_class(src_bucket, src_key, dest_bucket, dest_key, new_storage_class)
        verify_object(dest_bucket, dest_key, data, new_storage_class)

@attr(resource='object')
@attr(method='copy')
@attr(operation='test copy object to object with different storage class')
@attr('storage_class')
@attr('fails_on_aws')
def test_object_storage_class_copy():
    _do_test_object_storage_class_copy(_populate_key, size=9*1024*1024)

@attr(resource='object')
@attr(method='copy')
@attr(operation='test changing objects storage class')
@attr('storage_class')
@attr('fails_on_aws')
def test_object_storage_class_copy_multipart():
    _do_test_object_storage_class_copy(_populate_multipart_key, size=9*1024*1024)

class FakeFile(object):
    """
    file that simulates seek, tell, and current character
    """
    def __init__(self, char='A', interrupt=None):
        self.offset = 0
        self.char = char
        self.interrupt = interrupt

    def seek(self, offset, whence=os.SEEK_SET):
        if whence == os.SEEK_SET:
            self.offset = offset
        elif whence == os.SEEK_END:
            self.offset = self.size + offset;
        elif whence == os.SEEK_CUR:
            self.offset += offset

    def tell(self):
        return self.offset

class FakeWriteFile(FakeFile):
    """
    file that simulates interruptable reads of constant data
    """
    def __init__(self, size, char='A', interrupt=None):
        FakeFile.__init__(self, char, interrupt)
        self.size = size

    def read(self, size=-1):
        if size < 0:
            size = self.size - self.offset
        count = min(size, self.size - self.offset)
        self.offset += count

        # Sneaky! do stuff before we return (the last time)
        if self.interrupt != None and self.offset == self.size and count > 0:
            self.interrupt()

        return self.char*count

class FakeFileVerifier(object):
    """
    file that verifies expected data has been written
    """
    def __init__(self, char=None):
        self.char = char
        self.size = 0

    def write(self, data):
        size = len(data)
        if self.char == None:
            self.char = data[0]
        self.size += size
        eq(data, self.char*size)

def _verify_atomic_key_data(key, size=-1, char=None):
    """
    Make sure file is of the expected size and (simulated) content
    """
    fp_verify = FakeFileVerifier(char)
    key.get_contents_to_file(fp_verify)
    if size >= 0:
        eq(fp_verify.size, size)

def _test_atomic_dual_conditional_write(file_size):
    """
    create an object, two sessions writing different contents
    confirm that it is all one or the other
    """
    bucket = get_new_bucket()
    objname = 'testobj'
    key = bucket.new_key(objname)

    fp_a = FakeWriteFile(file_size, 'A')
    key.set_contents_from_file(fp_a)
    _verify_atomic_key_data(key, file_size, 'A')
    etag_fp_a = key.etag.replace('"', '').strip()

    # get a second key object (for the same key)
    # so both can be writing without interfering
    key2 = bucket.new_key(objname)

    # write <file_size> file of C's
    # but before we're done, try to write all B's
    fp_b = FakeWriteFile(file_size, 'B')
    fp_c = FakeWriteFile(file_size, 'C',
        lambda: key2.set_contents_from_file(fp_b, rewind=True, headers={'If-Match': etag_fp_a})
        )
    # key.set_contents_from_file(fp_c, headers={'If-Match': etag_fp_a})
    e = assert_raises(boto.exception.S3ResponseError, key.set_contents_from_file, fp_c,
                      headers={'If-Match': etag_fp_a})
    eq(e.status, 412)
    eq(e.reason, 'Precondition Failed')
    eq(e.error_code, 'PreconditionFailed')

    # verify the file
    _verify_atomic_key_data(key, file_size, 'B')

@attr(resource='object')
@attr(method='put')
@attr(operation='write one or the other')
@attr(assertion='1MB successful')
@attr('fails_on_aws')
def test_atomic_dual_conditional_write_1mb():
    _test_atomic_dual_conditional_write(1024*1024)

@attr(resource='object')
@attr(method='put')
@attr(operation='write file in deleted bucket')
@attr(assertion='fail 404')
@attr('fails_on_aws')
def test_atomic_write_bucket_gone():
    bucket = get_new_bucket()

    def remove_bucket():
        bucket.delete()

    # create file of A's but delete the bucket it's in before we finish writing
    # all of them
    key = bucket.new_key('foo')
    fp_a = FakeWriteFile(1024*1024, 'A', remove_bucket)
    e = assert_raises(boto.exception.S3ResponseError, key.set_contents_from_file, fp_a)
    eq(e.status, 404)
    eq(e.reason, 'Not Found')
    eq(e.error_code, 'NoSuchBucket')

def _multipart_upload_enc(bucket, s3_key_name, size, part_size=5*1024*1024,
                          do_list=None, init_headers=None, part_headers=None,
                          metadata=None, resend_parts=[]):
    """
    generate a multi-part upload for a random file of specifed size,
    if requested, generate a list of the parts
    return the upload descriptor
    """
    upload = bucket.initiate_multipart_upload(s3_key_name, headers=init_headers, metadata=metadata)
    s = ''
    for i, part in enumerate(generate_random(size, part_size)):
        s += part
        transfer_part(bucket, upload.id, upload.key_name, i, part, part_headers)
        if i in resend_parts:
            transfer_part(bucket, upload.id, upload.key_name, i, part, part_headers)

    if do_list is not None:
        l = bucket.list_multipart_uploads()
        l = list(l)

    return (upload, s)



@attr(resource='object')
@attr(method='put')
@attr(operation='multipart upload with bad key for uploading chunks')
@attr(assertion='successful')
@attr('encryption')
def test_encryption_sse_c_multipart_invalid_chunks_1():
    bucket = get_new_bucket()
    key = "multipart_enc"
    content_type = 'text/bla'
    objlen = 30 * 1024 * 1024
    init_headers = {
        'x-amz-server-side-encryption-customer-algorithm': 'AES256',
        'x-amz-server-side-encryption-customer-key': 'pO3upElrwuEXSoFwCfnZPdSsmt/xWeFa0N9KgDijwVs=',
        'x-amz-server-side-encryption-customer-key-md5': 'DWygnHRtgiJ77HCm+1rvHw==',
        'Content-Type': content_type
    }
    part_headers = {
        'x-amz-server-side-encryption-customer-algorithm': 'AES256',
        'x-amz-server-side-encryption-customer-key': '6b+WOZ1T3cqZMxgThRcXAQBrS5mXKdDUphvpxptl9/4=',
        'x-amz-server-side-encryption-customer-key-md5': 'arxBvwY2V4SiOne6yppVPQ=='
    }
    e = assert_raises(boto.exception.S3ResponseError,
                      _multipart_upload_enc, bucket, key, objlen,
                      init_headers=init_headers, part_headers=part_headers,
                      metadata={'foo': 'bar'})
    eq(e.status, 400)

@attr(resource='object')
@attr(method='put')
@attr(operation='multipart upload with bad md5 for chunks')
@attr(assertion='successful')
@attr('encryption')
def test_encryption_sse_c_multipart_invalid_chunks_2():
    bucket = get_new_bucket()
    key = "multipart_enc"
    content_type = 'text/plain'
    objlen = 30 * 1024 * 1024
    init_headers = {
        'x-amz-server-side-encryption-customer-algorithm': 'AES256',
        'x-amz-server-side-encryption-customer-key': 'pO3upElrwuEXSoFwCfnZPdSsmt/xWeFa0N9KgDijwVs=',
        'x-amz-server-side-encryption-customer-key-md5': 'DWygnHRtgiJ77HCm+1rvHw==',
        'Content-Type': content_type
    }
    part_headers = {
        'x-amz-server-side-encryption-customer-algorithm': 'AES256',
        'x-amz-server-side-encryption-customer-key': 'pO3upElrwuEXSoFwCfnZPdSsmt/xWeFa0N9KgDijwVs=',
        'x-amz-server-side-encryption-customer-key-md5': 'AAAAAAAAAAAAAAAAAAAAAA=='
    }
    e = assert_raises(boto.exception.S3ResponseError,
                      _multipart_upload_enc, bucket, key, objlen,
                      init_headers=init_headers, part_headers=part_headers,
                      metadata={'foo': 'bar'})
    eq(e.status, 400)

@attr(resource='bucket')
@attr(method='get')
@attr(operation='Test Bucket Policy for a user belonging to a different tenant')
@attr(assertion='succeeds')
@attr('bucket-policy')
def test_bucket_policy_different_tenant():
    bucket = get_new_bucket()
    key = bucket.new_key('asdf')
    key.set_contents_from_string('asdf')
    l = bucket.list()
    resource1 = "arn:aws:s3::*:" + bucket.name
    resource2 = "arn:aws:s3::*:" + bucket.name + "/*"
    policy_document = json.dumps(
    {
        "Version": "2012-10-17",
        "Statement": [{
        "Effect": "Allow",
        "Principal": {"AWS": "*"},
        "Action": "s3:ListBucket",
        "Resource": [
            "{}".format(resource1),
            "{}".format(resource2)
          ]
        }]
     })
    bucket.set_policy(policy_document)

    new_conn = boto.s3.connection.S3Connection(
        aws_access_key_id=s3['tenant'].aws_access_key_id,
        aws_secret_access_key=s3['tenant'].aws_secret_access_key,
        is_secure=s3['tenant'].is_secure,
        port=s3['tenant'].port,
        host=s3['tenant'].host,
        calling_format=s3['tenant'].calling_format,
        )
    bucket_name = ":" + bucket.name
    b = new_conn.get_bucket(bucket_name)
    b.get_all_keys()

@attr(resource='bucket')
@attr(method='put')
@attr(operation='Test put condition operator end with ifExists')
@attr('bucket-policy')
def test_bucket_policy_set_condition_operator_end_with_IfExists():
    bucket = _create_keys(keys=['foo'])
    policy = '''{
      "Version":"2012-10-17",
      "Statement": [{
        "Sid": "Allow Public Access to All Objects",
        "Effect": "Allow",
        "Principal": "*",
        "Action": "s3:GetObject",
        "Condition": {
                    "StringLikeIfExists": {
                        "aws:Referer": "http://www.example.com/*"
                    }
                },
        "Resource": "arn:aws:s3:::%s/*"
      }
     ]
    }''' % bucket.name
    eq(bucket.set_policy(policy), True)
    res = _make_request('GET', bucket.name, bucket.get_key("foo"),
                        request_headers={'referer': 'http://www.example.com/'})
    eq(res.status, 200)
    res = _make_request('GET', bucket.name, bucket.get_key("foo"),
                        request_headers={'referer': 'http://www.example.com/index.html'})
    eq(res.status, 200)
    res = _make_request('GET', bucket.name, bucket.get_key("foo"))
    eq(res.status, 200)
    res = _make_request('GET', bucket.name, bucket.get_key("foo"),
                        request_headers={'referer': 'http://example.com'})
    eq(res.status, 403)

def _make_arn_resource(path="*"):
    return "arn:aws:s3:::{}".format(path)

@attr(resource='object')
@attr(method='put')
@attr(operation='Deny put obj requests without encryption')
@attr(assertion='success')
@attr('encryption')
@attr('bucket-policy')
def test_bucket_policy_put_obj_enc():

    bucket = get_new_bucket()

    deny_incorrect_algo = {
        "StringNotEquals": {
          "s3:x-amz-server-side-encryption": "AES256"
        }
    }

    deny_unencrypted_obj = {
        "Null" : {
          "s3:x-amz-server-side-encryption": "true"
        }
    }

    p = Policy()
    resource = _make_arn_resource("{}/{}".format(bucket.name, "*"))

    s1 = Statement("s3:PutObject", resource, effect="Deny", condition=deny_incorrect_algo)
    s2 = Statement("s3:PutObject", resource, effect="Deny", condition=deny_unencrypted_obj)
    policy_document = p.add_statement(s1).add_statement(s2).to_json()

    bucket.set_policy(policy_document)

    key1_str ='testobj'
    key1  = bucket.new_key(key1_str)
    check_access_denied(key1.set_contents_from_string, key1_str)

    sse_client_headers = {
        'x-amz-server-side-encryption' : 'AES256',
        'x-amz-server-side-encryption-customer-algorithm': 'AES256',
        'x-amz-server-side-encryption-customer-key': 'pO3upElrwuEXSoFwCfnZPdSsmt/xWeFa0N9KgDijwVs=',
        'x-amz-server-side-encryption-customer-key-md5': 'DWygnHRtgiJ77HCm+1rvHw=='
    }


    key1.set_contents_from_string(key1_str, headers=sse_client_headers)

@attr(resource='object')
@attr(method='put')
@attr(operation='put obj with RequestObjectTag')
@attr(assertion='success')
@attr('tagging')
@attr('bucket-policy')
def test_bucket_policy_put_obj_request_obj_tag():

    bucket = get_new_bucket()

    tag_conditional = {"StringEquals": {
        "s3:RequestObjectTag/security" : "public"
    }}

    p = Policy()
    resource = _make_arn_resource("{}/{}".format(bucket.name, "*"))

    s1 = Statement("s3:PutObject", resource, effect="Allow", condition=tag_conditional)
    policy_document = p.add_statement(s1).to_json()

    bucket.set_policy(policy_document)

    new_conn = _get_alt_connection()
    bucket1 = new_conn.get_bucket(bucket.name, validate=False)
    key1_str ='testobj'
    key1  = bucket1.new_key(key1_str)
    check_access_denied(key1.set_contents_from_string, key1_str)

    headers = {"x-amz-tagging" : "security=public"}
    key1.set_contents_from_string(key1_str, headers=headers)

