#!/bin/bash

CLOUDFRONT_OAI_ID=
set -euxo pipefail

parse_flags() {
  # Check prereqs
  which aws
  which jq
  which uuid
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
    aws s3api delete-bucket-policy --bucket $BUCKET --region $AWS_REGION
   
    someuuid=$(uuid | cut -f 1 -d-)
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

update_cloudfront_distribution() {

    # Get the ETAG (required to update)
    mkdir -p tmp
    aws cloudfront get-distribution --id $DISTRIBUTION  > tmp/existing-cloudfront-distribution.json
    ETAG=$(cat tmp/existing-cloudfront-distribution.json | jq -r .ETag)
    # Update the fields we need
    cat tmp/existing-cloudfront-distribution.json | jq '.Distribution.DistributionConfig.Origins.Items[0].S3OriginConfig.OriginAccessIdentity = "origin-access-identity/cloudfront/'"${CLOUDFRONT_OAI_ID}"'"' > tmp/updated-cloudfront-distribution.json
    #aws cloudfront update-distribution --id $DISTRIBUTION --if-match $ETAG --distribution-config file://tmp/cloudfront-distribution-secure.json > tmp/cloudfront-distribution-secure-info.json
    cat tmp/updated-cloudfront-distribution.json | jq .Distribution.DistributionConfig > tmp/updated-cloudfront-distribution-config.json
    aws cloudfront update-distribution --id $DISTRIBUTION --if-match $ETAG --distribution-config file://tmp/updated-cloudfront-distribution-config.json > tmp/updated-cloudfront-distribution-info.json
    # This isn't used anywhere in this script, but still is good to report to the user
    CLOUDFRONT_DOMAIN=$(cat tmp/updated-cloudfront-distribution-info.json | jq -r ".Distribution.DomainName" | cut -f 1 -d .)
    echo "Setup cloudfront distribution with domain ${CLOUDFRONT_DOMAIN}"
}

deploy_secrets() {
    # Update the amazon secret to include: cloudfront-keypair-id and cloudfront-private-key
    kubectl create -f newsecret.yaml
}


parse_flags "${@}"
update_bucket_policy
update_cloudfront_distribution
#deploy_secrets
