# Splitting Data for Distributed Processing

As described in the [distributed computing with Pachyderm docs](http://pachyderm.readthedocs.io/en/latest/fundamentals/distributed_computing.html), Pachyderm allows you to parallelize computations over data as long as that data can be split up into multiple "datums."  However, in many cases, you might have a data set that you want or need to commit into Pachyderm as a single file, rather than a bunch of smaller files (e.g., one per record) that are easily mapped to datums.  In these cases, Pachyderm provides an easy way to automatically split your data set for subsequent distributed computing.

Let's say that we have a data set consisting of information about our users.  This data is in CSV format in a single file, `user_data.csv`,  with one record per line:

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

If we just put this into Pachyderm as a single file, we could not subsequently process each of these user records in parallel as separate "datums" (see [this guide](http://pachyderm.readthedocs.io/en/latest/fundamentals/distributed_computing.html) for more information on datums and distributed computing).  Of course, you could manually separate out each of these user records into separate files before you commit them into the `users` repo or via a pipeline stage dedicated to this splitting task.  This would work, but Pachyderm actually makes it much easier for you.

The `put-file` API includes an option for splitting up the file into separate datums automatically.  You can do this with the `pachctl` CLI tool via the `--split` flag on `put-file`.  For example, to automatically split the `user_data.csv` file up into separate datums for each line, you could execute the following:

```
$ pachctl put-file users master -c -f user_data.csv --split line --target-file-datums 1
```  

The `--split line` argument specifies that Pachyderm should split this file on lines, and the `--target-file-datums 1` arguments specifies that each resulting file should include at most one "datum" (or one line).  Note, that Pachyderm will still show the `user_data.csv` entity to you as one entity in the repo:

```
$ pachctl list-file users master
NAME                 TYPE                SIZE                
user_data.csv   dir                 5.346 KiB
```

But, this entity is now a directory containing all of the split records:

```
$ pachctl list-file users master user_data.csv
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

A pipeline that then takes the repo `users` as input with a glob pattern of `/user_data.csv/*` would process each user record (i.e., each line of the CSV) in parallel.  

This is, of course, just one example.  Right now, Pachyderm supports this type of splitting on lines or on JSON blobs.  Here are a few more examples:

```
# Split a json file on json blobs, putting
# each json blob into it's own file.
$ pachctl put-file users master -c -f user_data.json --split json --target-file-datums 1

# Split a json file on json blobs, putting
# 3 json blobs into each split file.
$ pachctl put-file users master -c -f user_data.json --split json --target-file-datums 3

# Split a file on lines, putting each 100 
# bytes chunk into the split files.
$ pachctl put-file users master -c -f user_data.txt --split line --target-file-bytes 100
```  

## Specifying Header/Footer

Additionally, if your data has a common header or footer, you can specify these
manually via `pachctl put-header` or `pachctl put-footer`. This is helpful for CSV data.

To do this, you'll need to specify the header/footer on the parent directory of your data. Below we have an example of splitting a CSV with a header, then setting the header explicitly. Notice that once we've set the header, whenever we get a file under that directory, the header is applied. You can still use glob patterns to get all the data under the directory, and in that case the header is still applied.

```
# Raw CSV
$ cat users.csv 
id,name,email
4,alice,aaa@place.com
7,bob,bbb@place.com
# Take the raw CSV data minus the header and split it into multiple files:
$ cat users.csv | tail -n +2 | pc put-file bar master users --split line
Reading from stdin.
$ pachctl list-file bar master
NAME  TYPE SIZE 
users dir  42B  
$ pachctl list-file bar master /users/
NAME                    TYPE SIZE 
/users/0000000000000000 file 22B  
/users/0000000000000001 file 20B  
# Before we set the header, we just see the raw data when we issue a get-file
$ pachctl get-file bar master /users/0000000000000000
4,alice,aaa@place.com
# Now we take the CSV header and apply it to the directory:
$ cat users.csv | head -n 1 | pachctl put-header bar master users 
# Now when we read an individual file, we see the header plus the contents
$ pachctl get-file bar master /users/0000000000000000
id,name,email
4,alice,aaa@place.com
$ pachctl get-file bar master /users
id,name,email
$ pachctl get-file bar master /users/**
id,name,email
4,alice,aaa@place.com
7,bob,bbb@place.com
```

For more info see `pachctl put-header --help`

## PG Dump / SQL Support

You can also ingest data from postgres using split file.

1) Generate your PG Dump file

```
$ pg_dump -t users -f users.pgdump
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


2) Ingest using split file

When you use `pachctl put-file --split ...` your pg dump file is split into
three parts - the header, rows, and the footer. The header contains all the SQL
statements in the pg dump that setup the schema and tables. The rows are split
into individual files (or if you specify the `--target-file-datums` or 
`--target-file-bytes` multiple rows per file). The footer contains the remaining
SQL statements for setting up the tables.

The header and footer are stored on the directory containing the rows. This way,
if you request a `get-file` on the directory, you'll get just the header and
footer. If you request an individual file, you'll see the header plus the row(s)
plus the footer. If you request all the files with a glob pattern, e.g.
`/directoryname/*`, you'll receive the header plus all the rows plus the footer,
recreating the full pg dump. In this way, you can construct full or partial 
pg dump files so that you can load full or partial data sets.

```
$ pachctl put-file data master -c -f users.pgdump --split sql
$ pachctl put-file data master users --split sql -f users.pgdump 
$ pachctl list-file data master
NAME         TYPE SIZE 
users        dir  914B 
$ pachctl list-file data master /users
NAME                           TYPE SIZE 
/users/0000000000000000        file 20B  
/users/0000000000000001        file 18B  
```

Then in your pipeline (where you've started and forked postgres), you can load
the data by doing something like:

```
$ cat /pfs/data/users/* | sudo -u postgres psql
```

And with a glob pattern like `/*` this code would load the raw postgres chunk
into your postgres instance for processing by your pipeline.

For this use case, you'll likely want to use `--target-file-datums` or 
`--target-file-bytes` since it's likely that you'll want to run your queries
against a bunch of rows at a time.
