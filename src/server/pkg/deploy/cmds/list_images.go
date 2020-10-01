package cmds

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/assets"
	"github.com/spf13/cobra"
)

func CreateListImagesCmd(preRun PreRun, opts *assets.AssetOpts, globalFlags *GlobalFlags) *cobra.Command {
	listImages := &cobra.Command{
		Short:  "Output the list of images in a deployment.",
		Long:   "Output the list of images in a deployment.",
		PreRun: preRun,
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			for _, image := range assets.Images(opts) {
				fmt.Println(image)
			}
			return nil
		}),
	}
	appendGlobalFlags(listImages, globalFlags)
	return listImages
}
