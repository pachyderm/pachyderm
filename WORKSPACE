workspace(name = "com_github_pachyderm_pachyderm")

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "com_google_protobuf",
    sha256 = "9bd87b8280ef720d3240514f884e56a712f2218f0d693b48050c836028940a42",
    strip_prefix = "protobuf-25.1",
    urls = ["https://github.com/protocolbuffers/protobuf/archive/v25.1.tar.gz"],
)

load("@com_google_protobuf//:protobuf_deps.bzl", "protobuf_deps")

protobuf_deps()

# rules_pyvenv contains functionality for creating a python virtual environment.
http_archive(
    name = "rules_pyvenv",
    sha256 = "3a3cc6e211850178de02b618d301f3f39d1a9cddb54d499d816ff9ea835a2834",
    strip_prefix = "rules_pyvenv-1.2",
    url = "https://github.com/cedarai/rules_pyvenv/archive/refs/tags/v1.2.tar.gz",
)

http_archive(
    name = "rules_nodejs",
    sha256 = "87c6171c5be7b69538d4695d9ded29ae2626c5ed76a9adeedce37b63c73bef67",
    strip_prefix = "rules_nodejs-6.2.0",
    url = "https://github.com/bazelbuild/rules_nodejs/releases/download/v6.2.0/rules_nodejs-v6.2.0.tar.gz",
)

load("@rules_nodejs//nodejs:repositories.bzl", "nodejs_register_toolchains")

nodejs_register_toolchains(
    name = "nodejs",
    node_version = "20.11.1",
)
