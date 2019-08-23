package config

import (
	"errors"
	"fmt"

	"github.com/pachyderm/pachyderm/src/client/pkg/erronce"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	pachContextOnce  erronce.ErrOnce
	pachContextValue *Context
)

func getUnstructuredString(us *unstructured.Unstructured, name string) (string, error) {
	value, ok := us.Object[name].(string)
	if !ok {
		return "", fmt.Errorf("the Pachyderm context contains an invalid value for `%s`", name)
	}
	return value, nil
}

// ReadPachConfig loads the Pachyderm context from the kube config.
func ReadPachContext() (*Context, error) {
	err := pachContextOnce.Do(func() error {
		kubeConfig := ReadKubeConfig()
		kubeConfigAccess := kubeConfig.ConfigAccess()
		kubeStartingConfig, err := kubeConfigAccess.GetStartingConfig()
		if err != nil {
			return fmt.Errorf("could not fetch kubernetes' starting config: %v", err)
		}

		if len(kubeStartingConfig.CurrentContext) == 0 {
			return errors.New("kubernetes' current context has not been set")
		}

		kubeContext, ok := kubeStartingConfig.Contexts[kubeStartingConfig.CurrentContext]
		if !ok {
			return errors.New("kubernetes' current config refers to one that does not exist")
		}

		pachContextValue = &Context{}

		obj, ok := kubeContext.Extensions["pachyderm-context-v1"]
		if !ok {
			return nil
		}

		us, ok := obj.(*unstructured.Unstructured)
		if !ok || us == nil {
			return nil
		}

		pachdAddress, err := getUnstructuredString(us, "pachd_address")
		if err != nil {
			return err
		}
		pachContextValue.PachdAddress = pachdAddress

		serverCAs, err := getUnstructuredString(us, "server_cas")
		if err != nil {
			return err
		}
		pachContextValue.ServerCAs = serverCAs

		sessionToken, err := getUnstructuredString(us, "session_token")
		if err != nil {
			return err
		}
		pachContextValue.SessionToken = sessionToken

		activeTransaction, err := getUnstructuredString(us, "active_transaction")
		if err != nil {
			return err
		}
		pachContextValue.ActiveTransaction = activeTransaction

		source, err := getUnstructuredString(us, "source")
		if err != nil {
			return err
		}
		switch source {
		case "CONFIG_V1":
			pachContextValue.Source = ContextSource_CONFIG_V1
		default:
			pachContextValue.Source = ContextSource_NONE
		}

		return nil
	})

	return pachContextValue, err
}

// Write writes the context in 'c' to the current kube config.
func (c *Context) Write() error {
	kubeConfig := ReadKubeConfig()
	kubeConfigAccess := kubeConfig.ConfigAccess()
	kubeStartingConfig, err := kubeConfigAccess.GetStartingConfig()
	if err != nil {
		return fmt.Errorf("could not fetch kubernetes' starting config: %v", err)
	}

	if len(kubeStartingConfig.CurrentContext) == 0 {
		return errors.New("kubernetes' current context has not been set")
	}

	kubeContext, ok := kubeStartingConfig.Contexts[kubeStartingConfig.CurrentContext]
	if !ok {
		return errors.New("kubernetes' current config refers to one that does not exist")
	}

	kubeContext.Extensions["pachyderm-context-v1"] = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"pachd_address":      c.PachdAddress,
			"server_cas":         c.ServerCAs,
			"session_token":      c.SessionToken,
			"active_transaction": c.ActiveTransaction,
		},
	}

	if err := clientcmd.ModifyConfig(kubeConfigAccess, *kubeStartingConfig, true); err != nil {
		return fmt.Errorf("could not modify kubernetes config: %v", err)
	}

	return nil
}
