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
Pachyderm defines 5 type of User: 

- A **Root User**: This special user is created when Auth is activated. The root user **always has [clusterAdmin](#roles) permissions on the cluster**.

- An **IdP User**: Any user or group of users authenticated by your Identity Provider to access Pachyderm.

- A **Robot User**: A Service account used for third party applications/systems integrating with Pachyderm APIs/Clients.

- A **Pipeline User**: An internal Service Account used for Pipelines when interacting with Pachyderm ressources.

- A **All Cluster Users** (`allClusterUsers`) : A general subject that represents **everyone who has logged in to a cluster**.
## Resources
Pachyderm has 2 types of resources: **Repositories**: `repo`, **Clusters**: `cluster`. 

!!! Coming soon
    Two additionnal tiers: A `project` tier between the cluster and repo levels, and the `enterprise` tier, above all clusters, at the enterprise server level, are in the works. Clusters contain one to many projects. Projects contain one to many repositories.

## Roles
Pachyderm has a number of predefined roles granting permissions to its Resources.
Those Roles are listed here in no specific hierarchical order. 
Some might inherit a set of permissions from another.

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
        repoReader, repoWriter, and repoOwner can be set at all levels: cluster, and repo. 

- **secretAdmin**: A secretAdmin has the ability to create, update, delete Kubernetes secrets on a cluster.

    !!! Note
        secretAdmin is set at the cluster level. 


- **clusterAdmin**: A clusterAdmin can perform any
action on the cluster including appointing other clusterAdmins.
By default, the activation of Auth (`pachctl auth activate`) creates a Root User
with irrevocable ClusterAdmin rights. This Role must be set at the cluster level only.

    !!! Info
        When Pachyderm Auth is not enabled,
        all users are clusterAdmins.

    The `clusterAdmin` role is broken up in a set of finer grained roles, so users can delegate specific tasks without giving full admin privileges to a cluster.
    Those roles are listed below. The union of them all is equivalent to the ClusterAdmin role.

    !!! Note
        None of those sub-roles have shared privileges.

- **idpAdmin**: An ipdAdmin can configure identity providers. They can perform operations such as `pachctl idp create-connector` or `pachctl idp update-connector`.

- **identityAdmin**: An identityAdmin has the ability to configure the identity service.  In particular, they can set an issuer: `pachctl idp set-config --issuer http://<some IP>:658` ?? //TODO What are the pachctl command available to the identityAdmin?

- **debugger**: A debugger role has the ability to produce debug dumps.

- **licenseAdmin**: This role grant the ability to register new clusters with the license server, as well as manage and update the enterprise license. For example, this role can perform a `pachctl enterprise register`, `pachctl license activate` or `pachctl license delete-cluster`. //TODO get validation

- **oidcAppAdmin**: An oidcAppAdmin can configure oidc apps between the Identity service and a cluster. In other terms, they can perform operations such as `pachctl idp create-client`. This command is rarely necessary to most users. Usually, a `pachctl deploy` will invoke a `pachctl idp create-client` under the hood and register Pachyderm's apps like the IDE, for example. However, there might be cases where this registration of new oidc clients needs do be made explicitely. This could happen in the case were someone decides to reinstall Pachyderm manually, component by component, without the automated `pachctl deploy` for example???????????. //TODO Uhm.... So not clear. Get a clearer picture of when this create-client command would be useful, when it is implicitly called, and what other commands this role has access to. 

    A use case for fine grained admin access would be Hub where we want to give `allClusterUsers` the debugger and `repoOwner` roles but not full cluster admin privileges. ????? //TODO needs validation

    !!! Note
        The above roles constituting the clusterAdmin role can be set at the cluster level only???? //TODO I do not understand how the licenseAdmin, who seem to have access to enterprise  server commands, is part of the clusterAdmin split up of roles when the clusterAdmin role is set on a cluster... Me confusion.
