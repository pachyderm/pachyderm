#!/usr/bin/env python3
"""Script to generate the documentation for the pachyderm_sdk package."""
from pathlib import Path
from shutil import rmtree
from sys import stderr

import betterproto
import pdoc

betterproto.PLACEHOLDER = None  # This cleans up docs for generated Message classes.

MODULE = "pachyderm_sdk"
OUTPUT_DIR = Path(__file__).parent
TEMPLATE_DIR = str(OUTPUT_DIR / "templates")
CONFIG = {"sort_identifiers": False}


def main():
    # Write to stderr since the containerized workflow outputs to stdout.
    print(f"Generating HTML documentation for module: {MODULE}", file=stderr)
    rmtree(OUTPUT_DIR / MODULE, ignore_errors=True)
    pdoc.tpl_lookup.directories.insert(0, TEMPLATE_DIR)
    module = pdoc.Module(MODULE, docfilter=None, skip_errors=False)
    pdoc.link_inheritance()

    def recursive_write_files(m: pdoc.Module):
        html = m.html(**CONFIG)

        filepath = OUTPUT_DIR.joinpath(m.url())
        print(f"Writing {filepath.relative_to(OUTPUT_DIR.parent)}", file=stderr)
        filepath.parent.mkdir(parents=True, exist_ok=True)
        filepath.write_text(html)

        for submodule in m.submodules():
            recursive_write_files(submodule)

    recursive_write_files(module)


if __name__ == "__main__":
    main()
