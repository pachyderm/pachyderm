from contextlib import contextmanager
from typing import ContextManager, Optional

from dotenv import load_dotenv

from ...constants import DOTENV_PATH_WORKER
from . import WorkerStub as _GeneratedWorkerStub


class WorkerStub(_GeneratedWorkerStub):
    __error: Optional[str] = None  # used by batch_datums

    @contextmanager
    def batch_datum(self) -> ContextManager:
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
        should automatically do this for the user.

        Examples
        --------
        >>> from pachyderm_sdk import Client
        >>>
        >>> worker = Client().worker
        >>>
        >>> # Perform an expensive computation here before
        >>> #   you being processing your datums
        >>> #   i.e. initializing a model.
        >>>
        >>> while True:
        >>>     with worker.batch_datum():
        >>>         # process datums
        >>>         pass
        """
        self.next_datum(error=self.__error or "")
        load_dotenv(DOTENV_PATH_WORKER, override=True)

        self.__error = None
        try:
            yield
        except Exception as error:
            self.__error = repr(error)
            print(f"{self.__error}\nReporting above error to worker.")
