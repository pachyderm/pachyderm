./conformance/s3-tests:
	git submodule update --init --recursive
	cd ./conformance/s3-tests \
		&& ./bootstrap \
		&& source virtualenv/bin/activate \
		&& pip install nose-exclude==0.5.0

./integration/python/venv:
	virtualenv -v integration/python/venv -p python3
	cd integration/python && . venv/bin/activate && pip install -r requirements.txt
