# Notes about Bazel in this project.

### Dependencies and Virtual Environment.

Bazel is concerned with (hermetic builds)[https://bazel.build/basics/hermeticity], and this requires
a more careful and rigorous handling of dependencies than would normally be required for a python
project.

Prior to the introduction of bazel, the dependencies for this project were specified within the
setup.cfg file, but as this is a somewhat antiquated location there are no Bazel rules that are
compatible with this location. These dependencies have been moved to the `requirements.txt` and
`requirements-dev.txt` files. However, for bazel to run "hermetic" builds for this project, these
third-party dependencies must be "locked" to ensure that all developers and tests are run using the
exact same versions of these dependencies. These lock files are `requirements-lock.txt` and
`requirements-dev-lock.txt`. The following command regenerate these lock files:

```
bazel run :requirements.update
bazel run :requirements-dev.update
```

To ensure developers are all using the same packages in their development environment, bazel can
also create a virtual environment which contains these locked dependencies:

```
bazel run :venv
```

Note that Python requires different versions of dependencies depending on the host platform, i.e.
Linux and Mac require different versions of modules. When you run `:requirements.update`, it only
updates for your current platform. So if you run this on a Mac, CI will fail because it's Linux, and
you should also run it on your Linux VM and check in those generated changes.


### Gazelle
The jupyter-extension project is now hooked into the gazelle tooling used throughout the repo.
Gazelle will mark intra-project dependencies automatically when invoked. To run gazelle:
```
bazel run //:gazelle
```
When adding a new python package as a dependency, the gazelle_python manifest will need to be updated.
This can be done with the following command:
```
bazel run :gazelle_python_manifest.update
```
