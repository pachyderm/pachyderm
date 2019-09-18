# s2

[![GoDoc](https://godoc.org/github.com/pachyderm/s2?status.svg)](https://godoc.org/github.com/pachyderm/s2)

*"It's one less than s3"*

This library facilitates the creation of servers with [S3-like APIs](https://docs.aws.amazon.com/AmazonS3/latest/API/Welcome.html) in go. With this library, you just need to implement a few interfaces to create an S3-like API - s2 handles all of the glue code and edge cases. Additionally, we provide a test runner that builds off of ceph's excellent [s3 compatibility tests](https://github.com/ceph/s3-tests) to provide fairly complete conformance tests with the s3 API.

See a complete example in the [example directory.](./example)

s2 is used in production for pachyderm's [s3gateway feature](http://docs.pachyderm.io/en/latest/enterprise/s3gateway.html).

## Adding new functionality

s2 only supports a small subset of the S3 API. When an unsupported feature is called, a `NotImplemented` error (with a 501 status code) is returned. If s2 is missing functionality that you'd like, we'd gladly take a PR that adds it!
