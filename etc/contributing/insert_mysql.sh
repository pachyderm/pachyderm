#!/bin/sh
# This script inserts a row into the test_data table in the test_db database
set -ve

MYSQL_HOST=127.0.0.1
MYSQL_USER=root
MYSQL_PASSWORD=root

kubectl exec -it mysql-0 -- mysql --host=$MYSQL_HOST --user=$MYSQL_USER --password=$MYSQL_PASSWORD --database=test_db --execute='INSERT INTO test_data (a) VALUES ("hello world");'
 
