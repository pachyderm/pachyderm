#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE DATABASE dex;
    GRANT ALL PRIVILEGES ON DATABASE dex TO postgres;
EOSQL
