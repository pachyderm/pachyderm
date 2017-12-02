#!/bin/bash

function dump() {
	date
	
	kubectl get all

    kubectl --namespace=kube-system get pod
    ls /etc/kubernetes/pki/ || true
    kubectl --namespace=kube-system get pod | grep apiserver | cut -f 1 -d " " | while read pod; do kubectl --namespace=kube-system exec $pod -- ps -a; done
    
    echo 'dns logs'
    kubectl --namespace=kube-system get pod | grep dns | cut -f 1 -d " " | while read pod; do kubectl --namespace=kube-system logs po/$pod kubedns; done
	
	# Get node status
	# kubectl get node | tail -n 1 | cut -f 1 -d " " | while read node; do kubectl get node/$node -o yaml; done
	
    # too verbose for now
	kubectl describe pod -l app=pachd
    echo 'pachd logs:'
	kubectl logs -l app=pachd

    minikube logs
}

while true; do
	dump
	sleep 10
done
