"""Shim for bazel to run check-manifest"""
from pathlib import Path
import os

import check_manifest

ROOT = os.environ.get("BUILD_WORKSPACE_DIRECTORY")
if ROOT:
    PROJECT = os.path.join(ROOT, "jupyter-extension")
else:
    PROJECT = str(Path(__file__).parent.parent)

print(os.environ)
print(f"checking directory: {PROJECT}")
assert check_manifest.check_manifest(source_tree=PROJECT)
