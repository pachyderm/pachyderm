#!/usr/bin/env python3

import os
import sys
import gzip
import stat
import shutil
import zipfile
import tempfile
import urllib.request

def main():
    version = sys.argv[1]

    bin_path = os.path.join(os.environ["GOPATH"], "bin", "pachctl")
    is_darwin = sys.platform == "darwin"
    archive_name = "pachctl_{}_{}_amd64".format(version, sys.platform)

    if is_darwin:
        archive_url = "https://github.com/pachyderm/pachyderm/releases/download/v{}/{}.zip".format(version, archive_name)
    else:
        archive_url = "https://github.com/pachyderm/pachyderm/releases/download/v{}/{}.tar.gz".format(version, archive_name)
    
    print("downloading {} to {}".format(archive_url, bin_path))

    with tempfile.TemporaryFile() as temp_file:
        with urllib.request.urlopen(archive_url) as response:
            shutil.copyfileobj(response, temp_file)

        with open(bin_path, "wb") as bin_file:
            if is_darwin:
                with zipfile.ZipFile(file=temp_file) as zip_dir:
                    with zip_dir.open("{}/pachctl".format(archive_name), "r") as zip_file:
                        shutil.copyfileobj(zip_file, bin_file)
            else:
                with gzip.GzipFile(fileobj=temp_file) as gzip_file:
                    shutil.copyfileobj(gzip_file, bin_file)

    st = os.stat(bin_path)
    os.chmod(bin_path, st.st_mode | stat.S_IEXEC)

if __name__ == "__main__":
    main()
