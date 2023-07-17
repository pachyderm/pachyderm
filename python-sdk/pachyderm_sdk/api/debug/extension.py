"""Handwritten classes/methods that augment the existing Debug API."""
from typing import Iterator, List, Optional, TYPE_CHECKING

from . import DebugStub
from . import DumpChunk, System

if TYPE_CHECKING:
    from ..pps import Pipeline


class ApiStub(DebugStub):
    # noinspection PyMethodOverriding
    def dump(
        self,
        *,
        system: "System" = None,
        pipelines: Optional[List["Pipeline"]] = None,
        input_repos: bool = False,
        timeout: int = 0,
    ) -> Iterator["DumpChunk"]:
        """Collect a standard set of debugging information using the DumpV2 API
          rather than the now deprecated Dump API.

        This method is intended to be used in tandem with the
          `debug.get_dump_v2_template` endpoint. However, if no system or pipelines
          are specified then this call will automatically be performed for the user.

        If no system or pipelines are specified, then debug information for all
          systems and pipelines will be returned.
        """
        if system is None and not pipelines:
            template = self.get_dump_v2_template()
            return self.dump_v2(
                system=template.request.system,
                pipelines=template.request.pipelines,
                input_repos=input_repos or template.request.input_repos,
                timeout=timeout or template.request.timeout,
            )
        return self.dump_v2(
            system=system,
            pipelines=pipelines,
            input_repos=input_repos,
            timeout=timeout,
        )
