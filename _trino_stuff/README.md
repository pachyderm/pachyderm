When working on this, I downloaded and extracted the Trino server to this directory, so that I could hack on it and run it. It's huge, so I couldn't check it in to git, but here's how to recreate my local state (if you want to continue development):

```
# Install trino client to _trino_stuff/bin
mkdir ./bin
curl -Lo bin/trino https://repo1.maven.org/maven2/io/trino/trino-cli/406/trino-cli-406-executable.jar
chmod +x bin/trino


# Install trino server to _trino_stuff/trino-server-406
curl -LO https://repo.maven.apache.org/maven2/io/trino/trino-server/406/trino-server-406.tar.gz
tar -xvzf ./trino-server-406.tar.gz


# To run the server:
/trino-server-406/bin/launcher run

# To run the client:
bin/trino --server=http://localhost:8080
```

## Example session:

```
$ bin/trino
trino> SHOW CATALOGS ;
 Catalog
---------
 pach
 system
...

trino> SHOW SCHEMAS FROM pach ;
   	Schema
--------------------
 example
 information_schema
 tpch
...

trino> SHOW TABLES FROM pach.tpch ;
  Table
----------
 lineitem
 orders
...

trino> SELECT COUNT(*) FROM pach.tpch.lineitem ;
 _col0
-------
   785
(1 row)

Query 20230208_214226_00034_zp3it, FINISHED, 1 node
Splits: 11 total, 11 done (100.00%)
1.46 [785 rows, 0B] [539 rows/s, 0B/s]
```
