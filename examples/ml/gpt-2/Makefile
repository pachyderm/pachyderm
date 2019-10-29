
docker-build:
	docker build --no-cache -t pachyderm/gpt-2-example .

docker-push: docker-build
	docker push pachyderm/gpt-2-example

deploy:
	pachctl update repo queries
	pachctl create pipeline -f tweets.json
	pachctl create pipeline -f train.json
	pachctl create pipeline -f generate.json
