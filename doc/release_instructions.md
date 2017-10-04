# Release procedure

NOTE! At the moment, we require the release script to be run on an ubuntu machine.

This is because of a dependency on CGO via [this bug](https://github.com/opencontainers/runc/issues/841)

(We don't want to enable CGO in part because it doesn't play nice w OSX for us)

If you're doing a custom release (off a branch that isn't master), [skip to the section at the bottom](#custom-release)

## Requirements:

You'll need the following credentials / tools:

- A GitHub *Personal Access Token* with **repo** access
  - You can get your personal oauth token here: https://github.com/settings/tokens
- goxc (`go get github.com/laher/goxc`)
- goxc configured ...
    - run: `make GITHUB_OAUTH_TOKEN=<persional access token from #1> goxc-generate-local`
- sha256sum (if you're on mac ... `brew install coreutils`)
- access to `homebrew-tap` and `www` repositories
- S3 credentials
- A dockerhub account, with write access to https://hub.docker.com/u/pachyderm/

## Releasing:

1) Make sure the HEAD commit (that you're about to release) has a passing build on travis.

2) Make sure that you have no uncommitted files in the current branch. Note that `make doc` (next step) will fail if there are any uncommited changes in the current branch

3) Update `src/client/version/client.go` version values, and **commit the change** (locally—you'll push it to GitHub in the next step, but this allows `make doc` to run)

4) Run `make doc` or `make VERSION_ADDITIONAL=<rc/version suffix> doc-custom` with the new version values.

  Note in particular:

  * `VERSION_ADDITIONAL` must be of the form `rc<N>` to comply with [PEP 440](https://www.python.org/dev/peps/pep-0440), which is needed to make ReadTheDocs create a new docs version for the release[\[1\]](http://docs.readthedocs.io/en/latest/versions.html). If you're building a custom release for a client, you don't need follow this form, but if you're building a release candidate, you do.

  * You can also run just `make release-custom` to use the commit hash as the version suffix. Note that this is not PEP 440-compliant, but may be useful for custom releases (see below)

  * Make sure you add any newly created (untracked) doc files, in addition to docs that have been updated (`git commit -a` might not get everything)

5) At this point, all of our auto-generated documentation should be updated. Push a new commit (to master) with:

  ```
  > git add <all docs files>
  > git commit -m"(Update version and) run make doc for <version> (point release|release candidate|etc)" (e.g. "Update version and run make doc for 1.0.0 point release" or "Run make doc for 1.0.0rc1 release candidate")
  > git push origin master
  ```

6) Run `docker login` (as the release script pushes new versions of the pachd and job-shim binaries to dockerhub)

7) Run `make point-release` or `make VERSION_ADDITIONAL=<rc/version suffix> release-custom`


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

## Custom Release

Occasionally we have a need for a custom release off a non master branch. This is usually because some features we need to supply to users that are incompatible w features on master, but the features on master we need to keep longer term.

Follow the instructions above, just run the make script off of your branch.

Then _after a successful release_:

- The tag created by goxc will point to master, and this is wrong. Opened an issue for this: https://github.com/laher/goxc/issues/112
- But the binaries built are correct (they're built off of your local code, on the branch you've checked out)
- So we'll have to manually create the tag + release + binaries

1) Download the binaries that goxc created

Do this for the release you just created.

[Here's an example URL](https://github.com/pachyderm/pachyderm/releases/tag/v1.2.5)

You should download all three binaries:

- pachctl_1.2.5_amd64.deb
- pachctl_1.2.5_darwin_amd64.zip
- pachctl_1.2.5_linux_amd64.tar.gz

2) Delete the release

- You'll have to go to http://github.com/pachyderm/pachyderm/releases, and manually delete the release that goxc created
- Also make sure you delete the tag. [You can see a list of tags here](https://github.com/pachyderm/pachyderm/tags) or here's an [example release tag](https://github.com/pachyderm/pachyderm/releases/tag/v1.2.5)

3) Manually tag your branch

```
git tag -d v1.2.6 # You may need to delete it locally
git tag v1.2.6
git push origin --tags
```

This will fail if you didn't delete the tag on Github in the previous step

4) Manually create the release on Github

- [Create a new release](https://github.com/pachyderm/pachyderm/releases/new)
- Use the tag you just pushed to Github
- Provide a custom message. [Here's an example](https://github.com/pachyderm/pachyderm/releases/tag/v1.2.6)
- Upload the binaries (the link is at the bottom) that you copied
- Publish the release

5) Check the docs

Note that ReadTheDocs builds docs from our GitHub master branch. If the docs changes you made aren't checked into the Pachyderm master branch, they won't show up.

If you have checked in your docs changes, but they're not showing up as the `latest` version of the docs, tag your version as 'active' on the readthedocs dashboard: https://readthedocs.org/projects/pachyderm/versions/
