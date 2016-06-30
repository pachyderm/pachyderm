all: setup run

get-input-data:
	curl -o dataset.tar.gz https://dl.dropboxusercontent.com/u/3940085/scratch/hiphop-lyrics.tar.gz

extract-input-data:
	gunzip dataset.tar.gz
	tar -xvf dataset.tar
	rm -rf data
	mkdir data
	mv *.txt data

clean-input-data:
	rm dataset.tar

load-input-data:
	pachctl create-repo hiphop_lyrics
	pachctl start-commit hiphop_lyrics > commitID
	umount ./mnt || true
	rm -rf ./mnt
	mkdir mnt
	pachctl mount ./mnt &
	sleep 5
	cp data/* ./mnt/hiphop_lyrics/`cat commitID`/
	pachctl finish-commit hiphop_lyrics `cat commitID`
	rm commitID

gather-input-data: get-input-data extract-input-data clean-input-data

verify-pachctl:
	which pachctl

input-data: verify-pachctl gather-input-data load-input-data

docker-build:
	docker build -t tensor_flow_rnn_hiphop .

setup: docker-build input-data

run: verify-pachctl docker-build
	pachctl create-pipeline -f pipeline.json

clean: verify-pachctl
	pachctl delete-pipeline -f pipeline.json
	pachctl delete-repo hiphop_lyrics 
