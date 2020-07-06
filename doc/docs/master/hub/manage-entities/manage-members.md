# Manage Members

Within an organization, each uniquely authenticated GitHub or Google
account is considered a member. All members in Pachyderm Hub get
administrative access to the corresponding Pachyderm Hub organization
and all workspaces.
In the current Pachyderm version, other roles are not yet supported.
The first member that has created the organization is responsible for
billing. This role can later be transferred to other members of the
organization.

Each type of account has a limit on the number of members that you can
invite to your organization. View biling details in your account
profile to learn how many users you can add.

## Add a New Member

You can invite new members into your organization when you sign up
for a paid account or any time later. Pachyderm Hub supports GitHub
and Google authentication, therefore each member that you invite must
have either GitHub or Google account.

Pending invites can be canceled before they are accepted.

To add a new member, complete the following steps:

1. Go to **Members**.
1. Click **Invite New Member**.
1. Type email addresses of the member or members that you want to invite.
1. Click **Invite**.

   Pachyderm Hub sends invitations to the provided emails. The new members
   will need to accept the invitations by logging in to Pachyderm.

## Modify Access to Pipelines

Pachyderm includes pipeline-level access control that enables you to restrict
user access to specific pipelines and repositories. Read more in
[Configure Access Controls](../../../enterprise/auth/auth/).

## Deactivate a Member

When you need to disable access to Pachyderm Hub for a specific member,
you can deactivate that user in the system. After deactivating, the user
will no longer have access to any workspaces, pipelines, or data in the
current organization. Members can be restored at any time.

To deactivate a member in Pachyderm Hub, complete the following steps:

1. Go to **Members**.
1. Select a member and click **Deactivate** under **ACTIONS**.

## Restore a Member

You can restore access to a Pachyderm Hub organization for a member
that was previously deactivated.

To restore a member, complete the following steps:

1. Go to **Members**.
1. Select a member and click **Reactivate** under **ACTIONS**.
