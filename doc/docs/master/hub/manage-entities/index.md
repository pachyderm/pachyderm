# Overview

When you interact with Pachyderm Hub, you create an account,
organizations, and other entities to gain access to Pachyderm
Hub resources. Creating an account is the first
step in this workflow.

An account allows you to create workspaces and access
them from anywhere. In Pachyderm Hub, you can create various types
of accounts, including a free account for testing purposes. For more
information, see [Pachyderm Hub Accounts](TBA)

When you create a Pachyderm Hub account, you accept our Terms and
Conditions and Privacy Policy.

## Manage Organizations

!!! note
    Organizations are available for paid accounts. 

An organization is a collection of users that are associated with
the same entity, such as a company, and belong to the same billing
profile. When you sign up for a paid account, you are prompted to
create an organization.
The first paid user authorized with Pachyderm Hub gets Admin permissions
and can invite new members. In an organization, you can create multiple
workspaces and have a more granular control over resources and members.

To create an organization, sign up for a Pro account as described in
[Create a Pro Account](TBA). 

## Manage Workspaces

A workspace is a collection of infrastructure resources that run
Pachyderm components needed to process your data. Workspace resource
limitations are defined by the type of your account. By default,
all members of an organization have access to all workspaces. You
can restrict user access by defining a access control list (ACL) for
each workspace.

Each workspace has isolated computing, processing, storage, and other
resources. Resource autoscaling is managed by Pachyderm behind the
scenes. Each workspace in an organization must have a unique name.

## Manage Members

A member is a user account of a Pachyderm Hub organization that represents
a person interacting with Pachyderm Hub. A member is associated with a
GitHub or Google account, therefore, you must have either of these
to log in to Pachyderm Hub. All users in Pachyderm Hub get administrative
access to the corresponding Pachyderm Hub organization and workspaces.
The first user that has created the organization is responsible for
billing. This role can later be transferred to other users.





