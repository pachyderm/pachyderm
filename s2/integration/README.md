# s2 integration tests

Integration tests for s2 implementations, using common s3 libs/bins. Each test
targets a different lib or bin, but they all test the same basic
functionality:

1) create a bucket
2) put a couple of objects, one that can be done in a simple PUT, and one that
   should require a multipart upload
3) list objects and verify results
4) get uploaded objects and verify results
5) delete objects
6) delete bucket

These integration tests are available in addition to conformance tests because
s3 libs/bins have subtlety different behavior, but the conformance tests only
check corner cases with boto3.