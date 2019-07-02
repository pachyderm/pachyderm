import gevent
import gevent.pool
import gevent.queue
import gevent.monkey; gevent.monkey.patch_all()
import itertools
import optparse
import os
import sys
import time
import traceback
import random
import yaml

import realistic
import common

NANOSECOND = int(1e9)

def reader(bucket, worker_id, file_names, queue, rand):
    while True:
        objname = rand.choice(file_names)
        key = bucket.new_key(objname)

        fp = realistic.FileValidator()
        result = dict(
                type='r',
                bucket=bucket.name,
                key=key.name,
                worker=worker_id,
                )

        start = time.time()
        try:
            key.get_contents_to_file(fp._file)
        except gevent.GreenletExit:
            raise
        except Exception as e:
            # stop timer ASAP, even on errors
            end = time.time()
            result.update(
                error=dict(
                    msg=str(e),
                    traceback=traceback.format_exc(),
                    ),
                )
            # certain kinds of programmer errors make this a busy
            # loop; let parent greenlet get some time too
            time.sleep(0)
        else:
            end = time.time()

            if not fp.valid():
                m='md5sum check failed start={s} ({se}) end={e} size={sz} obj={o}'.format(s=time.ctime(start), se=start, e=end, sz=fp._file.tell(), o=objname)
                result.update(
                    error=dict(
                        msg=m,
                        traceback=traceback.format_exc(),
                        ),
                    )
                print "ERROR:", m
            else:
                elapsed = end - start
                result.update(
                    start=start,
                    duration=int(round(elapsed * NANOSECOND)),
                    )
        queue.put(result)

def writer(bucket, worker_id, file_names, files, queue, rand):
    while True:
        fp = next(files)
        fp.seek(0)
        objname = rand.choice(file_names)
        key = bucket.new_key(objname)

        result = dict(
            type='w',
            bucket=bucket.name,
            key=key.name,
            worker=worker_id,
            )

        start = time.time()
        try:
            key.set_contents_from_file(fp)
        except gevent.GreenletExit:
            raise
        except Exception as e:
            # stop timer ASAP, even on errors
            end = time.time()
            result.update(
                error=dict(
                    msg=str(e),
                    traceback=traceback.format_exc(),
                    ),
                )
            # certain kinds of programmer errors make this a busy
            # loop; let parent greenlet get some time too
            time.sleep(0)
        else:
            end = time.time()

            elapsed = end - start
            result.update(
                start=start,
                duration=int(round(elapsed * NANOSECOND)),
                )

        queue.put(result)

def parse_options():
    parser = optparse.OptionParser(
        usage='%prog [OPTS] <CONFIG_YAML',
        )
    parser.add_option("--no-cleanup", dest="cleanup", action="store_false",
        help="skip cleaning up all created buckets", default=True)

    return parser.parse_args()

def write_file(bucket, file_name, fp):
    """
    Write a single file to the bucket using the file_name.
    This is used during the warmup to initialize the files.
    """
    key = bucket.new_key(file_name)
    key.set_contents_from_file(fp)

