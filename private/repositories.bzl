"""Module repositories imports non-bzlmod aware dependencies."""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive", "http_file")

def etc_proto_deps():
    """
    etc_proto_deps adds dependencies necessary to run "make proto", i.e. //etc/proto:run.
    """

    # protoc-gen-go
    gengo_build_file_content = """exports_files(["protoc-gen-go"])"""
    http_archive(
        name = "com_github_protocolbuffers_protobuf_go_protoc_gen_go_x86_64_linux",
        url = "https://github.com/protocolbuffers/protobuf-go/releases/download/v1.31.0/protoc-gen-go.v1.31.0.linux.amd64.tar.gz",
        sha256 = "04414c31a3af6f908d4359ff12a02f7ef864417978e303ccc62053af536ae13c",
        build_file_content = gengo_build_file_content,
    )
    http_archive(
        name = "com_github_protocolbuffers_protobuf_go_protoc_gen_go_aarch64_linux",
        url = "https://github.com/protocolbuffers/protobuf-go/releases/download/v1.31.0/protoc-gen-go.v1.31.0.linux.arm64.tar.gz",
        sha256 = "7963f59b873680b90e61c3be94bfe35e4731789516fc7377ead1be652212fcb1",
        build_file_content = gengo_build_file_content,
    )
    http_archive(
        name = "com_github_protocolbuffers_protobuf_go_protoc_gen_go_x86_64_macos",
        url = "https://github.com/protocolbuffers/protobuf-go/releases/download/v1.31.0/protoc-gen-go.v1.31.0.darwin.amd64.tar.gz",
        sha256 = "7890e2790dd68b181b1f6c33f306073e0abda3e7f360548e0b5ccb5fc20485a5",
        build_file_content = gengo_build_file_content,
    )
    http_archive(
        name = "com_github_protocolbuffers_protobuf_go_protoc_gen_go_aarch64_macos",
        url = "https://github.com/protocolbuffers/protobuf-go/releases/download/v1.31.0/protoc-gen-go.v1.31.0.darwin.arm64.tar.gz",
        sha256 = "c01ab747f9decfb9bc300c8506a8e741d35dc45860cbf3950c752572129b2139",
        build_file_content = gengo_build_file_content,
    )

    # protoc-gen-validate
    genvalidate_build_file_content = """exports_files(["protoc-gen-validate-go"])"""
    http_archive(
        name = "com_github_bufbuild_protoc_gen_validate_x86_64_linux",
        url = "https://github.com/bufbuild/protoc-gen-validate/releases/download/v1.0.2/protoc-gen-validate_1.0.2_linux_amd64.tar.gz",
        build_file_content = genvalidate_build_file_content,
        integrity = "sha256-XVMR6B95Kben+O1jY8hfg6Jbdw7gDzvUUCSYuR5tO3E=",
    )
    http_archive(
        name = "com_github_bufbuild_protoc_gen_validate_aarch64_linux",
        url = "https://github.com/bufbuild/protoc-gen-validate/releases/download/v1.0.2/protoc-gen-validate_1.0.2_linux_arm64.tar.gz",
        build_file_content = genvalidate_build_file_content,
        integrity = "sha256-kPUsHhhnC+8DdiffEJuyDiTWti7FTIMVBPB0PQLi+sI=",
    )
    http_archive(
        name = "com_github_bufbuild_protoc_gen_validate_x86_64_macos",
        url = "https://github.com/bufbuild/protoc-gen-validate/releases/download/v1.0.2/protoc-gen-validate_1.0.2_darwin_amd64.tar.gz",
        build_file_content = genvalidate_build_file_content,
        integrity = "sha256-CPPaRgwsEOlkgW0acOVctL4aLQe1wRpNbQEMXzr3QtQ=",
    )
    http_archive(
        name = "com_github_bufbuild_protoc_gen_validate_aarch64_macos",
        url = "https://github.com/bufbuild/protoc-gen-validate/releases/download/v1.0.2/protoc-gen-validate_1.0.2_darwin_arm64.tar.gz",
        build_file_content = genvalidate_build_file_content,
        integrity = "sha256-CtoZYRgLAU26WIkDDV+1Row6H4l0PcbN9TRG56vbVDU=",
    )

    # protoc-gen-doc
    gendoc_build_file_content = """exports_files(["protoc-gen-doc"])"""
    http_archive(
        name = "com_github_pseudomuto_protoc_gen_doc_x86_64_linux",
        url = "https://github.com/pseudomuto/protoc-gen-doc/releases/download/v1.5.1/protoc-gen-doc_1.5.1_linux_amd64.tar.gz",
        build_file_content = gendoc_build_file_content,
        integrity = "sha256-R81ysH5tqzQI1oamXTfTpqthbafYtWSyvSopY6crcv0=",
    )
    http_archive(
        name = "com_github_pseudomuto_protoc_gen_doc_aarch64_linux",
        url = "https://github.com/pseudomuto/protoc-gen-doc/releases/download/v1.5.1/protoc-gen-doc_1.5.1_linux_arm64.tar.gz",
        build_file_content = gendoc_build_file_content,
        integrity = "sha256-Fy5qGR2s7Y6zHrzZDUUjoa/6bQeQCom1SEIYI92nlv4=",
    )
    http_archive(
        name = "com_github_pseudomuto_protoc_gen_doc_x86_64_macos",
        url = "https://github.com/pseudomuto/protoc-gen-doc/releases/download/v1.5.1/protoc-gen-doc_1.5.1_darwin_amd64.tar.gz",
        build_file_content = gendoc_build_file_content,
        integrity = "sha256-9Cnlpd3Yhr+2gmXy+SwcalCXgLetyveos76UPyjhRLo=",
    )
    http_archive(
        name = "com_github_pseudomuto_protoc_gen_doc_aarch64_macos",
        url = "https://github.com/pseudomuto/protoc-gen-doc/releases/download/v1.5.1/protoc-gen-doc_1.5.1_darwin_arm64.tar.gz",
        build_file_content = gendoc_build_file_content,
        integrity = "sha256-boxzfZpnpqhzo/HTfti7KgqZlvbc9nAaoQSMe9eYqvk=",
    )

    # protoc-gen-openapiv2
    http_file(
        name = "com_github_grpc_ecosystem_grpc_gateway_protoc_gen_openapiv2_x86_64_linux",
        url = "https://github.com/grpc-ecosystem/grpc-gateway/releases/download/v2.18.1/protoc-gen-openapiv2-v2.18.1-linux-x86_64",
        integrity = "sha256-ABA8GJOn6widP5Ybn2QmN5l8rLizezUhomHlCPw+sE4=",
        executable = True,
        downloaded_file_path = "protoc-gen-openapiv2",
    )
    http_file(
        name = "com_github_grpc_ecosystem_grpc_gateway_protoc_gen_openapiv2_aarch64_linux",
        url = "https://github.com/grpc-ecosystem/grpc-gateway/releases/download/v2.18.1/protoc-gen-openapiv2-v2.18.1-linux-arm64",
        integrity = "sha256-0+TNVOGyBSxx6+Uikbt4sI6SJ7xTYfKdyzy5SRL23v0=",
        executable = True,
        downloaded_file_path = "protoc-gen-openapiv2",
    )
    http_file(
        name = "com_github_grpc_ecosystem_grpc_gateway_protoc_gen_openapiv2_x86_64_macos",
        url = "https://github.com/grpc-ecosystem/grpc-gateway/releases/download/v2.18.1/protoc-gen-openapiv2-v2.18.1-darwin-x86_64",
        integrity = "sha256-eHoFTKQC5E3iYmvhoglL/0QKA6m0UrnKxFn12+22jEw=",
        executable = True,
        downloaded_file_path = "protoc-gen-openapiv2",
    )
    http_file(
        name = "com_github_grpc_ecosystem_grpc_gateway_protoc_gen_openapiv2_aarch64_macos",
        url = "https://github.com/grpc-ecosystem/grpc-gateway/releases/download/v2.18.1/protoc-gen-openapiv2-v2.18.1-darwin-arm64",
        integrity = "sha256-QdkMfGznadc2KMO0ZZ8q9gQLcAp87nkdG2dRyrQ+ClE=",
        executable = True,
        downloaded_file_path = "protoc-gen-openapiv2",
    )

    # protoc-gen-grpc-gateway
    http_file(
        name = "com_github_grpc_ecosystem_grpc_gateway_protoc_gen_grpc_gateway_x86_64_linux",
        url = "https://github.com/grpc-ecosystem/grpc-gateway/releases/download/v2.18.1/protoc-gen-grpc-gateway-v2.18.1-linux-x86_64",
        integrity = "sha256-PCEXHYgPROUCI6NBXYHMPjWaxANgpYqrw6bXMl1168k=",
        executable = True,
        downloaded_file_path = "protoc-gen-grpc-gateway",
    )
    http_file(
        name = "com_github_grpc_ecosystem_grpc_gateway_protoc_gen_grpc_gateway_aarch64_linux",
        url = "https://github.com/grpc-ecosystem/grpc-gateway/releases/download/v2.18.1/protoc-gen-grpc-gateway-v2.18.1-linux-arm64",
        integrity = "sha256-d0Q4iffQ4eywC/lAeV+dxF2UfBLN4DwZ2MW/OSYkDuc=",
        executable = True,
        downloaded_file_path = "protoc-gen-grpc-gateway",
    )
    http_file(
        name = "com_github_grpc_ecosystem_grpc_gateway_protoc_gen_grpc_gateway_x86_64_macos",
        url = "https://github.com/grpc-ecosystem/grpc-gateway/releases/download/v2.18.1/protoc-gen-grpc-gateway-v2.18.1-darwin-x86_64",
        integrity = "sha256-yeFzNDsIHqcj7A+k6FfUcJL3tWag2+0vGeKzYabh9/k=",
        executable = True,
        downloaded_file_path = "protoc-gen-grpc-gateway",
    )
    http_file(
        name = "com_github_grpc_ecosystem_grpc_gateway_protoc_gen_grpc_gateway_aarch64_macos",
        url = "https://github.com/grpc-ecosystem/grpc-gateway/releases/download/v2.18.1/protoc-gen-grpc-gateway-v2.18.1-darwin-arm64",
        integrity = "sha256-ZKcjzAiohWlQXFa2pAvIk9+xKGnbQOpx3E0anmScSq0=",
        executable = True,
        downloaded_file_path = "protoc-gen-grpc-gateway",
    )

    # protoc-gen-grpc-gateway-ts
    gwts_build_file_content = """exports_files(["protoc-gen-grpc-gateway-ts"])"""
    http_archive(
        name = "com_github_grpc_ecosystem_grpc_gateway_protoc_gen_grpc_gateway_ts_x86_64_linux",
        url = "https://github.com/grpc-ecosystem/protoc-gen-grpc-gateway-ts/releases/download/v1.1.2/protoc-gen-grpc-gateway-ts_1.1.2_Linux_amd64.tar.gz",
        build_file_content = gwts_build_file_content,
        sha256 = "94c348e2554d8c76d10bb48439e5e015c0da54aae46a87adca44527c737ea341",
    )
    http_archive(
        name = "com_github_grpc_ecosystem_grpc_gateway_protoc_gen_grpc_gateway_ts_aarch64_linux",
        url = "https://github.com/grpc-ecosystem/protoc-gen-grpc-gateway-ts/releases/download/v1.1.2/protoc-gen-grpc-gateway-ts_1.1.2_Linux_arm64.tar.gz",
        build_file_content = gwts_build_file_content,
        sha256 = "f4a0f1f2e6cf3c92b14e5f92af3a3774e752f9ade76831a3f02568f037746042",
    )
    http_archive(
        name = "com_github_grpc_ecosystem_grpc_gateway_protoc_gen_grpc_gateway_ts_x86_64_macos",
        url = "https://github.com/grpc-ecosystem/protoc-gen-grpc-gateway-ts/releases/download/v1.1.2/protoc-gen-grpc-gateway-ts_1.1.2_Darwin_amd64.tar.gz",
        build_file_content = gwts_build_file_content,
        sha256 = "32ccc4ac6ec42b183a4ae806dac628daa47e1eb9dd35356f72f07ffc12967357",
    )
    http_archive(
        name = "com_github_grpc_ecosystem_grpc_gateway_protoc_gen_grpc_gateway_ts_aarch64_macos",
        url = "https://github.com/grpc-ecosystem/protoc-gen-grpc-gateway-ts/releases/download/v1.1.2/protoc-gen-grpc-gateway-ts_1.1.2_Darwin_arm64.tar.gz",
        build_file_content = gwts_build_file_content,
        sha256 = "de4aa0ae3734ac62737a982b39a7194f7ca4925c49f282e7570f0ab6597ba532",
    )

    # gopatch
    gopatch_build_file_content = """exports_files(["gopatch"])"""
    http_archive(
        name = "org_uber_go_gopatch_x86_64_linux",
        url = "https://github.com/uber-go/gopatch/releases/download/v0.3.0/gopatch_Linux_x86_64.tar.gz",
        build_file_content = gopatch_build_file_content,
        sha256 = "34f12f161209ce91010236a3781a7c90b3a391377f69be7376cce02586a71140",
    )
    http_archive(
        name = "org_uber_go_gopatch_aarch64_linux",
        url = "https://github.com/uber-go/gopatch/releases/download/v0.3.0/gopatch_Linux_arm64.tar.gz",
        build_file_content = gopatch_build_file_content,
        sha256 = "2f73999527945dda74323dddf991442a52fececd4ff7f810821ef46dcf33e3ad",
    )
    http_archive(
        name = "org_uber_go_gopatch_x86_64_macos",
        url = "https://github.com/uber-go/gopatch/releases/download/v0.3.0/gopatch_Darwin_x86_64.tar.gz",
        build_file_content = gopatch_build_file_content,
        sha256 = "b34750977ce6802c2ade0c7c302c79d065b86eb9e09bca86fa6f245b9d3fabe3",
    )
    http_archive(
        name = "org_uber_go_gopatch_aarch64_macos",
        url = "https://github.com/uber-go/gopatch/releases/download/v0.3.0/gopatch_Darwin_arm64.tar.gz",
        build_file_content = gopatch_build_file_content,
        sha256 = "2bb3914dbf273581c34e0db0913bd99f10cbbe622849ee0d8a460dbd03f27348",
    )

    # protoc binaries
    protoc_build_file_content = """exports_files(["bin/protoc"])"""
    http_archive(
        name = "com_github_protocolbuffers_protobuf_x86_64_linux",
        url = "https://github.com/protocolbuffers/protobuf/releases/download/v25.1/protoc-25.1-linux-x86_64.zip",
        build_file_content = protoc_build_file_content,
        integrity = "sha256-7Y/Kh6EciI/tMp1qWcNMfUNhZfZiosh1JG3bGsK23VA=",
    )
    http_archive(
        name = "com_github_protocolbuffers_protobuf_aarch64_linux",
        url = "https://github.com/protocolbuffers/protobuf/releases/download/v25.1/protoc-25.1-linux-aarch_64.zip",
        build_file_content = protoc_build_file_content,
        integrity = "sha256-mZdajBG4PNZcPhFRrhcUv5WavAUhrLZZv3IFJCdqsMg=",
    )
    http_archive(
        name = "com_github_protocolbuffers_protobuf_x86_64_macos",
        url = "https://github.com/protocolbuffers/protobuf/releases/download/v25.1/protoc-25.1-osx-x86_64.zip",
        build_file_content = protoc_build_file_content,
        integrity = "sha256-csbWsryFX/hojDt/sxKIzK/Qq1Ulb/g4LVcR7PzBH08=",
    )
    http_archive(
        name = "com_github_protocolbuffers_protobuf_aarch64_macos",
        url = "https://github.com/protocolbuffers/protobuf/releases/download/v25.1/protoc-25.1-osx-aarch_64.zip",
        build_file_content = protoc_build_file_content,
        integrity = "sha256-MgMIzhjDWVZJSHVPUXSN5BzwKk5+3wz0eoBbnThhDxY=",
    )

