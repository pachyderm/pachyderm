# Overview

Pachyderm Enterprise Edition can be deployed easily on top of an existing or new deployment of Pachyderm, and we have engineers available to help enterprise customers get up and running very quickly.  To get more information about Pachyderm Enterprise Edition, to ask questions, or to get access for evaluation, please contact us at [sales@pachyderm.io](mailto:sales@pachyderm.io) or on our [Slack](http://slack.pachyderm.io/). 

## Enterprise Features List

Pachyderm Enterprise Edition let's you scale and manage Pachyderm data pipelines in an enterprise setting. 
It delivers the most recent version of Pachyderm along with: 

### Additional Features

- **No scaling limitations**: The activation of the Enterprise Edition **lifts all [scaling limits of the Community Edition**](../../reference/scaling_limits/). You can run as many pipelines as you need and parallelize your jobs without constraints.
- Administrative and security features needed for enterprise-scale implementations of Pachyderm :
    - [**Authentication**](../auth/authentication/): Pachyderm allows for authentication against any OIDC provider. Users can authenticate to Pachyderm by logging into their favorite Identity Provider. 
    - [**Authorization**](../auth/authorization/): Enterprise-scale deployments require access controls.  Pachyderm Enterprise Edition gives teams the ability to control access to production pipelines and data.  Administrators can silo data, prevent unintended modifications to production pipelines, and support multiple data scientists or even multiple data science groups by controling the access of users Pachyderm resources.
    - [**Enterprise Server**](../auth/enterprise-server/): An organization can have **many Pachyderm clusters registered with one single Enterprise Server** that manages Enterprise Licensing and the integration with a company's Identity Provider.

### Tools

Pachyderm Enterprise comes with two complementary tools that will quickly become indispensable when designing and debugging pipelines.

- Pipeline visualization and data exploration with [**Pachyderm Console**](#console): a visual interfaces to Pachyderm:

    This first iteration of Pachyderm Console provides an intuitive visualization of your DAG, an easy way to drill into pipelines and job details, explore data and read through logs.  
    
- Run experiments and explore your data with [**Notebooks**](#notebooks): a customization of JupiterLab UI running on your Pachyderm cluster (beta version). 


## Console

Pachyderm Enterprise Edition includes a full UI for visualizing pipelines and exploring data.  Pachyderm Enterprise will automatically infer the structure of data scientists' DAG pipelines and display them visually.  Data scientists and cluster admins can click on individual segments of pipelines and repos to see how many jobs have run, explore data or access Pachyderm logs. 

![Console Pipeline](../images/console-pipeline.png)

!!! Note
    Console requires a separate [installation]()

## Notebooks
Pachyderm Enterprise Edition includes a JupiterLab UI allowing you to run your pipelines and data experiments from your favorite Jupiter notebooks.

![Notebooks](../images/notebooks.png)

!!! Note
    Notebooks require a seaparate [installation]()







