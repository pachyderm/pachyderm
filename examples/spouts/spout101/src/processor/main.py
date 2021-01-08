#!/usr/local/bin/python3
import glob
import os
from datetime import datetime

file_paths = glob.glob(os.path.join("/pfs/spout", "*.txt"))

# logs the files in the input repo based on their size.
for file_path in file_paths:
    size = os.stat(file_path).st_size
    now = datetime.now().time()
    if (size == 1024):
        with open("/pfs/out/"+"1K"+".txt", 'w') as onek_files:
            onek_files.write(str(now) + " " + file_path + "\n")
    elif (size == 2048):
        with open("/pfs/out/"+"2K"+".txt", 'w') as twok_files:
            twok_files.write(str(now) + " " + file_path + "\n")
    else:
        print("Not matching size" + size + " " + file_path)
