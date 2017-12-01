#!/bin/bash

function dump() {
	date
	
	kubectl get all

    kubectl --namespace=kube-system get pod
    ls /etc/kubernetes/pki/
    kubectl --namespace=kube-system get pod | grep apiserver | cut -f 1 -d " " | while read pod; do kubectl --namespace=kube-system exec $pod -- ps -a; done
    
	
	# Get node status
	kubectl get node | tail -n 1 | cut -f 1 -d " " | while read node; do kubectl get node/$node -o yaml; done
	
	kubectl describe pod -l app=pachd
	kubectl logs -l app=pachd
}

while true; do
	dump
	sleep 10
done
