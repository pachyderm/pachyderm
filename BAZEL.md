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

Note: you will need a C++ compiler installed. `apt install build-essential` on Debian, or do the
xcode dance on Mac OS. The C++ compiler is used to build some internal Bazel tools.

### Setup at Pachyderm

If you'd like to use the shared build cache, join the Buildbuddy organization by logging in with
your pachyderm.io Google account:

https://pachyderm.buildbuddy.io/join/

From there, you'll get an API key. Add a line to `.bazelrc.local` like this:

    build --remote_header=x-buildbuddy-api-key=<key>

Then you will be using our shared cache, and your invocation URLs will be shareable with your
teammates. Note that this feature leaks your username, hostname, and the names of environment
variables (but not the values!) on your workstation to other team members. If any of those are
work-inappropriate, maybe fix that before you add your key!

## Use

=======

## Regenerating protos

Right now, `make proto` calls out to `bazel run //:make_proto` to generate the Go protos. When you
run `make proto`, you're getting your first taste of Bazel. CI checks that you remembered to run
this after editing protos.

=======

## Managing dependencies

Our repository is compatible with the traditional `go` toolchain in addition to Bazel. If you add or
remove dependencies, run `bazel run //:go mod tidy` to update go.mod and go.sum, and run
`bazel run //:gazelle` to update BUILD files. Then run a build or test. Also consider running
`bazel run :buildifier` to format the BUILD files. Sometimes the `buildozer` commands suggested by
Gazelle lead to lint errors; buildifier will auto-fix them for you.

Tests ensure you did all of this correctly. CI runs them and you can run them locally.
TODO(jrockway): List these.

## Run Go

If you'd like to invoke the version of Go used internally, run `bazel run //:go`. For example
`bazel run //:go mod tidy` will tidy go.mod and go.sum (which Bazel uses to load dependencies).

## Tools

### Gazelle

[Gazelle](https://github.com/bazelbuild/bazel-gazelle) is a tool that generates BUILD files based on
Go and Python source code. `bazel run //:gazelle` will run it. You'll need to run it when you add
new Go source files, packages, or change dependencies.

### Buildifier

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

# If you somehow find yourself responsible for bazelifying the repo, `bazel run //:gazelle` will run [Gazelle](https://github.com/bazelbuild/bazel-gazelle).

### Buildozer

If the build prints something like **You can use the following buildozer command to fix these
issues:**, use `bazel run //:buildozer` to invoke
[buildozer](https://github.com/bazelbuild/buildtools/blob/master/buildozer/README.md) as per the
printed instructions. This mostly happens when Go modules are added or removed; `MODULE.bazel`
contains a duplicate list of what's in `go.mod`, and the recommended `buildozer` command keeps them
in sync. If you forget to run this, nothing will work, so it's unlikely that you'll forget.

## Binaries

### pachctl

To run pachctl, `bazel run //:pachctl`. You can also build the binary and copy it to $PATH, if you
like. After `bazel run //:pachctl` or `bazel build //:pachctl`, it will be in
`bazel-bin/src/server/cmd/pachctl/pachctl_/pachctl`.

This version of pachctl has the current version number, based on `workspace_status.sh` baked into
it. Thus, `pachctl version` on the binary will lead you to the exact code it was built with!

## Hints

`bazel run //target -- ...` prevents `...` from being interpreted as an argument to Bazel, which is
useful if you are passing flags to something you're `bazel run`ning.

`bazel query 'deps("//some:target")'` will list all dependencies of `//some:target`.

`bazel query 'path("//some:target", "@@some_library//:whatever")` will show a dependency chain from
`//some:target` to `@@some_library//:whatever`.

`bazel query --output=build ...` will show a BUILD file representing the matched rules.
