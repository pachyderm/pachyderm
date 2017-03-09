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
    echo "pachctl is installed here:"
    which pachctl
    # Travis doesn't come w an ssh key
    # kops needs one in place (because it enables ssh access to nodes w it)
    # for now we'll just generate one on the fly
    # travis supports adding a persistent one if we pay: https://docs.travis-ci.com/user/private-dependencies/#Generating-a-new-key

    echo '                                              


' | ssh-keygen -t rsa -b 4096 -C "buildbot@pachyderm.io"


    echo "generated ssh keys:"
    ls ~/.ssh
    cat ~/.ssh/id_rsa.pub

	sudo -E make bench
else
	echo "Running tests"
	make test
fi
