# Authorization using Role-Based Access Control

Pachyderm authorization system implements a **Role Based Access Control** model (RBAC) to provides fine-grained access management of its **Ressources**.

The access control is based on **Roles**  assigned to **Users** (or **Subjects**)  granting them a set of permissions on Pachyderm's **Ressources** (**Role Bindings**). 

In this page we will 
- List Pachyderm's Type of Users, Ressources, and Roles
- Detail how Users are granted permissions on Pachyderm Ressources
- Take a look at Users revocation


## Users Types
Pachyderm defines 4 type of User: 

- A **Root User**: This special user created when Auth is activated. The root user **always has clusterAdmin permissions on the cluster**.

- An **IdP User**: Any user or group of users authenticated by your Identity Provider to access Pachyderm.

- A **Robot User**: A Service account used for third party applications/systems integrating with Pachyderm APIs/Clients.

- A **Pipeline User**: An internal Service Account used for Pipelines when interacting with Pachyderm ressources.


## Resources
Pachyderm has 3 types of resources: **Repositories**, **Projects** and **Clusters**.

Clusters contain one to many projects, and projects contain one to many repositories.


## Roles
Pachyderm has 4 predefined roles (//TODO more for projects) granting privileges to its Resources.
Those Roles are listed in such an order that any following Role adds a set of permissions to its predescessor.

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

- **clusterAdmin**: A clusterAdmin can perform any
action on the cluster including appointing other clusterAdmins.
By default, the activation of Auth (`pachctl auth activate`) creates a Root User
with irrevocable ClusterAdmin rights.

!!! Note
    When Pachyderm Auth is not enabbled,
    all users are clusterAdmins.


## Role Binding
Role Binding (the act of granting a given User a specific Role on a given Resource) applies to IdP Users only.
Any other type of User is either autocatically created with a predefined Role (for example the Root User and the Robot User are default clusterAdmins) or inherit their Roles from the ressource they subscribe to (This is the case of the Pipeline User). We will cover the case of the Pipeline User in detail.
###
### Example
### Pipeline users
Each Role can be binded to any Ressource and assigned a given IdP User as we will see in the following example.
Remember, an IdP user can be a group of users.





- RBAC
- Users/Groups (Principal)
- Roles
- Role Binding
- Access Control configuration
    - API
    - UI