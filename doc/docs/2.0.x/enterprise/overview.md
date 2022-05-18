# Features Overview

!!! Note
     To get more information about Pachyderm Enterprise Edition, to ask questions, or to get access for evaluation, don't hesitate to get in touch with us at [sales@pachyderm.io](mailto:sales@pachyderm.io) or on our [Slack](https://www.pachyderm.com/slack/){target=_blank}. 

## Enterprise Features List

Pachyderm Enterprise Edition helps you scale and manage Pachyderm data pipelines in an enterprise setting.

It delivers the most recent version of the Community Edition of Pachyderm along with additional features, a UI (Console) for visualizing pipelines and exploring data, and a customized JupyterHub deployment (Notebooks) to experiment with your data and pipelines from a Jupyter Notebook.

!!! Warning "THE ENTERPRISE EDITION LIFTS **ALL SCALING LIMITATIONS**"
     Note that the activation of the Enterprise Edition [**lifts all scaling limits of the Community Edition**](../../reference/scaling-limits/). You can run as many pipelines as you need and parallelize your jobs without constraints.


### Additional Features

Pachyderm Enterprise unlocks a series of additional administrative and security features needed for enterprise-scale deployments of Pachyderm, namely:

- [**Authentication**](../auth/authentication/idp-dex): Pachyderm allows for authentication **against any OIDC provider**. Users can authenticate to Pachyderm by logging into their favorite Identity Provider. 
- [**Role-Based Access Control - RBAC**](../auth/authorization/): Enterprise-scale deployments require access control.  Pachyderm Enterprise Edition gives teams the ability to control access to production pipelines and data.  Administrators can silo data, prevent unintended modifications to production pipelines, and support multiple data scientists or even multiple data science groups by controlling users' access to Pachyderm resources.
- [**Enterprise Server**](../auth/enterprise-server/setup/): An organization can have **many Pachyderm clusters registered with one single Enterprise Server** that manages the Enterprise licensing and the integration with a company's Identity Provider.

### Tooling

Pachyderm Enterprise also comes with a complementary tool that will quickly become indispensable when designing and debugging pipelines: [**Pachyderm Console**](#console) - A visual interface for pipeline visualization and data exploration.

This first iteration of Pachyderm Console provides an intuitive visualization of your DAGs, an easy way to drill into pipelines and job details, explore data, and read through logs.  
    
## Console
Pachyderm Enterprise Edition includes a full UI for visualizing pipelines and exploring data.  It automatically infers the structure of data scientists' DAGs and displays them visually. Data scientists and cluster admins can click on individual segments of pipelines and repos to see how many jobs have run, explore commits and data, or access Pachyderm logs. Console is an indispensable tool when designing and troubleshooting your data workflow.

![Console Pipeline](../images/console-pipeline.png)

You can [deploy Console with Pachyderm](../../deploy-manage/deploy/console/) by adding the relevant fields to your Helm values. A production environment requires setting up an [Ingress Controller, and a DNS](../../deploy-manage/deploy/ingress/). You can also choose to deploy Console [locally](../../deploy-manage/deploy/console/#deploy-locally) to experiment with the product.






