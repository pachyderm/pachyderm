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


* Do we want to bubble up the manual methods to the client level?
* Auth
  * Implement full OIDC auth process (TODO: link ticket)
* Debug
  * debug_cpu (do we need it?)
* Health
  * health_check (do we need it? can we just use version?)
* PFS
  * Revisit mounting
* PPS

* Transaction
  * Using the API is very verbose. Can we fix this?
  * Add test for running a transaction in an OpenCommit. I bet it will fail.
    * Is this something that is reasonable for a user to do?
* Fix Tony issue where get_distribution fails in betterproto
  * Should be try-except.

TODO:
  * Port over examples.
  * Setup CI.
  * Investigate if we can make docs better.
    * At the very least we could generate a docstring saying
    "see <InputType> docstring for more information."
    * It would be nice to move generated Message docstrings down to methods,
    but this would require substantial betterproto work.


Bugs?
* DeleteProject<force=True> errors if project does not exist. DeleteRepo<force=True> does not have this behaviour
* ListFile give unintuitive error message if Repo.type=None
  * >>> file = pfs.File.from_uri("master@images:/")
    >>> file.commit.branch.repo.type = None
    >>> list(client.pfs.list_file(file=file))
