
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

.PHONY: launch-dev launch-prod setup-auth test install build ci prune-deps circle-docker-build circle-docker-manifest e2e

bunyan:
	npm exec --prefix backend bunyan

launch-dev:
	npm run start:dev --prefix ./backend & npm run start --prefix ./frontend

launch-prod:
	npm run build --prefix ./frontend && npm run build --prefix ./backend && npm start --prefix ./backend

setup-auth:
	npm run setupAuth:local

setup-ci-auth: 
	npm run setupAuth

install:
	npm install && \
	npm install --prefix ./backend && \
	npm install --prefix ./frontend && \
	npm install --prefix ./cypress

prune-deps:
	rm -rf ./frontend/node_modules && npm ci --prefix ./backend --only=production

docker-ci:
	npm ci --prefix ./backend && npm ci --prefix ./frontend

clean-deps:
	rm -rf ./node_modules ./frontend/node_modules ./backend/node_modules ./cypress/node_modules

build:
	npm run build --prefix ./frontend && npm run build --prefix ./backend

test:
	npm test --prefix ./backend && npm test --prefix ./frontend

circle-docker-build:
	cd .. && docker buildx create --name hab --driver docker-container
	cd .. && docker buildx build --provenance false --builder hab --platform linux/$(TARGET_ARCH) --build-arg DOCKER_TAG=$(DOCKER_TAG) --push -t pachyderm/haberdashery:$(DOCKER_TAG)-$(TARGET_ARCH) -f Dockerfile.console .

circle-docker-manifest:
	docker manifest create pachyderm/haberdashery:$(DOCKER_TAG) docker.io/pachyderm/haberdashery:$(DOCKER_TAG)-amd64 docker.io/pachyderm/haberdashery:$(DOCKER_TAG)-arm64
	docker manifest push pachyderm/haberdashery:$(DOCKER_TAG)

e2e-ce: e2e

e2e: 
	npm run cypress:local

e2e-auth: 
	npm run cypress:local-auth
	