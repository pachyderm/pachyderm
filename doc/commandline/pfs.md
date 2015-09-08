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
    
    Mounts a repo in the distributed file system onto the local mountpoint. 
    
    Return:

Once you mount a repo in pfs to a local mountpoint, any reads or writes pointed at the local mountpoint percolate to the repo in the distributed file system. 

##### Example
    # mount the `user_data` repo in pfs to your working directory
    $pfs mount user_data

    # mount the `user_data` repo in pfs to your home directory
    $pfs mount user_data ~



### Commands that can be called on a repository:
#### init
    Usage: pfs init REPOSITORY
    
    Create a new repository
    
    Return: REPOSITORY
##### Example
    # Create  the repository `user_data`
    $pfs init user_data
    user_data

#### list-commits
    Usage: pfs list-commits REPOSITORY
    
    Lists all commits in a repository
    
    Return format: Commit_ID
##### Example
    # List all commits in the repository `user_data`
    $pfs list-commits user_data
    Commit_ID
    UUID1
    UUID2

### Commands that can be called on a commit:
#### branch
    Usage: pfs branch REPOSITORY [PARENT_COMMIT]
    
    PARENT-COMMIT       The local directory used to access the pfs repo. 
                        Commit will have no parent if left unspecified
                        
    Creates an open commit with the specified parent
    
    Returns: COMMIT_ID
##### Example
    # Create a new, writable commit with the parent commit `UUID1` in the `user_data` repository
    $pfs branch user_data UUID1 
    UUID2
#### commit
    Usage: pfs commit REPOSITORY COMMIT_ID
    
    Turns an writable commit into a read-only commit
    
    Returns: COMMIT_ID
##### Example
    # Make the writable commit `UUID2` into read-only in the repository `user_data`
    $pfs commit user_data UUID2
    UUID2
#### commit-info
    Usage: pfs commit-info REPOSITORY COMMIT_ID
    
    Returns metadata about the specified commit
    
    Return format: ID  timestamp  status  parent  size  ETC,ETC,ETC?
##### Example
    # Get infomation about commit `UUID2` in repository `user_data`
    $pfs commit-info user_data UUID2
    ID        timestamp       status        parent      diff-size
    UUID2     00:00:00:00     read-only     UUID1       81.2MB

### Commands that can be called on a file or directory:
#### mkdir
    Usage: pfs mkdir REPOSITORY COMMIT_ID DIRECTORY_PATH
    
    Creates a new empty directory

Use `/` to add heirarchy to directories. Works recursively. You can only create a new directory on a writable commit.

##### Example
    # Create directory `bar` inside directory `foo` on commit `UUID2` in repository `user_data`
    $pfs mkdir user_data UUID2 foo/bar
    
#### ls
    Usage: pfs ls REPOSITORY COMMIT_ID DIRECTORY_PATH
    
    Lists all of the files and directories in the specified path. Works just like the `ls` command in shell. 

##### Example
    # List all files and directories in the directory `foo`, commit `UUID2`, repository `user_data`
    $pfs ls user_data UUID2 foo
    name        type
    bar         directory
    file1       file
    file2       file
    
#### put 
    Usage: FILE | pfs ls REPOSITORY COMMIT_ID FILE_PATH
    
    Adds a file into pfs
    
    Return: FILE_PATH

Including a `/` in the file name will create the directory and add the file. Files can only be added to writable commits. 

##### Example
    # Add the file `file1` to commit `UUID2` in the repository `user_data`
    $ local_file | pfs put user_data UUID2 file1
    file1
    
    # Create the directory `foo` and add the file `file1` to that directory.
    $ local_file | pfs put user_data UUID2 foo/file1
    foo/file1

#### get
    Usage: pfs get REPOSITORY COMMIT_ID FILE_PATH
    
    Reads a file out of pfs
    
    Return: file contents

##### Example
    # Get the file `file1` from commit `UUID2` in the repository `user_data`
    $ pfs get user_data UUID2 file1
    file1
    
#### file-info
    Usage: pfs file-info REPOSITORY COMMIT_ID FILE_PATH
    
    Returns metadata about the specified file
    
    Return format: name  size  date_added  commit_added

##### Example
    # Get the file `file1` from commit `UUID2` in the repository `user_data`
    $ pfs file-info user_data UUID2 file1
    name        size        date_added      commit_added
    file1       3.1MB       1/1/2015        UUID1
