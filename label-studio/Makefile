build-test-image:
	git clone https://github.com/pachyderm/label-studio.git
	cd label-studio && docker build -t pachyderm/label-studio:local-testing .
	rm -rf label-studio

run-test-image:
	docker run -d -it --net="host" pachyderm/label-studio:local-testing label-studio -b \
    --username user@domain \
    --password pass \
    --user-token abcdef

run-test-program:
	label-studio -b --username user@domain --password pass --user-token abcdef &

install-test-deps:
	python -m venv env
	. env/bin/activate
	pip install -r requirements.txt

run-tests:
	# Add PACHD_ADDRESS env variable to below line
	LABEL_STUDIO_USER_TOKEN=abcdef python -m pytest tests

docker-test: build-test-image run-test-image install-test-deps run-tests
direct-test: run-test-program install-test-deps run-tests

.PHONY: build-test-image run-test-image install-test-deps run-tests docker-test direct-test
