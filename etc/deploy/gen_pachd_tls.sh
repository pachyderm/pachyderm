#!/bin/bash
# This script generates a self-signed TLS cert to be used by pachd in tests

hostport=$1
output_prefix=${2:-pachd}
# shellcheck disable=SC2001
host="$(echo "$hostport" | sed -e 's,:.*,,g')"
if [[ "${host}" =~ [0-9]+\.[0-9]+\.[0-9]+\.[0-9]+ ]]; then
  ip=${host}
else
  dns=${host}
fi

# Define a minimal openssl config for our micro-CA
read -d '' -r tls_config <<EOF
[ req ]
default_md         = sha256 # MD = message digest. md5 is the openSSL default in 1.1.0 (see 'man req')
prompt             = no     # use values in [dn] directly
distinguished_name = dn
x509_extensions    = exn    # Since we're making self-signed certs. For CSRs, use req_extensions

[ dn ]
CN = ${dns:-localhost}

[ exn ]
EOF

if [[ -n "${ip}" ]]; then
  tls_config+=$'\n'"subjectAltName = IP:${ip}"
fi

echo "${tls_config}"

# Set other openssl options
tls_opts=(
  # Immediately self-sign the generated CSR and output that, instead of
  # outputting the CSR itself
  -x509

  # Don't encrypt (DES) the resulting cert (dangerous, non-prod only)
  -nodes

  # signed cert should be valid for 1 year
  -days 365

  # Generate the cert's private key as well (instead of receiving one)
  -newkey rsa:2048

  # Output the private key here # Output the private key here
  -keyout "${output_prefix}.key"

  # Output PEM-encoded cert (this is the default, and this flag is unnecessary,
  # but PEM is required by kubernetes and this makes explicit the fact that
  # we're meeting that requirement
  -outform PEM

  # Output path for the self-signed cert
  -out "${output_prefix}.pem"
)

# Generate self-signed cert
openssl req "${tls_opts[@]}" -config <(echo "${tls_config}")

# Print instructions for using new cert and key
echo "New cert and key are in '${output_prefix}.pem' and '${output_prefix}.key'"
echo "Deploy pachd to present the new self-signed cert and key by running:"
echo ""
echo "  pachctl undeploy # remove any existing cluster"
echo "  pachctl deploy <destination> --tls=\"${output_prefix}.pem,${output_prefix}.key\""
echo ""
