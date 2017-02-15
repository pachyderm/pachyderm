#!/bin/bash

set -euxo pipefail

deploy_k8s_on_aws() {
    # Check prereqs
    
    which aws
    aws configure
    aws iam list-users
    which jq
    
    # Setup K8S DNS hosted zone
    
    # Common config
    
    export AWS_REGION=us-east-1
    export AWS_AVAILABILITY_ZONE=us-east-1a
    export STATE_STORE_NAME=k8scom-state-store-pachyderm-${RANDOM}
    
    # Setup basic kops dependencies
    
    aws s3api create-bucket --bucket $STATE_STORE_NAME --region $AWS_REGION
    
    # KOPS options
    
    export NAME=bpachydermcluster.kubernetes.com
    export KOPS_STATE_STORE=s3://$STATE_STORE_NAME
    export NODE_SIZE=r4.xlarge
    export MASTER_SIZE=r4.xlarge
    export NUM_NODES=3
    
    kops create cluster \
        --node-count $NUM_NODES \
        --zones $AWS_AVAILABILITY_ZONE \
        --master-zones $AWS_AVAILABILITY_ZONE \
        --dns private \
        --dns-zone kubernetes.com \
        --node-size $NODE_SIZE \
        --master-size $NODE_SIZE \
        $NAME
    
    kops update cluster $NAME --yes --state $KOPS_STATE_STORE
    
    
    
    # NON AUTOMATIC PART
    
    # kubectl context now set to 'apachydermcluster.kubernetes.com'
    # which tries to connect to the k8s server at https://api.apachydermcluster.kubernetes.com
    # which of course doesn't resolve
    
    # So I just used an etc/hosts hack:
    # 107.22.153.120 api.apachydermcluster.kubernetes.com
    
    # and now its up!
}



##################################3333
########## Deploy Pach cluster
######

deploy_pachyderm_on_aws() {

    # got IP from etc hosts file / master external IP on aws console
    KUBECTLFLAGS="-s 107.22.153.120"
    
    # top 2 shared w k8s deploy script
    export AWS_REGION=us-east-1
    export AWS_AVAILABILITY_ZONE=us-east-1a
    export STORAGE_SIZE=100
    export BUCKET_NAME=fdy-pachyderm-norm-test41
    
    # Omit location constraint if us-east
    # aws s3api create-bucket --bucket ${BUCKET_NAME} --region ${AWS_REGION} --create-bucket-configuration LocationConstraint=${AWS_REGION}
    
    aws s3api create-bucket --bucket ${BUCKET_NAME} --region ${AWS_REGION}
    
    STORAGE_NAME=`aws ec2 create-volume --size ${STORAGE_SIZE} --region ${AWS_REGION} --availability-zone ${AWS_AVAILABILITY_ZONE} --volume-type gp2 | grep VolumeId | cut -d "\"" -f 4`
    
    ## Redeploying a cluster, so need to reuse the vol name
    #export STORAGE_NAME="vol-064e19d4dc8042d04"
    #export STORAGE_NAME="vol-01d86f7d128981b6d"
    #export STORAGE_NAME="vol-06872bcc37b8367e3"
    #export STORAGE_NAME="vol-0b136c33257a51036"
    
    echo "volume storage: ${STORAGE_NAME}"
    
    # Since my user should have the right access:
    AWS_KEY=`cat ~/.aws/credentials | head -n 2 | tail -n 1 | cut -d " " -f 3`
    AWS_ID=`cat ~/.aws/credentials | tail -n 1 | cut -d " " -f 3`
    
    
    # Omit token since im using my personal creds
    # pachctl deploy amazon ${BUCKET_NAME} ${AWS_ID} ${AWS_KEY} ${AWS_TOKEN} ${AWS_REGION} ${STORAGE_NAME} ${STORAGE_SIZE}
    # Also ... need to escape them in quotes ... or pachctl will complain only 6 args not 7 (and some special chars in tokesn are bad)
    pachctl deploy amazon ${BUCKET_NAME} "${AWS_ID}" "${AWS_KEY}" " " ${AWS_REGION} ${STORAGE_NAME} ${STORAGE_SIZE}

}
