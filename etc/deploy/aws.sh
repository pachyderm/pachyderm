#!/bin/bash

set -euxo pipefail

deploy_k8s_on_aws() {
    # Check prereqs
    
    which aws
    aws configure list
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
    set +euxo pipefail
    mkdir tmp
    echo "KOPS_STATE_STORE=$KOPS_STATE_STORE" >> tmp/$NAME.sh
    echo $KOPS_STATE_STORE > tmp/current-benchmark-state-store.txt
    echo $NAME > tmp/current-benchmark-cluster.txt
    set -euxo pipefail

    wait_for_k8s_master_ip
    update_sec_group
    wait_for_nodes_to_come_online

}

update_sec_group() {
    export sec_group_id=`cat tmp/$NAME.instances.json | jq ".Reservations | .[] | .Instances | .[] | select( .Tags | .[]? | select( .Value | contains(\"masters.$NAME\") ) ) | .SecurityGroups[0].GroupId" | head -n 1 | cut -f 2 -d "\""`
    # Note - groupname may be sufficient and is just master.$NAME
    #aws ec2 authorize-security-group-ingress --group-id $sec_group_id --protocol tcp --port 30650 --cidr "0.0.0.0/0"

    # For k8s access
    #aws ec2 authorize-security-group-ingress --group-name "masters.$NAME" --protocol tcp --port 8080 --cidr "0.0.0.0/0" --region $AWS_REGION
    aws ec2 authorize-security-group-ingress --group-id $sec_group_id --protocol tcp --port 8080 --cidr "0.0.0.0/0" --region $AWS_REGION
    # For pachyderm direct access:
    aws ec2 authorize-security-group-ingress --group-id $sec_group_id --protocol tcp --port 30650 --cidr "0.0.0.0/0" --region $AWS_REGION
}

wait_for_k8s_master_ip() {
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
    # This is the only operation that requires sudo privileges
    sudo echo " " >> /etc/hosts # Some files dont contain newlines ... I'm looking at you travisCI
    sudo echo "$masterk8sip api.${NAME}" >> /etc/hosts
    echo "state of /etc/hosts:"
    cat /etc/hosts
}

wait_for_nodes_to_come_online() {
    # Wait until all nodes show as ready, and we have as many as we expect
    set +euxo pipefail
    check_all_nodes_ready
    while [ $? -ne 0 ]; do
        sleep 1
        check_all_nodes_ready
    done
    set -euxo pipefail
    rm nodes.txt
}

check_all_nodes_ready() {
    echo "Checking k8s nodes are ready"
    kubectl get nodes > nodes.txt
    if [ $? -ne 0 ]; then
        return 1
    fi

    master=`cat nodes.txt | grep master | wc -l`
    if [ $master != "1" ]; then
        echo "no master nodes found"
        return 1
    fi

    NUM_NODES=3
    TOTAL_NODES=$(($NUM_NODES+1))
    ready_nodes=`cat nodes.txt | grep -v NotReady | grep Ready | wc -l`
    echo "total $TOTAL_NODES, ready $ready_nodes"
    if [ $ready_nodes == $TOTAL_NODES ]; then
        echo "all nodes ready"
        return 0
    fi
    return 1
}

get_k8s_master_domain() {
    sleep 1
    echo "Retrieving ec2 instance list to get k8s master domain name"
    aws ec2 describe-instances --filters "Name=instance-type,Values=${NODE_SIZE}" --region $AWS_REGION > tmp/$NAME.instances.json
    export masterk8sdomain=`cat tmp/$NAME.instances.json | jq ".Reservations | .[] | .Instances | .[] | select( .Tags | .[]? | select( .Value | contains(\"masters.$NAME\") ) ) | .PublicDnsName" | head -n 1 | cut -f 2 -d "\""`
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
    export STORAGE_SIZE=100
    export BUCKET_NAME=${RANDOM}-pachyderm-store
    
    # Omit location constraint if us-east
    # TODO - check the $AWS_REGION value and Do The Right Thing
    # aws s3api create-bucket --bucket ${BUCKET_NAME} --region ${AWS_REGION} --create-bucket-configuration LocationConstraint=${AWS_REGION}
    
    aws s3api create-bucket --bucket ${BUCKET_NAME} --region ${AWS_REGION}
    
    # Since my user should have the right access:
    AWS_KEY=`cat ~/.aws/credentials | grep aws_secret_access_key | cut -d " " -f 3`
    AWS_ID=`cat ~/.aws/credentials | grep aws_access_key_id  | cut -d " " -f 3`

    # Omit token since im using my personal creds
    pachctl deploy amazon ${BUCKET_NAME} "${AWS_ID}" "${AWS_KEY}" " " ${AWS_REGION} ${STORAGE_SIZE} --dynamic-etcd-nodes=3
}

if [ "$EUID" -ne 0 ]; then
  echo "Cowardly refusing to deploy cluster. Please run as root"
  echo "Please run this command like 'sudo -E make launch-bench'"
  exit 1
fi

set +euxo pipefail
which pachctl
if [ $? -ne 0 ]; then
    echo "pachctl not found on path"
    exit 1
fi
set -euxo pipefail

deploy_k8s_on_aws
deploy_pachyderm_on_aws
