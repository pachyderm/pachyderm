#!/usr/bin/env python

import boto3, queue, os, subprocess, sys, threading, time, traceback

processed = 0
client_mutex = threading.Lock()

def make_client():
    global client_mutex
    client_mutex.acquire()
    client = boto3.client('s3')
    client_mutex.release()
    return client

num_workers = 40
def do_worker(done, work_queue, from_bucket, to_bucket):
    global processed
    client = make_client()
    while True:
        try:
            path = work_queue.get_nowait()
            client.copy({'Bucket': from_bucket, 'Key': path}, to_bucket, path[::-1])
            processed += 1
        except queue.Empty:
            if done.is_set():
                return
            time.sleep(1)

def do_reverse(work_queue, from_bucket):
    client = make_client()
    kwargs = {'Bucket': from_bucket}
    while True:
        res = client.list_objects_v2(**kwargs)
        kwargs['ContinuationToken'] = res['NextContinuationToken']

        for item in res['Contents']:
            work_queue.put(item['Key'])

def do_progress(done):
    while not done.is_set():
        buf = f'reversed: {processed}'
        sys.stdout.write(buf)
        sys.stdout.flush()
        sys.stdout.write('\b' * len(buf))

usage = f'{sys.argv[0]} <from-bucket> <to-bucket>'
if __name__ == '__main__':
    if len(sys.argv) != 3:
        print(usage)
        sys.exit(1)

    from_bucket = sys.argv[1]
    to_bucket = sys.argv[2]

    done = threading.Event()
    work_queue = queue.Queue(maxsize=num_workers * 2)

    workers = [threading.Thread(target=lambda: do_worker(done, work_queue, from_bucket, to_bucket)) for i in range(num_workers)]
    workers.append(threading.Thread(target=lambda: do_progress(done)))
    [x.start() for x in workers]

    try:
        do_reverse(work_queue, from_bucket)
    except KeyboardInterrupt:
        pass
    except Exception as ex:
        traceback.print_exc()

    try:
        done.set()
        [x.join() for x in workers]
    except KeyboardInterrupt:
        os._exit()

    sys.exit(0)

