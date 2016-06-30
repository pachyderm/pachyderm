## Tl;Dr

You probably want to add a missing package / update a package. Do:

```
cd src/server
govendor list +missing
```

To see what you're missing, then add it via:

```
govendor add yourpkg
```

Or add all missing packages by:

```
govendor add +missing
```

To update existing packages, read on ...

## How Govendor Works

1) The `vendor.json` is the source of truth

Don't add things there manually.

The only package that isn't in there is he symlink to the client library.

2) To check the state of the world:

Run `govendor list`

It'll list all the packages w certain flags ... which they call 'status'

(More on that [here](https://github.com/kardianos/govendor))

Things to note in particular are packages marked 'm' for missing

- these are dependencies that the code relies on but doesn't have
- so you'll want to add them

Also note any dependencies that are marked 'e' for external

- these are dependencies that are external to our repo

3) To add / update dependencies

- `govendor add somepkg` adds it to the repo
- `govendor update somepkg` is local
  - it uses whatever version of the pkg is available on $GOPATH/src
  - it DOES NOT ever use the network to get a package
- `govendor fetch somepkg` is non local
  - it uses the network to fetch a new package version
  - I believe it bypasses any copies in $GOPATH/src

4) Blow it all away

This will remove all dependencies and you'll start from scratch. It's a sledgehammer but may be better than fighting any conflicts for too long.

In my experience, if I'm tracking down more than a few conflicting errors, I'll blow away and restart.

Note -- you'll have to replace the client symlink at `src/server/vendor/github.com/pachyderm/pachyderm/src/client`

```
rm -rf src/server/vendor
cd src/server/vendor
govendor init
govendor add +external
```

Then replace the client symlink and debug / get things compiling and passing again.

5) Other good things to know

- govendor does not understand what a 'repo' is -- only a single go package
- so we have two options - update all sub packages, or do them by hand

## Troubleshooting

### Internal Packages

If you see a message like this:

```shell
CGOENABLED=0 GO15VENDOREXPERIMENT=1 go test -cover -v -short $(go list ./src/server/... | grep -v '/src/server/vendor/')
package github.com/pachyderm/pachyderm/src/server/cmd/job-shim
        imports github.com/pachyderm/pachyderm/src/client
        imports github.com/pachyderm/pachyderm/src/client/pfs
        imports github.com/gengo/grpc-gateway/runtime
        imports google.golang.org/grpc
        imports google.golang.org/grpc/internal: use of internal package not allowed
```

What's happening is that while google.golang.org.grpc is vendored (but double check by looking under `src/server/vendor/google.golang.org/`), its looking for
`google.golang.org/grpc/internal` and if it can't find it under the vendored directory, it will default to looking under $GOPATH, which fails the internal requirement.


To fix this, just do:

    govendor add google.golang.org/grpc/internal


### Travis Breaks, But Locally Tests Work

It probably means you didn't blow away the library in $GOPATH and locally were using a copy that wasn't vendored. Repeat the steps above under the 'Manually' heading.

