#!/bin/bash

url=$(aws cloudfront sign --url https://d3q2p4tuagkx3p.cloudfront.net/pach/block/7249a7d412dc4a97bbb851b262afb752 --key-pair-id APKAIQZFQIDVCMO6YHXA --private-key file://cf-keypair-private.pem --date-less-than 2017-09-01)

curl -v ''"$url"''
