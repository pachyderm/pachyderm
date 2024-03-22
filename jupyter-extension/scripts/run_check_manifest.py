"""Shim for bazel to run check-manifest"""
import check_manifest
import os

assert check_manifest.check_manifest(
    source_tree=os.environ.get("BUILD_WORKING_DIRECTORY", "."),
)  # or BUILD_WORKSPACE_DIRECTORY?
