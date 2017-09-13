# Access Controls

The access control features of Pachyderm Enterprise let you create and manage various users that interact with your Pachyderm cluster.  You can restrict access to individual data repositories on a per user basis and, as a result, restrict the subscription of pipelines to those data respositories.  

These docs will guide you through:

1. [Activating access control features (aka "auth" features).](#activating-access-control)
2. [Logging into Pachyderm.](#logging-into-pachyderm)
3. [Managing/updating user access to data repositories.](#managing-and-updating-user-access)

We will also discuss:

- [The behavior of pipelines when using access control](#behavior-of-pipelines-as-related-to-access-control)
- [The behavior of a cluster when access control is de-activated or an enterprise token expires](#activation-code-expiration-and-de-activation)

## Activating access control

First, you will need to make sure that your cluster has Pachyderm Enterprise Edition activated (you can follow [this guide](deployment.md) to activate Enterprise Edition).  The status of the Enterprise features can be verified with `pachctl`:

```
$ pachctl enterprise get-state
ACTIVE
```

Next, we need to activate the Enterprise access control features. This can be done with `pachctl auth activate ...`.  However, before executing that command, we should decide on at least one user that will have admin priviledges on the cluster. Pachyderm admins will be able to modify the scope of access for any non-admin users on the cluster.

All users in Pachyderm are identified by their GitHub usernames.  Thus, if we wanted to activate access control on a cluster and set the GitHub user `dwhitena` as an admin, we would execute the following command:

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

When visiting this link in a browser, you will be presented with an option to "Authorize Pachyderm" (assuming that you haven't authorized Pachyderm via GitHub previously). Once you authorize Pachyderm, you will be presented with a Pachyderm user token:

![alt tag](auth.png)

Copy and paste this token back into the terminal, as requested by `pachctl`, and press enter.  You are now logged in to Pachyderm!  

## Managing and updating user access

Let's suppose that we create a repository call `myrepo` with the user `dwhitena`.  Because, the user `dwhitena` created this repository, `dwhitena` will have full read/write access to the repo.  This can be confirmed via the `pachctl auth check ...` command:

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

An OWNER of `myrepo` or a cluster admin can then set other user's scope of access to the repo.  This can be done via the `pachctl auth set ...` command.  For example, to give the GitHub user `JoeyZwicker` READER (but not WRITER or OWNER) access to `myrepo` we execute:

```
$ pachctl auth set JoeyZwicker reader myrepo
```

This updated scope can then be verified with `pachctl auth get ...`:

```
$ pachctl auth get JoeyZwicker myrepo                        
READER
```

## Behavior of pipelines as related to access control

In Pachyderm, you don't explicitly set the scope of access for users on pipelines.  Rather, pipelines infer access from the repositories that are input to the pipeline, as follows:

- An OWNER, WRITER, or READER of a repo may subscribe a pipeline to that repo.
- When a user subscribes a pipeline to a repo, they will be set as an OWNER of that pipeline's output repo.
- The initial OWNER of a pipeline's output repo (or an admin) needs to set the scope of access for other users to that output repo.  

## Activation code expiration and de-activation

When an enterprise activation code expires, an auth-activated Pachyderm cluster goes into an "admin only" state.  In this state, only admins will have access to data that is in Pachyderm.  This safety measure keeps sensitive data protected, even when an enterprise subscription becomes stale. As soon as the enterprise activation code is updated (via the dashboard or via `pachctl enterprise activate ...`), the Pachyderm cluster will return to it's previous state.

When access controls are de-activated on a Pachyderm cluster via `pachctl auth deactivate`, the cluster returns to being a non-access controlled Pachyderm cluster.  That is, all ACLs are deleted and the cluster returns to being a blank slate in regards to access control.
