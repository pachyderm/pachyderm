import os
from typing import Dict, Generator

from . import WorkerStub as _GeneratedWorkerStub


class WorkerStub(_GeneratedWorkerStub):

    def batch_datums(self) -> Generator:
        """A Generator that, on iteration, calls the NextDatum endpoint
        within the worker to step forward during datum batching. This
        generator will also prepare the environment for user and report
        and errors that occur.

        Note: The API stub must have an open gRPC channel with the worker
        for NextDatum to function correctly. The ``Client`` object
        should automatically do this for the user. Accordingly, you are
        able to call this method on both instantiated and uninstantiated
        Client objects.

        Examples
        --------
        >>> from pachyderm_sdk import Client
        >>> while Client.worker.batch_datums():
        >>>     # process datums
        >>>     pass
        """
        env_vars: Dict[str, str] = dict()
        while True:
            response = self.next_datum()
            for _var in response.env:
                # TODO: set env vars here and update env_vars dict
                pass

            try:
                yield response
            except Exception as error:
                # If an error happens that is not caught by the user,
                #   alert the worker and reraise.
                self.next_datum(error=str(error))
                raise error

            for key in env_vars.keys():
                del os.environ[key]
            env_vars.clear()
