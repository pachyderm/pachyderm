# Release procedure

NOTE! At the moment, we require the release script to be run on an ubuntu machine.

This is because of a dependency on CGO via [this bug](https://github.com/opencontainers/runc/issues/841)

(We don't want to enable CGO in part because it doesn't play nice w OSX for us)

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

2) Update `src/client/version/version.go` version values

3) Run `make doc` with the new version values.

4) At this point, all of our auto-generated documentation should be updated. Push a new commit (to master) with:

```
> VERSION=<new version, e.g. 1.2.3>
> git commit -a -m"Update version and run make doc for ${VERSION} point release"
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


