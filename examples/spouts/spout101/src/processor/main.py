#!/usr/local/bin/python3
import glob
import os
from datetime import datetime


def main():
    file_paths = glob.glob(os.path.join("/pfs/spout", "*.txt"))
    os.makedirs("/pfs/out/1K.txt", exist_ok=True)
    os.makedirs("/pfs/out/2K.txt", exist_ok=True)

    # separates the files from the input repo based on their size by
    # writing each to a new file in either the 1K or 2K directory
    for file_path in file_paths:
        size = os.stat(file_path).st_size
        now = datetime.now().time()
        base_name = os.path.basename(file_path)

        if size == 1024:
            with open("/pfs/out/1K.txt/" + base_name, "w") as onek_files:
                onek_files.write(str(now) + " " + file_path + "\n")
        elif size == 2048:
            with open("/pfs/out/2K.txt/" + base_name, "w") as twok_files:
                twok_files.write(str(now) + " " + file_path + "\n")
        else:
            print("Not matching size" + size + " " + file_path)


if __name__ == "__main__":
    main()
