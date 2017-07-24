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
