# Bazel

You may have noticed some `BUILD.bazel` files throughout the repository. We are in the process of
building everything here with [Bazel](https://bazel.build/), so that day 1 setup is simplified, and
so that test tooling can be shared between languages.

Right now, only proto generation requires Bazel. In the future, everything will be possible to work
on with normal Go tools, but it will also be possible to use Bazel. It is likely that most
Pachyderm-specific development tasks, like deploying a dev environment and syncing your code changes
to it, will require Bazel. We'll probably also use it for CI, because of the test result caching.
But, because people depend on `github.com/pachyderm/pachyderm/v2` with "go get", we can never fully
convert to only using Bazel as the build system. Rather, we will use both Bazel and go.mod for the
foreseeable future. Hence, we check in all of our generated code.

Additionally, the relase process will also use `goreleaser` for a short while.

TL;DR: Bazel is strictly an overlay unless you actively work on this codebase day-to-day.

## Setup

To run `make proto`, you will now need Bazel. The best way to install Bazel is by installing
Bazelisk as the `bazel` binary in your path:

https://github.com/bazelbuild/bazelisk/blob/master/README.md

`brew install bazelisk` does this for you automatically. If you download and install the binary
manually, move it to somewhere in `$PATH`, set it executable, and name it `bazel`.

Once Bazelisk is installed, it will fetch the version of Bazel specified in the file
`.bazelversion`, and use that. If the repository needs features in a newer version, you will
automatically update the next time you run a Bazel command. You shouldn't need to ever update
Bazelisk, but feel free of course.

### Setup at Pachyderm

If you'd like to use the shared cache, join the Buildbuddy organization by logging in with your
pachyderm.io Google account:

https://pachyderm.buildbuddy.io/join/

From there, you'll get an API key. Add a line to `.bazelrc.local` like this:

    build --remote_header=x-buildbuddy-api-key=<key>

Then you will be using our shared cache, and your invocation URLs will be shareable with your
teammates. Note that this feature leaks your username, hostname, and the names of environment
variables (but not the values!) on your workstation to other team members. If any of those are
work-inappropriate, maybe fix that before you add your key!

## Use

Right now, `make proto` calls out to `bazel run //:make_proto` to generate the Go protos. When you
run `make proto`, you're getting your first taste of Bazel. CI checks that you remembered to run
this after editing protos.

If you edit BUILD files, run `bazel test //:buildifier_test` (potentially with `--test_output=all`)
to see if you introduced any lint errors. If you did, run `bazel run //:buildifier` to auto-fix
them. You can also install
[buildifier](https://github.com/bazelbuild/buildtools/blob/master/buildifier/README.md) and have
your editor run it on save for `*.bazel` files. Be aware that the lint rules change based on the
dialect of Starlark (raw Starlark, BUILD.bazel, WORKSPACE, etc.), so your editor needs to tell
`buildifier` the filename it's editing. If not, you will still have lint errors in your saved files.

If you'd like to invoke the version of Go used for proto generation, run `bazel run //:go`.

If the build prints something like "The following buildozer command can fix this...", use
`bazel run //:buildozer` to invoke
[buildozer](https://github.com/bazelbuild/buildtools/blob/master/buildozer/README.md).

If you somehow find yourself responsible for bazelifying the repo, `bazel run //:gazelle` will run
[Gazelle](https://github.com/bazelbuild/bazel-gazelle).

## Hints

`bazel run //target -- ...` prevents `...` from being interpreted as an argument to Bazel, which is
useful if you are passing flags to something you're `bazel run`ning.