def dumb_init_deps():
    http_file(
        name = "com_github_yelp_dumb_init_x86_64_linux",
        url = "https://github.com/Yelp/dumb-init/releases/download/v1.2.5/dumb-init_1.2.5_x86_64",
        integrity = "sha256-6HS1XzJ5ykFBXSkMUSp7qdCPmAQbKK58KssZpUXxxN8=",
        executable = True,
        downloaded_file_path = "dumb-init",
    )
    http_file(
        name = "com_github_yelp_dumb_init_aarch64_linux",
        url = "https://github.com/Yelp/dumb-init/releases/download/v1.2.5/dumb-init_1.2.5_aarch64",
        integrity = "sha256-t9ZI+XFUqZxTm2PFWXnNKfAF+IQw+zgwB/40WDQLeV4=",
        executable = True,
        downloaded_file_path = "dumb-init",
    )

def helm_deps():
    """Binaries for helm."""
    build = """exports_files(["helm"])"""
    http_archive(
        name = "sh_helm_x86_64_linux",
        url = "https://get.helm.sh/helm-v3.14.2-linux-amd64.tar.gz",
        build_file_content = build,
        sha256 = "0885a501d586c1e949e9b113bf3fb3290b0bbf74db9444a1d8c2723a143006a5",
        strip_prefix = "linux-amd64",
    )
    http_archive(
        name = "sh_helm_aarch64_linux",
        url = "https://get.helm.sh/helm-v3.14.2-linux-arm64.tar.gz",
        build_file_content = build,
        sha256 = "c65d6a9557bb359abc2c0d26670de850b52327dc3976ad6f9e14c298ea3e1b61",
        strip_prefix = "linux-arm64",
    )
    http_archive(
        name = "sh_helm_x86_64_macos",
        url = "https://get.helm.sh/helm-v3.14.2-darwin-amd64.tar.gz",
        build_file_content = build,
        sha256 = "64c633ae194bde77b7e7b7936a2814a7417817dc8b7bb7d270bd24a7a17b8d12",
        strip_prefix = "darwin-amd64",
    )
    http_archive(
        name = "sh_helm_aarch64_macos",
        url = "https://get.helm.sh/helm-v3.14.2-darwin-arm64.tar.gz",
        build_file_content = build,
        sha256 = "ff502fd39b06497fa3d5a51ec2ced02b9fcfdb0e9a948d315fb1b2f13ddc39fb",
        strip_prefix = "darwin-arm64",
    )

