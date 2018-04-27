# Contributing guidelines

## Filing issues

If you have a question or have a problem using Pachyderm, you can [file an issue](https://github.com/pachyderm/pachyderm/issues), join our [users channel on Slack](http://slack.pachyderm.io), or email us at [support@pachyderm.io](mailto:support@pachyderm.io) and we can help you right away.

General questions can usually be answered on our [Documentation Portal](http://pachyderm.readthedocs.io) or [FAQ](http://pachyderm.readthedocs.io/en/latest/FAQ.html).

## Submitting Pull Requests

We'd love to accept your Pull Requests! There are a few boxes you need to check before we can merge your PR though:

- We need you to sign our [Contributor License Agreement
(CLA)](https://pachyderm.wufoo.com/forms/pachyderm-contributor-license-agreement/).
Without this we can't legally distribute the code, you only have to do it once.


- Your code must pass the golinter, this can be checked by running `make lint`
it will spit out lint errors it finds, if it outputs nothing you're good to go.

- Your code must pass CI. CI will always fail initially for user submitted PRs
because part of our CI process contains sensitive credentials we can't run by
default for outside PRs. Pachyderm devs will help you with this one.

