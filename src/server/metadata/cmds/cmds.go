package cmds

import (
	"fmt"
	"os"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/kvparse"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachctl"
	"github.com/pachyderm/pachyderm/v2/src/metadata"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protopath"
	"google.golang.org/protobuf/reflect/protorange"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Cmds returns a slice containing metadata commands.
func Cmds(pachCtx *config.Context, pachctlCfg *pachctl.Config) []*cobra.Command {
	var commands []*cobra.Command

	editMetadata := &cobra.Command{
		Use:   `{{alias}} [<object type: cluster|project|repo|branch|commit> <object picker> <operation: add|edit|delete|replace> <data: key=value, key, '{"key":"value","key2":"value2"}'>]...`,
		Short: "Edits an object's metadata",
		Long: `Edits an object's metadata.

Each metadata change is a sequence of 4 arguments, the object type, the object, the operation, and
the edit to perform.  A command-line may contain zero or more of these changes.  One command-line is
run as one transaction; either all of the changes succeed, or all of the changes fail.

The objects available are:

  * cluster: Adjust metadata on the cluster itself.  The metadata is visible on an "InspectCluster" RPC, or
    "pachctl inspect cluster --raw".  The "object picker" is ignored, as there is only one cluster.
    By convention, "." is used as the object picker.

  * project: Adjust metadata on a project.  The metadata is visible on an "InspectProject" RPC, or
    "pachctl inspect project <name> --raw".

  * repo: Adjust metadata on a repo.  The metadata is visible on an "InspectRepo" RPC, or
    "pachctl inspect repo <name> --raw".

  * branch: Adjust metadata on a branch.  The metadata is visible on an "InspectBranch" RPC, or
    "pachctl inspect branch <repo>@<name> --raw".

  * commit: Adjust metadata on a commit (not a commitset).  The metadata is visible on an "InspectCommit" RPC,
    or "pachctl inspect commit <repo>@<commit> --raw".

Object pickers are used throughout Pachyderm.  A summary:

  * cluster: Ignored, any string matches the current cluster.

  * project: The name of the project, for example, "default".

  * repo: The project and name of the repo, like "default/images".  If the project is omitted, the current project
    is assumed.  Special repo types may be selected, like "default/edges.meta" for the meta repo associated with
    the pipeline "edges".

  * branch: A repo picker, with @ and a branch name, like "default/images@master".  If the project is omitted, the
    current project is assumed.

  * commit: Like a branch picker, but identifying a commit instead.  For example, "default/images@master", or
    "default/images@fc337b93997e4bc7b167222cb2482460".

The operations available are:

  * add: Add a new key/value pair to the metadata.  The operation fails if the key already exists.
    Add takes a single key/value pair, in key=value syntax.

  * edit: Change the value of a key.  It's OK if they key doesn't already exist.
    Edit takes a single key/value pair, in key=value syntax.

  * delete: Remove a key/value pair.  Delete takes the name of the key to delete.

  * replace: Replace the entire metadata structure with a new set of key/value pairs.
    Replace takes zero or more key/value pairs.  This can be represented as JSON, like
    {"key1":"value1","key2":"value2"}, or as our custom format,
    "key1=value1;key2=value2".

Key/value pair format:

  A single key-value pair is represented as "key=value".  Key and value can be double-quoted or single-quoted.
  Standard escapes apply inside double-quoted strings.  No escapes are allowed in single-quoted strings.
  For example, "two lines"="one\ntwo", 'key with a \'=value.

  Multiple key-value pairs follow the same format and are separated with a semicolon.  For example,
  "key1"='value1';key2=value2;'key3'="value3".  You may also supply a normal JSON object.  Keys must
  be strings, and values must be strings.  No other JSON syntax is allowed.
`,
		Example: "\t- {{alias}} edit metadata cluster . add environment=production \n" +
			"\t - {{alias}} edit metadata project myproject edit support_contact=you@example.com \n" +
			"\t - {{alias}} edit metadata commit images@master add verified_by=you@example.com \\ \n" +
			"\t \t \t commit edges@master add verified_by=you@example.com",
		Run: cmdutil.Run(func(cmd *cobra.Command, args []string) error {
			req, err := parseEditMetadataCmdline(args, pachCtx.Project)
			if err != nil {
				return errors.Wrap(err, "parse cmdline")
			}

			c, err := pachctlCfg.NewOnUserMachine(cmd.Context(), false)
			if err != nil {
				return err
			}
			defer c.Close()
			if _, err := c.MetadataClient.EditMetadata(c.Ctx(), req); err != nil {
				return errors.Wrap(err, "invoke EditMetadata")
			}
			fmt.Fprintf(os.Stderr, "ok\n")
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(editMetadata, "edit metadata"))

	return commands
}

func parseEditMetadataCmdline(args []string, defaultProject string) (*metadata.EditMetadataRequest, error) {
	if len(args)%4 != 0 {
		return nil, errors.New("must supply arguments in multiples of 4")
	}
	var errs error
	result := &metadata.EditMetadataRequest{}
	for i := 0; len(args) > 0; i++ {
		edit := &metadata.Edit{}
		kind, picker, op, data := args[0], args[1], args[2], args[3]
		args = args[4:]
		switch strings.ToLower(kind) {
		case "project":
			var pp pfs.ProjectPicker
			if err := pp.UnmarshalText([]byte(picker)); err != nil {
				errors.JoinInto(&errs, errors.Errorf("arg set %d: unable to parse project picker %q", i, picker))
				continue
			}
			edit.Target = &metadata.Edit_Project{
				Project: &pp,
			}
		case "commit":
			var cp pfs.CommitPicker
			if err := cp.UnmarshalText([]byte(picker)); err != nil {
				errors.JoinInto(&errs, errors.Errorf("arg set %d: unable to parse commit picker %q", i, picker))
				continue
			}
			fixupProjects(&cp, defaultProject)
			edit.Target = &metadata.Edit_Commit{
				Commit: &cp,
			}
		case "branch":
			var bp pfs.BranchPicker
			if err := bp.UnmarshalText([]byte(picker)); err != nil {
				errors.JoinInto(&errs, errors.Errorf("arg set %d: unable to parse branch picker %q", i, picker))
				continue
			}
			fixupProjects(&bp, defaultProject)
			edit.Target = &metadata.Edit_Branch{
				Branch: &bp,
			}
		case "repo":
			var rp pfs.RepoPicker
			if err := rp.UnmarshalText([]byte(picker)); err != nil {
				errors.JoinInto(&errs, errors.Errorf("arg set %d: unable to parse repo picker %q", i, picker))
				continue
			}
			fixupProjects(&rp, defaultProject)
			edit.Target = &metadata.Edit_Repo{
				Repo: &rp,
			}
		case "cluster":
			edit.Target = &metadata.Edit_Cluster{}
		default:
			errors.JoinInto(&errs, errors.Errorf("arg set %d: unknown object type %q", i, kind))
			continue
		}
		switch strings.ToLower(op) {
		case "add":
			kv, err := kvparse.ParseOne(data)
			if err != nil {
				errors.JoinInto(&errs, errors.Errorf("arg set %d: parse error", err))
				continue
			}
			edit.Op = &metadata.Edit_AddKey_{
				AddKey: &metadata.Edit_AddKey{
					Key:   kv.Key,
					Value: kv.Value,
				},
			}
		case "edit":
			kv, err := kvparse.ParseOne(data)
			if err != nil {
				errors.JoinInto(&errs, errors.Errorf("arg set %d: parse error", err))
				continue
			}
			edit.Op = &metadata.Edit_EditKey_{
				EditKey: &metadata.Edit_EditKey{
					Key:   kv.Key,
					Value: kv.Value,
				},
			}
		case "delete":
			edit.Op = &metadata.Edit_DeleteKey_{
				DeleteKey: &metadata.Edit_DeleteKey{
					Key: data,
				},
			}
		case "replace":
			kvs, err := kvparse.ParseMany(data)
			if err != nil {
				errors.JoinInto(&errs, errors.Wrapf(err, "arg set %d: invalid 'replace' expression", i))
				continue
			}
			md := make(map[string]string)
			mdOK := true
			for _, kv := range kvs {
				if _, ok := md[kv.Key]; ok {
					errors.JoinInto(&errs, errors.Errorf("duplicate key %q", kv.Key))
					mdOK = false
				}
				md[kv.Key] = kv.Value
			}
			if !mdOK {
				continue
			}
			edit.Op = &metadata.Edit_Replace_{
				Replace: &metadata.Edit_Replace{
					Replacement: md,
				},
			}
		default:
			errors.JoinInto(&errs, errors.Errorf("arg set %d: unknown operation %q", i, op))
			continue
		}
		result.Edits = append(result.Edits, edit)
	}
	return result, errs
}

// fixupProjects recursively visits every ProjectPicker and replaces empty (not nil) ones with
// defaultProject.
func fixupProjects(msg proto.Message, defaultProject string) {
	if err := protorange.Range(msg.ProtoReflect(), func(v protopath.Values) error {
		for i, p := range v.Path {
			if d := p.FieldDescriptor(); d != nil && d.Kind() == protoreflect.MessageKind && d.Message().FullName() == "pfs_v2.ProjectPicker" {
				m := v.Index(i - 1).Value.Message()
				projectName := m.Get(d).Message().Get(
					(&pfs.ProjectPicker{}).ProtoReflect().Descriptor().Fields().ByName("name"),
				).String()
				if projectName == "" {
					m.Set(d, protoreflect.ValueOf((&pfs.ProjectPicker{
						Picker: &pfs.ProjectPicker_Name{
							Name: defaultProject,
						},
					}).ProtoReflect()))
				}
			}
		}
		return nil
	}); err != nil {
		panic("reflection-based request edit failed; specify project name explicitly")
	}
}
