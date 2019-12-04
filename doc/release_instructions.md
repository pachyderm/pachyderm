# Release procedure

Types of Releases

|ReleaseType|Example Version|Built off master|Can build off any branch| Updates docs| Can host multiple install versions |
|---|---|---|---|---|---|
|Point Release| v1.7.2| Y | N | Y | N |
|Release Candidate| v1.8.0rc1 | Y | N | Y | N |
|Custom Release | v1.8.1-aeeff234982735987affee | N | Y | N | Y |

## Requirements:

NOTE! At the moment, we require the release script to be run on an ubuntu machine.

This is because of a dependency on CGO via [this bug](https://github.com/opencontainers/runc/issues/841)

(We don't want to enable CGO in part because it doesn't play nice with macOS for us)

You'll need the following credentials / tools:

- A GitHub *Personal Access Token* with **repo** access
  - You can get your personal oauth token here: https://github.com/settings/tokens
- goxc (`go get github.com/laher/goxc`)
- goxc configured ...
    - run: `make GITHUB_OAUTH_TOKEN=<persional access token from #1> goxc-generate-local`
- sha256sum (if you're on mac ... `brew install coreutils`)
- access to `homebrew-tap` and `www` repositories
- S3 credentials
- A dockerhub account, with write access to https://hub.docker.com/u/pachyderm/ (run `docker login`)

If you're doing a custom release (off a branch that isn't master), [skip to the section at the bottom](#custom-release)

## Releasing:

1) Make sure the HEAD commit (that you're about to release) has a passing build on travis.

2) Make sure that you have no uncommitted files in the current branch. Note that `make doc` (next step) will fail if there are any uncommited changes in the current branch

3) Update `src/client/version/client.go` version values, build a new local version of pachctl, and **commit the change** (locally—you'll push it to GitHub in the next step, but this allows `make doc` to run):
    ```
    > make VERSION_ADDITIONAL= install
    > git add src/client/version/client.go
    > git commit -m"Increment version for $(pachctl version --client-only) point release"
    ```

4) Run `make doc` or `make VERSION_ADDITIONAL=<rc/version suffix> doc-custom` with the new version values.

  Note in particular:

  * `VERSION_ADDITIONAL` must be of the form `rc<N>` to comply with [PEP 440](https://www.python.org/dev/peps/pep-0440), which is needed to make ReadTheDocs create a new docs version for the release[\[1\]](http://docs.readthedocs.io/en/latest/versions.html). If you're building a custom release for a client, you don't need follow this form, but if you're building a release candidate, you do.

  * You can also run just `make release-custom` to use the commit hash as the version suffix. Note that this is not PEP 440-compliant, but may be useful for custom releases (see below)

  * Make sure you add any newly created (untracked) doc files, in addition to docs that have been updated (`git commit -a` might not get everything)

5) At this point, all of our auto-generated documentation should be updated. Push a new commit (to master) with:

  ```
  > git add doc
  > git commit -m"Run make doc for $(pachctl version --client-only) (point release|release candidate|etc)"
  > git push origin master
  ```

6) Run `docker login` (as the release script pushes new versions of the pachd and job-shim binaries to dockerhub)

7) Run `make GITHUB_OAUTH_TOKEN=<persional access token from #1> goxc-generate-local` if you have not yet (or the next step will fail)

8) Run `make point-release` or `make VERSION_ADDITIONAL=rc1 release-candidate`

9) Commit the changes (the dash compatibility file will have been newly created), e.g.:

    ```
    > git status
    On branch master
    ....
    Untracked files:
      (use "git add <file>..." to include in what will be committed)

            etc/compatibility/1.6.4

    nothing added to commit but untracked files present (use "git add" to track)
    > git add etc/compatibility/$(pachctl version --client-only) 
    > git commit -m "Update dash compatibility for pachctl $(pachctl version --client-only)"
    > git push origin master
    ```

### If the release failed
You'll need to do two things: remove the relevant tags in GitHub, and re-build the docs in ReadTheDocs

1) Removing the tag in GitHub

  You'll need to delete the *release* and the *release tag* in github. Navigate to
  `https://www.github.com/pachyderm/pachyderm` and click on the *Releases* tab.
  Click on the big, blue version number corresponding to the release you want to
  delete, and you should be redirected to a page with just that release, and red
  "Delete" button on the top right. Click the delete button

  From here, go back to the list of Pachyderm releases, and click "tags". Click
  on the tag for the release you want to delete, and then click "delete" again to
  delete the tag.

  At this point, you can re-run the release process when you're ready.

2) Updating ReadTheDocs

  * Repeat the release process until you're happy with the tagged release in GitHub.
  * Navigate to the [Builds](https://readthedocs.org/projects/pachyderm/builds/) page in ReadTheDocs, select the version corresponding to this release next to the "Build Version" button, and then click "Build Version".
  * Check the updated ReadTheDocs page for the release, and make the docs (particularly the download link under "Local Installation") are correct.

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

## Custom Release

Occasionally we have a need for a custom release off a non master branch. This is usually because some features we need to supply to users that are incompatible w features on master, but the features on master we need to keep longer term.

Often times we can simply cut custom pachd/worker images for a customer. To do that, just run `make custom-images`. Otherwise, if the user needs a custom version of `pachctl`, do the following:

1) Run `docker login` (as the release script pushes new versions of the pachd and job-shim binaries to dockerhub)

2) Run `make custom-release`

Which will create a release like `v1.2.3-2342345aefda9879e87ad`

Which can be installed like:

```
$ curl -o /tmp/pachctl.deb -L https://github.com/pachyderm/pachyderm/releases/download/v1.9.9/pachctl_1.9.9_amd64.deb && sudo dpkg -i /tmp/pachctl.deb
```

Or for mac/brew:

```
# Where 1.7 is the major.minor version of the release you just did,
# and you use the right commit SHA as well in the URL
$ brew install https://raw.githubusercontent.com/pachyderm/homebrew-tap/1.7.0-5a590ad9d8e9a09d4029f0f7379462620cf589ee/pachctl@1.7.rb
```

_After a successful release_, you'll need to manually update the [release](https://github.com/pachyderm/pachyderm/releases) with the tag and publish as a workaround for [this issue](https://github.com/laher/goxc/issues/112).

Then check the docs. Note that ReadTheDocs builds docs from our GitHub master branch. If the docs changes you made aren't checked into the Pachyderm master branch, they won't show up. If you have checked in your docs changes, but they're not showing up as the `latest` version of the docs, tag your version as 'active' on the readthedocs dashboard: https://readthedocs.org/projects/pachyderm/versions/
