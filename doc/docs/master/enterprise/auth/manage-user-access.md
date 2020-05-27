## Manage and update user access

You can manage user access in the UI and CLI.
For example, you are logged in to Pachyderm as the user `dwhitena`
and have a repository called `test`.  Because the user `dwhitena` created
this repository, `dwhitena` has full `OWNER`-level access to the repo.
You can confirm this in the dashboard by navigating to or clicking on
the repo `test`:

![alt tag](../../assets/images/auth_dash4.png)


 Alternatively, you can confirm your access by running the
 `pachctl auth get ...` command.

!!! example

    ```
    pachctl auth get dwhitena test
    ```

    **System response:**

    ```bash
    OWNER
    ```

An OWNER of `test` or a cluster admin can then set other user’s
level of access to the repo by using
the `pachctl auth set ...` command or through the dashboard.

For example, to give the GitHub users `JoeyZwicker` and
`msteffen` `READER`, but not `WRITER` or `OWNER`, access to
`test` and `jdoliner` `WRITER`, but not `OWNER`, access,
click on **Modify access controls** under the repo details
in the dashboard. This functionality allows you to add
the users easily one by one:

![alt tag](../../assets/images/auth_dash5.png)
