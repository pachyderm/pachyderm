# Python-SDK

This directory contains the auto-generated python SDK for interacting with 
  a Pachyderm cluster. This currently uses a fork of betterproto that can
  be found here: https://github.com/bbonenfant/python-betterproto/tree/grpcio-support


## Notes
To rebuild docker image (cd pachyderm/python-sdk)
```bash
docker build -t pachyderm_python_proto:python-sdk .
```

To regenerate proto files (cd pachyderm)
```bash
rm -rf python-sdk/pachyderm_sdk/api
find src -regex ".*\.proto" \
  | grep -v 'internal' \
  | grep -v 'server' \
  | xargs tar cf - \
  | docker run -i pachyderm_python_proto:python-sdk \
  | tar -C python-sdk/pachyderm_sdk -xf -
```


* Test if interceptors are working. At the very least error catching isn't working like it did before.
* Do we want to bubble up the manual methods to the client level?
* Admin
  * Overwrite inspect cluster to send the client version automatically.
* Auth
  * Implement full OIDC auth process (TODO: link ticket)
  * Should we implement AuthServiceNotActivated (also: Identity, License)
* Debug
  * debug_cpu (do we need it?)
* Health
  * health_check (do we need it? can we just use version?)
* PFS
  * We probably don't need ModifyFileClient (but should point in docs to new impl in api.pfs.extensions)
  * GetFile and GetFileTar will need to have docs if we don't overwrite them.
  * Revisit mounting
  * We could make the object returned by the commit context manager friendlier.
    * put_file and copy_file methods could return a pfs.File object.
* PPS
  * InspectPipeline (branch in logic)
  * CreateSecret (formats data for user)
* Transaction
  * Overall tricky because we need to update the RPC metadata being sent.
  * Options include:
    * Custom class that hides start/finish and includes contextmanager
    * Extend transaction.ApiStub methods to require passing client.



Bugs?
* DeleteProject<force=True> errors if project does not exist. DeleteRepo<force=True> does not have this behaviour
* ListFile give unintuitive error message if Repo.type=None
  * >>> file = pfs.File.from_uri("master@images:/")
    >>> file.commit.branch.repo.type = None
    >>> list(client.pfs.list_file(file=file))
