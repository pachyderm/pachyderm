package helm

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/client/pkg/config"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"

	log "github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func configureHelm(context *config.Context, overrideNamespace string) (*cli.EnvSettings, *action.Configuration, error) {
	envSettings := cli.New()

	actionConfig := new(action.Configuration)

	if overrideNamespace == "" {
		overrideNamespace = context.Namespace
	}

	configFlags := &genericclioptions.ConfigFlags{
		ClusterName:  &context.ClusterName,
		AuthInfoName: &context.AuthInfo,
		Namespace:    &overrideNamespace,
	}

	if err := actionConfig.Init(configFlags, overrideNamespace, "", func(format string, v ...interface{}) {
		log.Debugf(format, v...)
	}); err != nil {
		return nil, nil, errors.Wrapf(err, "could not init helm config")
	}

	return envSettings, actionConfig, nil
}
