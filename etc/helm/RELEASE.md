# Creating a new release

Merge your commits to master

When ready to made a new release, create a git commit incrementing `version` in `pachyderm/Chart.yaml` and commit that to master.

Now create a tag in the form `pachyderm-x.x.x` for that commit and push it to the repo. Circle should pick up on the tag and automatically publish a release and update the chart repo.

<!-- SPDX-FileCopyrightText: Pachyderm, Inc. <info@pachyderm.com>
SPDX-License-Identifier: Apache-2.0 -->