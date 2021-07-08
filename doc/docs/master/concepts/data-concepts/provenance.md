# Provenance

Data versioning enables Pachyderm users to go back in time and see the state
of a dataset or repository at a particular moment in time. 

Data provenance (from the French *provenir* which means *the place of origin*),
also known as data lineage, tracks the dependencies and relationships
between datasets. Provenance answers not only the question of
where the data comes from, but also how the data was transformed along
the way. 

Data scientists use provenance in root cause analysis to improve
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
as well as for possible prejudice or model bias.
**Data provenance creates a complete audit trail** that enables data scientists
to track the data from its origin to the final decision and make
appropriate changes that address issues. With the adoption of General Data
Protection Regulation (GDPR) compliance requirements, monitoring data lineage
is becoming a necessity for many organizations that work with sensitive data.

**Pachyderm implements provenance for both commits and repositories**.
You can track revisions of the data and
understand the connection between the data stored in one repository
and the results in the other repository.

Collaboration takes data provenance even further. Provenance enables teams
of data scientists across the globe to build on each other work, share,
transform, and update datasets while automatically maintaining a
complete audit trail so that all results are reproducible.

The following diagram demonstrates how provenance works:

![Provenance example](../../assets/images/provenance.svg)

In the diagram above, you can see two input repositories called `params`
and `data`. The `data` repository continuously collects
data from an outside source. The training model pipeline combines the
data from these two repositories, trains many models, and runs tests to
select the best one.

Provenance helps you to understand how and why the best model was
selected and enables you to track the origin of the best model.
In the diagram above, the best model is represented with a purple
circle. By using provenance, you can find that the best model was
created from the commit **1a** in the `data` repository
and the commit **2a** in the `params` repository.

## Tracking Direct Provenance in Pachyderm

Pachyderm provides the `pachctl inspect` command that enables you to track
the direct provenance of your commits and learn where the data in the repository
originates in.

!!! example
    ```shell
    pachctl inspect commit edges@71c791f3252c492a8f8ad9a51e5a5cd5 --raw
    ```

    **System Response:**

    ```shell
    {
        "commit": {
            "branch": {
            "repo": {
                "name": "edges",
                "type": "user"
            },
            "name": "master"
            },
            "id": "71c791f3252c492a8f8ad9a51e5a5cd5"
        },
        "origin": {
            "kind": "AUTO"
        },
        "parent_commit": {
            "branch": {
            "repo": {
                "name": "edges",
                "type": "user"
            },
            "name": "master"
            },
            "id": "b6fc0ab2d8d04972b0c31b0e35133323"
        },
        "started": "2021-07-07T19:53:17.242981574Z",
        "finished": "2021-07-07T19:53:19.598729672Z",
        "direct_provenance": [
            {
            "repo": {
                "name": "edges",
                "type": "spec"
            },
            "name": "master"
            },
            {
            "repo": {
                "name": "images",
                "type": "user"
            },
            "name": "master"
            }
        ],
        "details": {
            "size_bytes": "22754"
        }
    }
    ```

In the example above, you can see that the commit `71c791f3252c492a8f8ad9a51e5a5cd5` 
on the master branch of the `edges` repo was **automatically** produced (`origin`) from a **user** input on
the master branch of the `images` repo processed by the `edges` pipeline (`direct_provenance`).

Additionally, the parent of the commit `71c791f3252c492a8f8ad9a51e5a5cd5`  in the `edges` repo has the id `b6fc0ab2d8d04972b0c31b0e35133323`.

## Traversing Provenance and Subvenance

In Pachyderm, all the related steps in a DAG share the same identifier,
making it easy to traverse the provenance and subvenance of any commit.

All it takes is to run `pachctl inspect commitset <commitID>`
to get the full list of all the commits created along with it due to provenance relationships.


Visit the [Global ID Page](../../advanced-concepts/globalID/) for more details.

