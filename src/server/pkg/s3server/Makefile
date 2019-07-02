./conformance/s3-tests:
	git submodule update --init --recursive
	cd ./conformance/s3-tests \
		&& ./bootstrap \
		&& source virtualenv/bin/activate \
		&& pip install nose-exclude==0.5.0
