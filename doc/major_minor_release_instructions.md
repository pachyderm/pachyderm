# New major or minor releases

In the case of a new major or minor release (x.0.0 or 1.x.0), you will need
to make a couple of additional changes beyond those outlined in the
[release instructions.](./release_instructions.md). Make sure to follow these
steps after the normal release instructions.

## Branch off

Create a new branch off master called `<major>.<minor>.x` and push it to
origin:

```shell
git checkout master
git branch <major>.<minor>.x
git push origin -u <major>.<minor>.x
```

## Update code on master

Now, on the master branch, you'll need to make some changes to prepare the
codebase for the _next_ major or minor release. To see what a complete diff
should look like with these changes, see
[this PR.](https://github.com/pachyderm/pachyderm/pull/5569)

### Update the version

On master, update `src/client/version/client.go` again with the _next_ major
or minor release:

```shell
make VERSION_ADDITIONAL= install
git add src/client/version/client.go
git commit -m"Increment version for $(pachctl version --client-only) release"
```

### Regenerate the golden manifests

```shell
make regenerate-test-deploy-manifests
```

### Write up the extract/restore functionality

Every extract/restore is a bit different, so there's no one-size-fits-all
instructions here. But if you were, e.g. just released 1.12 and are updating
master to be 2.0, you'd do the following:

#### Update protos

Make a copy of the protos:

```shell
mkdir -p src/client/admin/v1_12/pfs
cp src/client/pfs/pfs.proto src/client/admin/v1_12/pfs/pfs.proto
mkdir -p src/client/admin/v1_12/pps
cp src/client/pps/pps.proto src/client/admin/v1_12/pps/pps.proto
mkdir -p src/client/admin/v1_12/auth
cp src/client/auth/auth.proto src/client/admin/v1_12/auth/auth.proto
mkdir -p src/client/admin/v1_12/enterprise
cp src/client/enterprise/enterprise.proto src/client/admin/v1_12/enterprise/enterprise.proto
```

Then, in the copied over protos:

* Change the package names to be suffixed with the version, e.g. in
  `pfs.proto` change `package pfs`; to `package pfs_1_12;`. Also update the
  `go_package` field.
* Update the imports to reference the copied protos. e.g. in `pps.proto`,
  change `import client/pfs/pfs.proto` to
  `import client/admin/v1_12/pfs/pfs.proto`, and any uses of `pfs.` to
  `pfs_1_12.`.

In `src/client/admin/admin.proto`:

* Import the copied PFS and PPS protos
* Rename `Op1_12` (which should already exist) to `Op2_0`.
* Add a new `Op1_12`, which is basically a copy of `Op2_0` referencing the
  copied protos:

    ```protobuf
    message Op1_12 {
      pfs_1_12.PutObjectRequest object = 2;
      pfs_1_12.CreateObjectRequest create_object = 9;
      pfs_1_12.TagObjectRequest tag = 3;
      pfs_1_12.PutBlockRequest block = 10;
      pfs_1_12.CreateRepoRequest repo = 4;
      pfs_1_12.BuildCommitRequest commit = 5;
      pfs_1_12.CreateBranchRequest branch = 6;
      pps_1_12.CreatePipelineRequest pipeline = 7;
      pps_1_12.CreateJobRequest job = 8;
      auth_1_12.SetACLRequest set_acl = 11;
      auth_1_12.ModifyClusterRoleBindingRequest set_cluster_role_binding = 12;
      auth_1_12.SetConfigurationRequest set_auth_config = 13;
      auth_1_12.ActivateRequest activate_auth = 14;
      auth_1_12.RestoreAuthTokenRequest restore_auth_token = 15;
      enterprise_1_12.ActivateRequest activate_enterprise = 16;
      CheckAuthToken check_auth_token = 17;
    }
    ```

* Add `Op2_0 op2_0 = 7` in `message Op { ... }`. Assign the next sequential number to op2_0.

Finally, run `make proto` to rebuild the protos.

#### Wire in the proto changes

Copy over the converters:

```shell
cp src/server/admin/server/convert1_11.go src/server/admin/server/convert1_12.go
```

Then update `convert1_11.go`:

* Update imports for pfs, pps, auth, enterprise to point to 1.12 version. Eg. `import client/pfs` to `import client/admin/v1_12/pfs`
* Replace all instances of `pfs.` with `pfs1_12.`
* Replace all instances of `pps.` with `pps1_12.`
* Replace all instances of `auth.` with `auth1_12.`
* Replace all instances of `enterprise.` with `enterprise1_12.`

Then update `convert1_12.go`:

* Replace all instances of `1_12` with `2_0`
* Replace all instances of `1_11` with `1_12`
* Make any appropriate changes necessary for the conversion process (sometimes
  just the above replacements are sufficient)

Update the admin client (`src/client/admin.go`) and admin server
(`src/server/admin/server/api_server.go`). Mostly this just involves looking
for instances of `1_11` or `1_12` in those files and updating the code and
add `applyOp2_0` function which should be a copy of `applyOp1_12`

Update `convert1_7.go`, around line 309, to include `convert1_12Objects`.

Finally `make install docker-build` to ensure everything compiles, then commit
the updates and push.
