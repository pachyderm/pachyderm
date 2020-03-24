package helm

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/client/pkg/config"
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
		return nil, fmt.Errorf("could not locate helm chart: %v", err)
	}

	chart, err := loader.Load(chartPath)
	if err != nil {
		return nil, fmt.Errorf("could not load helm chart: %v", err)
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
			return nil, fmt.Errorf("install failed: %v", err)
		}
		return rel, nil
	} else if err != nil {
		return nil, fmt.Errorf("could not fetch helm history: %v", err)
	}

	// upgrade
	rel, err := upgrade.Run(installName, chart, values)
	if err != nil {
		return nil, fmt.Errorf("upgrade failed: %v", err)
	}
	return rel, nil
}