def kind_deps():
    http_file(
        name = "io_k8s_sigs_kind_x86_64_linux",
        url = "https://github.com/kubernetes-sigs/kind/releases/download/v0.22.0/kind-linux-amd64",
        sha256 = "e4264d7ee07ca642fe52818d7c0ed188b193c214889dd055c929dbcb968d1f62",
        executable = True,
        downloaded_file_path = "kind",
    )
    http_file(
        name = "io_k8s_sigs_kind_aarch64_linux",
        url = "https://github.com/kubernetes-sigs/kind/releases/download/v0.22.0/kind-linux-arm64",
        sha256 = "4431805115da3b54290e3e976fe2db9a7e703f116177aace6735dfa1d8a4f3fe",
        executable = True,
        downloaded_file_path = "kind",
    )
    http_file(
        name = "io_k8s_sigs_kind_x86_64_macos",
        url = "https://github.com/kubernetes-sigs/kind/releases/download/v0.22.0/kind-darwin-amd64",
        sha256 = "28a9f7ad7fd77922c639e21c034d0f989b33402693f4f842099cd9185b144d20",
        executable = True,
        downloaded_file_path = "kind",
    )
    http_file(
        name = "io_k8s_sigs_kind_aarch64_macos",
        url = "https://github.com/kubernetes-sigs/kind/releases/download/v0.22.0/kind-darwin-arm64",
        sha256 = "c8dd3b287965150ae4db668294edc48229116e95d94620c306d8fae932ee633f",
        executable = True,
        downloaded_file_path = "kind",
    )

