# File

A file is a filesystem object that stores data. Unlike source-code
version-control systems that most suitable for storing plain text
files, you can store any type of file in Pachyderm, including
binary files. Very often, data scientists operate with
comma-separated values (CSV), JavaScript Object Notation (JSON),
Structured Query Language (SQL), and other plain text and binary file
formats. Storing and versioning large sized datasets in traditional
version control systems might not be possible because many of them
have file size limitations.

Pachyderm stores files in repositories that are directories in the
Pachyderm etcd container.

To upload your files to a Pachyderm repository, run the
`pachctl put file <name>` command. By using the `pachctl put file`
command, you can put both files and directories into a Pachyderm repository.

## File Processing Strategies

Pachyderm provides the following file processing strategies:

Appending files
 By default, Pachyderm appends all new files to the existing files.
 For example, you have an `A.csv` file in a repository. If you upload the
 same file to that repository, Pachyderm *appends* the data to the existing
 file which results in the file `A.csv` having twice the data from its
 original size.

 **Example:**

 1. View the list of files:

    ```bash
    $ pachctl list file images@master
    NAME   TYPE SIZE
    /A.csv file 258B
    ```

 1. Add the `A.csv` file once again:

    ```bash
    $ pachctl put file images@master -f A.csv
    ```

 1. Verify that the file has doubled in size:

   ```bash
   $ pachctl list file images@master
   NAME   TYPE SIZE
   /A.csv file 516B
   ```

Overwriting files
 When you enable the overwrite mode by using the `--overwrite`
 flag or `-o`, the file replaces the existing file instead of appending to
 it. For example, you have an `A.csv` file in a repository. If you upload
 the same file to that repository with the `--overwrite` flag, Pachyderm
 *overwrites* the whole file.

 **Example:**

 1. View the list of files:

    ```bash
    $ pachctl list file images@master
    NAME   TYPE SIZE
    /A.csv file 258B
    ```

 1. Add the `A.csv` file once again:

    ```bash
    $ pachctl put file --overwrite images@master -f A.csv
    ```

 1. Check the file size:

    ```bash
    $ pachctl list file images@master
    NAME   TYPE SIZE
    /A.csv file 258B
    ```
