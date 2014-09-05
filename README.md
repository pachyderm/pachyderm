Pachyderm File System Daemon

This program runs an http server that servers the pfsd http api which works as
follows:

# init

Initializing a new `filesystem`:

curl localhost:5656/`filesystem` -XPUT -d cmd="init"

# cat

Read `file` from `filesystem`:

curl localhost:5656/`filesystem`/`file`

Writing `local_file` to `file` in `filesystem`:

curl localhost:5656/`filesystem`/`file` -XPUT -d @`local_file` 

Create `commit` on `filesystem`:

curl localhost:5656/`filesystem`/`file` -XPUT -d cmd="commit"

Read `file` from `filesystem` at `commit`:

curl localhost:5656/`filesystem`@`commit`/`file`

Create `branch` of `filesystem` at `commit`:

curl localhost:5656/`filesystem`@`commit` -XPUT -d cmd="branch" -d branch=`branch`


