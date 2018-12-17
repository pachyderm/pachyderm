package client

import (
	"fmt"
	"io"
	"net/http"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/client-go/rest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DEFAULT_DAEMON_LOCAL_PORT = 30650
	DEFAULT_SAML_ACS_LOCAL_PORT = 30654
	DEFAULT_DASH_UI_PORT = 30080
	DEFAULT_DASH_WEBSOCKET_PORT = 30081
)

type PortForwarder struct {
	client rest.Interface
	config *rest.Config
	namespace string
	podName string
	localPort int
	remotePort int
	stdout io.Writer
	stderr io.Writer
	stopChan chan struct{}
}

func NewPortForwarder(config *rest.Config, namespace string, selector map[string]string, localPort, remotePort int, stdout, stderr io.Writer) (*PortForwarder, error) {
	if namespace == "" {
		namespace = "default"
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	core := client.CoreV1()

	podList, err := core.Pods(namespace).List(metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(selector)),
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		},
	})
	if err != nil {
		return nil, err
	}
	if len(podList.Items) != 1 {
		return nil, fmt.Errorf("Incorrect number of pods returned for selector %v: %d", selector, len(podList.Items))
	}

	return &PortForwarder {
		client: core.RESTClient(),
		config: config,
		namespace: namespace,
		podName: podList.Items[0].Name,
		localPort: localPort,
		remotePort: remotePort,
		stdout: stdout,
		stderr: stderr,
		stopChan: make(chan struct{}, 1),
	}, nil
}

func (f *PortForwarder) Run() error {
	url := f.client.Post().
		Resource("pods").
		Namespace(f.namespace).
		Name(f.podName).
		SubResource("portforward").
		URL()

	transport, upgrader, err := spdy.RoundTripperFor(f.config)
	if err != nil {
		return err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", url)
	ports := []string{fmt.Sprintf("%d:%d", f.localPort, f.remotePort)}
	readyChan := make(chan struct{}, 1)
	fw, err := portforward.New(dialer, ports, f.stopChan, readyChan, f.stdout, f.stderr)
	if err != nil {
		return err
	}

	errChan := make(chan error, 1)
	go func() { errChan <- fw.ForwardPorts() }()

	select {
	case err = <- errChan:
		return fmt.Errorf("port forwarding failed: %v", err)
	case <- fw.Ready:
		return nil
	}
}

func (f *PortForwarder) Close() {
	close(f.stopChan)
}

func DaemonForwarder(config *rest.Config, namespace string, localPort int, stdout, stderr io.Writer) (*PortForwarder, error) {
	if localPort == 0 {
		localPort = DEFAULT_DAEMON_LOCAL_PORT
	}
	selector := map[string]string{"app": "pachd"}
	return NewPortForwarder(config, namespace, selector, localPort, 650, stdout, stderr)
}

func SAMLACSForwarder(config *rest.Config, namespace string, localPort int, stdout, stderr io.Writer) (*PortForwarder, error) {
	if localPort == 0 {
		localPort = DEFAULT_SAML_ACS_LOCAL_PORT
	}
	// TODO(ys): using a suite selector because the original code had that.
	// check if it is necessary.
	selector := map[string]string{"suite": "pachyderm", "app": "pachd"}
	return NewPortForwarder(config, namespace, selector, localPort, 654, stdout, stderr)
}

func DashUIForwarder(config *rest.Config, namespace string, localPort int, stdout, stderr io.Writer) (*PortForwarder, error) {
	if localPort == 0 {
		localPort = DEFAULT_DASH_UI_PORT
	}
	selector := map[string]string{"app": "dash"}
	return NewPortForwarder(config, namespace, selector, localPort, 8080, stdout, stderr)
}

func DashWebSocketForwarder(config *rest.Config, namespace string, localPort int, stdout, stderr io.Writer) (*PortForwarder, error) {
	if localPort == 0 {
		localPort = DEFAULT_DASH_UI_PORT
	}
	selector := map[string]string{"app": "dash"}
	return NewPortForwarder(config, namespace, selector, localPort, 8081, stdout, stderr)
}
