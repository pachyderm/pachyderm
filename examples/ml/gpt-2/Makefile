
docker-build:
	docker build -t jdoliner/gpt-2-example:5 .

docker-push: docker-build
	docker push jdoliner/gpt-2-example:5

deploy:
	pachctl update repo queries
	pachctl update pipeline -f tweets.json
	pachctl update pipeline -f train.json
	pachctl update pipeline -f generate.json --reprocess
