package main

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"

	"github.com/crewjam/saml"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
)

type appEnv struct {
	// cert
	PublicCert string `env:"PUBLIC_CERT,default="`
	Port       int    `env:METADATA_PORT,default=80`
}

func main() {
	cmdutil.Main(serveIDPMetadata, &appEnv{})
}

func serveIDPMetadata(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)
	if appEnv.PublicCert == "" {
		return fmt.Errorf("IdP metadata server cannot start if PUBLIC_CERT env var is empty")
	}

	// Create saml.IdentityProvider struct, which generates the metadata bytes
	// we'll need to serve
	// TODO(msteffen): get name of the k8s service in front of this server from
	// environment variables and use that in the address here
	idpHost := fmt.Sprintf("0.0.0.0:%d", appEnv.Port)
	block, _ := pem.Decode([]byte(appEnv.PublicCert))
	if block == nil {
		// no error message, unfortunately
		return fmt.Errorf("could not parse public cert")
	}
	idpCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("could not parse public cert: %v", err)
	}

	idp := &saml.IdentityProvider{
		MetadataURL: url.URL{
			Scheme: "http",
			Host:   idpHost,
		},
		Certificate: idpCert,
	}

	// Start listener & metadata http server
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", appEnv.Port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			idp.ServeMetadata(w, req)
		}),
	}
	return server.ListenAndServe()
}
