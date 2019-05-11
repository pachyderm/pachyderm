.PHONY: kafka

build-docker-images:
	docker build -t pachyderm/kafka-example-consumer -f ./consumer.Dockerfile .
	docker build -t pachyderm/kafka-example-producer -f ./producer.Dockerfile .
	docker push pachyderm/kafka-example-consumer
	docker push pachyderm/kafka-example-producer

clean:
	kubectl delete -f ./producer.yaml
	cd ../../.. && make clean-launch-kafka

launch:
	cd ../../.. && make launch-kafka
	kubectl apply -f ./producer.yaml
	pachctl create pipeline -f ./consumer.json
