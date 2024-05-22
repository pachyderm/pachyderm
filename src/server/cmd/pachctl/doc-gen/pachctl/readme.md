Want to edit or update some of our PachCTL CLI documentation? This is a quick guide to help you get started.

## List of commands categories and their locations:

- **taskcmds** /src/internal/task/cmds
- **admincmds** /src/server/admin/cmds
- **authcmds** /src/server/auth/cmds
- **configcmds** /src/server/config
- **debugcmds** /src/server/debug/cmds
- **enterprisecmds** /src/server/enterprise/cmds
- **identitycmds** /src/server/identity/cmds
- **licensecmds** /src/server/license/cmds
- **misccmds** /src/server/misc/cmds
- **pfscmds** /src/server/pfs/cmds
- **ppscmds** /src/server/pps/cmds
- **txncmds** /src/server/transaction/cmds

<!-- You can find quicklinks to these from /cmd/cmd.go in the import list. -->

## How to Edit 

1. Navigate to the corresponding `cmds.go` file for the command you want to edit.
2. Update the `Short` and `Long` descriptions for the command you want to edit.  Long descriptions can contain examples; this section is also valuable for our external documentation. 

<!-- Not sure where something lives? Find it in our [documentation](https://docs.pachyderm.com/latest/run-commands) and then search the repo for a matching short description. -->

## How to Test the Pachctl CLI's docs output

1. Navigate to `/src/server/cmd/pachctl` and run `go build`. 
2. Execute `./pachctl` to see the updated CLI and interact with it.
3. Explore other command outputs in the same way you normally would, for example: `./pachctl list`, `./pachctl put file`, etc. 

## How to Build the Pachctl CLI Markdown documentation

1. Navigate to `/src/server/cmd/pachctl-doc` and run `go build`. 
2. Execute `./pachctl-doc` to generate documentation in a ./docs folder
3. Move the contents of that folder to the docs-content repo's latest release (`/run-commands` directory).