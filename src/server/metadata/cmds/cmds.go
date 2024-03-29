package cmds

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
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
		Use:   `{{alias}} [<object type: project|repo|branch|commit> <object picker> <operation: add|edit|delete|set> <data: key=value, key, '{"key":"value","key2":"value2"}'>]...`,
		Short: "Edits an object's metadata",
		Long:  "Edits an object's metadata.",
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

func parseKV(data string) (k, v string) {
	// TODO(jrockway, CORE-2228): Allow keys and values with = in them.
	parts := strings.SplitN(data, "=", 2)
	return parts[0], parts[1]
}

func parseData(data string) (result map[string]string, _ error) {
	if len(data) == 0 || data[0] != '{' {
		return nil, errors.New("data must be a JSON object")
	}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return nil, errors.Wrap(err, "parse json data")
	}
	return
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
		default:
			errors.JoinInto(&errs, errors.Errorf("arg set %d: unknown object type %q", i, kind))
			continue
		}
		switch strings.ToLower(op) {
		case "add":
			k, v := parseKV(data)
			edit.Op = &metadata.Edit_AddKey_{
				AddKey: &metadata.Edit_AddKey{
					Key:   k,
					Value: v,
				},
			}
		case "edit":
			k, v := parseKV(data)
			edit.Op = &metadata.Edit_EditKey_{
				EditKey: &metadata.Edit_EditKey{
					Key:   k,
					Value: v,
				},
			}
		case "delete":
			edit.Op = &metadata.Edit_DeleteKey_{
				DeleteKey: &metadata.Edit_DeleteKey{
					Key: data,
				},
			}
		case "set":
			md, err := parseData(data)
			if err != nil {
				errors.JoinInto(&errs, errors.Wrapf(err, "arg set %d: invalid 'set' expression", i))
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
