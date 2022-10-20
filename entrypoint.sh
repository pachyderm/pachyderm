#!/bin/bash
# Based on https://raw.githubusercontent.com/brainsam/pgbouncer/master/entrypoint.sh

set -e

print_validation_error() {
    echo "$1"
    exit 1
}

# Here are some parameters. See all on
# https://pgbouncer.github.io/config.html

PG_CONFIG_DIR=/tmp

# Write the password with MD5 encryption, to avoid printing it during startup.
# Notice that `docker inspect` will show unencrypted env variables.
_AUTH_FILE="${AUTH_FILE:-$PG_CONFIG_DIR/userlist.txt}"

# Workaround userlist.txt missing issue
# https://github.com/edoburu/docker-pgbouncer/issues/33
if [ ! -e "${_AUTH_FILE}" ]; then
    touch "${_AUTH_FILE}"
fi

pass="md5$(echo -n "$POSTGRESQL_PASSWORD$POSTGRESQL_USERNAME" | md5sum | cut -f 1 -d ' ')"
echo "\"$POSTGRESQL_USERNAME\" \"$pass\"" >> ${PG_CONFIG_DIR}/userlist.txt
echo "Wrote authentication credentials to ${PG_CONFIG_DIR}/userlist.txt"

# TLS Checks (server)
if [[ -n "$PGBOUNCER_SERVER_TLS_CERT_FILE" ]] && [[ ! -f "$PGBOUNCER_SERVER_TLS_CERT_FILE" ]]; then
    print_validation_error "The X.509 server certificate file in the specified path ${PGBOUNCER_SERVER_TLS_CERT_FILE} does not exist"
fi
if [[ -n "$PGBOUNCER_SERVER_TLS_KEY_FILE" ]] && [[ ! -f "$PGBOUNCER_SERVER_TLS_KEY_FILE" ]]; then
    print_validation_error "The server private key file in the specified path ${PGBOUNCER_SERVER_TLS_KEY_FILE} does not exist"
fi
if [[ -n "$PGBOUNCER_SERVER_TLS_CA_FILE" ]] && [[ ! -f "$PGBOUNCER_SERVER_TLS_CA_FILE" ]]; then
    print_validation_error "The server CA X.509 certificate file in the specified path ${PGBOUNCER_SERVER_TLS_CA_FILE} does not exist"
fi


if [ ! -f ${PG_CONFIG_DIR}/pgbouncer.ini ]; then
    echo "Create pgbouncer config in ${PG_CONFIG_DIR}"

    # Config file is in “ini” format. Section names are between “[” and “]”.
    # Lines starting with “;” or “#” are taken as comments and ignored.
    # The characters “;” and “#” are not recognized when they appear later in the line.
    # shellcheck disable=SC2059
    printf "\
    ################## Auto generated ##################
    [databases]
    *=host=${POSTGRESQL_HOST:?"Setup pgbouncer config error! You must set DB_HOST env"} \
    port=${POSTGRESQL_PORT:-5432} user=${POSTGRESQL_USERNAME:-postgres}

    [pgbouncer]
    listen_addr = ${LISTEN_ADDR:-0.0.0.0}
    listen_port = ${LISTEN_PORT:-5432}
    auth_file = ${AUTH_FILE:-$PG_CONFIG_DIR/userlist.txt}
    auth_type = ${AUTH_TYPE:-md5}
    ${PGBOUNCER_POOL_MODE:+pool_mode = ${PGBOUNCER_POOL_MODE}\n}\
    ${PGBOUNCER_MAX_CLIENT_CONN:+max_client_conn = ${PGBOUNCER_MAX_CLIENT_CONN}\n}\
    ${PGBOUNCER_IGNORE_STARTUP_PARAMETERS:+ignore_startup_parameters = ${PGBOUNCER_IGNORE_STARTUP_PARAMETERS}\n}\
    admin_users = ${POSTGRESQL_USERNAME:-postgres}
    ${PGBOUNCER_IDLE_TRANSACTION_TIMEOUT:+idle_transaction_timeout = ${PGBOUNCER_IDLE_TRANSACTION_TIMEOUT}\n}\

    # TLS settings
    ${PGBOUNCER_SERVER_TLS_SSLMODE:+server_tls_sslmode = ${PGBOUNCER_SERVER_TLS_SSLMODE}\n}\
    ${PGBOUNCER_SERVER_TLS_CA_FILE:+server_tls_ca_file = ${PGBOUNCER_SERVER_TLS_CA_FILE}\n}\
    ${PGBOUNCER_SERVER_TLS_KEY_FILE:+server_tls_key_file = ${PGBOUNCER_SERVER_TLS_KEY_FILE}\n}\
    ${PGBOUNCER_SERVER_TLS_CERT_FILE:+server_tls_cert_file = ${PGBOUNCER_SERVER_TLS_CERT_FILE}\n}\
    ################## end file ##################
    " > ${PG_CONFIG_DIR}/pgbouncer.ini
    cat ${PG_CONFIG_DIR}/pgbouncer.ini
    echo "Starting $*..."
fi

exec "$@"