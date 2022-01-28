#!/bin/sh
# This script can be run inside the mysql pod to setup the database
# kubectl exec -t mysql-0 -- /bin/bash -c $(cat <THIS SCRIPT>)
set -ve

MYSQL_HOST=127.0.0.1
MYSQL_USER=root
MYSQL_PASSWORD=root

kubectl exec -it mysql-0 -- mysql --host=$MYSQL_HOST --user=$MYSQL_USER --password=$MYSQL_PASSWORD --execute="CREATE DATABASE test_db;"

kubectl exec -it mysql-0 -- mysql --host=$MYSQL_HOST --user=$MYSQL_USER --password=$MYSQL_PASSWORD --database=test_db --execute="CREATE TABLE test_data (id SERIAL PRIMARY KEY, a VARCHAR(200));"