def kubectl_deps():
    http_file(
        name = "io_k8s_dl_kubectl_x86_64_linux",
        url = "https://dl.k8s.io/release/v1.29.2/bin/linux/amd64/kubectl",
        sha256 = "7816d067740f47f949be826ac76943167b7b3a38c4f0c18b902fffa8779a5afa",
        executable = True,
        downloaded_file_path = "kubectl",
    )
    http_file(
        name = "io_k8s_dl_kubectl_aarch64_linux",
        url = "https://dl.k8s.io/release/v1.29.2/bin/linux/arm64/kubectl",
        sha256 = "3507ecb4224cf05ae2151a98d4932253624e7762159936d5347b19fe037655ca",
        executable = True,
        downloaded_file_path = "kubectl",
    )
    http_file(
        name = "io_k8s_dl_kubectl_x86_64_macos",
        url = "https://dl.k8s.io/release/v1.29.2/bin/darwin/amd64/kubectl",
        sha256 = "bb04d9450d9c9fa120956c5cc7c8dfaa700297038ff9c941741e730b02bbd1f3",
        executable = True,
        downloaded_file_path = "kubectl",
    )
    http_file(
        name = "io_k8s_dl_kubectl_aarch64_macos",
        url = "https://dl.k8s.io/release/v1.29.2/bin/darwin/arm64/kubectl",
        sha256 = "ce030f86625df96560402573d86d4e6f4b8b956ca3e3b9df57cb8ccf2b9a540c",
        executable = True,
        downloaded_file_path = "kubectl",
    )

