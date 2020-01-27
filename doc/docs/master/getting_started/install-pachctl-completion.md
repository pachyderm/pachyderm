# Install `pachctl` Autocompletion

Pachyderm autocompletion allows you to automatically finish
partially typed commands by pressing `TAB`. Autocompletion needs
to be installed separately when `pachctl` is already
available on your client machine.

Pachyderm autocompletion is supported for both `bash` and `zsh`.
You must have either of them preinstalled
before installing Pachyderm autocompletion.

!!! tip
    Type `pachctl completion --help` to display help information about
    the command.

To install `pachctl` autocompletion, perform the following steps:

1. Verify that `bash` or `zsh` completion is installed on your machine.
   For example, if you have installed bash completion by using Homebrew,
   type:

   ```bash
   brew info bash-completion
   ```

   This command returns information about the directory in which
   `bash-completion` and bash completion scripts are installed.
   For example,  `/usr/local/etc/bash_completion.d/`. Unless it is
   the default `/etc/bash_completion.d/` location, you need to specify
   the path to `bash_completion.d`. Also, the output of the info
   command, might have a suggestion to include the path to
   `bash-completion` into your `~/.bash_profile` file.

1. Install `pachctl` autocompletion:

   * If you are using `bash`, run the following command:

     ```bash
     pachctl completion bash --install --path <path-to-bash>
     ```

   For example, if you specify the path to `bash-completion` as
   `/usr/local/etc/bash_completion.d/pachctl`, your system response
   looks like this:

   **System response:**

   ```bash
   Bash completions installed in /usr/local/etc/bash_completion.d/pachctl, you must restart bash to enable completions.
   ```

   * If you are using `zsh`, run the following command:

     ```sh
     pachctl completion zsh --install
     ```

     **System response:**

     ```sh
     Completions installed in "_pachctl", you must restart your terminal to enable them.
     ```

1. Restart your terminal.

   `pachctl` autocomplete should now be enabled in your system.

!!! note "See Also"

    [Pachyderm Shell](TBA)
