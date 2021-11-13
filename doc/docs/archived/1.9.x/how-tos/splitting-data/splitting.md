# Splitting Data for Distributed Processing

Before you read this section, make sure that you understand
the concepts described in
[Distributed Computing](../distributed_computing.md).

Pachyderm enables you to parallelize computations over data as long as
that data can be split up into multiple *datums*.  However, in many
cases, you might have a dataset that you want or need to commit
into Pachyderm as a single file rather than a bunch of smaller
files that are easily mapped to datums, such as one file per record.
For such cases, Pachyderm provides an easy way to prepare your dataset
for subsequent distributed computing by splitting it upon uploading
to a Pachyderm repository.

In this example, you have a dataset that consists of information about your
users and a repository called `user`.
This data is in `CSV` format in a single file called `user_data.csv`
with one record per line:

```
$ head user_data.csv
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

If you put this data into Pachyderm as a single
file, Pachyderm processes them a single datum.
It cannot process each of
these user records in parallel as separate `datums`.
Potentially, you can manually separate
these user records into standalone files before you
commit them into the `users` repository or through
a pipeline stage dedicated to this splitting task.
However, Pachyderm provides an optimized way of completing
this task.

The `put file` API includes an option for splitting
the file into separate datums automatically. You can use
the `--split` flag with the `put file` command.

To complete this example, follow the steps below:

1. Create a `users` repository by running:

   ```shell
   $ pachctl create repo users
   ```

1. Create a file called `user_data.csv` with the
contents listed above.

1. Put your `user_data.csv` file into Pachyderm and
automatically split it into separate datums for each line:

   ```shell
   $ pachctl put file users@master -f user_data.csv --split line --target-file-datums 1
   ```

   The `--split line` argument specifies that Pachyderm
   splits this file into lines, and the `--target-file-datums 1`
   argument specifies that each resulting file must include
   at most one datum or one line.

1. View the list of files in the master branch of the `users`
repository:

   ```shell
   $ pachctl list file users@master
   NAME                 TYPE                SIZE
   user_data.csv   dir                 5.346 KiB
   ```

   If you run `pachctl list file` command for the master branch
   in the `users` repository, Pachyderm
   still shows the `user_data.csv` entity to you as one
   entity in the repo
   However, this entity is now a directory that contains all
   of the split records.

1. To view the detailed information about
the `user_data.csv` file, run the command with the file name
specified after a colon:

   ```shell
   $ pachctl list file users@master:user_data.csv
   NAME                             TYPE                SIZE
   user_data.csv/0000000000000000   file                43 B
   user_data.csv/0000000000000001   file                39 B
   user_data.csv/0000000000000002   file                37 B
   user_data.csv/0000000000000003   file                34 B
   user_data.csv/0000000000000004   file                35 B
   user_data.csv/0000000000000005   file                41 B
   user_data.csv/0000000000000006   file                32 B
   etc...
   ```

   Then, a pipeline that takes the repo `users` as input
   with a glob pattern of `/user_data.csv/*` processes each
   user record, such as each line in the CSV file in parallel.

### JSON and Text File Splitting Examples

Pachyderm supports this type of splitting for lines or
JSON blobs as well. See the examples below.

* Split a `json` file on `json` blobs by putting each `json`
blob into a separate file.

  ```shell
  $ pachctl put file users@master -f user_data.json --split json --target-file-datums 1
  ```

* Split a `json` file on `json` blobs by putting three `json`
blobs into each split file.

  ```shell
  $ pachctl put file users@master -f user_data.json --split json --target-file-datums 3
  ```

* Split a file on lines by putting each 100-bytes chunk into
the split files.

  ```shell
  $ pachctl put file users@master -f user_data.txt --split line --target-file-bytes 100
  ```

## Specifying a Header

If your data has a common header, you can specify it
manually by using `pachctl put file` with the `--header-records` flag.
You can use this functionality with JSON and CSV data.

To specify a header, complete the following steps:

1. Create a new or use an existing data file. For example, the `user_data.csv`
from the section above with the following header:

   ```shell
   NUMBER,EMAIL,IP_ADDRESS
   ```

1. Create a new repository or use an existing one:

   ```shell
   $ pachctl create repo users
   ```

1. Put your file into the repository by separating the header from
other lines:

   ```shell
   $ pachctl put file users@master -f user_data.csv --split=csv --header-records=1 --target-file-datums=1
   ```

1. Verify that the file was added and split:

   ```shell
   $ pachctl list file users@master:/user_data.csv
   ```

   **Example:**

   ```shell
   NAME                            TYPE SIZE
   /user_data.csv/0000000000000000 file 70B
   /user_data.csv/0000000000000001 file 66B
   /user_data.csv/0000000000000002 file 64B
   /user_data.csv/0000000000000003 file 61B
   /user_data.csv/0000000000000004 file 62B
   /user_data.csv/0000000000000005 file 68B
   /user_data.csv/0000000000000006 file 59B
   /user_data.csv/0000000000000007 file 59B
   /user_data.csv/0000000000000008 file 71B
   /user_data.csv/0000000000000009 file 65B
   ```

1. Get the first file from the repository:

   ```shell
   $ pachctl get file users@master:/user_data.csv/0000000000000000
   NUMBER,EMAIL,IP_ADDRESS
   1,cyukhtin0@stumbleupon.com,144.155.176.12
   ```
1. Get all files:

   ```csv
   $ pachctl get file users@master:/user_data.csv/*
   NUMBER,EMAIL,IP_ADDRESS
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

For more information, type `pachctl put file --help`.

## Ingesting PostgresSQL data

Pachyderm supports direct data ingestion from PostgreSQL.
You need first extract your database into a script file
by using `pg_dump` and then add the data from the file
into Pachyderm by running the `pachctl put file` with the
`--split` flag.

When you use `pachctl put file --split sql ...`, Pachyderm
splits your `pgdump` file into three parts - the header, rows,
and the footer. The header contains all the SQL statements
in the `pgdump` file that set up the schema and tables.
The rows are split into individual files, or if you specify
the `--target-file-datums` or `--target-file-bytes`, multiple
rows per file. The footer contains the remaining
SQL statements for setting up the tables.

The header and footer are stored in the directory that contains
the rows. If you request a `get file` on that directory, you
get just the header and footer. If you request an individual
file, you see the header, the row or rows, and the footer.
If you request all the files with a glob pattern, for example,
`/directoryname/*`, you receive the header, all the rows, and
the footer recreating the full `pgdump`. Therefore, you can
construct full or partial `pgdump` files so that you can
load full or partial datasets.

To put your PostgreSQL data into Pachyderm, complete the following
steps:

1. Generate a `pgdump` file:

   **Example:**

   ```shell
   $ pg_dump -t users -f users.pgdump
   ```

1. View the `pgdump` file:

???+ note "Example"

    ```shell
    $ cat users.pgdump
    --
    -- PostgreSQL database dump
    --

    -- Dumped from database version 9.5.12
    -- Dumped by pg_dump version 9.5.12

    SET statement_timeout = 0;
    SET lock_timeout = 0;
    SET client_encoding = 'UTF8';
    SET standard_conforming_strings = on;
    SELECT pg_catalog.set_config('search_path', '', false);
    SET check_function_bodies = false;
    SET client_min_messages = warning;
    SET row_security = off;

    SET default_tablespace = '';

    SET default_with_oids = false;

    --
    -- Name: users; Type: TABLE; Schema: public; Owner: postgres
    --

    CREATE TABLE public.users (
        id integer NOT NULL,
        name text NOT NULL,
        saying text NOT NULL
    );


    ALTER TABLE public.users OWNER TO postgres;

    --
    -- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
    --

    COPY public.users (id, name, saying) FROM stdin;
    0	wile E Coyote	...
    1	road runner	\\.
    \.


    --
    -- PostgreSQL database dump complete
    --
    ```

3.  Ingest the SQL data by using the `pachctl put file` command
    with the `--split` file:

    ```shell
    $ pachctl put file data@master -f users.pgdump --split sql
    $ pachctl put file data@master:users --split sql -f users.pgdump
    ```

4. View the information about your repository:

   ```shell

   $ pachctl list file data@master
   NAME         TYPE SIZE
   users        dir  914B
   ```

   The `users.pgdump` file is added to the master branch in the `data`
   repository.

5. View the information about the `users.pgdump` file:

   ```shell

   $ pachctl list file data@master:users
   NAME                           TYPE SIZE
   /users/0000000000000000        file 20B
   /users/0000000000000001        file 18B
   ```

6. In your pipeline, where you have started and forked PostgreSQL,
you can load the data by running the following or a similar script:

   ```
   $ cat /pfs/data/users/* | sudo -u postgres psql
   ```

   By using the glob pattern `/*`, this code loads each raw PostgreSQL chunk
   into your PostgreSQL instance for processing by your pipeline.


!!! tip
    For this use case, you might want to use `--target-file-datums` or
    `--target-file-bytes` because these commands enable your queries to run
    against many rows at a time.
