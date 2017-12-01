#!/bin/bash

function dump() {
	date
	
	kubectl get all
	
	# Get node status
	kubectl get node | tail -n 1 | cut -f 1 -d " " | while read node; do kubectl get node/$node -o yaml; done
	
	kubectl describe pod -l app=pachd
	kubectl logs -l app=pachd
}

while true; do
	dump
	sleep 10
done
