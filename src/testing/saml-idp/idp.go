package main

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/crewjam/saml"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
)

type appEnv struct {
	// cert
	PublicCert string `env:"PUBLIC_CERT,default="`
	Port       int    `env:"METADATA_PORT,default=80"`
}

func main() {
	cmdutil.Main(serveIDPMetadata, &appEnv{})
}

func serveIDPMetadata(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)
	log.Printf("appEnv: %v", appEnv)
	if appEnv.PublicCert == "" {
		return errors.Errorf("IdP metadata server cannot start if PUBLIC_CERT env var is empty")
	}
	if appEnv.Port == 0 {
		return errors.Errorf("IdP metadata server cannot serve on port 0")
	}

	// Create saml.IdentityProvider struct, which generates the metadata bytes
	// we'll need to serve
	// TODO(msteffen): get name of the k8s service in front of this server from
	// environment variables and use that in the address here
	idpHost := fmt.Sprintf("0.0.0.0:%d", appEnv.Port)
	block, _ := pem.Decode([]byte(appEnv.PublicCert))
	if block == nil {
		// no error message, unfortunately
		return errors.Errorf("could not parse public cert")
	}
	idpCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return errors.Wrap(err, "could not parse public cert")
	}

	idp := &saml.IdentityProvider{
		MetadataURL: url.URL{
			Scheme: "http",
			Host:   idpHost,
		},
		Certificate: idpCert,
	}

	// Start listener & metadata http server
	log.Printf("Listening on port %d", appEnv.Port)
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", appEnv.Port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			log.Print("serving metadata...")
			idp.ServeMetadata(w, req)
		}),
	}
	return server.ListenAndServe()
}
