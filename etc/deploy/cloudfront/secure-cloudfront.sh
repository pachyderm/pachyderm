#!/bin/bash

set -euxo pipefail

parse_flags() {
  # Check prereqs
  which aws
  which jq
  # Common config
  export AWS_REGION=us-east-1
  export AWS_AVAILABILITY_ZONE=us-east-1a

  # Parse flags
  eval "set -- $( getopt -l "state:,region:,zone:,bucket:,distribution:" "--" "${0}" "${@}" )"
  while true; do
      case "${1}" in
          --region)
            export AWS_REGION="${2}"
            shift 2
            ;;
          --zone)
            export AWS_AVAILABILITY_ZONE="${2}"
            shift 2
            ;;
          --bucket)
            export BUCKET="${2}"
            shift 2
            ;;
          --distribution)
            export DISTRIBUTION="${2}"
            shift 2
            ;;
          --)
            shift
            break
            ;;
      esac
  done

  echo "Region: ${AWS_REGION}"
  zone_suffix=${AWS_AVAILABILITY_ZONE#$AWS_REGION}
  if [[ ${#zone_suffix} -gt 3 ]]; then
    echo "Availability zone \"${AWS_AVAILABILITY_ZONE}\" may not be in region \"${AWS_REGION}\""
    echo "Try setting both --region and --zone"
    echo "Exiting to be safe..."
    exit 1
  fi
}


# Update bucket access policy

update_bucket_policy() {
    aws s3api get-bucket-policy --bucket $BUCKET --region $AWS_REGION > tmp/existing-bucket-policy.json
    aws s3api delete-bucket-policy --bucket $BUCKET --region $AWS_REGION
   
    sed 's/XXCallerReferenceXX/'$someuuid'/' etc/deploy/cloudfront/origin-access-identity.json.template > tmp/cloudfront-origin-access-identity.json
    sed -i 's/XXBucketNameXX/'$BUCKET'/' tmp/cloudfront-origin-access-identity.json
    aws cloudfront create-cloud-front-origin-access-identity --cloud-front-origin-access-identity-config file://tmp/cloudfront-origin-access-identity.json > tmp/cloudfront-origin-access-identity-info.json
    CLOUDFRONT_OAI_CANONICAL=$(cat tmp/cloudfront-origin-access-identity-info.json | jq -r ".CloudFrontOriginAccessIdentity.S3CanonicalUserId")
    echo "Got Cloudfront Origin Access Identity Canonical user id : ${CLOUDFRONT_OAI_CANONICAL}"
    CLOUDFRONT_OAI_ID=$(cat tmp/cloudfront-origin-access-identity-info.json | jq -r ".CloudFrontOriginAccessIdentity.Id")
    echo "Got Cloudfront Origin Access Identity ID : ${CLOUDFRONT_OAI_ID}"
  
    sed 's/XXBUCKET_NAMEXX/'$BUCKET'/' etc/deploy/cloudfront/bucket-policy-secure.json.template > tmp/bucket-policy-secure.json
    sed -i 's/XXCLOUDFRONT_OAI_CANONICALXX/'$CLOUDFRONT_OAI_CANONICAL'/' tmp/bucket-policy-secure.json
    aws s3api put-bucket-policy --bucket $BUCKET --policy file://tmp/bucket-policy-secure.json --region=${AWS_REGION}

}

# Update cloudfront distribution

#  sed 's/XXCallerReferenceXX/'$someuuid'/' etc/deploy/cloudfront/origin-access-identity.json.template > tmp/cloudfront-origin-access-identity.json
#  sed -i 's/XXBucketNameXX/'$BUCKET'/' tmp/cloudfront-origin-access-identity.json
#  aws cloudfront create-cloud-front-origin-access-identity --cloud-front-origin-access-identity-config file://tmp/cloudfront-origin-access-identity.json > tmp/cloudfront-origin-access-identity-info.json
#  CLOUDFRONT_OAI_CANONICAL=$(cat tmp/cloudfront-origin-access-identity-info.json | jq -r ".CloudFrontOriginAccessIdentity.S3CanonicalUserId")
#  echo "Got Cloudfront Origin Access Identity Canonical user id : ${CLOUDFRONT_OAI_CANONICAL}"
#  CLOUDFRONT_OAI_ID=$(cat tmp/cloudfront-origin-access-identity-info.json | jq -r ".CloudFrontOriginAccessIdentity.Id")
#  echo "Got Cloudfront Origin Access Identity ID : ${CLOUDFRONT_OAI_ID}"
#
#  sed 's/XXCallerReferenceXX/'$someuuid'/' etc/deploy/cloudfront/distribution.json.template > tmp/cloudfront-distribution.json
#  sed -i 's/XXBucketNameXX/'$BUCKET'/' tmp/cloudfront-distribution.json
#
#  sed -i 's/XXCLOUDFRONT_OAI_IDXX/'$CLOUDFRONT_OAI_ID'/' tmp/cloudfront-distribution.json
#  aws cloudfront create-distribution --distribution-config file://tmp/cloudfront-distribution.json > tmp/cloudfront-distribution-info.json
#  CLOUDFRONT_DOMAIN=$(cat tmp/cloudfront-distribution-info.json | jq -r ".Distribution.DomainName" | cut -f 1 -d .)
#  echo "Setup cloudfront distribution with domain ${CLOUDFRONT_DOMAIN}"
#
#  sed 's/XXBUCKET_NAMEXX/'$BUCKET'/' etc/deploy/cloudfront/bucket-policy.json.template > tmp/bucket-policy.json
#  sed -i 's/XXCLOUDFRONT_OAI_CANONICALXX/'$CLOUDFRONT_OAI_CANONICAL'/' tmp/bucket-policy.json
#  aws s3api put-bucket-policy --bucket $BUCKET --policy file://tmp/bucket-policy.json --region=${AWS_REGION}





parse_flags "${@}"
update_bucket_policy
