# File

A file is a Unix filesystem object, which is a directory or
file, that stores data. Unlike source code
version-control systems that are most suitable for storing plain text
files, you can store any type of file in Pachyderm, including
binary files. Often, data scientists operate with
comma-separated values (CSV), JavaScript Object Notation (JSON),
images, and other plain text and binary file
formats. Pachyderm supports all file sizes and formats and applies
storage optimization techniques, such as deduplication, in the
background.

To upload your files to a Pachyderm repository, run the
`pachctl put file` command. By using the `pachctl put file`
command, you can put both files and directories into a Pachyderm repository.

## File Processing Strategies

Pachyderm provides the following file processing strategies:

**Appending files**
:   By default, when you put a file into a Pachyderm repository and a
    file by the same name already exists in the repo, Pachyderm appends
    the new data to the existing file.
    For example, you have an `A.csv` file in a repository. If you upload the
    same file to that repository, Pachyderm *appends* the data to the existing
    file, which results in the `A.csv` file having twice the data from its
    original size.

!!! example

    1. View the list of files:

       ```shell
       pachctl list file images@master
       ```

       **System Response:**

       ```shell
       NAME   TYPE SIZE
       /A.csv file 258B
       ```

    1. Add the `A.csv` file once again:

       ```shell
       pachctl put file images@master -f A.csv
       ```

    1. Verify that the file has doubled in size:

       ```shell
       pachctl list file images@master
       ```

       **System Response:**

       ```shell
       NAME   TYPE SIZE
       /A.csv file 516B
       ```

**Overwriting files**
:   When you enable the overwrite mode by using the `--overwrite`
    flag or `-o`, the file replaces the existing file instead of appending to
    it. For example, you have an `A.csv` file in the `images` repository.
    If you upload the same file to that repository with the
    `--overwrite` flag, Pachyderm *overwrites* the whole file.

!!! example

    1. View the list of files:

       ```shell
       pachctl list file images@master
       ```

       **System Response:**

       ```shell
       NAME   TYPE SIZE
       /A.csv file 258B
       ```

    1. Add the `A.csv` file once again:

       ```shell
       pachctl put file --overwrite images@master -f A.csv
       ```

    1. Check the file size:

       ```shell
       pachctl list file images@master
       ```

       **System Response:**

       ```shell
       NAME   TYPE SIZE
       /A.csv file 258B
       ```
