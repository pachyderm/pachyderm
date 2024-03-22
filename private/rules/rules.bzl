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

_PROTOTAR_APPLY_TMPL = """#!/usr/bin/env bash

set -euo pipefail
IFS=$'\n\t'

{BASH_RLOCATION_FUNCTION}

cd "$BUILD_WORKSPACE_DIRECTORY"
exec "$(rlocation _main/src/proto/prototar/prototar_/prototar)" --delete-jsonschemas=false apply {SRC}
"""

def _copy_to_workspace_impl(ctx):
    copier = ctx.actions.declare_file("%s_extract.sh" % (ctx.label.name))
    ctx.actions.write(
        output = copier,
        content = _PROTOTAR_APPLY_TMPL.format(
            MSG = repr(""),
            SRC = ctx.attr.src.files.to_list()[0].path,
            BASH_RLOCATION_FUNCTION = BASH_RLOCATION_FUNCTION,
        ),
        is_executable = True,
    )
    runfiles = ctx.runfiles(ctx.files.src + [copier])
    for r in ctx.attr._runfiles:
        runfiles = runfiles.merge(r.default_runfiles)
    return DefaultInfo(
        executable = copier,
        runfiles = runfiles,
    )

_copy_to_workspace = rule(
    implementation = _copy_to_workspace_impl,
    attrs = {
        "src": attr.label(
            doc = "Tar to extract into the workspace.",
        ),
        "_runfiles": attr.label_list(default = ["@bazel_tools//tools/bash/runfiles", "//src/proto/prototar"]),
    },
    toolchains = [
        "@bazel_tools//tools/sh:toolchain_type",
    ],
    executable = True,
)

_PROTOTAR_TEST_TMPL = """#!/usr/bin/env bash

set -euo pipefail
IFS=$'\n\t'

{BASH_RLOCATION_FUNCTION}

{CD_WORKSPACE}

exec "$(rlocation _main/src/proto/prototar/prototar_/prototar)" --delete-jsonschemas=false --failure-message={MESSAGE} test {SRC}
"""

def _copy_to_workspace_test_impl(ctx):
    test = ctx.actions.declare_file("%s.sh" % (ctx.label.name))

    cd = ""
    output_related_runfiles = ctx.runfiles(ctx.files.outs)
    if ctx.file.workspace:
        cd = "cd \"$(dirname \"$(realpath {})\")\"".format(repr(ctx.file.workspace.path))
        output_related_runfiles = ctx.runfiles([ctx.file.workspace])

    ctx.actions.write(
        output = test,
        content = _PROTOTAR_TEST_TMPL.format(
            MESSAGE = repr(ctx.attr.message),  # repr because starlark quotes are shell quotes, right?
            SRC = "$(rlocation {})".format(repr(ctx.workspace_name + "/" + ctx.file.src.short_path)),
            CD_WORKSPACE = cd,
            BASH_RLOCATION_FUNCTION = BASH_RLOCATION_FUNCTION,
        ),
        is_executable = True,
    )
    runfiles = ctx.runfiles([ctx.file.src, test])
    runfiles = runfiles.merge(output_related_runfiles)
    for r in ctx.attr._runfiles:
        runfiles = runfiles.merge(r.default_runfiles)
    return DefaultInfo(
        executable = test,
        runfiles = runfiles,
    )

_copy_to_workspace_test = rule(
    implementation = _copy_to_workspace_test_impl,
    attrs = {
        "src": attr.label(
            allow_single_file = True,
            doc = "Tar to extract into the workspace.",
            mandatory = True,
        ),
        "message": attr.string(
            doc = "Message to print when the test fails.",
            default = "The file in your working copy is out of date, please regenerate it.",
        ),
        "outs": attr.label_list(
            allow_files = True,
            doc = "If set, the outputs that the copy_to_workspace rule generates.  This allows the test to be cacheable.",
        ),
        "workspace": attr.label(
            allow_single_file = True,
            doc = "Label of the WORKSPACE/MODULE.bazel file, if this is a non-hermetic test.  This is ideal for the case where you don't know what files your bundle contains.",
        ),
        "_runfiles": attr.label_list(default = ["@bazel_tools//tools/bash/runfiles", "//src/proto/prototar"]),
    },
    toolchains = [
        "@bazel_tools//tools/sh:toolchain_type",
    ],
    test = True,
)

def copy_to_workspace(name = "update", src = "", message = "", outs = [], **kwargs):
    """
    Copies a generated tarball into the working copy.

    Args:
        name: The name of the rule.
        src: The target that generates a tarball.
        message: The message to print when the test fails because the working copy is out of date.
        outs: If set, the outputs that the update operation produces.  This allows the up-to-date test to be cached.
        **kwargs: Any other args to pass to the underlying rules, like "tags".
    """
    _copy_to_workspace(
        name = name,
        src = src,
        **kwargs
    )

    tags = ["style"]
    workspace = None
    if not outs:
        # If the test specifies the outputs, then it can be hermetic.
        tags = ["external", "manual", "no-cache", "no-sandbox", "style"]
        workspace = "//:MODULE.bazel"
    if "tags" in kwargs:
        tags = kwargs.pop("tags")
    _copy_to_workspace_test(
        name = name + "_test",
        src = src,
        message = message,
        tags = tags,
        workspace = workspace,
        outs = outs,
        **kwargs
    )
