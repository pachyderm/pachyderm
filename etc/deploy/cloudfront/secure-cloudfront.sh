#!/bin/bash

CLOUDFRONT_OAI_ID=
set -euxo pipefail

parse_flags() {
  # Check prereqs
  command -v aws
  command -v jq
  command -v uuid
  # Common config
  export AWS_REGION=us-east-1
  export AWS_AVAILABILITY_ZONE=us-east-1a

  # Parse flags
  eval "set -- $( getopt -l "state:,region:,zone:,bucket:,cloudfront-distribution-id:,cloudfront-keypair-id:,cloudfront-private-key-file:" "--" "${0}" "${@}" )"
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
          --cloudfront-distribution-id)
            export CLOUDFRONT_DISTRIBUTION_ID="${2}"
            shift 2
            ;;
          --cloudfront-keypair-id)
            export CLOUDFRONT_KEYPAIR_ID="${2}"
            shift 2
            ;;
          --cloudfront-private-key-file)
            export CLOUDFRONT_PRIVATE_KEY_FILE="${2}"
            shift 2
            ;;
          --)
            shift
            break
            ;;
      esac
  done


  set +euxo pipefail

  if [ -z "$CLOUDFRONT_KEYPAIR_ID" ]; then
	echo "--cloudfront-keypair-id must be set"
	exit 1	
  fi

  if [ -z "$CLOUDFRONT_PRIVATE_KEY_FILE" ]; then
	echo "--cloudfront-private-key-file must be set"
	exit 1	
  fi

  if [ -z "$BUCKET" ]; then
	echo "--bucket must be set"
	exit 1	
  fi

  if [ -z "$CLOUDFRONT_DISTRIBUTION_ID" ]; then
	echo "--cloudfront-distribution-id must be set"
	exit 1	
  fi

  set -euxo pipefail

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
    aws s3api delete-bucket-policy --bucket "$BUCKET" --region "$AWS_REGION"
   
    # Create Origin Access Identity
    someuuid=$(uuid | cut -f 1 -d-)
    mkdir -p tmp
    jq ".CallerReference = \"$someuuid\" | .Comment = \"$BUCKET auto generated OAI\"" < etc/deploy/cloudfront/origin-access-identity.json.template > tmp/cloudfront-origin-access-identity.json
    aws cloudfront create-cloud-front-origin-access-identity --cloud-front-origin-access-identity-config file://tmp/cloudfront-origin-access-identity.json > tmp/cloudfront-origin-access-identity-info.json
    CLOUDFRONT_OAI_CANONICAL=$(jq -r ".CloudFrontOriginAccessIdentity.S3CanonicalUserId" < tmp/cloudfront-origin-access-identity-info.json)
    echo "Got Cloudfront Origin Access Identity Canonical user id : ${CLOUDFRONT_OAI_CANONICAL}"
    CLOUDFRONT_OAI_ID=$(jq -r ".CloudFrontOriginAccessIdentity.Id" < tmp/cloudfront-origin-access-identity-info.json)
    echo "Got Cloudfront Origin Access Identity ID : ${CLOUDFRONT_OAI_ID}"
  
    # Create secure bucket policy w the new OAI
    jq ".Statement[0].Resource = \"arn:aws:s3:::$BUCKET/*\" | .Statement[0].Principal.CanonicalUser = \"$CLOUDFRONT_OAI_CANONICAL\"" < etc/deploy/cloudfront/bucket-policy-secure.json.template > tmp/bucket-policy-secure.json
    aws s3api put-bucket-policy --bucket "$BUCKET" --policy file://tmp/bucket-policy-secure.json --region="${AWS_REGION}"

}

# Update cloudfront distribution

update_cloudfront_distribution() {

    # Get the ETAG (required to update)
    mkdir -p tmp
    aws cloudfront get-distribution --id "$CLOUDFRONT_DISTRIBUTION_ID" > tmp/existing-cloudfront-distribution.json
    ETAG=$(jq -r .ETag < tmp/existing-cloudfront-distribution.json)
    # Update the fields we need
    jq ".Distribution.DistributionConfig.Origins.Items[0].S3OriginConfig.OriginAccessIdentity = \"origin-access-identity/cloudfront/${CLOUDFRONT_OAI_ID}\"" < tmp/existing-cloudfront-distribution.json > tmp-result.json
    # There doesn't seem to be a way to use jq to replace 'in place':
    mv tmp-result.json tmp/updated-cloudfront-distribution.json
    jq '.Distribution.DistributionConfig.DefaultCacheBehavior.TrustedSigners = {"Enabled": true,"Items": ["self"],"Quantity":1}' < tmp/updated-cloudfront-distribution.json > tmp-result.json
    mv tmp-result.json tmp/updated-cloudfront-distribution.json

    jq .Distribution.DistributionConfig < tmp/updated-cloudfront-distribution.json > tmp/updated-cloudfront-distribution-config.json
    aws cloudfront update-distribution --id "$CLOUDFRONT_DISTRIBUTION_ID" --if-match "$ETAG" --distribution-config file://tmp/updated-cloudfront-distribution-config.json > tmp/updated-cloudfront-distribution-info.json
    # This isn't used anywhere in this script, but still is good to report to the user
    CLOUDFRONT_DOMAIN=$(jq -r ".Distribution.DomainName" < tmp/updated-cloudfront-distribution-info.json | cut -f 1 -d .)
    echo "Setup cloudfront distribution with domain ${CLOUDFRONT_DOMAIN}"
}

deploy_secrets() {
    # Update the amazon secret to include: cloudfront-keypair-id and cloudfront-private-key
    echo "Updating secrets for cluster:"
    kubectl cluster-info
    mkdir -p tmp
    kubectl get secrets/pachyderm-storage-secret -o json > tmp/existing-amazon-secrets.json
    ENCODED_KEYPAIR=$(echo "$CLOUDFRONT_KEYPAIR_ID" | base64 -w0)
    jq ".data.cloudfrontKeyPairId = \"$ENCODED_KEYPAIR\"" < tmp/existing-amazon-secrets.json > tmp-result.json
    mv tmp-result.json tmp/updated-amazon-secrets.json
    CLOUDFRONT_PRIVATE_KEY=$(base64 -w0 < "$CLOUDFRONT_PRIVATE_KEY_FILE")
    jq ".data.cloudfrontPrivateKey = \"$CLOUDFRONT_PRIVATE_KEY\"" < tmp/updated-amazon-secrets.json > tmp-result.json
    mv tmp-result.json tmp/updated-amazon-secrets.json
    kubectl replace -f tmp/updated-amazon-secrets.json
}

parse_flags "${@}"
update_bucket_policy
update_cloudfront_distribution
deploy_secrets


echo "Waiting on distribution to enter 'deployed' state"
echo "This is the last step of the script ... if it times out, thats ok. Just make sure that your CF distribution's status is 'Deployed' on the UI before you run anything on pachyderm"
date
# We need to wait for this, otherwise if you try and access objects while it's 'pending' CF redirects you to S3 and it'll cache the 307 redirect responses
aws cloudfront wait distribution-deployed --id "$CLOUDFRONT_DISTRIBUTION_ID"
echo "Cloudfront distribution ($CLOUDFRONT_DISTRIBUTION_ID) deployed"
date

