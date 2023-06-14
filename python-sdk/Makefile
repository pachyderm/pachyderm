
test: # For running tests locally
	docker build . -t pachyderm/python-sdk-ci-testing:${USER}-local -f Dockerfile.datum-batching-test
	PYTHON_SDK_TESTING_IMAGE=pachyderm/python-sdk-ci-testing:${USER}-local poetry run pytest -vvv tests/test_worker.py

.PHONY: \
	test
