# Notes on vendoring

As-is we don't have a mechanism for revendoring everything automatically.

That's ok - we want to update dependencies as needed. But to do so is tricky.

Govendor has some bugs, and its unclear how go test / build can be resolving imports.

Generally, I've found the following approach helpful. But first, some disclaimers:

## Things I still don't understand about govendor

I don't understand how the basics work - the update/add/remove

The update/add commands never seem to do what I want. Adding a library (even after removing) sometimes doesn't update all the files? It's also unclear if its getching from GOPATH or an origin url, or if there is a fallback between them.

e.g. I'm updating bazil.org/fuse.

I tried 'govendor remove bazil.org/fuse' but it leaves files behind! I guess because it can't differentiate between when I want to remove a top level package and a subpackage? That's kind of crazy.

So, to actually remove / update a whole repo, you need to do the following:

```shell
govendor remove domain/reponame/* # this seems to remove the entries from the vendor.json file?
rm -rf vendor/domain/reponame
govendor add domain/reponame
govendor add domain/reponame/pkg1
govendor add domain/reponame/pkg2
```

Unfortunately, there doesn't seem to be a way to get govendor to add a whole repository in a single pass. Even trying `govendor add +missing` doesn't work. 

(Maybe we should just write a script to clone a repo? And update the vendor.json accordingly.)

It's also very easy to miss a package when vendoring if you still have a copy of hte library on your GOPATH. So I recommend removing (or moving) the copy on GOPATH while you do the vendor/go test/loop to be confident you have all the packages you need.

Except ... it does seem to want a copy in GOPATH?

e.g. 

    $ govendor add bazil.org/fuse/fs/fstestutil
    Error: Package "bazil.org/fuse/fs/fstestutil" not a go package or not in GOPATH.

So then you do `go get bazil.org/fuse/fs/fstestutil` and repeat the govendor add and it works.

## To update a package

### Step 0 - Update

- Try using `govendor update github.com/some/package` to update what you need. You'll probably have to run this command for all subpackages you need.
- If that's not working, you can try the removal/fetch steps below

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
