package images

import (
	"fmt"
	"io"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/assets"
)

// Export a tarball of the images needed by a deployment.
func Export(opts *assets.AssetOpts, out io.Writer) error {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return err
	}
	authConfigs, err := docker.NewAuthConfigurationsFromDockerCfg()
	if err != nil {
		return fmt.Errorf("error parsing auth: %s, try running `docker login`", err.Error())
	}
	if len(authConfigs.Configs) == 0 {
		return fmt.Errorf("didn't find any valud auth configurations")
	}
	images := assets.Images(opts)
	for _, image := range images {
		repository, tag := docker.ParseRepositoryTag(image)
		pulled := false
		var loopErr error
		for _, authConfig := range authConfigs.Configs {
			if err := client.PullImage(
				docker.PullImageOptions{
					Repository:        repository,
					Tag:               tag,
					InactivityTimeout: 10 * time.Millisecond,
				},
				authConfig,
			); err != nil {
				loopErr = err
				continue
			}
			pulled = true
			break
		}
		if !pulled {
			return fmt.Errorf("error pulling image: %+v", loopErr)
		}
	}
	return client.ExportImages(docker.ExportImagesOptions{
		Names:        images,
		OutputStream: out,
	})
}
