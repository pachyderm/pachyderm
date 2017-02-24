#!/bin/bash

# Check the cron value to see if this is a daily job
if [ "$TRAVIS_EVENT_TYPE" == "cron" ]; then
	echo "Running daily benchmarks"
    # Use the secrets in the travis environment to setup the aws creds for the aws command:
    echo "${AWS_ACCESS_KEY_ID}
${AWS_SECRET_ACCESS_KEY}

" | aws configure
	sudo -E make bench
else
	echo "Running tests"
	make test
fi
