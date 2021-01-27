package tls

import (
	"os"
	"path"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
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

	// CertCheckFrequency is how often we check for a renewed TLS certificate
	CertCheckFrequency = time.Hour
)

// GetCertPaths gets the paths to the cert and key files within a cluster
func GetCertPaths() (certPath string, keyPath string, err error) {
	certPath = path.Join(VolumePath, CertFile)
	if _, err = os.Stat(certPath); err != nil {
		err = errors.Wrapf(err, "could not stat public cert at %s", certPath)
		return
	}

	keyPath = path.Join(VolumePath, KeyFile)
	if _, err = os.Stat(keyPath); err != nil {
		err = errors.Wrapf(err, "could not stat private key at %s", keyPath)
		return
	}
	return
}
