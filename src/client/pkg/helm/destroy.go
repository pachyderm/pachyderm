package helm

import (
	"github.com/pachyderm/pachyderm/src/client/pkg/config"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"

	"helm.sh/helm/v3/pkg/action"
)

// Destroy uninstalls a helm chart. Note that it does not remove the
// kubernetes resources.
func Destroy(context *config.Context, installName, overrideNamespace string) error {
	_, actionConfig, err := configureHelm(context, overrideNamespace)

	uninstall := action.NewUninstall(actionConfig)
	_, err = uninstall.Run(installName)
	if err != nil {
		return errors.Wrapf(err, "failed to uninstall helm package")
	}

	return nil
}
