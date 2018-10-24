BUILD_IMAGE_NAME := pachyderm/saml-idp
IDP_IMAGE_NAME := pachyderm/saml-idp

clean:
	rm -rf ./_idp ./_build || true

docker-image:
	@make clean
	mkdir _idp _build
	cp ./*.go ./_build
	cp ./Dockerfile.build ./_build/Dockerfile
	docker build -t $(BUILD_IMAGE_NAME) ./_build
	docker run -t -a STDOUT -a STDERR \
		-v $(realpath $(PWD)/../vendor):/go/src/vendor \
		-v $(realpath $(PWD)/../../..):/go/src/vendor/github.com/pachyderm/pachyderm \
		-v $(PWD):/out \
		$(BUILD_IMAGE_NAME)
	mv ./saml-idp ./_idp
	cp Dockerfile.idp ./_idp/Dockerfile
	docker build -t $(IDP_IMAGE_NAME) ./_idp
	@make clean

push-docker-image: docker-image
	docker push $(IDP_IMAGE_NAME):latest

.PHONY: docker-image push-docker-image clean
