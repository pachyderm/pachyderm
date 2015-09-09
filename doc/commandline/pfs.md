#TODO:
- commit definition
- mount: add flags and lots of examples
- add to put/get?
- talk about standard unix error codes, 0, 1
- user_data --> repo
- UUID--> ID

# Pachyderm File System CLI

## Terms
__repository:__ A repository (aka: repo) is a set of data over which you want to track changes. Repos in pfs behave just like repos in Git, except they work over huge datasets instead of just source code. You can take snapshots (commits) on a repo and multiple repos can exist in a single cluster. For example, if you have multiple production databases dumping data into pfs on a hourly or daily basis, it may make sense for each of those to be a separate repo. 

__commit:__ A commit is an immutable snapshot of a repo at a given time. Commits can either be open or closed. An open commit is writable, allowing you to add, change, or remove files. A closed or "committed" commit is read-only and cannot be changed. 

commits reference each other in dag...

__file/directory:__ Files are the base unit of data in pfs. Files can be organized in directories, just like any normal file system. 


## Commands
#### help
    Usage: pfs help COMMAND
    
    Returns information about a pfs command
    
##### Example
    # Learn more about the `get` command
    $ pfs help get
    Get a file from stdout. commit-id must be a readable commit.
    Usage:
      pfs get repository-name commit-id path/to/file

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
    # mount the `user_data` repo in pfs to your working directory
    $ pfs mount user_data

    # mount the `user_data` repo in pfs to your home directory
    $ pfs mount user_data ~



### Commands that can be called on a repository:
#### init
    Usage: pfs init REPOSITORY
    
    Create a new repository
Repository names cannot contain `/`.

##### Example
    # Create the repository `user_data`
    $ pfs init user_data

#### list-commits
    Usage: pfs list-commits REPOSITORY
    
    Lists all commits in a repository
    
    Return format: ID  PARENT  STATUS  TIME_OPENED  TIME_CLOSED  TOTAL_SIZE  DIFF_SIZE
##### Example
    # List all commits in the repository `user_data`
    $ pfs list-commits user_data
    ID      PARENT      STATUS      TIME_OPENED         TIME_CLOSED     TOTAL_SIZE    DIFF_SIZE
    ID_2    ID_1        writable    about an hour ago                   801.2 GB      100 MB   
    ID_1    scratch     read-only   2 hours ago         2 hours ago     801.1 GB      801.1 GB                     
    

### Commands that can be called on a commit:
#### branch
    Usage: pfs branch REPOSITORY [PARENT_COMMIT]
    
    PARENT-COMMIT       Commit will have no parent if left unspecified  
                    
    Creates an open commit with the specified parent. 
    
    Return: COMMIT_ID
    
Commits start out containing exactly the contents of the parent commit and then record changes.
    
   
##### Example
    # Create a new, writable commit with the parent commit `ID_1` in the `user_data` repository
    $ pfs branch user_data ID_1 
    ID_2
    
#### commit
    Usage: pfs commit REPOSITORY COMMIT_ID
    
    Turns a writable commit into a read-only commit
    
##### Example
    # Make the writable commit `ID_2` into read-only in the repository `user_data`
    $ pfs commit user_data ID_2

#### commit-info
    Usage: pfs commit-info REPOSITORY COMMIT_ID
    
    Returns metadata about the specified commit
    
    Return format:  ID  PARENT  STATUS  TIME_OPENED  TIME_CLOSED  TOTAL_SIZE  DIFF_SIZE

##### Example
    # Get infomation about commit `ID_2` in repository `user_data`
    $ pfs commit-info user_data ID_2
    ID      PARENT      STATUS      TIME_OPENED         TIME_CLOSED     TOTAL_SIZE    DIFF_SIZE
    ID_2    ID_1        writable    about an hour ago                   801.2 GB      100 MB   

### Commands that can be called on a file or directory:
#### mkdir
    Usage: pfs mkdir REPOSITORY COMMIT_ID DIRECTORY_PATH
    
    Creates a new empty directory

Use `/` to add heirarchy to directories. Works recursively (see example). You can only create a new directory on a writable commit.

##### Example
    # Create directory `bar` inside directory `foo` on commit `ID_2` in repository `user_data`.
    # `foo` need not exist. If it doesn't, it will also be created. 
    $ pfs mkdir user_data ID_2 foo/bar
    
#### ls
    Usage: pfs ls REPOSITORY COMMIT_ID [DIRECTORY_PATH]
    
        DIRECTORY_PATH       Defaults to `/`, the root of the file system  
                    
    Lists all of the files and directories in the specified path. Similar to the `ls` command in shell. 
    
    Return format: NAME  TYPE  MODIFIED  CREATED  LAST_COMMIT_MODIFIED  SIZE  PERMISSIONS

##### Example
    # List all files and directories in the directory `foo`, commit `ID_2`, repository `user_data`
    $ pfs ls user_data ID_2 foo
    NAME    TYPE    MODIFIED        LAST_COMMIT_MODIFIED    SIZE        PERMISSIONS
    bar     dir     35 minutes ago  ID_1                    4K          448
    file1   file    35 minutes ago  ID_1                    5.1 MB      420 
    file2   file    35 minutes ago  ID_1                    14.8 GB     420
    
#### put 
    Usage: pfs put REPOSITORY COMMIT_ID FILE_PATH
    
    Adds a file into pfs. Reads contents from stdin.

Including a `/` in the file name will create directories as needed. Files can only be added to writable commits. 

##### Example
    # Add the contents of `local_file` to pfs. Name it `file1` in commit `ID_2` and repository `repo`
    $ pfs put repo ID_2 file1 <local_file
    
    # Create the directory `foo` and dump the contents of a Postgres database into the file `dump`
    $ pg_dump database | pfs put repo ID_2 foo/dump

#### get
    Usage: pfs get REPOSITORY COMMIT_ID FILE_PATH
    
    Reads a file out of pfs and outputs the content to stdout
    
    Return: file contents

##### Example
    # Get the file `file1` from commit `ID_2` in the repository `repo`
    $ pfs get repo ID_2 file1
    <contents of file1>
    
#### file-info
    Usage: pfs file-info REPOSITORY COMMIT_ID FILE_PATH
    
    Returns metadata about the specified file
    
    Return format: NAME  TYPE  MODIFIED  LAST_COMMIT_MODIFIED  SIZE  PERMISSIONS


##### Example
    # Get the file `file1` from commit `UUID2` in the repository `user_data`
    $ pfs file-info user_data UUID2 file1
    NAME    TYPE    MODIFIED        LAST_COMMIT_MODIFIED    SIZE        PERMISSIONS
    file1   file    35 minutes ago  ID_1                    5.1 MB      420 
