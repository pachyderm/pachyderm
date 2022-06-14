
UNAME_P := $(shell uname -p)
ifeq ($(UNAME_P),x86_64)
	TARGET_ARCH = amd64
endif
ifneq ($(filter %86,$(UNAME_P)),)
	TARGET_ARCH = amd64
endif
ifneq ($(filter arm%,$(UNAME_P)),)
	TARGET_ARCH = arm64
endif
ifneq ($(filter aarch%,$(UNAME_P)),)
	TARGET_ARCH = arm64
endif

ifdef CIRCLE_TAG 
	DOCKER_TAG = $(CIRCLE_TAG)
else
	DOCKER_TAG = $(CIRCLE_SHA1)
endif

.PHONY: launch-dev launch-prod setup-auth test graphql install build ci prune-deps circle-docker-build circle-docker-manifest e2e

launch-dev:
	npm run start:dev --prefix ./backend & npm run start --prefix ./frontend

launch-prod:
	npm run build --prefix ./frontend && npm run build --prefix ./backend && npm start --prefix ./backend

launch-mock:
	LOG_LEVEL=debug npm run mock-and-start --prefix ./backend & npm run start:mock  --prefix ./frontend

setup-auth:
	npm run setup:local

setup-ci-auth: 
	npm run setup

install:
	npm install && npm install --prefix ./backend && npm install --prefix ./frontend

ci:
	npm ci && npm ci --prefix ./backend && npm ci --prefix ./frontend

prune-deps:
	rm -rf ./frontend/node_modules && npm ci --prefix ./backend --only=production

docker-ci:
	npm ci --prefix ./backend && npm ci --prefix ./frontend

clean-deps:
	rm -rf ./node_modules && rm -rf ./frontend/node_modules && rm -rf ./backend/node_modules

build:
	npm run build --prefix ./frontend && npm run build --prefix ./backend

graphql:
	npm run generate-types && npm run lint --prefix ./backend -- --fix && npm run lint:js --prefix ./frontend -- --fix

test:
	npm test --prefix ./backend && npm test --prefix ./frontend

circle-docker-build:
	mkdir -p /tmp/docker-cache
	echo "$(DOCKERHUB_PASS)" | docker login -u "$(DOCKERHUB_USERNAME)" --password-stdin
	docker buildx create --name hab --driver docker-container
	docker buildx build --builder hab --platform linux/$(TARGET_ARCH) --cache-from=type=local,src=/tmp/docker-cache --cache-to=type=local,dest=/tmp/docker-cache --push -t pachyderm/haberdashery:$(DOCKER_TAG)-$(TARGET_ARCH) .

circle-docker-manifest:
	echo "$(DOCKERHUB_PASS)" | docker login -u "$(DOCKERHUB_USERNAME)" --password-stdin
	docker manifest create pachyderm/haberdashery:$(DOCKER_TAG) docker.io/pachyderm/haberdashery:$(DOCKER_TAG)-amd64 docker.io/pachyderm/haberdashery:$(DOCKER_TAG)-arm64
	docker manifest push pachyderm/haberdashery:$(DOCKER_TAG)
e2e: 
	npm run cypress:local
