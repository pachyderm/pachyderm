"""Module jsonnet-lint provides a jsonnet_test rule for testing that jsonnet files are syntactically valid."""

_LINTER_TMPL = """#!/usr/bin/env bash

set -euo pipefail
IFS=$'\n\t'

exec {JSONNET_LINT} {SRC}
"""

def _jsonnet_lint_test_impl(ctx):
    linter = ctx.actions.declare_file("%s_lint.sh" % (ctx.label.name))
    ctx.actions.write(
        output = linter,
        content = _LINTER_TMPL.format(
            JSONNET_LINT = ctx.attr._runfiles[0][DefaultInfo].files_to_run.executable.short_path,
            SRC = ctx.attr.src.files.to_list()[0].short_path,
        ),
        is_executable = True,
    )
    runfiles = ctx.runfiles(ctx.files.deps + ctx.files.src + [linter])
    for r in ctx.attr._runfiles:
        runfiles = runfiles.merge(r.default_runfiles)
    return DefaultInfo(
        executable = linter,
        runfiles = runfiles,
    )

jsonnet_lint_test = rule(
    implementation = _jsonnet_lint_test_impl,
    attrs = {
        "src": attr.label(
            allow_single_file = True,
            doc = "A jsonnet program to check for validity.",
            mandatory = True,
        ),
        "deps": attr.label_list(
            allow_files = True,
            doc = "Any jsonnet libraries needed by the jsonnet program to be validated.",
        ),
        "_runfiles": attr.label_list(default = ["@jsonnet_go//cmd/jsonnet-lint:jsonnet-lint"]),
    },
    toolchains = [
        "@bazel_tools//tools/sh:toolchain_type",
    ],
    executable = True,
    test = True,
)
