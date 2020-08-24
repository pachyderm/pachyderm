# New major or minor releases

In the case of a new major or minor release (x.0.0 or 1.x.0), you will need
to make a couple of additional changes beyond those outlined in the
[release instructions.](./release_instructions.md). Make sure to follow these
steps after the normal release instructions.

## Branch off

Create a new branch off master called `<major>.<minor>.x` and push it to
origin:

```bash
git checkout master
git branch <major>.<minor>.x
git push origin -u <major>.<minor>.x
```

## Update code on master

Now, on the master branch, you'll need to make some changes to prepare the
codebase for the _next_ major or minor release. To see what a complete diff
should look like with these changes, see
[this PR.](https://github.com/pachyderm/pachyderm/pull/5111)

### Update the version

On master, update `src/client/version/client.go` again with the _next_ major
or minor release:

```bash
make VERSION_ADDITIONAL= install
git add src/client/version/client.go
git commit -m"Increment version for $(pachctl version --client-only) release"
```

### Regenerate the golden manifests

```bash
make regenerate-test-deploy-manifests
```

### Write up the extract/restore functionality

Every extract/restore is a bit different, so there's no one-size-fits-all
instructions here. But if you were, e.g. just released 1.11 and are updating
master to be 1.12, you'd do the following:

#### Update protos

Make a copy of the protos:

```bash
mkdir -p src/client/admin/v1_11/pfs
cp src/client/pfs/pfs.proto src/client/admin/v1_11/pfs/pfs.proto
mkdir -p src/client/admin/v1_11/pps
cp src/client/pps/pps.proto src/client/admin/v1_11/pps/pps.proto
mkdir -p src/client/admin/v1_11/auth
cp src/client/auth/auth.proto src/client/admin/v1_11/auth/auth.proto
```

Then, in the copied over protos:

* Change the package names to be suffixed with the version, e.g. in
  `pfs.proto` change `package pfs`; to `package pfs_1_11;`. Also update the
  `go_package` field.
* Update the imports to reference the copied protos. e.g. in `pps.proto`,
  change `import client/pfs/pfs.proto` to
  `import client/admin/v1_11/pfs/pfs.proto`, and any uses of `pfs.` to
  `pfs_1_11.`.

In `src/client/admin/admin.proto`:

* Import the copied PFS and PPS protos
* Rename `Op1_11` (which should already exist) to `Op1_12`.
* Add a new `Op1_11`, which is basically a copy of `Op1_12` referencing the
  copied protos:

    ```protobuf
    message Op1_11 {
      pfs_1_11.PutObjectRequest object = 2;
      pfs_1_11.CreateObjectRequest create_object = 9;
      pfs_1_11.TagObjectRequest tag = 3;
      pfs_1_11.PutBlockRequest block = 10;
      pfs_1_11.CreateRepoRequest repo = 4;
      pfs_1_11.BuildCommitRequest commit = 5;
      pfs_1_11.CreateBranchRequest branch = 6;
      pps_1_11.CreatePipelineRequest pipeline = 7;
      pps_1_11.CreateJobRequest job = 8;
    }
    ```

Finally, run `make proto` to rebuild the protos.

#### Wire in the proto changes

Copy over the converters:

```bash
cp src/server/admin/server/convert1_10.go src/server/admin/server/convert1_11.go
```

Then update `convert1_11.go`:

* Replace all instances of `1_11` with `1_12`
* Replace all instances of `1_10` with `1_11`
* Make any appropriate changes necessary for the conversion process (sometimes
  just the above replacements are sufficient)

Update the admin client (`src/client/admin.go`) and admin server
(`src/server/admin/server/api_server.go`). Mostly this just involves looking
for instances of `1_10` or `1_11` in those files and updating the code.

Update `convert1_7.go`, around line 309, to include `convert1_11Objects`.

Finally `make install docker-build` to ensure everything compiles, then commit
the updates and push.
