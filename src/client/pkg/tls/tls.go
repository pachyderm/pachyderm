package tls

import (
	"fmt"
	"os"
	"path"
)

const (
	// VolumePath is the path at which the tls cert and private key (if any)
	// will be mounted in the pachd pod
	VolumePath = "/pachd-tls-cert"

	// CertFile is the name of the mounted file containing a TLS certificate
	// that identifies pachd
	CertFile = "tls.crt"

	// KeyFile is the name of the mounted file containing a private key
	// corresponding to the public certificate in TLSCertFile
	KeyFile = "tls.key"
)

// GetCertPaths gets the paths to the cert and key files within a cluster
func GetCertPaths() (certPath string, keyPath string, err error) {
	certPath = path.Join(VolumePath, CertFile)
	if _, err = os.Stat(certPath); err != nil {
		err = fmt.Errorf("could not stat public cert at %s: %v", certPath, err)
		return
	}

	keyPath = path.Join(VolumePath, KeyFile)
	if _, err = os.Stat(keyPath); err != nil {
		err = fmt.Errorf("could not stat private key at %s: %v", keyPath, err)
		return
	}
	return
}
