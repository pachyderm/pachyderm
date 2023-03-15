"""Implementation of a classproperty since stacking of @classmethod and @property
decorators is deprecated as of Python 3.10

Reference:
    https://github.com/python/cpython/issues/89519#issuecomment-1397534245
"""
import sys
from typing import Generic, Optional, TypeVar

if sys.version_info < (3, 9):
    from typing import Callable, Type
else:
    from builtins import type as Type
    from collections.abc import Callable

T = TypeVar("T")
RT = TypeVar("RT")


# noinspection PyPep8Naming
class classproperty(Generic[T, RT]):
    """
    Class property attribute (read-only).

    Same usage as @property, but taking the class as the first argument.

        class C:
            @classproperty
            def x(cls):
                return 0

        print(C.x)    # 0
        print(C().x)  # 0
    """

    def __init__(self, func: Callable[[Type[T]], RT]) -> None:
        # For using `help(...)` on instances in Python >= 3.9.
        self.__doc__ = func.__doc__
        self.__module__ = getattr(func, "__module__", None)
        self.__name__ = func.__name__
        self.__qualname__ = func.__qualname__
        # Consistent use of __wrapped__ for wrapping functions.
        self.__wrapped__: Callable[[Type[T]], RT] = func

    def __set_name__(self, owner: Type[T], name: str) -> None:
        # Update based on class context.
        self.__module__ = owner.__module__
        self.__name__ = name
        self.__qualname__ = owner.__qualname__ + "." + name

    def __get__(self, instance: Optional[T], owner: Optional[Type[T]] = None) -> RT:
        if owner is None:
            owner = type(instance)
        return self.__wrapped__(owner)
