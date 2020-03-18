package images

import (
	"io"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
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
		return errors.Wrapf(err, "error parsing auth, try running `docker login`")
	}
	if len(authConfigs.Configs) == 0 {
		return errors.Errorf("didn't find any valid auth configurations")
	}
	images := assets.Images(opts)
	for _, image := range images {
		repository, tag := docker.ParseRepositoryTag(image)
		pulled := false
		var loopErr []error
		for registry, authConfig := range authConfigs.Configs {
			if err := client.PullImage(
				docker.PullImageOptions{
					Repository:        repository,
					Tag:               tag,
					InactivityTimeout: 5 * time.Second,
				},
				authConfig,
			); err != nil {
				loopErr = append(loopErr, errors.Wrapf(err, "error pulling from %s", registry))
				continue
			}
			pulled = true
			break
		}
		if !pulled {
			errStr := ""
			for _, err := range loopErr {
				errStr += err.Error() + "\n"
			}
			return errors.Errorf("errors pulling image %s:%s:\n%s", repository, tag, errStr)
		}
	}
	return client.ExportImages(docker.ExportImagesOptions{
		Names:        images,
		OutputStream: out,
	})
}

// Import a tarball of the images needed by a deployment such as the one
// created by Export and push those images to the registry specific in opts.
func Import(opts *assets.AssetOpts, in io.Reader) error {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return err
	}
	authConfigs, err := docker.NewAuthConfigurationsFromDockerCfg()
	if err != nil {
		return errors.Wrapf(err, "error parsing auth, try running `docker login`")
	}
	if len(authConfigs.Configs) == 0 {
		return errors.Errorf("didn't find any valid auth configurations")
	}
	if err := client.LoadImage(docker.LoadImageOptions{
		InputStream: in,
	}); err != nil {
		return err
	}
	registry := opts.Registry
	opts.Registry = "" // pretend we're using default images so we can get targets to tag
	images := assets.Images(opts)
	opts.Registry = registry
	for _, image := range images {
		repository, tag := docker.ParseRepositoryTag(image)
		registryRepo := assets.AddRegistry(opts.Registry, repository)
		if err := client.TagImage(image, docker.TagImageOptions{
			Repo: registryRepo,
			Tag:  tag,
		},
		); err != nil {
			return errors.Wrapf(err, "error tagging image")
		}
		pushed := false
		var loopErr []error
		for registry, authConfig := range authConfigs.Configs {
			if err := client.PushImage(
				docker.PushImageOptions{
					Name:              registryRepo,
					Tag:               tag,
					Registry:          opts.Registry,
					InactivityTimeout: 5 * time.Second,
				},
				authConfig,
			); err != nil {
				loopErr = append(loopErr, errors.Wrapf(err, "error pushing to %s", registry))
				continue
			}
			pushed = true
			break
		}
		if !pushed {
			if len(loopErr) > 0 {
				return loopErr[0]
			} else {
				return errors.Errorf("failed to push images because there are no auth configs")
			}
		}
	}
	return nil
}
