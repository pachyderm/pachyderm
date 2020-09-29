package cmds

import (
	"os"

	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/assets"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/images"
	"github.com/spf13/cobra"
)

func CreateExportImagesCmd(preRun PreRun, opts *assets.AssetOpts) *cobra.Command {
	exportImages := &cobra.Command{
		Use:    "{{alias}} <output-file>",
		Short:  "Export a tarball (to stdout) containing all of the images in a deployment.",
		Long:   "Export a tarball (to stdout) containing all of the images in a deployment.",
		PreRun: preRun,
		Run: cmdutil.RunFixedArgs(1, func(args []string) (retErr error) {
			file, err := os.Create(args[0])
			if err != nil {
				return err
			}
			defer func() {
				if err := file.Close(); err != nil && retErr == nil {
					retErr = err
				}
			}()
			return images.Export(opts, file)
		}),
	}
	return exportImages
}
