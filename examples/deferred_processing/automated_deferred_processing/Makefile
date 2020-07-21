
# The executable paths below are set to generic values.
# Modify them for your system by setting environment variables, using a command like this
# to both fetch the version of pachyderm you want to use and execute it.
# "../etc/fetch_release_pachctl.py 1.10.0 ; env PACHCTL=${GOPATH}/bin/pachctl make -e opencv"

SHELL := /bin/bash
PACHCTL := pachctl
KUBECTL := kubectl
DOCKER := docker

DEFERRED_PROCESSING_PIPELINES = ../deferred_processing_plus_transactions/edges_dp.json \
				../deferred_processing_plus_transactions/montage_dp.json 

DEFERRED_PROCESSING_1_FILES = ../deferred_processing_plus_transactions/images.txt \
				../deferred_processing_plus_transactions/images2.txt

DEFERRED_PROCESSING_2_FILES = ../deferred_processing_plus_transactions/images3.txt


CONTAINER_VERSION := 1.11.0
DOCKER_ACCOUNT := pachyderm
CONTAINER_NAME := pachctl
CONTAINER_TAG := $(DOCKER_ACCOUNT)/$(CONTAINER_NAME):$(CONTAINER_VERSION)

# The pipeline uses a timed auth token for authentication.
# The auth token uses a golang duration, which allows h for hours, m for minutes, s for seconds.
# This default is 26 weeks.
# The token will only be valid for this duration if it is *shorter* than the lifetime of the session
# for the user who's logged into the cluster from where this makefile is run.
# See pachctl auth use-auth-token --help for more details.
# If you'd like to experiment with other TTLs, you can customize the value like this:
# env AUTH_TOKEN_TTL=1h make -e create-branch-mover
AUTH_TOKEN_TTL := 624h

# these are used by jq in  rules below to transform manifests and specs
AUTH_TOKEN_TRANSFORM = ".data.auth_token=\"$$(${PACHCTL} auth get-auth-token --ttl "${AUTH_TOKEN_TTL}" | grep Token | awk '{print $$2}' | base64 -e | tr -d '\r\n')\""

PIPELINE_TRANSFORM := .transform.image=\"$(CONTAINER_TAG)\"

# Create your own json patch using these instructions:
# https://docs.pachyderm.com/latest/how-tos/mount-volume/
# This patch adds environment variables in addition to
# mounting a volume from a secret.
json-patch.jpatch: json-patch.template commands.sed
	sed -E -f commands.sed json-patch.template | tr '\n' ' ' > $@

branch-mover.json: branch-mover.pipeline json-patch.jpatch pachyderm-user-secret.secret
	jq "$(PIPELINE_TRANSFORM)|.pod_patch=\"$$(cat json-patch.jpatch)\"" branch-mover.pipeline > $@

branch-mover-no-auth.json: branch-mover.pipeline 
	jq "$(PIPELINE_TRANSFORM)" branch-mover.pipeline > $@

create-deferred-processing-cluster: $(DEFERRED_PROCESSING_PIPELINES) $(DEFERRED_PROCESSING_1_FILES) $(DEFERRED_PROCESSING_2_FILES)
	${PACHCTL} create repo images_dp_1
	${PACHCTL} create repo images_dp_2
	cat $(DEFERRED_PROCESSING_PIPELINES) | ${PACHCTL} create pipeline -f -
	for file in $(DEFERRED_PROCESSING_1_FILES) ; \
	do \
	${PACHCTL} put file images_dp_1@master -i $$file; \
	done
	for file in $(DEFERRED_PROCESSING_2_FILES) ; \
	do \
	${PACHCTL} put file images_dp_2@master -i $$file; \
	done


delete-deferred-processing-cluster:
	-@${PACHCTL} delete pipeline montage_dp
	-@${PACHCTL} delete pipeline edges_dp
	-@${PACHCTL} delete repo images_dp_1
	-@${PACHCTL} delete repo images_dp_2

create-branch-mover: branch-mover.json
	${PACHCTL} create pipeline -f $<

update-branch-mover: branch-mover.json
	${PACHCTL} update pipeline -f $<

delete-branch-mover:
	-@rm -f branch-mover.json
	-@${PACHCTL} delete pipeline branch-mover

create-branch-mover-no-auth: branch-mover-no-auth.json
	${PACHCTL} create pipeline -f $<

update-branch-mover-no-auth: branch-mover-no-auth.json
	${PACHCTL} update pipeline -f $<

delete-branch-mover-no-auth:
	-@rm -f branch-mover-no-auth.json
	-@${PACHCTL} delete pipeline branch-mover

pachyderm-user-secret.secret: pachyderm-user-secret.clear
	jq "$(AUTH_TOKEN_TRANSFORM)" $< > $@
	-${PACHCTL} delete secret $$(jq .metadata.name $@ | tr -d '"')
	${PACHCTL} create secret -f $@

clean: delete-branch-mover delete-deferred-processing-cluster 
	-@rm -f  branch-mover*.json pachyderm-user-secret.secret json-patch.jpatch


