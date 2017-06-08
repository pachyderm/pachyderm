#!/bin/bash

#url=$(aws cloudfront sign --url https://d3q2p4tuagkx3p.cloudfront.net/pach/block/7249a7d412dc4a97bbb851b262afb752 --key-pair-id APKAIQZFQIDVCMO6YHXA --private-key file://cf-keypair-private.pem --date-less-than 2017-09-01)

url=$(aws cloudfront sign --url http://d1isn5su8jq4o0.cloudfront.net/pach/object/8c0f56e81a3b4b646d2b01e429a5fea494398d258c4a02b140cfdc75bd5b1841e78ce0ca05b03c48a50fd2969953b8b379d8ada9ce675cc74ebb8cb77d5c6af1 --key-pair-id APKAIQZFQIDVCMO6YHXA --private-key file://cf-keypair-private.pem --date-less-than 2017-09-01)

echo $url

curl -v ''"$url"''
