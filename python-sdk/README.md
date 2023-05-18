# Python-SDK

This directory contains the auto-generated python SDK for interacting with 
  a Pachyderm cluster. This currently uses a fork of betterproto that can
  be found here: https://github.com/bbonenfant/python-betterproto/tree/grpcio-support

See [example.py](./example.py) for the ported OpenCV example script.

## Notes
* Do we want to bubble up the handwritten methods to the client level?
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
  * Investigate type-hinting context managers:
    * For Pycharm to understand the return types we wrap them in a typing.ContextManager
      object. This isn't technically correct, so we should see if VSCode works as well. 


Bugs?
* DeleteProject<force=True> errors if project does not exist. DeleteRepo<force=True> does not have this behaviour
* ListFile give unintuitive error message if Repo.type=None
  * >>> file = pfs.File.from_uri("master@images:/")
    >>> file.commit.branch.repo.type = None
    >>> list(client.pfs.list_file(file=file))


Developer Guide

Running Tests:

```
pytest -vvv tests
```