#!/usr/bin/python
"""Regenerate tests for only the files that have changed."""

import argparse
import logging
import os
import subprocess

TOP_LEVEL_EXCLUDES = ["docs", "hack", "tests", "kfdef"]

def generate_test_name(repo_root, package_dir):
  """Generate the name of the go file to write the test to.

  Args:
    repo_root: Root of the repository
    package_dir: Full path to the kustomize directory.
  """

  rpath = os.path.relpath(package_dir, repo_root)

  pieces = rpath.split(os.path.sep)
  name = "-".join(pieces) + "_test.go"

  return name

def get_changed_dirs():
  """Return a list of directories of changed kustomization packages."""
  # Generate a list of the files which have changed with respect to the upstream
  # branch

  # TODO(jlewi): Upstream doesn't seem to work in some cases. I think
  # upstream might end up referring to the
  origin = os.getenv("REMOTE_ORIGIN", "@{upstream}")
  logging.info("Using %s as remote origin; you can override using environment "
               "variable REMOTE_ORIGIN", origin)
  modified_files = subprocess.check_output(
    ["git", "diff", "--name-only", origin])

  repo_root = subprocess.check_output(["git", "rev-parse", "--show-toplevel"])
  repo_root = repo_root.decode()
  repo_root = repo_root.strip()

  modified_files = modified_files.decode().split()
  changed_dirs = set()
  for f in modified_files:
    changed_dir = os.path.dirname(os.path.join(repo_root, f))
    # Check if its a kustomization directory by looking for a kustomization
    # file
    if not os.path.exists(os.path.join(repo_root, changed_dir,
                                       "kustomization.yaml")):
      continue

    changed_dirs.add(changed_dir)

    # If the base package changes then we need to regenerate all the overlays
    # We assume that the base package is named "base".

    if os.path.basename(changed_dir) == "base":
      parent_dir = os.path.dirname(changed_dir)
      for root, _, files in os.walk(parent_dir):
        for f in files:
          if f == "kustomization.yaml":
            changed_dirs.add(root)

  return changed_dirs

def find_kustomize_dirs(root):
  """Find all kustomization directories in root"""

  changed_dirs = set()
  for top in os.listdir(root):
    if top.startswith("."):
      logging.info("Skipping directory %s", os.path.join(root, top))
      continue

    if top in TOP_LEVEL_EXCLUDES:
      continue

    for child, _, files in os.walk(os.path.join(root, top)):
      for f in files:
        if f == "kustomization.yaml":
          changed_dirs.add(child)

  return changed_dirs

def remove_unmatched_tests(repo_root, package_dirs):
  """Remove any tests that don't map to a kustomization.yaml file.

  This ensures tests don't linger if a package is deleted.
  """

  # Create a set of all the expected test names
  expected_tests = set()

  for d in package_dirs:
    expected_tests.add(generate_test_name(repo_root, d))

  tests_dir = os.path.join(repo_root, "tests")
  for name in os.listdir(tests_dir):
    if not name.endswith("_test.go"):
      continue

    # TODO(jlewi): we should rename and remove "_test.go" from the filename.
    if name in ["kusttestharness_test.go"]:
      logging.info("Ignoring %s", name)
      continue

    if not name in expected_tests:
      test_file = os.path.join(tests_dir, name)
      logging.info("Deleting test %s; doesn't map to kustomization file",
                   test_file)
      os.unlink(test_file)

if __name__ == "__main__":
  logging.basicConfig(
      level=logging.INFO,
      format=('%(levelname)s|%(asctime)s'
              '|%(pathname)s|%(lineno)d| %(message)s'),
      datefmt='%Y-%m-%dT%H:%M:%S',
  )
  logging.getLogger().setLevel(logging.INFO)

  parser = argparse.ArgumentParser()

  parser.add_argument(
      "--all",
      dest = "all_tests",
      action = "store_true",
      help="Regenerate all tests")

  parser.set_defaults(all_tests=False)

  args = parser.parse_args()

  repo_root = subprocess.check_output(["git", "rev-parse", "--show-toplevel"])
  repo_root = repo_root.decode()
  repo_root = repo_root.strip()

  # Get a list of package directories
  package_dirs = find_kustomize_dirs(repo_root)

  remove_unmatched_tests(repo_root, package_dirs)

  if args.all_tests:
    changed_dirs = package_dirs
  else:
    changed_dirs = get_changed_dirs()

  for full_dir in changed_dirs:
    test_name = generate_test_name(repo_root, full_dir)
    logging.info("Regenerating test %s for %s ", test_name, full_dir)

    test_path = os.path.join(repo_root, "tests", test_name)

    with open(test_path, "w") as test_file:
      subprocess.check_call(["./hack/gen-test-target.sh", full_dir],
                            stdout=test_file, cwd=repo_root)
