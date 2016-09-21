all: setup run

get-input-data:
	curl -o dataset.tar.gz https://pachyderm.io/data/GoT-scripts.tar.gz

extract-input-data:
	gunzip dataset.tar.gz
	tar -xvf dataset.tar
	rm -rf data
	mkdir data
	mv *.txt data

clean-input-data:
	rm dataset.tar

load-input-data:
	pachctl create-repo GoT_scripts
	pachctl start-commit GoT_scripts > commitID
	# For now, use the CLI to input the data because of OS X Fuse bug (https://github.com/pachyderm/pachyderm/issues/629)
	# this is a heavy handed way to do a copy:
	ls -1 data/* | while read -r file; do cat $$file | pachctl put-file GoT_scripts `cat commitID` `basename $$file`; done
	pachctl finish-commit GoT_scripts `cat commitID`
	rm commitID

gather-input-data: get-input-data extract-input-data clean-input-data

verify-pachctl:
	which pachctl

input-data: verify-pachctl gather-input-data load-input-data

docker-build:
	docker build -t tensor_flow_rnn_got .

setup: docker-build input-data

run: verify-pachctl docker-build
	pachctl create-pipeline -f pipeline.json

clean: verify-pachctl
	pachctl delete-pipeline -f pipeline.json
	pachctl delete-repo GoT_scripts 
