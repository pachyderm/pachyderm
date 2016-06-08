# How To Update a Single Dependency

As-is we don't have a mechanism for revendoring everything automatically. That's ok - we want to update dependencies as needed.

## Notes

- govendor does not understand what a 'repo' is -- only a single go package
- so we have two options - update all sub packages, or do them by hand

## Update all sub-packages

`go list -f '{{join .Deps "\n"}}' |  xargs go list -f '{{if not .Standard}}{{.ImportPath}}{{end}} | xargs govendor update`

## Manually

#### Setup:

```shell
# remove the library from your $GOPATH so that you can be confident you're using only vendored copies
rm -rf $GOPATH/src/domain/reponame
# this seems to remove the entries from the vendor.json file:
govendor remove domain/reponame/* 
# while this actually removes the files:
rm -rf vendor/domain/reponame

```

#### Vendoring:

Remember, govendor only understands packages. So you'll have to install each one by hand. Start w the top level package:

```shell
go get domain/reponame
govendor add domain/reponame
```
And then iterate through each sub-package:

```shell
go get domain/reponame/pkg1
govendor add domain/reponame/pkg1
```

I recommend re-running 'go test ...' in between adding each package, until the compilation and tests succeed. Then you should feel confident you have all the packages you need.


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

