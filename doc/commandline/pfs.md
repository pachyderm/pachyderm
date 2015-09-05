## Pachyderm File System CLI

### Terms
__repository:__ A repository (aka: repo) is a grouping of data that you want to track changes over. Repos in pfs behave just like repos in Git, except they work over huge datasets instead of just source code. You can take snapshots (commits) on a repo. Multiple repos can exist in a single cluster, for example, if you have multiple production databases dumping data into pfs, it may make sense for each of those to be separate repos. 

__commit:__ A commit is an immutable snapshot of a repo at a given time. Commits can either be open or closed. An open commit, also called "dirty", is writable so you can add, change or remove files. A closed, or "committed", commit is read-only and cannot be changed. 

__file/directory:__ Files are the base unit of data in pfs. Files can be organized in directories, just like any normal file system. 



### Commands
##### version
##### mount
##### help

#### Commands that can be called on a repository:
##### init
##### list-commits

#### Commands that can be called on a commit:
##### branch
##### commit
##### commit-info

#### Commands that can be called on a file:
##### mkdir
##### ls
##### put
##### get
