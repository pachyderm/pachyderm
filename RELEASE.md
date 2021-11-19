# Making a new release of jupyterlab_pachyderm

The extension can be published to `PyPI` and `npm` manually.

## Manual release

Assuming we want to release a new version `v1.2.3`:

- From `main` branch, run `git checkout -b v1.2.3-release`
- Bump up `"version"` in `package.json`
  - run `npm install` to update `package-lock.json`
- Update `CHANGELOG.md`
  - What to include? You can try to summarize the changes by comparing the previous release to the latest main, something like `https://github.com/pachyderm/jupyterlab-pachyderm/compare/v1.2.2...main`
  - You can also add `[Changes](https://github.com/pachyderm/jupyterlab-pachyderm/compare/v1.2.2...v1.2.3)` to the changelog
- `git commit -m "v1.2.3 Release"`, make PR, then **Squash and Merge** into `main`
- back in `main`, run `git tag -a v1.2.3 -m "Release version 1.2.3" && git push origin v1.2.3`

### Python package

This extension can be distributed as Python
packages. All of the Python
packaging instructions in the `pyproject.toml` file to wrap your extension in a
Python package. Before generating a package, we first need to install `build`.

```bash
pip install build twine
```

To create a Python source package (``.tar.gz``) and the binary package (`.whl`) in the `dist/` directory, do:

```bash
python -m build
```

> `python setup.py sdist bdist_wheel` is deprecated and will not work for this package.

Then to upload the package to PyPI, do:

```bash
twine upload dist/*
```

Or try uploading to TestPyPI, first:

```bash
twine upload --repository testpypi dist/*
```

### NPM package

To publish the frontend part of the extension as a NPM package, do:

```bash
npm login
npm publish --access public
```
### GitHub Releases

Follow [instructions](https://docs.github.com/en/repositories/releasing-projects-on-github/managing-releases-in-a-repository)

- Choose the tag we just created
- The Python package should be at `dist/*.tar.gz`, whish is produced by the `python -m build` command
  - Upload the tar file via the GitHub Web GUI ([Step 8])[https://docs.github.com/en/repositories/releasing-projects-on-github/managing-releases-in-a-repository#:~:text=generate%20release%20notes.-,Optionally,-%2C%20to%20include%20binary]
- If we produce a npm package, upload that as well

## Publishing to `conda-forge`

If the package is not on conda forge yet, check the documentation to learn how to add it: https://conda-forge.org/docs/maintainer/adding_pkgs.html

Otherwise a bot should pick up the new version publish to PyPI, and open a new PR on the feedstock repository automatically.
