# Release procedure


## Pachctl

[Follow the release instructions for pachctl](https://github.com/pachyderm/homebrew-tap/blob/master/README.md)

## Pachd

- [find the tag/release on github you want to release](https://github.com/pachyderm/pachyderm/releases)
  - CI needs to pass
  - and it needs to be merged into master to generate a tag
- make sure on GH its marked as a release
- checkout the commit locally, then run:

    git fetch --tags
    make release-pachd

To test / verify:


