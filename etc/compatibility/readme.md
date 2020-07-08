## What

This directory is a compatibility table between dash and pachctl releases.

Each file, e.g. `1.6.0` contains a list of dash image versions that it is
compatible with.

Do not modify these files manually.

## How it works

As part of the dash CI, the latest pachctl version is tested against the latest
dash release.

If the tests pass, that dash image version is concatenated to the file
corresponding to the pachctl version it was tested against.

## Why

This allows us to retrieve the lastest dash image a user can use with their
version of pachctl, enabling:

```
# Easy upgrades
pachctl update-dash
# Implicit upgrades each time deploy is used
pachctl deploy ...

```
