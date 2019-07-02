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

def writer(bucket, objname, fp, queue):
    key = bucket.new_key(objname)

    result = dict(
        type='w',
        bucket=bucket.name,
        key=key.name,
        )

    start = time.time()
    try:
        key.set_contents_from_file(fp, rewind=True)
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
        chunks=fp.last_chunks,
        )
    queue.put(result)


def reader(bucket, objname, queue):
    key = bucket.new_key(objname)

    fp = realistic.FileVerifier()
    result = dict(
            type='r',
            bucket=bucket.name,
            key=key.name,
            )

    start = time.time()
    try:
        key.get_contents_to_file(fp)
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
            result.update(
                error=dict(
                    msg='md5sum check failed',
                    ),
                )

    elapsed = end - start
    result.update(
        start=start,
        duration=int(round(elapsed * NANOSECOND)),
        chunks=fp.chunks,
        )
    queue.put(result)

def parse_options():
    parser = optparse.OptionParser(
        usage='%prog [OPTS] <CONFIG_YAML',
        )
    parser.add_option("--no-cleanup", dest="cleanup", action="store_false",
        help="skip cleaning up all created buckets", default=True)

    return parser.parse_args()

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
        if 'roundtrip' not in config:
            raise RuntimeError('roundtrip section not found in config')
        for item in ['readers', 'writers', 'duration', 'files', 'bucket']:
            if item not in config.roundtrip:
                raise RuntimeError("Missing roundtrip config item: {item}".format(item=item))
        for item in ['num', 'size', 'stddev']:
            if item not in config.roundtrip.files:
                raise RuntimeError("Missing roundtrip config item: files.{item}".format(item=item))

        seeds = dict(config.roundtrip.get('random_seed', {}))
        seeds.setdefault('main', random.randrange(2**32))

        rand = random.Random(seeds['main'])

        for name in ['names', 'contents', 'writer', 'reader']:
            seeds.setdefault(name, rand.randrange(2**32))

        print 'Using random seeds: {seeds}'.format(seeds=seeds)

        # setup bucket and other objects
        bucket_name = common.choose_bucket_prefix(config.roundtrip.bucket, max_len=30)
        bucket = conn.create_bucket(bucket_name)
        print "Created bucket: {name}".format(name=bucket.name)
        objnames = realistic.names(
            mean=15,
            stddev=4,
            seed=seeds['names'],
            )
        objnames = itertools.islice(objnames, config.roundtrip.files.num)
        objnames = list(objnames)
        files = realistic.files(
            mean=1024 * config.roundtrip.files.size,
            stddev=1024 * config.roundtrip.files.stddev,
            seed=seeds['contents'],
            )
        q = gevent.queue.Queue()

        logger_g = gevent.spawn(yaml.safe_dump_all, q, stream=real_stdout)

        print "Writing {num} objects with {w} workers...".format(
            num=config.roundtrip.files.num,
            w=config.roundtrip.writers,
            )
        pool = gevent.pool.Pool(size=config.roundtrip.writers)
        start = time.time()
        for objname in objnames:
            fp = next(files)
            pool.spawn(
                writer,
                bucket=bucket,
                objname=objname,
                fp=fp,
                queue=q,
                )
        pool.join()
        stop = time.time()
        elapsed = stop - start
        q.put(dict(
                type='write_done',
                duration=int(round(elapsed * NANOSECOND)),
                ))

        print "Reading {num} objects with {w} workers...".format(
            num=config.roundtrip.files.num,
            w=config.roundtrip.readers,
            )
        # avoid accessing them in the same order as the writing
        rand.shuffle(objnames)
        pool = gevent.pool.Pool(size=config.roundtrip.readers)
        start = time.time()
        for objname in objnames:
            pool.spawn(
                reader,
                bucket=bucket,
                objname=objname,
                queue=q,
                )
        pool.join()
        stop = time.time()
        elapsed = stop - start
        q.put(dict(
                type='read_done',
                duration=int(round(elapsed * NANOSECOND)),
                ))

        q.put(StopIteration)
        logger_g.get()

    finally:
        # cleanup
        if options.cleanup:
            if bucket is not None:
                common.nuke_bucket(bucket)
