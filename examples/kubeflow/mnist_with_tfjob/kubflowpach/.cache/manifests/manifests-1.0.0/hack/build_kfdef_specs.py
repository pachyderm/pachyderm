"""Generate KFDef YAML from kustomize packages.

This is a helper tool aimed at generating the RAW Yaml for KFDef specs into
kubeflow/manifests.

We use kustomize to make it easier to generate KFDef YAML files corresponding
to different KF versions but we don't want users to be exposed to that.
"""

import fire
import logging
import os
import subprocess
import tempfile
import yaml

RESOURCE_PREFIX = "kfdef.apps.kubeflow.org_v1_kfdef_"

class KFDefBuilder:
  @staticmethod
  def run():
    root = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))

    kfdef_dir = os.path.join(root, "kfdef")
    source_dir = os.path.join(root, "kfdef", "source")

    # Walk over all versions
    for base, dirs, _ in os.walk(source_dir):
      for version in dirs:
        package_dir = os.path.join(base, version)

        # Create a temporary directory to write all the kustomize output to
        temp_dir = tempfile.mkdtemp()

        subprocess.check_call(["kustomize", "build", package_dir, "-o",
                               temp_dir])

        for f in os.listdir(temp_dir):
          new_name = f[len(RESOURCE_PREFIX):]

          # To preserve the existing pattern for now master files are just
          # named kfctl_?.Yaml
          # whereas version files are named kfctl_?.version.yaml
          # in subsequent PRs we might change that

          if version == "master":
            ext = ".yaml"
          else:
            ext = "." + version + ".yaml"

          basename, _ = os.path.splitext(new_name)
          new_name = basename + ext

          new_file = os.path.join(kfdef_dir, new_name.replace("-", "_"))
          logging.info(f"Processing file: {f} -> {new_file}")

          with open(os.path.join(temp_dir, f)) as hf:
            spec = yaml.load(hf)

          # Remove the name. Kustomize requires a name but we don't want
          # a name so that kfctl will fill it in based on the app directory
          del spec["metadata"]["name"]

          with open(new_file, "w") as hf:
            yaml.safe_dump(spec, hf)

if __name__ == "__main__":

  logging.basicConfig(level=logging.INFO,
                      format=('%(levelname)s|%(asctime)s'
                              '|%(message)s|%(pathname)s|%(lineno)d|'),
                      datefmt='%Y-%m-%dT%H:%M:%S',
                      )

  fire.Fire(KFDefBuilder)
