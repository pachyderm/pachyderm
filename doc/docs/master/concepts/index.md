# Concepts

Pachyderm is an enterprise-grade, open source data science platform that
makes explainable, repeatable, and scalable Machine Learning (ML) and
Artificial Intelligence (AI) a reality. The Pachyderm platform brings
together version control for data with the tools to build scalable
end-to-end ML/AI pipelines while empowering users to develop their
code in any language, framework, or tool of their choice. Pachyderm
has been proven to be the ideal foundation for teams looking to
use ML and AI to solve real-world problems in a reliable way.

The Pachyderm platform includes the following main components:

- Pachyderm File System (PFS)
- Pachyderm pipelines

To start, you need to understand the foundational concepts of Pachyderm's
data versioning and pipeline semantics. After you have a good grasp of
the basics, you can use advanced concepts and features for more
complicated challenges.

This section describes the following Pachyderm concepts:

<div class="row">
  <div class="column-2">
    <div class="card-square mdl-card mdl-shadow--2dp">
      <div class="mdl-card__title mdl-card--expand">
        <h4 class="mdl-card__title-text">Versioned Data Concepts &nbsp;&nbsp; &nbsp;<i class="fa fa-rocket"></i></h4>
      </div>
      <div class="mdl-card__supporting-text">
        Learn about the main Pachyderm abstractions that
        you will operate with when using Pachyderm.
      </div>
      <div class="mdl-card__actions mdl-card--border">
          <ul>
            <li><a href="data-concepts/" class="md-typeset md-link">
            Versioned Data Concepts Overview
            </a>
            </li>
          </ul>
      </div>
    </div>
  </div>
  <div class="column-2">
    <div class="card-square mdl-card mdl-shadow--2dp">
      <div class="mdl-card__title mdl-card--expand">
        <h4 class="mdl-card__title-text">Pipeline Concepts &nbsp;&nbsp;&nbsp;<i class="fa fa-cogs"></i></h4>
      </div>
      <div class="mdl-card__supporting-text">
        Learn the main concepts of the Pachyderm
        pipeline system.
      </div>
      <div class="mdl-card__actions mdl-card--border">
          <ul>
            <li><a href="pipeline-concepts/" class="md-typeset md-link">
            Pipeline Concepts Overview
           </a>
          </li>
       </div>
     </div>
  </div>
</div>
<div class="row">
  <div class="column-2">
    <div class="card-square mdl-card mdl-shadow--2dp">
      <div class="mdl-card__title mdl-card--expand">
        <h4 class="mdl-card__title-text">Advanced Concepts &nbsp;&nbsp; &nbsp;<i class="fa fa-rocket"></i></h4>
      </div>
      <div class="mdl-card__supporting-text">
        More about Pachyderm abstractions: Global IDs, 
        deferred processing, and distributed computing.
      </div>
      <div class="mdl-card__actions mdl-card--border">
          <ul>
            <li><a href="../concepts/advanced-concepts/globalID/" class="md-typeset md-link">
            Advanced Concepts
            </a>
            </li>
          </ul>
      </div>
    </div>
  </div>
  <div class="column-2">
  </div>
</div>





# Pachyderm

[Pachyderm](https://www.pachyderm.com/) provides the data layer that allows machine learning teams to productionize and scale their machine learning lifecycle. With Pachyderm’s industry-leading data versioning, pipelines, and lineage, teams gain data-driven automation, petabyte scalability, and end-to-end reproducibility. Teams using Pachyderm get their ML projects to market faster, lower data processing and storage costs, and meet regulatory compliance requirements more easily.

Read about how our Customers leveraged Pachyderm to face their ML/AI challenges and operationalize their pipelines: [Case studies](https://www.pachyderm.com/case-studies/)

## Features include

- **Data-driven pipelines**
    - Our containerized (language-agnostic) pipelines execute whenever new data is committed
    - Pachyderm is "Kubernetes native" and autoscale with parallel processing of data without writing additional code
    - Incremental processing saves compute by only processing differences and automatically skipping duplicate data

- **Automated data versioning**
    Pachyderm’s Data Versioning gives teams an automated and performant way to keep track of all data changes.
    - Our file-based, git-like versioning core feature provides a complete audit trail for all data and artifacts across pipeline stages, including intermediate results
    - Our optimized storage framework supports petabytes of unstructured data while minimizing storage costs

- **Immutable data lineage**
    We provide an immutable record for all activities and assets in the ML lifecycle:
    - We track every version of your code, models, and data and manage relationships between historical data states so you can maintain the reproducibility of data and code for compliance
    - Our Global IDs feature makes it easy for teams to track any result all the way back to its raw input, including all analysis, parameters, code, and intermediate results

- Our **Console** provides an intuitive visualization of your DAG (directed acyclic graph)

- Our **Notebooks** provide an easy way to interact with Pachyderm data versioning and pipelines via Jupyter notebooks bridging the worlds of data science and data engineering

## Installing Pachyderm Operator

### Prerequisites
Before installing Pachyderm Operator, you will need to install `pachctl`, the command-line tool that you will use to interact with a Pachyderm cluster in your terminal.
Follow the instructions in our [documentation](https://docs.pachyderm.com/2.0.x-rc/getting_started/local_installation/#install-pachctl) .

### Install The Operator
Pachyderm Operator has a **Red Hat marketplace listing**.

- Subscribe to the operator on Marketplace
  https://marketplace.redhat.com/en-us/products/pachyderm
- Install the operator and validate
  https://marketplace.redhat.com/en-us/documentation/operators

### Post Deployment
Finally, connect your client to your cluster.
In your terminal:

1. Create a new Pachyderm context with the embedded Kubernetes context by running:

    ```
    pachctl config import-kube <name-your-new-pachyderm-context> -k `kubectl config current-context`
    ```

1. Verify that the context was successfully created:

    ```
    pachctl config get context <name-your-new-pachyderm-context>
    ```

1. Activate the new Pachyderm context:

    ```
    pachctl config set active-context <name-your-new-pachyderm-context>
    ```

1. Verify that the new context has been activated:

    ```
    pachctl config get active-context
    ```

1. You should be all set to start using Pachyderm

    * Run the following command to make sure that your cluster is responsive:
        ```
        pachctl version
        ```
        **System Response**
        ```
        COMPONENT           VERSION
        pachctl             2.0.0
        pachd               2.0.0
        ```
    * Then run your first pipelines by following the following [tutorial](https://docs.pachyderm.com/latest/getting_started/beginner_tutorial/) or check our Quick Start in the `Enabled` Menu of your OpenShift Data Science Console.

    