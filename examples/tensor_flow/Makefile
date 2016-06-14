

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
	pachctl delete-repo tensor_ptb || true
	pachctl create-repo tensor_ptb || true
	pachctl start-commit tensor_ptb > commitID
	rm -rf ./mnt
	mkdir mnt
	pachctl mount ./mnt &
	cp -R data/data ./mnt/tensor_ptb/`cat commitID`/
	pachctl finish-commit tensor_ptb `cat commitID`

verify-pachctl:
	which pachctl

input-data: verify-pachctl get-input-data extract-input-data clean-input-data load-input-data

docker-build:
	docker build -t tensor_flow_rnn .


