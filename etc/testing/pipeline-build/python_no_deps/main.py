import os
import sys
import argparse

LENGTH = 4
INPUT_DIRECTORY = "/pfs/in"
OUTPUT_DIRECTORY = "/pfs/out"

def main():
    pad_char = "0" if len(sys.argv) <= 1 else sys.argv[1]
    
    for fname in os.listdir(INPUT_DIRECTORY):
        if os.path.isfile(os.path.join(INPUT_DIRECTORY, fname)):
            with open(os.path.join(INPUT_DIRECTORY, fname), "r") as f_in:
                contents = f_in.read()
                if len(contents) < LENGTH:
                    contents = (pad_char * (LENGTH - len(contents))) + contents
                with open(os.path.join(OUTPUT_DIRECTORY, fname), "w") as f_out: 
                    f_out.write("{}".format(contents))

if __name__ == "__main__":
    main()
