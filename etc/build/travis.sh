#!/bin/bash

# Check the cron value to see if this is a daily job
if [ "$TRAVIS_EVENT_TYPE" == "cron" ]; then
	echo "Running daily benchmarks"
	make bench
else
	echo "Running tests"
	make test
fi
