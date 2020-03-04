#!/usr/bin/env python3
import time
import tarfile
import os
import stat
import io

def open_pipe(path_to_file, attempts=0, timeout=2, sleep_int=5):
    if attempts < timeout :
        flags = os.O_WRONLY  # Refer to "man 2 open".
        mode = stat.S_IWUSR  # This is 0o400.
        umask = 0o777 ^ mode  # Prevents always downgrading umask to 0.
        umask_original = os.umask(umask)
        try:
            file = os.open(path_to_file, flags, mode)
            # you must open the pipe as binary to prevent line-buffering problems.
            return os.fdopen(file, "wb")
        except OSError as oe:
            print ('{0} attempt of {1}; error opening file: {2}'.format(attempts + 1, timeout, oe))
            os.umask(umask_original)
            time.sleep(sleep_int)
            return open_pipe(path_to_file, attempts + 1)
        finally:
            os.umask(umask_original)
    return None
try:
    with open("/pfs/mymark", "r") as f:
        lines = f.read().split("\n")
except:
    lines = []
while True:
    lines.append((lines[-1] if len(lines) > 0 else "") + ".")
    contents = "\n".join(lines).encode("utf-8")
    time.sleep(30)
    attempts = 0
    try:
        spout = open_pipe('/pfs/out')
        with tarfile.open(fileobj=spout, mode="w|", encoding="utf-8") as tar_stream:
            tar_info = tarfile.TarInfo("mymark")
            tar_info.size = len(contents)
            tar_info.mode = 0o600
            tar_stream.addfile(tarinfo=tar_info, fileobj=io.BytesIO(contents))
        spout.close()
    except:
        attempts += 1
        if attempts == 5:
            raise
        time.sleep(0.1)
