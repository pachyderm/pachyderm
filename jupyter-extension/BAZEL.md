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
`requirements-dev-lock.txt` (ideally these would "lock" to the same file). The following command
regenerate these lock files:

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
