# Authorization
!!! Note
    Go back to our [Enterprise landing page](https://docs.pachyderm.com/latest/enterprise/){target=_blank} if you do not have a key.
    Before setting up a Role-Based access control to Pachyderm resources, verify that:

    - the [User Access Management](../#activate-user-access-management) feature is on by running `pachctl auth whoami`. The command should return `You are "pach:root"` (i.e., your are the **Root User** with `clusterAdmin` privileges). Run `pachctl auth use-auth-token` to login as a Root User.
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

## Roles
Pachyderm has a number of predefined roles granting permissions to its Resources.
Those Roles are listed here in no specific hierarchical order. 
Some might inherit a set of permissions from another.

### Repo Roles

These roles can be granted at the repo level or at the cluster level - if a user is granted a role at the cluster level, it applies to all repos, including any new repos that are created after the grant. 

- **repoReader**: A repoReader can consume data from a repo, but cannot edit them.
repoReader can execute commands such as `pachctl get file` and
`pachctl list file`, as well as create pipelines that use data
from this repo. 

- **repoWriter**: A repoWriter can read and modify data in a repo by
adding, deleting, or updating the files in the repo. The
`repoWriter` role can perform operations such as `pachctl put file` or
`pachctl delete commit`...

- **repoOwner**: A repoOwner can read and modify data in a repo, 
update the role bindings for that repo, and delete the repo.

### Cluster Roles

These roles are only applicable at the cluster level. `clusterAdmin` is a catch-all role which allows a user to perform any operation on the cluster, while the others allow delegation of specific privileges depending on a users needs.

- **clusterAdmin**: A clusterAdmin can perform any action on the cluster.

!!! Info
    Activating auth (`pachctl auth activate`) creates a Root User with irrevocable `clusterAdmin` rights. This Role must be set at the cluster level only.

- **secretAdmin**: A secretAdmin has the ability to create, update, delete Kubernetes secrets on a cluster.

- **idpAdmin**: An ipdAdmin can configure identity providers. They can perform operations such as `pachctl idp create-connector` or `pachctl idp update-connector`.

- **identityAdmin**: An identityAdmin has the ability to configure the identity service. They can perform operations like `pachctl idp set-config` and `pachctl idp get-config`.  

- **debugger**: A debugger has the ability to produce debug dumps using `pachctl debug`.

- **licenseAdmin**: This role grant the ability to register new clusters with the license server, as well as manage and update the enterprise license. For example, this role can perform a `pachctl enterprise register`, `pachctl license activate` or `pachctl license delete-cluster`. 

- **oidcAppAdmin**: An oidcAppAdmin can configure oidc apps between the Identity service and a cluster. They can perform operations such as `pachctl idp create-client`. This role is necessary to deploy pachd, console or other apps that need to be registered with the identity service.

- **robotUser**: A robotUser has the ability to create robot users and generate auth tokens for them.

- **logReader**: A logReader can access the logs for the pachd pod using `pachctl logs`, which may contain repo names, filenames and other metadata about the contents of the cluster.
