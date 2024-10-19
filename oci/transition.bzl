"a rule transitioning an oci_image to multiple platforms"

# This is from: https://github.com/bazel-contrib/rules_oci/blob/main/examples/multi_architecture_image/transition.bzl

def _multiarch_transition(_, attr):
    return [
        {"//command_line_option:platforms": str(platform)}
        for platform in attr.platforms
    ]

multiarch_transition = transition(
    implementation = _multiarch_transition,
    inputs = [],
    outputs = ["//command_line_option:platforms"],
)

def _impl(ctx):
    return DefaultInfo(files = depset(ctx.files.image))

multi_arch = rule(
    implementation = _impl,
    attrs = {
        "image": attr.label(cfg = multiarch_transition),
        "platforms": attr.label_list(),
        "_allowlist_function_transition": attr.label(
            default = "@bazel_tools//tools/allowlists/function_transition_allowlist",
        ),
    },
)
