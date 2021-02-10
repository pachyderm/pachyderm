#!/usr/bin/env python

import boto3, queue, os, re, subprocess, sys, threading, time

top_level_dirs = ['chunk', 'pach', 'artifacts']

forward_regexes = [re.compile(f'^{x}/.*$') for x in top_level_dirs]
reverse_regexes = [re.compile(f'^.*/{x[::-1]}$') for x in top_level_dirs]

num_workers = 10
def do_worker(done, work_queue, from_bucket, to_bucket):
    client = boto3.client('s3')
    while True:
        try:
            path = work_queue.get_nowait()

            print(f'Reversing {path} -> {path[::-1]}')
            client.copy({'Bucket': from_bucket, 'Key': path}, to_bucket, path[::-1])
        except queue.Empty:
            if done.is_set():
                return
            time.sleep(1)

def do_reverse(unreverse, work_queue):
    if unreverse:
        print('This will change all reversed pachyderm object paths to normal pachyderm object paths')
    else:
        print('This will change all normal pachyderm object paths to reversed pachyderm object paths')
    if input('Are you sure? (y/N) ').lower() not in ['y', 'yes']:
        print('Aborted')
        sys.exit(1)

    i = 0
    with open('objects.txt') as f:
        for line in f:
            (date, time, size, path) = line.strip().split()
            if any([x.match(path) for x in forward_regexes]):
                if not unreverse:
                    work_queue.put(path)
                    i += 1
            elif any([x.match(path) for x in reverse_regexes]):
                if unreverse:
                    work_queue.put(path)
                    i += 1
            else:
                print(f'Unknown object: {path}')

            if i == 100:
                return

unreverse_values = {'unreverse': True, 'reverse': False}

usage = f'{sys.argv[0]} (unreverse | reverse) <from-bucket> <to-bucket>'
if __name__ == '__main__':
    if len(sys.argv) != 4 or sys.argv[1] not in unreverse_values.keys():
        print(usage)
        sys.exit(1)

    unreverse = unreverse_values[sys.argv[1]]
    from_bucket = sys.argv[2]
    to_bucket = sys.argv[3]

    done = threading.Event()
    work_queue = queue.Queue(maxsize=num_workers * 2)

    workers = [threading.Thread(target=lambda: do_worker(done, work_queue, from_bucket, to_bucket)) for i in range(num_workers)]
    [x.start() for x in workers]

    try:
        do_reverse(unreverse, work_queue)
    except KeyboardInterrupt:
        pass

    try:
        done.set()
        [x.join() for x in workers]
    except KeyboardInterrupt:
        os._exit()

    sys.exit(0)

