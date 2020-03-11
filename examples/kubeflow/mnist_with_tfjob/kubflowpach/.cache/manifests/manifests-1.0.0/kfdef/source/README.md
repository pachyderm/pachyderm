This directory contains kustomization packages that are used to generate the YAML specs for the KFDef.

The script `hack/build_kfdef_specs.py` is used to run kustomize and output the YAML files.

## Subdirectories

Each sub-directory corresponds to a kustomize package corresponding to a different release
of Kubeflow.

* **master**: This is the base kustomization package
  * In general when adding new applications or making other KFDef specs that should be carried throughout
    future versions you would make here

* **vX.Y.Z**: This is the kustomization package for Kubeflow release x.y.x. It will
   typically reference another version as its base and define patches to apply the appropriate
   modifications.
