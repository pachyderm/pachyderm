
test:
	docker build . -t pachyderm/python-sdk-ci-testing:${USER}-local
	PYTHON_SDK_TESTING_IMAGE=pachyderm/python-sdk-ci-testing:${USER}-local poetry run pytest -vvv tests/test_worker.py

.PHONY: \
	test