def skopeo_deps():
    # Skopeo is an enormous can of worms.  Like a neutron star made only of worms.  The original
    # authors apparently do not believe in binary releases, apparently because one time Redhat broke
    # libc or something?  That's fine, whatever, it's just Go, so we'll build it ourselves like we
    # do for all the crazy proto compilers!  But as it turns out, it depends on a bunch of C
    # libraries that I do not want to figure out how to Bazel-ify.  At some point it needs headers
    # from btrfs!  Where am I going to get THAT from!
    #
    # Instead, someone on the Internet just builds the binaries with Github Actions. Great, we'll
    # use that.  Unfortunately, that person doesn't build it for Mac.  But worry not, someone forked
    # the repo and does build it for Mac!  That's what we're using.
    #
    # Something else we could do is use the underlying libraries, specifically
    # github.com/containers/v5/copy, directly, right?  They're the whole problem.  You can make it
    # build without a C compiler, but you need to pass a ton of build tags everywhere; gazelle needs
    # them, MODULE.bazel needs them as Gazelle overrides, and Bazel targets that depend on the
    # library need the build tags passed on the bazel build/run command line with --define
    # gotags=....  Adding to the fun, that module also depends on mutually incompatible versions of
    # Open Telemetry somehow, which breaks etcd, somehow.  I tried making it work.  It can be made
    # to work; just upgrade every module you see in the error messages from "go mod tidy" to the
    # latest version.  But it was just too big of a mess for what amounts to copying 3 files between
    # two local directories.  Like, what?  You need to inspect btrfs filesystems during that?
    #
    # So we'll do this terrible thing instead.  Take that, the software supply chain!
    http_file(
        name = "com_github_containers_skopeo_x86_64_linux",
        url = "https://github.com/tgolsson/skopeo-binary/releases/download/v1.14.2/skopeo-linux-amd64",
        sha256 = "51218f93a2b079e36a36f7fbe2d2d86778be0a6947653031b4f9e254e2469224",
        executable = True,
        downloaded_file_path = "skopeo",
    )
    http_file(
        name = "com_github_containers_skopeo_aarch64_linux",
        url = "https://github.com/tgolsson/skopeo-binary/releases/download/v1.14.2/skopeo-linux-arm64",
        sha256 = "364c46085de31edf4b312f13587442f4eade1f181bc5a9ea2ab2ffab5b575916",
        executable = True,
        downloaded_file_path = "skopeo",
    )
    http_file(
        name = "com_github_containers_skopeo_x86_64_macos",
        url = "https://github.com/tgolsson/skopeo-binary/releases/download/v1.14.2/skopeo-darwin-amd64",
        sha256 = "5a82082f439508255fd2588e1aecbe9b61bf5ad62442cde2bab5d3ba665ad63f",
        executable = True,
        downloaded_file_path = "skopeo",
    )
    http_file(
        name = "com_github_containers_skopeo_aarch64_macos",
        url = "https://github.com/tgolsson/skopeo-binary/releases/download/v1.14.2/skopeo-darwin-arm64",
        sha256 = "77a31136eaa04643120b4999e526da35735e01ccc80f1c97752a33b3fed64547",
        executable = True,
        downloaded_file_path = "skopeo",
    )

