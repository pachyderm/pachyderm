"""Module rules provides pachyderm-specific build rules."""

load("@aspect_bazel_lib//lib:paths.bzl", "BASH_RLOCATION_FUNCTION")
load("@bazel_skylib//rules:native_binary.bzl", "native_binary")

def host_native_binary(name, repo, target, **kwargs):
    """
    Wraps a pre-built external binary.  (See @bazel_skylib//rules:native_binary.bzl.)

    You can "bazel run" this binary, or use the binary in another dependency.  The output will have the
    same path regardless of the host architeture.
    """
    native_binary(
        name = name + "_bin",
        src = select({
            "//:is_x86_64_linux": repo + "_x86_64_linux//" + target,
            "//:is_aarch64_linux": repo + "_aarch64_linux//" + target,
            "//:is_x86_64_macos": repo + "_x86_64_macos//" + target,
            "//:is_aarch64_macos": repo + "_aarch64_macos//" + target,
        }),
        out = name,
        **kwargs
    )

_INSTALLER_TMPL = """#!/usr/bin/env bash

{BASH_RLOCATION_FUNCTION}

set -euo pipefail
IFS=$'\n\t'


if [[ $# -ge 1 ]]; then
    exec $(rlocation _main/src/cmd/install-tool/install-tool_/install-tool) -name="{NAME}" -src="{SRC}" -dst="$1"
else
    exec $(rlocation _main/src/cmd/install-tool/install-tool_/install-tool) -name="{NAME}" -src="{SRC}"
fi
"""

def _installable_binary_impl(ctx):
    installer = ctx.actions.declare_file("%s_%s.sh" % (ctx.label.name, ctx.attr.installed_name))
    ctx.actions.write(
        output = installer,
        content = _INSTALLER_TMPL.format(
            NAME = ctx.attr.installed_name,
            SRC = ctx.attr.target.files.to_list()[0].short_path,
            BASH_RLOCATION_FUNCTION = BASH_RLOCATION_FUNCTION,
        ),
        is_executable = True,
    )
    runfiles = ctx.runfiles(ctx.files.target + [installer])
    for r in ctx.attr._runfiles:
        runfiles = runfiles.merge(r.default_runfiles)
    return DefaultInfo(
        executable = installer,
        runfiles = runfiles,
    )

installable_binary = rule(
    implementation = _installable_binary_impl,
    attrs = {
        "target": attr.label(
            doc = "Binary target to install.",
        ),
        "binary": attr.string(
            doc = "Name of the binary that the target creates.",
        ),
        "installed_name": attr.string(
            doc = "Name of the binary when it's installed.",
        ),
        "_runfiles": attr.label_list(default = ["@bazel_tools//tools/bash/runfiles", "//src/cmd/install-tool"]),
    },
    toolchains = [
        "@bazel_tools//tools/sh:toolchain_type",
    ],
    executable = True,
)
