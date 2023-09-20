"""A high-level decorator for pipeline code that uses the datum-batching feature."""
from functools import wraps
from typing import Callable

from . import Client

PIPELINE_FUNC = Callable[..., None]


def batch_all_datums(user_code: PIPELINE_FUNC) -> PIPELINE_FUNC:
    """A decorator that will repeatedly call the wrapped function until
    all datums have been processed. Before calling the wrapped function,
    this decorator will call the NextDatum endpoint within the worker
    and set any environment variables specified by the worker.

    Any exceptions raised during the execution of the wrapped function
    will be reported back to the worker. See the pachyderm documentation
    for more information on how the datum batching feature works.

    Note: This can only be used within a Pachyderm worker.

    Examples
    --------
    >>> from pachyderm_sdk import batch_all_datums
    >>>
    >>> @batch_all_datums
    >>> def pipeline():
    >>>     # process datums
    >>>     pass
    >>>
    >>> if __name__ == '__main__':
    >>>   # Perform an expensive computation here before
    >>>   #   entering your datum processing function
    >>>   #   i.e. initializing a model.
    >>>   pipeline()

    Check the following link for a more substatial example:
        github.com/pachyderm/examples/tree/master/object-detection
    """

    @wraps(user_code)
    def wrapper(*args, **kwargs) -> None:
        worker = Client().worker
        while True:
            with worker.batch_datum():
                user_code(*args, **kwargs)

    return wrapper
