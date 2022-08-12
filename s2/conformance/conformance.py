#!/usr/bin/env python3

import os
import re
import sys
import glob
import time
import argparse
import datetime
import subprocess
import collections

ROOT = os.path.dirname(os.path.abspath(__file__))
S3TESTS_ROOT = os.path.join(ROOT, "s3-tests")

RAN_PATTERN = re.compile(r"^Ran (\d+) tests in [\d\.]+s")
FAILED_PATTERN = re.compile(r"^FAILED \((SKIP=(\d+))?(, )?(errors=(\d+))?(, )?(failures=(\d+))?\)")

ERROR_PATTERN = re.compile(r"^(FAIL|ERROR): (.+)")

TRACEBACK_PREFIXES = [
    "Traceback (most recent call last):",
    "----------------------------------------------------------------------",
    "  ",
]

# Ignore tests with these attributes. They're ignored because s2 itself
# doesn't support this functionality.
BLACKLISTED_ATTRIBUTES = [
    "list-objects-v2",
    "cors",
    "lifecycle",
    "encryption",
    "bucket-policy",
    "tagging",
    "object-lock",
    "appendobject",
]

def compute_stats(filename):
    ran = 0
    skipped = 0
    errored = 0
    failed = 0

    with open(filename, "r") as f:
        for line in f:
            match = RAN_PATTERN.match(line)
            if match:
                ran = int(match.groups()[0])
                continue

            match = FAILED_PATTERN.match(line)
            if match:
                groups = match.groups()
                skipped_str = groups[1]
                errored_str = groups[4]
                failed_str = groups[7]
                skipped = int(skipped_str) if skipped_str is not None else 0
                errored = int(errored_str) if errored_str is not None else 0
                failed = int(failed_str) if failed_str is not None else 0

    if ran != 0:
        return (ran - skipped - errored - failed, ran - skipped)
    else:
        return (0, 0)

def run_nosetests(config, test=None, env=None, stderr=None):
    config = os.path.abspath(config)
    if not os.path.exists(config):
        print("config file does not exist: {}".format(config), file=sys.stderr)
        sys.exit(1)

    all_env = dict(os.environ)
    all_env["S3TEST_CONF"] = config
    if env is not None:
        all_env.update(env)

    pwd = os.path.join(ROOT, "s3-tests")
    args = [os.path.join("virtualenv", "bin", "nosetests"), "-a", ",".join("!{}".format(a) for a in BLACKLISTED_ATTRIBUTES)]
    if test is not None:
        args.append(test)

    proc = subprocess.run(args, env=all_env, cwd=pwd, stderr=stderr)
    print("Test run exited with {}".format(proc.returncode))

def print_failures(runs_dir):
    log_files = sorted(glob.glob(os.path.join(runs_dir, "*.txt")))

    if len(log_files) == 0:
        print("No log files found", file=sys.stderr)
        sys.exit(1)

    old_stats = None
    if len(log_files) > 1:
        old_stats = compute_stats(log_files[-2])

    filepath = log_files[-1]
    stats = compute_stats(filepath)

    if old_stats:
        print("Overall results: {}/{} (vs last run: {}/{})".format(*stats, *old_stats))
    else:
        print("Overall results: {}/{}".format(*stats))

    failing_test = None
    causes = collections.defaultdict(lambda: [])
    with open(filepath, "r") as f:
        for line in f:
            line = line.rstrip()

            if failing_test is None:
                match = ERROR_PATTERN.match(line)
                if match is not None:
                    failing_test = match.groups()[1]
            else:
                if not any(line.startswith(p) for p in TRACEBACK_PREFIXES):
                    causes[line].append(failing_test)
                    failing_test = None

    causes = sorted(causes.items(), key=lambda i: len(i[1]), reverse=True)
    for (cause_name, failing_tests) in causes:
        if len(cause_name) > 160:
            print("{} [...]:".format(cause_name[:160]))
        else:
            print("{}:".format(cause_name))
        
        for failing_test in failing_tests:
            print("- {}".format(failing_test))

def main():
    parser = argparse.ArgumentParser(description="Runs the s2 conformance test suite.")
    parser.add_argument("--no-run", default=False, action="store_true", help="Disables a test run, and just prints failure data from the last test run")
    parser.add_argument("--test", default="", help="Run a specific test")
    parser.add_argument("--s3tests-config", required=True, help="Path to the s3-tests config file")
    parser.add_argument("--ignore-config", default=None, help="Path to the ignore config file")
    parser.add_argument("--runs-dir", default=None, help="Path to the directory holding test runs")
    args = parser.parse_args()

    if (not args.runs_dir) and (args.no_run or not args.test):
        print("Must specify `--runs-dir`", file=sys.stderr)
        sys.exit(1)

    if args.no_run:
        print_failures(args.runs_dir)
        return

    if args.test:
        print("Running test {}".format(args.test))

        # In some places, nose and its plugins expect tests to be
        # specified as testmodule.testname, but here, it's expected to be
        # testmodule:testname. This replaces the last . with a : so that
        # the testmodule.testname format can be used everywhere, including
        # here.
        if "." in args.test and not ":" in args.test:
            test = ":".join(args.test.rsplit(".", 1))
        else:
            test = args.test

        run_nosetests(args.s3tests_config, test=test)
    else:
        print("Running tests")

        if args.ignore_config:
            # This uses the `nose-exclude` plugin to exclude tests for
            # unsupported features. Note that `nosetest` does have a built-in
            # way of excluding tests, but it only seems to match on top-level
            # modules, rather than on specific tests.
            ignores = []

            with open(args.ignore_config, "r") as f:
                for line in f:
                    line = line.strip()
                    if line and not line.startswith("#"):
                        ignores.append(line)

            env = {
                "NOSE_EXCLUDE_TESTS": ";".join(ignores)
            }
        else:
            env = None

        filepath = os.path.join(args.runs_dir, datetime.datetime.now().strftime("%Y-%m-%d-%H-%M-%S.txt"))
        with open(filepath, "w") as f:
            run_nosetests(args.s3tests_config, env=env, stderr=f)

        print_failures(args.runs_dir)
                
if __name__ == "__main__":
    main()
