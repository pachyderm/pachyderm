#!/usr/bin/env python

import argparse
import uuid
import subprocess
import os
import re
import time

class Failed(Exception):
    pass

def launch_pachyderm(args, env):
    try:
        manifest = subprocess.check_output("make google-cluster-manifest", env=env, shell=True)
    except subprocess.CalledProcessError:
        raise Failed()

    # Use the user-specified images
    manifest = re.sub('"pachyderm/pachd:.+"', '"{}"'.format(args.pachd_image), manifest)
    manifest = re.sub('"pachyderm/job-shim:.+"', '"{}"'.format(args.job_shim_image), manifest)
    manifest = manifest.replace('IfNotPresent', 'Always')
    tmp_manifest = '/tmp/pachyderm_benchmark_manifest'
    with open(tmp_manifest, 'w') as f:
        f.write(manifest)

    # deploy pachyderm
    if subprocess.call('kubectl create -f {}'.format(tmp_manifest), shell=True) != 0:
        raise Failed()

    # scale pachd
    if subprocess.call('kubectl scale rc pachd --replicas={}'.format(args.cluster_size), shell=True) != 0:
        raise Failed()

    # wait for all pachd nodes to be ready
    while subprocess.call("etc/kube/check_pachd_ready.sh") != 0:
        time.sleep(5)

def create_cluster(env):
    if subprocess.call("make google-cluster", env=env, shell=True) != 0:
        raise Failed()

def clean_cluster(env):
    if subprocess.call("make clean-google-cluster", env=env, shell=True) != 0:
        raise Failed()

def run_benchmark(env):
    if subprocess.call('kubectl run -t -i bench --image="{}" --restart=Never -- go test ./src/server -run=XXX -bench={}'.format(args.pachyderm_compile_image, args.benchmark), shell=True) != 0:
        raise Failed()

def gce(args):
    env = os.environ.copy()
    env['CLUSTER_NAME'] = args.cluster_name
    env['CLUSTER_SIZE'] = str(args.cluster_size)
    env['BUCKET_NAME'] = args.bucket_name
    env['STORAGE_NAME'] = args.volume_name
    env['STORAGE_SIZE'] = '10'  # defaults to 10GB

    for i in range(args.runs):
        try:
            create_cluster(env)
            launch_pachyderm(args, env)
            run_benchmark(env)
            clean_cluster(env)
        except Failed:
            print("something went wrong... removing the cluster...")
            clean_cluster(env)

def aws(args):
    print('AWS benchmark is not currently supported')

description = '''Run a Pachyderm benchmark on a cloud provider.

The typical workflow looks like this:

1. You write some code.  Maybe even add some new benchmarks.
2. You `make docker-build` to build the pachd and job-shim images.
3. You `docker tag` the images with your own namespace.  For instance, I'd tag "pachyderm/pachd:latest" with "derekchiang/pachd:latest".
4. You push the tagged images to your own namespace.
5. You run this script with --pachd-image, --job-shim-image, and --pachyderm-compile-image flags pointing to your own images.  For instance:

    ./bench.py --pachd-image derekchiang/pachd:latest --job-shim-image derekchiang/job-shim:latest --pachyderm-compile-image derekchiang/pachyderm_compile:latest
'''

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Run a Pachyderm benchmark on a cloud provider.', formatter_class=argparse.ArgumentDefaultsHelpFormatter)
    parser.add_argument('--provider', default='GCE', choices=['GCE', 'AWS'], help='the cloud provider to run the benchmark on.')
    parser.add_argument('--cluster-name', default='pachyderm-benchmark-{}'.format(str(uuid.uuid4())[:8]), help="the name of the cluster to run the benchmark on; one will be created.")
    parser.add_argument('--cluster-size', default=4, type=int, help='the number of nodes to run the benchmark on.')
    parser.add_argument('--bucket-name', default='pachyderm-benchmark-bucket-{}'.format(uuid.uuid4()), help='the GCS/S3 bucket to use with the benchmark; one will be created.')
    parser.add_argument('--volume-name', default='pachyderm-benchmark-volume-{}'.format(uuid.uuid4()), help='the persistent volume to use with the benchmark; one will be created.')
    parser.add_argument('--benchmark', default='.', help='a regex expression that specifies the benchmark to run; runs all benchmarks by default.')
    parser.add_argument('--runs', default=1, type=int, help='how many times the benchmark runs.')
    parser.add_argument('--pachd-image', default="pachyderm/pachd:latest", help='the pachd image to use.')
    parser.add_argument('--job-shim-image', default="pachyderm/job-shim:latest", help='the job-shim image to use.')
    parser.add_argument('--pachyderm-compile-image', default="pachyderm/pachyderm-compile:latest", help='the pachyderm-compile image to use.')

    args = parser.parse_args()

    print("running the benchmark with the following arguments:")
    print(args)

    if args.provider == 'GCE':
        gce(args)
    elif args.provider == 'AWS':
        aws(args)

