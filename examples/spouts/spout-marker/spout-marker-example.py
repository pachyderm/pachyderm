#!/usr/bin/env python3
import time
import tarfile
import os
import stat
import io
import argparse

def open_pipe(path_to_file, attempts=0, timeout=3, sleep_int=5):
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
            print ('{0} attempt of {1}; error opening spout: {2}'.format(attempts + 1, timeout, oe))
            os.umask(umask_original)
            time.sleep(sleep_int)
            return open_pipe(path_to_file, attempts + 1)
        finally:
            os.umask(umask_original)
    return None

def open_marker(path_to_file, attempts=0, timeout=3, sleep_int=5):
    if attempts < timeout :
        try:
            file = open(path_to_file, "r")
            return file.read().split("\n")
        except Exception as oe:
            print ('{0} attempt of {1}; error opening marker: {2}'.format(attempts + 1, timeout, oe))
            time.sleep(sleep_int)
            return open_marker(path_to_file, attempts + 1)
    print ('no marker file, using empty array')
    return []

def main():
    output_character = os.getenv('OUTPUT_CHARACTER', '.')
    parser = argparse.ArgumentParser(description='Add a single character every 10 seconds to a file called output.')
    parser.add_argument('-c', '--output_character', required=False,
                        help="""Character to output.""",
                        default=output_character)
    args = parser.parse_args()

    # Set the character we're going to output.

    print ('going to output: {0}'.format(args.output_character))
        
    # Open the named pipe to start the spout's output.
    print ('opening output pipe /pfs/out')
    spout = open_pipe('/pfs/out')

    # Read in what we've output so far from the marker.
    # Each line is a history of the output.
    # The last line is the most recent.
    print ('opening marker file /pfs/mymark')
    lines = open_marker('/pfs/mymark')

    if spout != None:
        while True:
            lines.append((lines[-1] if len(lines) > 0 else "") + args.output_character)
            marker_contents = "\n".join(lines).encode("utf-8")
            output_contents=lines[-1].encode("utf-8")
            print ('marker contains: {0}'.format(marker_contents))
            print ('output contains: {0}'.format(output_contents))
            time.sleep(10)
            with tarfile.open(fileobj=spout, mode="w|", encoding="utf-8") as tar_stream:
                # This is our marker.  It'll be hidden from the output repo, but
                # will be visible to us in /pfs/mymark when the spout starts.
                tar_info_marker = tarfile.TarInfo("mymark")
                tar_info_marker.size = len(marker_contents)
                tar_info_marker.mode = 0o600
                tar_stream.addfile(tarinfo=tar_info_marker, fileobj=io.BytesIO(marker_contents))
                # This is the current output.  This file is named "output".
                # It will be visible to any pipelines which use this spout as their input.
                # It will contain the last line of contents, above.
                tar_info_output = tarfile.TarInfo("output");
                tar_info_output.size = len(output_contents)
                tar_info_output.mode = 0o600
                tar_stream.addfile(tarinfo=tar_info_output, fileobj=io.BytesIO(output_contents))
                tar_stream.close()
        else:
            print ("couldn't open /pfs/out")

if __name__== "__main__":
  main()
