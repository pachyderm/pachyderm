#!/bin/bash
# This script generates a self-signed TLS cert to be used by pachd in tests

eval "set -- $( getopt -l "dns:,ip:,port:" -o "o:" "--" "${0}" "${@:-}" )"
output_prefix=pachd
while true; do
  case "${1}" in
    --dns)
      dns="${2}"
      shift 2
      ;;
    --ip)
      ip="${2}"
      shift 2
      ;;
    --port)
      port="${2}"
      shift 2
      ;;
    -o)
      output_prefix="${2}"
      shift 2
      ;;
    --)
      shift
      break
      ;;
  esac
done

# Validate flags
if [[ -z "${dns}" ]] && [[ -z "${ip}" ]]; then
  cat <<EOF >/dev/fd/2
Error: You must set either --dns or --ip

Usage:
Generate a self-signed TLS certificate for Pachyderm

Args:
  --dns  Set the domain name in the certificate (i.e. the common name with
         which pachd will authenticate)

  --ip   The IP address in the certificate (which you plan to use to connect
         to pachd. Without this, enabling TLS while connecting to pachd over
         an IP will not work)

   -o <prefix>
        This script generates a public TLS cert at <prefix>.pem and <prefix>.key
        Setting this arg determines the output file names.
EOF
  exit 1
fi
if [[ -z "${port}" ]]; then
  echo "Warning, --port is unset. Assuming :30650 (.v1.pachd_address in your "
  echo "Pachyderm config must be updated if this is not the correct port, or "
  echo "pachd will fail to connect)"
fi
port="${port:-30650}"

# Define a minimal openssl config for our micro-CA
read -d '' -r tls_config <<EOF
[ req ]
default_md         = sha256 # MD = message digest. md5 is the openSSL default in 1.1.0 (see 'man req')
prompt             = no     # use values in [dn] directly
distinguished_name = dn
x509_extensions    = exn    # Since we're making self-signed certs. For CSRs, use req_extensions

[ dn ]
CN = ${dns:-localhost} # TODO(msteffen) better default domain name

[ exn ]
EOF

# If 'ip' is set, include IP in TLS cert
if [[ -n "${ip}" ]]; then
  tls_config+=$'\n'"subjectAltName = IP:${ip}"
  if [[ -n "${dns}" ]]; then
    tls_config+=", DNS:${dns}"
  fi
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
  -keyout ${output_prefix}.key

  # Output PEM-encoded cert (this is the default, and this flag is unnecessary,
  # but PEM is required by kubernetes and this makes explicit the fact that
  # we're meeting that requirement
  -outform PEM

  # Output path for the self-signed cert
  -out ${output_prefix}.pem
)

# Generate self-signed cert
openssl req "${tls_opts[@]}" -config <(echo "${tls_config}")

# Print instructions for using new cert and key
host="${dns}"
if [[ -n "${ip}" ]]; then
  host="${ip}"
fi
echo "New cert and key are in '${output_prefix}.pem' and '${output_prefix}.key'"
echo "Deploy pachd to present the new self-signed cert and key by running:"
echo ""
echo "  pachctl undeploy # remove any existing cluster"
echo "  pachctl deploy <destination> --tls=\"${output_prefix}.pem,${output_prefix}.key\""
echo ""
echo "Configure pachctl to trust the new self-signed cert by running:"
echo ""
echo "  pachctl config update context \\"
echo "    --pachd-address=\"grpcs://${host}:${port}\" \\"
echo "    --server-cas=\"\$(cat ./${output_prefix}.pem | base64)\""
echo ""
