# Provenance

Data versioning enables Pachyderm users to go back in time and see the state
of a dataset or repository at a particular moment in time. Data provenance
(from the French *provenir* which means *the place of origin*),
also known as data lineage, tracks the dependencies and relationships
between datasets. Provenance answers not only the question of
where the data comes from, but also how the data was transformed along
the way. Data scientists use provenance in root cause analysis to improve
their code, workflows, and understanding of the data and its implications
on final results. Data scientists need
to have confidence in the information with which they operate. They need
to be able to reproduce the results and sometimes go through the whole
data transformation process from scratch multiple times, which makes data
provenance one of the most critical aspects of data analysis. If your
computations result in unexpected numbers, the first place to look
is the historical data that gives insights into possible flaws in the
transformation chain or the data itself.

For example, when a bank makes a decision about a mortgage
application, many factors are taken into consideration, including the
credit history, annual income, and loan size. This data goes through multiple
automated steps of analysis with numerous dependencies and decisions made
along the way. If the final decision does not satisfy the applicant,
the historical data is the first place to look for proof of authenticity,
as well as for possible prejudice or model bias against the applicant.
Data provenance creates a complete audit trail that enables data scientists
to track the data from its origin to the final decision and make
appropriate changes that address issues. With the adoption of General Data
Protection Regulation (GDPR) compliance requirements, monitoring data lineage
is becoming a necessity for many organizations that work with sensitive data.

Pachyderm implements provenance for both commits and repositories.
You can track revisions of the data and
understand the connection between the data stored in one repository
and the results in the other repository.

Collaboration takes data provenance even further. Provenance enables teams
of data scientists across the globe to build on each other work, share,
transform, and update datasets while automatically maintaining a
complete audit trail so that all results are reproducible.

The following diagram demonstrates how provenance works:

![Provenance example](../../assets/images/provenance.svg)

In the diagram above, you can see two input repositories called `parameters`
and `training-data`. The `training-data` repository continuously collects
data from an outside source. The training model pipeline combines the
data from these two repositories, trains many models, and runs tests to
select the best one.

Provenance helps you to understand how and why the best model was
selected and enables you to track the origin of the best model.
In the diagram above, the best model is represented with a purple
circle. By using provenance, you can find that the best model was
created from the commit **2** in the `training-data` repository
and the commit **1** in the `parameters` repository.

## Tracking Provenance in Pachyderm

Pachyderm provides the `pachctl inspect` command that enables you to track
provenance of your commits and learn where the data in the repository
originates in.

!!! example
    ```shell
    $ pachctl inspect commit split@master
    Commit: split@f71e42704b734598a89c02026c8f7d13
    Original Branch: master
    Started: 4 minutes ago
    Finished: 3 minutes ago
    Size: 0B
    Provenance:  __spec__@8c6440f52a2d4aa3980163e25557b4a1 (split)  raw_data@ccf82debb4b94ca3bfe165aca8d517c3 (master)
    ```

In the example above, you can see that the latest commit in the master
branch of the split repository tracks back to the master branch in the
`raw_data` repository.

## Tracking Provenance Downstream

Pachyderm provides the `flush commit` command that enables you
to track provenance downstream. Tracking downstream means that instead of
tracking the origin of a commit, you can learn in which output repository
a certain input has resulted.

For example, you have the `ccf82debb4b94ca3bfe165aca8d517c3` commit in
the `raw_data` repository. If you run the `pachctl flush commit` command
for this commit, you can see in which repositories and commits that data
resulted.

!!! example
    ```shell
    $ pachctl flush commit raw_data@ccf82debb4b94ca3bfe165aca8d517c3
    REPO        BRANCH COMMIT                           PARENT STARTED        DURATION       SIZE
    split       master f71e42704b734598a89c02026c8f7d13 <none> 52 minutes ago About a minute 0B
    split       stats  9b46d7abf9a74bf7bf66c77f2a0da4b1 <none> 52 minutes ago About a minute 15.39MiB
    pre_process master a99ab362dc944b108fb33544b2b24a8c <none> 48 minutes ago About a minute 0B
    ```
