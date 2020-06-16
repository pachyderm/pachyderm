# Mount a Repo to a Local Computer

!!! warninig
    Because Pachyderm uses FUSE to mount a repo and
    macOS does not fully support FUSE, you might
    not be able to use this functionality on macOS.

Pachyderm enables you to mount a repository
as a local filesystem on your computer by using the
`pachctl mount` command. This command
uses Filesystem in Userspace (FUSE) user interface to export a Pachyderm
File System (PFS) to a Unix computer system.
This functionality is useful when you want to expose the results
of your pipeline for further consumption or modify the files
in the input repository directly.

You can mount a Pachyderm repo in one of the following modes:

* `Read-only` — the users can view files in all
repositories, but cannot modify them.
* `Read-write` — the users can view files
in all repositories and modify the files in input repositories.

## Mounting Repositories in Read-Only Mode

By default, when you run the `pachctl mount` command, Pachyderm mounts
all Pachyderm repositories in read-only mode. You can access the
files through your file browser or enable third-party applications
access. Read-only access enables you to explore and experiment with
the data, without modifying it. For example, you can mount your
repo to a local computer and then open that directory in a Jupyter
Notebook for exploration. 

## Mounting Specific Branches and Commits

The `pachctl mount` command allows you to mount not only the default
branch, typically a `master` branch, but also other Pachyderm
branches. By default, Pachyderm mounts the `master` branch. However,
if you add a branch to the name of the repo, the `HEAD` of that branch
will be mounted.

**Example:**

```bash
pachctl mount images --repos images@staging+w
```

You need to add the `+w` to the branch to enable write access.

Also, you can mount a specific commit. Because commits
might be on multiple branches, modifying them might result in data deletion
in the `HEAD` of the branches. That is why you can only mount commits in
read-only mode.

## Mounting Repositories in Read-Write Mode

Running the `pachctl mount` command with the `--write` flag grants you
write access to the mounted repositories, which means that you can
open the files for editing and putting them back to the Pachyderm
repository. 

For example, you have an [OpenCV pipeline](../../../getting_started/beginner_tutorial/#image-processing-with-opencv)
up and running,
and you need to edit files in the `images` repository. Let's
say you want to experiment with brightness and contrast
settings in `liberty.png`. Then, you want the pipelines
run for the updated file and see the result in the end.
If you do not mount the `images` repo, you would have to
first download the files to your computer, edit them,
and then put them back to the repository. The `pachctl mount`
command automates all these steps for you. You can mount just the
`images` repo or all Pachyderm repositories as directories
on you machine, edit as needed, and, when done,
exit the `pachctl mount` command. Upon exiting the `pachctl mount`
command, Pachyderm uploads all the changes to the corresponding
repository.

If someone else modifies the files while you are working on them
locally, their changes will likely be overwritten when you exit
`pachctl mount`. This happens because your changes are saved to
the Pachyderm repository only after you interrupt the `pachctl mount`
command. Therefore, make sure that you do not work on the
same files while someone else is working on them.

!!! warning
    Use writable mount **ONLY** when you have sole ownership
    over the mounted data. Otherwise, merge conflicts or
    unexpected data overwrites can occur.

Because output repositories are created by the Pachyderm
pipelines, they are immutable. Only a pipeline
can change and update files in these repositories. If you try to change
a file in an output repo, you will get an error message.

## Prerequisites

You must have the following configured for this functionality to work:

* Unix or Unix-like operating system, such as Ubuntu 16.04 or macOS
Yosemite or later.
* FUSE for your operating system installed:

  * On macOS, run:

    ```bash
    brew cask install osxfuse
    ```

    ```bash
    brew install sshfs
    ```

  * On Ubuntu, run:

    ```bash
    sudo apt-get install sshfs
    ```

    For more information, see:

    * [FUSE for macOS](https://osxfuse.github.io/)
    * [FUSE on Ubuntu](https://gist.github.com/cstroe/e83681e3510b43e3f618)

  !!! note
      macOS has limited support for FUSE and might not be stable.

## Mount a Pachyderm Repo

Before you can mount a Pachyderm repo, verify that you have all the
[Prerequisites](#prerequisites).

To mount a Pachyderm repo on a local computer, complete the following
steps:

1. In a terminal, go to a directory in which you want to mount a
Pachyderm repo. It can be any new empty directory on your local computer.
For example, `pfs`.

1. Run `pachctl mount` for a repository and branch that you want to mount:

   ```bash
   pachctl mount <path-on-your-computer> [flags]
   ```

   **Example:**

   * If you want to mount all the repositories in your Pachyderm cluster 
   to a `pfs` directory on your computer and give `WRITE` access to them, run:

   ```bash
   pachctl mount pfs --write
   ```

   * If you want to mount the master branch of the `images` repo
   and enable file editing in this repository, run:

   ```bash
   pachctl mount images --repos images@master+w
   ```

   To give read-only access, omit `+w`.

   **System Response:**

   ```bash
   ro for images: &{Branch:master Write:true}
   ri: repo:<name:"montage" > created:<seconds:1591812554 nanos:348079652 > size_bytes:1345398 description:"Output repo for pipeline montage." branches:<repo:<name:"montage" > name:"master" >
   continue
   ri: repo:<name:"edges" > created:<seconds:1591812554 nanos:201592492 > size_bytes:136795 description:"Output repo for pipeline edges." branches:<repo:<name:"edges" > name:"master" >
   continue
   ri: repo:<name:"images" > created:<seconds:1591812554 nanos:28450609 > size_bytes:244068 branches:<repo:<name:"images" > name:"master" >
   MkdirAll /var/folders/jl/mm3wrxqd75l9r1_d0zktphdw0000gn/T/pfs201409498/images
   ```

   The command runs in your terminal until you terminate it
   by pressing `CTRL+C`.

1. You can check that the repo was mounted by running the mount command
in your terminal:

   ```bash hl_lines="7"
   mount
   /dev/disk1s1 on / (apfs, local, read-only, journaled)
   devfs on /dev (devfs, local, nobrowse)
   /dev/disk1s2 on /System/Volumes/Data (apfs, local, journaled, nobrowse)
   /dev/disk1s5 on /private/var/vm (apfs, local, journaled, nobrowse)
   map auto_home on /System/Volumes/Data/home (autofs, automounted, nobrowse)
   pachctl@osxfuse0 on /Users/testuser/pachyderm/images (osxfuse, nodev, nosuid, synchronous, mounted by testuser)
   ```

1. Access your mountpoint.

   For example, in macOS, open Finder, press
   `CMD + SHIFT + G`, and type the mountpoint location. If you have mounted
   the repo to `~/pachyderm/images`, type `~/pachyderm/images`.

   ![finder-repo-mount](../../assets/images/s_finder_repo_mount.png)

1. Edit the files as needed.
1. When ready, add your changes to the Pachyderm repo by stopping
the `pachctl mount` command with `CTRL+C` or by running `pachctl unmount
<mountpoint>`.

   If you have mounted a writable Pachyderm share interrupting the
   `pachctl mount` command results in uploads
   your changes to the corresponding repo and branch, which is equivalent
   to running the `pachctl put file` command. You can check that
   Pachyderm runs a new job for this work by listing current jobs with
   `pachctl list job`.