def pachdev_deps():
    http_file(
        name = "com_github_kubernetes_sigs_metrics_server_helm_chart",
        url = "https://github.com/kubernetes-sigs/metrics-server/releases/download/metrics-server-helm-chart-3.11.0/metrics-server-3.11.0.tgz",
        sha256 = "e7cba5a2d7e2b0949b3ea1aa48fc3780b71eba17e1edccc8273f5da67048164b",
        downloaded_file_path = "metrics-server.tgz",
    )

def kubeconform_deps():
    """https://github.com/yannh/kubeconform"""
    build_file_content = """exports_files(["kubeconform"])"""
    http_archive(
        name = "com_github_yannh_kubeconform_x86_64_linux",
        url = "https://github.com/yannh/kubeconform/releases/download/v0.6.4/kubeconform-linux-amd64.tar.gz",
        sha256 = "2b4ebeaa4d5ac4843cf8f7b7e66a8874252b6b71bc7cbfc4ef1cbf85acec7c07",
        build_file_content = build_file_content,
    )
    http_archive(
        name = "com_github_yannh_kubeconform_aarch64_linux",
        url = "https://github.com/yannh/kubeconform/releases/download/v0.6.4/kubeconform-linux-arm64.tar.gz",
        sha256 = "582a8a3eb3c33feb065928728ddac6e544c4f0b1234fd78fc1e4c4175a422b22",
        build_file_content = build_file_content,
    )
    http_archive(
        name = "com_github_yannh_kubeconform_x86_64_macos",
        url = "https://github.com/yannh/kubeconform/releases/download/v0.6.4/kubeconform-darwin-amd64.tar.gz",
        sha256 = "805bbc2ab900925be01b8e72039d3689c0a6d4f8aa4b2fad24dd1243387a2b18",
        build_file_content = build_file_content,
    )
    http_archive(
        name = "com_github_yannh_kubeconform_aarch64_macos",
        url = "https://github.com/yannh/kubeconform/releases/download/v0.6.4/kubeconform-darwin-arm64.tar.gz",
        sha256 = "95f9097e54c2c1abe9099d0fedc57a2d5c674b2e32231b901a6e60411899afdd",
        build_file_content = build_file_content,
    )

