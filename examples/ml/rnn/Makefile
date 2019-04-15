all: setup run

# Used by Pachyderm CI
# Easiest way to make the test model run shorter is to truncate the input size
test: verify-pachctl gather-input-data
	cat data/ptb.test.txt | head -n 100 > data/ptb.test.txt.short
	mv data/ptb.test.txt.short data/ptb.test.txt
	cat data/ptb.train.txt | head -n 100 > data/ptb.train.txt.short
	mv data/ptb.train.txt.short data/ptb.train.txt
	cat data/ptb.valid.txt | head -n 100 > data/ptb.valid.txt.short
	mv data/ptb.valid.txt.short data/ptb.valid.txt
	rm data/all.txt
	cat data/ptb.test.txt > data/all.txt
	cat data/ptb.train.txt >> data/all.txt
	cat data/ptb.valid.txt >> data/all.txt
	make load-input-data
	make run

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
	pachctl create repo GoT_scripts
	pachctl put file GoT_scripts@master -f ./data/ -r

gather-input-data: get-input-data extract-input-data clean-input-data

verify-pachctl:
	which pachctl

input-data: verify-pachctl gather-input-data load-input-data

docker-build:
	# have to use rsync instead of cp so that we can skip the "doc"
	# directory which contains this path itself
	rsync -av ../../../ pachyderm --exclude doc
	docker build -t tensor_flow_rnn_got .
	rm -rf pachyderm

setup: docker-build input-data

run: verify-pachctl docker-build
	pachctl create pipeline -f pipeline-combined.json

clean: verify-pachctl
	pachctl delete pipeline -f pipeline.json
	pachctl delete repo GoT_scripts 
