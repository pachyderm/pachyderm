#!/usr/bin/env python3

import os
import sys
import stat
import shutil
import tarfile
import zipfile
import tempfile
import urllib.request

def fetch_darwin(version, bin_path):
    archive_name = "pachctl_{}_darwin_amd64".format(version)
    archive_url = "https://github.com/pachyderm/pachyderm/releases/download/v{}/{}.zip".format(version, archive_name)

    with tempfile.TemporaryFile() as temp_file:
        with urllib.request.urlopen(archive_url) as response:
            shutil.copyfileobj(response, temp_file)

        with zipfile.ZipFile(file=temp_file) as zip_dir:
            with zip_dir.open("{}/pachctl".format(archive_name), "r") as zip_file:
                with open(bin_path, "wb") as bin_file:
                    shutil.copyfileobj(zip_file, bin_file)

def fetch_linux(version, bin_path):
    archive_name = "pachctl_{}_linux_amd64".format(version)
    archive_url = "https://github.com/pachyderm/pachyderm/releases/download/v{}/{}.tar.gz".format(version, archive_name)

    with tempfile.TemporaryFile() as temp_file:
        with urllib.request.urlopen(archive_url) as response:
            shutil.copyfileobj(response, temp_file)

        temp_file.seek(0)

        with tarfile.open(fileobj=temp_file, mode="r:gz") as tar_file:
            with tar_file.extractfile("pachctl_{}_linux_amd64/pachctl".format(version)) as f:
                with open(bin_path, "wb") as bin_file:
                    shutil.copyfileobj(f, bin_file)

def main():
    version = sys.argv[1]
    # TODO(msteffen) replace this with 'go env GOPATH', in case go's default
    # changes
    go_path = os.getenv("GOPATH", os.path.join(os.environ["HOME"], "go"))
    go_bin = os.getenv("GOBIN", os.path.join(go_path, "bin"))
    pachctl_path = os.path.join(go_bin, "pachctl")

    if sys.platform == "darwin":
        fetch_darwin(version, pachctl_path)
    else:
        fetch_linux(version, pachctl_path)

    st = os.stat(pachctl_path)
    os.chmod(pachctl_path, st.st_mode | stat.S_IEXEC)

if __name__ == "__main__":
    main()
