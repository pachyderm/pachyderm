#!/usr/bin/env pyspark

"""
Estimate Pi

This uses a random "dart throwing" approach, with sampling spread across a Spark cluster, then writes out result into PFS.

The number of samples to take is sourced from a config file versioned in a Pachyderm repo
"""

import random

# check if sc is already defined from pyspark
try:
    sc
except NameError:
    from pyspark import SparkContext
    sc = SparkContext(appName="Estimate_Pi")

def inside(p):
    x, y = random.random(), random.random()
    return x*x + y*y < 1

try:
    num_samples = int(open('/pfs/estimate_pi_config/num_samples').read())
except:
    print 'no config found in pfs, falling back to 100000 samples'
    num_samples = 100000

count = sc.parallelize(range(0, num_samples)).filter(inside).count()

pi = 4.0 * count / num_samples

print 'pi estimate:', pi
open('/pfs/out/pi_estimate', 'w').write(str(pi))

# vim: tabstop=8 expandtab shiftwidth=4 softtabstop=4
