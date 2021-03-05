# Getting Started with Pachyderm Hub

Pachyderm Hub is a **SaaS platform** that 
gives you access to all of **Pachyderm's functionalities
without the burden of deploying and maintaining it** locally
or in a third-party cloud platform. 

This section walks you through
the steps of creating a cluster in Pachyderm Hub so that
you do not need to worry about the underlying infrastructure
and can get started right away.

!!! Note
    Let us know how to improve Pachyderm Hub. Join our [Slack channel](http://slack.pachyderm.io).
## Get started in 4 simple steps
![Pachyderm Hub Steps](../images/pachhub_steps.png)
## Before you start
Log in with your GitHub or Google account to start using [hub.pachyderm.com](https://hub.pachyderm.com). 
![Pachyderm Hub Login](../images/pachhub_login.png)
## 1- Create a Cluster 
Click the **Create a 4-hr Workspace** button and fill out the form.
![PachHub cluster](../images/pachhub_create_workspace.png)

Your just provisioned a one-node cluster that you can now use for
a limited time for free!

!!! Note
      While Pachyderm maintains a few clusters that are instantly
      available, none may be available during periods of high traffic. If
      you see your cluster is in a *starting* state, you might have to wait a few
      minutes for it to be ready.

## 2- Install pachctl
Pachyderm Hub enables you to access your cluster through Pachyderm 
CLI `pachctl` and the web interface called the Dashboard.
Although you can perform most simple actions directly in the Dashboard,
`pachctl` provides full functionality. Most likely, you will use
`pachctl` for any operation beyond the most basic workflow.
We recommend that you use `pachctl` for all data operations and
the Dashboard to view your data and graphical representation of your
pipelines.

After your cluster creation, open a terminal window and [install 'pachctl'](https://docs.pachyderm.com/latest/getting_started/local_installation/#install-pachctl).

!!! Warning
    `kubectl` commands are not supported for the clusters deployed
    on Pachyderm Hub.
## 3-4 Configure your Pachyderm context and login to your cluster by using a one-time authentication token
1. To configure a Pachyderm context and log in to your cluster
(i.e. have your`pachctl` point to your new cluster), click the **Connect** link on your cluster name in the Pachyderm Hub UI.

      ![Pachyderm Cluster running](../images/pachhub_cluster_running.png)

      In your terminal window, copy, paste, and run the commands 1,2,3 listed in the instructions.

      ![Pachyderm Cluster connect](../images/pachhub_cluster_connect.png)

      **Note**
            Your token has a limited lifetime. The following error means that your authentication
            token has expired. Close the pop-up window and re-connect.
            ```shell
               error authenticating with Pachyderm cluster: otp is invalid or has expired
            ```

1. Verify that you have set the correct context:

      ```shell
      pachctl config get active-context
      ```
      **System response**
      ```
      jolly-caribou
      ```

1. Create a test repo:

      ```shell
      pachctl create repo test
      ```
      ```shell
      pachctl list repo
      ```
      **System response**
      ```
      NAME CREATED       SIZE (MASTER) ACCESS LEVEL
      test 5 seconds ago 0B            OWNER    
      ```

1. Check the repo in your dashboard:

      In the Pachyderm Hub UI, click the **Details** button next to your cluster name,
      then **Dash**. Your dashboard opens in a new window.

      ![repo_ready](../images/pachhub_dash_ready.png)

## Next Step

You have successfully deployed and configured a Pachyderm
cluster in Pachyderm Hub.

Next: [Beginner Tutorial](../getting_started/beginner_tutorial.md)
