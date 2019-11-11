.PHONY: 
PROJNAME := pachkafka

# customize these to your kafka deployment and needs
KAFKA_HOST := kafka.kafka
KAFKA_TOPIC := test
KAFKA_GROUP_ID := test3
KAFKA_PORT := 9092
KAFKA_TIMEOUT := 5
NAMED_PIPE := /pfs/out
VERBOSE_LOGGING := true

PIPELINES_DIR := pipelines
SOURCE_DIR := source
ADDITIONAL_MANIFESTS_DIR := additional_manifests
PROJ_DIRS := $(SOURCE_DIR) $(PIPELINES_DIR)
CONTAINER_VERSION := 0.50
DOCKER_ACCOUNT := papposilenus
CONTAINER_NAME := kafka_spout
CONTAINER_TAG := $(DOCKER_ACCOUNT)/$(CONTAINER_NAME):$(CONTAINER_VERSION)
JQ := jq
JQFLAGS :=
DOCKER := docker
ECHO := echo
TOUCH := touch
CP := cp
RM := rm
PIPELINE_TRANSFORM := .transform.image="$(CONTAINER_TAG)"|.transform.env.KAFKA_HOST="$(KAFKA_HOST)"|.transform.env.KAFKA_TOPIC="$(KAFKA_TOPIC)"|.transform.env.KAFKA_GROUP_ID="$(KAFKA_GROUP_ID)"|.transform.env.KAFKA_PORT="$(KAFKA_PORT)"|.transform.env.KAFKA_TIMEOUT="$(KAFKA_TIMEOUT)"|.transform.env.NAMED_PIPE="$(NAMED_PIPE)"|.transform.env.VERBOSE_LOGGING="$(VERBOSE_LOGGING)"
PIPELINE_SOURCE_FILES = $(shell find $(PROJ_DIRS) -type f -name \*.go)
PIPELINE_DEF_FILES = $(shell find $(PROJ_DIRS) -type f -name \*.pipeline)
KAFKA_FILES = $(sort $(shell find $(ADDITIONAL_MANIFESTS_DIR) -type f -name [0-9][0-9][0-9]\*.yaml))
PIPELINE_JSON_FILES := $(patsubst %.pipeline,%.json,$(PIPELINE_DEF_FILES))
DAG := kafka_spout 
PACHCTL := pachctl
PACHCTL_FLAGS :=
#this needs to be recursively expanded because it's used in a foreach loop
PACHCTL_FILE = -f $(PIPELINES_DIR)/$(stage).json

KUBECTL := kubectl
KUBECTL_FLAGS :=

pachctl-pipeline = @$(foreach stage,$(2), $(PACHCTL) $(1) pipeline $(PACHCTL_FILE) $(PACHCTL_FLAGS) ; )
kubectl-apply = @$(foreach manifest,$(1), $(KUBECTL) apply -f $(manifest) $(KUBECTL_FLAGS) ; )

%.json: %.pipeline Makefile 
	@$(JQ) $(JQFLAGS) '$(PIPELINE_TRANSFORM)'  $< > $@

kafka: $(KAFKA_FILES)
	@$(TOUCH) $@
	@$(call kubectl-apply,$(KAFKA_FILES))

docker-image: Dockerfile $(PIPELINE_JSON_FILES) $(PIPELINE_SOURCE_FILES)
	@$(TOUCH) $@
	@$(CP) -R ../../../vendor/github.com/segmentio/kafka-go $(SOURCE_DIR)
	@$(DOCKER) build -t $(CONTAINER_TAG) .|| { $(RM) -r docker-image $(SOURCE_DIR)/kafka-go; exit 1; }
	@$(DOCKER) push $(CONTAINER_TAG)
	@$(RM) -r $(SOURCE_DIR)/kafka-go	

create-dag: kafka docker-image $(PIPELINE_JSON_FILES) $(PIPELINE_SOURCE_FILES)
	@$(call pachctl-pipeline,create,$(DAG))

update-dag: kafka docker-image $(PIPELINE_JSON_FILES) $(PIPELINE_SOURCE_FILES)
	@$(call pachctl-pipeline,update,$(DAG))

clean:
	-@$(RM) $(wildcard $(PIPELINE_JSON_FILES) docker-image secrets kafka)


