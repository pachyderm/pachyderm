# Deploying Pachyderm On-Premises With Non-Cloud Object Stores

Coming soon.
This document, when complete, will discuss common configurations for on-premises objects stores.

## General information on non-cloud object stores

Please see [the on-premises introduction to object stores](../on_premises/#deploying-an-object-store) for some general information on object stores and how they're used with Pachyderm.

### EMC ECS
Coming soon.

### MinIO
Coming soon.

### SwiftStack
Coming soon.

## Notes
### S3 API Signature Algorithms and Regions

The S3 protocol has two different ways of authenticating requests through its api.
`S3v2` has been [officially deprecated by Amazon](https://docs.aws.amazon.com/AmazonS3/latest/dev/UsingAWSSDK.html#UsingAWSSDK-sig2-deprecation),
so it's not likely that you'll run into it.
You don't need to know the details of how they work
(though you can follow these links, S3v2 & [S3v4](https://docs.aws.amazon.com/AmazonS3/latest/API/sig-v4-authenticating-requests.html), if you're curious),
but you may run into issues with either mismatch of the signature method or availability regions.

If you have trouble getting Pachyderm to run,
check your Kubernetes logs for the `pachd` pod,
Use `kubectl get pods` to find the name of the `pachd` pod and
`kubectl logs <pachd-pod-name>` to get the logs.

If you see an error beginning with
```
INFO error starting grpc server pfs.NewBlockAPIServer
```

It could have either of two causes.

#### Availability Region Mismatch
If the error is of the form
```
INFO error starting grpc server pfs.NewBlockAPIServer: storage is unable to discern NotExist errors, "The authorization header is malformed; the region 'us-east-1' is wrong; expecting 'z1-a'" should count as NotExist
```
It may be [a known issue](https://github.com/pachyderm/pachyderm/issues/3544) with hardcoded region `us-east-1` in the minio libraries.
You can correct this by either using the `--isS3V2` flag on your the `pachctl deploy custom ...` command
or by changing the region of your storage to `us-east-1`.

#### Signature version mismatch

You're not likely to run into this in an on-premises deployment
unless your object store has deliberately been set up to use the deprecated `S3v2` signature or
you are running your on-premises deployment against Google Cloud Storage,
which is not recommended. See [Infrastructure in general](../on_premises/#infrastructure-in-general).

You'll need to determine what signature algorithm your object store uses in its S3-compatible API:  `S3v2` or `S3v4`.
If it's `S3V2`,
you can solve this by using the `--isS3V2` flag on your the `pachctl deploy custom ...` command.

