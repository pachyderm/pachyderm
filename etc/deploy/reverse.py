#!/usr/bin/env python

# This script copies all the objects from one S3 bucket to another, while
# reversing the object keys. This is used to migrate from a pachyderm
# deployment using the object-storage option 'reverse=True' to 'reverse=False'
# (or vice-versa).
#
# Usage: reverse.py <from-bucket> <to-bucket>
#
# This script will add an object to the target bucket to track when the
# migration started, so that it can be run multiple times on a running
# deployment without having to redo work. The migration should not be
# considered complete unless it has run on a stopped pachyderm deployment,
# though, this merely allows you to do most of the work while your cluster is
# still running. If the migration errors or is interrupted, no progress from
# that run will be saved. To completely reset the progress, delete the
# 'reverse_migration' object from the target bucket.

import boto3, os, queue, subprocess, sys, threading, time, traceback
from botocore.exceptions import ClientError
from datetime import datetime, timedelta
import multiprocessing as mp

num_workers = 100
def do_worker(done, work_queue, finished_queue, from_bucket, to_bucket):
    client = boto3.client('s3')
    try:
        while True:
            try:
                path = work_queue.get_nowait()
                client.copy({'Bucket': from_bucket, 'Key': path}, to_bucket, path[::-1])
                finished_queue.put(path)
            except queue.Empty:
                if done.is_set():
                    return
                time.sleep(1)
    except KeyboardInterrupt:
        pass

migration_state_path = 'reverse_migration'
def start_migration(client, to_bucket):
    prev_start_time = None
    try:
        # Load the previous start time to know when we need to migrate objects starting from
        res = client.head_object(Bucket=to_bucket, Key=migration_state_path)
        raw = res['Metadata']['migration_start_time']
        if raw is not None and len(raw) > 0:
            prev_start_time = datetime.fromisoformat(raw)
    except ClientError as ex:
        if ex.response['Error']['Code'] != '404':
            raise

    # Write an empty object to update the LastModified so we can get the current S3 time
    if prev_start_time is None:
        res = client.put_object(Bucket=to_bucket, Key=migration_state_path, Metadata={'migration_start_time': ''})
    else:
        res = client.put_object(Bucket=to_bucket, Key=migration_state_path, Metadata={'migration_start_time': prev_start_time.isoformat()})

    # Inspect the object to get the canonical 'start time'
    res = client.head_object(Bucket=to_bucket, Key=migration_state_path)
    return prev_start_time, res['LastModified']

def end_migration(client, to_bucket, start_time):
    # Write an empty object to indicate the migration was completed for the given start time
    res = client.put_object(Bucket=to_bucket, Key=migration_state_path, Metadata={'migration_start_time': start_time.isoformat()})

def print_progress(done, scanned, enqueued, finished_queue):
    finished = 0
    while not done.is_set():
        while True:
            try:
                _ = finished_queue.get_nowait()
                finished += 1
            except queue.Empty:
                break

        buf = f'scanned: {scanned.value} enqueued: {enqueued.value} finished: {finished}...'
        sys.stdout.write(buf)
        sys.stdout.flush()
        sys.stdout.write('\b' * len(buf))
        time.sleep(0.1)

def all_objects(client, from_bucket, from_time, finished_queue):
    kwargs = {'Bucket': from_bucket}
    scanned = mp.Value('l', 0)
    enqueued = mp.Value('l', 0)
    done = mp.Event()

    progress = mp.Process(target=print_progress, args=(done, scanned, enqueued, finished_queue))
    progress.start()

    while True:
        res = client.list_objects_v2(**kwargs)
        for item in res['Contents']:
            scanned.value += 1
            if from_time is None or item['LastModified'] >= from_time:
                enqueued.value += 1
                yield (item['Key'], item['LastModified'])

        if res['IsTruncated']:
            kwargs['ContinuationToken'] = res['NextContinuationToken']
        else:
            break

    done.set()
    progress.join()
    print(f'\nSkipped objects: {scanned.value - enqueued.value}\nProcessed objects: {enqueued.value}')

def do_reverse(from_bucket, to_bucket):
    complete = False
    done = mp.Event()
    work_queue = mp.Queue(maxsize=1001) # Slightly more than one batch of results from list_object
    finished_queue = mp.Queue(maxsize=1001)

    workers = [mp.Process(target=do_worker, args=(done, work_queue, finished_queue, from_bucket, to_bucket)) for i in range(num_workers)]
    [x.start() for x in workers]

    try:
        client = boto3.client('s3')
        from_time, start_time = start_migration(client, to_bucket)
        print(f'Previous start time: {from_time}')
        print(f'Current start time: {start_time}')

        for path, modified in all_objects(client, from_bucket, from_time, finished_queue):
            work_queue.put(path)
        complete = True
    except KeyboardInterrupt:
        pass
    except Exception as ex:
        traceback.print_exc()

    try:
        done.set()
        [x.join() for x in workers]
    except KeyboardInterrupt:
        os._exit(1)

    if complete:
        print('Writing out successful migration timestamp')
        end_migration(client, to_bucket, start_time)

if __name__ == '__main__':
    if len(sys.argv) != 3:
        print(f'Usage: {sys.argv[0]} <from-bucket> <to-bucket>')
        sys.exit(1)
    do_reverse(from_bucket=sys.argv[1], to_bucket=sys.argv[2])