docker-image:
	docker build -t pachyderm/saml-idp .

push-docker-image: docker-image
	docker push pachyderm/saml-idp:latest

.PHONY: docker-image push-docker-image
