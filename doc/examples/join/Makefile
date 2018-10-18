
setup:
	pachctl create-repo data
	@#pachctl create-pipeline map.json
	@#pachctl create-pipeline join.json
	@#pachctl create-pipeline shuffle.json
	@#pachctl create-pipeline reduce.json

teardown:
	pachctl delete-repo data

generate-input-data:
	./gendata.rb

upload:
	pachctl delete-repo data || true
	pachctl create-repo data
	pachctl start-commit data master
	cat sales1.csv | tail -n+2 | pachctl put-file data master /sales/1.csv 
	cat authors.csv | tail -n+2 | pachctl put-file data master /authors/1.csv 
	cat books.csv | tail -n+2 | pachctl put-file data master /books/1.csv 
	pachctl finish-commit data master

input: generate-input-data upload

docker-build-map:
	docker build . -f Dockerfile.map -t map
	docker tag map pachyderm/map
	docker push pachyderm/map

map: docker-build-map
	pachctl create-pipeline -f map.json || \
		pachctl update-pipeline -f map.json --push-images --reprocess

docker-build-join:
	docker build . -f Dockerfile.join -t join
	docker tag join pachyderm/join
	docker push pachyderm/join

join: docker-build-join
	pachctl create-pipeline -f join.json || \
		pachctl update-pipeline -f join.json --push-images --reprocess

docker-build-shuffle:
	docker build . -f Dockerfile.shuffle -t shuffle
	docker tag shuffle pachyderm/shuffle
	docker push pachyderm/shuffle

shuffle: docker-build-shuffle
	pachctl create-pipeline -f shuffle.json || \
		pachctl update-pipeline -f shuffle.json --push-images --reprocess

docker-build-reduce:
	docker build . -f Dockerfile.reduce -t reduce
	docker tag reduce pachyderm/reduce
	docker push pachyderm/reduce

reduce: docker-build-reduce
	pachctl create-pipeline -f reduce.json || \
		pachctl update-pipeline -f reduce.json --push-images --reprocess

all: input map join shuffle reduce
