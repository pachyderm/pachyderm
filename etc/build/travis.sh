#!/bin/bash

# Check the cron value to see if this is a daily job

# for now always run bench since thats what I'm trying to test
TRAVIS_EVENT_TYPE=cron

if [ "$TRAVIS_EVENT_TYPE" == "cron" ]; then
	echo "Running daily benchmarks"
    # Use the secrets in the travis environment to setup the aws creds for the aws command:
    echo "${AWS_ACCESS_KEY_ID}
${AWS_SECRET_ACCESS_KEY}

" | aws configure
    make install
	sudo -E make bench
else
	echo "Running tests"
	make test
fi
