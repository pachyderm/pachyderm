#!make
# Various config files
CONFIG_FILE=pipeline.conf
SECRETS_YAML=secrets.yaml
ENCODED_ENV=.env.encoded
DOCKER_TEST_ENV=docker.test.env

# Various folders
CONFIG=config
SRC=src
TARGET=target
TEST=test

# Import variables from config file and make them available
include $(CONFIG)/$(CONFIG_FILE)
export

PWD=$(shell pwd)

.PHONY: pipe.config pipe.clean pipe.create pipe.delete
.PHONY: docker.secrets docker.secrets.clean
.PHONY: docker.secrets.registry docker.secrets.registry.clean
.PHONY: docker.secrets.container docker.secrets.container.clean
.PHONY: docker.push docker.build docker.clean docker.test

# Prepares everything for install
all: pipe.config docker.build

install: docker.push pipe.create

test: docker.test

clean: pipe.clean docker.clean
	rm -rf $(TARGET)

# Creates the pipeline configuration file
PIPE_CONFIG=$(TARGET)/$(PIPELINE_BUILD_CONFIG_TARGET_FILE)
$(PIPE_CONFIG):
	mkdir -p $(TARGET); \
	envsubst '$${PIPELINE_NAME} $${PIPELINE_DOCKER_IMAGE} \
	$${PIPELINE_DOCKER_HOME} $${PIPELINE_DOCKER_SCRIPT} \
	$${PIPELINE_DOCKER_SECRETS} $${PIPELINE_DOCKER_SECRETS_PATH} \
	$${PIPELINE_DOCKER_REGISTRY_SECRETS} $${PIPELINE_REPO}' \
	< $(CONFIG)/$(PIPELINE_BUILD_CONFIG_TEMPLATE_FILE) \
	> $(PIPE_CONFIG)
# Just makes it easier to type and test
pipe.config: $(PIPE_CONFIG)
pipe.clean:
	rm -f $(PIPE_CONFIG);

# Creates the pipeline in pachyderm
pipe.create:
	pachctl create pipeline -f $(PIPE_CONFIG)

# Deletes the pipeline from pacyderm
pipe.delete:
	pachctl delete pipeline ${PIPELINE_NAME}

docker.secrets: docker.secrets.registry docker.secrets.container
# Create the secrets to access the docker registry
docker.secrets.registry:
	if kubectl get secret $(PIPELINE_DOCKER_REGISTRY_SECRET); then \
		echo "$$PIPELINE_DOCKER_REGISTRY_SECRET already exists, skipping"; \
	else \
		kubectl create secret docker-registry $(PIPELINE_DOCKER_REGISTRY_SECRET) \
		--docker-server=$(PIPELINE_DOCKER_REGISTRY) \
		--docker-username=$(PIPELINE_DOCKER_REGISTRY_USERNAME) \
		--docker-password=$(PIPELINE_DOCKER_REGISTRY_PASSWORD) \
		--docker-email=$(PIPELINE_DOCKER_REGISTRY_EMAIL); \
	fi
# Cleanup
docker.secrets.registry.clean:
	if kubectl get secret $(PIPELINE_DOCKER_REGISTRY_SECRETS); then \
		kubectl delete secret $(PIPELINE_DOCKER_REGISTRY_SECRETS); \
	fi

# Export env, rename variables to *_ENCODED, envsubst the secrets file
# All this effort because they kubectl requires them to be base64 encoded ...
docker.secrets.container: $(TARGET)/$(ENCODED_ENV) $(TARGET)/$(SECRETS_YAML)
# Export env and rename everything to *_ENCODED
$(TARGET)/$(ENCODED_ENV):
	mkdir -p $(TARGET); \
	env | awk '{split($$0,a,"="); foo="echo \""a[2]"\\c\" | base64"; foo | getline b; \
		close(foo); printf("export %s_ENCODED=%s\n", a[1], b);}' > $(TARGET)/$(ENCODED_ENV); \

$(TARGET)/$(SECRETS_YAML):
	. $(TARGET)/$(ENCODED_ENV); \
	envsubst < $(CONFIG)/$(SECRETS_YAML) > $(TARGET)/$(SECRETS_YAML); \
	if kubectl get secret $(PIPELINE_DOCKER_SECRETS); then \
		echo "$$PIPELINE_DOCKER_SECRETS already exists, skipping"; \
	else \
		kubectl create -f $(TARGET)/$(SECRETS_YAML); \
	fi

docker.secrets.container.clean:
	rm -f $(TARGET)/$(SECRETS_YAML); \
	rm -f $(TARGET)/$(ENCODED_ENV); \
	if kubectl get secret $(PIPELINE_DOCKER_SECRETS); then \
		kubectl delete secret $(PIPELINE_DOCKER_SECRETS); \
	fi

docker.secrets.clean: docker.secrets.container.clean docker.secrets.registry.clean

# Build the docker image, some env variables must be passed along
docker.build:
	docker build -t $(PIPELINE_DOCKER_IMAGE) \
	--label "$(PIPELINE_DOCKER_IMAGE)" \
	--build-arg SRC=$(SRC) \
	--build-arg PIPELINE_HOME=$(PIPELINE_DOCKER_HOME) \
	--build-arg PIPELINE_INPUT=$(PIPELINE_INPUT) \
	--build-arg PIPELINE_OUTPUT=$(PIPELINE_OUTPUT) $(PWD)

docker.push: docker.secrets
	docker push $(PIPELINE_DOCKER_IMAGE)

# Remove docker image(s)
docker.image.clean:
	-docker rmi -f $(shell docker images --filter \
		"label=$(PIPELINE_DOCKER_IMAGE)" -q); \

# Remove docker contaner(s)
docker.container.clean:
	-docker rm -f $(shell docker ps -a --filter \
		"label=$(PIPELINE_DOCKER_IMAGE):test" -q); \

# Remove all trace of docker
docker.clean: docker.secrets.clean docker.container.clean docker.image.clean docker.test.clean

# Performs a test by running the docker container locally and mounting local folders
# Again, make sure all environmental variables are present, e.g. passwords etc.
docker.test.env:
	envsubst < $(CONFIG)/$(DOCKER_TEST_ENV) > $(TEST)/$(DOCKER_TEST_ENV);
docker.test.env.clean:
	rm -f $(TEST)/$(DOCKER_TEST_ENV);

docker.test: docker.build docker.test.env
	docker run \
	--label "$(PIPELINE_DOCKER_IMAGE):test" \
	-e PIPELINE_INPUT=$(PIPELINE_INPUT) \
	-e PIPELINE_OUTPUT=$(PIPELINE_OUTPUT) \
	-e PIPELINE_HOME=$(PIPELINE_DOCKER_HOME) \
	--env-file $(TEST)/$(DOCKER_TEST_ENV) \
	--mount src="$(PWD)"/$(PIPELINE_TEST_INPUT),dst=$(PIPELINE_INPUT),type=bind \
	--mount src="$(PWD)"/$(PIPELINE_TEST_OUTPUT),dst=$(PIPELINE_OUTPUT),type=bind \
	$(PIPELINE_DOCKER_IMAGE) \
	/bin/bash -c "$(PIPELINE_DOCKER_HOME)/$(PIPELINE_DOCKER_SCRIPT)"

docker.test.clean: docker.container.clean docker.test.env.clean

# Display some information about the pipeline
verify:
	pachctl inspect pipeline $(PIPELINE_NAME); \
	pachctl list file $(PIPELINE_NAME)@master

