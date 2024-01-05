"""Module rules provides pachyderm-specific build rules."""

load("@bazel_skylib//rules:native_binary.bzl", "native_binary")

def host_native_binary(name, repo, target):
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
    )
