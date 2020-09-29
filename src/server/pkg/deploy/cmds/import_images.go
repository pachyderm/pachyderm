package cmds

import (
	"os"

	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/assets"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/images"
	"github.com/spf13/cobra"
)

func CreateImportImagesCmd(preRun PreRun, opts *assets.AssetOpts) *cobra.Command {
	importImages := &cobra.Command{
		Use:    "{{alias}} <input-file>",
		Short:  "Import a tarball (from stdin) containing all of the images in a deployment and push them to a private registry.",
		Long:   "Import a tarball (from stdin) containing all of the images in a deployment and push them to a private registry.",
		PreRun: preRun,
		Run: cmdutil.RunFixedArgs(1, func(args []string) (retErr error) {
			file, err := os.Open(args[0])
			if err != nil {
				return err
			}
			defer func() {
				if err := file.Close(); err != nil && retErr == nil {
					retErr = err
				}
			}()
			return images.Import(opts, file)
		}),
	}

	return importImages
}
