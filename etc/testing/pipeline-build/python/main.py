import os
import argparse
from leftpad import left_pad

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--content", default="0", help="What content to left pad with")
    parser.add_argument("--length", default=4, type=int, help="Pad target length")
    parser.add_argument("--input", default="/pfs/in", help="Input directory")
    parser.add_argument("--output", default="/pfs/out", help="Output directory")
    args = parser.parse_args()

    for fname in os.listdir(args.input):
        if os.path.isfile(os.path.join(args.input, fname)):
            with open(os.path.join(args.input, fname), "r") as f_in:
                with open(os.path.join(args.output, fname), "w") as f_out: 
                    for line in f_in:
                        f_out.write("{}\n".format(left_pad(line, args.length, args.content)))

if __name__ == "__main__":
    main()
