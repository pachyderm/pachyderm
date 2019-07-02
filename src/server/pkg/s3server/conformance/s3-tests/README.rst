========================
 S3 compatibility tests
========================

This is a set of unofficial Amazon AWS S3 compatibility
tests, that can be useful to people implementing software
that exposes an S3-like API. The tests use the Boto2 and Boto3 libraries.

The tests use the Nose test framework. To get started, ensure you have
the ``virtualenv`` software installed; e.g. on Debian/Ubuntu::

	sudo apt-get install python-virtualenv

and then run::

	./bootstrap

You will need to create a configuration file with the location of the
service and two different credentials. A sample configuration file named
``s3tests.conf.SAMPLE`` has been provided in this repo. This file can be
used to run the s3 tests on a Ceph cluster started with vstart.

Once you have that file copied and edited, you can run the tests with::

	S3TEST_CONF=your.conf ./virtualenv/bin/nosetests

You can specify which directory of tests to run::

	S3TEST_CONF=your.conf ./virtualenv/bin/nosetests s3tests.functional

You can specify which file of tests to run::

	S3TEST_CONF=your.conf ./virtualenv/bin/nosetests s3tests.functional.test_s3

You can specify which test to run::

	S3TEST_CONF=your.conf ./virtualenv/bin/nosetests s3tests.functional.test_s3:test_bucket_list_empty

To gather a list of tests being run, use the flags::

	 -v --collect-only

Some tests have attributes set based on their current reliability and
things like AWS not enforcing their spec stricly. You can filter tests
based on their attributes::

	S3TEST_CONF=aws.conf ./virtualenv/bin/nosetests -a '!fails_on_aws'

Most of the tests have both Boto3 and Boto2 versions. Tests written in
Boto2 are in the ``s3tests`` directory. Tests written in Boto3 are
located in the ``s3test_boto3`` directory.

You can run only the boto3 tests with::

        S3TEST_CONF=your.conf ./virtualenv/bin/nosetests -v -s -A 'not fails_on_rgw' s3tests_boto3.functional

