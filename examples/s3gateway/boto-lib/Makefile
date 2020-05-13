.PHONY: init run

venv:
	virtualenv -v venv -p python3
	. venv/bin/activate && pip install -r requirements.txt

init: venv
	cd ../../opencv && make opencv

run: venv
	. venv/bin/activate && ./main.py
