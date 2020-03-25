package helm

import (
	"github.com/pachyderm/pachyderm/src/client/pkg/config"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
)

func Deploy(context *config.Context, installName, chartName, chartVersion string, values map[string]interface{}) (*release.Release, error) {
	envSettings, actionConfig, err := configureHelm(context, "")

	upgrade := action.NewUpgrade(actionConfig)
	upgrade.Version = chartVersion
	upgrade.Namespace = context.Namespace
	// upgrade.Install = true

	chartPath, err := upgrade.ChartPathOptions.LocateChart(chartName, envSettings)
	if err != nil {
		return nil, errors.Wrapf(err, "could not locate helm chart")
	}

	chart, err := loader.Load(chartPath)
	if err != nil {
		return nil, errors.Wrapf(err, "could not load helm chart")
	}

	history := action.NewHistory(actionConfig)
	history.Max = 1
	if _, err := history.Run(installName); err == driver.ErrReleaseNotFound {
		// install
		install := action.NewInstall(actionConfig)
		install.ChartPathOptions = upgrade.ChartPathOptions
		install.Version = chartVersion
		install.Namespace = context.Namespace
		install.ReleaseName = installName
		rel, err := install.Run(chart, values)
		if err != nil {
			return nil, errors.Wrapf(err, "install failed")
		}
		return rel, nil
	} else if err != nil {
		return nil, errors.Wrapf(err, "could not fetch helm history")
	}

	// upgrade
	rel, err := upgrade.Run(installName, chart, values)
	if err != nil {
		return nil, errors.Wrapf(err, "upgrade failed")
	}
	return rel, nil
}
