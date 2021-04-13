# Authorization
!!! Note
    Before setting up a Role-Based access control to Pachyderm resources, verify that
    - the [User Access Management](../index.md/#activate-user-access-management) feature is on by running `pachctl auth whoami`. The command should return `You are "pach:root" (i.e., your are the **Root User** with `clusterAdmin` privileges). Run `pachctl auth use-auth-token` to login as a Root User.
    - your [IdP connector is set](../authentication/idp-dex.md).

Pachyderm authorization feature follows a **Role Based Access Control** model (RBAC).
The access control is based on **Roles**  assigned to **Users**, granting them a set of permissions on Pachyderm's **Ressources** (**Role Bindings**). 

In this page we will:

- List Pachyderm's Type of Users, Ressources, and Roles.
- Detail how Users are granted permissions on Pachyderm Ressources.
- Take a look at Users revocation.


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


## Role Binding

This chapter will detail how to:
- Grant/modify permissions (Roles) on given Resources to a User (Idp or Robot User).
- Remove all permissions on a Ressource from a User (Idp or Robot User).

!!! Note "Default Privileges"
    - The case of the Root User: The activation of the Authentication and Authorization feature generates a **Root User** with **unalterable and unrevokable `clusterAdmin` privileges**. 
    - The case of the Robot User: **Robot users do not have any permission by default**. They will need to be set by a `clusterAdmin`.
    - The case of the Pipeline User: In Pachyderm, **you do not explicitly grant users access to pipelines**, they get set for you when you create or update a pipeline. 

!!! Warning "Rules to keep in mind"
    - A user or group can have one or more roles on a specific Resource.
    - Roles are inherited: if a user has a role on a cluster, they have that role for all projects and repos in that cluster.
    - The creator of a repo becomes its `repoOwner`.
    - To update a pipeline, you must have at least `repoReader`-level access to all pipeline inputs
        and `repoWriter`-level access to the pipeline output. 
        This is because pipelines read from their input repos and write
        to their output repos.
    - When a user subscribes a pipeline to a repo, Pachyderm sets
        that user as an `repoOwner` of that pipeline's output repo.
        If additional users need access to the output repository,
        the initial `repoOwner` of a pipeline's output repo, or a `clusterAdmin`,
        needs to grant that user access to the repo.
      
      
### Set Roles to Users

- A **clusterAdmin** can grant admin privileges on a cluster or any lower level ressources to other users.
- A **repoOwner** of a given repository (or a **clusterAdmin** as mentioned above) can set any level of access to "their" repo to users by running the command:

```shell
$ pachctl auth set <ressource> <ressource name> [role1,role2 | none ] <prefix:subject>
```

To keep using our Auth0 example and illustrate the attribution of a given Role to a User,
let's have our `Root User` (with default clusterAdmin privileges) give access to a repo to our `one-pachyderm-user@gmail.com` user:

1. Connect as our Root User again.
1. Create a repo named `testinput` containing one text file.
1. Grant access on this repo to our user `one-pachyderm-user@gmail.com` registered with our IdP (Auth0).
1. See what happens when `one-pachyderm-user@gmail.com` tries to write in the repo without the proper writing access.

- 1- Connect as our Root User:
    ```shell
    $ pachctl auth use-auth-token
    ```
    You will be asked to re-enter your Root token.

- 2- Create a Repo:
    ```shell
    $ mkdir -p ./testinput 
    $ printf "this is a test" >./testinput/test.txt
    $ pachctl create repo testinput
    $ cd testinput && pachctl put file testinput@master -f test.txt

    ```
    A quick `pachctl list repo` will list your new repo and display your access level on that repo as a **clusterAdmin**.
    ![Admin Repo Access Level](../images/clusteradmin-repo-access.png)

- 3- Grant access on this repo to our user `one-pachyderm-user@gmail.com`:
    For example, let's give our user a **repoReader** access to the repo `testinput`.

    ```shell
    $ pachctl auth set repo testinput repoReader user:one-pachyderm-user@gmail.com
    ```
    ... and take a quick look at his access level:
    ```shell
    $ pachctl auth get repo testinput
    ```
    The command returns the list of users granted access to this repo and their associated access level. 
    ```
    user:one-pachyderm-user@gmail.com: [repoReader]
    pach:root: [repoOwner]
    ```

    !!! Note
        Note that the user `one-pachyderm-user@gmail.com` has a prefix `user`.
        Pachyderm defines 4 prefixes depending on the type of user:
            - robot
            - user
            - group
            - pipeline (as mentioned above, this prefix will not be used in the context of granting privileges to users. However, it does exist. We are listing it here to give an exhauxtiv list of all prefixes.)

- 4- Have `one-pachyderm-user@gmail.com` try to add a file to `testinput` without proper writing access:
    ```shell
    # Login as `one-pachyderm-user@gmail.com`
    $ pachctl auth login
    # Try to write into testinput repo
    $ printf "this is another test" >./testinput/anothertest.txt
    $ cd testinput && pachctl put file testinput@master -f anothertest.txt
    ```
    The command returns an error message:
    ```
    user:one-pachyderm-user@pachyderm.io is not authorized to perform this operation - needs permissions [REPO_WRITE] on REPO testinput
    ```

!!! Info
    Use `--help` to display the list of all available commands, arguments, and flags of the command `pachctl auth set`.

!!! Note
    - To alter a user's privileges, simply re-run the `pachctl auth set` command above with a different set of Roles. 
    For example, 
    ```shell
    $ pachctl auth set repo testinput repoWriter user:one-pachyderm-user@gmail.com
    ```
    will give one-pachyderm-user@gmail.com `repoWriter` privileges when she/he was  inially granted `repoReader` access.

    - You can remove all access level on a repo to a user by using the `none` keyword.
    For example,
    ```shell
    $ pachctl auth set repo testinput none user:one-pachyderm-user@gmail.com
    ```
    will remove any previous granted rights on the repo `testinput` to the user one-pachyderm-user@gmail.com.

### Set Roles to Groups

If your IdP enables group support,
you can grant access on Pachyderm ressources to a group of users.

Let's keep using our Auth0 example as an illustration, and:

1. As a `clusterAdmin`, create a Group in Auth0.
1. Assign our user to the newly created group.
1. Update our connector accordingly.
1. Grant the group an owner access to a specific repo in Pachyderm.

!!! Info
    To enable the Group creation in Auth0, you will need to install an [`Authorization Extension`](https://auth0.com/docs/extensions/authorization-extension) to Auth0:

    - Go to **Auth0 Dashboard > Extensions**.
    - Select **Auth0 Authorization** and answer the prompt to install.
    - Choose where you would like to store your data: **Webtask Storage** for this example and click **Install**


- 1- Group creation

    An Authorization link should now show on your Auth0 webpage.
    In **Authorization/Groups**, create a group. Here `testgroup`:
    ![Group creation](../images/auth0-create-group.png)

- 2- Add your user to your group

    In **Authorization/Users**, select your user one-pachyderm-user@gmail.com and add her/him to your `testgroup` as follow.
    ![Add User to Group](../images/auth0-add-user-to-group.png)

    In **User Mangement/Users**, you user should now show the following addition to their app_metadata:
    ```json
    {
        "authorization": {
            "groups": [
            "testgroup"
            ]
        }
    }
    ```
- 3- Update your connector

    === "oidc-dex-connector.json"

        ``` json
        {
            "issuer": "https://dev-k34x5yjn.us.auth0.com/",
            "clientID": "hegmOc5rTotLPu5ByRDXOvBAzgs3wuw5",
            "clientSecret": "7xk8O71Uhp5T-bJp_aP2Squwlh4zZTJs65URPma-2UT7n1iigDaMUD9ArhUR-2aL",
            "redirectURI": "http://<ip>:30658/callback",
            "scopes": ["groups"],
            "claimMapping":
            {
            "groups": "authorization:groups"
            }
        }
        ```

    === "oidc-dex-connector.yaml"

        ``` yaml
            connectors:
            - type: oidc
            id: auth0
            name: Auth0
            version: 1
            config:
                # Canonical URL of the provider, also used for configuration discovery.
                # This value MUST match the value returned in the provider config discovery.
                #
                # See: https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderConfig
                issuer: https://dev-k34x5yjn.us.auth0.com/

                # Connector config values starting with a "$" will read from the environment.
                clientID: hegmOc5rTotLPu5ByRDXOvBAzgs3wuw5
                clientSecret: 7xk8O71Uhp5T-bJp_aP2Squwlh4zZTJs65URPma-2UT7n1iigDaMUD9ArhUR-2aL

                # Dex's issuer URL + "/callback"
                redirectURI: http://<id>:30658/callback
                scopes:
                - groups
                claimMapping:
                    groups: "authorization:groups"
        ```

    Note the addition of the `scopes` and `claimMapping` fields to your original connector configuration file.
    Update your connector:
    ```shell
    $ pachctl idp update-connector auth0 --version 2
    ```
    You are all set to grant this new group access to Pachyderm's ressources.

- 4- Grant the group an admin access to a specific repo in Pachyderm.

    ```shell
    $ pachctl auth set repo testinput repoOwner group:testgroup
    ```
    A quick check at this repo should give you its updated list of users an their access level:
    ```shell
    $ pachctl auth get repo testinput
    ```
    //TODO test out once group support implemented all the way in alpha.12

## Example

![Role binding example](../images/role-binding-example.svg)

In this diagram, the `data-scientists` group has been assigned the `repoReader` role on the cluster. This gives them permissions to **read all repos in all projects**.

The IdP user `one-pachyderm-user@company.io` has been assigned the `repoOwner` role on the `nlp` project. This gives them permission to **read, write and grant permissions for repos within the nlp project**. 
It does not give them any permission on the `image-recognition` project, or on the `cluster` itself.

If `one-pachyderm-user@company.io` was a member of the `data-scientists` group, then they would cumulate both roles: 'repoReader' on all repo and `repoOwner` on the `nlp` project.

The IdP user `another-pachyderm-user@company.io` has been assigned the `repoWriter` role on the repo `categorize-text`. This gives them permission to **read and write in that repo**, but not to access any other repo, project, or the cluster itself.

## User Revocation
//TODO Coming soon -> In dev