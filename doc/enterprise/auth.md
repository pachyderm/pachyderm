# Access Controls

The Pachyderm Enterprise Edition access control features will let you create and manage various users interacting with your Pachyderm cluster.  You can restrict access to individual data repositories on a per user basis and, as a result, restrict the subscription of pipelines to those data respositories.  

These docs will guide you through:

1. [Activating access control features (aka "auth" features).](#activating-access-control)
2. [Logging into Pachyderm as a user.](#logging-into-pachyderm)
3. [Managing/updating user access to data repositories.](#managing-and-updating-user-access)

We will also discuss:

- [The behavior of pipelines when using access control](#the-behavior-of-pipelines-when-using-ACLs)
- [The behavior of a cluster when access control is de-activated or an enterprise token expires](#behavior-when-activation-codes-expire)

## Activating access control

First, you will need to make sure that your cluster has Pachyderm Enterprise Edition activated (you can follow [this guide](deployment.md) to activate the Enterprise Edition).  The status of the Enterprise features can be verified with `pachctl`:

```
$ pachctl enterprise get-state
ACTIVE
```

Next, we need to activate the Enterprise access control features. This can be done with `pachctl auth activate ...`.  However, before executing that command, we should decide on at least one user that will have admin priviledges on the cluster. admins on the Pachyderm cluster will be able to modify the access of any non-admin users on the cluster.

All users in Pachyderm are identified by their GitHub user names.  Thus, if we wanted to activate access control on a cluster and set the GitHub user `dwhitena` as an admin, we would execute the following command:

```
$ pachctl auth activate --admins=dwhitena
```

Your Pachyderm cluster can have more than one admin if you like.  You would just need to specify them here as a comma separated list. 

## Logging into Pachyderm

Now that we have activated access control via `pachctl auth`, we can login to our cluster.  Let's login as the GitHub user `dwhitena` using `pachctl auth login <username>`.  When we execute this command, `pachctl` will provide us with a GitHub link to authenticate ourselves as the provided GitHub user, as shown below:

```
$ pachctl auth login dwhitena
(1) Please paste this link into a browser:

https://github.com/login/oauth/authorize?client_id=d3481e92b4f09ea74ff8&redirect_uri=https%3A%2F%2Fpachyderm.io%2Flogin-hook%2Fdisplay-token.html

(You will be directed to GitHub and asked to authorize Pachyderm's login app on Github. If you accept, you will be given a token to paste here, which will give you an externally verified account in this Pachyderm cluster)

(2) Please paste the token you receive from GitHub here:

```

When visiting this link in a browser, you will be presented with an aption to "Authorize Pachyderm" (assuming that you haven't authorized Pachyderm via GitHub previously). Once you authorize Pachyderm, you will be presented with a Pachyderm user token:

![alt tag](auth.png)

Copy and paste this token back into the terminal as requested by `pachctl` and press enter.  You are now logged in to Pachyderm!  

## Managing and updating user access

Now, let's suppose that we create a repository call `myrepo` with the user `dwhitena`.  Because, the user `dwhitena` created this repository, `dwhitena` will have full read/write access to the repo.  This can be confirmed via the `pachctl auth check ...` command:

```
$ pachctl auth check owner myrepo
true
$ pachctl auth check reader myrepo
true
$ pachctl auth check writer myrepo
true
```

Or via the `pachctl auth get ...` command:

```
$ pachctl auth get dwhitena myrepo
OWNER
```

An OWNER of `myrepo` or cluster admin, can then set other user's scope of access to the repo.  This can be done via the `pachctl auth set ...` command.  For example, to give the GitHub user `JoeyZwicker` reader (but not writer or owner) access to `myrepo` we execute:

```
$ pachctl auth set JoeyZwicker reader myrepo
```

This updated scope can then be verified with `pachctl auth get ...`:

```
$ pachctl auth get JoeyZwicker myrepo                        
READER
```

## The behavior of pipelines when using ACLs

In Pachyderm, you don't explicitly set the scope of access for users on pipelines.  Rather, pipelines infer access from the repositories that are input to the pipeline.  

## Behavior when activation codes expire

When an enterprise activation code expires, an auth-activated Pachyderm cluster goes into an "admin only" state.  In this state, only admins will access access to data that is in Pachyderm.  This safety measure keeps sensitive data protected, even when an enterprise subscription becomes stale.
