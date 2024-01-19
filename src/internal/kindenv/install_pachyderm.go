package kindenv

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

type InstallConfig struct {
	AdditionalValuesFiles []string
}

// InstallPachyderm installs or upgrades Pachyderm.
func (c *Cluster) InstallPachyderm(ctx context.Context, config *ClusterConfig, valuesFiles []string) error {
	k, err := c.GetKubeconfig(ctx)
	if err != nil {
		return errors.Wrap(err, "get kubeconfig")
	}

	chart := "etc/helm/pachyderm"
	return k.HelmCommand(ctx, "upgrade", "--install", "pachyderm", chart).Run()
}
