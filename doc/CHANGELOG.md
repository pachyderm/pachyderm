# Changelog

## 1.0.0 (5/4/2016)

1.0.0 is the first generally available release of Pachyderm.
It's a complete rewrite of the 0.* series of releases, sharing no code with them.
The following major architectural changes have happened since 0.*:

- All network communication and serialization is done using protocol buffers and GRPC.
- BTRFS has been removed, instead build on object storage, s3 and GCS are currently supported.
- Everything in Pachyderm is now scheduled on Kubernetes, this includes Pachyderm services and user jobs.
- We now have several access methods, you can use `pachctl` from the command line, our go client within your own code and the FUSE filesystem layer