def ibazel_deps():
    """https://github.com/bazelbuild/bazel-watcher"""
    http_file(
        name = "com_github_bazelbuild_bazel_watcher_ibazel_x86_64_linux",
        url = "https://github.com/bazelbuild/bazel-watcher/releases/download/v0.24.0/ibazel_linux_amd64",
        sha256 = "32979dfb172e4695ed92e240ec4374d73e01efac140e74ec054aa08c1a10235b",
        downloaded_file_path = "ibazel",
        executable = True,
    )
    http_file(
        name = "com_github_bazelbuild_bazel_watcher_ibazel_aarch64_linux",
        url = "https://github.com/bazelbuild/bazel-watcher/releases/download/v0.24.0/ibazel_linux_arm64",
        sha256 = "90aa01f1473650d2587c8c353ecb65c1bc748e5e9bf7bb45716ec213bc0987e6",
        downloaded_file_path = "ibazel",
        executable = True,
    )
    http_file(
        name = "com_github_bazelbuild_bazel_watcher_ibazel_x86_64_macos",
        url = "https://github.com/bazelbuild/bazel-watcher/releases/download/v0.24.0/ibazel_darwin_amd64",
        sha256 = "354bab8a58e81d7c8ba0df50799d14d40f56e57323129c94a983b39b755a9d40",
        downloaded_file_path = "ibazel",
        executable = True,
    )
    http_file(
        name = "com_github_bazelbuild_bazel_watcher_ibazel_aarch64_macos",
        url = "https://github.com/bazelbuild/bazel-watcher/releases/download/v0.24.0/ibazel_darwin_arm64",
        sha256 = "3af24483cd7afd6155711acc3e97b36dd408710c8773ebab45729c139cfa2bc4",
        downloaded_file_path = "ibazel",
        executable = True,
    )

