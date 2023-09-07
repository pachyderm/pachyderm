load("build", "download_file", "go_binary", "path")

dumb_init = download_file(
    name = "dumb-init",
    by_platform = {
        "linux/amd64": {
            "url": "https://github.com/Yelp/dumb-init/releases/download/v1.2.5/dumb-init_1.2.5_x86_64",
            "digest": "blake3:9a520c3860a67bca23323e2dfa9e263f8dd54000b1c890b44db2a5316c607284",
        },
        "linux/arm64": {
            "url": "https://github.com/Yelp/dumb-init/releases/download/v1.2.5/dumb-init_1.2.5_aarch64",
            "digest": "blake3:8df4e75473552405410b17e555da4b4c493857627bfdb3c0fd9be45c79827182",
        },
    },
)

worker = go_binary(
    workdir = path(".."),
    target = "./src/server/cmd/worker",
)

worker_init = go_binary(
    workdir = path(".."),
    target = "./etc/worker",
)
