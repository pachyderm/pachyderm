# Python-SDK

This directory contains the auto-generated python SDK for interacting with 
  a Pachyderm cluster. This currently uses a fork of betterproto that can
  be found here: https://github.com/bbonenfant/python-betterproto/tree/grpcio-support

See [example.py](./example.py) for the ported OpenCV example script.

## Notes
* Do we want to bubble up the handwritten methods to the client level?
* Auth
  * Implement full OIDC auth process (TODO: link ticket)
* Transaction
  * Using the API is very verbose. Can we fix this?
  * Add test for running a transaction in an OpenCommit. I bet it will fail.
    * Is this something that is reasonable for a user to do?

TODO:
  * Port over examples.
  * Investigate if we can make docs better.
    * At the very least we could generate a docstring saying
    "see <InputType> docstring for more information."
    * It would be nice to move generated Message docstrings down to methods,
    but this would require substantial betterproto work.
  * Investigate type-hinting context managers:
    * For Pycharm to understand the return types we wrap them in a typing.ContextManager
      object. This isn't technically correct, so we should see if VSCode works as well.

## Developer Guide

Running Tests:

```
pytest -vvv tests
```