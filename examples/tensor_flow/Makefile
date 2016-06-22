all: setup run

get-input-data:
	curl -o dataset.tar.gz http://www.fit.vutbr.cz/~imikolov/rnnlm/simple-examples.tgz

extract-input-data:
	gunzip dataset.tar.gz
	tar -xvf dataset.tar
	rm -rf data
	mv simple-examples data

clean-input-data:
	rm dataset.tar

load-input-data:
	pachctl create-repo tensor_ptb_data
	pachctl start-commit tensor_ptb_data > commitID
	umount ./mnt || true
	rm -rf ./mnt
	mkdir mnt
	pachctl mount ./mnt &
	sleep 5
	cp -R data/data/* ./mnt/tensor_ptb_data/`cat commitID`/
	pachctl finish-commit tensor_ptb_data `cat commitID`
	rm commitID

verify-pachctl:
	which pachctl

input-data: verify-pachctl get-input-data extract-input-data clean-input-data load-input-data

docker-build:
	docker build -t tensor_flow_rnn .

setup: docker-build input-data

run: verify-pachctl docker-build
	pachctl create-pipeline -f pipeline.json

clean: verify-pachctl
	pachctl delete-pipeline -f pipeline.json
	pachctl delete-repo tensor_ptb_data
