#!/bin/sh

pachctl update pipeline --jsonnet ./src/templates/sql_ingest_cron.jsonnet \
	--arg name=myingest \
	--arg url="mysql://root@mysql:3306/test_db" \
	--arg query="SELECT * FROM test_data" \
	--arg cronSpec="@every 30s" \
	--arg secretName="mysql-creds" \
	--arg format=json
