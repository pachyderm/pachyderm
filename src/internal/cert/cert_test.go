package cert

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

// TestBasic generates an x509 cert and then uses it to verify itself
func TestBasic(t *testing.T) {
	dnsName := "testing.pachyderm.com"

	// Generate self-signed cert
	cert, err := GenerateSelfSignedCert(dnsName, nil)
	require.NoError(t, err)
	pool := x509.NewCertPool()
	pool.AddCert(cert.Leaf)

	// Verify self-signed cert
	_, err = cert.Leaf.Verify(x509.VerifyOptions{
		CurrentTime: time.Now(),
		DNSName:     dnsName,
		Roots:       pool,
	})
	require.NoError(t, err)
}

// TestTLS sets up a local server and then uses a client to communicate with it
// over TLS
func TestTLS(t *testing.T) {
	dnsName := "testing.pachyderm.com"
	cert, err := GenerateSelfSignedCert(dnsName, nil)
	require.NoError(t, err)
	pool := x509.NewCertPool()
	pool.AddCert(cert.Leaf)

	// Implement a simple echo server
	l := NewTestListener()
	server := http.Server{
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{*cert},
		},
		Addr: dnsName,

		// Server is a simple echo server
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			b, err := ioutil.ReadAll(req.Body)
			if err != nil {
				w.Write([]byte("err: " + err.Error()))
				return
			}
			w.Write(b)
		}),
	}
	// Note: no need to provide cert/key files, as they're set in the TLSConfig
	go server.ServeTLS(l, "", "")

	// Create a client for the server above
	c := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: pool,
			},
			DisableCompression: true, // No extra headers
			DisableKeepAlives:  true, // TestListener only allows one connection

			// Because TestListener ignores the address it's given, even though we'll
			// dial testing.pachyderm.com, it'll just connect to the server above (but
			// the TLS cert will be signed for the right domain)
			DialContext: l.Dial,
		},
	}

	// Send a secret message to the server, and make sure we get the expected
	// response back
	message := []byte("secret message")
	resp, err := c.Post("https://"+dnsName, "text/plain", bytes.NewReader(message))
	require.NoError(t, err)
	respText, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, respText, message)

	// Note: To make this test fail, you need to both:
	// - change server.ServeTLS() above to server.Serve() (disable TLS on server)
	// - change c.Post("https://"...) to c.Post("http://"...) (disable on client)
	// OR
	// - change dnsName after generating the certificate (causes test to hang
	//   rather than fail immediately. Not sure why yet)
	require.False(t, bytes.Contains(l.ClientToServerLog(), message))
	require.False(t, bytes.Contains(l.ServerToClientLog(), message))
}

