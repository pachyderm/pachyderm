package helm

import (
	"github.com/pachyderm/pachyderm/src/client/pkg/config"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"
	"helm.sh/helm/v3/pkg/storage/driver"
)

// Deploy installs a helm chart.
func Deploy(context *config.Context, repoName, repoURL, installName, chartName, chartVersion string, values map[string]interface{}) (*release.Release, error) {
	envSettings, actionConfig, err := configureHelm(context, "")
	if err != nil {
		return nil, err
	}

	upgrade := action.NewUpgrade(actionConfig)
	upgrade.Version = chartVersion
	upgrade.Namespace = context.Namespace

	repoEntry := repo.Entry{
		Name: repoName,
		URL:  repoURL,
	}

	chartRepository, err := repo.NewChartRepository(&repoEntry, getter.All(envSettings))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to construct helm chart repository")
	}
	if _, err := chartRepository.DownloadIndexFile(); err != nil {
		return nil, errors.Wrapf(err, "failed to download helm index file at %q", chartRepository.Config.URL)
	}

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
