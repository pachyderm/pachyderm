"""Module rules provides pachyderm-specific build rules."""

load("@bazel_skylib//rules:native_binary.bzl", "native_binary")

def host_native_binary(name, repo, target):
    """
    Wraps a pre-built external binary.  (See @bazel_skylib//rules:native_binary.bzl.)

    You can "bazel run" this binary, or use the binary in another dependency.  The output will have the
    same path regardless of the host architeture.
    """
    for arch in ["_x86_64_linux", "_aarch64_linux", "_x86_64_macos", "_aarch64_macos"]:
        native_binary(
            name = name + arch,
            src = repo + arch + "//" + target,
            out = name,
        )

    native.alias(
        name = name,
        actual = select({
            "//:is_x86_64_linux": name + "_x86_64_linux",
            "//:is_aarch64_linux": name + "_aarch64_linux",
            "//:is_x86_64_macos": name + "_x86_64_macos",
            "//:is_aarch64_macos": name + "_aarch64_macos",
        }),
    )
