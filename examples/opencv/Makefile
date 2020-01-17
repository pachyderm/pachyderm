# This Makefile is a reference for setting up and running various parts of this demo quickly

# Try to find the pachyderm files we'll need
data_dir := $(strip $(if $(shell ls ./edges.json 2>/dev/null), "."))
data_dir := $(strip $(if $(data_dir), $(data_dir), $(shell ls -d $${HOME}/pachyderm/examples/opencv 2>/dev/null)))
data_dir := $(strip $(if $(data_dir), $(data_dir), $(shell ls -d $${HOME}/go/src/github.com/pachyderm/pachyderm/examples/opencv 2>/dev/null)))

echo:
	echo "$(data_dir)"

opencv:
	pachctl create repo images
	pachctl create pipeline -f $(data_dir)/edges.json
	pachctl create pipeline -f $(data_dir)/montage.json
	pachctl put file images@master -i $(data_dir)/images.txt
	pachctl put file images@master -i $(data_dir)/images2.txt

delete:
	yes | pachctl delete all

minikube:
	minikube start
	@echo -n "Waiting for minikube to start"; \
	until minikube ip 2>/dev/null; do \
	    echo -n "."; \
	    sleep 1; \
	done; \
	echo
	pachctl deploy local
	@echo -n "Waiting for pachyderm to start"; \
	until \
	    kubectl get po \
	      -l suite=pachyderm,app=dash \
	      -o jsonpath='{.items[0].status.conditions[?(@.type=="Ready")].status}' >/dev/null 2>&1; \
	do \
	    echo -n "."; \
	    sleep 1; \
	done; \
	echo; \
	echo -n "Waiting for the dash to be ready (may take a minute or two)"; \
	until [ "$$( \
	    kubectl get po \
	      -l suite=pachyderm,app=dash \
	      -o jsonpath='{.items[0].status.conditions[?(@.type=="Ready")].status}')" = True ]; \
	do \
	    echo -n "."; \
	    sleep 1; \
	done; \
	echo
	pachctl port-forward &

.PHONY: \
	echo \
	opencv \
	delete \
	minikube

