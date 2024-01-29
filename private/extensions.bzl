"""Module extensions interfaces repositories.bzl with bzlmod."""

load("//private:repositories.bzl", "dumb_init_deps", "etc_proto_deps", "helm_deps", "kind_deps", "kubectl_deps", "pachdev_deps", "skopeo_deps")

def _non_module_deps_impl(_ctx):
    etc_proto_deps()
    dumb_init_deps()
    helm_deps()
    kind_deps()
    kubectl_deps()
    skopeo_deps()
    pachdev_deps()

non_module_deps = module_extension(
    implementation = _non_module_deps_impl,
)
