# Using the Pachyderm Shell

The Pachyderm Shell is a special-purpose shell for Pachyderm that provides
auto-suggesting as you type. New Pachyderm users will find this user-friendly
shell especially appealing as it helps to learn `pachctl`, type commands
faster, and displays useful information about the objects you are interacting
with. This new shell does not supersede the classic use of `pachctl` shell
in your standard terminal, but is a compelling convenience for power users
and beginners alike. If you prefer to use just `pachctl`, you can continue to
do so.

To enter the Pachyderm Shell, type:

```shell
pachctl shell
```

When you enter `pachctl` shell, your prompt changes to display your current
Pachyderm context, as well as displays a list of available commands in a
drop-down list.

![Pachyderm Shell](../../assets/images/s_pach_shell.png)

To scroll through the list, press `TAB` and then use arrows to move up or
down. Press `SPACE` to select a command.

When in the Pachyderm Shell, you do not need to prepend your commands with
`pachctl` because Pachyderm does that for you automatically behind the
scenes. For example, instead of running `pachctl list repo`, run `list
repo`:

![Pachyderm Shell list repo](../../assets/images/s_pach_shell_list_repo.png)

With nested commands, `pachctl shell` can do even more. For example, if you
type `list file <repo>@<branch>/`, you can preview and select files from that
branch:

![Pachyderm Shell list file](../../assets/images/s_pach_shell_list_file.png)

Similarly, you can select a commit:

![Pachyderm Shell list commit](../../assets/images/s_pach_shell_list_commit.png)

### Exit the Pachyderm Shell

To exit the Pachyderm Shell, press `CTRL-D` or type `exit`.

### Clearing Cached Completions

To optimize performance and achieve faster response time,
the Pachyderm Shell caches completion results. You can clear this cache
by pressing **F5** forcing the Pachyderm Shell to send requests to the
server for new completions.


## Limitations

The Pachyderm Shell does not support standard UNIX commands or `kubectl` commands.
To run them, exit the Pachyderm Shell or run the commands in a different terminal
window.
