# Manage Workspaces

In Pachyderm Hub, all data and pipelines are isolated and run in workspaces.
A workspace is a collection of infrastructure resources that run
Pachyderm components needed to process your data. Workspace resource
limitations are defined by your account type and provisioned automatically.
For example, in the free tier, you can only create one workspacei at a time,
while a Pro account allows you to create multiple workspaces.

Each workspace has isolated computing, processing, storage, and other
resources that are autoscaled and managed by Pachyderm behind the
scenes.
Each workspace in an organization must have a unique name and you cannot
move data or pipelines between workspaces excep `extract` and `restore`.

## Create a Workspace

You need a workspace to start adding data and building pipelines
with Pachyderm. 

To create a workspace, complete the following steps:

1. Click **Create New Workspace** or **Create a 4-Hour Workspace**.
1. Fill out the required fields and accept Terms and Conditions.
1. Click **Create**.

   Your workspace should be created instantly.

   **Note:** While Pachyderm maintains a few clusters that are instantly
   available, none may be available during periods of high traffic. If
   you see your cluster is in a starting state, you might have to wait
   a few minutes for it to be ready.

## Connect to Your Workspace

Pachyderm Hub enables you to access your workspace through a command-line
interface (CLI) called `pachctl` and the web interface called the Dashboard.
Although you can perform most simple actions directly in the dashboard,
`pachctl` provides full functionality. Most likely, you will use
`pachctl` for any operation beyond the most basic workflow.
Pachyderm recommends that you use `pachctl` for all data operations and
the dashboard to view your data and graphical representation of your
pipelines.

To connect to a workspace, complete the following steps:

* To connect using `pachctl`:

  1. On your local computer, open a terminal window.
  1. Install or upgrade  `pachctl` as described in
     [Install pachctl](../../getting_started/local_installation/#install-pachctl).
 
  1. Verify your `pachctl` version:

     ```bash
     pachctl version --client-only
     ```

     **System Response:**

     ```bash
     1.11.x
     ```

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

  1. Verify that you can communicate with your workspace by using
  `pachctl` commands in your terminal:

  1. Create a repo called `test`:

     ```bash
     pachctl create repo test
     ```

  1. Verify that the repo was created:

     ```bash
     pachctl list repo
     ```

     **System Response:**

     ```bash
     NAME   CREATED       SIZE (MASTER) ACCESS LEVEL
     test   3 seconds ago 0B            OWNER
     ```

* To connect by using the dashboard:

  1. On the **Workspaces** page, click **PachDash** next to your workspace.

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