// TestToPEM tests the PEM-serialization helper functions in cert.go
func TestToPEM(t *testing.T) {
	// RSA private key (generated separately with openssl)
	privateKey := []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEAnn4t95creCwrQVvg0omFja2sfSCAZGrpU5Nxhex0k6crP7Vh
82wcb/y1Bxsd/F+6begMXHq0+Zt1cka2vlmKsucTkMCSirtn77FXK7Ib7u2VAYa8
kDEVHoeEHjGhVrPkl3jWagl2gr9VM5r5eq1YajS4s6DjwBf56y636/9lsy0WXfbl
MzwSnKkZyIFDLBa/GBZdcFT3ev9xYxJ7nPgKxvjkBX8tji18BSnBjV1ZWix5JIsd
2GW+STuFAy5UAym5vzyOkHEQrVWKCMJ9UnwAEI/Yup8MSme2Ob5mpOmEOBAfGYqE
/2TZBoXqCxtHz5vAWKb0WkPsGhXWYNPwf+BLYwIDAQABAoIBAQCT4FDNMIuLXVKi
cbIrXcpxLTjBqoCAsMuwgeIqvTrrxM5ya67PavCBgDv7PE7W+Q49m4NlCcwvE+AZ
1maM5YimcTltFm/j5wULu+AEUfMEE0Gyod7vfgwhZvlbHp1VAxVmSoVrfBbJ2PEK
7C6XSoMy3Kv0VUoKIZS53OYX2DwwVooLOHgUA/Q+NDTlFiMwQRAuuZAAu9NvcMQA
HNsDIS8KKmltPmqD+CU9mHiXwuj4UtTBODGzuhR+39BVk0F6VaD9if8s/B0RTfg7
HiFjrP1DyjqsTZESg3m0+nDchcVxRVXHScFOVfXzpsT3dpR6TYHn2XA4f/cJeFoo
Q7Iy6IoxAoGBAM4w/OKNRtMemMal75Sbmj0qu1wOxlH2BbvEy5IXfMscE2f+bu/U
9CKhxXwqqtbTAQu7jAn4uNUm2DAam7nBgJPS9/eW8nRhMPb2k07QOZ29BZ5KMi+r
UcYySx7qa+TOae456QYmP3qO0ZtUpimu8r6X1yihRhPiLiVjKJPKYI2pAoGBAMTH
fH5tn17og/AspFhBXceueSzncTa9crSY5MKgrmEb4iQdK2WIWLTHggDzKuGAMyFl
vlxqdG+vO8AY4klY69TaLoXwc0KmyZjDtWPyo7YCeN1Hudo2ujdIn3zNdSuH3XM3
bS1Gh2xEsyMGNNP4YoUT9O9jiIQj/vh5nYCgtIArAoGBAK62u9GMPHMv/ex1NqkJ
oIwr5U6ABnP0r68Hdid4V3oTdC4uXfpCzAt8YEZyMQiPCtfSNztL0fJrU8yO/11L
JZQcs5jMAu2yXTcmgHPL5MZQIK6b2CKkXEpA2356zKm4bfI6h8V6K1fCJMIl3BZ9
85qkNuBqp2K5yLhNaVixp1bhAoGAK3cU3KhCJ6icXBTASG5H1K+JPI3yx/CYwaN0
BDmRywlprihzSX4Qef4HjUYpFp5GrP3YSnmJNpIyVIAqm6D0lpOK6zLtgq9soD26
d1VFLBLnt5j8SGMGRufXsq1/UBo2pBh+GR4XE6cpGndoe9nFiTebRrVpliaNTz0t
uRfGRvkCgYEAuDXErZ3tb98nBqg8BD2X9MYwyLMCNOBsyEqCaERKsKtDiJTpRcpa
gHZ/5BN5nz5oCsMtCrW9JZDauCby+MlPIAhmy3gjY0fZbqdHbyoMaIbrtm9TOwFg
maI1ml7Sfxr0dyd7f4+Co5TP+MhHlv83rCVUfB1SMnP+QJkZwVQyfdQ=
-----END RSA PRIVATE KEY-----
`)
	// Self-signed TLS cert for testing.pachyderm.com (generated with openssl)
	cert := []byte(`-----BEGIN CERTIFICATE-----
MIIDGDCCAgCgAwIBAgIJAJu/qP84dDeFMA0GCSqGSIb3DQEBCwUAMCAxHjAcBgNV
BAMMFXRlc3RpbmcucGFjaHlkZXJtLmNvbTAgFw0xODA5MDYyMDUwMjBaGA80NzU2
MDgwMzIwNTAyMFowIDEeMBwGA1UEAwwVdGVzdGluZy5wYWNoeWRlcm0uY29tMIIB
IjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAnn4t95creCwrQVvg0omFja2s
fSCAZGrpU5Nxhex0k6crP7Vh82wcb/y1Bxsd/F+6begMXHq0+Zt1cka2vlmKsucT
kMCSirtn77FXK7Ib7u2VAYa8kDEVHoeEHjGhVrPkl3jWagl2gr9VM5r5eq1YajS4
s6DjwBf56y636/9lsy0WXfblMzwSnKkZyIFDLBa/GBZdcFT3ev9xYxJ7nPgKxvjk
BX8tji18BSnBjV1ZWix5JIsd2GW+STuFAy5UAym5vzyOkHEQrVWKCMJ9UnwAEI/Y
up8MSme2Ob5mpOmEOBAfGYqE/2TZBoXqCxtHz5vAWKb0WkPsGhXWYNPwf+BLYwID
AQABo1MwUTAdBgNVHQ4EFgQUj8Wfd80glMBSTNJ74rOZDjpctEYwHwYDVR0jBBgw
FoAUj8Wfd80glMBSTNJ74rOZDjpctEYwDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG
9w0BAQsFAAOCAQEAbLQVgpcZbfIvp0X6D7AoZIi+ZqQtQXLxWDkIYM99ZSatS64o
7JKoXGOQMdt/wIgLD5YvjHfXABBJgOtvEDPv4KFxm+XVvb4lDtTH4PLR0dBVw1zl
ltyhsG07jznpEMmBr5eMgKti9fPAeOmS/Nv3oRuSVtOf3pVMk9CxPzvyKKCAg0ee
gAFHGMGADvyIOcZMUDn7MCSl08ciwEyjZZDa6Fgbpihm77rRrp4udR58q8VtE0m6
f3vxVUn0ZJ54JbIWKeJnJ5Svelzm2JBg/sWcJAPR4btMKv9Jie/THQd53QGNyE/I
TF+qRCQa5ciV5EtqZWgI5xpVSPDxf6C+Mm17vw==
-----END CERTIFICATE-----
`)
	tlsCert, err := tls.X509KeyPair(cert, privateKey)
	require.NoError(t, err)
	require.Equal(t, privateKey, KeyToPEM(&tlsCert))
	require.Equal(t, cert, PublicCertToPEM(&tlsCert))
}
