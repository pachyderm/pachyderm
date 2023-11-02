
docs:
	docker build -t pachyderm_sdk_docs -f Dockerfile.docs-generation .
	rm -rf docs/pachyderm_sdk
	tar cf - pachyderm_sdk | docker run -i pachyderm-sdk-docs | tar -xf -

proto:
	./proto/generate-protos.sh

test:
	poetry run pytest -vvv tests

.PHONY: \
	docs proto test
