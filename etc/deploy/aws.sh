#!/bin/bash

set -euxo pipefail

deploy_k8s_on_aws() {
    # Check prereqs
    
    which aws
    aws configure
    aws iam list-users
    which jq
    which uuid
    
    # Setup K8S DNS hosted zone
    
    # Common config
    
    export AWS_REGION=us-east-1
    export AWS_AVAILABILITY_ZONE=us-east-1a
    export STATE_STORE_NAME=k8scom-state-store-pachyderm-${RANDOM}
    
    # Setup basic kops dependencies
    
    aws s3api create-bucket --bucket $STATE_STORE_NAME --region $AWS_REGION
    
    # KOPS options
    
    export NAME=`uuid | cut -f 1 -d "-"`-pachydermcluster.kubernetes.com
    export KOPS_STATE_STORE=s3://$STATE_STORE_NAME
    echo "kops state store: $KOPS_STATE_STORE"
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
    
    # This will allow us to cleanup the cluster afterwards
    # ... if we don't do it explicitly its annoying to kill the cluster
    echo "KOPS_STATE_STORE=$KOPS_STATE_STORE" >> $NAME.sh
    echo $KOPS_STATE_STORE > current-benchmark-state-store.txt
    echo $NAME > current-benchmark-cluster.txt

    # Get the IP of the k8s master node and hack /etc/hosts so we can connect
    # Need to retry this in a loop until we see the instance appear

    set +euxo pipefail
    get_k8s_master_domain
    while [ $? -ne 0 ]; do
        get_k8s_master_domain
    done
    echo "Master k8s node is up and lives at $masterk8sdomain"
    set -euxo pipefail
    masterk8sip=`dig +short $masterk8sdomain`
    sudo echo "$masterk8sip api.${NAME}" >> /etc/hosts

    kubectl get nodes --show-labels

    # Wait until all nodes show as ready, and we have as many as we expect
}

get_k8s_master_domain() {
    sleep 1
    echo "Retrieving ec2 instance list to get k8s master domain name"
    aws ec2 describe-instances --filters "Name=instance-type,Values=${NODE_SIZE}" --region $AWS_REGION > $NAME.instances.json
    export masterk8sdomain=`cat $NAME.instances.json | jq ".Reservations | .[] | .Instances | .[] | select( .Tags | .[]? | select( .Value | contains(\"masters.$NAME\") ) ) | .PublicDnsName" | head -n 1 | cut -f 2 -d "\""`
    if [ -n "$masterk8sdomain" ]; then
        return 0
    fi
    return 1
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


deploy_k8s_on_aws
