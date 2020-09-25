.PHONY: docker-image kafka
PROJNAME := pachkafka

# customize these to your kafka deployment and needs
KAFKA_HOST := kafka.kafka
KAFKA_TOPIC := test_topic
KAFKA_GROUP_ID := test_group
KAFKA_PORT := 9092
KAFKA_TIMEOUT := 5
NAMED_PIPE := /pfs/out
VERBOSE_LOGGING := true

PIPELINES_DIR := pipelines
SOURCE_DIR := source
ADDITIONAL_MANIFESTS_DIR := additional_manifests
PROJ_DIRS := $(SOURCE_DIR) $(PIPELINES_DIR)
JQFLAGS :=

# You should set these to appropriate values.
CONTAINER_VERSION := $(shell pachctl version --client-only)
DOCKER_ACCOUNT := pachyderm
CONTAINER_NAME := kafka_spout
CONTAINER_TAG := $(DOCKER_ACCOUNT)/$(CONTAINER_NAME):$(CONTAINER_VERSION)


PIPELINE_TRANSFORM := .transform.image="$(CONTAINER_TAG)"|.transform.env.KAFKA_HOST="$(KAFKA_HOST)"|.transform.env.KAFKA_TOPIC="$(KAFKA_TOPIC)"|.transform.env.KAFKA_GROUP_ID="$(KAFKA_GROUP_ID)"|.transform.env.KAFKA_PORT="$(KAFKA_PORT)"|.transform.env.KAFKA_TIMEOUT="$(KAFKA_TIMEOUT)"|.transform.env.NAMED_PIPE="$(NAMED_PIPE)"|.transform.env.VERBOSE_LOGGING="$(VERBOSE_LOGGING)"
PIPELINE_SOURCE_FILES = $(shell find $(PROJ_DIRS) -type f -name \*.go)
PIPELINE_DEF_FILES = $(shell find $(PROJ_DIRS) -type f -name \*.pipeline)
KAFKA_FILES = $(sort $(shell find $(ADDITIONAL_MANIFESTS_DIR) -type f -name [0-9][0-9][0-9]\*.yaml))
PIPELINE_JSON_FILES = $(patsubst %.pipeline,%.json,$(PIPELINE_DEF_FILES))
DAG := kafka_spout 
#this needs to be recursively expanded because it's used in a foreach loop
PACHCTL_FILE = -f $(PIPELINES_DIR)/$(stage).json


pachctl-pipeline = @$(foreach stage,$(2), pachctl $(1) pipeline $(PACHCTL_FILE) ; )
pachctl-delete-pipeline = @$(foreach stage,$(1), pachctl delete pipeline $(stage) ; )
kubectl-apply = @$(foreach manifest,$(1), kubectl apply -f $(manifest)  ; )
kubectl-delete = @$(foreach manifest,$(1), kubectl delete -f $(manifest)  ; )
reverse = $(if $(1),$(call reverse,$(wordlist 2,$(words $(1)),$(1)))) $(firstword $(1))

%.json: %.pipeline Makefile 
	@jq $(JQFLAGS) '$(PIPELINE_TRANSFORM)'  $< > $@

kafka: $(KAFKA_FILES)
	@$(call kubectl-apply,$(KAFKA_FILES))

docker-image: Dockerfile $(PIPELINE_JSON_FILES) $(PIPELINE_SOURCE_FILES)
	@docker build -t $(CONTAINER_TAG) .
	@docker push $(CONTAINER_TAG)

create-dag: kafka docker-image $(PIPELINE_JSON_FILES) $(PIPELINE_SOURCE_FILES)
	@$(call pachctl-pipeline,create,$(DAG))

update-dag: kafka docker-image $(PIPELINE_JSON_FILES) $(PIPELINE_SOURCE_FILES)
	@$(call pachctl-pipeline,update,$(DAG))

clean:
	-@$(call kubectl-delete,$(KAFKA_FILES))
	-@$(call pachctl-delete-pipeline,$(call reverse,$(DAG)))
	-@rm -f $(PIPELINE_JSON_FILES)