def main():
    # parse options
    (options, args) = parse_options()

    if os.isatty(sys.stdin.fileno()):
        raise RuntimeError('Need configuration in stdin.')
    config = common.read_config(sys.stdin)
    conn = common.connect(config.s3)
    bucket = None

    try:
        # setup
        real_stdout = sys.stdout
        sys.stdout = sys.stderr

        # verify all required config items are present
        if 'readwrite' not in config:
            raise RuntimeError('readwrite section not found in config')
        for item in ['readers', 'writers', 'duration', 'files', 'bucket']:
            if item not in config.readwrite:
                raise RuntimeError("Missing readwrite config item: {item}".format(item=item))
        for item in ['num', 'size', 'stddev']:
            if item not in config.readwrite.files:
                raise RuntimeError("Missing readwrite config item: files.{item}".format(item=item))

        seeds = dict(config.readwrite.get('random_seed', {}))
        seeds.setdefault('main', random.randrange(2**32))

        rand = random.Random(seeds['main'])

        for name in ['names', 'contents', 'writer', 'reader']:
            seeds.setdefault(name, rand.randrange(2**32))

        print 'Using random seeds: {seeds}'.format(seeds=seeds)

        # setup bucket and other objects
        bucket_name = common.choose_bucket_prefix(config.readwrite.bucket, max_len=30)
        bucket = conn.create_bucket(bucket_name)
        print "Created bucket: {name}".format(name=bucket.name)

        # check flag for deterministic file name creation
        if not config.readwrite.get('deterministic_file_names'):
            print 'Creating random file names'
            file_names = realistic.names(
                mean=15,
                stddev=4,
                seed=seeds['names'],
                )
            file_names = itertools.islice(file_names, config.readwrite.files.num)
            file_names = list(file_names)
        else:
            print 'Creating file names that are deterministic'
            file_names = []
            for x in xrange(config.readwrite.files.num):
                file_names.append('test_file_{num}'.format(num=x))

        files = realistic.files2(
            mean=1024 * config.readwrite.files.size,
            stddev=1024 * config.readwrite.files.stddev,
            seed=seeds['contents'],
            )
        q = gevent.queue.Queue()


        # warmup - get initial set of files uploaded if there are any writers specified
        if config.readwrite.writers > 0:
            print "Uploading initial set of {num} files".format(num=config.readwrite.files.num)
            warmup_pool = gevent.pool.Pool(size=100)
            for file_name in file_names:
                fp = next(files)
                warmup_pool.spawn(
                    write_file,
                    bucket=bucket,
                    file_name=file_name,
                    fp=fp,
                    )
            warmup_pool.join()

        # main work
        print "Starting main worker loop."
        print "Using file size: {size} +- {stddev}".format(size=config.readwrite.files.size, stddev=config.readwrite.files.stddev)
        print "Spawning {w} writers and {r} readers...".format(w=config.readwrite.writers, r=config.readwrite.readers)
        group = gevent.pool.Group()
        rand_writer = random.Random(seeds['writer'])

        # Don't create random files if deterministic_files_names is set and true
        if not config.readwrite.get('deterministic_file_names'):
            for x in xrange(config.readwrite.writers):
                this_rand = random.Random(rand_writer.randrange(2**32))
                group.spawn(
                    writer,
                    bucket=bucket,
                    worker_id=x,
                    file_names=file_names,
                    files=files,
                    queue=q,
                    rand=this_rand,
                    )

        # Since the loop generating readers already uses config.readwrite.readers
        # and the file names are already generated (randomly or deterministically),
        # this loop needs no additional qualifiers. If zero readers are specified,
        # it will behave as expected (no data is read)
        rand_reader = random.Random(seeds['reader'])
        for x in xrange(config.readwrite.readers):
            this_rand = random.Random(rand_reader.randrange(2**32))
            group.spawn(
                reader,
                bucket=bucket,
                worker_id=x,
                file_names=file_names,
                queue=q,
                rand=this_rand,
                )
        def stop():
            group.kill(block=True)
            q.put(StopIteration)
        gevent.spawn_later(config.readwrite.duration, stop)

        # wait for all the tests to finish
        group.join()
        print 'post-join, queue size {size}'.format(size=q.qsize())

        if q.qsize() > 0:
            for temp_dict in q:
                if 'error' in temp_dict:
                    raise Exception('exception:\n\t{msg}\n\t{trace}'.format(
                                    msg=temp_dict['error']['msg'],
                                    trace=temp_dict['error']['traceback'])
                                   )
                else:
                    yaml.safe_dump(temp_dict, stream=real_stdout)

    finally:
        # cleanup
        if options.cleanup:
            if bucket is not None:
                common.nuke_bucket(bucket)
