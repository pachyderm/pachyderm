# S3Gateway

Pachyderm includes s3gateway, which allows you to interact with PFS storage
through an HTTP API that imitates a subset of AWS S3's API. This allows you to
reuse a number of tools and libraries built to work with object stores
(e.g. [minio](https://docs.minio.io/docs/minio-client-quickstart-guide.html),
[Boto](https://github.com/boto/boto3)) to interact with PFS, rather than
through its gRPC interface.
