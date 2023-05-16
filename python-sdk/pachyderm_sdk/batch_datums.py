import os
from functools import wraps
from typing import Callable, Dict, Optional

from . import Client

PIPELINE_FUNC = Callable[..., None]


def batch_datums(user_code: PIPELINE_FUNC) -> PIPELINE_FUNC:
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
    >>> from pachyderm_sdk import batch_datums
    >>>
    >>> @batch_datums
    >>> def pipeline():
    >>>     # process datums
    >>>     pass
    """

    @wraps(user_code)
    def wrapper(*args, **kwargs) -> None:
        client = Client()
        while True:
            with client.worker.batch_datums():
                user_code(*args, **kwargs)

    return wrapper