def loki_deps():
    """loki and logcli binaries, for //tools."""
    http_archive(
        name = "com_github_grafana_loki_loki_x86_64_linux",
        url = "https://github.com/grafana/loki/releases/download/v2.6.1/loki-linux-amd64.zip",
        sha256 = "f9c577d54e35e5f46b1a9d0e0788299a0c1e69975ee999d27c3dc711942cc973",
        build_file_content = """exports_files(["loki-linux-amd64"])""",
    )
    http_archive(
        name = "com_github_grafana_loki_loki_aarch64_linux",
        url = "https://github.com/grafana/loki/releases/download/v2.6.1/loki-linux-arm64.zip",
        sha256 = "8679a7a6260ae67bdc50d50ba51cfa2817b6aba9ededf63b62dc8670552507a1",
        build_file_content = """exports_files(["loki-linux-arm64"])""",
    )
    http_archive(
        name = "com_github_grafana_loki_loki_x86_64_macos",
        url = "https://github.com/grafana/loki/releases/download/v2.6.1/loki-darwin-amd64.zip",
        sha256 = "ab18da94a900edcda0bee6b59d10f385fd47486282ebf08accc7adb012edd93e",
        build_file_content = """exports_files(["loki-darwin-amd64"])""",
    )
    http_archive(
        name = "com_github_grafana_loki_loki_aarch64_macos",
        url = "https://github.com/grafana/loki/releases/download/v2.6.1/loki-darwin-arm64.zip",
        sha256 = "dd63acd073ef5305f0f90651d75d6fe2810d5f2811ec366bf05e7545eb8aa6ff",
        build_file_content = """exports_files(["loki-darwin-arm64"])""",
    )
    http_archive(
        name = "com_github_grafana_loki_logcli_x86_64_linux",
        url = "https://github.com/grafana/loki/releases/download/v2.6.1/logcli-linux-amd64.zip",
        sha256 = "8fe59a3b2f83c02920f29251a97ad1f6d4a83a8bbbc20097b73fb839d946cb75",
        build_file_content = """exports_files(["logcli-linux-amd64"])""",
    )
    http_archive(
        name = "com_github_grafana_loki_logcli_aarch64_linux",
        url = "https://github.com/grafana/loki/releases/download/v2.6.1/logcli-linux-arm64.zip",
        sha256 = "e2c3cb728402087cadc120e190d617db8e0b39eb21bbff73f0d040e535822ca8",
        build_file_content = """exports_files(["logcli-linux-arm64"])""",
    )
    http_archive(
        name = "com_github_grafana_loki_logcli_x86_64_macos",
        url = "https://github.com/grafana/loki/releases/download/v2.6.1/logcli-darwin-amd64.zip",
        sha256 = "f2c8583460e86fd8ba07cd055340b9f7432cf8df05d5afd333f6277dc420b009",
        build_file_content = """exports_files(["logcli-darwin-amd64"])""",
    )
    http_archive(
        name = "com_github_grafana_loki_logcli_aarch64_macos",
        url = "https://github.com/grafana/loki/releases/download/v2.6.1/logcli-darwin-arm64.zip",
        sha256 = "81dafec9d6357fa3303bc3d913df45e11a364dbc2577fa6b948f2bd0620ab565",
        build_file_content = """exports_files(["logcli-darwin-arm64"])""",
    )
