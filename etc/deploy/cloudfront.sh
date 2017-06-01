#!/bin/bash

set -euxo pipefail

which aws
aws configure list > /dev/null
aws iam list-users

# Creates the distribution and proper error caching rules, and updates the underlying bucket policy
create_distribution() {
    
}

# Updates the bucket policy for an S3 bucket
# To restrict access to only cloudfront
create_origin_access_identity() {

}

# Generates the signed
generate_signed_cookie() {

}
