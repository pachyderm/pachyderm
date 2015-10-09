#TODO:
- mount: better decription, add flags and lots of examples. Talk about umount and how other pfs commands do/don't work the same after mount.
- add more details to put/get? Any specifics or nuances that people could get stuck on. 
- add to error codes section

# Pachyderm File System CLI

* [Terms](#terms)
* [Error Codes](#error-codes)
* [Commands](#commands)
    * [help] (#help)
    * [version] (#version)
    * [mount] (#mount)
    * [list-repos] (#list-repos)
    * [create-repo] (#create-repo)
    * [inspect-repo] (#inspect-repo)
    * [delete-repo] (#delete-repo)
    * [list-commits] (#list-commits)
    * [start-commit] (#start-commit)
    * [finish-commit] (#finish-commit)
    * [inspect-commit] (#inspect-commit)
    * [mkdir] (#mkdir)
    * [list-files] (#list-files)
    * [put-file] (#put-file)
    * [get-file] (#get-file)
    * [inspect-file] (#inspect-file)
    * [delete-file] (#delete-file)

## Terms
__repository:__ A repository (aka: repo) is a set of data over which you want to track changes. Repos in pfs behave just like repos in Git, except they work over huge datasets instead of just source code. You can take snapshots (commits) on a repo and multiple repos can exist in a single cluster. For example, if you have multiple production databases dumping data into pfs on a hourly or daily basis, it may make sense for each of those to be a separate repo. 

__commit:__ A commit is an immutable snapshot of a repo at a given time. Commits can be in two states, started or finished. Starting a commit creates a new commit that is writable, meaning you can add, modify, or remove files. Finishing a commit turns a writable commit into an immutable read-only state. Finished commits are fully replicated, but writable commits are considered in a "dirty" state and are not replicated until they are finished. Commits reference each other in a tree structure. 

__file/directory:__ Files are the base unit of data in pfs. Files can be organized in directories, just like any normal file system. 

## Error codes

## Commands
#### help
    Usage: pfs help COMMAND
    
    Returns information about a pfs command
    
##### Example
    # Learn more about the `get` command
    $ pfs help get
    Get a file from stdout. commit-id must be a readable commit.
    Usage:
      pfs get repository-name commit-id path

#### version
    Usage: pfs version
    
    Returns the currently running version of pfs
    
##### Example
    # Find out the running version of Pachyderm
    $pfs version
    Client: v0.9
    Server: v0.9

#### mount
    Usage: pfs mount MOUNTPOINT REPOSITORY [COMMIT_ID] [OPTIONS]
    
    MOUNTPOINT          The local directory used to access the pfs repo. 
                        Defaults to your working directory if unspecified.
    COMMIT_ID           Only mount the data for a specific commit
    -s, --shard=0       
    -m, --modulus=1     
    
    Mounts a repo in the distributed file system onto the local mountpoint. 
    
Mounting a repo in pfs lets you access its contents as if it was a local file system. Any reads or writes pointed at the local mountpoint are applied to the repo in the distributed file system. 

##### Example
    # mount the `repo` repository in pfs to your working directory
    $ pfs mount repo

    # mount the `repo` repo in pfs to your home directory
    $ pfs mount repo ~

#### list-repos
Alias: lr

    Usage: pfs list-repos
    
    Lists all repositories in pfs
    
    Return format: NAME  TIME_CREATED  NUM_COMMITS  TOTAL_SIZE
    
##### Example
    # List all repositories in pfs
    $ pfs list-repos
    NAME    TIME_CREATED        NUM_COMMITS     TOTAL_SIZE    
    repo    about a year ago    520187          468.4 TB
    repo2   106 days ago        129840          977.9 GB     

### Commands that can be called on a repository:
#### create-repo
Alias: cr

    Usage: pfs create-repo REPOSITORY
    
    Create a new repository
Repository names cannot contain `/`.


##### Example
    # Create the repository `repo`
    $ pfs create-repo repo

#### inspect-repo
Alias: ir

    Usage: pfs inspect-repo REPOSITORY
    
    Returns additional information about a repository
    
    Return format: NAME  TIME_CREATED  NUM_COMMITS  TOTAL_SIZE
##### Example
    # List all repositories in pfs
    $ pfs inspect-repo
    NAME    TIME_CREATED        NUM_COMMITS     TOTAL_SIZE    
    repo    about a year ago    520187          468.4 TB

#### delete-repo
Alias: dr

    Usage: pfs delete-repo REPOSITORY
    
    Deletes a repository including all its commits and data

##### Example
    # Delete the repository `repo`
    $ pfs delete-repo repo
    
#### list-commits
Alias: lc

    Usage: pfs list-commits REPOSITORY
    
    Lists all commits in a repository
    
    Return format: ID  PARENT  STATUS  TIME_OPENED  TIME_CLOSED  TOTAL_SIZE  DIFF_SIZE
##### Example
    # List all commits in the repository `repo`
    $ pfs list-commits repo
    ID      PARENT      STATUS      TIME_OPENED         TIME_CLOSED     TOTAL_SIZE    DIFF_SIZE
    ID_2    ID_1        writable    about an hour ago                   801.2 GB      100 MB
    ID_1    scratch     read-only   2 hours ago         2 hours ago     801.1 GB      801.1 GB
    

### Commands that can be called on a commit:
#### start-commit
Alias: sc

    Usage: pfs start-commit REPOSITORY [PARENT_COMMIT]
    
    PARENT-COMMIT       Commit will have no parent if left unspecified  
                    
    Creates a writable commit with the specified parent. 
    
    Return: COMMIT_ID
    
Commits start out containing exactly the contents of the parent commit and then record changes.
    
   
##### Example
    # Create a new, writable commit with the parent commit `ID_1` in the `repo` repository
    $ pfs start-commit repo ID_1 
    ID_2
    
#### finish-commit
Alias: fc

    Usage: pfs finish-commit REPOSITORY COMMIT_ID
    
    Turns a writable commit into a read-only commit
    
##### Example
    # Make the writable commit `ID_2` read-only in the repository `repo`
    $ pfs finish-commit repo ID_2

#### inspect-commit
Alias: ic

    Usage: pfs inspect-commit REPOSITORY COMMIT_ID
    
    Returns metadata about the specified commit
    
    Return format:  ID  PARENT  STATUS  TIME_OPENED  TIME_CLOSED  TOTAL_SIZE  DIFF_SIZE

##### Example
    # Get infomation about commit `ID_2` in repository `repo`
    $ pfs inspect-commit repo ID_2
    ID      PARENT      STATUS      TIME_OPENED         TIME_CLOSED     TOTAL_SIZE    DIFF_SIZE
    ID_2    ID_1        writable    about an hour ago                   801.2 GB      100 MB   

### Commands that can be called on a file or directory:
#### mkdir
Alias: md

    Usage: pfs mkdir REPOSITORY COMMIT_ID PATH
    
    Creates a new empty directory

Use `/` to add heirarchy to directories. Works recursively (see example). You can only create a new directory on a writable commit.

##### Example
    # Create directory `bar` inside directory `foo` on commit `ID_2` in repository `repo`.
    # `foo` need not exist. If it doesn't, it will also be created. 
    $ pfs mkdir repo ID_2 foo/bar
    
#### list-files
Alias: ls, lf

    Usage: pfs ls REPOSITORY COMMIT_ID [PATH]
    
        PATH       If unspecified, will return all files and directories recursively.
                    
    Lists all of the files and directories in the specified path. The PATH `/` will return
    the top-level directories and files without recursing. Similar to the `ls` shell command.
    
    Return format: NAME  TYPE  MODIFIED  CREATED  LAST_COMMIT_MODIFIED  SIZE  PERMISSIONS

We currently do not have a recurse flag for specifed paths, but will be adding it in the near future.
##### Example
    # List all files and directories in the directory `foo`, commit `ID_2`, repository `repo`
    $ pfs ls repo ID_2 foo
    NAME    TYPE    MODIFIED        LAST_COMMIT_MODIFIED    SIZE        PERMISSIONS
    bar     dir     35 minutes ago  ID_1                    4K          448
    file1   file    35 minutes ago  ID_1                    5.1 MB      420 
    file2   file    35 minutes ago  ID_1                    14.8 GB     420
    
#### put-file
Alias: pf, put
    Usage: pfs put-file REPOSITORY COMMIT_ID PATH
    
    Adds a file into pfs. Reads contents from stdin.

Including a `/` in the file name will create directories as needed. Files can only be added to writable commits. 

##### Example
    # Add the contents of `local_file` to pfs. Name it `file1` in commit `ID_2` and repository `repo`
    $ pfs put-file repo ID_2 file1 <local_file
    
    # Create the directory `foo` and dump the contents of a Postgres database into the file `dump`
    $ pg_dump database | pfs put-file repo ID_2 foo/dump

#### get-file
Alias: gf, get
    Usage: pfs get-file REPOSITORY COMMIT_ID PATH
    
    Reads a file out of pfs and outputs the content to stdout
    
    Return: file contents

##### Example
    # Read the file `file1` from commit `ID_2` in the repository `repo`
    $ pfs get-file repo ID_2 file1
    <contents of file1>

#### inspect-file
Alias: if
    Usage: pfs inspect-file REPOSITORY COMMIT_ID PATH
    
    Returns metadata about the specified file
    
    Return format: NAME  TYPE  MODIFIED  LAST_COMMIT_MODIFIED  SIZE  PERMISSIONS

##### Example
    # Get information about the file `file1` from commit `ID_2` in the repository `repo`
    $ pfs inspect-file repo ID_2 file1
    NAME    TYPE    MODIFIED        LAST_COMMIT_MODIFIED    SIZE        PERMISSIONS
    file1   file    35 minutes ago  ID_1                    5.1 MB      420 

#### delete-file
Alias: df

    Usage: pfs delete-file REPOSITORY COMMIT_ID PATH
    
    Deletes a file in pfs
    
##### Example
    # Delete the file `file1` from commit `ID_2` in the repository `repo`
    $ pfs delete-file repo ID_2 file1
