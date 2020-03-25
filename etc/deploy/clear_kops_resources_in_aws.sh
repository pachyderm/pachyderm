#!/bin/bash
# This script deletes all of the resources and registrations created by kops in a given region.
# Note that it terminates all instances and deletes all volumes in the region, so it's not
# restricted to only kops clusters nor is it generally safe. I'm preserving it as a reference,
# though, so that if we're ever trying to clean up after a kops cluster again, and we need to
# figure out what to delete, we'll have a place to start.

# Parse flags
eval "set -- $( getopt -l "zone:" "--" "${0}" "${@}" )"
while true; do
    case "${1}" in
        --zone)
          export AWS_AVAILABILITY_ZONE="${2}"
          len_zone_minus_one="$(( ${#AWS_AVAILABILITY_ZONE} - 1 ))"
          export AWS_REGION=${AWS_AVAILABILITY_ZONE:0:${len_zone_minus_one}}
          shift 2
          ;;
        --)
          shift
          break
          ;;
    esac
done

if [[ -z "${AWS_AVAILABILITY_ZONE}"  ]]; then
  echo "No aws availability zone set, must set --zone"
  echo "Exiting to be safe..."
  exit 1
fi

# Delete autoscaling group
aws --region="${AWS_REGION}" autoscaling describe-auto-scaling-groups \
    | jq --raw-output '.AutoScalingGroups[].AutoScalingGroupName' \
    | while read -r g; do \
        echo "${g}"
        aws --region="${AWS_REGION}" autoscaling delete-auto-scaling-group --auto-scaling-group-name="${g}" --force-delete; \
      done

# Delete launch configurations
aws --region="${AWS_REGION}" autoscaling describe-launch-configurations \
    | jq --raw-output '.LaunchConfigurations[].LaunchConfigurationName' \
    | grep 'pachydermcluster' \
    | while read -r lc; do \
        echo "${lc}"; \
        aws --region="${AWS_REGION}" autoscaling delete-launch-configuration --launch-configuration-name="${lc}"; \
      done

# Wait until all instances are terminated
WHEEL='\|/-'
spin() {
  echo -n "${WHEEL:0:1}"
  WHEEL="${WHEEL:1}${WHEEL:0:1}"
}

first="true"
while true; do
  group_count="$(
    aws --region="${AWS_REGION}" ec2 describe-instances \
      | jq --raw-output '.Reservations[].Instances[] | select(.State.Name != "terminated") | .InstanceId' \
      | wc -l
  )"
  if [[ "${group_count}" -eq 0 ]]; then
    echo "All instances deleted"
    break
  fi
  [[ "${first}" ]] && first="" || echo -en "\e[F"
  spin
  echo " ${group_count} instances remaining"
  sleep 1
done

# Delete volumes
aws --region="${AWS_REGION}" ec2 describe-volumes \
  | jq --raw-output .Volumes[].VolumeId \
  | while read -r v; do \
      echo "${v}"; \
      aws --region="${AWS_REGION}" ec2 delete-volume --volume-id="${v}"; \
done

# Delete ELBs
aws --region="${AWS_REGION}" elb describe-load-balancers \
  | jq --raw-output '.LoadBalancerDescriptions[].LoadBalancerName' \
  | while read -r elb; do
      cmd=( aws elb delete-load-balancer "--load-balancer-name=${elb}" )
      echo "${cmd[@]}"
      "${cmd[@]}"
done

# Delete routes
aws --region="${AWS_REGION}" ec2 describe-route-tables \
  | jq --raw-output '.RouteTables[].RouteTableId' \
  | while read -r id; do \
      aws ec2 describe-route-tables --route-table-ids="${id}" \
        | jq --raw-output '.RouteTables[].Routes[] | select(.GatewayId != "local") | .DestinationCidrBlock' \
        | while read -r cidr; do \
            echo aws ec2 delete-route --route-table-id="${id}" --destination-cidr-block="${cidr}"
      done
done

# Delete subnets
aws --region="${AWS_REGION}" ec2 describe-subnets \
  | jq --raw-output '.Subnets[].SubnetId' \
  | while read -r s; \
      do echo "${s}"; \
      aws --region="${AWS_REGION}" ec2 delete-subnet --subnet-id="${s}"; \
done

# Delete any routing tables that aren't the main routing table for their VPC
aws --region="${AWS_REGION}" ec2 describe-route-tables \
  | jq --raw-output '.RouteTables[]
                     | select([ .Associations[].Main | not ] | all)
                     | .RouteTableId' \
  | while read -r t; do \
      echo "${t}"; \
      aws --region="${AWS_REGION}" ec2 delete-route-table --route-table-id="${t}"; \
done

