import os
from contextlib import contextmanager
from typing import ContextManager, Dict, Optional

from . import WorkerStub as _GeneratedWorkerStub


class WorkerStub(_GeneratedWorkerStub):

    __env: Dict[str, str] = dict()  # used by batch_datums
    __error: Optional[str] = None   # used by batch_datums

    @contextmanager
    def batch_datums(self) -> ContextManager:
        """A ContextManager that, when entered, calls the NextDatum
        endpoint within the worker to step forward during datum batching.
        This context manager will also prepare the environment for the user
        and report errors that occur back to the worker.

        This context manager expects to be called within an infinite while
        loop -- see the examples section. This context can only be entered
        from within a Pachyderm worker and the worker will terminate your
        code when all datums have been processed.

        Note: The API stub must have an open gRPC channel with the worker
        for NextDatum to function correctly. The ``Client`` object
        should automatically do this for the user. Accordingly, you are
        able to call this method on both instantiated and uninstantiated
        Client objects.

        Examples
        --------
        >>> from pachyderm_sdk import Client
        >>> while True:
        >>>     with Client.worker.batch_datums():
        >>>         # process datums
        >>>         pass
        """
        for key in self.__env.keys():
            del os.environ[key]
        self.__env.clear()

        response = self.next_datum(error=self.__error or "")

        for _var in response.env:
            # TODO: set env vars here and update __env dict
            pass

        self.__error = None
        try:
            yield
        except Exception as error:
            self.__error = repr(error)
            # TODO: Probably want better logging here than a print statement.
            print(f"{self.__error}\nReporting above error to worker.")
