# Export Your Data with `pachctl`

The `pachctl get file` command enables you to get the contents
of a file in a Pachyderm repository. You need to know the file
path to specify it in the command.

To export your data with pachctl:

1. Get the list of files in the repository:

   ```bash
   pachctl list file <repo>@<branch>
   ```

   **Example:**

   ```bash
   pachctl list commit data@master
   ```

   **System Response:**

   ```bash
   REPO   BRANCH COMMIT                           PARENT                           STARTED           DURATION           SIZE
   data master 230103d3c6bd45b483ab6d0b7ae858d5 f82b76f463ca4799817717a49ab74fac 2 seconds ago  Less than a second 750B
   data master f82b76f463ca4799817717a49ab74fac <none>                           40 seconds ago Less than a second 375B
   ```

1. Get the contents of a specific file:

   ```bash
   pachctl get file <repo>@<branch>:<path/to/file>
   ```

   **Example:**

   ```bash
   pachctl get file data@master:user_data.csv
   ```

   **System Response:**

   ```bash
   1,cyukhtin0@stumbleupon.com,144.155.176.12
   2,csisneros1@over-blog.com,26.119.26.5
   3,jeye2@instagram.com,13.165.230.106
   4,rnollet3@hexun.com,58.52.147.83
   5,bposkitt4@irs.gov,51.247.120.167
   6,vvenmore5@hubpages.com,161.189.245.212
   7,lcoyte6@ask.com,56.13.147.134
   8,atuke7@psu.edu,78.178.247.163
   9,nmorrell8@howstuffworks.com,28.172.10.170
   10,afynn9@google.com.au,166.14.112.65
   ```

   Also, you can view the parent, grandparent, and any previous
   revision by using the caret (`^`) symbol with a number that
   corresponds to an ancestor in sequence:

   * To view a parent of a commit:

     1. List files in the parent commit:

        ```bash
        pachctl list commit <repo>@<branch-or-commit>^:<path/to/file>
        ```

     1. Get the contents of a file:

        ```bash
        pachctl get file <repo>@<branch-or-commit>^:<path/to/file>
        ```

   * To view an `<n>` parent of a commit:

     1. List files in the parent commit:

        ```bash
        pachctl list commit <repo>@<branch-or-commit>^<n>:<path/to/file>
        ```

        **Example:**

        ```bash
        NAME           TYPE SIZE
        /user_data.csv file 375B
        ```

     1. Get the contents of a file:

        ```bash
        pachctl get file <repo>@<branch-or-commit>^<n>:<path/to/file>
        ```

        **Example:**

        ```bash
        pachctl get file datas@master^4:user_data.csv
        ```

     You can specify any number in the `^<n>` notation. If the file
     exists in that commit, Pachyderm returns it. If the file
     does not exist in that revision, Pachyderm displays the following
     message:

     ```bash
     pachctl get file <repo>@<branch-or-commit>^<n>:<path/to/file>
     ```

     **System Response:**

     ```bash
     file "<path/to/file>" not found
     ```
