# Splitting Data for Distributed Processing

Before your read this section, make sure that you understand
the concepts described in
[Distributed Computing](http://pachyderm.readthedocs.io/en/latest/fundamentals/distributed_computing.html).

Pachyderm enables you to parallelize computations over data as long as
that data can be split up into multiple *datums*.  However, in many
cases, you might have a dataset that you want or need to commit
into Pachyderm as a single file rather than a bunch of smaller
files that are easily mapped to datums, such as one file per record.
For these cases, Pachyderm provides an easy way to automatically
split your dataset for subsequent distributed computing.

For example, you have a dataset that consists of information about your
users. This data is in `CSV` format in a single file called `user_data.csv`
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
file, you cannot process each of
these user records in parallel as separate `datums`.
Potentially, you can manually separate each of
these user records into separate files before you
commit them into the `users` repository or through
a pipeline stage dedicated to this splitting task.
However, Pachyderm provides an optimized way of completing
this task.

The `put file` API includes an option for splitting
the file into separate datums automatically. You can use
the `--split` flag with the `put file` command. For
example, to automatically split the `user_data.csv` file
into separate datums for each line, run the
following command:

```
$ pachctl put file users@master -f user_data.csv --split line --target-file-datums 1
```

The `--split line` argument specifies that Pachyderm
splits this file into lines, and the `--target-file-datums 1`
argument specifies that each resulting file must include
at most one datum or one line.
If you run `pachctl list file` command for the master branch
in the users repository, Pachyderm
still shows the `user_data.csv` entity to you as one
entity in the repo:

```
$ pachctl list file users@master
NAME                 TYPE                SIZE
user_data.csv   dir                 5.346 KiB
```

However, this entity is now a directory that contains all
of the split records. To view the detailed information about
the `user_data.csv` file, run the command with the file name
specified after a colon: 

```
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

### Examples

Pachyderm supports this type of splitting on lines or on
JSON blobs as well. See the examples below.

* Split a `json` file on `json` blobs by putting each `json`
blob into its own file.

  ```bash
  $ pachctl put file users@master -f user_data.json --split json --target-file-datums 1
  ```

* Split a `json` file on `json` blobs by putting three `json`
blobs into each split file.

  ```bash
  $ pachctl put file users@master -f user_data.json --split json --target-file-datums 3
  ```

* Split a file on lines by putting each 100-bytes chunk into
the split files.

  ```bash
  $ pachctl put file users@master -f user_data.txt --split line --target-file-bytes 100
  ```

## Specifying a Header or Footer

Additionally, if your data has a common header or footer, you can specify these
manually via `pachctl put-header` or `pachctl put-footer`. This is helpful for CSV data.

To do this, you need to specify the header and footer in the
`_parent directory_` of your data. By specifying the header or
footer or both you are embedding them into the directory. Then
Pachyderm applies that header or footer or both to all the files in
that directory.

The example below demonstrates splitting of an CSV file
with a header and then setting the header explicitly.
Once you set the header, whenever you get a file under that directory,
the header is applied. You can still use glob patterns to get all
the data under the directory. In that case, the header is still applied.

1. View a raw CSV file:

   ```bash
   $ cat users.csv

   id,name,email
   4,alice,aaa@place.com
   7,bob,bbb@place.com
   ```

1. Take the raw CSV data minus the header and split it into multiple
files:

   ``` bash
   $ cat users.csv | tail -n +2 | pachctl put file bar@master:users --split line
   Reading from stdin.
   ```
1. View the repository:

   ```bash
   $ pachctl list file bar@master
   NAME  TYPE SIZE
   users dir  42B

1. View the detailed information about the file:

   ```bash
   $ pachctl list file bar@master:users/
   NAME                    TYPE SIZE 
   /users/0000000000000000 file 22B  
   /users/0000000000000001 file 20B  
   ```
1. Read the file:

   ```bash
   $ pachctl get file bar@master:users/0000000000000000
   4,alice,aaa@place.com
   ```
   Before you set the header, you see raw data when you run `get file`.

1. Apply a CSV header to the directory:

   ```bash
   $ cat users.csv | head -n 1 | pachctl put-header bar master users
   ```

1. Re-read the file:

   ```bash
   $ pachctl get file bar@master:users/0000000000000000
   id,name,email
   4,alice,aaa@place.com
   ```

   When you read an individual file now, you see the header and the contents.

#  If you issue a 'get file' on the directory, it returns just the header/footer

1. Run `get file` on the directory:

   ```bash

   $ pachctl get file bar@master:users
   id,name,email
   ```

   If you issue a 'get file' on the directory, it returns just the header or
   footer, or both.

1. Use the glob pattern flag to get the entire CSV file:

   ```bash
   $ pachctl get file bar@master:users/*
   id,name,email
   4,alice,aaa@place.com
   7,bob,bbb@place.com
   ```

1. To delete the existing header, run the following command

   ```bash
   $ echo "" | pachctl put-header repo branch path -f -
   ```

1. Get the file after deleting the header:

   ```
   $ pachctl get file bar@master:users/*
   4,alice,aaa@place.com
   7,bob,bbb@place.com
   ```

For more information about operations with headers and footers,
see `pachctl put-header --help`.

## Ingesting PostgresSQL data

Pachyderm supports direct data ingestion from PostgreSQL.
You need first extract your database into a script file
by using `pg_dump` and then add the data from the file
into Pachyderm by running the `pachctl put file` with the
`--split` flag.

When you use `pachctl put file --split sql ...`, Pachyderm
splits you `pgdump` file into three parts - the header, rows,
and the footer. The header contains all the SQL statements
in the `pgdump` file that set up the schema and tables.
The rows are split into individual files, or, if you specify
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

   ```bash
   $ pg_dump -t users -f users.pgdump
   ```

1. View the `pddump` file

   **Example:**

   ```bash
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

1. Ingest SQL data using split file

   ```bash
   $ pachctl put file data@master -f users.pgdump --split sql
   $ pachctl put file data@master:users --split sql -f users.pgdump 
   ```

1. View the information about your repository:

   ```bash

   $ pachctl list file data@master
   NAME         TYPE SIZE 
   users        dir  914B 
   ```

   The `users` pgdump file is added to the master branch in the `data`
   repository.

1. View the information about the `users` pgdump file:

   ```bash

   $ pachctl list file data@master:users
   NAME                           TYPE SIZE 
   /users/0000000000000000        file 20B  
   /users/0000000000000001        file 18B  
   ```

1. In your pipeline, where you have started and forked PostgreSQL,
you can load the data by running the following or a similar script:

   ```
   $ cat /pfs/data/users/* | sudo -u postgres psql
   ```

   By using the glob pattern `/*` this code loads each raw PostgreSQL chunk
   into your PostgreSQL instance for processing by your pipeline.


   **Tip:** For this use case, you might want to use `--target-file-datums` or 
   `--target-file-bytes` because you might want to run your queries
   against many rows at a time.
