# Pachyderm File System CLI

## Terms
__repository:__ A repository (aka: repo) is a grouping of data that you want to track changes over. Repos in pfs behave just like repos in Git, except they work over huge datasets instead of just source code. You can take snapshots (commits) on a repo and multiple repos can exist in a single cluster. For example, if you have multiple production databases dumping data into pfs on a hourly or daily basis, it may make sense for each of those to be a separate repo. 

__commit:__ A commit is an immutable snapshot of a repo at a given time. Commits can either be open or closed. An open commit is writable, allowing you to add, change, or remove files. A closed or "committed" commit is read-only and cannot be changed. 

__file/directory:__ Files are the base unit of data in pfs. Files can be organized in directories, just like any normal file system. 


## Commands
#### help
    Usage: pfs help COMMAND
    
    Returns information about a pfs command

#### version
    Usage: pfs version
    
    Returns the currently running version of pfs

#### mount
    Usage: pfs mount REPOSITORY [MOUNTPOINT]
    
    MOUNTPOINT          The local directory used to access the pfs repo. 
                        Defaults to your working directory if unspecified.
    
    mounts a repo in the distributed file system onto the local mountpoint. 
    
    Return:

Once you mount a repo in pfs to a local mountpoint, any reads or writes pointed at the local mountpoint percolate to the repo in the distributed file system. 

##### Examples
    # mount the `user_data` repo in pfs to your working directory
    $pfs mount user_data

    # mount the `user_data` repo in pfs to your home directory
    $pfs mount user_data ~



### Commands that can be called on a repository:
#### init
    Usage: pfs init REPOSITORY
    
    Create a new repository
    
    Return: REPOSITORY
##### Examples
    # Create  the repository `user_data`
    $pfs init user_data
    user_data

#### list-commits
    Usage: pfs list-commits REPOSITORY
    
    Lists all commits in a repository
    
    Return format: ID  timestamp  status
##### Examples
    # List all commits in the repository `user_data`
    $pfs list-commits user_data
    ID        timestamp       status
    UUID1     00:00:00:00     writable
    UUID2     00:00:00:00     read-only

### Commands that can be called on a commit:
#### branch
#### commit
#### commit-info

### Commands that can be called on a file:
#### mkdir
#### ls
#### put
#### get
