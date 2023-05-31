from functools import reduce
from typing import Iterable


def count(stream: Iterable) -> int:
    """Consume a stream and count the messages."""
    return reduce(lambda total, _: total + 1, stream, 0)
