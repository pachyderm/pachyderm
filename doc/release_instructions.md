# Release procedure

## Types of releases

|ReleaseType|Example Version|Built off master|Can build off any branch| Updates docs| Can host multiple install versions |
|---|---|---|---|---|---|
|Point Release| v1.7.2| Y | N | Y | N |
|Release Candidate| v1.8.0rc1 | Y | N | Y | N |
|Custom Release | v1.8.1-aeeff234982735987affee | N | Y | N | Y |

## Requirements

NOTE! At the moment, we require the release script to be run on an ubuntu
machine.

This is because of a dependency on CGO via
[this bug](https://github.com/opencontainers/runc/issues/841), as we don't
want to enable CGO in part because it doesn't play nice with macOS for us.

You'll need the following credentials / tools:

- A GitHub *Personal Access Token* with **repo** access
  - You can get your personal oauth token here: https://github.com/settings/tokens
- `goxc` (`go get github.com/laher/goxc`)
- `goxc` configured ...
    - run: `make GITHUB_OAUTH_TOKEN=<persional access token from #1> goxc-generate-local`
- `sha256sum`
- access to `homebrew-tap` and `www` repositories
- S3 credentials
- A dockerhub account, with write access to https://hub.docker.com/u/pachyderm/ (run `docker login`)
- `silversearcher`
    - on linux: `apt-get install -y silversearcher-ag`
    - on mac: `brew install the_silver_searcher`

If you're doing a custom release (off a branch that isn't master), [skip to the section at the bottom](#custom-release)

## Releasing

The following is the procedure for point releases, rcs, or anything else off
of master; i.e. for non-custom releases. For custom release instructions, see
below.

### Prerequisites

1) Make sure the HEAD commit (that you're about to release) has a passing build on travis.
2) Make sure that you have no uncommitted files in the current branch.

### Update client version

Update `src/client/version/client.go` version values.

- for a major release, increment the MajorVersion and set the MinorVersion and MicroVersion to 0; e.g. `2.0.0`.
- for a minor release, leave the MajorVersion unchanged, increment the MinorVersion, and set the MicroVersion to 0; e.g. `1.10.0`.
- for a patch release, leave the MajorVersion and MinorVersion unchanged and increment the MicroVersion; e.g. `1.9.8`.

Commit these changes locally (you will push to GitHub in a later step):

```bash
make VERSION_ADDITIONAL= install
git add src/client/version/client.go
git commit -m"Increment version for $(pachctl version --client-only) release"
```

### Update dash compatibility version

Commit these changes locally (you will push to GitHub in a later step):

```bash
make dash-compatibility
git add etc/compatibility
git commit -m"Update dash compatibility for $(pachctl version --client-only) release"
```

Note: The update to "latest" will cause dash CI to default run with the
release pointed to be latest The latest link is only update for
Major/Minor/Point releases. In order to test a new version of dash with
RC/Alpha/Beta/Custom release, modify the deployment manifest to test it
manually.

### Rebuild docs

Run `make doc`. Make sure you add any newly created (untracked) doc files, in addition to docs that have been updated (`git commit -a` might not get everything):

```bash
git add doc
git commit -m"Run make doc for $(pachctl version --client-only)"
```

### Regenerate golden deployment manifests

```bash
make regenerate-test-deploy-manifests
git commit -a -m"Regenerate golden deployment manifests for $(pachctl version --client-only)"
```

### Update the changelog

Update the changelog in the branch and commit it locally.

### Push changes

```bash
git push
```

### Release!

* To release a point version, run `make point-release`
* To release an alpha version, run e.g. `make VERSION_ADDITIONAL=-alpha1 release-candidate`
* To release a beta version, run e.g. `make VERSION_ADDITIONAL=-beta1 release-candidate`
* To release an rc, run e.g. `make VERSION_ADDITIONAL=-rc1 release-candidate`

Then update the [release's notes](https://github.com/pachyderm/pachyderm/releases), and post the update on the #users channel.

### New major or minor releases

In the case of a new major or minor release (x.0.0 or 1.x.0), you will need
to make a couple of additional changes:

#### Write up the extract/restore functionality

- Copy the protobuf from the prior release into `src/client/admin`.
- Update `src/client/admin/admin.proto` to include the operations for the prior release.
- Run `make proto` to rebuild the protos.
- Add a converter to `src/server/admin/server`, e.g. `convert_1_11.go`.
- Update the admin client (`src/client/admin.go`) and admin server (`src/server/admin/server/api_server.go`.)

Look to the extract/restore functionality for other versions as a basis to
build off of. Frequently, it's just a matter of copy/pasting that code and
updating some names.

Then push the changes:

```bash
git commit -am "Added placeholder files for extract/restore functionality for <next version>"
git push
```

#### Branch off

Create a new branch off master called `<major>.<minor>.x` and push it to origin:

```bash
git checkout master
git branch <major>.<minor>.x
git push origin -u <major>.<minor>.x
```

## Custom releases

Occasionally we have a need for a custom release off a non-master branch. This
is usually because some features we need to supply to users that are
incompatible with features on master, but the features on master we need to
keep longer-term.

Assuming the prerequisites are met, making a custom release should simply be a
matter of running `make custom-release`. This will create a release like
`v1.2.3-2342345aefda9879e87ad`, which can be installed like:

```bash
curl -o /tmp/pachctl.deb -L https://github.com/pachyderm/pachyderm/releases/download/v1.11.0/pachctl_1.11.0_amd64.deb && sudo dpkg -i /tmp/pachctl.deb
```

Or for mac/brew:

```bash
# Where 1.7 is the major.minor version of the release you just did,
# and you use the right commit SHA as well in the URL
brew install https://raw.githubusercontent.com/pachyderm/homebrew-tap/1.7.0-5a590ad9d8e9a09d4029f0f7379462620cf589ee/pachctl@1.7.rb
```

## If the release failed

You'll need to delete the *release* and the *release tag* in github. Navigate to
`https://www.github.com/pachyderm/pachyderm` and click on the *Releases* tab.
Click on the big, blue version number corresponding to the release you want to
delete, and you should be redirected to a page with just that release, and red
"Delete" button on the top right. Click the delete button

From here, go back to the list of Pachyderm releases, and click "tags". Click
on the tag for the release you want to delete, and then click "delete" again to
delete the tag.

At this point, you can re-run the release process when you're ready.

## Rolling back a release

If a release has a problem and needs to be withdrawn, the steps in rolling back a release are similar to the steps under "If the release failed". In general, you'll need to:
- Delete the tag and GitHub Release for both the bad release *and the most recent good release*
- Re-release the previous version (to update homebrew)

All of these can be accomplished by:
- Following the steps under "If the release failed" for deleting the tag and GitHub release for both the bad release
- Checking out the git commit associated with the most recent good release (`git checkout tags/v<good release>`). Save this commit SHA (`git rev-list tags/v<good> --max-count=1`), in case you need it later, as we'll be deleting the tag.
- Delete the tag and GitHub release for the last good release (the one you just checked out)
- Syncing your local Git tags with the set of tags on Github (either re-clone the Pachyderm repo, or run `git tag -l | xargs git tag -d; git fetch origin master --tags`). This prevents the release process from failing with `tag already exists`.
- Run `make point-release` (or follow the release process for custom releases)
