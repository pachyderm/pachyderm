# Bazel

You may have noticed some `BUILD.bazel` files throughout the repository. We are in the process of
building everything here with [Bazel](https://bazel.build/), so that day 1 setup is simplified and
so that test tooling can be shared between languages. We also want to cache tests so that editing
README.md doesn't involve running all of our Go tests.

Currently, Bazel is only for environment setup, `make proto`, and Go tests that don't require
Kubernetes. Because people depend on installing `github.com/pachyderm/pachyderm/v2` with "go get",
we can never fully convert to only using Bazel as the build system. Rather, we will use both Bazel
and `go.mod` for the foreseeable future. Hence, we check in all of our generated code, and you can
use `go test` for everything if you hate Bazel for some reason.

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
xcode dance on Mac OS. The C++ compiler is used to build some internal Bazel tools like the shell
script wrapper.

Note: Every time you update Mac OS, you need to run
`/usr/sbin/softwareupdate --install-rosetta --agree-to-license`. Some of our dependencies don't
build `darwin_arm64` binaries, and we fall back to `darwin_amd64` binaries. Rosetta is what lets
M1/M2/M3 Macs run `amd64` binaries. You will see `Bad CPU type in executable` if Rosetta isn't
working. (You may need to run `bazel shutdown` after installing Rosetta, though this is only
suggested by Apple and doesn't appear to be required 100% of the time.)

Note: you will need a Python 3 interpreter installed as `python3`. `apt install python3` on Debian.
This is because `rules_python` uses the host `python3` to find the version of Python installed via
`rules_python` in `MODULE.bazel`.

### Setup at Pachyderm

If you'd like to use the shared build cache, join the Buildbuddy organization by logging in with
your pachyderm.io Google account:

https://pachyderm.buildbuddy.io/join/

From there, you'll get an API key. Add lines to `.bazelrc.local` like this:

    build --config=remotecache
    build --remote_header=x-buildbuddy-api-key=<key>

Then you will be using our shared cache, and your invocation URLs (which contain build logs, test
logs, profiles, etc.) will be share-able with your teammates.

Note: this feature leaks your username, hostname, and the names of environment variables (but not
the values!) on your workstation to other team members. If any of those are work-inappropriate,
maybe fix that before you add your key!

### Setup in the office

If you work in the office, you won't be able to use the remote cache. GRPC is disallowed on the
network and the remote cache uses GRPC.

You will also need to tell Bazel to trust the TLS MITM certificate, as the build will download
dependencies from the Internet. On Debian or Ubuntu, install `ca-certificates-java`, add the MITM
proxy's certificate file to `/etc/ssl/certs`, then run `sudo update-ca-certficates`.
`update-ca-certificates` copies all certs in `/etc/ssl/certs` into the Java "trust store", which is
how Java programs (which Bazel is) get trusted certs. Then tell Bazel to use the system trust store
by adding to `/etc/bazelrc` or `$HOME/.bazelrc`:

    startup --host_jvm_args=-Djavax.net.ssl.trustStore=/etc/ssl/certs/java/cacerts \
            --host_jvm_args=-Djavax.net.ssl.trustStorePassword=changeit

(Yes, it appears that you need a password to read certs on your own system, and the password is
literally `changeit`.)

Note that just grabbing a Java trust store with the MITM certificate isn't sufficient. `bazelisk`
downloads `bazel` whenever we use a new version of Bazel, and `bazelisk` is written in Go, which
needs the cert in `/etc/ssl/certs`.

## Use

## Write Go code

Writing Go code works like it always did. Your editor can understand the entire source tree without
any help. Whenever you add or delete files, or change dependencies, run `bazel run //:gazelle` to
update the associated BUILD files for Bazel.

## Regenerating protos

Right now, `make proto` calls out to `bazel run //:make_proto` to generate the Go protos. When you
run `make proto`, you're getting your first taste of Bazel. CI checks that you remembered to run
this after editing protos.

## Managing dependencies

Our repository is compatible with the traditional `go` toolchain in addition to Bazel. If you add or
remove dependencies, run `bazel run //:go mod tidy` to update go.mod and go.sum, and run
`bazel run //:gazelle` to update BUILD files. Then run a build or test. Also consider running
`bazel run :buildifier` to format the BUILD files. Sometimes the `buildozer` commands suggested by
Gazelle lead to lint errors; buildifier will auto-fix them for you.

The `bazel-style-tests` in CircleCI ensure you did all of this correctly.

## Run Go

If you'd like to invoke the version of Go used internally, run `bazel run //:go`. For example
`bazel run //:go mod tidy` will tidy go.mod and go.sum (which Bazel uses to load dependencies).

## Build containers

The containers are defined in the `oci/` directory. `bazel build //oci:pachd_image`, for example,
will build a container image compiled for linux on your host machine's architecture. You can copy
this into a registry to pull with `skopeo copy oci:bazel-bin/oci/pachd_image ...`.

Very soon, pushing to a dev environment will be automated.

## Run Tests

`bazel test ...` will run all tests. Test caching means tests whose results couldn't have changed
since the last time they were run will not actually be run, so if you're just editing one file it's
not much slower to run this. You can also pick the test you want to run, like
`bazel test //src/internal/archiveserver:archiveserver_test`. `--test_output=streamed` will show you
the messages from the test as they run (overwhelming if `...`), otherwise, bazel will print the path
of the test logs for failing tests. They are also available in the BuildBuddy UI.

For more Go details, see
https://github.com/bazelbuild/rules_go/blob/master/docs/go/core/rules.md#go_test

In the past, many tests had dependencies on a freshly-built `pachctl` or `match` binary. This is not
the case if you run the tests with Bazel; Bazel automatically includes the version of those binaries
based on your working tree in the test's environment.

### Run one test case

`bazel test //the:test --test_filter=TestCaseIWantToRun`

## Tools

### Gazelle

[Gazelle](https://github.com/bazelbuild/bazel-gazelle) is a tool that generates BUILD files based on
Go and Python source code. `bazel run //:gazelle` will run it. You'll need to run it when you add
new Go source files, packages, or change dependencies.

It addresses this sort of error:

```
compilepkg: missing strict dependencies:
        /tmp/bazel-working-directory/_main/src/server/auth/server/testing/admin_test.go: import of "github.com/pachyderm/pachyderm/v2/src/internal/pctx"
No dependencies were provided.
Check that imports in Go sources match importpath attributes in deps.
```

### Buildifier

If you edit BUILD files, run `bazel test //:buildifier_test` (potentially with `--test_output=all`)
to see if you introduced any lint errors. If you did, run `bazel run //:buildifier` to auto-fix
them. You can also install
[buildifier](https://github.com/bazelbuild/buildtools/blob/master/buildifier/README.md) and have
your editor run it on save for `*.bazel` files. Be aware that the lint rules change based on the
dialect of Starlark (raw Starlark, BUILD.bazel, WORKSPACE, etc.), so your editor needs to tell
`buildifier` the filename it's editing. If not, you will still have lint errors in your saved files.

If you'd like to invoke the version of Go used for proto generation, run `bazel run //:go`.

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

## Hints

`bazel run //target -- ...` prevents `...` from being interpreted as an argument to Bazel, which is
useful if you are passing flags to something you're `bazel run`ning.

`bazel query 'deps("//some:target")'` will list all dependencies of `//some:target`.

`bazel query 'somepath("//some:target", "@@some_library//:whatever")` will show a dependency chain
from `//some:target` to `@@some_library//:whatever`. `allpaths` will show all the chains.

`bazel query --output=build ...` will show a BUILD file representing the matched rules.

### protoc

Gazelle likes to regenerate the protos included with go modules. We have a lot of entries in
`MODULE.bazel` to suppress this behavior:

```starlark
go_deps.gazelle_override(
    directives = [
        "gazelle:proto disable_global",
    ],
    path = "github.com/opentracing/basictracer-go",
)
```

To find more of these to add, do something like:

    bazel query 'somepath("...", "@rules_go//proto:protoc")'

(Note that ... is 3 literal dots, it's not a suggestion to type something else. The first argument
is the "Universe" in which to search for dependencies.)

To find even more, do something like:

    bazel query 'rdeps("...", "@@zlib~1.3//:zutil.h")'

(protoc depends on zlib, which has a file called zutil.h; this shows anything in ... that ends up
depending on that file.)

### realenv

Realenv tests require that `CGO_ENABLED=0`. This means that tests that use realenv will have to set
`pure = "on"` in the `go_test` target. To do that, find the reverse dependencies of realenv:

    $ bazel query 'kind("go_test", rdeps("...", "//src/internal/testpachd/realenv"))'
    //src/internal/collection:collection_test
    //src/internal/fileserver:fileserver_test
    ...

Then use buildozer to adjust those targets:

    $ bazel build //:buildozer
    $ bazel query 'kind("go_test", rdeps("...", "//src/internal/testpachd/realenv"))' | \
      xargs bazel-bin/buildozer 'set pure "on"'
    fixed /home/jrockway/pach/gazelle/src/server/auth/server/testing/BUILD.bazel
    fixed /home/jrockway/pach/gazelle/src/server/pfs/server/testing/BUILD.bazel
    ...

This will not print anything if everything is already fixed, so you can run this freely and know
whether or not it affected anything.

### pulumi

Pulumi eventually depends on github.com/cloudflare/circl, which will require this workaround:
https://github.com/bazelbuild/bazel-gazelle/issues/1421#issuecomment-1424075874

### Profiling

Sometimes you want a whole-system profile of some tests running or something. Build the Go binaries
in debug mode with `-c dbg` to get the symbols, and run the tests outside of the Bazel sandbox with
`--spawn_strategy=local` so that `hotspot` can later find the binaries to inspect.

For example, to see what is making a database-based test slow, pick one at random and run 64 copies
at the same time:

    perf record -F 99 -e cpu-clock -ag -- bazel test //src/server/auth/server/testing:auth_test -c dbg \
    --test_filter=TestListRepoNotLoggedInError --runs_per_test=64 --jobs=64 --local_test_jobs=64

Then look at `perf.out` with `hotspot`:

    hotspot perf.out

This did not help with debugging database speed, but I wanted to document it anyway.

### ssh-ing to CI

Sometimes your tests will fail in CI, but not on your workstation. Traditionally, you would click
"Rerun with SSH" in Circle. You can still do that, but you can also `bazel run //etc/ci-image:run`
to be placed in a shell that is configured identically to CI; CI uses the same Docker image. Your
current working copy of Pachyderm will be mounted in `/home/circleci/project`, just like CI. You can
edit files on your workstation and they are immediately there; it's just a read-only bind mount. You
can run `git` or `bazel` inside this environment as well. While converting to Bazel, I've had a lot
of weird failures on CI, and this has always solved them without wasting time waiting for CI to
restart!
