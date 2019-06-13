
docker-build:
	docker build -t pachyderm/gpt-2-example .

docker-push: docker-build
	docker push pachyderm/gpt-2-example

deploy:
	pachctl update repo queries
	pachctl update pipeline -f tweets.json
	pachctl update pipeline -f train.json
	pachctl update pipeline -f generate.json --reprocess
