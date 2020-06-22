
docker-build:
	docker build -t pachyderm/gpt-2-example .
	docker build -t pachyderm/tweet-scraper tweets

docker-push: docker-build
	docker push pachyderm/gpt-2-example

deploy:
	pachctl update repo queries
	pachctl create pipeline -f tweets.json
	pachctl create pipeline -f train.json
	pachctl create pipeline -f generate.json
