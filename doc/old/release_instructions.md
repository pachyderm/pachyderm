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

## Releasing:

1) Make sure your commit has a passing build on travis

2) Update `src/client/version/version.go` version values

3) Release

To specify an additional version string:

```shell
make VERSION_ADDITIONAL=RC1 release
```

Otherwise,

```shell
make point-release
```

Afterwards, you'll be prompted to push your changes to master. Please do so.


