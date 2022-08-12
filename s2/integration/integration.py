#!/usr/bin/env python3

import os
import sys
import argparse
import subprocess
from urllib.parse import urlparse

ROOT = os.path.dirname(os.path.abspath(__file__))
TESTDATA = os.path.join(ROOT, "testdata")

def main():
    parser = argparse.ArgumentParser(description="Runs the s2 integration test suite.")
    parser.add_argument("address", help="Address of the s2 instance")
    parser.add_argument("--access-key", default="", help="Access key")
    parser.add_argument("--secret-key", default="", help="Secret key")
    parser.add_argument("--test", default=None, help="Run a specific test")
    args = parser.parse_args()

    suite_filter = None
    test_filter = None
    if args.test is not None:
        parts = args.test.split(":", maxsplit=1)
        if len(parts) == 1:
            suite_filter = parts[0]
        else:
            suite_filter, test_filter = parts

    # Create some sample data if it doesn't exist yet
    if not os.path.exists(TESTDATA):
        os.makedirs(TESTDATA)
        with open(os.path.join(TESTDATA, "small.txt"), "w") as f:
            f.write("x")
        with open(os.path.join(TESTDATA, "large.txt"), "w") as f:
            f.write("x" * (10 * 1024 * 1024))

    url = urlparse(args.address)

    env = dict(os.environ)
    env["S2_HOST_ADDRESS"] = args.address
    env["S2_HOST_NETLOC"] = url.netloc
    env["S2_HOST_SCHEME"] = url.scheme
    env["S2_ACCESS_KEY"] = args.access_key
    env["S2_SECRET_KEY"] = args.secret_key

    def run(cwd, *args):
        subprocess.run(args, cwd=os.path.join(ROOT, cwd), env=env, check=True)

    try:
        if suite_filter is None or suite_filter == "python":
            args = ["-k", test_filter] if test_filter is not None else []
            run("python", os.path.join("venv", "bin", "pytest"), "test.py", *args)
        if suite_filter is None or suite_filter == "go":
            args = ["-count=1"]
            if test_filter is not None:
                args.append("-run={}".format(test_filter))
            args.append("./...")
            run("go", "go", "test", *args)
        if suite_filter is None or suite_filter == "cli":
            run("cli", "bash", "test.sh")
    except subprocess.CalledProcessError:
        sys.exit(1)

if __name__ == "__main__":
    main()
