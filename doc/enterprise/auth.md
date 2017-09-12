# Access Controls

The Pachyderm Enterprise Edition access control features will let you create and manage various users interacting with your Pachyderm cluster.  You can restrict access to individual data repositories on a per user basis and, as a result, restrict the subscription of pipelines to those data respositories.  

These docs will guide you through:

1. [Activating access control features (aka "auth" features).](#activating-access-control)
2. [Logging into Pachyderm as a user.](#logging-into-pachyderm)
3. [Updating user access to data repositories.](#updating-user-access)

We will also discuss:

- [The behavior of pipelines when using access control](#the-behavior-of-pipelines-when-using-ACLs)
- [The behavior of a cluster when access control is de-activated or an enterprise token expires](#behavior-when-ACLs-are-de-activated-or-activation-codes-expire)

## Activating access control

First, you will need to make sure that your cluster has Pachyderm Enterprise Edition activated (you can follow [this guide](deployment.md) to activate the Enterprise Edition).  The status of the Enterprise features can be verified with `pachctl`:

```
$ pachctl enterprise get-state
ACTIVE
```

Next, we need to activate the Enterprise access control features. This can be done with `pachctl auth activate ...`.  However, before executing that command, we should decide on at least one user that will have admin priviledges on the cluster. admins on the Pachyderm cluster will be able to modify the access of any non-admin users on the cluster.

All users in Pachyderm are identified by their GitHub user names.  Thus, if we wanted to activate access control on a cluster and set the user `dwhitena` as an admin, we would execute the following command:

```

```

## Logging into Pachyderm

## Updating user access

## The behavior of pipelines when using ACLs

## Behavior when ACLs are de-activated or activation codes expire
