package helm

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/client/pkg/config"
	"helm.sh/helm/v3/pkg/action"
)

func Destroy(context *config.Context, installName, overrideNamespace string) error {
	_, actionConfig, err := configureHelm(context, overrideNamespace)

	uninstall := action.NewUninstall(actionConfig)
	_, err = uninstall.Run(installName)
	if err != nil {
		return fmt.Errorf("failed to uninstall helm package: %v", err)
	}

	return nil
}
