Deploying Pachyderm
===================

## Setup

First time user? [Follow the setup instructions](./deploying_setup.html)

## Inputting data into PFS

To start using Pachyderm, you'll need to input some of your data into a [PFS](./pachyderm_file_system.html) repo.

There are a handful of ways to do this:

1) You can do this via a [PFS](./pachyderm_file_system.html) mount

The [Fruit Stand](https://github.com/pachyderm/pachyderm/tree/master/examples/fruit_stand) example uses this method. This is probably the easiest to use if you're poking around and make some dummy repo's to get the hang of Pachyderm.

2) You can do it via the [pachctl](./pachctl.html) CLI

This is probably the most reasonable if you're scripting the input process.

3) You can use the golang client

You can find the docs for the client [here](https://godoc.org/github.com/pachyderm/pachyderm/src/client)

Again, helpful if you're scripting something, and helpful if you're already using go.

4) You can dump SQL into pachyderm

You'll probably use the golang client to do this. Example coming soon.

## Generating your custom image


## Tools


## Going to Production

When you're ready to deploy your pipelines in production, we recommend making sure you have the following setup:

1) [the Kubernetes Dashboard](https://github.com/kubernetes/dashboard)
2) A real object store

By default, a local k8s cluster / pachyderm cluster will use the local filesystem for storage. This is fine for development, but you should make sure you have S3/GCS setup to go into production. [Refer to the setup instructions](./deploying_setup.html) for more info.


3) The code behind your images under source control


