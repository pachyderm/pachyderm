# Authorization
!!! Note
    Go back to our [Enterprise landing page](https://docs.pachyderm.com/latest/enterprise/) if you do not have a key.
    Before setting up a Role-Based access control to Pachyderm resources, verify that:

    - the [User Access Management](../index.md/#activate-user-access-management) feature is on by running `pachctl auth whoami`. The command should return `You are "pach:root"` (i.e., your are the **Root User** with `clusterAdmin` privileges). Run `pachctl auth use-auth-token` to login as a Root User.
    - your [IdP connector is set](../authentication/idp-dex.md).

Pachyderm authorization feature follows a **Role Based Access Control** model (RBAC).
The access control is based on **Roles**  assigned to **Users**, granting them a set of permissions on Pachyderm's **Ressources** (**Role Bindings**). 

In this page we will introduce Pachyderm's various Type of Users, Ressources, and Roles.

This preamble will set the contextual knowledge to better understand how to:

- Grant Users permissions on Pachyderm Ressources.
- Revoke Users.

in the next [**Role Binding**](role-binding.md) page.


## Users Types
Pachyderm defines 4 type of User: 

- A **Root User**: This special user created when Auth is activated. The root user **always has clusterAdmin permissions on the cluster**.

- An **IdP User**: Any user or group of users authenticated by your Identity Provider to access Pachyderm.

- A **Robot User**: A Service account used for third party applications/systems integrating with Pachyderm APIs/Clients.

- A **Pipeline User**: An internal Service Account used for Pipelines when interacting with Pachyderm ressources.


## Resources
Pachyderm has 3 types of resources: **Repositories**: `repo`, **Projects**: `project`???, **Clusters**: `cluster`. - //TODO Clarify `enterprise`? project?

Clusters contain one to many projects, and projects contain one to many repositories.

## Roles
Pachyderm has 4 predefined roles (//TODO more for projects) granting privileges to its Resources.
Those Roles are listed in such an order that any following Role adds a set of permissions to its predecessor.

- **repoReader**: Can consume data from a repo, but cannot edit them.
repoReader can execute commands such as `pachctl get file` and
`pachctl list file`, as well as create pipelines that use data
from this repo. 

- **repoWriter**: Can read and modify data in a repository by
adding, deleting, or updating the files in the repo. The
`repoWriter` role can perform operations such as `pachctl put file` or
`pachctl delete commit`...

- **repoOwner**: Additionaly to having the repoReader and the repoWriter Roles,
a RepoOwner can grant permission to users on the Repository.

!!! Note
    repoReader, repoWriter, and repoOwner can be set at all levels: cluster, project, and repo. 

- **clusterAdmin**: A clusterAdmin can perform any
action on the cluster including appointing other clusterAdmins.
By default, the activation of Auth (`pachctl auth activate`) creates a Root User
with irrevocable ClusterAdmin rights. Set at the cluster level only.

!!! Note
    When Pachyderm Auth is not enabled,
    all users are clusterAdmins.

