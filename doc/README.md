# Docs

These docs are rendered and searchable on our [Developer Documentation Portal](http://pachyderm.readthedocs.io/en/stable). Here are a few section links for quick access.

- [Getting started with Pachyderm](http://pachyderm.readthedocs.io/en/stable/getting_started/getting_started.html) including installation and a beginner tutorial
- [Analyzing your own data](http://pachyderm.readthedocs.io/en/stable/deployment/analyze_your_data.html) and creating custom pipelines
- [Advanced features](http://pachyderm.readthedocs.io/en/stable/advanced/advanced.html) of Pachyderm such as provenance and using diffs of data for processing. 
- [Pachctl API Documentation](http://pachyderm.readthedocs.io/en/stable/pachctl/pachctl.html)
- [FAQ](http://pachyderm.readthedocs.io/en/stable/FAQ.html)

Note on updating:

We keep documentation on multiple versions of pachyderm under `docs/<version>`.
Any docs, new or historical, can be updated by editing the files under the
appropriate version's directory. We also maintain a "latest" version of the
docs, which is created when `build.sh` generates a symlink from `site/latest`
to another version under `site`. Currently, the latest version is 1.9.x.
