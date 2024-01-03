"""Module extensions interfaces repositories.bzl with bzlmod."""

load("//private:repositories.bzl", "dumb_init_deps", "etc_proto_deps")

def _non_module_deps_impl(_ctx):
    etc_proto_deps()
    dumb_init_deps()

non_module_deps = module_extension(
    implementation = _non_module_deps_impl,
)
