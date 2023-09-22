
docker-build:
	docker build -t pachyderm/gpt-2-example .

docker-push: docker-build
	docker push pachyderm/gpt-2-example

deploy:
	pachctl update repo queries
	pachctl create pipeline -f tweets.pipeline.json
	pachctl create pipeline -f train.pipeline.json
	pachctl create pipeline -f generate.pipeline.json
