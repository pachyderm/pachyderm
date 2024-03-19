package cmds

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachctl"
	"github.com/pachyderm/pachyderm/v2/src/metadata"
	"github.com/pachyderm/pachyderm/v2/src/pfs"

	"github.com/spf13/cobra"
)

// Cmds returns a slice containing metadata commands.
func Cmds(pachctlCfg *pachctl.Config) []*cobra.Command {
	var commands []*cobra.Command

	editMetadata := &cobra.Command{
		Use:   `{{alias}} [<object type: project> <object picker> <operation: add|edit|delete|set> <data: key=value, key, '{"key":"value","key2":"value2"}'>]...`,
		Short: "Edits an object's metadata",
		Long:  "Edits an object's metadata",
		Run: cmdutil.Run(func(cmd *cobra.Command, args []string) error {
			req, err := parseEditMetadataCmdline(args)
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
	// TODO(jrockway): Allow keys and values with = in them.
	parts := strings.SplitN(data, "=", 2)
	return parts[0], parts[1]
}

func parseData(data string) (result map[string]string, _ error) {
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return nil, errors.Wrap(err, "parse json data")
	}
	return
}

func parseEditMetadataCmdline(args []string) (*metadata.EditMetadataRequest, error) {
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
			edit.Target = &metadata.Edit_Project{
				Project: &pfs.ProjectPicker{
					Picker: &pfs.ProjectPicker_Name{
						Name: picker,
					},
				},
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
