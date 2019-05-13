
docker-build:
	docker build -t gpt-2-example .

deploy:
	pachctl update repo users
	pachctl update pipeline -f tweets.json
	pachctl update pipeline -f train.json
	pachctl update pipeline -f generate.json
