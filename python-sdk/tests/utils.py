from typing import Iterable


def count(stream: Iterable) -> int:
    """Count the messages in a stream."""
    return len(list(stream))
