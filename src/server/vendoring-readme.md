# Notes on vendoring

As-is we don't have a mechanism for revendoring everything automatically.

That's ok - we want to update dependencies as needed. But to do so is tricky.

Govendor has some bugs, and its unclear how go test / build can be resolving imports.

Generally, I've found the following approach helpful:


## To update a package

### Step 0 - Update

Try using `govendor update github.com/some/package` to update what you need. You'll probably have to run this command for all subpackages you need.

To run it for sub commands you can do:

`go list -f '{{join .Deps "\n"}}' |  xargs go list -f '{{if not .Standard}}{{.ImportPath}}{{end}} | xargs govendor update`

If that's not working, you can try ...

### Step 1 - Removal

- do 'govendor remove github.com/example/package'
- if that doesn't panic (it seems to always for me) continue on
- if it does panic, do `rm -rf vendor/github.com/example/package`, and
- remove the relevant fields from `vendor/vendor.json`

### Step 2 - Fetching / Adding

- do a `go get` on the package/repo in question
- then do `govendor add ...`


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


### General Debugging

I've found it useful to move or blow away packages under $GOPATH/src so that you can be confident you're loading the libraries you think you are. It's disappointing the information printed isn't a bit more helpful, but by doing this you'll know. And if this changes your error, then you'll know that you weren't using a vendored copy (and you probably should be).


### Sledgehammer

If the package you're importing has a ton of sub packages / its complicated to do manually one by one via `govendor add ...`, then what I've done is:

- do `govendor add` for the top level package
- this way the manifest will contain the checksum / etc for the repo
- then do the removal step above
- and then something like `cp -r $GOPATH/src/github.com/big/package vendor/github.com/big/package`
