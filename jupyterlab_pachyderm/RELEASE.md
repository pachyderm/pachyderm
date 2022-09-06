# Making a new release of jupyterlab-pachyderm

The extension can be published to `PyPI` via CI.

## Automated Release using CircleCI

### Major and Minor Releases

Major and minor version releases are published off of the `main` branch.

> Note for pre-releases, the steps are almost the same, except you skip steps 3-5, where you are supposed to bump versions to prepare for the next release.

1. Validate release info
    - Ensure that the `"version"` key in `package.json` contains the version to be released.
    - Also ensure that `CHANGELOG.md` is up-to-date.
    - Update `"version"` and `CHANGELOG.md` in a new PR if necessary.
    - Jot down the `version` and `commit_sha` that you want to release.
    - Make sure the version in `Dockerfile` is up to date
1. Create tag and push to GitHub.
   ```
   git tag -a v<version> <commit_sha> -m "Release version <version>"
   git push origin tag v<version>
   ```
   > This will trigger CI for building and publishing the new package to TestPyPI. You will need to go to CircleCI to approve publishing to PyPI.
1. Create a new branch for patch releases with the scheme `vA.B.x`. For example, if releasing `0.1.0`, create a new branch called `v0.1.x` to hold future patches.
1. In the new patch release branch, bump up the `"version"` in `package.json` to contain the next *patch* version. For example `0.1.0 -> 0.1.1`.
1. Create a PR to bump up the version in the `main` branch to contain the next *minor* version.

### Patch Releases

Sometimes we need to patch minor versions to fix bugs. Typically, we make a PR to fix the bug and merge into the patch release branch. Once that is merged, we are ready to publish a new patch release.

1. Checkout the relevant `vA.B.x` patch release branch.
1. Validate release info:
    - `"version"` in `package.json`
    - `CHANGELOG.md`
    - Update `"version"` and `CHANGELOG.md` in a PR if necessary
1. Jot down the `patch_version` and `commit_sha`.
1. Create tag and push to GitHub.
    ```
    git tag -a v<patch_version> <commit_sha> -m "Release version <patch_version>"
    git push origin tag v<patch_version>
    ```
    > This will trigger CI for building and publishing the new package to TestPyPI. You will need to go to CircleCI to approve publishing to PyPI.

## GitHub Releases

[Release instructions primer](https://docs.github.com/en/repositories/releasing-projects-on-github/managing-releases-in-a-repository)

1. Go to [tags on GitHub](https://github.com/pachyderm/jupyterlab-pachyderm/tags), and choose the tag we just created
1. Click on the tag > then click `Create release from tag` button
1. Name it after version number without the `v` prefix
1. Autogenerate the change doc by comparing the commits between this tag and the previous release.
1. If this is a pre-release, then make sure to check that box as well.

## [todo] Publishing to `conda-forge`

If the package is not on conda forge yet, check the documentation to learn how to add it: https://conda-forge.org/docs/maintainer/adding_pkgs.html

Otherwise a bot should pick up the new version publish to PyPI, and open a new PR on the feedstock repository automatically.
