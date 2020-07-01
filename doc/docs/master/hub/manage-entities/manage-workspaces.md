# Manage Workspaces

In Pachyderm Hub, you run your data science pipelines in workspaces.
A workspace is a collection of infrastructure resources that run
Pachyderm components needed to process your data. Workspace resource
limitations are defined by your account type and provisioned automatically.
For example, in the free tier, you can only create one workspace, while
a Pro account allows you to create multiple workspaces.

Each workspace has isolated computing, processing, storage, and other
resources. Resource autoscaling is managed by Pachyderm behind the
scenes. You cannot move data or pipelines between workspaces.
Each workspace in an organization must have a unique name.

## Create a Workspace

You need a workspace to star building your pipelines and working
with Pachyderm. 

To create a workspace, complete the following steps:

1. Click **Create New Workspace** or **Create a 4-Hour Workspace**.
1. Fill out the required fields and accept Terms and Conditions.
1. Click **Create**.

   Your workspace should be created instantly.

## Connect to Your Workspace

You can access your workspace either by using the command-line interface (CLI)
called `pachctl` or by using the PachDash UI. Although some of the
basic tasks, such as pipeline and repo creation, file browsing, and others,
can be done through the dashboard. More complex tasks require you to use
`pachctl`.

To connect to a workspace, complete the following steps:

* To connect by using the dashboard:

  1. On the **Workspaces** page, click **PachDash** next to your workspace.

     The PachDash UI is opening.

* To connect using `pachctl`:

  1. On the **Workspaces** page, click the **Connect** button next to your
  workspace.
  1. Copy the `echo` script.
  1. Log in to your terminal and paste the copied script.
  1. Copy other commands one by one and run them in your terminal.
  1. Verify that you are connected to your workspace:

     ```bash
     pachctl config get active-context
     ```

     This command must return the name of your workspace.

  You can now create your Pachyderm repositories and pipelines for
  your workflow. If this is your first time using Pachyderm, 
  try our [Beginner Tutorial](../../../getting_started/beginner_tutorial/).

## Delete a Workspace

When you delete a workspace, the data in your workspace is erased. If this is
not your intention, back up your cluster as described in [Back up and Restore](../../../deploy-manage/manage/backup_restore/).
You can only delete a workspace in the Pachyderm Hub UI. The `pachctl undeploy`
command is not supported for workspaces.

To delete a workspace, complete the following steps:

1. Go to **Workspaces**.
1. Click **Delete** next to the workspace name.