# Delete key pairs
aws --region="${AWS_REGION}" ec2 describe-key-pairs \
  | jq --raw-output '.KeyPairs[].KeyName' \
  | while read -r k; do \
      echo "${k}"; \
      aws --region="${AWS_REGION}" ec2 delete-key-pair --key-name="${k}"; \
done

# Detach internet gateways from their respective VPC, then delete them
aws --region="${AWS_REGION}" ec2 describe-internet-gateways \
  | jq --raw-output '.InternetGateways[].InternetGatewayId' \
  | while read -r g; do \
      echo "${g}"; \
      aws --region="${AWS_REGION}" ec2 detach-internet-gateway \
        --internet-gateway-id="${g}" \
        --vpc-id="$( \
            aws --region="${AWS_REGION}" ec2 describe-internet-gateways \
                --internet-gateway-id="${g}" \
              | jq --raw-output '.InternetGateways[].Attachments[0].VpcId'\
          )"; \
done

aws --region="${AWS_REGION}" ec2 describe-internet-gateways \
  | jq --raw-output '.InternetGateways[].InternetGatewayId' \
  | while read -r g; do \
      echo "${g}"; \
      aws --region="${AWS_REGION}" ec2 delete-internet-gateway --internet-gateway-id="${g}"; \
done

# Clear all security group ACLs (so that no security group is referenced by
# another security group) and then delete all non-default security groups
aws --region="${AWS_REGION}" ec2 describe-security-groups \
  | jq --raw-output '.SecurityGroups[].GroupId' \
  | while read -r sg; do \
      aws --region="${AWS_REGION}" ec2 describe-security-groups --group-id="${sg}" \
        | jq --raw-output '.SecurityGroups[0].IpPermissions[].UserIdGroupPairs[].GroupId' \
        | while read -r other; do \
            echo "${other} -> ${sg}"; \
            aws --region="${AWS_REGION}" ec2 revoke-security-group-ingress --group-id="${sg}" --source-group="${other}" --protocol=all; \
      done; \
done

aws --region="${AWS_REGION}" ec2 describe-security-groups \
  | jq --raw-output '.SecurityGroups[] | select(.GroupName != "default") | .GroupId' \
  | while read -r sg; do \
      echo "${sg}"; \
      aws --region="${AWS_REGION}" ec2 delete-security-group --group-id="${sg}"; \
done

# Delete all VPCs
aws --region="${AWS_REGION}" ec2 describe-vpcs \
  | jq --raw-output .Vpcs[].VpcId \
  | while read -r v; do \
      echo "${v}"; \
      aws --region="${AWS_REGION}" ec2 delete-vpc --vpc-id="${v}"; \
done

# Delete all VPC associations with the hosted zone 'kubernetes.com'
zone_name=kubernetes.com # trailing period is omitted (here, not below)
zone_id="$( aws route53 list-hosted-zones | jq --raw-output '.HostedZones[] | select(.Name == "'${zone_name}'.") | .Id' )"
aws route53 get-hosted-zone  --id="${zone_id}" \
  | jq -c --monochrome-output ".VPCs[] | select(.VPCRegion == \"${AWS_REGION}\")" \
  | while read -r vpc; do \
      echo "${zone_id} -> ${vpc}"; \
      aws route53 disassociate-vpc-from-hosted-zone --hosted-zone-id="${zone_id}" --vpc="${vpc}"; \
    done

# Delete all route53 records in 'kubernetes.com'
aws route53 list-resource-record-sets \
  --hosted-zone-id="${zone_id}" \
  | jq -c --monochrome-output '.ResourceRecordSets[] | select(.Name | test("pachydermcluster\\.kubernetes\\.com"))' \
  | while read -r rs; do \
      echo "${rs}"; \
      aws route53 change-resource-record-sets \
        --hosted-zone-id="${zone_id}" \
        --change-batch="{
            \"Changes\":[{\"Action\":\"DELETE\",\"ResourceRecordSet\":\"${rs}\"}]
          }"; \
done

# Delete elastic IPs
aws ec2 describe-addresses \
  | jq -c '.Addresses[] | if .Domain == "vpc" then { AllocationId: .AllocationId } else { PublicIp: .PublicIp } end' \
  | while read -r eip; do
      aws ec2 release-address --cli-input-json="${eip}"
    done
