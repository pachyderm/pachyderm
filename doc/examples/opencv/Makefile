# This Makefile is a reference for setting up and running various parts of this demo quickly

# Try to find the pachyderm files we'll need
data_dir := $(strip $(if $(shell ls ./edges.json 2>/dev/null), "."))
data_dir := $(strip $(if $(data_dir), $(data_dir), $(shell ls -d $${HOME}/pachyderm/doc/examples/opencv 2>/dev/null)))
data_dir := $(strip $(if $(data_dir), $(data_dir), $(shell ls -d $${HOME}/go/src/github.com/pachyderm/pachyderm/doc/examples/opencv 2>/dev/null)))

echo:
	echo "$(data_dir)"

opencv:
	pachctl create-repo images
	pachctl create-pipeline -f $(data_dir)/edges.json
	pachctl create-pipeline -f $(data_dir)/montage.json
	pachctl put-file images master -i $(data_dir)/images.txt
	pachctl put-file images master -i $(data_dir)/images2.txt

delete:
	yes | pachctl delete-all

minikube:
	minikube start
	@WHEEL="-\|/"; \
	until minikube ip 2>/dev/null; do \
	    WHEEL=$${WHEEL:1}$${WHEEL:0:1}; \
	    echo -en "\e[G\e[K$${WHEEL:0:1} waiting for minikube to start..."; \
	    sleep 1; \
	done
	pachctl deploy local
	@until "$$(
	    kc get po \
	      -l suite=pachyderm,app=dash \
	      -o jsonpath='{.items[0].status.conditions[?(@.type=="Ready")].status}')" = True; \
	do \
	    WHEEL=$${WHEEL:1}$${WHEEL:0:1}; \
	    echo -en "\e[G\e[K$${WHEEL:0:1} waiting for pachyderm to start..."; \
	    sleep 1; \
	done
	pachctl port-forward &

.PHONY: \
	echo \
	opencv \
	delete \
	minikube

