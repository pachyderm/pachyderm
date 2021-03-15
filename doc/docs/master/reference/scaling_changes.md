# Pipeline Scaling Limitation in Pachyderm's Community Edition
!!! We are here to help! 
    Scaling Pachyderm pipelines is easy to do but hard for users to perfect. 

It can be difficult to balance the breaking down of data science systems into multiple pipelines,horizontally scaling those pipelines to optimize the processing time and effectively understandingPachyderm's unique data-first approach to development in connecting those pipelines.

Starting with this new 1.13.0 release, we introduce a **limit to the number of pipelines and their parallelization** Community Edition users can deploy.

These limits have been set high enough for most of our users to enjoy Pachyderm capabilities while allowing us to **better assist in designing, scaling, and running pipelines** for those whose workload and pipeline numbers require finer tuning to handle their workload efficiently.

## Scaling Limitation

Our new limitations have been set on:

- The **number of pipelines** deployed: Community Users can deploy **up to 16 pipelines**.
- The **number of workers** for each pipeline: Community Users can run **up to 8 workers in parallel** on each pipeline.

Our Customer Success Experience has shown us that hitting those limits means that you are likely to be needing our assistance in optimizing your DAGs.

We provide an easy way to request a **2 weeks Enterprise Token** that will help you experiment with Pachyderm without those limitations. //TODO insert a link to the Request token page here

Because those limits are enforced at runtime, we will provide explicit messages and direct you to the right ressources.

!!! Info 
    Pachyderm offers readily available activation keys for proofs-of-concept, startups, academic, nonprofit, or open-source projects. Tell us about your project, get in touch. //TODO Add link to contact sales

## What happens when you exceed those limits?
!!! General rule 
    Pass the limit, Get an error message, follow the link.

### Pipelines limit
Well, for once, each `pachctl` command that requires an enterprise key check (For example: `pachctl auth,` `pachctl deploy ide`...) will generate an alert message in your STDERR with a link to the Enterprise landing page. Second, all of the `pachctl deploy` commands will generate an alert message in your STDERR with a link to the Form to request an Enterprise key in the case where no enterprise key has been found.

### Workers limit