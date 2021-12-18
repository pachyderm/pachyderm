# Making a new release of jupyterlab-pachyderm

The extension can be published to `PyPI` via CI.

## Automated Release using CircleCI

Below is an example release for `v1.2.3`:

1. From `main` branch, run `git checkout -b release-v1.2.3`
1. Changes to do in this branch:
    1. Bump up `"version"` in `package.json`
    1. Run `npm install` to update `package-lock.json`
    1. Update `CHANGELOG.md` by summarizing the changes. You can see the changes via `https://github.com/pachyderm/jupyterlab-pachyderm/compare/v1.2.2...main`
1. `git commit -m "Release v1.2.3"`, push, and make a PR

Once the PR is approved
1. On the release branch `release-v1.2.3`, run
    ```
    git tag -a v1.2.3 -m "Release version 1.2.3"
    git push origin tag v1.2.3
    ```
    > This will kick off the automated release process to TestPyPI. You will need to go to [CircleCI](https://app.circleci.com/pipelines/github/pachyderm/jupyterlab-pachyderm) to approve pushing to PyPI.

1. Merge the PR once the release is created and assets are published to PyPI.

### GitHub Releases

[Release instructions primer](https://docs.github.com/en/repositories/releasing-projects-on-github/managing-releases-in-a-repository)

1. Go to [tags on GitHub](https://github.com/pachyderm/jupyterlab-pachyderm/tags), and choose the tag we just created
1. Click on the tag > then click `Create release from tag` button
1. Name it after version number without the `v` prefix
1. Autogenerate the change doc by comparing the commits between this tag and the previous release.
1. If this is a pre-release, then make sure to check that box as well.

## [todo] Publishing to `conda-forge`

If the package is not on conda forge yet, check the documentation to learn how to add it: https://conda-forge.org/docs/maintainer/adding_pkgs.html

Otherwise a bot should pick up the new version publish to PyPI, and open a new PR on the feedstock repository automatically.
