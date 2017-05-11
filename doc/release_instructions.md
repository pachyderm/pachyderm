# Release procedure

NOTE! At the moment, we require the release script to be run on an ubuntu machine.

This is because of a dependency on CGO via [this bug](https://github.com/opencontainers/runc/issues/841)

(We don't want to enable CGO in part because it doesn't play nice w OSX for us)

If you're doing a custom release (off a branch that isn't master), [skip to the section at the bottom](#custom-release) 

## Requirements:

You'll need the following credentials / tools:

- goxc (`go get github.com/laher/goxc`)
- goxc configured ...
    - run: `make GITHUB_OAUTH_TOKEN=12345 goxc-generate-local`
    - You can get your personal oauth token here: https://github.com/settings/tokens
- sha256sum (if you're on mac ... `brew install coreutils`)
- access to `homebrew-tap` and `www` repositories
- S3 credentials
- A dockerhub account, with write access to https://hub.docker.com/u/pachyderm/

## Releasing:

1) Make sure your commit has a passing build on travis

2) Update `src/client/version/client.go` version values, commit the change

3) Run `make doc` with the new version values. If you're doing an RC or need to specify an additional version string, run it like `make VERSION_ADDITIONAL=RC1 doc`

Make sure you add any new doc files, e.g:

```
$ git status
...
Untracked files:
  (use "git add <file>..." to include in what will be committed)

	doc/pachctl/pachctl_rerun-pipeline.md
$ git add doc/pachctl/pachctl_rerun-pipeline.md
```

4) At this point, all of our auto-generated documentation should be updated. Push a new commit (to master) with:

```
> git commit -a -m"Update version and run make doc for VERSION point release"
> git push origin master
```

5) Run `docker login` (as the release script pushes new versions of the pachd and job-shim binaries to dockerhub)

6) Run `make release` or `make point-release`

To specify an additional version string:

```shell
make VERSION_ADDITIONAL=RC1 release
```

Otherwise,

```shell
make point-release
```

Afterwards, you'll be prompted to push your changes to master. Please do so.

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

The docs version may not show up. If this is the case, tag your version as 'active' on the readthedocs dashboard: https://readthedocs.org/projects/pachyderm/versions/

