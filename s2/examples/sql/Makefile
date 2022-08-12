.PHONY: run test integration-test conformance-test

run:
	go run main.go

test: integration-test conformance-test

integration-test:
	cd ../.. && make ./integration/python/venv
	../../integration/integration.py http://localhost:8080 --access-key=homer --secret-key=donuts

conformance-test:
	cd ../.. && make ./conformance/s3-tests
	../../conformance/conformance.py --s3tests-config=test/s3tests.yaml --ignore-config=test/ignore.conf --runs-dir=test/runs
